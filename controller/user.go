package controller

import (
	"digital-book-lending/models"
	"digital-book-lending/repository"
	"digital-book-lending/utils"
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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserCtrl struct {
	DBBookLending *gorm.DB
	RdbCache      *redis.Client
	RdbTemp       *redis.Client
}

func NewUserController(dbBookLend *gorm.DB, rdbCache, rdbTemp *redis.Client) *UserCtrl {
	return &UserCtrl{
		DBBookLending: dbBookLend,
		RdbCache:      rdbCache,
		RdbTemp:       rdbTemp,
	}
}

func (cc *UserCtrl) Register(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.Register
		err       error

		user models.Users
	)
	userRepo := repository.NewUserRepo(cc.DBBookLending)

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][UserController][Register]", logId)

	if err = ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bcrypt.GenerateFromPassword; Error: %+v", logPrefix, err))

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	user = models.Users{
		Id:        utils.CreateUUID(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPwd),
		CreatedAt: time.Now(),
	}

	if err = userRepo.Store(user); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; userRepo.Store; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
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
	return
}

func (cc *UserCtrl) Login(ctx *gin.Context) {
	var (
		logId     uuid.UUID
		logPrefix string
		req       request.Login
		err       error

		user models.Users
	)
	userRepo := repository.NewUserRepo(cc.DBBookLending)

	logId = utils.GenerateLogId(ctx)
	logPrefix = fmt.Sprintf("[%s][UserController][Login]", logId)

	if err = ctx.BindJSON(&req); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))

		res := response.Response(http.StatusBadRequest, utils.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	user, err = userRepo.GetByEmail(req.Email)
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; userRepo.GetByEmail; ERROR: %s;", logPrefix, err))
		if errors.Is(err, gorm.ErrRecordNotFound) || reflect.DeepEqual(user, models.Users{}) {
			res := response.Response(http.StatusBadRequest, utils.InvalidCred, logId, nil)
			res.Errors = response.Errors{Code: http.StatusBadRequest, Message: utils.MsgCredential}
			ctx.JSON(http.StatusBadRequest, res)
			return
		}

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; bcrypt.CompareHashAndPassword; ERROR: %s;", logPrefix, err))

		res := response.Response(http.StatusBadRequest, utils.InvalidCred, logId, nil)
		res.Errors = response.Errors{Code: http.StatusBadRequest, Message: utils.MsgCredential}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	token, err := utils.GenerateJwt(user.Id, logId.String())
	if err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; GenerateJwt; ERROR: %s;", logPrefix, err))

		res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
		res.Error = err.Error()
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := response.Response(http.StatusOK, "success", logId, fmt.Sprintf("token: %s", token))
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("%s; Success: %+v;", logPrefix, utils.JsonEncode(token)))
	ctx.JSON(http.StatusOK, res)
	return
}
