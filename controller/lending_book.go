package controller

import (
	"digital-book-lending/models"
	"digital-book-lending/repository"
	"digital-book-lending/utils"
	"digital-book-lending/utils/functions"
	"digital-book-lending/utils/response"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LendingCtrl struct {
	DBBookLending *gorm.DB
}

func NewLendingController(dbBookLend *gorm.DB) *LendingCtrl {
	return &LendingCtrl{DBBookLending: dbBookLend}
}

// BorrowBook godoc
// @Summary Borrow a book
// @Description Borrow a book
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path string true "Book ID"
// @Success 201 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 422 {object} response.Error
// @Security ApiKeyAuth
// @Router /books/{id}/borrow [post]
func (c *LendingCtrl) BorrowBook(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		err       error

		newLendingRecord models.LendingRecord
	)
	bookRepo := repository.NewBookRepo(c.DBBookLending)
	lendingRepo := repository.NewLendingRepo(c.DBBookLending)

	authData := getAuthData(ctx)
	userId := utils.InterfaceString(authData["user_id"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][LendingBook][Borrow]", logId)

	bookId, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}

	err = c.DBBookLending.Transaction(func(tx *gorm.DB) error {
		book, err := bookRepo.GetByIdForUpdate(tx, bookId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("book not found")
			}
			return err
		}

		if book.Quantity < 1 {
			return errors.New("book is out of stock")
		}

		_, err = lendingRepo.GetActiveByUserAndBook(tx, userId, bookId)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("you have already borrowed this book")
		}

		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		recentBorrows, err := lendingRepo.CountBorrowsByUser(tx, userId, sevenDaysAgo)
		if err != nil {
			return err
		}
		if recentBorrows >= 5 {
			return errors.New("borrowing limit exceeded: you have borrowed 5 books in the last 7 days")
		}

		bookDataUpdate := map[string]interface{}{"quantity": book.Quantity - 1}
		if _, err := bookRepo.Update(tx, book, bookDataUpdate); err != nil {
			return err
		}

		record := models.LendingRecord{
			Id:         utils.CreateUUID(),
			UserId:     userId,
			BookId:     bookId,
			BorrowDate: time.Now(),
			Status:     utils.Borrowed,
		}

		newLendingRecord, err = lendingRepo.Store(tx, record)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Error: %+v", logPrefix, err.Error()))
		res := response.Response(http.StatusUnprocessableEntity, utils.MsgFail, logId, nil)
		res.Errors = response.Errors{Code: http.StatusUnprocessableEntity, Message: err.Error()}
		ctx.JSON(http.StatusUnprocessableEntity, res)
		return
	}

	res := response.Response(http.StatusCreated, "Book borrowed successfully", logId, newLendingRecord)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Success: %+v;", logPrefix, utils.JsonEncode(newLendingRecord)))
	ctx.JSON(http.StatusCreated, res)
	return
}

// ReturnBook godoc
// @Summary Return a book
// @Description Return a book
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path string true "Lending ID"
// @Success 200 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 422 {object} response.Error
// @Security ApiKeyAuth
// @Router /books/{id}/return [post]
func (c *LendingCtrl) ReturnBook(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		err       error
	)
	bookRepo := repository.NewBookRepo(c.DBBookLending)
	lendingRepo := repository.NewLendingRepo(c.DBBookLending)

	authData := getAuthData(ctx)
	userId := utils.InterfaceString(authData["user_id"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][LendingBook][ReturnBook]", logId)

	lendingId, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}

	err = c.DBBookLending.Transaction(func(tx *gorm.DB) error {
		record, err := lendingRepo.GetBorrowedById(tx, lendingId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("active lending record not found or already returned")
			}
			return err
		}
		if record.UserId != userId {
			return errors.New("you are not authorized to return this book")
		}

		book, err := bookRepo.GetByIdForUpdate(tx, record.BookId)
		if err != nil {
			return err
		}
		bookDataUpdate := map[string]interface{}{"quantity": book.Quantity + 1}
		if _, err := bookRepo.Update(tx, book, bookDataUpdate); err != nil {
			return err
		}

		lendingDataUpdate := map[string]interface{}{
			"status":      utils.Returned,
			"return_date": time.Now(),
		}
		if err := lendingRepo.Update(tx, record, lendingDataUpdate); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Error: %+v", logPrefix, err.Error()))
		res := response.Response(http.StatusUnprocessableEntity, utils.MsgFail, logId, nil)
		res.Errors = response.Errors{Code: http.StatusUnprocessableEntity, Message: err.Error()}
		ctx.JSON(http.StatusUnprocessableEntity, res)
		return
	}

	res := response.Response(http.StatusOK, "Book returned successfully", logId, nil)
	ctx.JSON(http.StatusOK, res)
	return
}
