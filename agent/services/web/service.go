package web

import (
	"net/http"
	"os"

	"github.com/ThingiverseIO/logger"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/gorilla/websocket"
	"github.com/joernweissenborn/eventual2go"
	"github.com/spf13/pflag"
)

var log = logger.New("TVIO WEB SERVER")

type clientArrived struct{}
type clientLeft struct{}

type Service struct {
	*eventual2go.Reactor
	cfg     *Config
	clients map[uuid.UUID]*client
}

func (c *Service) Name() string { return "Web Service" }

func (c *Service) RegisterFlags(flags *pflag.FlagSet) {
	flags.IntP("port", "p", DefaultPort, "Port to serve")
	flags.StringP("interface", "i", DefaultInterface, "Interface to serve")
	flags.StringP("directory", "d", DefaultDirectory, "Directory to serve")
	flags.Bool("nodir", false, "If set, no directory is served, only websockets")
}

func (c *Service) Start(flags *pflag.FlagSet) (err error) {
	cfg := Configure(flags)
	log.Debug("Debug Mode Enabled")
	log.Init("Starting up")
	c.Reactor = eventual2go.NewReactor()
	c.cfg = cfg
	c.clients = map[uuid.UUID]*client{}

	c.React(clientArrived{}, c.onClientArrived)
	c.React(clientLeft{}, c.onClientLeft)
	if !c.cfg.NoDir {
		c.initWebserver()
	}
	c.initWebSocketListener()
	go c.serve()
	return
}

func (c *Service) serve() {
	err := http.ListenAndServe(c.cfg.Address(), nil)
	if err != nil {
		log.Error("Error Starting Webserver: ", err)
		p,_:=os.FindProcess(os.Getpid())
		p.Signal(os.Interrupt)
	}
	return
}

func (c *Service) initWebSocketListener() {
	log.Init("Initializing Websocket Server")
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Info("New connection ")
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if err != nil {
			log.Error(err)
			return
		}
		c.Fire(clientArrived{}, newClient(conn))
	}
	http.HandleFunc("/ws", handler)
	log.Init("Done")
}

func (c *Service) initWebserver() {
	log.Init("Initializing Webserver")
	fs := http.FileServer(http.Dir(c.cfg.Directory))
	http.Handle("/", fs)
	log.Init("Done")
}

func (c *Service) onClientArrived(d eventual2go.Data) {
	cli := d.(*client)
	c.clients[cli.UUID()] = cli
	c.AddFuture(clientLeft{}, cli.Leave())

	log.Infof("Client %s registered successfully", cli.UUID())

	c.printStatus()
}

func (c *Service) onClientLeft(d eventual2go.Data) {
	id := d.(uuid.UUID)
	delete(c.clients, id)
	log.Infof("Client %s left", id)
	c.printStatus()
}

func (c *Service) printStatus() {
	log.Debugf("Total Clients Registered: %d", len(c.clients))
}
