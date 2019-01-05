package message

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"mcenter/config"
	"mcenter/queue"
	"mcenter/util"
	"net/http"
	"strings"
)

const (
	NotifySuccess = 1
	NotifyFailure = 0
)

type Message struct {
	QueueConfig  *queue.QueueConfig
	AmqpDelivery *amqp.Delivery
	ResponseCode int
}

type HttpResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (m *Message) Notify(client *http.Client) *Message {
	m.ResponseCode = 0
	//加密
	mstring, err := util.AesEncrypt(m.AmqpDelivery.Body, []byte(util.Md5(config.AesToken)[16:]))
	if err != nil {
		util.LogOnError(fmt.Sprintf("AesEncrypt error:%s", err))
		return m
	}
	util.LogOnString(m.QueueConfig.Project.NotifyBase + m.QueueConfig.NotifyPath)
	req, err := http.NewRequest("POST", m.QueueConfig.Project.NotifyBase+m.QueueConfig.NotifyPath, strings.NewReader("msg="+base64.StdEncoding.EncodeToString(mstring)))
	if err != nil {
		util.LogOnError(err)
		return m
	}
	resp, err := client.Do(req)
	if err != nil {
		util.LogOnError(err)
		return m
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogOnError(err)
		return m
	}
	response := HttpResponse{}
	json.Unmarshal(body, &response)
	if response.Code == 200 {
		m.ResponseCode = 1
	} else {
		err := fmt.Errorf("http response code is %d , msg is %s", response.Code, response.Msg)
		util.LogOnError(err)
		return m
	}

	defer resp.Body.Close()
	return m
}

func (m *Message) Ack() error {
	util.LogOnString(fmt.Sprintf("Message-Ack: body: %s\n", string(m.AmqpDelivery.Body)))
	err := m.AmqpDelivery.Ack(false)
	util.LogOnError(err)
	return err
}

func (m Message) Reject() error {
	util.LogOnString(fmt.Sprintf("Message-Reject: body: %s\n", string(m.AmqpDelivery.Body)))
	err := m.AmqpDelivery.Reject(false)
	util.LogOnError(err)
	return err
}

func (m *Message) IsMaxRetry() bool {
	retries := m.CurrentMessageRetries()

	maxRetries := m.QueueConfig.RetryTimes
	util.LogOnString(fmt.Sprintf("retries:%d,maxRetries:%d\n", retries, maxRetries))
	return retries >= maxRetries
}

func (m *Message) CloneAndPublish(channel *amqp.Channel) error {
	msg := m.AmqpDelivery
	qc := m.QueueConfig

	errMsg := util.CloneToPublishMsg(msg)
	err := channel.Publish(qc.ErrorExchangeName(), msg.RoutingKey, false, false, *errMsg)
	util.LogOnError(err)
	return err
}

/*
获取消息重试的次数
*/
func (m *Message) CurrentMessageRetries() int {
	msg := m.AmqpDelivery

	xDeathArray, ok := msg.Headers["x-death"].([]interface{})
	if !ok {
		return 0
	}

	if len(xDeathArray) <= 0 {
		return 0
	}

	for _, h := range xDeathArray {
		xDeathItem := h.(amqp.Table)

		if xDeathItem["reason"] == "rejected" {
			return int(xDeathItem["count"].(int64))
		}
	}

	return 0
}
