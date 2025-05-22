package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rootuj/async_queue_with_rest/utils"
)

func Receiver(){
	messageQueue, err := utils.DailAndDeclareQuque()
	defer messageQueue.Conn.Close()
	defer messageQueue.Channel.Close()

	utils.HandleErrorWithLog(err)
	
	messages, err := messageQueue.Channel.Consume(messageQueue.Queue.Name, "", false, false, false, false, nil)

	utils.HandleErrorWithLog(err)
	
	var forever chan struct{}

	go func () {
		for i := range messages {
			fmt.Printf("%s", i.Body)
			DoHttpRequest(i.Body)
			i.Nack(false, false)
		}
	}()
	fmt.Println("waiting for the message press crtl + c to stop the queue clinet")
	<-forever
}

func DoHttpRequest(jdata []byte) {
	type jsonData struct {
		Url string `json:"url"`
		Method string `json:"method"` 
		RequestBody []byte `json:"requestbody"`   
	}

	var unMarshelData jsonData

	err:=json.Unmarshal(jdata, &unMarshelData)
	utils.HandleErrorWithLog(err)

	req, err := http.NewRequest(unMarshelData.Method, unMarshelData.Url, bytes.NewBuffer(unMarshelData.RequestBody))
	utils.HandleErrorWithLog(err)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	utils.HandleErrorWithLog(err)

	defer resp.Body.Close()
	reponseBody, err := io.ReadAll(resp.Body)

	fmt.Printf("%s ", reponseBody)
}

func main() {
	fmt.Println("---startubg broker---")
	Receiver()
}


