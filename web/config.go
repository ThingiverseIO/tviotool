package web

import (
	"fmt"
	"os/user"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Interface string
	Port      int
	Directory string
	NoDir bool
}

func (c Config) Address() string {
	return ifaceAddrToString(c.Interface, c.Port)
}

func ifaceAddrToString(iface string, port int) string {
	return fmt.Sprintf("%s:%d", iface, port)
}

var (
	DefaultPort      = 8085
	DefaultInterface = "127.0.0.1"
	DefaultDirectory = "."
)

func Configure(flags ...*pflag.FlagSet) (cfg *Config) {
	v := viper.New()

	v.SetDefault("port", DefaultPort)
	v.SetDefault("interface", DefaultInterface)
	v.SetDefault("directory", DefaultDirectory)

	//Configfile
	v.SetConfigName("config")
	v.AddConfigPath(".") // First look in CWD

	usr, err := user.Current()
	if err != nil {
		v.AddConfigPath(usr.HomeDir) // Then in user home

	}

	for _, f := range flags {
		v.BindPFlags(f)
	}

	cfg = &Config{}
	v.Unmarshal(cfg)
	return

}
