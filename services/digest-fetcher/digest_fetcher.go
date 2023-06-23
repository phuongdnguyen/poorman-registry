package digestfetcher

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nduyphuong/reverse-registry/config"
	repository "github.com/nduyphuong/reverse-registry/repository"
	containerregistry "github.com/nduyphuong/reverse-registry/services/container-registry"
	"github.com/nduyphuong/reverse-registry/utils"
	"github.com/sirupsen/logrus"
)

type Interface interface {
	Fetch([]config.Image) error
}

type client struct {
	storage       repository.Interface
	log           *logrus.Logger
	registry      containerregistry.Interface
	fetchInterval time.Duration
}

type Options struct {
	Storage       repository.Interface
	Registry      containerregistry.Interface
	Log           *logrus.Logger
	FetchInterval time.Duration
}

func New(opt Options) Interface {
	return &client{
		storage:       opt.Storage,
		registry:      opt.Registry,
		log:           opt.Log,
		fetchInterval: opt.FetchInterval,
	}
}

type Index struct {
	Manifests []Manifest `json:"manifests"`
}

type Manifest struct {
	Digest   string   `json:"digest"`
	Platform Platform `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
}

func (c *client) Fetch(images []config.Image) error {
	for {
		var wg sync.WaitGroup
		for _, v := range images {
			v := v
			wg.Add(1)
			go func() {
				defer wg.Done()
				idx, err := c.registry.ManifestOrIndex(v.Name)
				if err != nil {
					c.log.Errorf("fetching manifest or index %v", err)
				}
				var i Index
				json.Unmarshal(idx, &i)
				c.log.Debugf("unmarshalled index: %s", i)
				for _, k := range i.Manifests {
					if k.Platform.Architecture == "amd64" {
						c.log.Info("begin loop")

						nameFromRepo := utils.SplitAndGetLast("/", v.Name)
						mainPkgName, err := utils.SelectNotEmpty(nameFromRepo, v.MainPackage)
						if err != nil {
							c.log.Errorf("can not construct main package name %v", err)
						}
						tag, err := c.registry.VersionFromSbom(mainPkgName, v.Name)
						if err != nil {
							c.log.Errorf("version from sbom %v", err)
							return
						}

						img := utils.MakeImageName(v.Name, tag)

						digest := sha256.Sum256(idx)
						if err != nil {
							c.log.Errorf("hash index %v", err)
						}
						if err := c.storage.SaveDigest(img, "sha256:"+fmt.Sprintf("%x", digest)); err != nil {
							c.log.Errorf("save digest to db %v", err)
							break
						}
						c.log.Infof("saved to db %s %s", v.Name, tag)
					}
				}
			}()
		}
		wg.Wait()
		c.log.Infof("sleep for %v", c.fetchInterval)
		time.Sleep(c.fetchInterval)
	}
}
