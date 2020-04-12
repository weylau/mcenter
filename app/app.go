package app

import (
	"fmt"
	"github.com/juju/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"mcenter/app/config"
	"mcenter/app/loger"
	"mcenter/app/message"
	"mcenter/app/queue"
	"mcenter/app/util"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	receiveChan chan *message.Message
	arkChan     chan *message.Message
	resendChan  chan *message.Message
	exit        chan struct{}
	receiveExit chan struct{}
	workExit    chan struct{}
	arkExit     chan struct{}
	resendExit  chan struct{}
	queues      []*queue.Queues
	wc          *sync.WaitGroup
}

func Default(configFile string) *App {
	err := config.Default(configFile)
	if err != nil {
		panic(errors.ErrorStack(err))
	}
	loger.Default()
	app := &App{
		receiveChan: make(chan *message.Message, 1),
		arkChan:     make(chan *message.Message, 1),
		resendChan:  make(chan *message.Message, 1),
		exit:        make(chan struct{}),
		receiveExit: make(chan struct{}),
		workExit:    make(chan struct{}),
		arkExit:     make(chan struct{}),
		resendExit:  make(chan struct{}),
		wc:          &sync.WaitGroup{},
	}

	for _, project := range config.Conf.Projects {
		for _, q := range project.Queues {
			q.ProjectConfig = project
			app.queues = append(app.queues, q)
		}
	}
	return app
}

func (app *App) Run() {
	app.initQueue()
	app.start()
	loger.Loger.Infof("==== running ====")
	select {
	case sig := <-initSignalNotify():
		loger.Loger.Infof("==== exit (signal: %v) ====", sig)
	}
	app.close()
}

func (app *App) initQueue() {
	for _, project := range config.Conf.Projects {
		for _, q := range project.Queues {
			_, channel, err := util.SetupChannel(config.Conf.AmqpHost)
			if err != nil {
				panic(errors.ErrorStack(errors.Trace(err)))
			}
			q.InitQueue(channel)
		}
	}
}

func initSignalNotify() (chan os.Signal) {
	sc := make(chan os.Signal, 1)
	//Notify函数让signal包将输入信号转发到c。如果没有列出要传递的信号，会将所有输入信号传递到c；否则只传递列出的输入信号。
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	return sc
}

func (app *App) start() {
	app.receive()
	app.work()
	app.ack()
	app.resend()
}

//接收消息
func (app *App) receive() {
	wg := &sync.WaitGroup{}
	for _, qc := range app.queues {
		fmt.Println("QueueName：", qc.QueueName)
		fmt.Println("BindingExchange：", qc.BindingExchange)
		fmt.Println("RoutingKey：", qc.RoutingKey)
		for i := 0; i < config.Conf.ReceiveNum; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
			RECONNECT:
				for {
					_, channel, err := util.SetupChannel(config.Conf.AmqpHost)
					if err != nil {
						panic(errors.ErrorStack(errors.Trace(err)))
					}

					msgs, err := channel.Consume(
						qc.QueueName,
						"",
						false,
						false,
						false,
						false,
						nil)
					if err != nil {
						panic(errors.ErrorStack(errors.Trace(err)))
					}

					for {
						select {
						case <-app.exit:
							return
						case msg, ok := <-msgs:
							if !ok {
								loger.Loger.Infof("receiver: channel is closed, maybe lost connection")
								time.Sleep(5 * time.Second)
								continue RECONNECT
							}
							uuidStr := uuid.NewV4()
							msg.MessageId = fmt.Sprintf("%s", uuidStr)
							newMessage := message.Message{
								QueueConfig:  qc,
								AmqpDelivery: &msg,
							}
							app.receiveChan <- &newMessage
						}
					}
				}
			}()
		}
	}
	go func() {
		wg.Wait()
		close(app.receiveChan)
		app.receiveExit <- struct{}{}
		loger.Loger.Infof("receiveExit")
	}()
}

//处理消息回调
func (app *App) work() {
	wg := &sync.WaitGroup{}
	client := util.NewHttpClient(config.Conf.HttpMaxIdleConns, config.Conf.HttpMaxIdleConnsPerHost, config.Conf.HttpIdleConnTimeout)
	for i := 0; i < config.Conf.WorkNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				for m := range app.receiveChan {
					wg.Add(1)
					go func(client *http.Client, m *message.Message) {
						defer wg.Done()
						_, err := m.Notify(client)
						if err != nil {
							loger.Loger.Error(errors.ErrorStack(err))
						}
						app.arkChan <- m
					}(client, m)
				}
				break
			}
		}()
	}
	go func() {
		wg.Wait()
		close(app.arkChan)
		app.workExit <- struct{}{}
		loger.Loger.Infof("workExit")
	}()
}

func (app *App) ack() {
	wg := &sync.WaitGroup{}
	for i := 0; i < config.Conf.AckNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				for m := range app.arkChan {
					if m.ResponseCode == message.NotifySuccess {
						err := m.Ack()
						if err != nil {
							loger.Loger.Error(errors.ErrorStack(err))
						}
					} else if m.IsMaxRetry() {
						//已经超过重试次数另外处理
						err := m.Ack()
						if err != nil {
							loger.Loger.Error(errors.ErrorStack(err))
						}
						app.resendChan <- m
					} else {
						err := m.Reject()
						if err != nil {
							loger.Loger.Error(errors.ErrorStack(err))
						}
					}
				}
				break

			}

		}()
	}
	go func() {
		wg.Wait()
		close(app.resendChan)
		app.arkExit <- struct{}{}
		loger.Loger.Infof("arkExit")
	}()
}

func (app *App) resend() {
	wg := &sync.WaitGroup{}
	for i := 0; i < config.Conf.ResendNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		RECONNECT:
			for {
				conn, channel, err := util.SetupChannel(config.Conf.AmqpHost)
				if err != nil {
					panic(errors.ErrorStack(errors.Trace(err)))
				}
				for m := range app.resendChan {
					err := m.CloneAndPublish(channel)
					if err == amqp.ErrClosed {
						time.Sleep(5 * time.Second)
						continue RECONNECT
					}

				}
				_ = conn.Close()
				break
			}
		}()
	}
	go func() {
		wg.Wait()
		app.resendExit <- struct{}{}
		loger.Loger.Infof("resendExit")
	}()
}

func (app *App) close() {
	close(app.exit)
	<-app.receiveExit
	<-app.workExit
	<-app.arkExit
	<-app.resendExit
	loger.Loger.Infof("==== App closed ====")
}
