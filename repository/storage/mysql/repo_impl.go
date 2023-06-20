package mysql

import (
	"github.com/xxxibgdrgnmm/reverse-registry/model"
	repository "github.com/xxxibgdrgnmm/reverse-registry/repository/storage"
	"gorm.io/gorm"
)

type MySQLStorage struct {
	db *gorm.DB
}

func NewMySQLStorage(db *gorm.DB) repository.Interface {
	return &MySQLStorage{
		db,
	}
}

func (s *MySQLStorage) FindByNameTag(nameWithTag string) (*model.ImageModel, error) {
	var iM model.ImageModel
	query := s.db.Model(&model.ImageModel{})
	query = query.Where("name=?", nameWithTag)
	err := query.Find(&iM).Error
	if err != nil {
		return nil, err
	}
	return &iM, nil
}

func (s *MySQLStorage) FindByDigest(hashedIndex string) (*model.ImageModel, error) {
	var iM model.ImageModel
	query := s.db.Model(&model.ImageModel{})
	query = query.Where("hashed_index=?", hashedIndex)
	err := query.Find(&iM).Error
	if err != nil {
		return nil, err
	}
	return &iM, nil
}

func (s *MySQLStorage) SaveDigest(nameWithTag string, hashedIndex string) error {
	var iM model.ImageModel
	iM.Name = nameWithTag
	iM.HashedIndex = hashedIndex
	if err := s.db.Save(&iM).Error; err != nil {
		return err
	}
	return nil
}
