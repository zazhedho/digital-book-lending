package models

import (
	"database/sql"
	"time"
)

func (LendingRecord) TableName() string {
	return "lending_records"
}

type LendingRecord struct {
	Id         string       `json:"id" gorm:"column:id;primaryKey"`
	UserId     string       `json:"user_id" gorm:"column:user_id"`
	BookId     string       `json:"book_id" gorm:"column:book_id"`
	BorrowDate time.Time    `json:"borrow_date" gorm:"column:borrow_date"`
	ReturnDate sql.NullTime `json:"return_date" gorm:"column:return_date"`
	Status     string       `json:"status" gorm:"column:status"`
	CreatedAt  time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time    `json:"updated_at" gorm:"column:updated_at"`
}
