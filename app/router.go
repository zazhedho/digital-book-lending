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
	"gorm.io/gorm"
)

type Routes struct {
	App           *gin.Engine
	DBBookLending *gorm.DB
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
	ctrlUser := controller.NewUserController(r.DBBookLending)
	ctrlBook := controller.NewBookController(r.DBBookLending)

	apiV1 := r.App.Group("/api/v1")
	{
		// user route
		user := apiV1.Group("/user")
		{
			user.POST("/register", ctrlUser.Register)
			user.POST("/login", ctrlUser.Login)
		}

		// book route
		book := apiV1.Group("/books")
		{
			book.GET("", ctrlBook.List)
			adminBook := book.Group("").Use(r.AuthMiddleware(), r.RoleMiddleware(utils.RoleAdmin))
			{
				adminBook.POST("", ctrlBook.Create)
				adminBook.PUT("/update/:id", ctrlBook.Update)
				adminBook.DELETE("delete/:id", ctrlBook.Delete)
			}
		}
	}
}

func (r *Routes) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err       error
			logId     uuid.UUID
			logPrefix string
		)

		logId = utils.GenerateLogId(ctx)
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

func (r *Routes) RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			logId     uuid.UUID
			logPrefix string
		)

		logId = utils.GenerateLogId(ctx)
		logPrefix = fmt.Sprintf("[%s][RoleMiddleware]", logId)

		authData, exists := ctx.Get(utils.CtxKeyAuthData)
		if !exists {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; AuthData not found", logPrefix))
			res := response.Response(http.StatusForbidden, utils.MsgFail, logId, nil)
			res.Error = "auth data not found"
			ctx.AbortWithStatusJSON(http.StatusForbidden, res)
			return
		}
		dataJWT := authData.(map[string]interface{})

		userRole, ok := dataJWT["role"].(string)
		if !ok {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; there is no role user", logPrefix))
			res := response.Response(http.StatusForbidden, utils.MsgFail, logId, nil)
			res.Error = "there is no role user"
			ctx.AbortWithStatusJSON(http.StatusForbidden, res)
			return
		}

		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; User with role '%s' tried to access a restricted route;", logPrefix, userRole))
			res := response.Response(http.StatusForbidden, utils.MsgFail, logId, nil)
			res.Errors = response.Errors{Code: http.StatusForbidden, Message: utils.AccessDenied}
			ctx.AbortWithStatusJSON(http.StatusForbidden, res)
			return
		}

		ctx.Next()
	}
}
