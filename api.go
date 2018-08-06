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
	cate_busi string
)

// 请赋值成自己的获取master addr的函数
var AddrFunc = func() (string, error) {
	return "127.0.0.1:28702", nil
}

func sendAgent(tube, content string) {
	addr, err := AddrFunc()
	if err != nil {
		return
	}
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	seg := 65000
	for bpos, epos, l := 0, 0, len(content); bpos < l; bpos += seg {
		epos = bpos + seg
		if epos > l {
			epos = l
		}
		out := tube + "," + content[bpos:epos]
		conn.Write([]byte(out))
	}
}

func Init(module, subcate string, level int, mode int) {
	if strings.Contains(module, ",") || strings.Contains(subcate, ",") {
		panic("clog Init error, module or subcate contains ','")
	}

	cate_dbg = strings.Join([]string{module, "logdbg", utils.LocalIp, subcate}, ",")
	cate_war = strings.Join([]string{module, "logwar", utils.LocalIp, subcate}, ",")
	cate_err = strings.Join([]string{module, "logerr", utils.LocalIp, subcate}, ",")
	cate_info = strings.Join([]string{module, "loginfo", utils.LocalIp, subcate}, ",")
	cate_busi = strings.Join([]string{module, "logbusi_%s", utils.LocalIp, subcate}, ",")

	Level = level
	Mode = mode
}

func Debug(format string, params ...interface{}) {
	if Level&1 != 0 {
		content := fmt.Sprintf(format, params...)
		if Mode&1 != 0 {
			log.Println("[DEBUG]", content)
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
			log.Println("[WARN]", content)
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
			log.Println("[ERROR]", content)
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
			log.Println("[INFO]", content)
		}
		if Mode&2 != 0 {
			sendAgent(cate_info, content)
		}
	}
}

func Busi(sub string, format string, params ...interface{}) {
	content := fmt.Sprintf(format, params...)
	if Mode&1 != 0 {
		log.Println("[BUSI]", sub, content)
	}
	if Mode&2 != 0 {
		sendAgent(fmt.Sprintf(cate_busi, sub), content)
	}
}
