package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/tbellembois/goi2c/model"
	"honnef.co/go/js/dom/v2"
)

var (
	Win dom.Window
	Doc dom.Document

	signal chan (int)

	i2cDeviceIdMap map[string]bool // keep the connected sockets
)

func keepAlive() {
	for {
		<-signal
	}
}

func init() {
	signal = make(chan int)
	i2cDeviceIdMap = make(map[string]bool)

	Win = dom.GetWindow()
	Doc = Win.Document()

}

func OnClose(this js.Value, args []js.Value) interface{} {

	println("Ws connection closed.")
	return nil

}

func OnOpen(this js.Value, args []js.Value) interface{} {

	println("Ws connection open.")
	return nil

}

func OnMessage(this js.Value, args []js.Value) interface{} {

	if len(args) > 0 {

		var (
			temperatureRecord *model.TemperatureRecord
			ok                bool
			err               error
		)

		jsTemperatureRecord := args[0].Get("data").String()

		temperatureRecord = &model.TemperatureRecord{}
		if err = json.Unmarshal([]byte(jsTemperatureRecord), temperatureRecord); err != nil {
			println(err)
		}

		if _, ok = i2cDeviceIdMap[temperatureRecord.Probe.I2cDeviceID]; !ok {
			i2cDeviceIdMap[temperatureRecord.Probe.I2cDeviceID] = true
			createRecord(temperatureRecord)
		} else {
			updateRecord(temperatureRecord)
		}

	}

	return nil

}

func createRecord(t *model.TemperatureRecord) {

	//println("creating ", t.Probe.I2cDeviceID)

	colDiv := Doc.CreateElement("div").(*dom.HTMLDivElement)
	colDiv.Class().SetString(strings.Join([]string{"col-sm-4", "mt-1"}, " "))

	cardDiv := Doc.CreateElement("div").(*dom.HTMLDivElement)
	cardDiv.Class().SetString(strings.Join([]string{"card", "text-center", "mt-sm-3"}, " "))
	cardDiv.Style().SetProperty("width", "18rem", "1")

	cardBodyDiv := Doc.CreateElement("div").(*dom.HTMLDivElement)
	cardBodyDiv.Class().SetString("card-body")

	spanRefresh := Doc.CreateElement("span").(*dom.HTMLSpanElement)
	spanRefresh.Class().SetString(strings.Join([]string{"mdi", "mdi-refresh"}, " "))
	spanRefresh.SetTextContent(t.Probe.SendFrequency)

	h5Title := Doc.CreateElement("h5").(*dom.HTMLHeadingElement)
	h5Title.Class().SetString("card-title")
	h5Title.SetID(fmt.Sprintf("title%s", t.Probe.I2cDeviceID))
	h5Title.SetTextContent(t.Probe.I2cDeviceID)

	h6Subtitle := Doc.CreateElement("h5").(*dom.HTMLHeadingElement)
	h6Subtitle.Class().SetString(strings.Join([]string{"card-subtitle", "mb-2", "text-muted"}, " "))
	h6Subtitle.SetID(fmt.Sprintf("description%s", t.Probe.I2cDeviceID))
	h6Subtitle.SetTextContent(t.Probe.I2cDeviceID)

	pTimestamp := Doc.CreateElement("p").(*dom.HTMLParagraphElement)
	pTimestamp.Class().SetString("font-italic")
	pTimestamp.SetID(fmt.Sprintf("timestamp%s", t.Probe.I2cDeviceID))
	pTimestamp.SetTextContent(t.Timestamp.Format("Jan 2 15:04:05"))

	pTemperatureHot := Doc.CreateElement("p").(*dom.HTMLParagraphElement)
	pTemperatureHot.Class().SetString("h1")
	pTemperatureHot.SetID(fmt.Sprintf("temperaturehot%s", t.Probe.I2cDeviceID))
	pTemperatureHot.SetTextContent(fmt.Sprintf("%v°C", t.TemperatureHot))

	divActions := Doc.CreateElement("div").(*dom.HTMLDivElement)

	buttonConfig := Doc.CreateElement("button").(*dom.HTMLButtonElement)
	buttonConfig.Class().SetString(strings.Join([]string{"btn", "btn-primary", "mb-2", "mr-2"}, " "))
	buttonConfig.SetTitle("configure probe")
	iconConfig := Doc.CreateElement("span").(*dom.HTMLSpanElement)
	iconConfig.Class().SetString(strings.Join([]string{"mdi", "mdi-cog-outline"}, " "))
	buttonConfig.AppendChild(iconConfig)

	buttonChart := Doc.CreateElement("button").(*dom.HTMLButtonElement)
	buttonChart.Class().SetString(strings.Join([]string{"btn", "btn-primary", "mb-2", "mr-2"}, " "))
	buttonChart.SetTitle("view chart")
	iconChart := Doc.CreateElement("span").(*dom.HTMLSpanElement)
	iconChart.Class().SetString(strings.Join([]string{"mdi", "mdi-chart-line"}, " "))
	buttonChart.AppendChild(iconChart)

	buttonExport := Doc.CreateElement("button").(*dom.HTMLButtonElement)
	buttonExport.Class().SetString(strings.Join([]string{"btn", "btn-primary", "mb-2", "mr-2"}, " "))
	buttonExport.SetTitle("export data")
	iconExport := Doc.CreateElement("span").(*dom.HTMLSpanElement)
	iconExport.Class().SetString(strings.Join([]string{"mdi", "mdi-export"}, " "))
	buttonExport.AppendChild(iconExport)

	colDiv.AppendChild(cardDiv)

	cardBodyDiv.AppendChild(h5Title)
	cardBodyDiv.AppendChild(h6Subtitle)
	cardBodyDiv.AppendChild(spanRefresh)
	cardBodyDiv.AppendChild(pTimestamp)
	cardBodyDiv.AppendChild(pTemperatureHot)

	divActions.AppendChild(buttonConfig)
	divActions.AppendChild(buttonChart)
	divActions.AppendChild(buttonExport)

	cardDiv.AppendChild(cardBodyDiv)
	cardDiv.AppendChild(divActions)

	Doc.GetElementByID("temperatureRecord").AppendChild(colDiv)

}

func updateRecord(t *model.TemperatureRecord) {

	//println("updating ", t.Probe.I2cDeviceID)

	Doc.GetElementByID(fmt.Sprintf("temperaturehot%s", t.Probe.I2cDeviceID)).SetTextContent(fmt.Sprintf("%v°C", t.TemperatureHot))
	Doc.GetElementByID(fmt.Sprintf("timestamp%s", t.Probe.I2cDeviceID)).SetTextContent(t.Timestamp.Format("Jan 2 15:04:05"))

	js.Global().Get("Hightlight").Invoke(fmt.Sprintf("#temperaturehot%s", t.Probe.I2cDeviceID))

}

func main() {

	js.Global().Set("OnMessage", js.FuncOf(OnMessage))
	js.Global().Set("OnOpen", js.FuncOf(OnOpen))
	js.Global().Set("OnClose", js.FuncOf(OnClose))

	js.Global().Get("DOMContentLoaded").Invoke()

	keepAlive()

}
