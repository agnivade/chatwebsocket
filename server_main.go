package main

import (
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"path"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", serveHome)

	http.HandleFunc("/ws", serveWs)

	log.Println("Starting server ..")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	lp := path.Join("templates", "master.html")

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		log.Printf("Error occured - ", err)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error occured - ", err)
	}

}

func serveWs(w http.ResponseWriter, r *http.Request) {

	log.Println("Received websocket request..")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			return
		}

		log.Println("Received this message - ", p)

		if err = ws.WriteMessage(messageType, p); err != nil {
			return
		}
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}

	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()

}
