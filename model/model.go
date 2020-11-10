package model

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen

import (
	"time"
)

type Probe struct {
	Id            uint64 `json:"id"`
	I2cDeviceID   string `json:"i2cdeviceid"`
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	SendFrequency string `json:"sendfrequency,omitempty"`
}

type TemperatureRecord struct {
	Id               uint64    `json:"id"`
	Probe            *Probe    `json:"probe"`
	Timestamp        time.Time `json:"timestamp"`
	TemperatureHot   float64   `json:"temperaturehot"`
	TemperatureCold  float64   `json:"temperaturecold"`
	TemperatureDelta float64   `json:"temperaturedelta"`
}
