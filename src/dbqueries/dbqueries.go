package dbqueries


import (
	"sync"
    "database/sql"
    "fmt"
    "strings"
    "time"
    "errors"
    "github.com/lib/pq"
    "gorestful/entity/pessoa"
    uuidGoogle "github.com/google/uuid"
)

var lingMap = make(map[string]int)
var pessoaMapCache sync.Map
var idPessoaMapCache sync.Map
var lastPessoaUpdate = time.Now()


func InitLingMap(db *sql.DB) {

	rows, err := db.Query(`SELECT id, ling FROM ling`)
	checkErr(err)
	defer rows.Close()
    
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
    defer rows.Close()
    var count string
    
    for rows.Next() {
    
	    err := rows.Scan(&count)    
  		checkErr(err)
    }
    
    return count
} 

func InsertPessoa(db *sql.DB, p pessoa.Pessoa) (*pessoa.Pessoa, error) {

    return runInsertPessoa(db, p)
}

func GetTerm(db *sql.DB, term string) map[string]pessoa.Pessoa {

    var mapP = runQueryTerm(db, term)
   
    return mapP
}

func GetPessoaById(db *sql.DB, id string) (*pessoa.Pessoa, error) {
    
    return runQueryPessoaById(db, id)
}

func runInsertPessoa(db *sql.DB, p pessoa.Pessoa) (*pessoa.Pessoa, error) {
    
    uuid := uuidGoogle.New().String()
    p.Id = &uuid
        
    pessoaMapCache.Store(p.Id, p)
	
    return &p, nil
}

func runQueryTerm(db *sql.DB, term string) map[string]pessoa.Pessoa {

    BulkInsert(db, true)

    rows_query_pessoas, err := db.Query(`
    	SELECT pessoa_select.id, apelido, nome, nascimento, ling 
			FROM ling RIGHT JOIN 
				(SELECT * FROM stack RIGHT JOIN 
			 		(SELECT * FROM pessoa WHERE apelido=$1 OR nome LIKE '%'||$1||'%') AS p ON id_pessoa = p.id) AS pessoa_select
	 	ON ling.id=pessoa_select.id_ling`, term)
    checkErr(err)
   	defer rows_query_pessoas.Close()
    
    rows_query_ling, err := db.Query(`
    	SELECT pessoa.id, apelido, nome, nascimento, ling
			FROM pessoa RIGHT JOIN 
				(SELECT * FROM stack LEFT JOIN 
					(SELECT * FROM ling WHERE ling=$1) AS ling_select ON id_ling=ling_select.id) AS stack_select
		ON pessoa.id=stack_select.id_pessoa`, term)
    checkErr(err)
   	defer rows_query_ling.Close()
    
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
        var ling sql.NullString

        var err = rows.Scan(&id, &apelido, &nome, &nascimento, &ling)
        // check errors
        checkErr(err)
                
        p := mapPessoa[id]
        
        p.Id = &id;
        p.Apelido = &apelido
        p.Nome = &nome
        p.Nascimento = &nascimento
        
        if(ling.Valid){
        	p.Stack = append(p.Stack, ling.String)
        }

		
		mapPessoa[id] = p
    }

	return mapPessoa
}
	
func runQueryPessoaById(db *sql.DB, id string) (*pessoa.Pessoa, error) {

    pCache, isMapped := pessoaMapCache.Load(id)
    _, isIdMapped := idPessoaMapCache.Load(id)
    
    if(isMapped){
    	pCachePessoa := pCache.(pessoa.Pessoa)
    	return &pCachePessoa, nil
    }
    
    if(!isIdMapped) {
    
    	return nil, errors.New("Id inexistente")
    }

    rows, err := db.Query(`
    	SELECT pessoa.id, apelido, nome, nascimento, ling FROM pessoa 
    		LEFT JOIN (SELECT * FROM stack LEFT JOIN ling ON id_ling = ling.id) AS stack ON pessoa.id=stack.id_pessoa 
    		WHERE pessoa.id=$1`, id)
	    
    checkErr(err)
    defer rows.Close()
    
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
        
        p.Id = &id;
        p.Apelido = &apelido
        p.Nome = &nome
        p.Nascimento = &nascimento
        
        stack = append(stack,ling)
    }
    
    p.Stack = stack
    
    return &p, nil
}

func BulkInsert(db *sql.DB, force bool) error {

	if( force) {
		time.NewTicker(500 * time.Millisecond)
	}
	
	err := runInsertPessoaBulk(db)
  	err = runInsertStackBulk(db)
	
	if( err != nil ){
		return err
	}
	
	pessoaMapCache.Range( func( key interface{}, value interface{}) bool {
		pessoaMapCache.Delete(key)
		return true
	})

	return nil
}


func runInsertStackBulk( db *sql.DB) error {

	var	values []interface{}
	var placeholders []string
	index := 0
		
	pessoaMapCache.Range(func( key interface{}, pI interface{}) bool {
	
		p := pI.(pessoa.Pessoa)
		for _, ling := range p.Stack {
	
			updateLingMap(db, ling)
			_, isIdMapped := idPessoaMapCache.Load(p.Id)
			if(isIdMapped) {
				placeholders = append(placeholders, fmt.Sprintf("($%d,$%d)", index*2+1,index*2+2))

				values = append(values, p.Id, lingMap[ling])
				index++
			}
		}
		
		return true
	})
	
	stringJoin := strings.Join(placeholders, ",")
	
	if (stringJoin == "" || len(stringJoin) == 0) {
		return nil
	}
	
	insertStatement := fmt.Sprintf("INSERT INTO stack(id_pessoa, id_ling) VALUES %s ON CONFLICT DO NOTHING", stringJoin)
	
	txn, err := db.Begin()
	
	if err != nil {
		return err
	}
	
	_, err = txn.Exec(insertStatement, values...)
	
	if err != nil {
		return err
	}
	err = txn.Commit(); 
	
	return nil
}


func runInsertPessoaBulk( db *sql.DB) error {

	var	values []interface{}
	var placeholders []string
	index := 0
	
	pessoaMapCache.Range(func( key interface{}, pI interface{}) bool {
	
		p := pI.(pessoa.Pessoa)
		_, isIdMapped := idPessoaMapCache.Load(p.Id)
		
		if(!isIdMapped) {
			
			placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d)", index*4+1,index*4+2, index*4+3, index*4+4))
			values = append(values, p.Id, p.Apelido, p.Nome, p.Nascimento)
			index++

		}		
		return true		
	})
	
	stringJoin := strings.Join(placeholders, ",")
	
	if (stringJoin == "" || len(stringJoin) == 0) {
		return nil
	}
	
	insertStatement := fmt.Sprintf("INSERT INTO pessoa (id, apelido, nome, nascimento) VALUES %s ON CONFLICT DO NOTHING", stringJoin)
	
	txn, err := db.Begin()

	if err != nil {
		return err
	}
	
	_, err = txn.Exec(insertStatement, values...)
	
	if err != nil {
		return err
	}
	err = txn.Commit();
	
	addIdToCache(db)
	
	return nil
}

func addIdToCache(db *sql.DB) {
	
	rows, err := db.Query(`SELECT pessoa.id FROM pessoa;`)
	    
    checkErr(err)
    defer rows.Close()
    
    for rows.Next() {
        var id string

        err = rows.Scan(&id)
        checkErr(err)
        
        idPessoaMapCache.Store(id, id)
    }
}


func updateLingMap(db *sql.DB, ling string) error {

	_, lingIsMapped := lingMap[ling]
		
	if(!lingIsMapped) {
	
		insertQuery, err := db.Query(`INSERT INTO ling(ling) VALUES($1);`, ling)
		
		if(insertQuery != nil) {
			defer insertQuery.Close()
		}
		
		if err, ok := err.(*pq.Error); ok {	
			
			switch err.Code.Name() {
				case "unique_violation":
					return nil
				default:
					return err
			}
		}
		checkErr(err)
		
		InitLingMap(db)
	}
	
	return nil
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
