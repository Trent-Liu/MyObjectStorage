package main

import (
	"awesomeProject/src/lib/mysql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//访问数据服务的DELETE对象接口进行散列值的删除。
func del(hash string) {
	log.Println("delete", hash)
	url := "http://" + os.Getenv("LISTEN_ADDRESS") + "/objects/" + hash
	request, _ := http.NewRequest("DELETE", url, nil)
	client := http.Client{}
	client.Do(request)
}

func main() {
	//获得$STORAGE_ROOT/objects/目录下所有文件。
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")

	//在for循环中遍历访问这些文件，从文件名中获得对象的散列值。
	for i := range files {
		hash := strings.Split(filepath.Base(files[i]), ".")[0]
		//检查元数据服务是否存在该散列值
		hashInMetadata, e := mysql.HasHash(hash)
		if e != nil {
			log.Println(e)
			return
		}
		//如果不存在，则del删除散列值。
		if !hashInMetadata {
			del(hash)
		}
	}
}
