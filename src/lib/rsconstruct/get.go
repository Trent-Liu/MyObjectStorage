package rsconstruct

import (
	"awesomeProject/src/lib/objectstream"
	"fmt"
	"io"
)

//内嵌decoder结构体。
type RSGetStream struct {
	*decoder
}

func NewRSGetStream(locateInfo map[int]string, hash string, size int64) (*RSGetStream, error) {

	//创建长度为6的io.Reader数组reader，读取6个分片的数据。
	readers := make([]io.Reader, ALL_SHARDS)
	//循环遍历6个分片的id，在locateInfo中查找该分片所在的数据服务节点地址。
	for i := 0; i < ALL_SHARDS; i++ {
		server := locateInfo[i]
		//如果数据服务节点存在，则调用objectstream.NewGetStream()打开一个对象读取流用于读取该分片数据。打开的流被保存在readers数组相应的元素中。
		reader, e := objectstream.NewGetStream(server, fmt.Sprintf("%s.%d", hash, i))
		if e == nil {
			readers[i] = reader
		}
	}

	writers := make([]io.Writer, ALL_SHARDS)
	ShardSize := (size + DATA_SHARDS - 1) / DATA_SHARDS
	var e error
	//再次遍历readers，如果某个元素为nil，则调用objectstream.NewTempPutStream创建相应的临时对象写入流用于回复分片。
	for i := range readers {
		if readers[i] == nil {
			//打开的流被保存在writers数组相应的元素中。
			writers[i], e = objectstream.NewTempPutStream(locateInfo[i], fmt.Sprintf("%s.%d", hash, i), ShardSize)
			if e != nil {
				return nil, e
			}
		}
	}

	//reader和writers数组形成互补的关系，对于某个分片id，要么在readers中存在相应的读取流，要么在writers中存在相应的写入流。
	//加对象的大小size作为参数调用newDecoder
	dec := NewDecoder(readers, writers, size)
	return &RSGetStream{dec}, nil
}

//遍历writers成员，如果某个分片的writer不为nil，则调用其Commit方法，参数为true，意味着临时对象将被转正。
func (s *RSGetStream) Close() {
	for i := range s.writers {
		if s.writers[i] != nil {
			s.writers[i].(*objectstream.TempPutStream).Commit(true)
		}
	}
}
