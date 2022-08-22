package mq

import (
	"encoding/json"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	//通道
	channel *amqp.Channel
	//连接对象
	conn     *amqp.Connection
	Name     string
	exchange string
}

func New(s string) *RabbitMQ {
	conn, e := amqp.Dial(s)
	if e != nil {
		panic(e)
	}

	ch, e := conn.Channel()
	if e != nil {
		panic(e)
	}

	q, e := ch.QueueDeclare(
		//队列名字
		"", // name
		//是否持久化
		false, // durable
		//不用的时候是否自动删除
		true, // delete when unused
		//用来指定是否独占队列
		false, // exclusive
		//no-wait
		false, // no-wait
		//其他参数
		nil, // arguments
	)
	if e != nil {
		panic(e)
	}

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.conn = conn
	mq.Name = q.Name
	return mq
}

//接收方，队列要绑定到交换机才能接收到消息。
func (q *RabbitMQ) Bind(exchange string) {
	e := q.channel.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil)
	if e != nil {
		panic(e)
	}
	q.exchange = exchange
}

//消息直接发送到队列
func (q *RabbitMQ) Send(queue string, body interface{}) {
	str, e := json.Marshal(body)
	if e != nil {
		panic(e)
	}
	e = q.channel.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	if e != nil {
		panic(e)
	}
}

//消息发送到交换机
func (q *RabbitMQ) Publish(exchange string, body interface{}) {
	str, e := json.Marshal(body)
	if e != nil {
		panic(e)
	}
	e = q.channel.Publish(
		exchange, //交换机
		"",       //队列名字
		false,    //是否强制性
		false,    //是否立即处理掉
		amqp.Publishing{
			ReplyTo: q.Name,      //传输类型
			Body:    []byte(str), //发送的消息
		})
	if e != nil {
		panic(e)
	}
}

//消费者读取队列消息
func (q *RabbitMQ) Consume() <-chan amqp.Delivery {
	c, e := q.channel.Consume(q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if e != nil {
		panic(e)
	}
	return c
}

func (q *RabbitMQ) Close() {
	q.channel.Close()
	q.conn.Close()
}
