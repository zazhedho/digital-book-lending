package interfaces

import (
	"digital-book-lending/models"

	"gorm.io/gorm"
)

type Lending interface {
	Store(tx *gorm.DB, m models.LendingRecord) (models.LendingRecord, error)
	GetActiveByUserAndBook(tx *gorm.DB, userId, bookId string) (models.LendingRecord, error)
}
