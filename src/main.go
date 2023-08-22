package main

import (
    //"database/sql"
    "os"
    "errors"
    "strings"
    "fmt"
    "io"
    "encoding/json"
    "log"
    "net/http"
	"strconv"
	"time"	
    "github.com/gorilla/mux"
   
    _ "github.com/lib/pq"
    
    "gorestful/entity/pessoa"
    dbConfig "gorestful/dbconfig"
    dbQueries "gorestful/dbqueries"
)

var DB = dbConfig.SetupDB()

func main() {
	
	go setUpTicker() 
	initRoutes()
}

func setUpTicker() {

    ticker := time.NewTicker(50 * time.Millisecond)
    for _ = range ticker.C {
          dbQueries.BulkInsert(DB)
	}
} 

func initRoutes() {


	router := mux.NewRouter()
	
	router.HandleFunc("/pessoas/{id}", getPessoas).Methods("GET")
	router.HandleFunc("/contagem-pessoas", getContagemPessoas).Methods("GET")
	router.HandleFunc("/pessoas", getTermo).Methods("GET")
	router.HandleFunc("/pessoas", postPessoas).Methods("POST")
	
	var port string
	port = strings.Join([]string{":", os.Getenv("HTTP_PORT")}, "")
	log.Fatal(http.ListenAndServe(port, router))
}

func postPessoas(w http.ResponseWriter, r *http.Request) {

	dec := json.NewDecoder(r.Body)
	
	var p pessoa.Pessoa
	
	err := dec.Decode(&p)
	
	if(err != nil){
		checkPostRequestErr(err, w, r)
		return 
	}
	
	if(p.Nome == nil || p.Apelido == nil || p.Nascimento == nil || len(*p.Nome) > 100 || len(*p.Apelido) > 32 || !checkDateIsValid(*p.Nascimento) || !checkStackIsValid(p.Stack)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	
	checkErr(err)
    pessoaInserted, errInsert := dbQueries.InsertPessoa(DB, p)
    
    if( errInsert != nil ){

		w.WriteHeader(http.StatusUnprocessableEntity)
		return		
    }

    w.Header().Set("Location", "/pessoas/"+*pessoaInserted.Id)    
    w.WriteHeader(http.StatusCreated)			

}

func checkDateIsValid( date string) bool {

	d := strings.Split(date, "-")
	
	if(len(d) != 3){
		return false
	}
	
	year := d[0]
	month := d[1]
	day := d[2]
	
	if(len(year) != 4) {
		return false
	}
		
	if(len(month) != 2 || month < "01" || month > "12") {
	    return false
	}
	
	if(len(day) != 2 || day < "01" || day > "31") {
		return false
	}
	
	if(month == "02" && day > "28"){
	
		yearInt, err := strconv.Atoi(year)
		
		if(err != nil){
			return false;
		}
	
		if(!isLeapYear(yearInt)){
			return false;		
		}
		
		if(day != "29"){
			return false;
		}
	}
	
	if(day > "30" && (month == "04" || month == "06" || month == "09" || month == "11")){
		return false
	}

 	return true
}

func isLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	} else if year%100 == 0 {
		return false
	} else if year%4 == 0 {
		return true
	}

	return false
}

func checkStackIsValid( stack []string ) bool {

	for i := 0; i < len(stack); i++  {
		if(len(stack[i]) > 32){
			return false;
		}
	}
		
 	return true;
}


func getTermo(w http.ResponseWriter, r *http.Request) {

    termQuery := r.URL.Query()["t"]
	
	if(len(termQuery) == 0 ||  len(termQuery) > 1) {
		
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	term := termQuery[0]
	mapPessoas := dbQueries.GetTerm(DB, term )
	
	listPessoas := []pessoa.Pessoa{}
	
	for  _, value := range mapPessoas {
		listPessoas = append(listPessoas, value)
	}

	json.NewEncoder(w).Encode(listPessoas)
}

func getContagemPessoas(w http.ResponseWriter, r *http.Request) {
	
	json.NewEncoder(w).Encode(dbQueries.CountPessoas(DB))
}

func getPessoas(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
    id := params["id"]
	
	p, err := dbQueries.GetPessoaById(DB, id)
	
	if(err != nil){
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(*p)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
func checkPostRequestErr(err error, w http.ResponseWriter, r *http.Request){


        var syntaxError *json.SyntaxError
        var unmarshalTypeError *json.UnmarshalTypeError

        switch {
        // Catch any syntax errors in the JSON and send an error message
        // which interpolates the location of the problem to make it
        // easier for the client to fix.
        case errors.As(err, &syntaxError):
            msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
            http.Error(w, msg, http.StatusBadRequest)

        // In some circumstances Decode() may also return an
        // io.ErrUnexpectedEOF error for syntax errors in the JSON. There
        // is an open issue regarding this at
        // https://github.com/golang/go/issues/25956.
        case errors.Is(err, io.ErrUnexpectedEOF):
            msg := fmt.Sprintf("Request body contains badly-formed JSON")
            http.Error(w, msg, http.StatusBadRequest)

        // Catch any type errors, like trying to assign a string in the
        // JSON request body to a int field in our Person struct. We can
        // interpolate the relevant field name and position into the error
        // message to make it easier for the client to fix.
        case errors.As(err, &unmarshalTypeError):
            msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
            http.Error(w, msg, http.StatusBadRequest)

        // An io.EOF error is returned by Decode() if the request body is
        // empty.
        case errors.Is(err, io.EOF):
            msg := "Request body must not be empty"
            http.Error(w, msg, http.StatusBadRequest)

        // Otherwise default to logging the error and sending a 500 Internal
        // Server Error response.
        default:
            log.Print(err.Error())
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        }
        return
}


