package main

import (
	"bytes"
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
			OFSRequest(i.Body)
		}
	}()
	fmt.Println("waiting for the message press crtl + c to stop the queue clinet")
	<-forever
}

func OFSRequest(jsonData []byte) {
	url := "http://localhost:6969/make_ofs"

	fmt.Println("this is the way")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
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


