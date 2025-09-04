package main

import (
	"database/sql"
	"digital-book-lending/app"
	"digital-book-lending/database"
	_ "digital-book-lending/docs"
	"digital-book-lending/repository"
	"digital-book-lending/services"
	"digital-book-lending/utils"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// @title Digital Book Lending API
// @version 1.0
// @description This is a sample server for a digital book lending service.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://www.linkedin.com/in/zaidus-zhuhur/
// @contact.email zaiduszhuhur@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	var (
		err         error
		sqlBookLend *sql.DB
	)
	if timeZone, err := time.LoadLocation("Asia/Jakarta"); err != nil {
		utils.WriteLog(utils.LogLevelError, "time.LoadLocation - Error: "+err.Error())
	} else {
		time.Local = timeZone
	}

	if err = godotenv.Load(".env"); err != nil && os.Getenv("APP_ENV") == "" {
		log.Fatalf("Error app environment")
	}

	myAddr := "unknown"
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myAddr = ipnet.IP.String()
				break
			}
		}
	}

	myAddr += strings.Repeat(" ", 15-len(myAddr))
	os.Setenv("ServerIP", myAddr)
	utils.WriteLog(utils.LogLevelInfo, "Server IP: "+myAddr)

	var port, appName string
	flag.StringVar(&port, "port", os.Getenv("PORT"), "port of the service")
	flag.StringVar(&appName, "appname", os.Getenv("APP_NAME"), "service name")
	flag.Parse()
	utils.WriteLog(utils.LogLevelInfo, "APP: "+appName+"; PORT: "+port)

	//Load app config
	confID := utils.GetAppConf("CONFIG_ID", "", nil)
	utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("ConfigID: %s", confID))

	runMigration()

	db, sqlBookLend := database.ConnDb()
	defer sqlBookLend.Close()

	// Repositories
	bookRepo := repository.NewBookRepo(db)
	userRepo := repository.NewUserRepo(db)
	blacklistRepo := repository.NewBlacklistRepo(db)
	lendingRepo := repository.NewLendingRepo(db)

	// Services
	bookService := services.NewBookService(bookRepo, db)
	userService := services.NewUserService(userRepo, blacklistRepo)
	lendingService := services.NewLendingService(lendingRepo, bookRepo, db)

	routes := app.NewRoutes(bookService, userService, lendingService, blacklistRepo)

	routes.BookLending()
	err = routes.App.Run(fmt.Sprintf(":%s", port))
	FailOnError(err, "Failed run service")
}

func runMigration() {
	m, err := migrate.New(
		utils.GetEnv("PATH_MIGRATE", "file://migrations").(string),
		fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s",
			utils.GetEnv("DB_USERNAME", "").(string),
			utils.GetEnv("DB_PASS", "").(string),
			utils.GetEnv("DB_HOST", "").(string),
			utils.GetEnv("DB_PORT", "").(string),
			utils.GetEnv("DB_NAME", "").(string)),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err.Error() != "no change" {
		log.Fatal(err)
	}
	utils.WriteLog(utils.LogLevelInfo, "Migration Success")
}
