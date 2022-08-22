package calculatehash

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
)

func CalculateHash(r io.Reader) string {
	//变量h类型是sha256.digest结构体，实现的接口为hash.Hash
	h := sha256.New()
	//参数r中读取数据并写入h，h会对写入的数据计算散列值。
	io.Copy(h, r)
	//读取散列值，在base64编码。
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
