package interfaces

import "digital-book-lending/models"

type Users interface {
	Store(m models.Users) error
	GetByEmail(email string) (models.Users, error)
}
