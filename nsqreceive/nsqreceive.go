//Nsq接收测试
package main

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/common/util"
	"github.com/nsqio/go-nsq"
)

// 消费者
type ConsumerT struct{}

// 主函数
func main() {
	channelName := util.GenerateUUID() + "#ephemeral"
	fmt.Println("ephermal channel name: " + channelName)
	//topicName := "sbx-cclog-a-1.0"
	topicName := "sbx-peerlog-peer0.org1.example.com"
	addr := "192.168.1.248:4161"
	InitConsumer(topicName, channelName, addr)
	for {
		time.Sleep(time.Second * 10)
	}
}

//处理消息
func (*ConsumerT) HandleMessage(msg *nsq.Message) error {
	fmt.Println("receive", msg.NSQDAddress, "message:", string(msg.Body))
	return nil
}

//初始化消费者
func InitConsumer(topic string, channel string, address string) {
	cfg := nsq.NewConfig()
	cfg.LookupdPollInterval = time.Second          //设置重连时间
	c, err := nsq.NewConsumer(topic, channel, cfg) // 新建一个消费者
	if err != nil {
		panic(err)
	}
	c.SetLogger(nil, 0)        //屏蔽系统日志
	c.AddHandler(&ConsumerT{}) // 添加消费者接口

	//建立NSQLookupd连接
	if err := c.ConnectToNSQLookupd(address); err != nil {
		panic(err)
	}

	//建立多个nsqd连接
	// if err := c.ConnectToNSQDs([]string{"127.0.0.1:4150", "127.0.0.1:4152"}); err != nil {
	//  panic(err)
	// }

	// 建立一个nsqd连接
	// if err := c.ConnectToNSQD("127.0.0.1:4150"); err != nil {
	//  panic(err)
	// }
}
