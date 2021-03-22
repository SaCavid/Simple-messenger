package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	SendMessages     uint64
	ReceivedMessages uint64
)

type Message struct {
	From   string
	To     string
	Data   string
	Users  []string
	Status bool // status of channel
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

	return nil
}

func TestTcp(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	err := godotenv.Load(".env")
	if err != nil {
		// unable to connect to database. Quit app
		log.Fatal("Failed to load env! ", err)
	}

	ma, err := strconv.ParseUint(os.Getenv("FAKEUSERS"), 10, 64)
	if err != nil {
		ma = 1000
	}

	for i := 0; i < int(ma); i++ {
		time.Sleep(1000 * time.Microsecond)
		go ClientWithTls(os.Getenv("TLSPORT"), csr, int(i))
		// go ClientNoTls(os.Getenv("TCPPORT"), int(ma) + i)
	}

	for {
		time.Sleep(3 * time.Second)
		log.Println("controller: ", SendMessages, ReceivedMessages)
	}
}

func ClientWithTls(addr string, rootCert string, i int) {

	time.Sleep(3 * time.Second)
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		log.Fatal("failed to parse root certificate")
	}

	config := &tls.Config{RootCAs: roots, ServerName: "home.com"}

	conn, err := tls.Dial("tcp", ":"+addr, config)
	if err != nil {
		log.Println(err.Error(), " ", i)
		return
	}

	m := NewMessage(fmt.Sprintf("tcpUser%d", i), fmt.Sprintf("tcpUser%d", i), "Looking for new solution", nil)

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

		log.Println("finished tls client")
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
			ReceivedMessages++
		}
	}()

	for {

		m := Message{
			From: fmt.Sprintf("tcpUser%d", i),
			To:   fmt.Sprintf("tcpUser%d", 2),
			Data: "Looking for new solution",
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

		time.Sleep(1 * time.Second)
		SendMessages++
	}
}

func ClientNoTls(addr string, i int) {

	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", ":"+addr)
	if err != nil {
		log.Println(err.Error(), " ", i)
		return
	}

	m := NewMessage(fmt.Sprintf("tcpUser%d", i), fmt.Sprintf("tcpUser%d", i), "Looking for new solution", nil)

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

		log.Println("finished tls client")
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
			ReceivedMessages++
		}
	}()

	for {

		m := Message{
			From: fmt.Sprintf("tcpUser%d", i),
			To:   fmt.Sprintf("tcpUser%d", 2),
			Data: "Looking for new solution",
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

		time.Sleep(1 * time.Second)
		SendMessages++
	}
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
