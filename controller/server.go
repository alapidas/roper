package controller

/*
import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alapidas/roper/model"
	"github.com/codegangsta/cli"
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
	n, err := InitHandler()
	if err != nil {
		log.Errorf("Unable to initialize web server: %s\n", err)
		return err
	}
	n.Run(":3001")
	return nil
}

// initRepos does one thing and it does it well - it initializes repo objects
// on the running server.
func (self *server) initRepos(locs []string) error {
	for _, loc := range locs {
		if exists, err := pathExists(loc); err != nil || !exists {
			return fmt.Errorf("Unable to locate directory %v", loc)
		}
		self.repos.AddRepo(model.NewRepo(loc, loc))
	}
	return nil
}

// initOSDeps initializes any OS-level dependencies that are needed by the
// server.  Most principally, the existence of the `createrepo` program in
// the $PATH.
func (self *server) initOSDeps() error {
	// Make sure createrepo is installed
	var err error
	self.crPath, err = exec.LookPath("createrepo")
	if err != nil {
		return errors.New("Unable to find 'createrepo' in $PATH, please install it")
	}
	log.WithFields(log.Fields{
		"crPath": self.crPath,
	}).Info("createrepo executable found")
	return nil
}
*/
