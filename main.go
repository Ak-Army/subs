package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/cli/command"

	_ "github.com/Ak-Army/subs/cmd"
)

func main() {
	c := cli.New("subs", fmt.Sprintf("%s, build time: %s", Version, BuildTime))
	cli.RootCommand().Authors = []string{"authors goes here"}
	cli.RootCommand().AddCommand("completion", &command.Completion{})
	c.SetDefault("download")
	c.Run(context.Background(), os.Args)
}
