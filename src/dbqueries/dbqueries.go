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

    var mapPSync = runQueryTerm(db, term)
    var mapP = make(map[string]pessoa.Pessoa)

    mapPSync.Range( func(key any, value any) bool {
    
    	mapP[key.(string)] = value.(pessoa.Pessoa)
    	return true
    })
   
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


func pessoaTermQuery(db *sql.DB, term string, mapPessoa *sync.Map, wg *sync.WaitGroup) {

    defer wg.Done()
	rows_query_pessoas, err := db.Query(`
    	SELECT pessoa_select.id, apelido, nome, nascimento, ling 
			FROM ling RIGHT JOIN 
				(SELECT * FROM stack RIGHT JOIN 
			 		(SELECT * FROM pessoa WHERE to_tsvector('english', apelido) @@ plainto_tsquery('english', $1)  OR nome LIKE '%'||$1||'%') AS p ON id_pessoa = p.id) AS pessoa_select
	 	ON ling.id=pessoa_select.id_ling`, term)

    checkErr(err)
   	defer rows_query_pessoas.Close()

	addToMap(rows_query_pessoas, mapPessoa)

}

func lingTermQuery(db *sql.DB, term string, mapPessoa *sync.Map, wg *sync.WaitGroup) {

    defer wg.Done()

	rows_query_ling, err := db.Query(`
    	SELECT pessoa.id, apelido, nome, nascimento, ling
			FROM pessoa RIGHT JOIN 
				(SELECT * FROM stack LEFT JOIN 
					(SELECT * FROM ling WHERE to_tsvector('english', ling) @@ plainto_tsquery('english', $1) ) AS ling_select ON id_ling=ling_select.id) AS stack_select 
					ON pessoa.id=stack_select.id_pessoa`, term)
    checkErr(err)

   	defer rows_query_ling.Close()
	addToMap(rows_query_ling, mapPessoa)
}

func runQueryTerm(db *sql.DB, term string) sync.Map {

    BulkInsert(db, true)
   
    var mapPessoa sync.Map
    var wg sync.WaitGroup
    
    wg.Add(1)
    go pessoaTermQuery(db, term, &mapPessoa, &wg)
    wg.Add(1)
    go lingTermQuery(db, term, &mapPessoa, &wg)

    wg.Wait()

    return mapPessoa
}

func addToMap(rows *sql.Rows, mapPessoa *sync.Map) {


	for rows.Next() {
	
        var id string
        var apelido string
        var nome string
        var nascimento string
        var ling sql.NullString

        var err = rows.Scan(&id, &apelido, &nome, &nascimento, &ling)
        // check errors
        checkErr(err)
                
        var p pessoa.Pessoa
        
        p.Id = &id;
        p.Apelido = &apelido
        p.Nome = &nome
        p.Nascimento = &nascimento
        
        if(ling.Valid){
        	p.Stack = append(p.Stack, ling.String)
        }

        objFromMap, ok := mapPessoa.Load(id)
		if(ok) {
		
			pessoaFromMap := objFromMap.(pessoa.Pessoa)
			pessoaFromMap.Id = p.Id;
    	    pessoaFromMap.Apelido = p.Apelido
    	    pessoaFromMap.Nome = p.Nome
	        pessoaFromMap.Nascimento = p.Nascimento
	        pessoaFromMap.Stack = p.Stack
		} else {
		
			mapPessoa.Store(*p.Id, p)
		}
    }
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
		time.NewTicker(100 * time.Millisecond)
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
