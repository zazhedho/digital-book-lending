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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BookCtrl struct {
	DBBookLending *gorm.DB
	RdbCache      *redis.Client
}

func NewBookController(dbBookLend *gorm.DB, rdbCache *redis.Client) *BookCtrl {
	return &BookCtrl{
		DBBookLending: dbBookLend,
		RdbCache:      rdbCache,
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
	if rows, err = bookRepo.Update(models.Book{ID: id}, book); err != nil {
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
