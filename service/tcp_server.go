package service

import (
	"../models"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

func (srv *Server) TcpServer(addr string) {

	l, err := net.Listen("tcp", "localhost:"+addr)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := l.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("started server on " + addr)

	go ClientNoTls(addr, 3)
	go ClientNoTls(addr, 4)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go Receiver(srv, conn)
	}
}

func ClientNoTls(addr string, i int) {

	time.Sleep(5 * time.Second)
	conn, err := net.Dial("tcp", "localhost:"+addr)
	if err != nil {
		log.Println(err.Error(), " ", i)
		return
	}

	m := models.NewMessage(fmt.Sprintf("tcpUser%d", i*2), fmt.Sprintf("tcpUser%d", i), "Looking for new solution", nil)

	d, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = conn.Write(d)
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer func() {

		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	go func() {
		for {
			m := models.Message{}
			d := json.NewDecoder(conn)

			err := d.Decode(&m)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("User: ", fmt.Sprintf("tcpUser%d", i), "Message from user:", m.From, "Data: ", m.Data)
		}
	}()

	for {

		m := models.Message{
			From: fmt.Sprintf("tcpUser%d", i),
			To:   fmt.Sprintf("tcpUser%d", 1),
			Data: "Looking for new solution",
		}

		if m.From == m.To {
			continue
		}

		d, err := json.Marshal(m)
		if err != nil {
			log.Println(err)
			break
		}

		_, err = conn.Write(d)
		if err != nil {
			log.Println(err.Error())
			break
		}

		time.Sleep(3 * time.Second)
	}

	log.Println("finished tls client")
}