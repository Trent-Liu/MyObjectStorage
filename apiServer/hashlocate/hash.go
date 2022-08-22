package hashlocate

import (
	"awesomeProject/apiServer/etcd"
)

//将一个字符串映射为1——8之间的任意一个数字
func BKDRHash(str string) uint64 {
	seed := uint64(131) // 31 131 1313 13131 131313 etc..
	hash := uint64(0)
	for i := 0; i < len(str); i++ {
		hash = (hash * seed) + uint64(str[i])
	}
	return ((hash & 0x7FFFFFFF) % 1024) % uint64(etcd.ServiceLength)
}
