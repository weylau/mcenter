package queue

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"mcenter/util"
)

type ProjectsConfig struct {
	Projects []ProjectConfig `json:"projects"`
}
type ProjectConfig struct {
	Name          string        `json:"name"`
	NotifyBase    string        `json:"notify_host"`
	NotifyTimeout int           `json:"notify_timeout"`
	Queues        []QueueConfig `json:"queues"`
}
type QueueConfig struct {
	QueueName       string   `json:"queue_name"`
	RoutingKey      []string `json:"routing_key"`
	NotifyPath      string   `json:"notify_path"`
	NotifyTimeout   int      `json:"notify_timeout"`
	RetryTimes      int      `json:"retry_times"`
	RetryDuration   int      `json:"retry_duration"`
	BindingExchange string   `json:"binding_exchange"`

	Project *ProjectConfig
}

func (qc *QueueConfig) InitQueue(channel *amqp.Channel) {
	qc.DeclareExchange(channel)
	qc.DeclareQueue(channel)
}

func (qc *QueueConfig) RetryQueueName() string {
	return fmt.Sprintf("%s-retry", qc.QueueName)
}
func (qc *QueueConfig) ErrorQueueName() string {
	return fmt.Sprintf("%s-error", qc.QueueName)
}
func (qc *QueueConfig) RetryExchangeName() string {
	return fmt.Sprintf("%s-retry", qc.QueueName)
}
func (qc *QueueConfig) RequeueExchangeName() string {
	return fmt.Sprintf("%s-retry-requeue", qc.QueueName)
}
func (qc *QueueConfig) ErrorExchangeName() string {
	return fmt.Sprintf("%s-error", qc.QueueName)
}

func (qc *QueueConfig) DeclareExchange(channel *amqp.Channel) {
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

func (qc *QueueConfig) DeclareQueue(channel *amqp.Channel) {
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

func ParserConfig(configFileName string) (*ProjectsConfig, error) {

	configFile, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil, err
	}
	projectsConfig := ProjectsConfig{}
	err = json.Unmarshal(configFile, &projectsConfig)
	if err != nil {
		return nil, err
	}
	return &projectsConfig, nil

}

//加载队列配置文件
func GetQuenus(configFileName string) ([]*QueueConfig, error) {
	allQueues := []*QueueConfig{}
	projectsConfig, err := ParserConfig(configFileName)
	if err != nil {
		return nil, err
	}
	projects := projectsConfig.Projects
	for i, _ := range projects {
		queues := projects[i].Queues
		for j, _ := range queues {
			queues[j].Project = &projects[i]
			allQueues = append(allQueues, &queues[j])
		}
	}
	return allQueues, nil
}
