package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
)

//RepoDAOInterface interface
type RepoDAOInterface interface {
	CreateOrUpdate(repo model.Repository) error
	All() ([]model.Repository, error)
	Delete(repo string) error
}

//RepoDAOImpl struct
type RepoDAOImpl struct {
	Db *gorm.DB
}

//CreateOrUpdate func
func (dao RepoDAOImpl) CreateOrUpdate(repo model.Repository) error {
	if dao.Db.Model(&repo).Where("ID = ?", repo.ID).Updates(&repo).RowsAffected == 0 {
		return dao.Db.Create(&repo).Error
	}
	return nil
}

//All func
func (dao RepoDAOImpl) All() ([]model.Repository, error) {
	list := []model.Repository{}
	return list, dao.Db.Find(&list).Error
}

//Delete func
func (dao RepoDAOImpl) Delete(repo string) error {
	return dao.Db.Where("name = ?", repo).Delete(&model.Repository{}).Error
}
