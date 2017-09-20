package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tvioweb",
	Short: "ThingiverseIO Webserver",
	Long:  `TODO`,
	// Run: func(cmd *cobra.Command, args []string) {
	//         log := logging.Get()
	//         log.Info("hello")
	//         // Do Stuff Here
	// },
}
