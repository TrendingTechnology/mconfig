package store

import "github.com/mhchlib/mconfig/core/event"

type Consumer struct {
}

func newConsumer() *Consumer {
	return &Consumer{}
}

func (consumer Consumer) AddEvent(e *event.Event) error {
	err := event.AddEvent(e)
	return err
}
