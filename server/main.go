// 数据收集服务.
// author: simplejia
// date: 2014/12/01
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/simplejia/clog/server/conf"
	"github.com/simplejia/clog/server/procs"
	"github.com/simplejia/lc"
)

type s struct {
	cate    string
	subcate string
	body    string
}

var tubes = make(map[string]chan *s)

func init() {
	lc.Init(1e5)
}

func main() {
	log.Println("main()")

	go ws()
	go udp()
	select {}
}

func hget(w http.ResponseWriter, r *http.Request) {
	fun := "main.hget"

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s ReadAll err: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var p *struct {
		Cate    string
		Subcate string
		Body    string
	}
	err = json.Unmarshal(body, &p)
	if err != nil || p == nil {
		log.Printf("%s Unmarshal err: %v, req: %s\n", fun, err, body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subcate := p.Subcate
	if subcate == "" {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		subcate = host
	}

	k := p.Cate + "," + subcate
	tube, ok := tubes[k]
	if !ok {
		tube = make(chan *s, 1e5)
		tubes[k] = tube
		go proc(k)
	}

	select {
	case tube <- &s{
		cate:    p.Cate,
		subcate: subcate,
		body:    p.Body,
	}:
	default:
	}

	w.WriteHeader(http.StatusOK)
	return
}

func ws() {
	http.HandleFunc("/clog/conf", conf.Cgi)
	http.HandleFunc("/clog/api", hget)

	addr := fmt.Sprintf("%s:%d", "0.0.0.0", conf.Get().AdminPort)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln("net.ListenAndServe error:", err)
	}
}

func udp() {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", conf.Get().Port))
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

		ss := strings.SplitN(string(request[:readLen]), ",", 5)
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
			tube = make(chan *s, 1e5)
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
