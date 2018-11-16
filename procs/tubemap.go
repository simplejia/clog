package procs

import (
	"log"
	"runtime/debug"

	"github.com/simplejia/clog/conf"
)

var Handlers = map[string]HandlerFunc{}

type HandlerFunc func(cate, subcate, body string, params map[string]interface{})

func Doit(cate, subcate, body string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Doit() recover err: %v, cate: %s, subcate: %s, body: %s, stack: %s\n", err, cate, subcate, body, debug.Stack())
		}
	}()

	if handlers, ok := conf.Get().Procs[cate]; ok {
		for _, p := range handlers.([]*conf.ProcAction) {
			if p.Handler == "" {
				continue
			}
			if handler, ok := Handlers[p.Handler]; ok {
				handler(cate, subcate, body, p.Params)
			} else {
				log.Printf("Doit() p.Handler not right: %s\n", p.Handler)
			}
		}
	} else {
		// default
		FileHandler("default__"+cate, subcate, body, nil)
	}
	return
}

func RegisterHandler(name string, handler HandlerFunc) {
	Handlers[name] = handler
}
