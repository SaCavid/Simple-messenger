package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var (

	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)
type Server struct {
	Port    int
	Mu sync.Mutex
	Clients map[string]net.Conn
}

type Message struct {
	From string
	To   string
	Data string
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

func main() {
	log.SetFlags(log.Lshortfile)

	err := godotenv.Load(".env")
	if err != nil {
		// unable to connect to database. Quit app
		log.Fatal("Failed to load env! ", err)
	}

	t := os.Getenv("TLSPORT")
	log.Println(t)

	go TlsServer()

	//for {
	//	time.Sleep(60 * time.Second)
	//	log.Println("Server online")
	//}

	r := mux.NewRouter()
	r.HandleFunc("/", Hello)
	r.HandleFunc("/ws", Ws)

	//r.PathPrefix("/").Handler(http.FileServer(http.Dir("./index.html"))).Methods("GET")
	http.Handle("/", r)
	log.Println("Starting up on 80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func Hello(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("index.html").ParseFiles("./index.html", "./header.html")
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		panic(err)
	}
}

func Ws(w http.ResponseWriter, r *http.Request) {

	//srv.deleteUserConnection(token.UserId)
	wsConn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("Connected from ip address: ", wsConn.RemoteAddr().String())
	m := Message{}
	for {
		err := wsConn.ReadJSON(m)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println(m)
	}
}

func TlsServer() {

	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		log.Fatal(err)
	}

	cer2, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		log.Fatal(err)
	}

	configServer := &tls.Config{Certificates: []tls.Certificate{cer}, ServerName: "Test"}
	configServer.Certificates = append(configServer.Certificates, cer2)

	l, err := tls.Listen("tcp", ":2500", configServer)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := l.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("started server on 2500", len(configServer.Certificates))

	srv := Server{
		Port:    0,
		Clients: make(map[string]net.Conn, 0),
	}

	go ClientWithTls(rootCert, 1)
	go ClientWithTls(rootCert, 2)
	//go ClientWithTls(rootCert, 3)
	//go ClientNoTLS(0)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go Receiver(&srv, conn)
	}
}

func Receiver(srv *Server, conn net.Conn) {

	c := make(chan Message)

	go Transmitter(srv, c)

	logged := false
	log.Println("receiver working", conn.RemoteAddr())
	for {
		m := Message{}
		d := json.NewDecoder(conn)

		err := d.Decode(&m)
		if err != nil {
			log.Println(err)
			err := conn.Close()
			if err != nil {
				log.Println(err)
			}

			return
		}

		err = m.ValidateMessage()
		if err != nil {
			log.Println(err)

			err := conn.Close()
			if err != nil {
				log.Println(err)
			}

			return
		}

		if !logged {
			srv.Mu.Lock()
			srv.Clients[m.From] = conn
			srv.Mu.Unlock()
			logged = true
		} else {
			c <- m
		}
	}
}

func Transmitter(srv *Server, c chan Message) {

	for {
		y := <-c

		srv.Mu.Lock()
		client := srv.Clients[y.To]
		srv.Mu.Unlock()

		if client != nil {
			d, err := json.Marshal(y)
			if err != nil {
				log.Println(err)
				return
			}
			_, err = client.Write(d)
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Println("User didnt connected")
		}
	}

}

func ClientWithTls(rootCert string, i int) {

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		log.Fatal("failed to parse root certificate")
	}

	//block, _ := pem.Decode([]byte(rootCert))
	//var cert* x509.Certificate
	//cert, _ = x509.ParseCertificate(block.Bytes)
	//rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	//fmt.Println(rsaPublicKey.N)
	//fmt.Println(rsaPublicKey.E)
		
	config := &tls.Config{RootCAs: roots, ServerName: "localhost"}

	time.Sleep(5 * time.Second)
	conn, err := tls.Dial("tcp", "localhost:2500", config)
	if err != nil {
		log.Println(err.Error(), " ", i)
		return
	}
	log.Println(fmt.Sprintf("%x",conn.ConnectionState().TLSUnique))
	m := Message{
		From: fmt.Sprintf("tcpUser%d", i),
		To:   fmt.Sprintf("tcpUser%d", i),
		Data: "Looking for new solution",
	}

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
			m := Message{}
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

		m := Message{
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

func ClientNoTls(i int) {

	time.Sleep(5 * time.Second)
	conn, err := net.Dial("tcp", "localhost:2500")
	if err != nil {
		log.Fatal(err)
	}

	//conn := tls.Client(connp, config)
	m, err := conn.Write([]byte("hi there from TCP"))
	if err != nil {
		log.Print(err.Error())
	}
	log.Print(m)

	defer func() {

		err = conn.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	for {
		n, err := io.WriteString(conn, fmt.Sprintf("Hello secure Server %d", i))
		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println(n)
		time.Sleep(3 * time.Second)
	}

}

const rootCert = `-----BEGIN CERTIFICATE-----
MIIB+TCCAZ+gAwIBAgIJAL05LKXo6PrrMAoGCCqGSM49BAMCMFkxCzAJBgNVBAYT
AkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRn
aXRzIFB0eSBMdGQxEjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0xNTEyMDgxNDAxMTNa
Fw0yNTEyMDUxNDAxMTNaMFkxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0
YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxEjAQBgNVBAMM
CWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHGaaHVod0hLOR4d
66xIrtS2TmEmjSFjt+DIEcb6sM9RTKS8TZcdBnEqq8YT7m2sKbV+TEq9Nn7d9pHz
pWG2heWjUDBOMB0GA1UdDgQWBBR0fqrecDJ44D/fiYJiOeBzfoqEijAfBgNVHSME
GDAWgBR0fqrecDJ44D/fiYJiOeBzfoqEijAMBgNVHRMEBTADAQH/MAoGCCqGSM49
BAMCA0gAMEUCIEKzVMF3JqjQjuM2rX7Rx8hancI5KJhwfeKu1xbyR7XaAiEA2UT7
1xOP035EcraRmWPe7tO0LpXgMxlh2VItpc2uc2w=
-----END CERTIFICATE-----
`

//const rootCert2 = `-----BEGIN CERTIFICATE-----
//MIICUDCCAfegAwIBAgIUE9/A+PSWFZTxeQz7KP2OHOlqy70wCgYIKoZIzj0EAwIw
//fjELMAkGA1UEBhMCQVoxDTALBgNVBAgMBEJha3UxDTALBgNVBAcMBEJha3UxDTAL
//BgNVBAoMBEJldGExETAPBgNVBAsMCEludGVybmV0MQ0wCwYDVQQDDARGUU5IMSAw
//HgYJKoZIhvcNAQkBFhFkZWx0YUB0ZWxlY29tLmNvbTAeFw0yMDA4MzExMTI0MDNa
//Fw0zMDA4MjkxMTI0MDNaMH4xCzAJBgNVBAYTAkFaMQ0wCwYDVQQIDARCYWt1MQ0w
//CwYDVQQHDARCYWt1MQ0wCwYDVQQKDARCZXRhMREwDwYDVQQLDAhJbnRlcm5ldDEN
//MAsGA1UEAwwERlFOSDEgMB4GCSqGSIb3DQEJARYRZGVsdGFAdGVsZWNvbS5jb20w
//WTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQBTHRQQ7GHBEEciduqWAkJgqbAQPwK
//8kTUFbU8TN+/CceRXwx5nf2+J+cZi0VlsbhzAPE2RcwFBD8a7aFo/e0Qo1MwUTAd
//BgNVHQ4EFgQUbOfIi6b5v71qYWzNlLP9uTaASO4wHwYDVR0jBBgwFoAUbOfIi6b5
//v71qYWzNlLP9uTaASO4wDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNHADBE
//AiB/W0I69w5KgXFFeCbWRrjpL7X5/gQfZMVFqYL7lE2PpwIge91A7MtGpFKArxS1
//dHESJcpyUb0ZiCLJPKm/dlMDr38=
//-----END CERTIFICATE-----
//`

const rootCert3 = `-----BEGIN CERTIFICATE-----
MIICZjCCAgugAwIBAgIUR4SvVVV3xBP+tWeCybvF9KXnmsYwCgYIKoZIzj0EAwIw
gYcxCzAJBgNVBAYTAmF6MRAwDgYDVQQIDAdhYnNlcm9uMQ0wCwYDVQQHDARiYWt1
MQ0wCwYDVQQKDARiZXRhMQ0wCwYDVQQLDARiZXRhMREwDwYDVQQDDAhkaXJlY3Rv
cjEmMCQGCSqGSIb3DQEJARYXaW5mb0BiZXRhLWR5bmFtaWNzLmNvcG0wHhcNMjAw
OTAzMDgwMDM4WhcNMzAwOTAxMDgwMDM4WjCBhzELMAkGA1UEBhMCYXoxEDAOBgNV
BAgMB2Fic2Vyb24xDTALBgNVBAcMBGJha3UxDTALBgNVBAoMBGJldGExDTALBgNV
BAsMBGJldGExETAPBgNVBAMMCGRpcmVjdG9yMSYwJAYJKoZIhvcNAQkBFhdpbmZv
QGJldGEtZHluYW1pY3MuY29wbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCSx
tKEGtMJ4o27RQcAbBcmcITeQX3DULsi2h1xovNItE5AYFb4vVdDO8M6PnY+G0IXf
iEqzMUAH2RO6QDm87hWjUzBRMB0GA1UdDgQWBBQsvnOLfPfcDv4q9b1dDXTyGOKV
kzAfBgNVHSMEGDAWgBQsvnOLfPfcDv4q9b1dDXTyGOKVkzAPBgNVHRMBAf8EBTAD
AQH/MAoGCCqGSM49BAMCA0kAMEYCIQDvw1NrW5a/fxIm7l3PZMXgZ+E6vfR7VQFd
6LNkpQ30CQIhAI1OcoQfIjQ+GE7ArY1pqLggpM38u7dEVC37u5L3tqu/
-----END CERTIFICATE-----
`

const serverKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHg+g2unjA5BkDtXSN9ShN7kbPlbCcqcYdDu+QeV8XWuoAoGCCqGSM49
AwEHoUQDQgAEcZpodWh3SEs5Hh3rrEiu1LZOYSaNIWO34MgRxvqwz1FMpLxNlx0G
cSqrxhPubawptX5MSr02ft32kfOlYbaF5Q==
-----END EC PRIVATE KEY-----
`

const serverCert = `-----BEGIN CERTIFICATE-----
MIIB+TCCAZ+gAwIBAgIJAL05LKXo6PrrMAoGCCqGSM49BAMCMFkxCzAJBgNVBAYT
AkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRn
aXRzIFB0eSBMdGQxEjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0xNTEyMDgxNDAxMTNa
Fw0yNTEyMDUxNDAxMTNaMFkxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0
YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxEjAQBgNVBAMM
CWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHGaaHVod0hLOR4d
66xIrtS2TmEmjSFjt+DIEcb6sM9RTKS8TZcdBnEqq8YT7m2sKbV+TEq9Nn7d9pHz
pWG2heWjUDBOMB0GA1UdDgQWBBR0fqrecDJ44D/fiYJiOeBzfoqEijAfBgNVHSME
GDAWgBR0fqrecDJ44D/fiYJiOeBzfoqEijAMBgNVHRMEBTADAQH/MAoGCCqGSM49
BAMCA0gAMEUCIEKzVMF3JqjQjuM2rX7Rx8hancI5KJhwfeKu1xbyR7XaAiEA2UT7
1xOP035EcraRmWPe7tO0LpXgMxlh2VItpc2uc2w=
-----END CERTIFICATE-----
`
