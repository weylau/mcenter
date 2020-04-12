package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"mcenter/app/util"
)

type Projects struct {
	ProjectName string    `toml:"project_name"`
	NotifyHost  string    `toml:"notify_host"`
	Queues      []*Queues `toml:"queues"`
}
type Queues struct {
	QueueName       string   `toml:"queue_name"`
	NotifyPath      string   `toml:"notify_path"`
	NotifyTimeout   int      `toml:"notify_timeout"`
	RetryTimes      int      `toml:"retry_times"`
	RetryDuration   int      `toml:"retry_duration"`
	BindingExchange string   `toml:"binding_exchange"`
	RoutingKey      []string `toml:"routing_key"`
	ProjectConfig   *Projects
}

func (qc *Queues) InitQueue(channel *amqp.Channel) {
	qc.DeclareExchange(channel)
	qc.DeclareQueue(channel)
}

func (qc *Queues) RetryQueueName() string {
	return fmt.Sprintf("%s-retry", qc.QueueName)
}
func (qc *Queues) ErrorQueueName() string {
	return fmt.Sprintf("%s-error", qc.QueueName)
}
func (qc *Queues) RetryExchangeName() string {
	return fmt.Sprintf("%s-retry", qc.QueueName)
}
func (qc *Queues) RequeueExchangeName() string {
	return fmt.Sprintf("%s-retry-requeue", qc.QueueName)
}
func (qc *Queues) ErrorExchangeName() string {
	return fmt.Sprintf("%s-error", qc.QueueName)
}

func (qc *Queues) DeclareExchange(channel *amqp.Channel) {
	exchanges := []string{
		qc.BindingExchange,
		qc.RetryExchangeName(),
		qc.ErrorExchangeName(),
		qc.RequeueExchangeName(),
	}

	for _, e := range exchanges {
		err := channel.ExchangeDeclare(e, "topic", true, false, false, false, nil)
		util.FailOnErr(err, "")
	}
}

func (qc *Queues) DeclareQueue(channel *amqp.Channel) {
	var err error

	// 定义重试队列
	retryQueueOptions := map[string]interface{}{
		"x-dead-letter-exchange": qc.RequeueExchangeName(),
		"x-message-ttl":          int32(qc.RetryDuration * 1000),
	}

	_, err = channel.QueueDeclare(qc.RetryQueueName(), true, false, false, false, retryQueueOptions)
	util.FailOnErr(err, "DeclareQueue:1")
	err = channel.QueueBind(qc.RetryQueueName(), "#", qc.RetryExchangeName(), false, nil)
	util.FailOnErr(err, "DeclareQueue:2")

	// 定义错误队列

	_, err = channel.QueueDeclare(qc.ErrorQueueName(), true, false, false, false, nil)
	util.FailOnErr(err, "")
	err = channel.QueueBind(qc.ErrorQueueName(), "#", qc.ErrorExchangeName(), false, nil)
	util.FailOnErr(err, "")

	// 定义工作队列

	workerQueueOptions := map[string]interface{}{
		"x-dead-letter-exchange": qc.RetryExchangeName(),
	}
	_, err = channel.QueueDeclare(qc.QueueName, true, false, false, false, workerQueueOptions)
	util.FailOnErr(err, "DeclareQueue:1")

	for _, key := range qc.RoutingKey {
		err = channel.QueueBind(qc.QueueName, key, qc.BindingExchange, false, nil)
		util.FailOnErr(err, "DeclareQueue:")
	}

	// 最后，绑定工作队列 和 requeue Exchange
	err = channel.QueueBind(qc.QueueName, "#", qc.RequeueExchangeName(), false, nil)
	util.FailOnErr(err, "")
}
