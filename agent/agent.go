package agent

import (
	"os"
	"os/signal"

	"github.com/ThingiverseIO/logger"
	"github.com/joernweissenborn/eventual2go"
	"github.com/spf13/pflag"
)

var (
	log = logger.New("TVIO AGENT")
)

func Run(flags *pflag.FlagSet, services ...Service) (err error) {
	shutdown := eventual2go.NewShutdown()
	for _, service := range services {
		log.Init(service.Name())
		err = service.Start(flags)
		if err != nil {
			log.Error(err)
			return
		}
		shutdown.Register(service)
		log.PrintSuccess()
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("Shutting down")
	shutdown.Do(nil)
	return
}
