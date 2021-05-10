package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hpcloud/tail"
)

type ProbeConfig struct {
	MinTemp        float64
	MaxTemp        float64
	LastAlert      time.Time
	AlertThreshold time.Duration
}

var (
	flagFacetteServerAddress string
	flagCollectdCSVDataDir   string
	flagMailServerAddress    string
	flagMailServerPort       string
	flagMailServerFrom       string
	flagMailServerTo         string
	probeConfigs             map[string]*ProbeConfig
	emailRegex               *regexp.Regexp

	sources []string
	tails   []*tail.Tail
)

func init() {

	flag.StringVar(&flagFacetteServerAddress, "facetteServerAddress", "http://127.0.0.1:12003", "server address (default http://127.0.0.1:12003)")
	flag.StringVar(&flagCollectdCSVDataDir, "collectdCSVDataDir", "/var/lib/collectd/csv", "collectd CSV plugin DataDir (default /var/lib/collectd/csv)")
	flag.StringVar(&flagMailServerAddress, "mailServerAddress", "localhost", "mail server address (default localhost)")
	flag.StringVar(&flagMailServerPort, "mailServerPort", "25", "mail server port (default 25)")
	flag.StringVar(&flagMailServerFrom, "mailServerFrom", "", "mail server from (default '')")
	flag.StringVar(&flagMailServerTo, "mailServerTo", "", "mail server to (default '')")
	flag.Parse()

	// Checking mandatory arguments.
	if len(flag.Args()) == 0 {
		fmt.Println("missing probe range arguments as probeName:minValue:maxValue:alertThreshold")
		os.Exit(1)
	}

	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	// Validating from.
	if !isEmailValid(flagMailServerFrom) {
		fmt.Printf("invalid from email syntax for %s", flagMailServerFrom)
		os.Exit(1)
	}
	// Validating to.
	for _, email := range strings.Split(flagMailServerTo, ",") {
		if !isEmailValid(email) {
			fmt.Printf("invalid from email syntax for %s", email)
			os.Exit(1)
		}
	}

	fmt.Printf("- facette server URL %s\n", flagFacetteServerAddress)
	fmt.Printf("- collectd CSV datadir %s\n", flagCollectdCSVDataDir)

	// We do not check the duration syntax. ParseDuration will do it later.
	probeConfigRegex := regexp.MustCompile(`^(?P<name>.+):(?P<min>-?\d+(\.\d+)?):(?P<max>-?\d+(\.\d+)?):(?P<duration>.+)$`)
	probeConfigRegexNames := probeConfigRegex.SubexpNames()

	probeConfigs = make(map[string]*ProbeConfig)
	for _, arg := range flag.Args() {

		result := probeConfigRegex.FindAllStringSubmatch(arg, -1)
		if result == nil {
			fmt.Println("probe range invalid syntax")
			os.Exit(1)
		}

		m := map[string]string{}
		for i, n := range result[0] {
			m[probeConfigRegexNames[i]] = n
		}

		var (
			min, max float64
			duration time.Duration
			err      error
		)
		if min, err = strconv.ParseFloat(m["min"], 64); err != nil {
			panic(err)
		}
		if max, err = strconv.ParseFloat(m["max"], 64); err != nil {
			panic(err)
		}
		if duration, err = time.ParseDuration(m["duration"]); err != nil {
			panic(err)
		}

		probeConfigs[m["name"]] = &ProbeConfig{
			MinTemp:        min,
			MaxTemp:        max,
			AlertThreshold: duration,
		}

	}

}

func isEmailValid(email string) bool {

	if len(email) < 3 && len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)

}

func sendMail(text string) {

	var (
		e       error
		client  *smtp.Client
		smtpw   io.WriteCloser
		message string
	)

	to := strings.Split(flagMailServerTo, ",")
	subject := text
	body := text

	message += fmt.Sprintf("From: %s\r\n", flagMailServerFrom)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(to, ","))
	message += "Content-Type: text/plain; charset=utf-8\r\n"
	message += fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n" + body + "\r\n"

	if client, e = smtp.Dial(flagMailServerAddress + ":" + flagMailServerPort); e != nil {
		fmt.Println(e)
		return
	}
	defer client.Close()

	if e = client.Mail(flagMailServerFrom); e != nil {
		fmt.Println(e)
		return
	}

	for _, t := range to {
		if e = client.Rcpt(t); e != nil {
			fmt.Println(e)
			return
		}
	}

	if smtpw, e = client.Data(); e != nil {
		fmt.Println(e)
		return
	}

	buf := bytes.NewBufferString(message)
	if _, e = buf.WriteTo(smtpw); e != nil {
		fmt.Println(e)
		return
	}
	smtpw.Close()

	_ = client.Quit()

}

func sendAlert(source string, temp float64, logTime time.Time) {

	lastAlert := probeConfigs[source].LastAlert
	alertThreshold := probeConfigs[source].AlertThreshold
	now := time.Now()

	if lastAlert.Add(alertThreshold).Before(now) {

		msg := fmt.Sprintf("Alert at %s for %s: temperature %f (expected %f < temp < %f)\n",
			logTime.Format(time.ANSIC),
			source,
			temp,
			probeConfigs[source].MinTemp,
			probeConfigs[source].MaxTemp)
		probeConfigs[source].LastAlert = now

		fmt.Println(msg)
		sendMail(msg)

	}

}

func checkAlert(source, log string) {

	// Log format: timestamp,temperature
	// exemple: 1619172608.372,28.625000
	s := strings.Split(log, ",")

	// Leaving the nanoseconds.
	intTimestamp, err := strconv.ParseInt(strings.Split(s[0], ".")[0], 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	logTime := time.Unix(intTimestamp, 0)

	temperature, err := strconv.ParseFloat(s[1], 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	if temperature > float64(probeConfigs[source].MaxTemp) || temperature < float64(probeConfigs[source].MinTemp) {
		sendAlert(source, temperature, logTime)
	}

}

func restartTails() {

	fmt.Println("- restarting tails")
	stopTails()
	startTails()

}

func stopTails() {

	for i := range tails {

		fmt.Printf("- stopping tailing on %s\n", tails[i].Filename)
		tails[i].Lines <- tail.NewLine("STOP")

	}
	tails = nil

}

func startTails() {

	today := time.Now().Format("2006-01-02")
	for _, source := range sources {

		csvFileName := path.Join(flagCollectdCSVDataDir, source, "digitemp", fmt.Sprintf("imost_temperature-%s", today))

		go func(csvFileName, source string) {

			fmt.Printf("- opening %s for probe %s\n", csvFileName, source)
			t, err := tail.TailFile(csvFileName, tail.Config{Follow: true})
			if err != nil {
				fmt.Println(err)
				return
			}

			tails = append(tails, t)

			firstLoop := true
			for line := range t.Lines {

				// Leaving header.
				if firstLoop {
					firstLoop = false
					continue
				}

				// fmt.Println(line.Text)

				if line.Text == "STOP" {
					fmt.Printf("- received STOP tail for %s\n", source)
					t.Stop()
					t.Cleanup()
					return
				}
				checkAlert(source, line.Text)

			}

		}(csvFileName, source)

	}

}

func main() {

	// Getting Facette sources from API.
	var (
		err  error
		req  *http.Request
		res  *http.Response
		body []byte
	)
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	if req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/catalog/sources", flagFacetteServerAddress), nil); err != nil {
		panic(err)
	}
	if res, err = httpClient.Do(req); err != nil {
		panic(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		panic(err)
	}
	if err = json.Unmarshal(body, &sources); err != nil {
		panic(err)
	}

	fmt.Printf("- tailing sources %+v\n", sources)
	startTails()

	fmt.Println("- starting cron")
	cronScheduler := gocron.NewScheduler(time.UTC)
	_, err = cronScheduler.Every(1).Day().At("00:01").Do(restartTails)
	if err != nil {
		panic(err)
	}
	cronScheduler.StartBlocking()

}
