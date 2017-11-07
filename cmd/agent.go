package cmd

import (
	"fmt"
	"strings"
	"tviotool/agent"
	"tviotool/agent/services/web"

	"github.com/spf13/cobra"
)

var services = map[string]agent.Service{
	"web": &web.Service{},
}

func names() (names []string) {
	for name := range services {
		names = append(names, name)
	}
	return
}

func init() {
	RootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentListCmd)
	for _, service := range services {
		service.RegisterFlags(agentCmd.PersistentFlags())
	}
}

var agentCmd = &cobra.Command{
	Use:       fmt.Sprintf("agent [%s]", strings.Join(names(), " | ")),
	ValidArgs: names(),
	Short:     "Starts an agent serving the given services",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}
		var srvs []agent.Service
		for _, name := range args {
			if service, ok := services[name]; ok {
				srvs = append(srvs, service)
			} else {
				cmd.Println("Invalid service", name)
				return cmd.Usage()
			}
		}
		cmd.Flags().Parse(args)
		return agent.Run(cmd.PersistentFlags(), srvs...)
	},
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "Serves the current Lists all available services",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Available Services")
		for _, service := range services {
			cmd.Println("\t-", service.Name())
		}
		return nil
	},
}
