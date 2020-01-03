package main

import (
	"os"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/subs/cmd"
)

func main() {
	c := cli.New("subs", "1.0.0")
	c.Authors = []string{"authors goes here"}
	c.Add(
		&cmd.Download{
			ConfigPath: "config.yml",
		})
	c.SetDefault("download")
	c.Run(os.Args)
}
