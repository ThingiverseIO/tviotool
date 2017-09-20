package web

import (
	tviomsg "github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type msgType int

const (
	NEWINPUT msgType = iota
	REMOVEINPUT
	NEWOUTPUT
	REMOVEOUTPUT
	CONNECTED
	REQUEST
	RESULT
	LISTENSTART
	LISTENSTOP
	OBSERVESTART
	OBSERVESTOP
	PROPERTYGET
	PROPERTYSET
	PROPERTYCHANGE
	PROPERTYUPDATE
)

func parseMessage(d eventual2go.Data) (msg eventual2go.Data) {
	data := d.([]byte)
	var m message
	decode(&m, data)
	return m
}

func isMessage(t msgType) messageFilter {
	return func(msg message) (is bool) {
		is = msg.Type == t
		return
	}
}

//go:generate evt2gogen -t message

type message struct {
	Type    msgType
	Payload []byte
}

func (m message) flatten() (data []byte) {
	return encode(m)
}

type newInput struct {
	Descriptor string
	UUID       uuid.UUID
	Error      string
}

type newOutput struct {
	Descriptor string
	UUID       uuid.UUID
}

type remove struct {
	UUID uuid.UUID
}

type connected struct {
	UUID      uuid.UUID
	Connected bool
}

type request struct {
	UUID      uuid.UUID
	CallType  tviomsg.CallType
	Function  string
	Parameter []byte
	ID        int //only for CALL(ALL)
}

type result struct {
	UUID      uuid.UUID
	CallType  tviomsg.CallType
	Function  string
	Parameter []byte
	ID        int //only for CALL(ALL)
}

type listenStart struct {
	UUID     uuid.UUID
	Function string
}

type listenStop struct {
	UUID     uuid.UUID
	Function string
}

type observeStart struct {
	UUID     uuid.UUID
	Property string
}

type observeStop struct {
	UUID     uuid.UUID
	Property string
}

type propertyGet struct {
	UUID     uuid.UUID
	Property string
	Value    []byte
}

type propertyChange struct {
	UUID     uuid.UUID
	Property string
	Value    []byte
}

type propertyUpdate struct {
	UUID     uuid.UUID
	Property string
	Value    []byte
}
