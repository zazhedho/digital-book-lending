package controller

import (
	"digital-book-lending/services"
	"digital-book-lending/utils"
	"digital-book-lending/utils/functions"
	"digital-book-lending/utils/request"
	"digital-book-lending/utils/response"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookCtrl struct {
	bookService *services.BookService
}

func NewBookController(bookService *services.BookService) *BookCtrl {
	return &BookCtrl{
		bookService: bookService,
	}
}

// Create godoc
// @Summary Create a new book
// @Description Create a new book
// @Tags books
// @Accept  json
// @Produce  json
// @Param book body request.AddBook true "Book details"
// @Success 201 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Security ApiKeyAuth
// @Router /books [post]
func (c *BookCtrl) Create(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.AddBook
	)
	authData := functions.GetAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][Create]", logId)

	if err := ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	book, err := c.bookService.CreateBook(req, username)
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookService.CreateBook; Error: %+v", logPrefix, err))
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
}

// Update godoc
// @Summary Update a book
// @Description Update a book
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path string true "Book ID"
// @Param book body request.UpdateBook true "Book details"
// @Success 200 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Security ApiKeyAuth
// @Router /books/update/{id} [put]
func (c *BookCtrl) Update(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.UpdateBook
	)
	authData := functions.GetAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][Update]", logId)

	if err := ctx.BindJSON(&req); err != nil {
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

	rows, err := c.bookService.UpdateBook(id, req, username)
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookService.UpdateBook; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; ISBN: '%s' already exists", logPrefix, req.ISBN))
			res := response.Response(http.StatusBadRequest, utils.MsgExists, logId, nil)
			res.Errors = response.Errors{Code: http.StatusBadRequest, Message: fmt.Sprintf("ISBN: '%s' is already exists", req.ISBN)}
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
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Book with ID: '%s' updated successfully; Data: %v", logPrefix, id, utils.JsonEncode(req)))
	ctx.JSON(http.StatusOK, res)
}

// Delete godoc
// @Summary Delete a book
// @Description Delete a book
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path string true "Book ID"
// @Success 200 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Security ApiKeyAuth
// @Router /books/delete/{id} [delete]
func (c *BookCtrl) Delete(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
	)
	authData := functions.GetAuthData(ctx)
	username := utils.InterfaceString(authData["username"])

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][Book][Delete]", logId)

	id, err := functions.ValidateUUID(ctx, logPrefix, logId)
	if err != nil {
		return
	}
	logPrefix += fmt.Sprintf("[%s][%s]", id, username)

	if err := c.bookService.DeleteBook(id, username); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bookService.DeleteBook; Error: %+v", logPrefix, err))

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusOK, fmt.Sprintf("Book with ID: '%s' deleted successfully", id), logId, nil)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Book with ID: '%s' deleted successfully", logPrefix, id))
	ctx.JSON(http.StatusOK, res)
}

// List godoc
// @Summary List books
// @Description List books
// @Tags books
// @Accept  json
// @Produce  json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param order_by query string false "Order by field"
// @Param order_direction query string false "Order direction (asc/desc)"
// @Param search query string false "Search query"
// @Success 200 {object} response.Pagination
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /books [get]
func (c *BookCtrl) List(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
	)
	authData := functions.GetAuthData(ctx)
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

	books, totalData, err := c.bookService.ListBooks(page, limit, orderBy, orderDir, search)
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
}
