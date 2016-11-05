// 数据收集服务.
// author: simplejia
// date: 2014/12/01
package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/simplejia/clog"
	"github.com/simplejia/clog/server/conf"
	"github.com/simplejia/clog/server/procs"
)

type s struct {
	cate    string
	subcate string
	body    string
}

var (
	buf   = make(chan string, 1e6)
	tubes = make(map[string]chan *s)
)

func init() {
	clog.Init("clog", "", conf.C.Clog.Level, conf.C.Clog.Mode)
}

func main() {
	log.Println("main()")

	go recv()
	go dispatch()

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
		case buf <- string(request[:readLen]):
		default:
		}
	}
}

func dispatch() {
	for d := range buf {
		ss := strings.SplitN(d, ",", 5)
		if len(ss) != 5 {
			continue
		}
		cate := strings.Join(ss[:2], "/")
		subcate := ss[2]
		if len(ss[3]) > 0 {
			subcate += "+" + ss[3]
		}
		body := ss[4]

		k := cate + "," + subcate
		tube, ok := tubes[k]
		if !ok {
			tube = make(chan *s, 1e6)
			tubes[k] = tube
			go proc(k)
		}

		select {
		case tube <- &s{
			cate:    cate,
			subcate: subcate,
			body:    body,
		}:
		default:
		}
	}
}

func proc(k string) {
	tube := tubes[k]
	for d := range tube {
		procs.Doit(d.cate, d.subcate, d.body)
	}
}
