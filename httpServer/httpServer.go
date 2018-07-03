package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

//"Build":"Development Build","Version":"1.2.0","Index":0,"LastContact":0,"KnownLeader":false
type VerTest struct {
	Build       string
	Version     string
	Index       uint8
	LastContact uint8
	KnownLeader bool
}

// hello world, the web server
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

func main() {
	http.HandleFunc("/v1/version", HelloServer)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
