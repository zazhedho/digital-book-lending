package repository

import (
	"digital-book-lending/models"

	"gorm.io/gorm"
)

type BlacklistRepo struct {
	DB *gorm.DB
}

func NewBlacklistRepo(db *gorm.DB) *BlacklistRepo {
	return &BlacklistRepo{
		DB: db,
	}
}

func (r *BlacklistRepo) Store(blacklist models.Blacklist) error {
	return r.DB.Create(&blacklist).Error
}

func (r *BlacklistRepo) GetByToken(token string) (models.Blacklist, error) {
	var blacklist models.Blacklist
	err := r.DB.Where("token = ?", token).First(&blacklist).Error
	return blacklist, err
}
