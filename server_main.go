package main

import (
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	go h.run()
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

	c := &connection{send: make(chan []byte, 256), ws: ws}

	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()

}
