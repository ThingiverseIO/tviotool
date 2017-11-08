package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ThingiverseIO/console"
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	setupCfg  = "configure"
	setupShow = "show"
)

func init() {
	RootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup ThingiverseIO",
	RunE: func(cmd *cobra.Command, args []string) error {

		aborted := false
		for !aborted {
			var action string
			action, aborted = console.AskOption("Choose Action", setupCfg, setupShow)
			switch action {
			case setupCfg:
				cfg := &config.UserConfig{}
				ifaces, err := net.Interfaces()
				if err != nil {
					console.Println(err)
					return nil
				}
				opts := map[string]interface{}{}
				for _, iface := range ifaces {
					addrs, err := iface.Addrs()
					if err != nil {
						console.Println(err)
						return nil
					}
					if len(addrs) == 0 {
						continue
					}
					addr := strings.Split(addrs[0].String(), "/")[0]
					name := fmt.Sprintf("%s => %s", iface.Name, addr)
					opts[name] = addr
				}
				var iiface interface{}
				console.Println("")
				_, iiface, aborted = console.AskOptionValue("Please select a network device", opts)
				if aborted {
					aborted = false
					continue
				}
				cfg.Interface = iiface.(string)
				console.Println("")
				console.Println("Selected Interface", cfg.Interface)
				console.Println("")
				cfg.Logger, aborted = console.AskOption("Please select a logger", "none", "stdout", "stderr", "file")
				if aborted {
					aborted = false
					continue
				}
				if cfg.Logger == "file" {
					cfg.Logger = console.AskString("Please specify filename: ")
				}
				console.Println("")
				console.Println("Selected Logger", cfg.Logger)
				console.Println("")

				cfg.Debug = console.AskYesOrNo("Enable debug?", false)

				console.Println("")
				console.Println("Selected Config")
				console.Println("")
				console.Println(cfg)
				console.Println("")

				if !console.AskYesOrNo("Save", false) {
					return nil
				}

				loc, aborted := console.AskOption("Where to save?", "Home", "Current Directory")
				if aborted {
					aborted = false
					continue
				}
				format, aborted := console.AskOption("Which format?", "JSON", "YAML")
				if aborted {
					aborted = false
					continue
				}

				switch loc {
				case "Home":
					usr, err := user.Current()
					if err != nil {
						console.Println(err)
						return nil
					}
					loc = usr.HomeDir
				case "Current Directory":
					loc = "."
				}

				var fname string
				var fcontent []byte
				switch format {
				case "JSON":
					fname = ".tvio.json"
					fcontent, err = json.MarshalIndent(cfg, "", "\t")
					if err != nil {
						console.Println(err)
						return nil
					}
				case "YAML":
					fname = ".tvio.yml"
					fcontent, err = yaml.Marshal(cfg)
					if err != nil {
						console.Println(err)
						return nil
					}
				}
				path := filepath.Join(loc, fname)

				err = ioutil.WriteFile(path, fcontent, os.ModePerm)
				if err != nil {
					console.Println(err)
					return nil
				}

				console.Println("done")

			case setupShow:
				console.Println("Reading Configuration")
				console.Println("")
				console.Println("The Current Configuration is:")
				console.Println("")
				console.Println(config.Configure())
				aborted = console.AskEnterOrAbort("Press return to continue, 'q' to quit", "q")
			}
		}
		return nil
	},
}
