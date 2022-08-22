package main

import (
	"awesomeProject/src/lib/mysql"
	"log"
)

//设置最大版本留存数量
const MIN_VERSION_COUNT = 5

func main() {

	//搜索出元数据服务中所有版本数量大于等于6的对象，保存在Bucket结构体的数组buckets里。
	buckets, e := mysql.SearchVersionStatus(MIN_VERSION_COUNT + 1)
	if e != nil {
		log.Println(e)
		return
	}
	//遍历buckets，并在一个for循环中调用es.DelMetadata,从该对象当前最小的版本号开始一一删除，直到最后还剩5个。
	for i := range buckets {
		bucket := buckets[i]
		for v := 0; v < bucket.Doc_count-MIN_VERSION_COUNT; v++ {
			mysql.DelMetadata(bucket.Key, bucket.User, v+int(bucket.Min_version.Value))
		}
	}

}
