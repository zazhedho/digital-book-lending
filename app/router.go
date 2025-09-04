package app

import (
	"digital-book-lending/controller"
	"digital-book-lending/interfaces"
	"digital-book-lending/middleware"
	"digital-book-lending/services"
	"digital-book-lending/utils"
	"digital-book-lending/utils/response"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type Routes struct {
	App            *gin.Engine
	BookService    *services.BookService
	UserService    *services.UserService
	LendingService *services.LendingService
	BlacklistRepo  interfaces.Blacklist
}

func NewRoutes(bookService *services.BookService, userService *services.UserService, lendingService *services.LendingService, blacklistRepo interfaces.Blacklist) *Routes {
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

	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &Routes{
		App:            app,
		BookService:    bookService,
		UserService:    userService,
		LendingService: lendingService,
		BlacklistRepo:  blacklistRepo,
	}
}

func (r *Routes) BookLending() {
	ctrlUser := controller.NewUserController(r.UserService)
	ctrlBook := controller.NewBookController(r.BookService)
	ctrlLending := controller.NewLendingController(r.LendingService)

	apiV1 := r.App.Group("/api/v1")
	{
		// user route
		user := apiV1.Group("/user")
		{
			user.POST("/register", ctrlUser.Register)
			user.POST("/login", ctrlUser.Login)
			user.POST("/logout", r.AuthMiddleware(), ctrlUser.Logout)
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

			lendingBook := book.Group("").Use(r.AuthMiddleware(), r.RoleMiddleware(utils.RoleAdmin, utils.RoleMember))
			{
				lendingBook.POST("/:id/borrow", ctrlLending.BorrowBook)
				lendingBook.POST("/:id/return", ctrlLending.ReturnBook)
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
		logPrefix += fmt.Sprintf("[%s][%s]", utils.InterfaceString(dataJWT["jti"]), utils.InterfaceString(dataJWT["user_id"]))

		// Check if token is blacklisted
		_, err = r.BlacklistRepo.GetByToken(tokenString)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; blacklistRepo.GetByToken; Error: %+v", logPrefix, err))
			res := response.Response(http.StatusInternalServerError, utils.MsgFail, logId, nil)
			res.Error = err.Error()
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
			return
		}

		//the token is valid but has been logged out
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Invalid Token: %s; Error: token is blacklisted;", logPrefix, tokenString))
			res := response.Response(http.StatusUnauthorized, utils.MsgFail, logId, nil)
			res.Error = "Please login and try again"
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
			return
		}

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

		isAllowed := slices.Contains(allowedRoles, userRole)
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
