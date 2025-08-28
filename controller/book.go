package controller

import (
	"digital-book-lending/models"
	"digital-book-lending/repository"
	"digital-book-lending/utils"
	"digital-book-lending/utils/functions"
	"digital-book-lending/utils/request"
	"digital-book-lending/utils/response"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookCtrl struct {
	DBBookLending *gorm.DB
}

func NewBookController(dbBookLend *gorm.DB) *BookCtrl {
	return &BookCtrl{
		DBBookLending: dbBookLend,
	}
}

func getAuthData(ctx *gin.Context) map[string]interface{} {
	jwtClaims, _ := ctx.Get(utils.CtxKeyAuthData)
	if jwtClaims != nil {
		return jwtClaims.(map[string]interface{})
	}
	return nil
}

func (c *BookCtrl) Create(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.AddBook
		err       error

		book models.Book
	)
	bookRepo := repository.NewBookRepo(c.DBBookLending)
	authData := getAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][Create]", logId)

	if err = ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	book = models.Book{
		ID:        utils.CreateUUID(),
		Title:     req.Title,
		Author:    req.Author,
		ISBN:      req.ISBN,
		Category:  req.Category,
		Quantity:  req.Quantity,
		CreatedAt: time.Now(),
		CreatedBy: username,
	}

	if err = bookRepo.Store(book); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookRepo.Store; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			res := response.Response(http.StatusConflict, utils.MsgExists, logId, nil)
			res.Errors = response.Errors{Code: http.StatusConflict, Message: fmt.Sprintf("book with ISBN: %s already exists", req.ISBN)}
			ctx.JSON(http.StatusConflict, res)
			return
		}

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusCreated, "Add book successfully", logId, book)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Success: %+v;", logPrefix, utils.JsonEncode(book)))
	ctx.JSON(http.StatusCreated, res)
	return
}

func (c *BookCtrl) Update(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.UpdateBook
		err       error
		rows      int64

		book models.Book
	)
	bookRepo := repository.NewBookRepo(c.DBBookLending)
	authData := getAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][Update]", logId)

	if err = ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	id, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}
	logPrefix += fmt.Sprintf("[%s][%s]", id, username)

	timeNow := time.Now()
	book.Title = req.Title
	book.Author = req.Author
	book.ISBN = req.ISBN
	book.Category = req.Category
	book.Quantity = req.Quantity
	book.UpdatedAt = &timeNow
	book.UpdatedBy = username
	if rows, err = bookRepo.Update(c.DBBookLending, models.Book{ID: id}, book); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookRepo.Update; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; ISBN: '%s' already exists", logPrefix, book.ISBN))
			res := response.Response(http.StatusBadRequest, utils.MsgExists, logId, nil)
			res.Errors = response.Errors{Code: http.StatusBadRequest, Message: fmt.Sprintf("ISBN: '%s' is already exists", book.ISBN)}
			ctx.JSON(http.StatusBadRequest, res)
			return
		}

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}
	if rows == 0 {
		res := response.Response(http.StatusNotFound, utils.MsgNotFound, logId, nil)
		res.Errors = response.Errors{Code: http.StatusNotFound, Message: utils.NotFound}
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := response.Response(http.StatusOK, fmt.Sprintf("Book with ID: '%s' updated successfully", id), logId, nil)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Book with ID: '%s' updated successfully; Data: %v", logPrefix, id, utils.JsonEncode(book)))
	ctx.JSON(http.StatusOK, res)
	return
}

func (c *BookCtrl) Delete(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		err       error
	)
	bookRepo := repository.NewBookRepo(c.DBBookLending)
	authData := getAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][Delete]", logId)

	id, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}
	logPrefix += fmt.Sprintf("[%s][%s]", id, username)

	// hard delete
	//if _, err = bookRepo.Delete(models.Book{ID: id}); err != nil {
	//	utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookRepo.Delete; Error: %+v", logPrefix, err))
	//
	//	res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
	//	res.Error = err.Error()
	//	ctx.JSON(http.StatusInternalServerError, res)
	//	return
	//}

	// soft delete
	if _, err = bookRepo.SoftDelete(models.Book{ID: id}, map[string]interface{}{"deleted_at": time.Now(), "deleted_by": username}); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookRepo.Delete; Error: %+v", logPrefix, err))

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusOK, fmt.Sprintf("Book with ID: '%s' deleted successfully", id), logId, nil)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Book with ID: '%s' deleted successfully", logPrefix, id))
	ctx.JSON(http.StatusOK, res)
	return
}

func (c *BookCtrl) List(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
	)
	bookRepo := repository.NewBookRepo(c.DBBookLending)
	authData := getAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][List][%s]", logId, username)

	//query parameters
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	orderBy := ctx.DefaultQuery("order_by", "updated_at")
	orderDir := ctx.DefaultQuery("order_direction", "desc")
	search := ctx.Query("search")

	books, totalData, err := bookRepo.Fetch(page, limit, orderBy, orderDir, search)
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Fetch; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := response.Response(http.StatusNotFound, utils.NotFound, logId, nil)
			res.Error = "Book data not found"
			ctx.JSON(http.StatusNotFound, res)
			return
		}

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.PaginationResponse(http.StatusOK, int(totalData), page, limit, logId, books)
	utils.WriteLog(utils.LogLevelInfo, fmt.Sprintf("%s; Success; List: %v", logPrefix, utils.JsonEncode(books)))
	ctx.JSON(http.StatusOK, res)
	return
}
