package pessoa

import (
    //"encoding/json"
    //"log"
)

type Pessoa struct {

	Id string `json:"id"`
	Apelido string `json:"apelido"`
	Nome string `json:"nome"`
	Nascimento string `json:"nascimento"`
	Stack []string
}


func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
