package cmd

import (
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"cfg"},
	Short:   "Shows the tvio configuration in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("The Current Configuration is:")
		fmt.Println("")
		fmt.Println(config.Configure())
		return nil
	},
}
