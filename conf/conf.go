package conf

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/utils"
)

type ProcAction struct {
	Handler string
	Params  map[string]interface{}
}

type Conf struct {
	Port  int
	Tpl   map[string]json.RawMessage
	Procs map[string]interface{}
	Clog  *struct {
		Name  string
		Mode  int
		Level int
	}
}

func Get() *Conf {
	return C.Load().(*Conf)
}

func Set(c *Conf) {
	C.Store(c)
}

var (
	Env string
	C   atomic.Value
)

func replaceTpl(src string, tpl map[string]json.RawMessage) (dst []byte) {
	oldnew := []string{}
	for k, v := range tpl {
		oldnew = append(oldnew, "\"$"+k+"\"", string(v))
	}

	r := strings.NewReplacer(oldnew...)
	dst = []byte(r.Replace(src))
	return
}

func reloadConf(content []byte) {
	lastbody := content

	for {
		time.Sleep(time.Second * 3)

		body, err := getcontents()
		if err != nil || len(body) == 0 {
			clog.Error("getcontents err: %v, body: %s", err, body)
			continue
		}

		if bytes.Compare(lastbody, body) == 0 {
			continue
		}

		if err := parse(body); err != nil {
			clog.Error("parse err: %v, body: %s", err, body)
			continue
		}

		if err := savecontents(body); err != nil {
			clog.Error("savecontents err: %v, body: %s", err, body)
			continue
		}

		lastbody = body
	}
}

func getcontents() (fcontent []byte, err error) {
	dir := "conf"
	for i := 0; i < 3; i++ {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		dir = filepath.Join("..", dir)
	}
	fcontent, err = ioutil.ReadFile(filepath.Join(dir, "conf.json"))
	if err != nil {
		return
	}
	return
}

func savecontents(fcontent []byte) (err error) {
	dir := "conf"
	for i := 0; i < 3; i++ {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		dir = filepath.Join("..", dir)
	}
	err = ioutil.WriteFile(filepath.Join(dir, "conf.json"), fcontent, 0644)
	if err != nil {
		return
	}
	return
}

func parse(fcontent []byte) (err error) {
	fcontent = utils.RemoveAnnotation(fcontent)

	var envs map[string]*Conf
	if err = json.Unmarshal(fcontent, &envs); err != nil {
		return
	}

	c := envs[Env]
	if c == nil {
		return fmt.Errorf("env not right: %s", Env)
	}

	cs, _ := json.Marshal(c)
	cs = replaceTpl(string(cs), c.Tpl)
	if err := json.Unmarshal(cs, &c); err != nil {
		return fmt.Errorf("conf.json wrong format:", err)
	}

	for k, proc := range c.Procs {
		var new_proc []*ProcAction
		_proc, _ := json.Marshal(proc)
		if err := json.Unmarshal(_proc, &new_proc); err != nil {
			return fmt.Errorf("conf.json wrong format(procs):", err)
		}
		c.Procs[k] = new_proc
	}

	Set(c)

	log.Printf("Env: %s\nC: %s\n", Env, utils.Iprint(c))
	return
}

func init() {
	flag.StringVar(&Env, "env", "prod", "set env")
	flag.Parse()

	fcontent, err := getcontents()
	if err != nil {
		log.Printf("get conf file contents error: %v\n", err)
		os.Exit(-1)
	}

	err = parse(fcontent)
	if err != nil {
		log.Printf("parse conf file error: %v\n", err)
		os.Exit(-1)
	}

	go reloadConf(fcontent)
}
