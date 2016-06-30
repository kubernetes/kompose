package project

import (
	"bytes"

	"github.com/Sirupsen/logrus"
)

var (
	infoEvents = map[EventType]bool{
		EventProjectDeleteDone:   true,
		EventProjectDeleteStart:  true,
		EventProjectDownDone:     true,
		EventProjectDownStart:    true,
		EventProjectRestartDone:  true,
		EventProjectRestartStart: true,
		EventProjectUpDone:       true,
		EventProjectUpStart:      true,
		EventServiceDeleteStart:  true,
		EventServiceDelete:       true,
		EventServiceDownStart:    true,
		EventServiceDown:         true,
		EventServiceRestartStart: true,
		EventServiceRestart:      true,
		EventServiceUpStart:      true,
		EventServiceUp:           true,
	}
)

type defaultListener struct {
	project    *Project
	listenChan chan Event
	upCount    int
}

// NewDefaultListener create a default listener for the specified project.
func NewDefaultListener(p *Project) chan<- Event {
	l := defaultListener{
		listenChan: make(chan Event),
		project:    p,
	}
	go l.start()
	return l.listenChan
}

func (d *defaultListener) start() {
	for event := range d.listenChan {
		buffer := bytes.NewBuffer(nil)
		if event.Data != nil {
			for k, v := range event.Data {
				if buffer.Len() > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(k)
				buffer.WriteString("=")
				buffer.WriteString(v)
			}
		}

		if event.EventType == EventServiceUp {
			d.upCount++
		}

		logf := logrus.Debugf

		if infoEvents[event.EventType] {
			logf = logrus.Infof
		}

		if event.ServiceName == "" {
			logf("Project [%s]: %s %s", d.project.Name, event.EventType, buffer.Bytes())
		} else {
			logf("[%d/%d] [%s]: %s %s", d.upCount, len(d.project.Configs), event.ServiceName, event.EventType, buffer.Bytes())
		}
	}
}
