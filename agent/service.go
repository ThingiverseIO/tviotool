package agent

import (
	"github.com/joernweissenborn/eventual2go"
	"github.com/spf13/pflag"
)

type Service interface {
	eventual2go.Shutdowner
	Name() string
	Start(flags *pflag.FlagSet) error
	RegisterFlags(flags *pflag.FlagSet)
}
