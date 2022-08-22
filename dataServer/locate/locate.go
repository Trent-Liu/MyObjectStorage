package locate

import (
	"awesomeProject/src/lib/mq"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type LocateMessage struct {
	Addr string
	Id   int
}

//以字符串为键，整型为值的map，用于缓存所有的对象
var objects = make(map[string]int)

//保护对objects的读写操作
var mutex sync.Mutex

//实际定位对象
func Locate(hash string) int {
	mutex.Lock()
	//利用map操作判断某个散列值是否存在于objects中
	id, ok := objects[hash]
	mutex.Unlock()
	if !ok {
		return -1
	}
	//返回一个整型，用于返回分片的id，对象不存在则返回-1
	return id
}

//将对象以及其分片id加入缓存
func Add(hash string, id int) {
	mutex.Lock()
	objects[hash] = id
	mutex.Unlock()
}

//将一个散列值移出缓存
func Del(hash string) {
	mutex.Lock()
	delete(objects, hash)
	mutex.Unlock()
}

//监听定位消息
func StartLocate() {
	//创建一个结构体
	q := mq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()

	//绑定DataServers exchange
	q.Bind("dataServers")

	//得到channel接收消息
	c := q.Consume()

	//遍历，并接收消息，消息正文是需要定位的对象名字
	for msg := range c {
		//解除json的编码，消息队列里收到的对象散列值作为locate参数
		hash, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			panic(e)
		}
		//检查该文件是否存在
		id := Locate(hash)
		if id != -1 {
			//如果该文件存在，则调用Send，返回本节点的监听地址，表示该对象存在于本服务节点上
			//且id不为-1，将自身的节点监听地址和id打包成结构体反馈types.LocateMessage
			q.Send(msg.ReplyTo, LocateMessage{Addr: os.Getenv("LISTEN_ADDRESS"), Id: id})
		}
	}
}

func CollectObjects() {
	//读取$STORAGE_ROOT/objects/目录里的所有文件
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")
	for i := range files {
		//调用filepath.Base获取基本文件名，也就是对象的散列值，加入objects缓存。
		file := strings.Split(filepath.Base(files[i]), ".")
		//获取对象的散列值hash以及分片id，加入定位缓存
		if len(file) != 3 {
			panic(files[i])
		}
		hash := file[0]
		id, e := strconv.Atoi(file[1])
		if e != nil {
			panic(e)
		}
		objects[hash] = id
	}
}
