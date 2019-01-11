package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

var conn *amqp.Connection
var channel *amqp.Channel
var count = 0

const (
	//queueName = "p.m2.test"
	exchange = "test_exchange"
	mqurl    = "queue://test:123456@192.168.0.252/fangzz"
)

func main() {
	c := make(chan bool)
	go func(c chan bool) {
		for {
			time.Sleep(time.Second)
			push(c)
		}
	}(c)

	for {
		select {
		case <-c:
			fmt.Printf("已发送 %d 条消息\n", count)
		}
	}
	fmt.Println("end")
	closeAll()

}

func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:%s", msg, err)
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

func mqConnect() {
	var err error
	conn, err = amqp.Dial(mqurl)

	failOnErr(err, "failed to connect tp rabbitmq")

	channel, err = conn.Channel()

	failOnErr(err, "failed to open a channel")
}

func closeAll() {
	channel.Close()
	conn.Close()
}

//连接rabbitmq server
func push(c chan bool) {

	if channel == nil {
		mqConnect()
	}

	//q, err := channel.QueueDeclare(
	//	queueName, // name
	//	true,      // durable
	//	false,     // delete when unused
	//	false,     // exclusive
	//	false,     // no-wait
	//	nil,       // arguments
	//)
	//failOnErr(err, "failed to open a channel")
	msgContent := "{\"code\":1,\"msg\":\"ok\"}"

	err := channel.Publish(exchange, "test.test.test", false, false, amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(msgContent),
		DeliveryMode: 2,
	})
	if err == nil {
		count++
		c <- true
	} else {
		fmt.Println("消息发送失败", err)
	}
}
