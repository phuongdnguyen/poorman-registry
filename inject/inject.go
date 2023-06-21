package inject

import (
	"sync"

	"github.com/nduyphuong/reverse-registry/config"
	"github.com/nduyphuong/reverse-registry/driver"
	"github.com/nduyphuong/reverse-registry/repository"
	containerregistry "github.com/nduyphuong/reverse-registry/services/container-registry"
)

var imageStorage *repository.Storage
var muImageStorage sync.Mutex

func GetStorage(conf config.Config) (repository.Interface, error) {
	muImageStorage.Lock()
	defer muImageStorage.Unlock()
	if imageStorage != nil {
		return imageStorage, nil
	}
	dbConfig := conf.DBConfig
	host := dbConfig.Host
	user := dbConfig.User
	password := dbConfig.Password
	dbName := dbConfig.DBName
	if conf.DB == "mysql" {
		db, err := driver.NewMySQLDB(host, user, password, dbName)
		if err != nil {
			return nil, err
		}
		imageStorage := repository.NewStorage(db)
		return imageStorage, nil
	}
	db, err := driver.NewSqliteDB()
	if err != nil {
		return nil, err
	}
	imageStorage := repository.NewStorage(db)
	return imageStorage, nil
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
