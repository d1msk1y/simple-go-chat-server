package Database

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var DB *sql.DB

func TryConnectDB() error {
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "chat",
		AllowNativePasswords: true,
	}

	var err error
	DB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
		return err
	}

	pingErr := DB.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		return err
	}
	fmt.Println("Connected!")
	return nil
}
