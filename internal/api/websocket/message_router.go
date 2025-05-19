package websocket

import (
	"errors"
)

// MessageHandler is a function type that processes a message payload and returns an error if processing fails.
// It's used by the router to handle different types of WebSocket messages.
type MessageHandler func(payload interface{}) error

// Router manages WebSocket message routing by mapping message types to their respective handlers.
// It provides a way to register handlers for specific message types and route incoming messages accordingly.
type Router struct {
	handlers map[MessageType]MessageHandler
}

// NewRouter creates a new message router with an empty handler map.
// The router is used to direct incoming WebSocket messages to the appropriate handler based on their type.
func NewRouter() *Router {
	return &Router{handlers: make(map[MessageType]MessageHandler)}
}

// Register associates a message handler with a specific message type.
// When a message of this type is received, the router will invoke the registered handler.
func (r *Router) Register(msgType MessageType, handler MessageHandler) {
	r.handlers[msgType] = handler
}

// Route processes an incoming message by invoking the handler associated with its type.
// Returns an error if no handler is registered for the given message type or if the handler fails.
func (r *Router) Route(msgType MessageType, payload interface{}) error {
	h, ok := r.handlers[msgType]
	if !ok {
		return errors.New("no handler for message type")
	}
	return h(payload)
}
