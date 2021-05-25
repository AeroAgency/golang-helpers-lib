package event

import (
	"fmt"
)

// диспетчер событий
type Dispatcher struct {
	jobs   chan job
	events map[Name]Listener
}

// конструктор
func NewDispatcher() *Dispatcher {
	d := &Dispatcher{
		jobs:   make(chan job),
		events: make(map[Name]Listener),
	}

	go d.consume()

	return d
}

// регистрация событий и слушателей событий
func (d *Dispatcher) Register(listener Listener, names ...Name) error {
	for _, name := range names {
		if _, ok := d.events[name]; ok {
			return fmt.Errorf("the '%s' event is already registered", name)
		}

		d.events[name] = listener
	}

	return nil
}

// отправка события
func (d *Dispatcher) Dispatch(name Name, event interface{}) error {
	if _, ok := d.events[name]; !ok {
		return fmt.Errorf("the '%s' event is not registered", name)
	}

	d.jobs <- job{eventName: name, eventType: event}

	return nil
}

func (d *Dispatcher) consume() {
	for job := range d.jobs {
		d.events[job.eventName].Listen(job.eventType)
	}
}
