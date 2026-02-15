package cmd

// Config commands (wp config — wp-config.php management)

type ConfigCmdGroup struct {
	Get    ConfigGetCmd    `cmd:"" help:"Get config value"`
	Set    ConfigSetCmd    `cmd:"" help:"Set config value"`
	Delete ConfigDeleteCmd `cmd:"" help:"Delete config constant"`
	Has    ConfigHasCmd    `cmd:"" help:"Check if constant exists"`
	List   ConfigListCmd   `cmd:"" help:"List config values"`
	Path   ConfigPathCmd   `cmd:"" help:"Show wp-config.php path"`
}

type ConfigGetCmd struct{ Name string `arg:"" help:"Constant name"` }
type ConfigSetCmd struct {
	Name  string `arg:"" help:"Constant name"`
	Value string `arg:"" help:"Constant value"`
}
type ConfigDeleteCmd struct{ Name string `arg:"" help:"Constant name"` }
type ConfigHasCmd struct{ Name string `arg:"" help:"Constant name"` }
type ConfigListCmd struct{}
type ConfigPathCmd struct{}

func (c *ConfigGetCmd) Run(g *Globals) error {
	return execPassthrough(g, "config", "get", c.Name)
}

func (c *ConfigSetCmd) Run(g *Globals) error {
	return execPassthrough(g, "config", "set", c.Name, c.Value)
}

func (c *ConfigDeleteCmd) Run(g *Globals) error {
	return execPassthrough(g, "config", "delete", c.Name)
}

func (c *ConfigHasCmd) Run(g *Globals) error {
	return execBoolCheck(g, "config", "has", c.Name, "Config constant")
}

func (c *ConfigListCmd) Run(g *Globals) error {
	return execPassthrough(g, "config", "list")
}

func (c *ConfigPathCmd) Run(g *Globals) error {
	return execPassthrough(g, "config", "path")
}
