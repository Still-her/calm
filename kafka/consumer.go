package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
  "github.com/Still-her/calm"
	"github.com/IBM/sarama"
)

// 接口
type Consumer interface {
	sarama.ConsumerGroupHandler
	Getbrokers() string
	Gettopics() string
	Getgroupname() string
	Getoffset() int64
}

// 示例
type cliexample struct {
	brokers   string
	topics    string
	groupname string
	offset    int64
}

// 示例
func (c cliexample) Getbrokers() string   { return c.brokers }
func (c cliexample) Gettopics() string    { return c.topics }
func (c cliexample) Getgroupname() string { return c.groupname }
func (c cliexample) Getoffset() int64     { return c.offset }

// 示例
func (cliexample) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (cliexample) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (c cliexample) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		sess.MarkMessage(msg, "")
		sess.Commit()
	}
	return nil
}

// 示例
func Newcliexample(brokers, topics, groupname string, offset int64) *cliexample {
	cli := &cliexample{
		brokers:   brokers,
		topics:    topics,
		groupname: groupname,
		offset:    offset,
	}
	return cli
}

// 示例
func TestConsumerTopic() {
	que := calm.CreateList()
	cli := Newcliexample("nn1.hadoop", "topic_test", "group_test", -1)
	go ConsumerGroup(cli)
	for {
		time.Sleep(time.Duration(1) * time.Second)
		if que.IsEmpty() {
			continue
		}
		data := que.GetFirst().Data
		if nil != data {
			val := data.(*sarama.ConsumerMessage)
			fmt.Println("sarama.ConsumerMessage info:", "[topic:", val.Topic, "] [partiton:", val.Partition, "] [offset:", val.Offset, "] [value:", string(val.Value), "] [time:", val.Timestamp, "]")
		}
		que.PopFront()
	}
}

/*
OffsetNewest(从服务端的offset开始消费) 和OffsetOldest 真正区别是什么？
创建一个group并来消费topic数据之前，这个topic可能就存在并已经被写入数据了, OffsetNewest只获取group被创建后没有被标记为消费的数据,因为才创建group，
所以该group的offset还为unknow，则OffsetOldest 消费这个分区从创建到现在的所有数据,当group 中有多个成员时，则每个成员只消费被分配到的分区上的数据.

如果消费，但是没有提交offset,当group中有新的成员加入，发生rebalance的时候，会自动把没有提交的offset数据再重复消费一遍
手动提交，先标记sess.MarkMessage(msg, "")， 当处理完数据完成消费， 再提交sess.Commit()
*/
func ConsumerGroup(handler Consumer) {

	//根据字符串解析地址列表
	addressList := strings.Split(handler.Getbrokers(), ",")
	if len(addressList) < 1 || addressList[0] == "" {
		return
	}

	topicList := strings.Split(handler.Gettopics(), ",")
	if len(topicList) < 1 || topicList[0] == "" {
		return
	}
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange // 分区分配策略
	config.Consumer.Offsets.Initial = handler.Getoffset()                  //sarama.OffsetNewest sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cg, err := sarama.NewConsumerGroup(addressList, handler.Getgroupname(), config)
	if err != nil {
		return
	}
	defer cg.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			/*
				![important]
				应该在一个无限循环中不停地调用 Consume()
				因为每次 Rebalance 后需要再次执行 Consume() 来恢复连接
				Consume 开始才发起 Join Group 请求 如果当前消费者加入后成为了 消费者组 leader,则还会进行 Rebalance 过程，从新分配
				组内每个消费组需要消费的 topic 和 partition，最后 Sync Group 后才开始消费
				具体信息见 https://github.com/lixd/kafka-go-example/issues/4
			*/
			err = cg.Consume(ctx, topicList, handler)
			if err == nil {

			}
			// 如果 context 被 cancel 了，那么退出
			if ctx.Err() != nil {

				return
			}
		}
	}()
	wg.Wait()
}
