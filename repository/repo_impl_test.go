package repository

import (
	"testing"

	"github.com/nduyphuong/reverse-registry/driver"
	"github.com/test-go/testify/assert"
)

func TestRepoImpl(t *testing.T) {
	db, err := driver.NewMySQLDB("localhost", "root", "my-secret-pw", "test")
	assert.NoError(t, err)
	imageModelStorage := NewStorage(db)
	err = imageModelStorage.SaveDigest("172.20.10.2:8080/nginx:1.25.1-r0", "sha256:81bed54c9e507503766c0f8f030f869705dae486f37c2a003bb5b12bcfcc713f")
	assert.NoError(t, err)
	res, err := imageModelStorage.FindByNameTag("172.20.10.2:8080/nginx:1.25.1-r0")
	assert.NoError(t, err)
	assert.ObjectsAreEqual(map[string]string{"172.20.10.2:8080/nginx:1.25.1-r0": "sha256:81bed54c9e507503766c0f8f030f869705dae486f37c2a003bb5b12bcfcc713f"}, *res)
	res, err = imageModelStorage.FindByDigest("sha256:81bed54c9e507503766c0f8f030f869705dae486f37c2a003bb5b12bcfcc713f")
	assert.NoError(t, err)
	assert.ObjectsAreEqual(map[string]string{"172.20.10.2:8080/nginx:1.25.1-r0": "sha256:81bed54c9e507503766c0f8f030f869705dae486f37c2a003bb5b12bcfcc713f"}, *res)
}
