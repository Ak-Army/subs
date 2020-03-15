package main

import (
	"fmt"
	"os"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/subs/cmd"
)

func main() {
	c := cli.New("subs", fmt.Sprintf("%s, build time: %s", Version, BuildTime))
	c.Authors = []string{"authors goes here"}
	c.Add(
		cmd.NewDownload("config.yml"),
		cmd.NewSeason("config.yml"),
	)
	c.SetDefault("download")
	c.Run(os.Args)
}
