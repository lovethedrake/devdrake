package main

import (
	"fmt"

	"github.com/lovethedrake/drakecore/config"
	"github.com/urfave/cli"
)

func list(c *cli.Context) error {
	configFile := c.GlobalString(flagFile)
	listPipelines := c.Bool(flagPipeline)
	config, err := config.NewConfigFromFile(configFile)
	if err != nil {
		return err
	}
	if listPipelines {
		for _, pipeline := range config.AllPipelines() {
			fmt.Println(pipeline.Name())
		}
	} else {
		for _, job := range config.AllJobs() {
			fmt.Println(job.Name())
		}
	}
	return nil
}
