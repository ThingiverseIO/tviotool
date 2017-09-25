package web

import (
	"encoding/hex"
	"sync"

	tvioconfig "github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/core"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	tviomsg "github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/gorilla/websocket"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/eventual2go/typedevents"
)

type addInput struct{}
type removeInput struct{}
type webRequestEvent struct{}
type tvioResultEvent struct{}
type tvioListenResultEvent struct{}

type client struct {
	*eventual2go.Reactor
	m        *sync.RWMutex
	in       map[uuid.UUID]core.InputCore
	conn     *websocket.Conn
	leave    *eventual2go.Completer
	messages *eventual2go.StreamController
	sender   eventual2go.ActorMessageStream
	uuid     uuid.UUID
}

func newClient(conn *websocket.Conn) (c *client) {
	c = &client{
		Reactor:  eventual2go.NewReactor(),
		m:        &sync.RWMutex{},
		in:       map[uuid.UUID]core.InputCore{},
		conn:     conn,
		leave:    eventual2go.NewCompleter(),
		messages: eventual2go.NewStreamController(),
		uuid:     uuid.New(),
	}

	messages := messageStream{c.messages.Stream().Transform(parseMessage)}

	messages.Where(isMessage(NEWINPUT)).Listen(c.onNewInput)
	messages.Where(isMessage(REMOVEINPUT)).Listen(c.onRemoveInput)
	messages.Where(isMessage(REQUEST)).Listen(c.onRequest)
	messages.Where(isMessage(LISTENSTART)).Listen(c.onStartListen)
	messages.Where(isMessage(LISTENSTOP)).Listen(c.onStopListen)
	messages.Where(isMessage(OBSERVESTART)).Listen(c.onStartObserve)
	messages.Where(isMessage(OBSERVESTOP)).Listen(c.onStopObserve)
	messages.Where(isMessage(PROPERTYGET)).Listen(c.onPropertyGet)

	c.sender, _ = eventual2go.SpawnActor(&sender{conn, c.uuid})

	go c.receive()
	return
}

func (c client) Leave() *eventual2go.Future {
	return c.leave.Future()
}

func (c client) UUID() uuid.UUID {
	return c.uuid
}

func (c *client) close() {
	c.conn.Close()

	for _, i := range c.in {
		i.Shutdown()
	}

	c.leave.Complete(c.UUID())
}

func (c *client) onNewInput(m message) {
	log.Infof("Client %s requests new input", c.UUID())
	var msg newInput
	decode(&msg, m.Payload)

	in, err := createInput(msg.Descriptor)
	if err != nil {
		log.Errorf("Client %s: failed to create input: %s", c.UUID(), err)
		msg.Error = err.Error()
	} else {
		log.Infof("Client %s: input created successfully with UUID %s", c.UUID(), in.UUID())
		msg.UUID = in.UUID()
	}

	c.send(message{
		Type:    NEWINPUT,
		Payload: encode(msg),
	})

	in.ConnectedObservable().OnChange(c.sendConnected(in.UUID()))

	in.ListenStream().Listen(c.onListenResult(in.UUID()))

	c.m.Lock()
	defer c.m.Unlock()
	c.in[in.UUID()] = in
}

func (c *client) onRemoveInput(m message) {

	var msg remove
	decode(&msg, m.Payload)

	log.Infof("Client %s: Removing input %s", c.UUID(), msg.UUID)

	c.m.Lock()
	defer c.m.Unlock()
	if in, ok := c.in[msg.UUID]; ok {
		in.Shutdown()
		delete(c.in, msg.UUID)
	} else {
		log.Errorf("Client %s: removing %s failed, input does not exist", c.UUID(), msg.UUID)
	}
}

func (c *client) onRequest(m message) {

	var msg request
	decode(&msg, m.Payload)

	log.Infof("Client %s: %s Request from input %s with ID %d", c.UUID(), msg.CallType, msg.UUID, msg.ID)

	c.m.RLock()
	defer c.m.RUnlock()
	if in, ok := c.in[msg.UUID]; ok {
		switch msg.CallType {
		case tviomsg.CALL:
			res, _, _, _ := in.Request(msg.Function, msg.CallType, msg.Parameter)
			res.Then(c.onCallResult(msg.ID))
		case tviomsg.CALLALL:
		default:
			in.Request(msg.Function, msg.CallType, msg.Parameter)
		}
	} else {
		log.Errorf("Client %s: removing %s failed, input does not exist", c.UUID(), msg.UUID)
	}
}

func (c *client) onCallResult(id int) tviomsg.ResultCompletionHandler {
	return func(r *tviomsg.Result) *tviomsg.Result {
		log.Infof("Client %s: Call result from input %s with ID %d", c.UUID(), r.Request.Input, id)

		msg := result{
			UUID:      r.Request.Input,
			CallType:  r.Request.CallType,
			Function:  r.Request.Function,
			Parameter: r.Parameter(),
			ID:        id,
		}
		c.send(message{
			Type:    RESULT,
			Payload: encode(msg),
		})

		return nil
	}
}

func (c *client) onListenResult(UUID uuid.UUID) tviomsg.ResultSubscriber {
	return func(r *tviomsg.Result) {
		log.Infof("Client %s: Listen result for input %s", c.UUID(), r.Request.Input)

		msg := result{
			UUID:      UUID,
			CallType:  r.Request.CallType,
			Function:  r.Request.Function,
			Parameter: r.Parameter(),
		}
		c.send(message{
			Type:    RESULT,
			Payload: encode(msg),
		})
	}
}

func (c *client) onStartListen(m message) {
	var msg listenStart
	decode(&msg, m.Payload)
	log.Infof("Client %s: Request to start listening to function '%s' on input %s", c.UUID(), msg.Function, msg.UUID)
	if in, ok := c.in[msg.UUID]; ok {
		if err := in.StartListen(msg.Function); err != nil {
			log.Errorf("Client %s: Error StartListen: %s", c.UUID(), msg)
		} else {
			log.Success("Started to listening.")
		}
	} else {
		log.Errorf("Client %s: Non-existing input %s", c.UUID(), msg.UUID)
	}
}

func (c *client) onStopListen(m message) {
	var msg listenStop
	decode(&msg, m.Payload)
	log.Infof("Client %s: Request to stop listening to function '%s' on input %s", c.UUID(), msg.Function, msg.UUID)
	if in, ok := c.in[msg.UUID]; ok {
		if err := in.StopListen(msg.Function); err != nil {
			log.Errorf("Client %s: Error StopListen: %s", c.UUID(), msg)
		} else {
			log.Success("Started to listening.")
		}
	} else {
		log.Errorf("Client %s: Non-existing input %s", c.UUID(), msg.UUID)
	}
}

func (c *client) onStartObserve(m message) {
	var msg observeStart
	decode(&msg, m.Payload)
	log.Infof("Client %s: Request to start observing property '%s' on input %s", c.UUID(), msg.Property, msg.UUID)
	if in, ok := c.in[msg.UUID]; ok {
		if err := in.StartObservation(msg.Property); err != nil {
			log.Errorf("Client %s: Error StartObserve: %s", c.UUID(), msg)
		} else {
			o, _ := in.GetProperty(msg.Property)
			o.OnChange(c.onPropertyChange(in.UUID(), msg.Property))
			log.Success("Started observing.")
		}
	} else {
		log.Errorf("Client %s: Non-existing input %s", c.UUID(), msg.UUID)
	}
}

func (c *client) onPropertyChange(uuid uuid.UUID, property string) eventual2go.Subscriber {
	return func(d eventual2go.Data) {
		msg := propertyChange{
			UUID:     uuid,
			Property: property,
			Value:    d.([]byte),
		}
		c.send(message{
			Type:    PROPERTYCHANGE,
			Payload: encode(msg),
		})
	}
}

func (c *client) onStopObserve(m message) {
	var msg observeStop
	decode(&msg, m.Payload)
	log.Infof("Client %s: Request to stop observing property '%s' on input %s", c.UUID(), msg.Property, msg.UUID)
	if in, ok := c.in[msg.UUID]; ok {
		if err := in.StopObservation(msg.Property); err != nil {
			log.Errorf("Client %s: Error StopObserve: %s", c.UUID(), msg)
		} else {
			log.Success("Stoped observing.")
		}
	} else {
		log.Errorf("Client %s: Non-existing input %s", c.UUID(), msg.UUID)
	}
}

func (c *client) onPropertyGet(m message) {
	var msg propertyGet
	decode(&msg, m.Payload)
	log.Infof("Client %s: Request to get property '%s' on input %s", c.UUID(), msg.Property, msg.UUID)
	if in, ok := c.in[msg.UUID]; ok {
		o, err := in.GetProperty(msg.Property)
		if err != nil {
			log.Errorf("Client %s: Error StopObserve: %s", c.UUID(), msg)
		} else {
			msg.Value = o.Value().([]byte)
			log.Success("Got Property.")
		}
		c.send(message{
			Type:    PROPERTYGET,
			Payload: encode(msg),
		})
	} else {
		log.Errorf("Client %s: Non-existing input %s", c.UUID(), msg.UUID)
	}
}

func (c *client) send(m message) {
	c.sender.Send(m)
}

func (c *client) receive() {

	defer c.close()

	for {
		_, msg, err := c.conn.ReadMessage()

		if err != nil {
			log.Debugf("Client %s: Error reading socket: %s", c.UUID(), err)
			return
		}
		log.Debugf("Client %s: New Message Received", c.UUID())
		c.messages.Add(msg)

	}

}

func (c *client) sendConnected(UUID uuid.UUID) typedevents.BoolSubscriber {
	return func(is bool) {
		c.send(message{
			Type:    CONNECTED,
			Payload: encode(connected{UUID: UUID, Connected: is}),
		})
	}
}

type sender struct {
	conn *websocket.Conn
	uuid uuid.UUID
}

func (s *sender) Init() error { return nil }

func (s *sender) OnMessage(d eventual2go.Data) {
	m := d.(message)
	log.Debugf("Client %s: sending message: \n%s", s.uuid, hex.Dump(m.flatten()))
	s.conn.WriteMessage(websocket.BinaryMessage, m.flatten())
}

func createInput(desc string) (i core.InputCore, err error) {
	d, err := descriptor.Parse(desc)
	if err != nil {
		return
	}
	cfg := tvioconfig.Configure()
	cfg.Debug = true
	tracker, provider := core.DefaultBackends()
	i, err = core.NewInputCore(d, cfg, tracker, provider...)
	return
}
