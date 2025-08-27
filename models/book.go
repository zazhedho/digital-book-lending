package models

import "time"

func (Book) TableName() string {
	return "books"
}

type Book struct {
	ID        string     `json:"id" gorm:"column:id"`
	Title     string     `json:"title" gorm:"column:title"`
	Author    string     `json:"author" gorm:"column:author"`
	ISBN      string     `json:"isbn" gorm:"column:isbn"`
	Category  string     `json:"category" gorm:"column:category"`
	Quantity  int        `json:"quantity" gorm:"column:quantity"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	CreatedBy string     `json:"created_by" gorm:"column:created_by"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy string     `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
	DeletedBy string     `json:"-" gorm:"column:deleted_by"`
}
