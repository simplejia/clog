package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/simplejia/clog/agent/conf"
)

var buf chan []byte = make(chan []byte, 1e6)

func main() {
	log.Println("main()")

	go recv()
	for i := 0; i < 50; i++ {
		go send()
	}

	select {}
}

func recv() {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", conf.C.Port))
	if err != nil {
		log.Fatalln("net.ResolveUDPAddr error:", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln("net.ListenUDP error:", err)
	}
	defer conn.Close()

	request := make([]byte, 1024*50)
	for {
		readLen, err := conn.Read(request)
		if err != nil || readLen <= 0 {
			continue
		}
		select {
		case buf <- append([]byte(nil), request[:readLen]...):
		default:
		}
	}
}

func send() {
	dur, err := time.ParseDuration(conf.C.Timeout)
	if err != nil {
		log.Fatalln("conf.C.Timeout error:", err)
	}
	conn, err := net.DialTimeout("udp", conf.C.Master, dur)
	if err != nil {
		log.Fatalln("net.DialTimeout error:", err)
	}
	defer conn.Close()

	for d := range buf {
		conn.SetWriteDeadline(time.Now().Add(dur))
		conn.Write(d)
	}
}
