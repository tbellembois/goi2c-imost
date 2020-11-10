//go:generate rice embed-go
package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/objectbox/objectbox-go/objectbox"
	"github.com/tbellembois/goi2c/model"
)

const (
	delimiter byte = 0xF // socket data delimiter
)

// Env is used to insert data to the http server template
type Env struct {
	ServerAddress string
	Base64Logo    string
}

var (
	flagSocketAddress string
	socketAddress     string
	flagSocketPort    int
	socketPort        string
	flagServerAddress string
	serverAddress     string
	flagServerPort    int
	serverPort        string

	serverConn net.Listener
	netConn    net.Conn
	err        error

	ws                       *websocket.Conn
	wsClients                map[*websocket.Conn]bool
	temperatureRecordChannel chan ([]byte)

	objectBox            *objectbox.ObjectBox
	boxTemperatureRecord *model.TemperatureRecordBox
	boxProbe             *model.ProbeBox
)

func init() {

	flag.StringVar(&flagSocketAddress, "socketAddress", "127.0.0.1", "socket address (default 127.0.0.1)")
	flag.IntVar(&flagSocketPort, "socketPort", 8080, "socket port (default 8080)")
	flag.StringVar(&flagServerAddress, "serverAddress", "127.0.0.1", "server address (default 127.0.0.1)")
	flag.IntVar(&flagServerPort, "serverPort", 8081, "server port (default 8081)")
	flag.Parse()

	wsClients = make(map[*websocket.Conn]bool)
	temperatureRecordChannel = make(chan []byte)

}

func main() {

	// Initializing variables.
	socketPort = strconv.Itoa(flagSocketPort)
	serverPort = strconv.Itoa(flagServerPort)
	socketAddress = strings.Join([]string{flagSocketAddress, socketPort}, ":")
	serverAddress = strings.Join([]string{flagServerAddress, serverPort}, ":")

	// Initializing objectbox.
	if objectBox, err = objectbox.NewBuilder().Model(model.ObjectBoxModel()).Build(); err != nil {
		log.Fatal(err)
	}
	defer objectBox.Close()

	// Loading object boxes.
	boxTemperatureRecord = model.BoxForTemperatureRecord(objectBox)
	boxProbe = model.BoxForProbe(objectBox)

	runSocketListener()
	runHTTPServer()

}
