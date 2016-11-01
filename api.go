package clog

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/simplejia/utils"
)

var (
	Level     int
	Mode      int
	cate_dbg  string
	cate_war  string
	cate_err  string
	cate_info string
)

// 请赋值成自己的获取master addr的函数
var AddrFunc = func() string {
	return "127.0.0.1:28702"
}

func sendAgent(tube, content string) {
	conn, err := net.Dial("udp", AddrFunc())
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write([]byte(tube + "," + content))
}

func Init(module, subcate string, level int, mode int) {
	if strings.Contains(module, ",") || strings.Contains(subcate, ",") {
		panic("clog Init error, module or subcate contains ','")
	}

	cate_dbg = strings.Join([]string{module, "logdbg", utils.GetLocalIp(), subcate}, ",")
	cate_war = strings.Join([]string{module, "logwar", utils.GetLocalIp(), subcate}, ",")
	cate_err = strings.Join([]string{module, "logerr", utils.GetLocalIp(), subcate}, ",")
	cate_info = strings.Join([]string{module, "loginfo", utils.GetLocalIp(), subcate}, ",")

	Level = level
	Mode = mode
}

func Debug(format string, params ...interface{}) {
	if Level&1 != 0 {
		content := fmt.Sprintf(format, params...)
		if Mode&1 != 0 {
			log.Println(content)
		}
		if Mode&2 != 0 {
			sendAgent(cate_dbg, content)
		}
	}
}

func Warn(format string, params ...interface{}) {
	if Level&2 != 0 {
		content := fmt.Sprintf(format, params...)
		if Mode&1 != 0 {
			log.Println(content)
		}
		if Mode&2 != 0 {
			sendAgent(cate_war, content)
		}
	}
}

func Error(format string, params ...interface{}) {
	if Level&4 != 0 {
		content := fmt.Sprintf(format, params...)
		if Mode&1 != 0 {
			log.Println(content)
		}
		if Mode&2 != 0 {
			sendAgent(cate_err, content)
		}
	}
}

func Info(format string, params ...interface{}) {
	if Level&8 != 0 {
		content := fmt.Sprintf(format, params...)
		if Mode&1 != 0 {
			log.Println(content)
		}
		if Mode&2 != 0 {
			sendAgent(cate_info, content)
		}
	}
}
