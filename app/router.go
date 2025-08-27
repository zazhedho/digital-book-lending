package app

import (
	"digital-book-lending/controller"
	"digital-book-lending/middleware"
	"digital-book-lending/utils"
	"digital-book-lending/utils/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Routes struct {
	App           *gin.Engine
	DBBookLending *gorm.DB
	RdbCache      *redis.Client
	RdbTemp       *redis.Client
}

func NewRoutes() *Routes {
	app := gin.Default()

	app.Use(middleware.CORS())
	app.Use(gin.CustomRecovery(middleware.ErrorHandler))
	app.Use(middleware.SetContextId())

	// health check
	app.GET("/healthcheck", func(ctx *gin.Context) {
		utils.WriteLog(utils.LogLevelDebug, "ClientIP: "+ctx.ClientIP())
		ctx.JSON(http.StatusOK, gin.H{
			"message": "OK!!",
		})
	})

	return &Routes{
		App: app,
	}
}

func (r *Routes) BookLending() {
	ctrlUser := controller.NewUserController(r.DBBookLending, r.RdbCache, r.RdbTemp)
	ctrlBook := controller.NewBookController(r.DBBookLending, r.RdbCache)

	apiV1 := r.App.Group("/api/v1")
	{
		user := apiV1.Group("/user")
		{
			user.POST("/register", ctrlUser.Register)
			user.POST("/login", ctrlUser.Login)
		}

		book := apiV1.Group("/books").Use(r.AuthMiddleware())
		{
			book.POST("", ctrlBook.Create)
			book.POST("/update/:id", ctrlBook.Update)
		}
	}
}

func (r *Routes) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			ok        bool
			err       error
			logId     uuid.UUID
			logPrefix string
		)

		if logId, ok = ctx.Value(utils.CtxKeyId).(uuid.UUID); !ok {
			if logId, err = uuid.NewV7(); err != nil {
				logId = uuid.New()
			}
		}
		logPrefix = fmt.Sprintf("[%s][AuthMiddleware]", logId)

		tokenString, dataJWT, err := utils.JwtClaims(ctx)
		if err != nil {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Invalid Token: %s; Error: %s;", logPrefix, tokenString, err.Error()))
			res := response.Response(http.StatusUnauthorized, utils.MsgFail, logId, nil)
			res.Error = err.Error()
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
			return
		}
		logPrefix += fmt.Sprintf("[%s][%s]", dataJWT["id"], utils.InterfaceString(dataJWT["user_id"]))

		ctx.Set(utils.CtxKeyAuthData, dataJWT)
		ctx.Set("token", tokenString)

		ctx.Next()
	}
}
