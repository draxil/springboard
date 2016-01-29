package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/draxil/springboard/watch"
	"os"
)

const version = "0.1.0"
const author = "Joe Higton"
const author_email = "draxil@gmail.com"

func main() {
	app := cli.NewApp()
	app.Name = "springboard"
	app.Usage = "Watch a directory for files and send them places"
	commands, flags, cfg := setup()
	app.Commands = commands
	app.Version = version
	app.Authors = []cli.Author{cli.Author{Name: author,
		Email: author_email}}
	app.Flags = flags
	app.Action = func(c *cli.Context) {
		fmt.Fprintln( os.Stderr, cfg.Debug)
		cli.ShowAppHelp(c)
	}
	app.Run(os.Args)
}

func setup() (c []cli.Command, f []cli.Flag, cfg * watch.Config) {
	cfg = &watch.Config{}
	f = global_flags( cfg )
	c = append(c, http_post_command(cfg, run_watch))
	c = append(c, echo_command(cfg, run_watch))
	return
}


func global_flags( cfg * watch.Config )( f []cli.Flag){
	f = []cli.Flag{
		cli.BoolFlag{
			Name : "debug",
			Usage : "enable verbose messaging",
			Destination : &cfg.Debug,
		},
	}
	return
}

func run_watch(c * watch.Config) {
	e := watch.Watch(c)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

func http_post_command(cfg *watch.Config, action func(*watch.Config)) cli.Command {
	var pa watch.PostAction
	var http_headers cli.StringSlice
	return cli.Command{
		Name:  "post",
		Usage: "post the file somewhere",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "header",
				Value: &http_headers,
				Usage: "set extra http headers, format is KEY:VAL",
			},
			cli.StringFlag{
				Name: "mime",
				Value: pa.Mime,
				Usage: "Force the mime type on the post",
			},
		},
		ArgsUsage: "URL DIR",
		Action: func(c *cli.Context) {

			args := c.Args()

			bail := func() {
				cli.ShowSubcommandHelp(c)
				os.Exit(1)
			}

			if !args.Present() {
				bail()
			}
			arg := 0

			next := func() string {
				val := args.Get(arg)
				arg++
				if val == "" {
					bail()
				}
				return val
			}

			pa.To = next()

			cfg.Actions = []watch.Action{
				&pa,
			}
			cfg.Dir = next()
			action(cfg)
		},
	}
}

func echo_command(cfg *watch.Config, action func(*watch.Config)) cli.Command {
	return cli.Command{
		Name:  "echo",
		Usage: "echo the full filepath",
		ArgsUsage: "DIR",
		Action: func(c *cli.Context) {

			args := c.Args()

			bail := func() {
				cli.ShowSubcommandHelp(c)
				os.Exit(1)
			}

			if !args.Present() {
				bail()
			}

			cfg.Actions = []watch.Action{
				&watch.EchoAction{
				},
			}
			cfg.Dir = args.First()
			action(cfg)
		},
	}
}
