package repository

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"

	"gorm.io/gorm"
)

func NewUserRepo(db *gorm.DB) interfaces.Users {
	return &repo{DB: db}
}

type repo struct {
	DB *gorm.DB
}

func (r *repo) Store(m models.Users) error {
	if err := r.DB.Create(&m).Error; err != nil {
		utils.WriteLog(utils.LogLevelError, "sqlUsers.Store; "+err.Error())
		return err
	}

	return nil
}

func (r *repo) GetByEmail(email string) (ret models.Users, err error) {
	if err = r.DB.Where("email = ?", email).First(&ret).Error; err != nil {
		utils.WriteLog(utils.LogLevelError, "sqlUsers.GetByEmail; "+err.Error())
		return models.Users{}, err
	}

	return ret, nil
}
