package controller

import (
	"fmt"
	"github.com/alapidas/roper/model"
	"os"
	"strings"
)

var (
	reposEnvVar = "REPO_LOCS"
	repos       model.Repos
)

func init() {
	Initialize()
}

// Initialize directories and register repositories to manage.  To be run
// at startup of the server.
func Initialize() error {
	repoDirString := os.Getenv(reposEnvVar)
	if repoDirString == "" {
		return fmt.Errorf("No repository dirs specified")
	}
	repoDirStrings := strings.Split(repoDirString, ",")
	for _, repoStr := range repoDirStrings {
		if exists, err := pathExists(repoStr); err != nil || !exists {
			return fmt.Errorf("Unable to locate directory %v", repoStr)
		}
		repos.AddRepo(model.Repo{LocalPath: repoStr})
	}
	return nil
}
