package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

//直接从io.Reader读取相应的正文，不需要管道适配，只需要一个成员
type GetStream struct {
	reader io.Reader
}

//获取数据流的HTTP服务地址
func newGetStream(url string) (*GetStream, error) {
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	return &GetStream{r.Body}, nil
}

//封装，函数只需要server和object两个字符串
func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" {
		return nil, fmt.Errorf("invalid server %s object %s", server, object)
	}
	return newGetStream("http://" + server + "/objects/" + object)
}

//读取reader成员。
func (r *GetStream) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
