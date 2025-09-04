package controller

import (
	"digital-book-lending/models"
	"digital-book-lending/services"
	"digital-book-lending/utils"
	"digital-book-lending/utils/request"
	"digital-book-lending/utils/response"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserCtrl struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserCtrl {
	return &UserCtrl{
		DB: db,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body request.Register true "User registration details"
// @Success 201 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /user/register [post]
func (cc *UserCtrl) Register(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.Register
	)
	userService := services.NewUserService(cc.DB)

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][UserController][Register]", logId)

	if err := ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	user, err := userService.RegisterUser(req)
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; userService.RegisterUser; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Error: email already exists", logPrefix))
			res := response.Response(http.StatusBadRequest, utils.MsgExists, logId, nil)
			res.Errors = response.Errors{Code: http.StatusBadRequest, Message: "email already exists"}
			ctx.JSON(http.StatusBadRequest, res)
			return
		}

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusCreated, "User registered successfully", logId, user)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Success: %+v;", logPrefix, utils.JsonEncode(user)))
	ctx.JSON(http.StatusCreated, res)
}

// Login godoc
// @Summary Login a user
// @Description Login a user
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body request.Login true "User login details"
// @Success 200 {object} response.Success
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /user/login [post]
func (cc *UserCtrl) Login(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.Login
	)
	userService := services.NewUserService(cc.DB)

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][UserController][Login]", logId)

	if err := ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	token, err := userService.LoginUser(req, logId.String())
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; userService.LoginUser; ERROR: %s;", logPrefix, err))
		if errors.Is(err, gorm.ErrRecordNotFound) || reflect.DeepEqual(models.Users{}, models.Users{}) {
			res := response.Response(http.StatusBadRequest, utils.InvalidCred, logId, nil)
			res.Errors = response.Errors{Code: http.StatusBadRequest, Message: utils.MsgCredential}
			ctx.JSON(http.StatusBadRequest, res)
			return
		}

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusOK, "success", logId, fmt.Sprintf("token: %s", token))
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Success: %+v;", logPrefix, utils.JsonEncode(token)))
	ctx.JSON(http.StatusOK, res)
}

// Logout godoc
// @Summary Logout a user
// @Description Logout a user
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Success
// @Failure 500 {object} response.Error
// @Security ApiKeyAuth
// @Router /user/logout [post]
func (cc *UserCtrl) Logout(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
	)
	userService := services.NewUserService(cc.DB)

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][UserController][Logout]", logId)

	token, ok := ctx.Get("token")
	if !ok {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; token not found in context", logPrefix))
		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = "token not found"
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	if err := userService.LogoutUser(token.(string)); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; userService.LogoutUser; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusOK, "User logged out successfully", logId, nil)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Success: User logged out successfully", logPrefix))
	ctx.JSON(http.StatusOK, res)
}
