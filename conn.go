package main

import (
	// "io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"golang.org/x/net/html"
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
		_, ws_msg, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("Error occured- ", err)
			break
		}
		log.Printf("Message received - %s\n", ws_msg)
		s := string(ws_msg[:])
		msg_array := strings.Split(s, "^")
		if msg_array[0] == "0" {
			// send userreq to userchan
			user := &userreq{conn: c, username: msg_array[1], password: msg_array[2]}
			h.userchan <- user
		} else if msg_array[0] == "1" { // send to broadcast chan
			// if the string contains an url, extract it
			// This is a simple check for the sake of time
			if strings.Contains(msg_array[2], "http:") {

				// get the html content from the page
				resp, err := http.Get(msg_array[2])
				if err != nil {
					log.Println("Error occured while fetching from url- ", err)
					break
				}
				defer resp.Body.Close()
				// body, err := ioutil.ReadAll(resp.Body)
				// log.Println(string(body))
				// continue
				msgToSend := ""
				// get the first image tag
				doc, err := html.Parse(resp.Body)
				if err != nil {
					log.Println("Error occured while parsing- ", err)
					break
				}
				var f func(*html.Node)
				f = func(n *html.Node) {
					if n.Type == html.ElementNode && n.Data == "img" {
						//Found the image node !
						for _, tag_attr := range n.Attr {
							if tag_attr.Key == "src" {
								//Found the src attribute !
								msgToSend = "^" + tag_attr.Val
								return
							}
						}
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}
				f(doc)
				// construct the message
				msg := &message{target_user: msg_array[1], message: msgToSend, conn: c}

				h.broadcast <- msg
			} else {
				msg := &message{target_user: msg_array[1], message: msg_array[2], conn: c}
				h.broadcast <- msg
			}
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
