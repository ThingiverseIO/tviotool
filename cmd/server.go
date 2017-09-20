package cmd

import (
	"tviotool/web"

	"github.com/spf13/cobra"
)

func init() {
	serverCmd.PersistentFlags().IntP("port", "p", web.DefaultPort, "Port to serve")
	serverCmd.PersistentFlags().StringP("interface", "i", web.DefaultInterface, "Interface to serve")
	serverCmd.PersistentFlags().StringP("directory", "d", web.DefaultDirectory, "Directory to serve")
	RootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"srv"},
	Short:   "Serves the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flags().Parse(args)
		w := web.New(web.Configure(cmd.PersistentFlags()))
		return w.Serve()
	},
}
