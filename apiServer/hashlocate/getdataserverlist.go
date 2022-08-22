package hashlocate

import (
	"awesomeProject/apiServer/etcd"
	"awesomeProject/src/lib/rsconstruct"
	"sort"
)

func GetServerList(hash string) (serverlist map[int]string) {
	//得到1——8的任意一个节点
	index := BKDRHash(hash)

	servicelist := etcd.Ser.GetServices()
	keys := make([]string, len(servicelist))
	i := 0
	for k, _ := range servicelist {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	serverlist = make(map[int]string)

	for i := 0; i < rsconstruct.ALL_SHARDS; i++ {
		k := keys[(uint64(i)+index)%uint64(etcd.ServiceLength)]

		//得到某个hash会绑定的所有地址
		serverlist[i] = servicelist[k]
	}

	return

}
