package main

import(
	// "errors"
	// "log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var dsn string

func initDatabase(dbname, address, username, password string) error {
	dsn = username + ":" + password + "@(" + address + ":3306)/" + dbname
	
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	
	logger.Println("db opened")
	
	//validate the database connection
	err = db.Ping()
	
	if err != nil{
		return err
	} else {
		logger.Println("Database connection validated")
	}
	
	
	return nil
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	
	if err != nil {
		return nil, err
	}
	
	return db, nil
}

