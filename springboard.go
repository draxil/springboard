package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/draxil/springboard/watch"
	"os"
)

const version = "0.2.1"
const author = "Joe Higton"
const author_email = "draxil@gmail.com"

func main() {
	app := app()
	app.Run(os.Args)
}

func app() *cli.App {
	app := cli.NewApp()
	app.Name = "springboard"
	app.Usage = "Watch a directory for files and send them places"
	commands, flags, _ := setup()
	app.Commands = commands
	app.Version = version
	app.Authors = []cli.Author{cli.Author{Name: author,
		Email: author_email}}
	app.Flags = flags
	return app
}

func setup() (c []cli.Command, f []cli.Flag, cfg *watch.Config) {
	cfg = &watch.Config{}
	f = global_flags(cfg)

	add_command := func(cmd cli.Command) {
		cmd = wrap_cmd(cfg, cmd)
		c = append(c, cmd)
	}

	add_command(http_post_command(cfg, run_watch))
	add_command(echo_command(cfg, run_watch))

	return
}

func wrap_cmd(cfg *watch.Config, c cli.Command) cli.Command {
	a := c.Action.(func(*cli.Context))
	c.Action = func(c *cli.Context) {
		setup_action(cfg, c)
		a(c)
	}
	return c
}

func setup_action(cfg *watch.Config, c *cli.Context) {
	sparanoia := c.GlobalString("paranoia")
	switch sparanoia {
	case "off":
		cfg.Paranoia = watch.NoParanoia
	case "basic":
		cfg.Paranoia = watch.BasicParanoia
	case "extra":
		cfg.Paranoia = watch.ExtraParanoia
	default:
		fmt.Fprintln(os.Stderr, "Invalid choice of paranoia=", c)
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}
}

func global_flags(cfg *watch.Config) (f []cli.Flag) {

	f = []cli.Flag{
		cli.StringFlag{
			Name:        "archive",
			Usage:       "move the file to this location after successful action",
			Destination: &cfg.ArchiveDir,
		},
		cli.StringFlag{
			Name:        "error-dir",
			Usage:       "move the file to this location after a failed action",
			Destination: &cfg.ErrorDir,
		},
		cli.StringFlag{
			Name:  "paranoia",
			Usage: "Do we take extra steps to ensure the file has been completely written? See documentation for full details. Values: off, basic, extra.",
			Value: "basic",
		},
		cli.BoolFlag{
			Name:        "process-existing",
			Usage:       "Process any pre-existing files in the directory on startup. Obviously best used alongside an archive option of some kind.",
			Destination: &cfg.ProcessExistingFiles,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "enable verbose debug output",
			Destination: &cfg.Debug,
		},
		cli.BoolTFlag{
			Name:        "log-errors",
			Usage:       "enable logging of errors (default=true)",
			Destination: &cfg.ReportErrors,
		},
		cli.BoolFlag{
			Name:        "log-actions",
			Usage:       "enable logging of actions",
			Destination: &cfg.ReportActions,
		},
		cli.StringSliceFlag{
			Name:  "testing",
			Usage: "Used to set testing options, usually only required for development & testing",
			Value: (*cli.StringSlice)(&cfg.TestingOptions),
		},
	}
	return
}

func run_watch(c *watch.Config) {
	e := watch.Watch(c)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

func http_post_command(cfg *watch.Config, action func(*watch.Config)) cli.Command {
	var pa watch.PostAction
	//	var http_headers cli.StringSlice
	return cli.Command{
		Name:  "post",
		Usage: "post the file somewhere",
		Flags: []cli.Flag{
			// cli.StringSliceFlag{
			// 	Name:  "header",
			// 	Value: &http_headers,
			// 	Usage: "Set extra http headers, format is KEY:VAL",
			// },
			cli.StringFlag{
				Name:        "mime",
				Destination: &pa.Mime,
				Usage:       "Force the mime type on the post",
			},
			cli.StringFlag{
				Name:        "uname",
				Destination: &pa.BasicAuthUsername,
				Usage:       "Triggers use of HTTP basic auth (See RFC 2617, Section 2.) with the provided username",
			},
			cli.StringFlag{
				Name:        "pass",
				Destination: &pa.BasicAuthPwd,
				Usage:       "Set the password for HTTP basic auth.",
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
		Name:      "echo",
		Usage:     "echo the full filepath",
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
				&watch.EchoAction{},
			}
			cfg.Dir = args.First()
			action(cfg)
		},
	}
}
