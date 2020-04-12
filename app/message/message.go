package message

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/juju/errors"
	"github.com/streadway/amqp"
	"io/ioutil"
	"mcenter/app/config"
	"mcenter/app/loger"
	"mcenter/app/queue"
	"mcenter/app/util"
	"net/http"
)

const (
	NotifySuccess = 1
	NotifyFailure = 0
)

type Message struct {
	QueueConfig  *queue.Queues
	AmqpDelivery *amqp.Delivery
	ResponseCode int
}

type HttpResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (m *Message) Notify(client *http.Client) (*Message,error) {
	m.ResponseCode = 0
	//加密
	loger.Loger.Infof("msg body:" + string(m.AmqpDelivery.Body))
	loger.Loger.Infof("msg token:" + config.Conf.AesToken)
	loger.Loger.Infof("msg MD5 token:" + util.Md5(config.Conf.AesToken)[16:])
	mstring, err := util.AesEncrypt(m.AmqpDelivery.Body, []byte(util.Md5(config.Conf.AesToken)[16:]))
	if err != nil {
		return m,errors.Trace(err)
	}
	msg := base64.StdEncoding.EncodeToString(mstring)
	url := m.QueueConfig.ProjectConfig.NotifyHost + m.QueueConfig.NotifyPath
	loger.Loger.Infof("msg callback url:" + url)
	loger.Loger.Infof("msg encode:" + msg)
	data := `{"msg":"` + msg + `"}`
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return m,errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return m,errors.Trace(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return m,errors.Trace(err)
	}
	response := HttpResponse{}
	err = json.Unmarshal(body, &response)
	if err == nil && response.Code == 200 {
		m.ResponseCode = 1
	} else {
		err := fmt.Errorf("http response code is %d , msg is %s", response.Code, response.Msg)
		return m,errors.Trace(err)
	}


	return m, nil
}

func (m *Message) Ack() error {
	loger.Loger.Infof(fmt.Sprintf("Message-Ack: body: %s\n", string(m.AmqpDelivery.Body)))
	err := m.AmqpDelivery.Ack(false)
	return errors.Trace(err)
}

func (m Message) Reject() error {
	loger.Loger.Infof(fmt.Sprintf("Message-Reject: body: %s\n", string(m.AmqpDelivery.Body)))
	err := m.AmqpDelivery.Reject(false)
	return errors.Trace(err)
}

func (m *Message) IsMaxRetry() bool {
	retries := m.CurrentMessageRetries()

	maxRetries := m.QueueConfig.RetryTimes
	loger.Loger.Infof(fmt.Sprintf("retries:%d,maxRetries:%d\n", retries, maxRetries))
	return retries >= maxRetries
}

func (m *Message) CloneAndPublish(channel *amqp.Channel) error {
	msg := m.AmqpDelivery
	qc := m.QueueConfig

	errMsg := util.CloneToPublishMsg(msg)
	err := channel.Publish(qc.ErrorQueueName(), msg.RoutingKey, false, false, *errMsg)
	return errors.Trace(err)
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
