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

	"github.com/simplejia/clog/api"
	"github.com/simplejia/clog/conf"
	"github.com/simplejia/clog/procs"
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

	clog.AddrFunc = func() (string, error) {
		return fmt.Sprintf("127.0.0.1:%d", conf.Get().Port), nil
	}
	c := conf.Get()
	clog.Init(c.Clog.Name, "", c.Clog.Level, c.Clog.Mode)
}

func main() {
	log.Println("main()")

	go udp()
	go ws()
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

	add(p.Cate, subcate, p.Body)

	w.WriteHeader(http.StatusOK)
	return
}

func ws() {
	http.HandleFunc("/clog/api", hget)

	addr := fmt.Sprintf(":%d", conf.Get().Port)
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

	if err := conn.SetReadBuffer(50 * 1024 * 1024); err != nil {
		log.Fatalln("conn.SetReadBuffer error:", err)
	}

	request := make([]byte, 1024*64)
	for {
		readLen, raddr, err := conn.ReadFrom(request)
		if err != nil || readLen <= 0 {
			continue
		}

		ss := strings.SplitN(string(request[:readLen]), ",", 5)
		if len(ss) != 5 {
			continue
		}
		cate, subcate := "", ""
		cate = strings.Join(ss[:2], "/") // module+level
		if str := ss[2]; str == "" {     // localip
			subcate, _, _ = net.SplitHostPort(raddr.String())
		} else {
			subcate = str
		}
		if str := ss[3]; str != "" { // subcate
			subcate += "+" + str
		}
		body := ss[4]

		add(cate, subcate, body)
	}
}

func add(cate, subcate, body string) {
	k := cate + "," + subcate
	tube, ok := tubes[k]
	if !ok {
		tube = make(chan *s, 1e5)
		tubes[k] = tube
		go proc(tube)
	}

	select {
	case tube <- &s{
		cate:    cate,
		subcate: subcate,
		body:    body,
	}:
	default:
		log.Println("add data full")
	}
}

func proc(tube chan *s) {
	for d := range tube {
		procs.Doit(d.cate, d.subcate, d.body)
	}
}
