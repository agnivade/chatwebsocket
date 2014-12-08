package main

import (
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Authenticated
	state bool
}

type userreq struct {
	conn     *connection
	username string
	password string
}

type message struct {
	conn        *connection
	target_user string
	message     string
}

func (c *connection) reader() {
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("Error occured- ", err)
			break
		}
		log.Printf("Message received - %s\n", msg)
		s := string(msg[:])
		msg_array := strings.Split(s, ":")
		if msg_array[0] == "0" {
			// send userreq to userchan
			user := &userreq{conn: c, username: msg_array[1], password: msg_array[2]}
			h.userchan <- user
		} else if msg_array[0] == "1" {
			// send to broadcast chan
			msg := &message{target_user: msg_array[1], message: msg_array[2], conn: c}
			h.broadcast <- msg
		}

	}
	log.Println("closing read websocket")
	c.ws.Close()
}

func (c *connection) writer() {
	for {
		message := <-c.send
		log.Printf("Sending message - %s\n", message)
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Error occured- ", err)
			break
		}
	}
	log.Println("closing write websocket")
	c.ws.Close()
}
