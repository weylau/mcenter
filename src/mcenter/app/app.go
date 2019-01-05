package app

import (
	"flag"
	"fmt"
	"github.com/facebookgo/pidfile"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"log"
	"mcenter/config"
	"mcenter/message"
	"mcenter/queue"
	"mcenter/util"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type App struct {
	sysConfig   *config.SysConfig
	queues      []*queue.QueueConfig
	receiveChan chan *message.Message
	arkChan     chan *message.Message
	resendChan  chan *message.Message
	exit        chan struct{}
	receiveExit chan struct{}
	workExit    chan struct{}
	arkExit     chan struct{}
	saveExit    chan struct{}
}

func NewApp() *App {
	return &App{
		receiveChan: make(chan *message.Message, 1),
		arkChan:     make(chan *message.Message, 1),
		resendChan:  make(chan *message.Message, 1),
		exit:        make(chan struct{}),
		receiveExit: make(chan struct{}),
		workExit:    make(chan struct{}),
		arkExit:     make(chan struct{}),
		saveExit:    make(chan struct{}),
	}
}

func (app *App) Run() {
	defer func() {
		if err := recover(); err != nil {
			util.LogOnError(err)
		}
	}()
	//加载配置文件
	wdPath := flag.String("wd", "", "Server work directory")
	confPath := flag.String("conf", "", "Server configuration file path")
	queuesPath := flag.String("queues", "", "Server configuration file path")
	flag.Parse()
	applicationDir := ""
	if *wdPath != "" {
		_, err := os.Open(*wdPath)
		if err != nil {
			panic(err)
		}
		os.Chdir(*wdPath)
		applicationDir, err = os.Getwd()
	} else {
		var err error
		applicationDir, err = os.Getwd()
		if err != nil {
			file, _ := exec.LookPath(os.Args[0])
			applicationPath, _ := filepath.Abs(file)
			applicationDir, _ = filepath.Split(applicationPath)
		}

	}

	defaultQueuesConfPath := fmt.Sprintf("%s/%s", applicationDir, config.QueunsFile)
	defaultConfPath := fmt.Sprintf("%s/%s", applicationDir, config.ConfFile)

	if *queuesPath == "" {
		*queuesPath = defaultQueuesConfPath
	}

	if *confPath == "" {
		*confPath = defaultConfPath
	}

	sysConfig, err := config.GetConfig(*confPath)
	if err != nil {
		panic("加载系统配置失败:" + *confPath)
	}
	pidfile.Write()
	config.AesToken = sysConfig.AesToken
	app.sysConfig = sysConfig
	allQueues, err := queue.GetQuenus(*queuesPath)
	if err != nil {
		panic("加载队列配置失败:" + *queuesPath)
	}
	app.queues = allQueues
	// create queues
	for _, qc := range allQueues {
		_, channel, err := util.SetupChannel(app.sysConfig.AmqpHost)
		if err != nil {
			util.FailOnErr(err, "")
		}
		qc.InitQueue(channel)
	}

	app.Receive()
	app.Work()
	app.Ack()
	app.Resend()
	util.LogOnString("==== App running ====")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	sig := <-c
	//如果一分钟都关不了则强制关闭
	timeout := time.NewTimer(time.Minute)
	wait := make(chan struct{})
	go func() {
		app.Destroy()
		wait <- struct{}{}
	}()
	select {
	case <-timeout.C:
		panic(fmt.Sprintf("==== gomsg close timeout (signal: %v) ====", sig))
	case <-wait:
		logstr := fmt.Sprintf("==== gomsg closing down (signal: %v) ====", sig)
		util.LogOnString(logstr)
	}
}

//接收消息
func (app *App) Receive() {
	var wg sync.WaitGroup
	for _, qc := range app.queues {
		for i := 0; i < app.sysConfig.ReceiveNum; i++ {
			wg.Add(1)
			go func(qc *queue.QueueConfig, exit chan struct{}) {
				defer wg.Done()
			RECONNECT:
				for {
					_, channel, err := util.SetupChannel(app.sysConfig.AmqpHost)
					if err != nil {
						util.FailOnErr(err, "")
					}

					msgs, err := channel.Consume(
						qc.QueueName,
						"",
						false,
						false,
						false,
						false,
						nil)
					util.FailOnErr(err, "")

					for {
						select {
						case msg, ok := <-msgs:
							if !ok {
								log.Printf("receiver: channel is closed, maybe lost connection")
								time.Sleep(5 * time.Second)
								continue RECONNECT
							}
							uuidStr, err := uuid.NewV4()
							if err != nil {
								util.LogOnError(err)
							} else {
								msg.MessageId = fmt.Sprintf("%s", uuidStr)
								newMessage := message.Message{
									QueueConfig:  qc,
									AmqpDelivery: &msg,
								}
								app.receiveChan <- &newMessage
							}
						case <-exit:
							return
						}
					}
				}
			}(qc, app.exit)
		}
	}
	go func() {
		wg.Wait()
		close(app.receiveChan)
		app.receiveExit <- struct{}{}
		util.LogOnString("receiveExit")
	}()
}

//处理消息回调
func (app *App) Work() {
	var wg sync.WaitGroup
	client := util.NewHttpClient(app.sysConfig.HttpMaxIdleConns, app.sysConfig.HttpMaxIdleConnsPerHost, app.sysConfig.HttpIdleConnTimeout)
	for i := 0; i < app.sysConfig.WorkNum; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				for m := range app.receiveChan {
					wg.Add(1)
					go func(client *http.Client, m *message.Message) {
						defer wg.Done()
						m.Notify(client)
						app.arkChan <- m
					}(client, m)
				}
				break
			}
		}(&wg)
	}
	go func() {
		wg.Wait()
		close(app.arkChan)
		app.workExit <- struct{}{}
	}()
}

func (app *App) Ack() {
	var wg sync.WaitGroup
	for i := 0; i < app.sysConfig.AckNum; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				for m := range app.arkChan {
					if m.ResponseCode == message.NotifySuccess {
						m.Ack()
					} else if m.IsMaxRetry() {
						//已经超过重试次数另外处理
						m.Ack()
						app.resendChan <- m
					} else {
						m.Reject()
					}
				}
				break

			}

		}(&wg)
	}
	go func() {
		wg.Wait()
		close(app.resendChan)
		app.arkExit <- struct{}{}
	}()
}

func (app *App) Resend() {
	var wg sync.WaitGroup
	for i := 0; i < app.sysConfig.ResendNum; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
		RECONNECT:
			for {
				conn, channel, err := util.SetupChannel(app.sysConfig.AmqpHost)
				if err != nil {
					util.FailOnErr(err, "")
				}
				for m := range app.resendChan {
					err := m.CloneAndPublish(channel)
					if err == amqp.ErrClosed {
						time.Sleep(5 * time.Second)
						continue RECONNECT
					}

				}
				conn.Close()
				break
			}
		}(&wg)
	}
	go func() {
		wg.Wait()
		app.saveExit <- struct{}{}
	}()
}

func (app *App) Destroy() {
	close(app.exit)
	<-app.receiveExit
	<-app.workExit
	<-app.arkExit
	<-app.saveExit
	util.LogOnString("==== App destroyed ====")
}
