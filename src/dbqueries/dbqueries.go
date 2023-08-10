package dbqueries

import (
    "database/sql"
    //"encoding/json"
    "log"
    //"net/http"

    //"github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "gorestful/entity/pessoa"
    uuidGoogle "github.com/google/uuid"
)

var lingMap = make(map[string]int)

func InitLingMap(db *sql.DB) {

	rows, err := db.Query(`SELECT id, ling FROM ling`)
	checkErr(err)
    
	for rows.Next() {
			
		var id int;
		var ling string;
		
		var err = rows.Scan(&id, &ling)
		checkErr(err)
		
		lingMap[ling] = id
	}
}

func CountPessoas(db *sql.DB) string {

	rows, err := db.Query(`SELECT count(id) FROM pessoa`)
    checkErr(err)
    
    var count string
    
    for rows.Next() {
    
	    err := rows.Scan(&count)    
  		checkErr(err)
    }

    
    return count
} 

func InsertPessoa(db *sql.DB, p pessoa.Pessoa){

    log.Println("Insert pessoa =", p.Nome)
    runInsertPessoa(db, p)
}

func GetTerm(db *sql.DB, term string) map[string]pessoa.Pessoa {

    log.Println("Getting pessoa by term =", term)
    var mapP = runQueryTerm(db, term)
   
    return mapP
}

func GetPessoaById(db *sql.DB, id string) pessoa.Pessoa {

    log.Println("Getting pessoa by id =", id)
    
    var p = runQueryPessoaById(db, id)
    return p
}

func runInsertPessoa(db *sql.DB, p pessoa.Pessoa) {
    
    uuid := uuidGoogle.New().String()
    _, err := db.Query(`INSERT INTO pessoa (id, apelido, nome, nascimento) 
    		VALUES ($1, $2, $3, $4);`, uuid, p.Apelido, p.Nome, p.Nascimento)	 	
    checkErr(err)
    
    runInsertStack(db, p.Stack, uuid)
}

func runInsertStack( db *sql.DB, stack []string, idPessoa string) {

	log.Println("Insert stack =", idPessoa)
	for _, ling := range stack {
	
		lingIndex, lingIsMapped := lingMap[ling]
		
		if(!lingIsMapped) {
		
			_, err := db.Query(`INSERT INTO ling(ling) VALUES ($1);`, ling)
			checkErr(err)
			
			InitLingMap(db)
			lingIndex, lingIsMapped = lingMap[ling]
		}
		
		_, err := db.Query(`INSERT INTO stack(id_pessoa, id_ling) VALUES ($1, $2)`, idPessoa, lingIndex)
		checkErr(err)
	}
}

func runQueryTerm(db *sql.DB, term string) map[string]pessoa.Pessoa {

    if( false ) {
		log.Println(`SELECT pessoa_select.id, apelido, nome, nascimento, ling 
			FROM ling RIGHT JOIN 
				(SELECT * FROM stack RIGHT JOIN 
			 		(SELECT * FROM pessoa WHERE apelido='`, term, `' OR nome LIKE '%'||`,term,`||'%') AS p ON id_pessoa = p.id) AS pessoa_select
	 	ON ling.id=pessoa_select.id_ling`)		
	}

    rows_query_pessoas, err := db.Query(`
    	SELECT pessoa_select.id, apelido, nome, nascimento, ling 
			FROM ling RIGHT JOIN 
				(SELECT * FROM stack RIGHT JOIN 
			 		(SELECT * FROM pessoa WHERE apelido=$1 OR nome LIKE '%'||$1||'%') AS p ON id_pessoa = p.id) AS pessoa_select
	 	ON ling.id=pessoa_select.id_ling`, term)
    checkErr(err)
    
    if( false ) {
		    log.Println(`SELECT pessoa.id, apelido, nome, nascimento, ling
					FROM pessoa RIGHT JOIN 
						(SELECT * FROM stack RIGHT JOIN 
							(SELECT * FROM ling WHERE ling=`,term,`) AS ling_select ON id_ling=ling_select.id) AS stack_select
				ON pessoa.id=stack_select.id_pessoa`)    
    }
    
    rows_query_ling, err := db.Query(`
    	SELECT pessoa.id, apelido, nome, nascimento, ling
			FROM pessoa RIGHT JOIN 
				(SELECT * FROM stack RIGHT JOIN 
					(SELECT * FROM ling WHERE ling=$1) AS ling_select ON id_ling=ling_select.id) AS stack_select
		ON pessoa.id=stack_select.id_pessoa`, term)
    checkErr(err)
    
    mapPessoa := make(map[string]pessoa.Pessoa)
    
    mapPessoa = addToMap(rows_query_pessoas, mapPessoa);
    mapPessoa = addToMap(rows_query_ling, mapPessoa);
    
    return mapPessoa
}

func addToMap(rows *sql.Rows, mapPessoa map[string]pessoa.Pessoa) map[string]pessoa.Pessoa {


	for rows.Next() {
	
        var id string
        var apelido string
        var nome string
        var nascimento string
        var ling string

        var err = rows.Scan(&id, &apelido, &nome, &nascimento, &ling)
        // check errors
        checkErr(err)
                
        p := mapPessoa[id]
        
        p.Id = id;
        p.Apelido = apelido
        p.Nome = nome
        p.Nascimento = nascimento
        
		p.Stack = append(p.Stack,ling)
		
		mapPessoa[id] = p
    }

	return mapPessoa
}
	
func runQueryPessoaById(db *sql.DB, id string) pessoa.Pessoa {

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
