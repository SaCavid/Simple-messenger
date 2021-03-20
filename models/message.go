package models

import "errors"

type Message struct {
	From  string
	To    string
	Data  string
	Users []string
}

func NewMessage(from, to, data string, users []string) *Message {
	return &Message{
		From:  from,
		To:    to,
		Data:  data,
		Users: users,
	}
}

func (msg *Message) ValidateMessage() error {
	if msg.From == "" {
		return errors.New("sender cant be null")
	}

	if msg.To == "" {
		return errors.New("receiver cant be null")
	}

	if msg.Data == "" {
		return errors.New("data is null nothing to return")
	}

	return nil
}
