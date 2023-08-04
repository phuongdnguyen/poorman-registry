package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nduyphuong/reverse-registry/config"
	"github.com/nduyphuong/reverse-registry/handler"
	"github.com/nduyphuong/reverse-registry/inject"
	digestfetcher "github.com/nduyphuong/reverse-registry/services/digest-fetcher"
	"github.com/nduyphuong/reverse-registry/utils"
	"github.com/sirupsen/logrus"
)

func RunAPI(conf config.Config) error {
	router, handlerFactory, err := setupRouterAndHandlerFactory(conf)
	if err != nil {
		return err
	}

	setupRoutes(router, handlerFactory)

	port := getPort()
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		return err
	}
	return nil
}

func RunFetcher(conf config.Config) error {
	storage, registryClient, err := setupStorageAndRegistryClient(conf)
	if err != nil {
		return err
	}

	d, err := time.ParseDuration(conf.WorkerFetchInterval)
	if err != nil {
		return err
	}

	fetcher := setupFetcher(storage, registryClient, d)

	return fetcher.Fetch(conf.Images)
}
