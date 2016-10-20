package procs

import (
	"log"
	"runtime/debug"

	"github.com/simplejia/clog/master/conf"
)

var Handlers = map[string]HandlerFunc{}

type HandlerFunc func(cate, subcate string, content []byte, params map[string]interface{})

func Doit(cate, subcate string, content []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Doit() recover err: %v, cate: %s, subcate: %s, content: %s, stack: %s\n", err, cate, subcate, content, debug.Stack())
		}
	}()

	if handlers, ok := conf.C.Procs[cate]; ok {
		for _, p := range handlers.([]*conf.ProcAction) {
			if p.Handler == "" {
				continue
			}
			if handler, ok := Handlers[p.Handler]; ok {
				handler(cate, subcate, content, p.Params)
			} else {
				log.Printf("Doit() p.Handler not right: %s\n", p.Handler)
			}
		}
	} else {
		// default
		FileHandler("default__"+cate, subcate, content, nil)
	}
	return
}

func RegisterHandler(name string, handler HandlerFunc) {
	Handlers[name] = handler
}
