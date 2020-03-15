package cmd

import (
	"os"
	"path"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/xlog"
)

const DefDir string = "."
const LogName string = "feliratok.log"

type base struct {
	*cli.Flagger
	ConfigPath string `flag:"config, Load config from this file"`
	Log        bool   `flag:"log, Create log file"`

	path   []string
	config *config.Config
	log    xlog.Logger
}

func (b *base) Parse(args []string) error {
	if err := b.FlagSet.Parse(args); err != nil {
		return err
	}
	b.path = b.FlagSet.Args()
	return nil
}

func (b *base) init() {
	b.log = xlog.New(xlog.Config{
		Output: xlog.MultiOutput{
			xlog.LevelOutput{
				Info: xlog.NewConsoleOutput(),
			},
		},
	})
	b.getConfig()
	if b.config.Log {
		file, err := os.OpenFile(path.Join(DefDir, LogName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			xlog.Fatal(err)
		}
		defer file.Close()
		b.log = xlog.New(xlog.Config{
			Output: xlog.MultiOutput{
				xlog.LevelOutput{
					Info: xlog.NewConsoleOutput(),
				},
				xlog.NewLogfmtOutput(file),
			},
		})
	}
}

func (b *base) getConfig() {
	var err error
	filePath := path.Join(DefDir, b.ConfigPath)

	b.config, err = config.NewConf(filePath)
	if err != nil {
		xlog.Fatal(err)
	}
	if b.Log {
		b.config.Log = b.Log
	}
	if b.config.DecompSuffix == "" {
		b.config.DecompSuffix = "_decomp"
	}
}
