package containerregistry

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	cdl "github.com/sigstore/cosign/v2/cmd/cosign/cli/download"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
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

func (c *Client) VersionFromSbom(mainPkg string, repo string) (string, error) {
	type Package struct {
		Name        string `json:"name"`
		VersionInfo string `json:"versionInfo"`
	}
	type Packages struct {
		Packages []Package `json:"packages"`
	}
	buf := new(bytes.Buffer)
	kc := authn.NewMultiKeychain(
		authn.DefaultKeychain,
	)

	o := &options.RegistryOptions{Keychain: kc}
	do := &options.SBOMDownloadOptions{Platform: "linux/amd64"}
	sboms, err := cdl.SBOMCmd(context.TODO(), *o, *do, repo, buf)
	if err != nil {
		return "", err
	}
	var ps Packages

	for _, s := range sboms {
		json.Unmarshal([]byte(s), &ps)
		for _, v := range ps.Packages {
			if v.Name == mainPkg {
				return v.VersionInfo, nil
			}
		}
	}
	return "", nil
}

func (c *Client) getAuthOpt() crane.Option {
	kc := authn.NewMultiKeychain(
		authn.DefaultKeychain,
	)
	return crane.WithAuthFromKeychain(kc)
}
