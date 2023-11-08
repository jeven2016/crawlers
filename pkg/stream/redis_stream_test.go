package stream

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
	"testing"
	"time"
)

const redisStream = "sourceStream"
const consumersGroup = "group"

func TestRedisStream(t *testing.T) {

	//send
	sendMsg()
	//
	//if receiveMsg() {
	//	return
	//}

	select {}
}

func receiveMsg() bool {
	//--------------------------------------------------------------------------------------------
	client := getClient()
	defer client.Close()

	//创建一个消费group, 不同消费相同的消息
	groups, err := client.XInfoGroups(context.Background(), redisStream).Result()
	if err != nil {
		log.Println(err)
		return true
	}

	var groupExists bool
	for _, g := range groups {
		if g.Name == consumersGroup {
			println("group exists")
			groupExists = true
			break
		}
	}

	if !groupExists {
		err = client.XGroupCreate(context.Background(), redisStream, consumersGroup, "0").Err()
		if err != nil {
			log.Println(err)
		}
	}
	go func() {
		client2 := getClient()
		defer client2.Close()
		read("thr1: ", client2)
	}()

	go func() {
		client3 := getClient()
		defer client3.Close()
		read("thr3: ", client3)
	}()

	go func() {
		client4 := getClient()
		defer client4.Close()
		read("thr4: ", client4)
	}()
	return false
}

func sendMsg() {
	go func() {
		client := getClient()
		defer client.Close()

		send(client)

	}()
}

func send(client *redis.Client) {
	tick := time.Tick(2 * time.Second)
	var i int
	for {
		if i > 5 {
			return
		}
		select {
		case <-tick:
			err := client.XAdd(context.Background(), &redis.XAddArgs{Stream: redisStream,
				NoMkStream: false,  // * 默认false,当为false时,key不存在，会新建
				MaxLen:     100000, // * 指定stream的最大长度,当队列长度超过上限后，旧消息会被删除，只保留固定长度的新消息
				Approx:     false,  // * 默认false,当为true时,模糊指定stream的长度
				ID:         "*",    // 消息 id，我们使用 * 表示由 redis 生成
				// MinID: "id",            // * 超过阈值，丢弃设置的小于MinID消息id【基本不用】
				// Limit: 1000,            // * 限制长度【基本不用】
				Values: map[string]string{
					"name": "wang" + strconv.Itoa(i),
				}}).Err()
			if err != nil {
				log.Println(err)
				return
			}
			println("send a message" + strconv.Itoa(i))
			i++
		}

	}
}

func read(prefix string, client *redis.Client) {
	for {
		entries, err := client.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    consumersGroup,
			Consumer: "abc2",

			Streams: []string{redisStream, ">"},
			Count:   2,
			Block:   0,
			//NoAck:   false,
		}).Result()
		if err != nil {
			println(err.Error())
			return
		}
		for i := 0; i < len(entries[0].Messages); i++ {
			messageID := entries[0].Messages[i].ID
			values := entries[0].Messages[i].Values
			name := fmt.Sprintf("%v", values["name"])
			println(prefix + " " + name)
			client.XAck(context.Background(), "redisStream", "consumer", messageID)

		}
	}
}

func getClient() (client *redis.Client) {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "pwd", // no password set
		DB:       0,     // use default DB
		PoolSize: 3,     // 默认一个CPU 10个连接
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Unable to connect to Redis", err)
	}
	return client
}
