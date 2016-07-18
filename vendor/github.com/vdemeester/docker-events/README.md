# docker-events
[![GoDoc](https://godoc.org/github.com/vdemeester/docker-events?status.png)](https://godoc.org/github.com/vdemeester/docker-events)
[![Build Status](https://travis-ci.org/vdemeester/docker-events.svg?branch=master)](https://travis-ci.org/vdemeester/docker-events)
[![Go Report Card](https://goreportcard.com/badge/github.com/vdemeester/docker-events)](https://goreportcard.com/report/github.com/vdemeester/docker-events)
[![License](https://img.shields.io/github/license/vdemeester/docker-events.svg)]()

A really small library with the intent to ease the use of `Events`
method of `engine-api`.

## Usage

It should be pretty straighforward to use :

```go
import "events"

// […]

cli, err := client.NewEnvClient()
if err != nil {
    // Do something..
}

cxt, cancel := context.WithCancel(context.Background())
// Call cancel() to get out of the monitor

errChan := events.Monitor(ctx, cli, types.EventsOptions{}, func(event eventtypes.Message) {
    fmt.Printf("%v\n", event)
})

if err := <-errChan; err != nil {
    // Do something
}
```

It's also possible to do a little more advanced stuff using
`EventHandler` :

```go
import "events"

// […]

cli, err := client.NewEnvClient()
if err != nil {
    // Do something..
}

// Setup the event handler
eventHandler := events.NewHandler(events.ByAction)
eventHandler.Handle("create", func(m eventtypes.Message) {
    // Do something in case of create message
})

stoppedOrDead := func(m eventtypes.Message) {
    // Do something in case of stop or die message as it might be the
    // same way to react.
}

eventHandler.Handle("die", stoppedOrDead)
eventHandler.Handle("stop", stoppedOrDead)

// The other type of message will be discard.

// Filter the events we wams so receive
filters := filters.NewArgs()
filters.Add("type", "container")
options := types.EventsOptions{
    Filters: filters,
}

cxt, cancel := context.WithCancel(context.Background())
// Call cancel() to get out of the monitor

errChan := events.MonitorWithHandler(ctx, cli, options, eventHandler)

if err := <-errChan; err != nil {
    // Do something
}
```
