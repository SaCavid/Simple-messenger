package service

import (
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

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		err = conn.SetReadDeadline(time.Now().Add(time.Second * 120))
		if err != nil {
			log.Println(err)
			return
		}

		go srv.Receiver(conn)
	}
}
