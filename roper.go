package main

import (
	"fmt"
	"github.com/alapidas/roper/controller"
	"github.com/codegangsta/cli"
	"os"
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
				myServer := controller.NewServer()
				if err := myServer.Start(c); err != nil {
					os.Exit(1)
				}
			},
		},
	}
	return app
}
