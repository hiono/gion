package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/tasuku43/gion/internal/config"
)

func runConfig(args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		printConfigHelp(os.Stdout)
	}

	var setProvider string
	fs.StringVar(&setProvider, "set-provider", "", "set the default provider")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	// If --set-provider is provided, update the config
	if setProvider != "" {
		cfg := &config.Config{Provider: setProvider}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Provider set to:", setProvider)
		return nil
	}

	// Otherwise, show current config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Print current config
	fmt.Println("Current Configuration")
	fmt.Print("  Provider: ")
	if cfg.Provider == "" {
		fmt.Println("(auto-detect)")
	} else {
		fmt.Println(cfg.Provider)
	}

	// Show config file path
	configPath, err := config.DefaultConfigPath()
	if err == nil {
		fmt.Printf("\n  Config file: %s\n", configPath)
	}

	return nil
}
