package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	. "snow/common"
	. "snow/config"
	"snow/internal/broadcast"
	"snow/internal/plumtree"
	"snow/util"
	"time"
)

func main() {
	util.DebugLog()
	////测试轮数
	rounds := 100
	benchmark(500, 4, rounds)

	fmt.Println("done!!!")
	// 主线程保持运行
	select {}
}
func benchmark(n int, k int, rounds int) {
	portList := make([]int, 0)
	//tool.ResetIPTable()
	time.Sleep(2 * time.Second)
	defer util.ResetIPTable()
	configPath := ""
	flag.StringVar(&configPath, "configPath", "./config/config.yml", "config file path")
	flag.Parse()
	//消息大小
	strLen := 100
	initPort := 40000
	testMode := []MsgType{EagerPush, GossipMsg, RegularMsg, ColoringMsg} //按数组中的顺序决定跑的时候的顺序
	serversAddresses := initAddress(n, initPort)
	util.Num = n
	util.InitPort = initPort
	msg := randomByteArray(strLen)

	//节点启动完之后再跑
	for _, mode := range testMode {
		//err := tool.ResetIPTable()
		//if err != nil {
		//	panic(err)
		//}
		serverList := make([]*plumtree.Server, 0)
		for i := 0; i < n; i++ {
			action := createAction(i + 1)
			f := func(config *Config) {
				config.Port = initPort + i
				config.FanOut = k
				config.DefaultServer = serversAddresses
			}
			config, err := NewConfig(configPath, f)
			if err != nil {
				panic(err)
				return
			}
			server, err := plumtree.NewServer(config, action)
			if err != nil {
				log.Println(err)
				return
			}
			serverList = append(serverList, server)
		}
		time.Sleep(1 * time.Second)
		for _, v := range serverList {
			v.StartHeartBeat()
		}
		time.Sleep(time.Duration(n/200) * time.Second)
		for i := range rounds {
			// 1秒一轮,节点可能还没有离开新的广播就发出了	4秒足够把消息广播到所有节点
			fmt.Printf("=== %d =====\n", i)
			time.Sleep(1000 * time.Millisecond)

			if mode == RegularMsg {
				serverList[0].RegularMessage(msg, UserMsg)
			} else if mode == ColoringMsg {
				serverList[0].ColoringMessage(msg, UserMsg)
			} else if mode == GossipMsg {
				serverList[0].GossipMessage(msg, UserMsg)
			} else if mode == EagerPush {
				serverList[0].PlumTreeBroadcast(msg, UserMsg)
			}
			if i != 0 && i%10 == 0 {
				dport := i
				port := serverList[dport].Config.Port
				portList = append(portList, port)
				////tool.DisableNode(port)
				//time.Sleep(1000 * time.Millisecond)
				serverList[dport].IsClosed.Store(true)
				serverList[dport].HeartbeatService.Stop()
				serverList[dport].UdpServer.Close()
				//serverList[dport].Close()
				//serverList[dport].ApplyLeave()
			}

		}
		time.Sleep(5 * time.Second)
		for _, v := range serverList {
			f := false
			for _, port := range portList {
				if v.Config.Port == port {
					f = true
				}
			}
			if f == true {
				continue
			}
			v.IsClosed.Store(true)
			v.HeartbeatService.Stop()
			v.UdpServer.Close()
		}
		for _, v := range serverList {
			v.Close()

		}
		//err = tool.ResetIPTable()
		//if err != nil {
		//	panic(err)
		//}
		time.Sleep(5 * time.Second)

	}
}

// 编号从0开始
func createAction(num int) broadcast.Action {
	syncAction := func(bytes []byte) bool {
		//随机睡眠时间，百分之5的节点是掉队者节点
		if util.IntHash(num)%20 == 0 {
			time.Sleep(1 * time.Second)
		} else {
			randInt := util.RandInt(10, 200)
			time.Sleep(time.Duration(randInt) * time.Millisecond)
		}

		return true
	}
	asyncAction := func(bytes []byte) {
		//s := string(bytes)
		//fmt.Println("这里是异步处理消息的逻辑：", s)
	}
	reliableCallback := func(isConverged bool) {
		fmt.Println("这里是：可靠消息回调------------------------------", isConverged)
	}
	action := broadcast.Action{
		SyncAction:       &syncAction,
		AsyncAction:      &asyncAction,
		ReliableCallback: &reliableCallback,
	}
	return action
}
func randomByteArray(length int) []byte {
	rand.Seed(time.Now().UnixNano()) // 设置随机种子
	bytes := make([]byte, length)

	for i := range bytes {
		bytes[i] = byte(rand.Intn(26) + 'a') // 生成 a-z 之间的随机字符
	}

	return bytes
}
func initAddress(n int, port int) []string {
	strings := make([]string, 0)
	for i := 0; i < n; i++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port+i)
		strings = append(strings, addr)
	}
	return strings
}
