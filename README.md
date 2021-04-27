# BIORCELL 3D temperature probe manager

Temperature temperature probe for [BIORCELL 3D](https://biorcell-3d.com/).

Build with a [MCP9600](https://shop.pimoroni.com/products/mcp9600-thermocouple-amplifier-breakout) thermocouple amplifier and a (RaspberryPi)[https://www.raspberrypi.org/].

Written in [Golang](https://golang.org/).

## Client

`client/collectd-probe`: Daemon gathering MCP9600 informations and sending them to the collectd server.

Compile for RaspberryPi with:
```bash
    env GOOS=linux GOARCH=arm GOARM=5 go build .
```

## Server

`server/collectd`: Collectd server configuration.

Uses 3 plugins:
- network: listening for collectd-probe informations
- rrdtool: writing informations in the rrd format (used by Facette)
- csv: writing informations in the csv format (used by the notifier)

```bash
    // configure collectd.conf
    collectd -C ./collectd.conf
```

`server/facette`: Facette graph frontend configuration (as a Docker container) using rrd as unique provider.

```bash
    // configure etc/facette.yaml
    // configure docker-compose.yml
    docker-compose up -d
```

`server/notifier`: Daemon reading the csv informations and sending notifications. Use the Facette API to retrieve probe names. Min, Max temperature and notification threshold are given as arguments.

```bash
    // mandatory argument format:
    // probeName:minTemp:maxTemp:thresholdRate
    // - name must exactly match a Facette source
    // - minTemp, maxTemp: int or float
    // - thresholdRate: a Golang time.Duration
    notifier probe-fridge1:-10.5:-2:10m
    // for additionnal flags
    notifier --help 
```

## Notes

The notifier should be removed in the future to use the built in collectd notification system. The collectd Golang library used by the collect-probe does not send notifications yet.
