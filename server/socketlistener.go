package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"

	"github.com/tbellembois/goi2c/model"
)

func handleConnection(conn net.Conn) {

	var (
		buffer      []byte
		err         error
		temperature model.TemperatureRecord
	)

	if buffer, err = bufio.NewReader(conn).ReadBytes(delimiter); err != nil {
		conn.Close()
		return
	}

	if len(buffer) > 1 {

		if err = json.Unmarshal(buffer[:len(buffer)-1], &temperature); err != nil {
			log.Fatal(err)
		}

		// Sending to the channel for ws consumption.
		temperatureRecordChannel <- buffer[:len(buffer)-1]

		// Database save.
		if _, err = boxTemperatureRecord.Put(&temperature); err != nil {
			log.Fatal(err)
		}

	}
	handleConnection(conn)

}

func runSocketListener() {

	go func() {

		// Opening socket.
		log.Printf("socket listening to %s", socketAddress)
		if serverConn, err = net.Listen("tcp", socketAddress); err != nil {
			log.Fatal(err)
		}
		defer serverConn.Close()

		for {

			if netConn, err = serverConn.Accept(); err != nil {
				log.Fatal(err)
			}

			log.Printf("client %s connected", netConn.RemoteAddr().String())

			go handleConnection(netConn)
		}
	}()

}
