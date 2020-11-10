package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"math/rand"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
	"github.com/tbellembois/goi2c/model"
	"github.com/tbellembois/goi2c/util"
)

const (
	maxFakeTemp int32 = 30  // maximum fake temperature
	minFakeTemp int32 = 10  // minimum fake temperature
	delimiter   byte  = 0xF // socket data delimiter
)

var (
	i2cConn         *i2c.I2C
	i2cDeviceID     string
	serverConn      net.Conn
	serverConnected bool
	data            []byte
	err             error

	flagSendFrequency         string
	sendFrequency             time.Duration
	flagRetryConnectFrequency string
	retryConnectFrequency     time.Duration
	flagServerAddress         string
	serverAddress             string
	flagServerPort            int
	serverPort                string
	flagI2cAddress            string
	i2cAddress                uint8
	flagI2cBus                int
	i2cBus                    int
	flagFakeData              bool   // generate fake data instead of querying the i2c probe
	flagFakeSend              bool   // print data instead of sending them
	flagFakeId                string // only if flagFakeData=true, i2c device id can be specified
)

func connectServer() {

	serverConnected = false

	log.Printf("connecting to %s", serverAddress)
	for !serverConnected {

		if serverConn, err = net.Dial("tcp", serverAddress); err == nil {
			log.Printf("connected to %s", serverAddress)
			serverConnected = true
		}

		if err != nil {
			log.Printf("-> %s", err.Error())
			log.Printf("retrying in %s seconds", flagRetryConnectFrequency)

			time.Sleep(retryConnectFrequency)
		}

	}

}

func fakeConnectServer() {

	log.Println("faking server connectiong")
	serverConnected = true

}

func init() {

	flag.StringVar(&flagSendFrequency, "sendFrequency", "60s", "data sending frequency in seconds (default 60s)")
	flag.StringVar(&flagRetryConnectFrequency, "retryConnectFrequency", "5s", "server connection retry frequency in seconds (default 5s)")
	flag.StringVar(&flagServerAddress, "serverAddress", "127.0.0.1", "server address (default 127.0.0.1)")
	flag.IntVar(&flagServerPort, "serverPort", 8080, "server port (default 8080)")
	flag.StringVar(&flagI2cAddress, "i2cAddress", "66", "i2c device adress (default 66)")
	flag.IntVar(&flagI2cBus, "i2cBus", 1, "i2c bus (default 1)")
	flag.BoolVar(&flagFakeData, "fakeData", false, "do not get data from the i2c device, simulate it instead")
	flag.StringVar(&flagFakeId, "fakeId", "string", "with fakeData, specify the i2c device id (autogenerated by default)")
	flag.BoolVar(&flagFakeSend, "fakeSend", false, "do not send data to the server, print it instead")
	flag.Parse()

	// Uncomment/comment next line to suppress/increase verbosity of output
	if err = logger.ChangePackageLogLevel("i2c", logger.InfoLevel); err != nil {
		log.Fatal(err)
	}

}

func main() {

	// Initializing variables.
	serverPort = strconv.Itoa(flagServerPort)
	serverAddress = strings.Join([]string{flagServerAddress, serverPort}, ":")
	if retryConnectFrequency, err = time.ParseDuration(flagRetryConnectFrequency); err != nil {
		log.Fatal(err)
	}
	if sendFrequency, err = time.ParseDuration(flagSendFrequency); err != nil {
		log.Fatal(err)
	}
	var i64 int64
	if i64, err = strconv.ParseInt(flagI2cAddress, 10, 8); err != nil {
		log.Fatal(err)
	}
	i2cBus = flagI2cBus
	i2cAddress = uint8(i64)

	// Waiting for a server.
	if flagFakeSend {

		fakeConnectServer()

	} else {

		connectServer()
		defer serverConn.Close()

	}

	// I2C connection.
	if flagFakeData {

		if flagFakeId == "" {
			i2cDeviceID = util.RandStringBytesMaskImprSrcUnsafe(8)
		} else {
			i2cDeviceID = flagFakeId
		}

	} else {

		log.Printf("connecting to i2c device at address %d bus %d", i2cAddress, i2cBus)
		if i2cConn, err = i2c.NewI2C(0x66, 1); err != nil {
			log.Fatal(err)
		}
		defer i2cConn.Close()

		// Getting device ID.
		if _, err = i2cConn.WriteBytes([]byte{0x20}); err != nil {
			log.Fatal(err)
		}
		if data, _, err = i2cConn.ReadRegBytes(0x20, 2); err != nil {
			log.Fatal(err)
		}
		i2cDeviceID = fmt.Sprintf("%v", data[0])
		// data[1] is the Major/Minor Revision ID.

		// Thermocouple Sensor Configuration.
		log.Println("-> setting up thermocouple sensor configuration")
		if _, err = i2cConn.WriteBytes([]byte{0x05, 0b00000111}); err != nil {
			log.Fatal(err)
		}

		// DeviceConfiguration.
		log.Println("-> setting up device configuration")
		if _, err = i2cConn.WriteBytes([]byte{0x06, 0b01111100}); err != nil {
			log.Fatal(err)
		}

	}

	log.Printf("-> i2c device ID: %s", i2cDeviceID)

	for {

		probe := &model.Probe{
			I2cDeviceID:   i2cDeviceID,
			SendFrequency: flagSendFrequency,
		}
		temperature := model.TemperatureRecord{
			Timestamp: time.Now(),
		}
		temperature.Probe = probe

		if flagFakeData {

			ranTtemperature := rand.Int31n(maxFakeTemp-minFakeTemp) + minFakeTemp
			temperature.TemperatureHot = float64(ranTtemperature)

		} else {

			//
			// Read hot command
			//
			if _, err = i2cConn.WriteBytes([]byte{0x00}); err != nil {
				log.Fatal(err)
			}
			if data, _, err = i2cConn.ReadRegBytes(0x00, 2); err != nil {
				log.Fatal(err)
			}

			// Calculating hot temperature.
			temperature.TemperatureHot = (float64(data[0]) * float64(16)) + (float64(data[1]) / float64(16))

			//
			// Read delta command
			//
			if _, err = i2cConn.WriteBytes([]byte{0x01}); err != nil {
				log.Fatal(err)
			}
			if data, _, err = i2cConn.ReadRegBytes(0x01, 2); err != nil {
				log.Fatal(err)
			}

			// Calculating delta temperature.
			temperature.TemperatureDelta = (float64(data[0]) * float64(16)) + (float64(data[1]) / float64(16))

			//
			// Read col command
			//
			if _, err = i2cConn.WriteBytes([]byte{0x02}); err != nil {
				log.Fatal(err)
			}
			if data, _, err = i2cConn.ReadRegBytes(0x02, 2); err != nil {
				log.Fatal(err)
			}

			// Calculating cold temperature.
			temperature.TemperatureCold = (float64(data[0]) * float64(16)) + (float64(data[1]) / float64(16))

		}

		var dataSent []byte
		if dataSent, err = json.Marshal(temperature); err != nil {
			log.Fatal(err)
		}

		if flagFakeSend {

			log.Printf("%v", temperature)

		} else {

			if _, err = serverConn.Write(dataSent); err != nil {
				log.Println(err)
				connectServer()
			}
			if _, err = serverConn.Write([]byte{delimiter}); err != nil {
				log.Println(err)
				connectServer()
			}

		}

		time.Sleep(sendFrequency)

	}

}
