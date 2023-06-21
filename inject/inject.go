package inject

import (
	"sync"

	"github.com/xxxibgdrgnmm/reverse-registry/config"
	"github.com/xxxibgdrgnmm/reverse-registry/driver"
	"github.com/xxxibgdrgnmm/reverse-registry/repository"
	containerregistry "github.com/xxxibgdrgnmm/reverse-registry/services/container-registry"
)

var imageMySQLStorage *repository.Storage
var muImageMySQLStorage sync.Mutex

func GetStorage(conf config.Config) (repository.Interface, error) {
	muImageMySQLStorage.Lock()
	defer muImageMySQLStorage.Unlock()
	if imageMySQLStorage != nil {
		return imageMySQLStorage, nil
	}
	dbConfig := conf.DBConfig
	host := dbConfig.Host
	user := dbConfig.User
	password := dbConfig.Password
	dbName := dbConfig.DBName
	db, err := driver.NewMySQLDB(host, user, password, dbName)
	if err != nil {
		return nil, err
	}
	imageMySQL := repository.NewStorage(db)
	return imageMySQL, nil
}

var registryClient *containerregistry.Client
var muRegistryClient sync.Mutex

func GetContainerRegistryClient() (containerregistry.Interface, error) {
	muRegistryClient.Lock()
	defer muRegistryClient.Unlock()
	if registryClient != nil {
		return registryClient, nil
	}
	c := containerregistry.New()
	return c, nil
}
