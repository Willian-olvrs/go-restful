package main

import (
    //"database/sql"
    //"encoding/json"
    "fmt"
    //"log"
    //"net/http"

    //"github.com/gorilla/mux"
    _ "github.com/lib/pq"
    
    dbConfig "gorestful/dbconfig"
    Pessoa "gorestful/entity/pessoa"
)


//var db *sql.DB
// DB set up


func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}


func main() {
	var db = dbConfig.SetupDB();
	
	Pessoa.GetPessoas(db)
	fmt.Println("Hello World!")
}
