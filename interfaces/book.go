package interfaces

import "digital-book-lending/models"

type Book interface {
	Store(m models.Book) error
	Update(m models.Book, data interface{}) (int64, error)
	Delete(m models.Book) (int64, error)
	GetByIsbn(isbn string) (models.Book, error)
}
