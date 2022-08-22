package rsconstruct

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

type decoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder //RS解码
	size      int64
	cache     []byte //缓存数据
	cacheSize int    //缓存数据
	total     int64  //当前已经读取的字节
}

func NewDecoder(readers []io.Reader, writers []io.Writer, size int64) *decoder {
	//创建4+2的RS解码器enc，并设置decoder结构体中相应的属性
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &decoder{readers, writers, enc, size, nil, 0, 0}
}

//RSGetStream的Read方法就是其内嵌结构体decoder的Read方法
func (d *decoder) Read(p []byte) (n int, err error) {
	//当cache中没有更多数据时会调用getData方法获取数据
	if d.cacheSize == 0 {
		e := d.getData()
		if e != nil {
			return 0, e
		}
	}

	//输入参数p的数组长度
	length := len(p)
	//如果length超出当前缓存的数据大小
	if d.cacheSize < length {
		//就令length等于缓存的数据大小，仅保留剩下的部分。
		length = d.cacheSize
	}
	d.cacheSize -= length
	copy(p, d.cache[:length])
	d.cache = d.cache[length:]
	//返回length，通知调用方本次读取一共有多少数据被复制到p中。
	return length, nil
}

func (d *decoder) getData() error {
	//首先判断当前已经解码的数据大小是否等于对象原始大小，如果已经相等说明所有数据都被读取，返回io.EOF
	if d.total == d.size {
		return io.EOF
	}
	//还有数据要被读取，创建一个长度为6的shards，以及一个长度为0的整型数组repairIds
	//保存相应分片中读取的数据
	shards := make([][]byte, ALL_SHARDS)
	repairIds := make([]int, 0)

	//遍历6个shards，以及一个长度为0的整型数组repairIds
	for i := range shards {
		//如果某个分片对应的reader是nil，则分片已丢失，需要在repairIds中添加分片的id
		if d.readers[i] == nil {
			repairIds = append(repairIds, i)
		} else {
			//如果分片不是nil，shards需要被初始化一个长度为8000的字节数组
			shards[i] = make([]byte, BLOCK_PER_SHARD)
			//reader中完整读取8000字节的数据保存在shards里
			n, e := io.ReadFull(d.readers[i], shards[i])
			if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
				//非EOF失败，被置为nil
				shards[i] = nil
			} else if n != BLOCK_PER_SHARD {
				//读取的数据长度n不到8000字节，将shards的实际长度缩减为n
				shards[i] = shards[i][:n]
			}
		}
	}

	//读取一轮后，shards中保存了读取自对应分片的数据，并且回恢复被置为nil的shards
	e := d.enc.Reconstruct(shards)

	if e != nil {
		//这一步如果返回错误，则对象已经遭到了不可修复的破坏，只能原样返回上层。
		return e
	}
	//6个shards中都保存了对应分片的正确数据，遍历repairIds，将需要恢复的分片的数据写入相应的writer
	for i := range repairIds {
		id := repairIds[i]
		d.writers[id].Write(shards[id])
	}
	//遍历4个数据分片，将每个分片的数据添加到缓存chache中。
	for i := 0; i < DATA_SHARDS; i++ {
		shardSize := int64(len(shards[i]))
		//修改缓存当前的大小cacheSize以及当前已经读取的全部数据的大小total
		if d.total+shardSize > d.size {
			shardSize -= d.total + shardSize - d.size
		}
		d.cache = append(d.cache, shards[i][:shardSize]...)
		d.cacheSize += int(shardSize)
		d.total += shardSize
	}
	return nil
}
