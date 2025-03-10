package main

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("新客户端连接:", conn.RemoteAddr())
}

func main() {
	go s(":7070")
	go s(":7071")
	time.Sleep(3 * time.Second) // 等待服务器启动
	go main2()

	// 防止主函数退出
	for {
	}
}

func s(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("监听失败:", err)
		return
	}
	defer listener.Close()
	fmt.Println("服务器启动，监听端口", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("接受连接失败:", err)
			continue
		}
		go handleConnection(conn) // 使用 goroutine 处理每个客户端
	}
}

func main2() {
	// 指定本地地址
	localAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:7000")
	if err != nil {
		fmt.Println("解析本地地址失败:", err)
		return
	}

	// 创建 Dialer，设置 Control 函数以允许端口复用
	dialer := net.Dialer{
		LocalAddr: localAddr,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				// 设置 SO_REUSEADDR
				err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				if err != nil {
					fmt.Println("设置 SO_REUSEADDR 失败:", err)
				}
			})
		},
	}

	// 连接到第一个服务器
	conn, err := dialer.Dial("tcp", "127.0.0.1:7070")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()
	fmt.Println("连接成功:", conn.LocalAddr())

	// 连接到第二个服务器
	conn2, err := dialer.Dial("tcp", "127.0.0.1:7071")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn2.Close()
	fmt.Println("连接成功:", conn2.LocalAddr())
}
