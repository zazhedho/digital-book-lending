package controller

import (
	"digital-book-lending/services"
	"digital-book-lending/utils"
	"digital-book-lending/utils/functions"
	"digital-book-lending/utils/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LendingCtrl struct {
	lendingService *services.LendingService
}

func NewLendingController(lendingService *services.LendingService) *LendingCtrl {
	return &LendingCtrl{lendingService: lendingService}
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
	)

	authData := functions.GetAuthData(ctx)
	userId := utils.InterfaceString(authData["user_id"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][LendingBook][Borrow]", logId)

	bookId, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}

	newLendingRecord, err := c.lendingService.BorrowBook(bookId, userId)
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
	)

	authData := functions.GetAuthData(ctx)
	userId := utils.InterfaceString(authData["user_id"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][LendingBook][ReturnBook]", logId)

	lendingId, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}

	if err := c.lendingService.ReturnBook(lendingId, userId); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Error: %+v", logPrefix, err.Error()))
		res := response.Response(http.StatusUnprocessableEntity, utils.MsgFail, logId, nil)
		res.Errors = response.Errors{Code: http.StatusUnprocessableEntity, Message: err.Error()}
		ctx.JSON(http.StatusUnprocessableEntity, res)
		return
	}

	res := response.Response(http.StatusOK, "Book returned successfully", logId, nil)
	ctx.JSON(http.StatusOK, res)
}
