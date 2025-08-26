package database

import (
	"database/sql"
	"digital-book-lending/utils"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type dbUtils struct {
	db  *gorm.DB
	sql *sql.DB
}

var dbInstanceConnDb *dbUtils
var dbOnceConnDb sync.Once

func ConnDb() (*gorm.DB, *sql.DB) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		utils.GetEnv("DB_USERNAME", "").(string),
		utils.GetEnv("DB_PASS", "").(string),
		utils.GetEnv("DB_HOST", "").(string),
		utils.GetEnv("DB_PORT", "").(string),
		utils.GetEnv("DB_NAME", "").(string))

	dbOnceConnDb.Do(func() {
		utils.WriteLog(utils.LogLevelDebug, fmt.Sprintf("ConnDb; Initialize db connection..."))

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
		if err != nil {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("ConnDb; %s Error: %s", dsn, err.Error()))
			log.Fatalln("ConnDb; Failed to conn db: ", err.Error())
			return
		}

		maxIdle := 10
		maxIdleTime := 5 * time.Minute
		maxConn := 100
		maxLifeTime := time.Hour

		sqlDB, err := db.DB()
		if err != nil {
			utils.WriteLog(utils.LogLevelError, fmt.Sprintf("ConnDb.sqlDB; %s Error: %s", dsn, err.Error()))
			log.Fatalln("ConnDb.sqlDB; Failed to conn db: ", err.Error())
			return
		}

		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetConnMaxIdleTime(maxIdleTime)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(maxConn)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(maxLifeTime)

		//db.Debug()

		dbInstanceConnDb = &dbUtils{
			db:  db,
			sql: sqlDB,
		}
	})

	return dbInstanceConnDb.db, dbInstanceConnDb.sql
}
