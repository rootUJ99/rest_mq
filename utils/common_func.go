package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	ampq "github.com/rabbitmq/amqp091-go"
)

const ORDERS_QUEUE string = "orders_quqeue"

func HandleErrorWithLog(err error) {
	if err != nil {
		log.Println(err)
		// log.Panic(err)
	}
}

type MessageQueue struct {
	Queue ampq.Queue
	Conn *ampq.Connection
	Channel *ampq.Channel
}


func DailAndDeclareQuque() (MessageQueue, error){
	conn, err := ampq.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return MessageQueue{}, nil
	}
	ch, err := conn.Channel()

	if err != nil {
		return MessageQueue{}, nil
	}
	queue, err := ch.QueueDeclare(ORDERS_QUEUE, false, false, false, false, nil)

	if err != nil {
		return MessageQueue{}, nil
	}

	messageQueue := MessageQueue{
		Queue: queue,
		Channel: ch,
		Conn: conn,

	}
	
	return messageQueue, nil
}

func SendToQueueWithHttp(url string, method string, data []byte){

	type jsonData struct {
		Url string `json:"url"`
		Method string `json:"method"` 
		RequestBody []byte `json:"requestbody"`   
	}

	finalRequestData := jsonData{
		Url : url,
		Method: method,
		RequestBody: data,
	}

	jsonDataToSend, err := json.Marshal(finalRequestData)

	HandleErrorWithLog(err)

	messageQueue, err := DailAndDeclareQuque()
	defer messageQueue.Conn.Close()
	defer messageQueue.Channel.Close()
	HandleErrorWithLog(err)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second *10)
	defer cancel()

	fmt.Printf("got here finally %s", jsonDataToSend)
	err = messageQueue.Channel.PublishWithContext(ctx, "", messageQueue.Queue.Name, false, false, ampq.Publishing{
		ContentType: "text/pain",
		Body: jsonDataToSend,
	})

	HandleErrorWithLog(err)
	
	fmt.Printf("\n message has been sent")
}

// func SendToQueue(data []byte){
//
// 	messageQueue, err := DailAndDeclareQuque()
// 	defer messageQueue.Conn.Close()
// 	defer messageQueue.Channel.Close()
// 	HandleErrorWithLog(err)
// 	
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second *10)
// 	defer cancel()
//
// 	fmt.Printf("got here finally %s", data)
// 	err = messageQueue.Channel.PublishWithContext(ctx, "", messageQueue.Queue.Name, false, false, ampq.Publishing{
// 		ContentType: "text/pain",
// 		Body: data,
// 	})
//
// 	HandleErrorWithLog(err)
// 	
// 	fmt.Printf("\n message has been sent")
// }


