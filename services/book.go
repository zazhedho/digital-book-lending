package services

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"
	"digital-book-lending/utils/request"
	"time"

	"gorm.io/gorm"
)

type BookService struct {
	bookRepo interfaces.Book
	DB       *gorm.DB
}

func NewBookService(bookRepo interfaces.Book, db *gorm.DB) *BookService {
	return &BookService{
		bookRepo: bookRepo,
		DB:       db,
	}
}

func (s *BookService) CreateBook(req request.AddBook, username string) (models.Book, error) {
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

	if err := s.bookRepo.Store(book); err != nil {
		return models.Book{}, err
	}

	return book, nil
}

func (s *BookService) UpdateBook(id string, req request.UpdateBook, username string) (int64, error) {
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

	rows, err := s.bookRepo.Update(s.DB, models.Book{ID: id}, book)
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func (s *BookService) DeleteBook(id string, username string) error {
	if _, err := s.bookRepo.SoftDelete(models.Book{ID: id}, map[string]interface{}{"deleted_at": time.Now(), "deleted_by": username}); err != nil {
		return err
	}

	return nil
}

func (s *BookService) ListBooks(page, limit int, orderBy, orderDir, search string) ([]models.Book, int64, error) {
	return s.bookRepo.Fetch(page, limit, orderBy, orderDir, search)
}
