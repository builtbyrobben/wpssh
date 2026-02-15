package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/builtbyrobben/wpssh/internal/cmd"
)

var version = "dev"

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
		kong.Vars{"version": version},
	)
	err := ctx.Run(&cli.Globals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wpgo: %v\n", err)
		os.Exit(1)
	}
}
