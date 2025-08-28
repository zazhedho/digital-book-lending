package interfaces

import (
	"digital-book-lending/models"
	"time"

	"gorm.io/gorm"
)

type Lending interface {
	Store(tx *gorm.DB, m models.LendingRecord) (models.LendingRecord, error)
	Update(tx *gorm.DB, m models.LendingRecord, data interface{}) error
	GetActiveByUserAndBook(tx *gorm.DB, userId, bookId string) (models.LendingRecord, error)
	CountBorrowsByUser(tx *gorm.DB, userId string, since time.Time) (int64, error)
	GetBorrowedById(tx *gorm.DB, id string) (models.LendingRecord, error)
}
