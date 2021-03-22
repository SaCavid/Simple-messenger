package main

import (
	"github.com/SaCavid/Simple-messenger/models"
	"github.com/SaCavid/Simple-messenger/routes"
	"github.com/SaCavid/Simple-messenger/service"
	//"./models"
	//"./routes"
	//"./service"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {

	log.SetFlags(log.Lshortfile)

	err := godotenv.Load(".env")
	if err != nil {
		// unable to connect to database. Quit app
		log.Fatal("Failed to load env! ", err)
	}
	bufferSize := os.Getenv("DEFAULTBUFFER")

	size, err := strconv.ParseUint(bufferSize, 10, 64)
	if err != nil {
		log.Println(err)
		size = 1024
	}

	srv := &service.Server{
		Port:            0,
		LoginChan:       make(chan *service.User, size),
		LogoutChan:      make(chan string, size),
		Clients:         make(map[string]chan models.Message, 0),
		DefaultDeadline: time.Second * 120,
	}

	go srv.Connections()
	go srv.TlsServer(os.Getenv("TLSPORT"))
	go srv.TcpServer(os.Getenv("TCPPORT"))

	go Monitor(srv)
	routes.Route(srv)
}

func Monitor(srv *service.Server) {
	var r runtime.MemStats
	for {
		runtime.ReadMemStats(&r)

		time.Sleep(3 * time.Second) // r.Mallocs-r.Frees,
		log.Println("System goroutines:", runtime.NumGoroutine()-int(srv.ReceiverRoutine)-int(srv.TransmitterRoutine), "Rx", srv.ReceiverRoutine, "Tx", srv.TransmitterRoutine, "Connected users:", len(srv.Clients), "Send messages:", srv.SendMessages, "Received Messages:", srv.ReceivedMessages, "Lost packages: ", srv.ReceivedMessages-srv.SendMessages)
	}
}
