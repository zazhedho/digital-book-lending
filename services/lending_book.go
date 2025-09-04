package services

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"
	"errors"
	"time"

	"gorm.io/gorm"
)

type LendingService struct {
	lendingRepo interfaces.Lending
	bookRepo    interfaces.Book
	DB          *gorm.DB
}

func NewLendingService(lendingRepo interfaces.Lending, bookRepo interfaces.Book, db *gorm.DB) *LendingService {
	return &LendingService{
		lendingRepo: lendingRepo,
		bookRepo:    bookRepo,
		DB:          db,
	}
}

func (s *LendingService) BorrowBook(bookId, userId string) (models.LendingRecord, error) {
	var newLendingRecord models.LendingRecord

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		book, err := s.bookRepo.GetByIdForUpdate(tx, bookId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("book not found")
			}
			return err
		}

		if book.Quantity < 1 {
			return errors.New("book is out of stock")
		}

		_, err = s.lendingRepo.GetActiveByUserAndBook(tx, userId, bookId)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("you have already borrowed this book")
		}

		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		recentBorrows, err := s.lendingRepo.CountBorrowsByUser(tx, userId, sevenDaysAgo)
		if err != nil {
			return err
		}
		if recentBorrows >= 5 {
			return errors.New("borrowing limit exceeded: you have borrowed 5 books in the last 7 days")
		}

		bookDataUpdate := map[string]interface{}{"quantity": book.Quantity - 1}
		if _, err := s.bookRepo.Update(tx, book, bookDataUpdate); err != nil {
			return err
		}

		record := models.LendingRecord{
			Id:         utils.CreateUUID(),
			UserId:     userId,
			BookId:     bookId,
			BorrowDate: time.Now(),
			Status:     utils.Borrowed,
		}

		newLendingRecord, err = s.lendingRepo.Store(tx, record)
		if err != nil {
			return err
		}

		return nil
	})

	return newLendingRecord, err
}

func (s *LendingService) ReturnBook(lendingId, userId string) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		record, err := s.lendingRepo.GetBorrowedById(tx, lendingId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("active lending record not found or already returned")
			}
			return err
		}
		if record.UserId != userId {
			return errors.New("you are not authorized to return this book")
		}

		book, err := s.bookRepo.GetByIdForUpdate(tx, record.BookId)
		if err != nil {
			return err
		}
		bookDataUpdate := map[string]interface{}{"quantity": book.Quantity + 1}
		if _, err := s.bookRepo.Update(tx, book, bookDataUpdate); err != nil {
			return err
		}

		lendingDataUpdate := map[string]interface{}{
			"status":      utils.Returned,
			"return_date": time.Now(),
		}
		if err := s.lendingRepo.Update(tx, record, lendingDataUpdate); err != nil {
			return err
		}

		return nil
	})

	return err
}
