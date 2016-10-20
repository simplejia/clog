// 数据收集服务.
// author: simplejia
// date: 2014/12/01
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"

	_ "github.com/simplejia/clog/master/clog"
	"github.com/simplejia/clog/master/conf"
	"github.com/simplejia/clog/master/procs"
)

var buf chan []byte = make(chan []byte, 1e6)

func main() {
	log.Println("main()")

	go recv()
	for i := 0; i < 50; i++ {
		go proc()
	}

	select {}
}

func recv() {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", conf.C.Port))
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

func proc() {
	for d := range buf {
		ss := bytes.SplitN(d, []byte{','}, 5)
		if len(ss) != 5 {
			continue
		}
		cate := string(bytes.Join(ss[:2], []byte{'/'}))
		subcate := string(ss[2])
		if len(ss[3]) > 0 {
			subcate += "+" + string(ss[3])
		}
		body := ss[4]
		procs.Doit(cate, subcate, body)
	}
}
