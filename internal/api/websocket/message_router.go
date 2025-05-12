package websocket

import (
	"errors"
)

type MessageHandler func(payload interface{}) error

type Router struct {
	handlers map[MessageType]MessageHandler
}

func NewRouter() *Router {
	return &Router{handlers: make(map[MessageType]MessageHandler)}
}

func (r *Router) Register(msgType MessageType, handler MessageHandler) {
	r.handlers[msgType] = handler
}

func (r *Router) Route(msgType MessageType, payload interface{}) error {
	h, ok := r.handlers[msgType]
	if !ok {
		return errors.New("no handler for message type")
	}
	return h(payload)
}
