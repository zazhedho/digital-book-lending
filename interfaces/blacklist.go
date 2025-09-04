package interfaces

import "digital-book-lending/models"

type Blacklist interface {
	Store(m models.Blacklist) error
	GetByToken(token string) (models.Blacklist, error)
}
