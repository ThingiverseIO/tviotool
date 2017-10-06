package web

import (
	"net/http"

	"github.com/ThingiverseIO/logger"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/gorilla/websocket"
	"github.com/joernweissenborn/eventual2go"
)

var log = logger.New("TVIO WEB SERVER")

type clientArrived struct{}
type clientLeft struct{}

type Core struct {
	*eventual2go.Reactor
	cfg     *Config
	clients map[uuid.UUID]*client
}

func New(cfg *Config) (c *Core) {
	log.Debug("Debug Mode Enabled")
	log.Init("Starting up")
	c = &Core{
		Reactor: eventual2go.NewReactor(),
		cfg:     cfg,
		clients: map[uuid.UUID]*client{},
	}

	c.React(clientArrived{}, c.onClientArrived)
	c.React(clientLeft{}, c.onClientLeft)

	if !cfg.NoDir {
		c.initWebserver()
	}
	c.initWebSocketListener()
	return
}

func (c *Core) Serve() (err error) {
	log.Infof("Webserver starting at %s", c.cfg.Address())
	err = http.ListenAndServe(c.cfg.Address(), nil)
	if err != nil {
		log.Error("Error Starting Webserver: ", err)
	}
	return
}

func (c *Core) initWebSocketListener() {
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

func (c *Core) initWebserver() {
	log.Init("Initializing Webserver")
	fs := http.FileServer(http.Dir(c.cfg.Directory))
	http.Handle("/", fs)
	log.Init("Done")
}

func (c *Core) onClientArrived(d eventual2go.Data) {
	cli := d.(*client)
	c.clients[cli.UUID()] = cli
	c.AddFuture(clientLeft{}, cli.Leave())

	log.Infof("Client %s registered successfully", cli.UUID())

	c.printStatus()
}

func (c *Core) onClientLeft(d eventual2go.Data) {
	id := d.(uuid.UUID)
	delete(c.clients, id)
	log.Infof("Client %s left", id)
	c.printStatus()
}

func (c *Core) printStatus() {
	log.Debugf("Total Clients Registered: %d", len(c.clients))
}
