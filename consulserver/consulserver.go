package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"net/http"

	consulapi "github.com/hashicorp/consul/api"
)

// sample demo in :
// https://github.com/changjixiong/goNotes/blob/master/consulnotes/server/main.go

const RECV_BUF_LEN = 1024

var (
	// Address for consul check
	Address string

	// Port for consul check
	Port int

	// Check action for consul check
	Check string
)

func init() {
	// Address for consul check
	Address = "192.168.1.197"

	// Port for consul check
	Port = 12345

	// Check action for consul check
	Check = "/v1/version"
}

func registerServer() {

	config := consulapi.DefaultConfig()
	config.Address = "192.168.1.248:8500"
	client, err := consulapi.NewClient(config)

	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	checkPort := Port

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "zydNode_1"
	registration.Name = "zydTestServer"
	registration.Port = Port
	registration.Tags = []string{"zouyudiserverNode"}
	registration.Address = Address
	registration.Check = &consulapi.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d%s", registration.Address, checkPort, Check),
		Timeout:                        "3s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务
	}

	err = client.Agent().ServiceRegister(registration)

	if err != nil {
		log.Fatal("register server error : ", err)
	}

}

// VerTest struct
type VerTest struct {
	Build       string
	Version     string
	Index       uint8
	LastContact uint8
	KnownLeader bool
}

// HelloServer the web server action
func HelloServer(w http.ResponseWriter, req *http.Request) {
	versionInfo := &VerTest{
		"Development Build",
		"1.0.0",
		0,
		1,
		false,
	}
	versionJSON, err := json.Marshal(versionInfo)
	if err != nil {
		panic(err)
	}

	io.WriteString(w, string(versionJSON))
}

func checkServer() {
	//start a http server
	http.HandleFunc(Check, HelloServer)
	//fmt.Sprintf("%s:%d", Address, Port)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func main() {

	go registerServer()

	checkServer()

	/* ln, err := net.Listen("tcp", "0.0.0.0:9527")

	if nil != err {
		panic("Error: " + err.Error())
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			panic("Error: " + err.Error())
		}

		go EchoServer(conn)
	} */

}

// EchoServer echo server info
func EchoServer(conn net.Conn) {
	buf := make([]byte, RECV_BUF_LEN)
	defer conn.Close()

	for {
		n, err := conn.Read(buf)
		switch err {
		case nil:
			log.Println("get and echo:", "EchoServer "+string(buf[0:n]))
			conn.Write(append([]byte("EchoServer "), buf[0:n]...))
		case io.EOF:
			log.Printf("Warning: End of data: %s\n", err)
			return
		default:
			log.Printf("Error: Reading data: %s\n", err)
			return
		}
	}
}
