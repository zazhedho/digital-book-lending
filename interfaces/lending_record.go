package interfaces

import (
	"digital-book-lending/models"
	"time"

	"gorm.io/gorm"
)

type Lending interface {
	Store(tx *gorm.DB, m models.LendingRecord) (models.LendingRecord, error)
	GetActiveByUserAndBook(tx *gorm.DB, userId, bookId string) (models.LendingRecord, error)
	CountBorrowsByUser(tx *gorm.DB, userId string, since time.Time) (int64, error)
}
