package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alapidas/roper/controller"
	//"github.com/codegangsta/cli"
	//"os"
)

func main() {
	log.Infof("Starting Server")
	// create controller
	rc, err := controller.Init("/Users/alapidas/Downloads/roper.db")
	if err != nil {
		log.Fatalf("Unable to initialize application: %s", err)
	}
	defer rc.Close()
	// Discover
	name := "Test Epel"
	path := "/Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/epel"
	if err = rc.Discover(name, path); err != nil {
		log.WithFields(log.Fields{
			name: name,
			path: path,
		}).Fatal("Unable to discover repo - exiting")
	}
	//app := makeApp()
	//app.RunAndExitOnError()
}

/*func makeApp() *cli.App {
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
				myServer := controller.NewServer()
				if err := myServer.Start(c); err != nil {
					os.Exit(1)
				}
			},
		},
	}
	return app
}*/
