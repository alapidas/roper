package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/repos", ReposHandler)
	router.HandleFunc("/repos/{repoId}", ReposHandler)
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":3001")
}

func ReposHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res.Write([]byte(fmt.Sprintf("hai there repoId %v", vars["repoId"])))
}
