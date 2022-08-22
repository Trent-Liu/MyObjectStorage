package rsconstruct

import (
	"awesomeProject/src/lib/objectstream"
	"fmt"
	"io"
)

type RSPutStream struct {
	*encoder
}

//dataServers是字符串数组，存6个数据服务节点的地址。
//hash和size分别是需要PUT的对象的散列值和大小。
func NewRSPutStream(dataServers []string, hash string, size int64) (*RSPutStream, error) {
	//如果数据服务节点不为6个则错误
	if len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("dataServers number error")
	}

	//计算出数据分片应该有的大小
	ShardSize := (size + DATA_SHARDS - 1) / DATA_SHARDS

	//长度为6的io.Writers数组，每个元素都是一个objectstream.TempPutStream，上传一个分片对象
	writers := make([]io.Writer, ALL_SHARDS)
	var e error

	//生成
	for i := range writers {
		//某个数据节点，<hash>.X, 计算的数据片大小
		//在数据节点上创建临时文件，等待校验后转正或者删除
		writers[i], e = objectstream.NewTempPutStream(dataServers[i],
			fmt.Sprintf("%s.%d", hash, i), ShardSize)
		if e != nil {
			return nil, e
		}
	}

	//创建一个encoder结构体的指针enc，作为RSPutStream的内嵌结构体返回。
	enc := NewEncoder(writers)

	return &RSPutStream{enc}, nil
}

//将临时对象转正或者删除
func (s *RSPutStream) Commit(success bool) {
	//调用内嵌结构体的encoder的Flush方法将缓存中最后的数据写入
	s.Flush()
	//然后对encoder的成员数组writers中的元组调用Commit方法将6个临时对象依次转正或者删除。
	for i := range s.writers {
		s.writers[i].(*objectstream.TempPutStream).Commit(success)
	}
}
