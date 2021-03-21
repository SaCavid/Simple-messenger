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

	l, err := net.Listen("tcp", ":"+addr)
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

		go srv.Receiver(conn)
	}
}

func ClientNoTls(addr string, i int) {

	time.Sleep(5 * time.Second)
	conn, err := net.Dial("tcp", ":"+addr)
	if err != nil {
		log.Println(err.Error(), " ", i)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
		log.Println("finished tls client")
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

}
