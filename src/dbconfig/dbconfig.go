package dbconfig

import (
    "database/sql"
     "os"
     "log"
    //"encoding/json"
    "fmt"
    //"log"
    //"net/http"

    //"github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

const DB_DRIVER = "postgres"
var (

	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT = os.Getenv("DB_PORT")
	DB_USER = os.Getenv("DB_USER")
	DB_PASSWORD = os.Getenv("DB_PASSWORD") 
	DB_NAME = os.Getenv("DB_NAME") 
)

func SetupDB() *sql.DB {

    dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
    log.Printf(dbinfo)
    db, err := sql.Open(DB_DRIVER, dbinfo)
    
    checkErr(err)
    fmt.Println("Connected!");
     
    return db
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
