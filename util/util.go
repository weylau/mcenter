package util

import (
	"bytes"
	"fmt"
	"github.com/streadway/amqp"
	"net/http"
	"os"
	"strings"
	"time"
)

func LogOnString(content string) {
	if content != "" {
		content := fmt.Sprintf(" Log - %s", content)
		WriteLog("./log.log", content)
	}
}

func LogOnError(err interface{}) {
	if err != nil {
		errinfo := fmt.Sprintf(" ERROR - %s", err)
		WriteLog("./log.log", errinfo)
	}
}

func FailOnErr(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

func WriteLog(name, content string) {
	fd, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd_time := time.Now().Format("2006-01-02 15:04:05")
	fd_content := strings.Join([]string{fd_time, content, "\n"}, "")
	buf := []byte(fd_content)
	fd.Write(buf)
	fd.Close()

}

func BytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}

func NewHttpClient(maxIdleConns, maxIdleConnsPerHost, idleConnTimeout int) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(idleConnTimeout) * time.Second,
	}

	client := &http.Client{
		Transport: tr,
	}

	return client
}

func CloneToPublishMsg(msg *amqp.Delivery) *amqp.Publishing {
	newMsg := amqp.Publishing{
		Headers: msg.Headers,

		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,

		Body: msg.Body,
	}

	return &newMsg
}

func SetupChannel(mqHost string) (*amqp.Connection, *amqp.Channel, error) {

	conn, err := amqp.Dial(mqHost)
	if err != nil {
		FailOnErr(err, "")
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		FailOnErr(err, "")
		return nil, nil, err
	}

	err = channel.Qos(1, 0, false)
	if err != nil {
		FailOnErr(err, "")
		return nil, nil, err
	}
	return conn, channel, nil
}
