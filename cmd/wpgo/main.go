package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/builtbyrobben/wpssh/internal/cmd"
	"github.com/builtbyrobben/wpssh/internal/errfmt"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("wpgo"),
		kong.Description("WordPress management over SSH"),
		kong.DefaultEnvars("WPGO"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoExpandSubcommands: true,
		}),
		kong.Vars{
			"version": version,
			"commit":  commit,
			"date":    date,
		},
	)
	err := ctx.Run(&cli.Globals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wpgo: %s\n", errfmt.Format(err))
		os.Exit(1)
	}
}
