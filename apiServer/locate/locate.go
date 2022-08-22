package locate

import (
	"awesomeProject/src/lib/mq"
	"awesomeProject/src/lib/rsconstruct"
	"encoding/json"
	"os"
	"time"
)

type LocateMessage struct {
	Addr string
	Id   int
}

//接受需要定位的对象的名字
func Locate(name string) (locateInfo map[int]string) {
	//创建一个新的消息队列
	q := mq.New(os.Getenv("RABBITMQ_SERVER"))

	//向dataServers群发这个对象名字的定位信息，所有的dataServer的消息队列都可以接收到此消息
	q.Publish("dataServers", name)
	//等待来自数据节点的反馈
	c := q.Consume()
	//启动匿名函数，2s后关闭这个临时消息队列，超时机制防止无休止的等待
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()

	//返回得到的数据节点的监听地址
	locateInfo = make(map[int]string)

	//循环获取最多6条信息，每条消息都包含了拥有某个分片的数据服务节点的地址和分片的id
	//rsconstruct.ALL_SHARDS为常数6
	for i := 0; i < rsconstruct.ALL_SHARDS; i++ {
		msg := <-c
		if len(msg.Body) == 0 {
			return
		}
		var info LocateMessage
		json.Unmarshal(msg.Body, &info)
		//并被放在输出参数的locateInfo变量中返回。
		locateInfo[info.Id] = info.Addr
	}
	return
}

//检查Locate结果
func Exist(name string) bool {
	return len(Locate(name)) >= rsconstruct.DATA_SHARDS
}
