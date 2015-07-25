package controller

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alapidas/roper/model"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
	"os/exec"
)

type server struct {
	repos  *model.Repos
	crPath string
}

func NewServer() *server {
	return &server{repos: model.NewRepos()}
}

// StartServer starts a server, and does all the things you might expect a
// server to do, except make you dinner.
func (self *server) Start(c *cli.Context) error {
	if err := self.initOSDeps(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Unable to initialize OS dependencies")
		return err
	}
	if err := self.initRepos(c.StringSlice("loc")); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Unable to initialize repositories")
		return err

	}
	log.WithFields(log.Fields{
		"locations": c.StringSlice("loc"),
	}).Info("Starting server")
	n, err := initHandler()
	if err != nil {
		// TODO log something good
	}
	n.Run(":3001")
	return nil
}

func (self *server) initRepos(locs []string) error {
	for _, loc := range locs {
		if exists, err := pathExists(loc); err != nil || !exists {
			return fmt.Errorf("Unable to locate directory %v", loc)
		}
		self.repos.AddRepo(model.NewRepo(loc, loc))
	}
	return nil
}

// initOSDeps initializes any OS-level dependencies that are needed by the server
func (self *server) initOSDeps() error {
	// Make sure createrepo is installed
	crPath, err := exec.LookPath("createrepo")
	if err != nil {
		return errors.New("Unable to find 'createrepo' in $PATH, please install it")
	}
	log.WithFields(log.Fields{
		"crPath": self.crPath,
	}).Info("createrepo executable found")
	self.crPath = crPath
	return nil
}

// TODO needs to change
func initHandler() (*negroni.Negroni, error) {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/repos", reposHandler)
	router.HandleFunc("/repos/{repoId}", reposHandler)
	n := negroni.Classic()
	n.UseHandler(router)
	return n, nil
}

func reposHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res.Write([]byte(fmt.Sprintf("hai there repoId %v", vars["repoId"])))
}
