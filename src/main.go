package main

import (
    //"database/sql"
    //"encoding/json"
    "log"
    //"net/http"

    //"github.com/gorilla/mux"
 //  "gorestful/entity/pessoa"
    _ "github.com/lib/pq"
    
    dbConfig "gorestful/dbconfig"
    dbQueries "gorestful/dbqueries"
)


//var db *sql.DB
// DB set up


func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}


func main() {
	var db = dbConfig.SetupDB()
	dbQueries.InitLingMap(db)
	
//	p := pessoa.Pessoa{ Apelido: "will3", Nome:"Willian Santos", Nascimento:"2023-03-03", Stack: []string{"Java"} }
//	dbQueries.InsertPessoa(db, p)
	
	log.Println(dbQueries.CountPessoas(db))

	dbQueries.GetPessoaById(db, "f7379ae8-8f9b-4cd5-8221-51efe19e721b")
	log.Println(dbQueries.GetTerm(db, "ana"))
	log.Println("Main Finished")
}

