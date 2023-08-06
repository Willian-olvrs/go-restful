package dbconfig

import (
    "database/sql"
    //"encoding/json"
    "fmt"
    //"log"
    //"net/http"

    //"github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

const (

	DB_HOST = "localhost"
	DB_PORT = "5455"
	DB_USER = "postgres"
	DB_PASSWORD = "password"
	DB_NAME = "gorestful"
)

func SetupDB() *sql.DB {

    dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
    db, err := sql.Open("postgres", dbinfo)
    
    checkErr(err)
    fmt.Println("Connected!");
     
    return db
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
