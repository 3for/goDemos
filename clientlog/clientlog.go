package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/fsouza/go-dockerclient"
	consulapi "github.com/hashicorp/consul/api"
	nsq "github.com/nsqio/go-nsq"
)

func readerToChan(reader *bytes.Buffer, exit <-chan bool) <-chan string {
	c := make(chan string)

	go func() {
		for {
			select {
			case <-exit:
				close(c)
				return
			default:
				line, err := reader.ReadString('\n')
				line = strings.TrimSpace(line)
				if err != nil && err != io.EOF {
					close(c)
					return
				} else if line != "" {
					c <- line
				}
			}
		}
	}()
	return c
}

var producer *nsq.Producer

//LOGMsg is the struct for log
type LOGMsg struct {
	EventType string `json:"eventType"`
	Subject   string `json:"subject"`
	Payload   string `json:"payload"`
}

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

// CheckHandler the web server action
func CheckHandler(w http.ResponseWriter, req *http.Request) {
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
	http.HandleFunc(Check, CheckHandler)
	//fmt.Sprintf("%s:%d", Address, Port)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {

	go registerServer()

	go checkServer()

	//The remote machine of 192.168.1.248 has changed the docker settings for remote connection, the port of 5555 can be changed .
	endpoint := "192.168.1.248:5555"

	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	containerName := "/peer0.org1.example.com"
	containers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			//"ancestor": []string{"hyperledger/fabric-peer"}, //filter containers by specific image.
			"name": []string{containerName},
		},
	})

	if err != nil {
		panic(err)
	}

	strIP1 := "192.168.1.248:4150"
	strIP2 := "192.168.1.248:5150"
	InitProducer(strIP1)
	//topicName := "sbx-cclog-a-1.0"
	/*
		IN SANDBOX DEPLOY CHAINCODE SITUATION
		get docker logs(filter out DEBUG ), structured and resend to the same topic as chainconsle consumers.
		So in this case, the chainconsole box will get a piece of DUPLICATED same logs
	*/
	topicName := "sbx-peerlog-peer0.org1.example.com"

	for _, ctr := range containers {
		fmt.Println("container ID: ", ctr.ID)
		var stdoutBuffer, stderrBuffer bytes.Buffer
		opts := docker.LogsOptions{
			Container:    ctr.ID,
			OutputStream: &stdoutBuffer,
			ErrorStream:  &stderrBuffer,
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
			Tail:         "0",
		}
		exit := make(chan bool)
		go func() {
			client.Logs(opts)
			close(exit)
		}()
		stdoutCh := readerToChan(&stdoutBuffer, exit)
		stderrCh := readerToChan(&stderrBuffer, exit)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for value := range stdoutCh {
				fmt.Println("stdout: " + value)
			}
		}()
		go func() {
			defer wg.Done()
			for value := range stderrCh {
				strlogs := strings.Split(value, " -> ")
				logMsg := ""
				if len(strlogs) > 1 {
					if len(strlogs[1]) > 5 {
						expectLog := strlogs[1]
						expectLevel := expectLog[0:4]
						//fmt.Println("expect log Level: " + expectLevel)
						// Filter out the DEBUG information.
						if expectLevel != "DEBU" {
							logMsg = value
							fmt.Println("stderr: " + value)
						}
					}

				} else {
					logMsg = value
					fmt.Println("stderr: " + value)
				}
				if logMsg != "" {
					eventLog := &LOGMsg{
						"event-type-log",
						topicName,
						logMsg}

					eventJSON, err := json.Marshal(eventLog)
					if err != nil {
						panic(err)
					}
					eventMsg := string(eventJSON)
					fmt.Println("eventMsg: " + eventMsg)

					for err := Publish(topicName, eventMsg); err != nil; err = Publish(topicName, eventMsg) {
						//切换IP重连
						strIP1, strIP2 = strIP2, strIP1
						InitProducer(strIP1)
					}
				}
			}
		}()
		wg.Wait()
	}

	//关闭
	defer producer.Stop()
}

// InitProducer 初始化生产者
func InitProducer(str string) {
	var err error
	fmt.Println("address: ", str)
	producer, err = nsq.NewProducer(str, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

// Publish 发布消息
func Publish(topic string, message string) error {
	var err error
	if producer != nil {
		if message == "" { //不能发布空串，否则会导致error
			return nil
		}
		err = producer.Publish(topic, []byte(message)) // 发布消息
		return err
	}
	return fmt.Errorf("producer is nil. Error: %s ", err)
}
