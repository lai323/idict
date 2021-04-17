package practice

import (

	// tea "github.com/charmbracelet/bubbletea"
	// "github.com/go-resty/resty/v2"
	idictconfig "github.com/lai323/idict/config"
)

func Start(config *idictconfig.Config) func() error {
	// return func() error {
	// 	client := resty.New()
	// if len(config.Proxy) > 0 { client.SetProxy(config.Proxy) }
	// 	p := tea.NewProgram(NewModel(*config, client))

	// 	p.EnableMouseCellMotion()
	// 	p.EnterAltScreen()
	// 	err := p.Start()
	// 	p.ExitAltScreen()
	// 	p.DisableMouseCellMotion()

	// 	return err
	// }
	return nil
}
