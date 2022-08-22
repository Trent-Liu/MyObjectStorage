package rsconstruct

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

type encoder struct {
	writers []io.Writer
	enc     reedsolomon.Encoder
	cache   []byte
}

func NewEncoder(writers []io.Writer) *encoder {
	//调用reedsolomon.New生成一个具有四个数据片加两个校验片的RS码编码器enc
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	//输入参数writers和enc作为生成的encoder结构体的成员返回。
	return &encoder{writers, enc, nil}
}

func (e *encoder) Write(p []byte) (n int, err error) {
	length := len(p)
	current := 0
	//将p中待写入的数据以块的形式放入缓存，如果缓存已满就调用Flush方法将缓存实际写入writers
	for length != 0 {
		next := BLOCK_SIZE - len(e.cache)
		if next > length {
			next = length
		}
		e.cache = append(e.cache, p[current:current+next]...)
		if len(e.cache) == BLOCK_SIZE {
			e.Flush()
		}
		current += next
		length -= next
	}
	return len(p), nil
}

//将所有的数据依次写入分片
func (e *encoder) Flush() {
	if len(e.cache) == 0 {
		return
	}
	//首先调用encoder的成员遍历enc的Split方法将缓存的数据切成4个数据片
	shards, _ := e.enc.Split(e.cache)
	//调用enc的Encode方法生成两个校验片
	e.enc.Encode(shards)
	//将6个片的数据依次写入writers并清空缓存。
	for i := range shards {
		e.writers[i].Write(shards[i])
	}
	e.cache = []byte{}
}
