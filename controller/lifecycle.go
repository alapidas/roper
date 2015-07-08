package controller

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alapidas/roper/model"
	"os"
	"os/exec"
	"strings"
)

var (
	reposEnvVar = "REPO_LOCS"
	repos       model.Repos
	crPath      string
)

func init() {
	Initialize()
}

// Initialize directories and register repositories to manage.  To be run
// at startup of the server.
func Initialize() error {

	// Make sure createrepo is installed
	crPath, err := exec.LookPath("createrepo")
	if err != nil {
		log.Fatal("Unable to find 'createrepo' in $PATH, please install it")
	}
	log.WithFields(log.Fields{
		"crPath": crPath,
	}).Info("createrepo executable found")

	// Initialize repo objects
	repoDirString := os.Getenv(reposEnvVar)
	if repoDirString == "" {
		return fmt.Errorf("No repository dirs specified")
	}
	repoDirStrings := strings.Split(repoDirString, ",")
	for _, repoStr := range repoDirStrings {
		if exists, err := pathExists(repoStr); err != nil || !exists {
			return fmt.Errorf("Unable to locate directory %v", repoStr)
		}
		repos.AddRepo(model.Repo{Name: repoStr, LocalPath: repoStr})
	}
	return nil
}
