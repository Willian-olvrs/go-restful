package main

import (
    //"database/sql"
    "encoding/json"
    "log"
    "net/http"

    "github.com/gorilla/mux"
 //  "gorestful/entity/pessoa"
    _ "github.com/lib/pq"
    
    dbConfig "gorestful/dbconfig"
    dbQueries "gorestful/dbqueries"
)

var DB = dbConfig.SetupDB()

func main() {

	dbQueries.InitLingMap(DB)
	initRoutes()
	
	
//	p := pessoa.Pessoa{ Apelido: "will3", Nome:"Willian Santos", Nascimento:"2023-03-03", Stack: []string{"Java"} }
//	dbQueries.InsertPessoa(db, p)
	
	log.Println(dbQueries.CountPessoas(DB))


	log.Println(dbQueries.GetTerm(DB, "ana"))
	log.Println("Main Finished")
	initRoutes()
}

func initRoutes() {


	router := mux.NewRouter()
	
	
	router.HandleFunc("/pessoas/{id}", getPessoas).Methods("GET")
	router.HandleFunc("/contagem-pessoas", getContagemPessoas).Methods("GET")
	
	log.Fatal(http.ListenAndServe(":8000", router))
}


func getContagemPessoas(w http.ResponseWriter, r *http.Request) {
	
	json.NewEncoder(w).Encode(dbQueries.CountPessoas(DB))
}

func getPessoas(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
    id := params["id"]
    log.Println("GET /pessoas/", id)
	
	json.NewEncoder(w).Encode(dbQueries.GetPessoaById(DB, id))
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}


