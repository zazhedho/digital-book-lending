package app

import (
	"digital-book-lending/controller"
	"digital-book-lending/middleware"
	"digital-book-lending/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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

	apiV1 := r.App.Group("/api/v1")
	{
		user := apiV1.Group("/user")
		{
			user.POST("/register", ctrlUser.Register)
			user.POST("/login", ctrlUser.Login)
		}
	}
}
