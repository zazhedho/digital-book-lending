package repository

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"
	"time"

	"gorm.io/gorm"
)

type repoLending struct {
	DB *gorm.DB
}

func NewLendingRepo(db *gorm.DB) interfaces.Lending {
	return &repoLending{DB: db}
}

func (r *repoLending) Store(tx *gorm.DB, m models.LendingRecord) (models.LendingRecord, error) {
	if err := tx.Create(&m).Error; err != nil {
		return m, err
	}

	return m, nil
}

func (r *repoLending) GetActiveByUserAndBook(tx *gorm.DB, userId, bookId string) (models.LendingRecord, error) {
	var m models.LendingRecord
	err := tx.Where("user_id = ? AND book_id = ? AND status = ?", userId, bookId, utils.Borrowed).
		First(&m).Error

	return m, err
}

func (r *repoLending) CountBorrowsByUser(tx *gorm.DB, userId string, since time.Time) (int64, error) {
	var count int64
	err := tx.Model(&models.LendingRecord{}).
		Where("user_id = ? AND borrow_date >= ?", userId, since).
		Count(&count).Error
	return count, err
}
