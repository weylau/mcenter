# Mcenter
> 消息回调中心主要是为了解决当队列服务过多时，不同项目或服务都需要各自实现自己的消息消费逻辑，浪费资源；消息回调中心可以配置所有的队列，消费端只需要通过实现http或rpc接口供消息回调中心调用即可，这样可以对消息队列统一管理；

# 配置文件
```toml
#mqhost
amqp_host = "amqp://mq:123456@172.16.57.110:5672/"

# 日志级别
log_level = "debug"

# debug 为true时，日志直接输出在控制台，false时写入文件
debug = true

#消息接收协程数
receive_num = 2

#Work协程数
work_num = 2

#Ack协程数
ack_num = 2

#消息重发协程数
resend_num = 1

#HttpMaxIdleConns
http_max_idle_conns = 100

#HttpMaxIdleConnsPerHost
http_max_idle_conns_per_host = 2

#HttpIdleConnTimeout
http_idle_conn_timeout = 30

#鉴权token
aes_token = "1234567"

# =============================队列信息配置===============================
[[projects]]
# 项目名
project_name = "test_project"
# 回调域名
notify_host = "http://172.16.57.110:9003"

[[projects.queues]]

# 队列名
queue_name = "queue_name"
error_queue_name = "error_queue_name"
# 回调路由
notify_path = "/callback/test"
# 回调超时时长
notify_timeout = 5

# 重试次数
retry_times = 5
retry_duration = 5

# 绑定exchange
binding_exchange = "binding_exchange"

# 绑定key
routing_key = ["test.test.test"]



```


## Http回调处理demo
[mcenter-callback-demo](https://github.com/weylau/mcenter-callback-demo "mcenter-callback-demo")
