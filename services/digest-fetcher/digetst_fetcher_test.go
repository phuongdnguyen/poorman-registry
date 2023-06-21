package digestfetcher

import (
	"testing"
	"time"

	"github.com/nduyphuong/reverse-registry/config"
	"github.com/nduyphuong/reverse-registry/inject"
	"github.com/sirupsen/logrus"
	"github.com/test-go/testify/assert"
)

func TestFetch(t *testing.T) {
	log := logrus.New()
	images := make([]config.Image, 0)
	images = append(images, config.Image{
		Name:        "cgr.dev/chainguard/nginx",
		Constraint:  "^1.2.*",
		MainPackage: "nginx",
	})

	conf := config.Config{
		DBConfig: config.MysqlConfig{
			Host:     "localhost",
			User:     "root",
			Password: "my-secret-pw",
			DBName:   "test",
		},
		Images:              images,
		WorkerFetchInterval: "10s",
	}
	d, err := time.ParseDuration(conf.WorkerFetchInterval)
	storage, err := inject.GetStorage(conf)
	assert.NoError(t, err)
	registryClient, err := inject.GetContainerRegistryClient()
	assert.NoError(t, err)
	assert.NoError(t, err)
	fetcher := New(Options{
		Storage:       storage,
		Registry:      registryClient,
		Log:           log,
		FetchInterval: d,
	})
	fetcher.Fetch(conf.Images)
}
