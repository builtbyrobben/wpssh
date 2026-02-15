package cmd

// Maintenance commands

type MaintenanceCmd struct {
	Activate   MaintenanceActivateCmd   `cmd:"" help:"Activate maintenance mode"`
	Deactivate MaintenanceDeactivateCmd `cmd:"" help:"Deactivate maintenance mode"`
	Status     MaintenanceStatusCmd     `cmd:"" help:"Show maintenance mode status"`
}

type MaintenanceActivateCmd struct{}
type MaintenanceDeactivateCmd struct{}
type MaintenanceStatusCmd struct{}

func (c *MaintenanceActivateCmd) Run(g *Globals) error {
	return execPassthrough(g, "maintenance-mode", "activate")
}

func (c *MaintenanceDeactivateCmd) Run(g *Globals) error {
	return execPassthrough(g, "maintenance-mode", "deactivate")
}

func (c *MaintenanceStatusCmd) Run(g *Globals) error {
	return execPassthrough(g, "maintenance-mode", "status")
}
