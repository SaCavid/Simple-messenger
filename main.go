package main

import (
	"./routes"
	"./service"
	"github.com/joho/godotenv"
	"log"
	"net"
	"os"
)

func main() {

	log.SetFlags(log.Lshortfile)

	err := godotenv.Load(".env")
	if err != nil {
		// unable to connect to database. Quit app
		log.Fatal("Failed to load env! ", err)
	}

	srv := &service.Server{
		Port:    0,
		Clients: make(map[string]net.Conn, 0),
	}

	go srv.TlsServer(os.Getenv("TLSPORT"))
	go srv.TcpServer(os.Getenv("TCPPORT"))

	routes.Route(srv)
}
