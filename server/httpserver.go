package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
	"github.com/tbellembois/goi2c/server/static/html"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func homeEndpoint(w http.ResponseWriter, r *http.Request) {

	env := Env{
		ServerAddress: serverAddress,
		Base64Logo:    html.Logo,
	}

	t := template.Must(template.New("index").Parse(html.Index))
	if err = t.ExecuteTemplate(w, "index", env); err != nil {
		log.Fatal(err)
	}

}

func wsRead(conn *websocket.Conn) {

	for {

		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))
		fmt.Println(string(messageType))

	}

}

func wsWrite() {

	for {

		// Waiting for a message from the socket.
		msg := <-temperatureRecordChannel

		for conn := range wsClients {

			// Transferring it into the ws connection
			if err = conn.WriteMessage(1, msg); err != nil {
				log.Println(err)
			}

		}

	}

}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {

	// Allow incoming request from a different domain.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade the connection to a WebSocket connection.
	if ws, err = upgrader.Upgrade(w, r, nil); err != nil {
		log.Fatal(err)
	}

	wsClients[ws] = true

	// Reading on each connection.
	go wsRead(ws)

}

func runHTTPServer() {

	cssRiceBox := rice.MustFindBox("static/css")
	cssFileServer := http.StripPrefix("/css/", http.FileServer(cssRiceBox.HTTPBox()))
	http.Handle("/css/", cssFileServer)

	jsRiceBox := rice.MustFindBox("static/js")
	jsFileServer := http.StripPrefix("/js/", http.FileServer(jsRiceBox.HTTPBox()))
	http.Handle("/js/", jsFileServer)

	fontRiceBox := rice.MustFindBox("static/fonts")
	fontFileServer := http.StripPrefix("/fonts/", http.FileServer(fontRiceBox.HTTPBox()))
	http.Handle("/fonts/", fontFileServer)

	staticRiceBox := rice.MustFindBox("static")
	staticFileServer := http.StripPrefix("/static/", http.FileServer(staticRiceBox.HTTPBox()))
	http.Handle("/static/", staticFileServer)

	wasmRiceBox := rice.MustFindBox("wasm")
	wasmFileServer := http.StripPrefix("/wasm/", http.FileServer(wasmRiceBox.HTTPBox()))
	http.Handle("/wasm/", wasmFileServer)

	http.HandleFunc("/", homeEndpoint)
	http.HandleFunc("/ws", wsEndpoint)

	// Writing on all connections.
	go wsWrite()

	log.Printf("http server listening to %s", serverAddress)
	if err = http.ListenAndServe(serverAddress, nil); err != nil {
		panic("error running the server")
	}

}
