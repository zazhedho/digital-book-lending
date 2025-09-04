package services

import (
	"digital-book-lending/models"
	"digital-book-lending/repository"
	"digital-book-lending/utils"
	"digital-book-lending/utils/request"
	"time"

	"gorm.io/gorm"
)

type BookService struct {
	DB *gorm.DB
}

func NewBookService(db *gorm.DB) *BookService {
	return &BookService{
		DB: db,
	}
}

func (s *BookService) CreateBook(req request.AddBook, username string) (models.Book, error) {
	bookRepo := repository.NewBookRepo(s.DB)

	book := models.Book{
		ID:        utils.CreateUUID(),
		Title:     req.Title,
		Author:    req.Author,
		ISBN:      req.ISBN,
		Category:  req.Category,
		Quantity:  req.Quantity,
		CreatedAt: time.Now(),
		CreatedBy: username,
	}

	if err := bookRepo.Store(book); err != nil {
		return models.Book{}, err
	}

	return book, nil
}

func (s *BookService) UpdateBook(id string, req request.UpdateBook, username string) (int64, error) {
	bookRepo := repository.NewBookRepo(s.DB)
	timeNow := time.Now()

	book := models.Book{
		Title:     req.Title,
		Author:    req.Author,
		ISBN:      req.ISBN,
		Category:  req.Category,
		Quantity:  req.Quantity,
		UpdatedAt: &timeNow,
		UpdatedBy: username,
	}

	rows, err := bookRepo.Update(s.DB, models.Book{ID: id}, book)
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func (s *BookService) DeleteBook(id string, username string) error {
	bookRepo := repository.NewBookRepo(s.DB)

	if _, err := bookRepo.SoftDelete(models.Book{ID: id}, map[string]interface{}{"deleted_at": time.Now(), "deleted_by": username}); err != nil {
		return err
	}

	return nil
}

func (s *BookService) ListBooks(page, limit int, orderBy, orderDir, search string) ([]models.Book, int64, error) {
	bookRepo := repository.NewBookRepo(s.DB)
	return bookRepo.Fetch(page, limit, orderBy, orderDir, search)
}
