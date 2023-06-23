package containerregistry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
)

type Interface interface {
	Head(imageName string) error
	ManifestOrIndex(repoName string) ([]byte, error)
	ListTagsWithConstraint(repoName, constraint string) ([]string, error)
	VersionFromSbom(mainPkg, repo string) (string, error)
}

type Client struct {
}

func New() Interface {
	return &Client{}
}

func (c *Client) Head(imageName string) error {
	opts := c.getAuthOpt()
	if _, err := crane.Head(imageName, opts); err != nil {
		return err
	}
	return nil
}

func (c *Client) ManifestOrIndex(image string) ([]byte, error) {
	opts := c.getAuthOpt()
	return crane.Manifest(image, opts)
}

func (c *Client) ListTagsWithConstraint(repoName string, constraint string) ([]string, error) {
	result := make([]string, 0)
	tags, err := crane.ListTags(repoName)
	if err != nil {
		return nil, err
	}
	r, _ := regexp.Compile(constraint)
	for _, tag := range tags {
		if r.MatchString(tag) {
			result = append(result, tag)
		}
	}
	return result, nil
}

func (c *Client) VersionFromSbom(mainPkg string, imageRef string) (string, error) {
	ctx := context.TODO()
	// type Package struct {
	// 	Name        string `json:"name"`
	// 	VersionInfo string `json:"versionInfo"`
	// }
	// type Packages struct {
	// 	Packages []Package `json:"packages"`
	// }
	// buf := new(bytes.Buffer)
	kc := authn.NewMultiKeychain(
		authn.DefaultKeychain,
	)

	regOpts := options.RegistryOptions{Keychain: kc}
	// do := &options.SBOMDownloadOptions{Platform: "linux/amd64"}
	// sboms, err := cdl.SBOMCmd(context.TODO(), *o, *do, repo, buf)
	// if err != nil {
	// 	return "", err
	// }
	// var ps Packages

	// for _, s := range sboms {
	// 	json.Unmarshal([]byte(s), &ps)
	// 	for _, v := range ps.Packages {
	// 		if v.Name == mainPkg {
	// 			return v.VersionInfo, nil
	// 		}
	// 	}
	// }
	attOptions := &options.AttestationDownloadOptions{
		PredicateType: "https://slsa.dev/provenance/v1",
		Platform:      "linux/amd64",
	}
	ref, err := name.ParseReference(imageRef, regOpts.NameOptions()...)
	if err != nil {
		return "", err
	}
	ociremoteOpts, err := regOpts.ClientOpts(ctx)
	if err != nil {
		return "", err
	}

	var predicateType string
	predicateType, err = options.ParsePredicateType(attOptions.PredicateType)
	if err != nil {
		return "", err
	}

	se, err := ociremote.SignedEntity(ref, ociremoteOpts...)
	if err != nil {
		return "", err
	}

	idx, isIndex := se.(oci.SignedImageIndex)

	// We only allow --platform on multiarch indexes
	if attOptions.Platform != "" && !isIndex {
		return "", fmt.Errorf("specified reference is not a multiarch image")
	}

	if attOptions.Platform != "" && isIndex {
		targetPlatform, err := v1.ParsePlatform(attOptions.Platform)
		if err != nil {
			return "", fmt.Errorf("parsing platform: %w", err)
		}
		platforms, err := getIndexPlatforms(idx)
		if err != nil {
			return "", fmt.Errorf("getting available platforms: %w", err)
		}

		platforms = matchPlatform(targetPlatform, platforms)
		if len(platforms) == 0 {
			return "", fmt.Errorf("unable to find an attestation for %s", targetPlatform.String())
		}
		if len(platforms) > 1 {
			return "nil", fmt.Errorf(
				"platform spec matches more than one image architecture: %s",
				platforms.String(),
			)
		}

		nse, err := idx.SignedImage(platforms[0].hash)
		if err != nil {
			return "", fmt.Errorf("searching for %s image: %w", platforms[0].hash.String(), err)
		}
		if nse == nil {
			return "", fmt.Errorf("unable to find image %s", platforms[0].hash.String())
		}
		se = nse
	}

	attestations, err := cosign.FetchAttestations(se, predicateType)
	if err != nil {
		return "", err
	}
	if len(attestations) > 1 {
		return "", fmt.Errorf("filtered attestation list is more than one")
	}
	var a attestationPayload
	att := attestations[0]
	pB, err := base64.StdEncoding.DecodeString(att.PayLoad)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(pB, &a); err != nil {
		return "", err
	}
	return a.Predicate.BuildDefinition.InternalParameters[mainPkg], nil
}

type attestationPayload struct {
	Predicate predicate `json:"predicate"`
}

type predicate struct {
	BuildDefinition buildDefinition `json:"buildDefinition"`
}

type buildDefinition struct {
	InternalParameters map[string]string `json:"internalParameters"`
}

//  cosign download attestation cgr.dev/chainguard/redis --predicate-type https://slsa.dev/provenance/v1  | jq '.payload | @base64d | fromjson | .predicate.buildDefinition.internalParameters.redis'

func (c *Client) getAuthOpt() crane.Option {
	kc := authn.NewMultiKeychain(
		authn.DefaultKeychain,
	)
	return crane.WithAuthFromKeychain(kc)
}

func getIndexPlatforms(idx oci.SignedImageIndex) (platformList, error) {
	im, err := idx.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("fetching index manifest: %w", err)
	}

	platforms := platformList{}
	for _, m := range im.Manifests {
		if m.Platform == nil {
			continue
		}
		platforms = append(platforms, struct {
			hash     v1.Hash
			platform *v1.Platform
		}{m.Digest, m.Platform})
	}
	return platforms, nil
}

// matchPlatform filters a list of platforms returning only those matching
// a base. "Based" on ko's internal equivalent while it moves to GGCR.
// https://github.com/google/ko/blob/e6a7a37e26d82a8b2bb6df991c5a6cf6b2728794/pkg/build/gobuild.go#L1020
func matchPlatform(base *v1.Platform, list platformList) platformList {
	ret := platformList{}
	for _, p := range list {
		if base.OS != "" && base.OS != p.platform.OS {
			continue
		}
		if base.Architecture != "" && base.Architecture != p.platform.Architecture {
			continue
		}
		if base.Variant != "" && base.Variant != p.platform.Variant {
			continue
		}

		if base.OSVersion != "" && p.platform.OSVersion != base.OSVersion {
			if base.OS != "windows" {
				continue
			} else { //nolint: revive
				if pcount, bcount := strings.Count(base.OSVersion, "."), strings.Count(p.platform.OSVersion, "."); pcount == 2 && bcount == 3 {
					if base.OSVersion != p.platform.OSVersion[:strings.LastIndex(p.platform.OSVersion, ".")] {
						continue
					}
				} else {
					continue
				}
			}
		}
		ret = append(ret, p)
	}

	return ret
}

type platformList []struct {
	hash     v1.Hash
	platform *v1.Platform
}

func (pl *platformList) String() string {
	r := []string{}
	for _, p := range *pl {
		r = append(r, p.platform.String())
	}
	return strings.Join(r, ", ")
}
