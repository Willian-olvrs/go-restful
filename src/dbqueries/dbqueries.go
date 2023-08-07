package dbqueries

import (
    "database/sql"
    //"encoding/json"
    "log"
    //"net/http"

    //"github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "gorestful/entity/pessoa"
)


func GetPessoaById(db *sql.DB, id string) pessoa.Pessoa {

    log.Println("Getting pessoa by id =", id)
    
    var p = runQuery(db, id)
    
    log.Println("Getting pessoa by id =", id, "finished")
    return p
}

func runQuery(db *sql.DB, id string) pessoa.Pessoa {

	if(false) {
	//TODO DEBUG env var
		log.Println(`SELECT pessoa.id, apelido, nome, nascimento, ling FROM pessoa 
				LEFT JOIN (SELECT * FROM stack LEFT JOIN ling ON id_ling = ling.id) AS stack ON pessoa.id=stack.id_pessoa 
				WHERE pessoa.id=`, id)
    }

    rows, err := db.Query(`
    	SELECT pessoa.id, apelido, nome, nascimento, ling FROM pessoa 
    		LEFT JOIN (SELECT * FROM stack LEFT JOIN ling ON id_ling = ling.id) AS stack ON pessoa.id=stack.id_pessoa 
    		WHERE pessoa.id=$1`, id)
    
    checkErr(err)
    
    var p pessoa.Pessoa
	var stack []string
	
    for rows.Next() {
        var id string
        var apelido string
        var nome string
        var nascimento string
        var ling string

        err = rows.Scan(&id, &apelido, &nome, &nascimento, &ling)
     
        // check errors
        checkErr(err)
        
        p.Id = id;
        p.Apelido = apelido
        p.Nome = nome
        p.Nascimento = nascimento
        
        stack = append(stack,ling)
    }
    
    p.Stack = stack
    
    return p
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
