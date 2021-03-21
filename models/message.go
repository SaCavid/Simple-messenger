package models

import "errors"

type Message struct {
	From   string
	To     string
	Data   string
	Status bool // status of channel
}

func NewMessage(from, to, data string) *Message {
	return &Message{
		From: from,
		To:   to,
		Data: data,
	}
}

func (msg *Message) ValidateMessage() error {
	if msg.From == "" {
		return errors.New("sender cant be null")
	}

	if msg.To == "" {
		return errors.New("receiver cant be null")
	}

	return nil
}
