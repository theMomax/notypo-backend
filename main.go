package main

import (
	"os"

	"github.com/theMomax/notypo-backend/api"
	"github.com/theMomax/notypo-backend/config"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "notypo api-backend-server"
	app.Version = config.Version + " (" + config.GitCommit + ") Built at: " + config.BuildTime
	app.Description = "This is notypo's RESTful api-backend. See more at github.com/theMomax/notypo-backend."
	app.Usage = ""
	app.Author = "Max Obermeier"
	app.Commands = []cli.Command{cli.Command{
		Name:   "serve",
		Usage:  "start api-server",
		Action: serve,
		Flags:  config.Options(),
		Before: config.Load,
	}}
	app.Run(os.Args)
}

func serve(ctx *cli.Context) {
	api.Serve()
}
