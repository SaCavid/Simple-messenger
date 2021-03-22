package service

import (
	"crypto/tls"
	"encoding/json"
	"github.com/SaCavid/Simple-messenger/models"
	//"../models"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	Port               int
	Mu                 sync.Mutex
	LoginChan          chan *User
	LogoutChan         chan string
	Clients            map[string]chan models.Message
	ReceiverRoutine    uint64
	TransmitterRoutine uint64
	SendMessages       uint64
	ReceivedMessages   uint64
	DefaultDeadline    time.Duration
}

type User struct {
	Name    string
	Channel chan models.Message
}

func (srv *Server) TlsServer(addr string) {

	cer, err := tls.X509KeyPair([]byte(csr), []byte(privateKey))
	if err != nil {
		log.Fatal(err)
	}

	cer2, err := tls.X509KeyPair([]byte(csr), []byte(privateKey))
	if err != nil {
		log.Fatal(err)
	}

	configServer := &tls.Config{Certificates: []tls.Certificate{cer}, ServerName: "Test"}
	configServer.Certificates = append(configServer.Certificates, cer2)

	l, err := tls.Listen("tcp", ":"+addr, configServer)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := l.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("started server on ", addr, ", lens of certificates", len(configServer.Certificates))

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		err = conn.SetReadDeadline(time.Now().Add(srv.DefaultDeadline))
		if err != nil {
			log.Println(err)
			return
		}

		go srv.Receiver(conn)
	}
}

func (srv *Server) Receiver(conn net.Conn) {

	logged := false
	var user string

	c := make(chan models.Message, 8)
	srv.ReceiverRoutine++
	defer func() {

		_ = conn.Close()

		srv.ReceiverRoutine--
	}()

	go srv.Transmitter(conn, c)

	for {
		m := models.Message{}
		d := json.NewDecoder(conn)

		err := d.Decode(&m)
		if err != nil {
			log.Println(err)
			return
		}

		err = m.ValidateMessage()
		if err != nil {
			log.Println(err)
			continue
		}

		if !logged {
			logged = true
			user = m.From
			srv.Login(user, c)
		} else {
			srv.Mu.Lock()

			receiver := srv.Clients[m.To]
			srv.ReceivedMessages++
			srv.Mu.Unlock()

			if receiver != nil {
				receiver <- m
			}

			err = conn.SetReadDeadline(time.Now().Add(srv.DefaultDeadline))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (srv *Server) WsReceiver(w http.ResponseWriter, r *http.Request) {

	var user string
	var logged bool

	wsConn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	srv.ReceiverRoutine++
	c := make(chan models.Message, 8)

	defer func() {

		err := wsConn.Close()
		if err != nil {
			log.Println(err)
		}

		srv.ReceiverRoutine--
	}()

	go srv.WsTransmitter(wsConn, c)

	for {
		m := models.Message{}
		err := wsConn.ReadJSON(&m)
		if err != nil {
			log.Println(err)
			return
		}
		if !logged {
			err := m.ValidateMessage()
			if err != nil {
				log.Println(err)
				return
			}

			user = m.From
			logged = true
			srv.Login(user, c)
		} else {
			srv.Mu.Lock()
			receiver := srv.Clients[m.To]
			srv.ReceivedMessages++
			srv.Mu.Unlock()
			if receiver != nil {
				receiver <- m
			}
		}
	}
}

func (srv *Server) Transmitter(conn net.Conn, c chan models.Message) {

	srv.TransmitterRoutine++
	defer func() {
		close(c)
		srv.TransmitterRoutine--
		_ = conn.Close()
	}()

	for {
		y := <-c

		if y.Status {
			return
		}

		srv.Mu.Lock()
		srv.SendMessages++
		srv.Mu.Unlock()
		d, err := json.Marshal(y)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.Write(d)
		if err != nil {
			log.Println(err)
			return
		}

		err = conn.SetDeadline(time.Now().Add(srv.DefaultDeadline))
		if err != nil {
			log.Println(err)
		}
	}
}

func (srv *Server) WsTransmitter(conn *websocket.Conn, c chan models.Message) {

	srv.TransmitterRoutine++
	defer func() {
		close(c)
		srv.TransmitterRoutine--
		_ = conn.Close()
	}()

	for {
		y := <-c

		if y.Status {
			return
		}

		srv.Mu.Lock()
		srv.SendMessages++
		srv.Mu.Unlock()

		err := conn.WriteJSON(y)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (srv *Server) Connections() {

	for {
		select {
		case s := <-srv.LoginChan:
			srv.Mu.Lock()
			if srv.Clients[s.Name] != nil {
				m := models.Message{}
				m.Status = true
				srv.Clients[s.Name] <- m
			}

			srv.Clients[s.Name] = s.Channel
			srv.Mu.Unlock()
		case s := <-srv.LogoutChan:
			srv.Mu.Lock()
			c := srv.Clients[s]

			m := models.Message{}
			m.Status = true
			c <- m
			delete(srv.Clients, s)
			srv.Mu.Unlock()
		}
	}
}

func (srv *Server) Login(user string, c chan models.Message) {
	u := &User{
		Name:    user,
		Channel: c,
	}

	srv.LoginChan <- u
}

func (srv *Server) Logout(user string) {
	srv.LogoutChan <- user
}

const csr = `-----BEGIN CERTIFICATE-----
MIIDnzCCAoegAwIBAgIUY/e9t5IORtG8dF84YOMwI5dTm5kwDQYJKoZIhvcNAQEL
BQAwQzERMA8GA1UEAwwIaG9tZS5jb20xCzAJBgNVBAYTAkFaMRIwEAYDVQQIDAlC
YWt1IENpdHkxDTALBgNVBAcMBEJha3UwHhcNMjEwMzIwMDgwOTQ4WhcNMjIwMzIw
MDgwOTQ4WjBDMREwDwYDVQQDDAhob21lLmNvbTELMAkGA1UEBhMCQVoxEjAQBgNV
BAgMCUJha3UgQ2l0eTENMAsGA1UEBwwEQmFrdTCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBALryxrIFFYiqoj+bccmwZPxSNfbhzfudDtUOSYkrly/JoSYD
aq9gIaYAqaL9dQuZ82JMo3o6WdYSR+i/mm9k5kxBT+2Sl4IK1aYLMeUuZx8LmF9Q
G4rr3fHpOrQz1XjkpeXB4912iLx/n7i5NkW7O5bQBtCwFG0pWAO+bXM+NI/4J1dI
4XollMIthcQQsA9GM4bLSoNY0AOetMSoiPch00SdQy9l/Y1kLVOB2w6KveoD1HNg
TMXrHf0bgiOsMycstwYLg7igvuCgbfBEKkPKZB63bd1r1LGdpsUE7Hqjzi29E6AC
etba6Zk6AiPQ/sVqQ+xCIPIVzxAMeHSXVCBWux0CAwEAAaOBijCBhzAdBgNVHQ4E
FgQUfyDIhtLrpOCgI3QJBpGI2PJhD2kwHwYDVR0jBBgwFoAUfyDIhtLrpOCgI3QJ
BpGI2PJhD2kwDgYDVR0PAQH/BAQDAgWgMCAGA1UdJQEB/wQWMBQGCCsGAQUFBwMB
BggrBgEFBQcDAjATBgNVHREEDDAKgghob21lLmNvbTANBgkqhkiG9w0BAQsFAAOC
AQEAsKmSNXPR64DTblGzqGhl+i4HVhnyhrFnxD0UOjddjpTVU4OxpmYafw8PcPdq
HLA5A2oBe0EPuI+roTF72uKqDMCKaQBXkQClHGkv4q7GHphHjEq0q9J607nlIUG9
5wwzbQ1FdVSacmGbn6m3hJc34CHe7geGi6J6bQdUKCuWN34Rb6yo3N1ALXAsY5MJ
ymo97fEdA865o+RzRX+x7FRFbhgshWi44Op/lcQ4JnHbiYtdEcdcuc2o3OgTS/A1
BiPfJybf+9psSqP064MwHCowHB58KQRWU2OU9bglpE63lMKTqJrehx8A62tWZM7e
tQPHNZvKjAl02Sv4cUrgJP2IeQ==
-----END CERTIFICATE-----`

const privateKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC68sayBRWIqqI/
m3HJsGT8UjX24c37nQ7VDkmJK5cvyaEmA2qvYCGmAKmi/XULmfNiTKN6OlnWEkfo
v5pvZOZMQU/tkpeCCtWmCzHlLmcfC5hfUBuK693x6Tq0M9V45KXlwePddoi8f5+4
uTZFuzuW0AbQsBRtKVgDvm1zPjSP+CdXSOF6JZTCLYXEELAPRjOGy0qDWNADnrTE
qIj3IdNEnUMvZf2NZC1TgdsOir3qA9RzYEzF6x39G4IjrDMnLLcGC4O4oL7goG3w
RCpDymQet23da9SxnabFBOx6o84tvROgAnrW2umZOgIj0P7FakPsQiDyFc8QDHh0
l1QgVrsdAgMBAAECggEANLoMeGEetbEKmc4JxczObqvxNHRzWCfv6v9gliOJPJ0t
qj8Ec/o1A1Dkh2fc/yyojGz5HpweglYdmfOQZyKaIZ+6H1NdD/xmTbKSnAT+aK8o
hplda00jB/uz5udHqhUzBR4uWmP4JNIKBluWhwxLvjll8q321OL4Q/YNgJdm08O6
HP08MqZ2BoBI33aGAz9ywSfEt7U4jBuHU4H7VUKxKykT9sDKOygbQFAwjMUZjF3B
wcG0yEe6hTNPar8n8fvfmO+/xloADqCjcxXcim6m0kidvPNLBY2M1m6neGML/Ux3
kdUC0v/UfNzmr4oLTMl3zh98ffJrOES7JpjfvSZiMQKBgQDnPT43xZUOWX9QIZr9
Wa9Rj8RAoOqaI0STj0uBQPWAs4Nj6HDWo2lRqZusnIAU3PMNHdc3IpQbqDFSMJyt
l92DaIoOmTouFO8XMhh7JJcCsNOtEeV7P/U41W5b9bsM6X7K5wSvdN1G6vdDhq7B
uyCEdi/oIzgMnu5Z8ulm67qynwKBgQDO920XSWKzUwr04FleXfugTvdyCthaFs75
gCJg3AnGb7g3JItTZnwuly6TiSz5zYg+1kpyvJhJOtoOvJ93lVm0oQFA9n3MyM1l
woduMJt1XCtwf/IoIH12Pvt5w2uMIkb4QESySWmBQ+4n+iOVh2ykJ7DnkqmOqWdr
obMoO6vUwwKBgQDEI5Fzsxc0nbs8l9SkUv8/emenvhZgecvAMgqEbzoOWbX394BG
v0MlLm1KY1DM4YETvhz/ukfQkcCMC4nKQQd2YCTCLzxHPCB1F1vmj+m7MYvKwGRb
P6vb8kVyoSNw11lh98RkowbSEZl8YHA5CWWSlcEa8Uyof+KCz2UklIy+1wKBgQCh
HlcruLKAnZY6+eg4oXuA2diiTDUPNRBdhVW+B64Ib/J94xIfk/n6nzDgI/sCYPG+
0T3VwmHfKFSXAlo2Uuspxele9EUMxgm4PU8HBgoPu/gJNWGDwX9KLU/CA9LWndyX
6BhSnvnmasadEorfHjUCOe/q5u7eo5xiWthI6uMi1wKBgH3gVUXuuGqu63SqBV4/
gSYH0uJvDmkQrSVPQuxPJ/DG6ezZa+OGDQ/FeV9QlK8/08EH+D8zehk+cosLTTjU
nlorXGgpc7E/qT0R/xWO1k5PhP7UmTxw0RkR6jh25GERdNDuD7XHpA/OZF0aVmod
G2Oj3/YMQodII85LtAN5ZXsY
-----END PRIVATE KEY-----`
