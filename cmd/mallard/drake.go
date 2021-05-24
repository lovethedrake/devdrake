package main

import (
	"fmt"
	"os"

	"github.com/lovethedrake/go-drake/config"
	"github.com/lovethedrake/mallard/pkg/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Mallard"
	app.HelpName = "mallard"
	app.Usage = "execute Drake jobs and pipelines using the local Docker daemon"
	app.Version = version.Version()
	if version.Commit() != "" {
		app.Version = fmt.Sprintf("%s+%s", app.Version, version.Commit())
	}
	app.Version = fmt.Sprintf(
		"%s supports DrakeSpec %s",
		app.Version,
		config.SupportedSpecVersions,
	)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  flagsFile,
			Usage: "specify the location of configuration",
			Value: "Drakefile.yaml",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "list",
			Aliases:   []string{"ls"},
			Usage:     "list all drake jobs or pipelines",
			UsageText: "drake list [options]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  flagsPipeline,
					Usage: "list pipelines instead of jobs",
				},
			},
			Action: list,
		},
		{
			Name:      "run",
			Usage:     "execute drake jobs(s) or pipeline(s)",
			UsageText: "drake run name... [options]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  flagsPipeline,
					Usage: "execute a pipeline instead of a job",
				},
				cli.BoolFlag{
					Name:  flagsDebug,
					Usage: "display debug info",
				},
				cli.IntFlag{
					Name:  flagsConcurrency,
					Usage: "maximum number of jobs to execute at once",
					Value: 1,
				},
				cli.StringFlag{
					Name:  flagsSecretsFile,
					Usage: "specify the location of drake secrets",
					Value: "Drakesecrets.yaml",
				},
			},
			Action: run,
		},
	}
	fmt.Println()
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
