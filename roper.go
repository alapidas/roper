package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	app := makeApp()
	app.RunAndExitOnError()
}

func makeApp() *cli.App {
	app := cli.NewApp()
	app.Name = "roper"
	app.Usage = "A repo manager that doesn't suck"
	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "start a server to manage a repo",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "loc, repoloc",
					Usage: "required - comma-separated location(s) of local repos to manage (ex., /var/repos/myRepo)",
				},
			},
			Before: func(c *cli.Context) error {
				if len(c.StringSlice("loc")) == 0 {
					return fmt.Errorf("Must include local repo(s) to manage")
				}
				return nil
			},
			Action: func(c *cli.Context) {
				startServer(c)
			},
		},
	}
	return app
}

func startServer(c *cli.Context) {
	log.WithFields(log.Fields{
		"locations": c.StringSlice("loc"),
	}).Info("Starting server")
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/repos", reposHandler)
	router.HandleFunc("/repos/{repoId}", reposHandler)
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":3001")
}

func reposHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res.Write([]byte(fmt.Sprintf("hai there repoId %v", vars["repoId"])))
}
