package main

import (
	"log"
)

type hub struct {
	// Registered connections.
	connections []*connection

	// User-connection map
	connmap map[string]*connection

	// Inbound messages from the connections.
	broadcast chan *message

	// Register requests from the connections.
	init chan *connection

	// Username channel
	userchan chan *userreq

	// Unregister channel
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan *message),
	init:        make(chan *connection),
	connections: make([]*connection, 5),
	connmap:     make(map[string]*connection),
	userchan:    make(chan *userreq),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.init:
			log.Println("Adding connection to the array")
			h.connections = append(h.connections, c)
		case u := <-h.userchan:
			h.connmap[u.username] = u.conn
		case m := <-h.broadcast:
			// convert string to byte array
			bytearr := []byte(m.message)
			h.connmap[m.target_user].send <- bytearr
		case r := <-h.unregister:
			// iterating connmap and deleting connection from it
			for key, value := range h.connmap {
				if value == r {
					delete(h.connmap, key)
					close(r.send)
				}
			}

		}
	}
}
