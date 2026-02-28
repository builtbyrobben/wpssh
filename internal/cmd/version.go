package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// VersionCmd prints the wpgo version.
type VersionCmd struct {
	Version string `hidden:"" default:"${version}"`
	Commit  string `hidden:"" default:"${commit}"`
	Date    string `hidden:"" default:"${date}"`
}

func (c *VersionCmd) Run(globals *Globals) error {
	if globals.JSON {
		out := map[string]string{
			"version": c.Version,
			"commit":  c.Commit,
			"date":    c.Date,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	if globals.Plain {
		fmt.Println(strings.Join([]string{c.Version, c.Commit, c.Date}, "\t"))
		return nil
	}

	fmt.Printf("wpgo %s", c.Version)
	if c.Commit != "" {
		fmt.Printf(" (%s", c.Commit)
		if c.Date != "" {
			fmt.Printf(" %s", c.Date)
		}
		fmt.Printf(")")
	}
	fmt.Println()
	return nil
}
