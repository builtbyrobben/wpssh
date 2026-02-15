package cmd

import "fmt"

// VersionCmd prints the wpgo version.
type VersionCmd struct {
	Version string `hidden:"" default:"${version}"`
}

func (c *VersionCmd) Run(globals *Globals) error {
	fmt.Printf("wpgo %s\n", c.Version)
	return nil
}
