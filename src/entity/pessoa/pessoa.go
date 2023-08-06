package pessoa

import (
    "database/sql"
    //"encoding/json"
    "fmt"
    //"log"
    //"net/http"

    //"github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

type Pessoa struct {

	id string `json:"id"`
	apelido string `json:"apelido"`
	nome string `json:"nome"`
	nascimento string `json:"nascimento"`
}

func GetPessoas(db *sql.DB ) []Pessoa {

    fmt.Println("Getting pessoas...")

    rows, err := db.Query("SELECT * FROM pessoa")
    
    // check errors
    checkErr(err)

    var pessoas []Pessoa

    for rows.Next() {
        var id string
        var apelido string
        var nome string
        var nascimento string

        err = rows.Scan(&id, &apelido, &nome, &nascimento)

        // check errors
        checkErr(err)
		fmt.Println("%s, %s", id, nome)
        pessoas = append(pessoas, Pessoa{id: id, apelido: apelido, nome: nome, nascimento:nascimento})
    }

    return pessoas
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
