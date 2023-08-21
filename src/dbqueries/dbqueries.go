package dbqueries


import (
	"sync"
    "database/sql"
    "fmt"
    "strings"
    "errors"
    _ "github.com/lib/pq"
    "gorestful/entity/pessoa"
    uuidGoogle "github.com/google/uuid"
)

var lingMap = make(map[string]int)
var pessoaMapCache sync.Map
var idPessoaMapCache sync.Map
var apelidoMapCache sync.Map

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

    return  runQueryTerm(db, term)
}

func GetPessoaById(db *sql.DB, id string) (*pessoa.Pessoa, error) {
    
    return runQueryPessoaById(db, id)
}

func runInsertPessoa(db *sql.DB, p pessoa.Pessoa) (*pessoa.Pessoa, error) {
    
    uuid := uuidGoogle.New().String()
    p.Id = &uuid
    
    _, isApelidoMapped := apelidoMapCache.Load(*p.Apelido)
			
	if(isApelidoMapped) {
		
		return nil, errors.New("Apelido já inserido")
	}
	
    pessoaMapCache.Store(*p.Id, p)

    return &p, nil
}


func runQueryTerm(db *sql.DB, term string) map[string]pessoa.Pessoa {

   	mapPessoa := make(map[string]pessoa.Pessoa)
   	
   	rows_query_pessoas, err := db.Query(`
    	SELECT id, apelido, nome, nascimento, stack
			FROM pessoa 
			WHERE search_p LIKE '%'||$1||'%'`, term)
	checkErr(err)
   	defer rows_query_pessoas.Close()
	for rows_query_pessoas.Next() {
	
		var id string
        var apelido string
        var nome string
        var nascimento string
        var stack sql.NullString

        var err = rows_query_pessoas.Scan(&id, &apelido, &nome, &nascimento, &stack)
       	checkErr(err)
       	
		p,isPMapped := mapPessoa[id]
		if(!isPMapped){
			var pC pessoa.Pessoa
			
			pC.Id = &id
			pC.Apelido = &apelido
			pC.Nome = &nome
			pC.Nascimento = &nascimento
			mapPessoa[id] = pC
		}
		
 		if(stack.Valid){
			p.Stack = strings.Split(stack.String, " ")
        }
	}
    return mapPessoa
}

func BulkInsert(db *sql.DB) error {

	
	err := runInsertPessoaBulk(db)
	
	if( err != nil ){
		return err
	}

	return nil
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
    	SELECT pessoa.id, apelido, nome, nascimento, stack 
    		FROM pessoa 
    		WHERE pessoa.id=$1`, id)
	    
    checkErr(err)
    defer rows.Close()
    
    var p pessoa.Pessoa
	var stackArray []string
	
    for rows.Next() {
        var id string
        var apelido string
        var nome string
        var nascimento string
        var stack string

        err = rows.Scan(&id, &apelido, &nome, &nascimento, &stack)
        checkErr(err)
        
        p.Id = &id;
        p.Apelido = &apelido
        p.Nome = &nome
        p.Nascimento = &nascimento
        
        stackArray = strings.Split(stack, " ")
    }
    
    p.Stack = stackArray
    
    return &p, nil
}

func runInsertPessoaBulk(db *sql.DB) error {

	var	values []interface{}
	var placeholders []string
	index := 0
	
	pessoaMapCache.Range(func( key interface{}, value interface{}) bool {
	
		p := value.(pessoa.Pessoa)
		search := *p.Nome + *p.Apelido
		
		_, isIdMapped := idPessoaMapCache.Load(*p.Id)
		
		_, isApelidoMapped := apelidoMapCache.Load(*p.Apelido)
		
		if(isIdMapped) {
			return true
		}
		
		if(isApelidoMapped) {
		
			return true
		}
		
	
		if( len(p.Stack) > 0 ) {
			stack := ""
			for _, ling := range p.Stack {
			
				search = search + ":" + ling
				stack = stack + " " + ling
			}
		
			pessoaInfo := fmt.Sprintf(`($%d,$%d,$%d,$%d`, index*6+1,index*6+2, index*6+3, index*6+4)
			placeholders = append(placeholders, fmt.Sprintf("%s,$%d,$%d)", pessoaInfo, index*6+5, index*6+6))
			values = append(values, *p.Id, *p.Apelido, *p.Nome, *p.Nascimento, stack, search)
			index++
		} else {
			pessoaInfo := fmt.Sprintf(`($%d,$%d,$%d,$%d`, index*6+1,index*6+2, index*6+3, index*6+4)			
			placeholders = append(placeholders, fmt.Sprintf("%s,$%d,$%d)", pessoaInfo, index*6+5,index*6+6))
			values = append(values, *p.Id, *p.Apelido, *p.Nome, *p.Nascimento, "null", search)
			index++
		}
		
		return true		
	})
	
	stringJoin := strings.Join(placeholders, ",")
	
	if (stringJoin == "" || len(stringJoin) == 0) {
		return nil
	}
	
	insertStatement := fmt.Sprintf("INSERT INTO pessoa (id, apelido, nome, nascimento,stack,search_p) VALUES %s ON CONFLICT DO NOTHING", stringJoin)
	
	txn, err := db.Begin()

	if err != nil {
		return err
	}
	
	_, err = txn.Exec(insertStatement, values...)
	
	if err != nil {
		return err
	}
	err = txn.Commit();
	
	pessoaMapCache.Range( func( key interface{}, value interface{}) bool {
		pessoaMapCache.Delete(key)
		return true
	})
	
	addIdApelidToCache(db)
	
	return nil
}


func addIdApelidToCache(db *sql.DB) {
	
	rows, err := db.Query(`SELECT id,apelido FROM pessoa;`)
	    
    checkErr(err)
    defer rows.Close()
    
    for rows.Next() {
        var id string
        var apelido string

        err = rows.Scan(&id, &apelido)
        checkErr(err)
        
        idPessoaMapCache.Store(id, id)
        apelidoMapCache.Store(apelido, apelido)
    }
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
