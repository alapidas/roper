package controller

import (
	//"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
)

func InitHandler() (*negroni.Negroni, error) {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/v1/repos", reposHandler)
	router.HandleFunc("/v1/repos/{repoId}", reposHandler)
	router.HandleFunc("/v1/repos/{repoId}/packages", packagesHandler)
	router.HandleFunc("/v1/repos/{repoId}/packages/{packageId}", packagesHandler)
	n := negroni.Classic()
	n.UseHandler(router)
	return n, nil
}

func reposHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res.Write([]byte(fmt.Sprintf("hai there repoId %v", vars["repoId"])))
}

func packagesHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res.Write([]byte(fmt.Sprintf("package %s in repo %s", vars["packageId"],
		vars["repoId"])))
}
