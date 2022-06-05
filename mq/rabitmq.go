package mq

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"log"
)

const (
	FollowQueueName   = "follow"
	FavoriteQueueName = "favoriteActionQueue"
)

var (
	followMQ   *RabbitMQ
	favoriteMQ *RabbitMQ
)

func GetFollowMQ() *RabbitMQ {
	return followMQ
}
func GetFavoriteMQ() *RabbitMQ {
	return favoriteMQ
}

//RabbitMQ 结构体
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机名称
	Exchange string
	//bind Key 名称
	Key string
	//连接信息
	Mqurl string
}

func InitMQ(mqUrl string) {
	followMQ = NewRabbitMQSimple(FollowQueueName, mqUrl)
	favoriteMQ = NewRabbitMQSimple(FavoriteQueueName, mqUrl)
}

// NewRabbitMQSimple 创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string, mqUrl string) *RabbitMQ {
	//创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, "", "", mqUrl)
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect rabbitmq!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string, mqUrl string) *RabbitMQ {
	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: mqUrl}
}

// Destroy
//断开channel 和 connection
func (r *RabbitMQ) Destroy() {
	err := r.channel.Close()
	if err != nil {
		logrus.Errorln(err)
	}
	err = r.conn.Close()
	if err != nil {
		logrus.Errorln(err)
	}
}

//错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		logrus.Fatalf("%s:%s", message, err)
		logrus.Panicln(fmt.Sprintf("%s:%s", message, err))
	}
}

// PublishSimple 直接模式队列生产
func (r *RabbitMQ) PublishSimple(message string) error {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		return err
	}
	//调用channel 发送消息到队列中
	err = r.channel.Publish(
		r.Exchange,
		r.QueueName,
		//如果为true，根据自身exchange类型和routekey规则无法找到符合条件的队列会把消息返还给发送者
		false,
		//如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return err
	}
	return nil
}

// ConsumeSimple
//simple模式下消费者
func (r *RabbitMQ) ConsumeSimple(method func(string) error) {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	//接收消息
	msgs, err := r.channel.Consume(
		q.Name, // queue
		//用来区分多个消费者
		"", // consumer
		//是否自动应答
		true, // auto-ack
		//是否独有
		false, // exclusive
		//设置为true，表示不能将同一个Conenction中生产者发送的消息传递给这个Connection中的消费者
		false, // no-local
		//列是否阻塞
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(msgs)
	//forever := make(chan bool)
	//启用协程处理消息
	go func() {
		for d := range msgs {
			//自定义处理消息method
			err = method(string(d.Body))
			if err != nil {
				//log.Fatalln("msg handle fail")
			}
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	//<-forever
}
