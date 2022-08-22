package main

import (
	"awesomeProject/apiServer/objects"
	"awesomeProject/src/lib/calculatehash"
	"awesomeProject/src/lib/mysql"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//对象的检查和修复。
//也要在数据服务节点上定期运行
func main() {
	//调用filepath.Glob获取$STORAGE_ROOT/object/目录下的所有文件
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")

	//在for循环中循环遍历访问这些文件，从文件名中获得对象的散列值，并调用verify检查数据
	for i := range files {
		hash := strings.Split(filepath.Base(files[i]), ".")[0]
		verify(hash)
	}
}

//调用es.SearchHashSize从元数据服务中获取该散列值对应的对象大小。
func verify(hash string) {
	log.Println("verify", hash)
	size, e := mysql.SearchHashSize(hash)
	if e != nil {
		log.Println(e)
		return
	}
	//以对象的散列值和大小为参数调用objects.GetStream创建一个对象数据流
	//objects.GetStream创建一个指向rs.RSGetStream结构体的指针，通过读取rs.RSGetStream并在最后关闭，底层会自动完成数据修复。
	//如果数据已经损坏的不可被修复，那么计算散列值的时候必定不能匹配。
	stream, e := objects.GetStream(hash, size)
	if e != nil {
		log.Println(e)
		return
	}
	//计算对象的散列值，检查是否一致。
	d := calculatehash.CalculateHash(stream)
	if d != hash {
		//如果不一致则报告错误。
		log.Printf("object hash mismatch, calculated=%s, requested=%s", d, hash)
	}
	stream.Close()
}
