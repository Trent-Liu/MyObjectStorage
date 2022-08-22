package etcd

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"log"
	"time"
)

//Lease是一种检测客户端存活状况的机制，如果etcd在给定的TTL时间内未收到keepAlive，则租约到期
//每个key最多附加一个租约，当租约到期或被撤销时，该租约所附加的所有key都将被删除

//ServiceRegister 创建租约注册服务
type ServiceRegister struct {
	cli     *clientv3.Client //etcd v3 client
	leaseID clientv3.LeaseID //租约ID
	//租约keepalieve相应chan
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string //key
	val           string //value
}

//NewServiceRegister 新建注册服务
//etcd服务器节点，键为节点位置，val值为其地址 IP:port
func NewServiceRegister(endpoints []string, key, val string, lease int64, dailTimeout int) (*ServiceRegister, error) {
	//注册ectd的客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,                                //etcd的多个节点服务地址
		DialTimeout: time.Duration(dailTimeout) * time.Second, //DialTimeout 创建client的首次连接超时时间。如果这么多时间内都没连接成功就返回err，一旦client创建成功，就不再关心后续底层的连接状态。
	})
	//判断是否连接失败
	if err != nil {
		return nil, err
	}

	//本节点注册服务
	ser := &ServiceRegister{
		cli: cli,
		key: key,
		val: val,
	}

	//申请租约设置时间keepalive
	if err := ser.putKeyWithLease(lease); err != nil {
		return nil, err
	}

	return ser, nil
}

//设置租约
func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	//创建一个新的租约，并设置ttl时间
	resp, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}

	//注册服务并绑定租约
	_, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//设置续租 定期发送需求请求
	//KeepAlive使给定的租约永远有效。 如果发布到通道的keepalive响应没有立即被使用，
	// 则租约客户端将至少每秒钟继续向etcd服务器发送保持活动请求，直到获取最新的响应为止。
	//etcd client会自动发送ttl到etcd server，从而保证该租约一直有效
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	s.leaseID = resp.ID
	log.Println(s.leaseID)
	s.keepAliveChan = leaseRespChan
	log.Printf("Put key:%s  val:%s  success!", s.key, s.val)
	return nil
}

//ListenLeaseRespChan 监听 续租情况
func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		//fmt.Println("续约成功", leaseKeepResp)
		leaseKeepResp.ID++
		leaseKeepResp.ID--
	}
	fmt.Println("关闭续租")
}

// Close 注销服务
func (s *ServiceRegister) Close() error {
	//撤销租约
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	log.Println("撤销租约")
	return s.cli.Close()
}
