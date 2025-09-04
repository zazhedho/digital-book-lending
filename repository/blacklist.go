package repository

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"

	"gorm.io/gorm"
)

type blacklistRepo struct {
	DB *gorm.DB
}

func NewBlacklistRepo(db *gorm.DB) interfaces.Blacklist {
	return &blacklistRepo{
		DB: db,
	}
}

func (r *blacklistRepo) Store(blacklist models.Blacklist) error {
	return r.DB.Create(&blacklist).Error
}

func (r *blacklistRepo) GetByToken(token string) (models.Blacklist, error) {
	var blacklist models.Blacklist
	err := r.DB.Where("token = ?", token).First(&blacklist).Error
	return blacklist, err
}
