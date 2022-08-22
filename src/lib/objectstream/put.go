package objectstream

import (
	"io"
)

//io.PipeWriter的指针和一个error的channel
type PutStream struct {
	writer *io.PipeWriter //实现Write方法
	c      chan error     //发生的错误传回主线程
}


//写入writer
func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

//关闭writer
func (w *PutStream) Close() error {
	w.writer.Close()
	return <-w.c
}
