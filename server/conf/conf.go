package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/simplejia/utils"
)

type ProcAction struct {
	Handler string
	Params  map[string]interface{}
}

type Conf struct {
	Port  int
	Procs map[string]interface{}
	Log   struct {
		Mode  int
		Level int
	}
}

var CONFS struct {
	Env  string
	Tpl  map[string]interface{}
	Envs map[string]*Conf
}

var Env string
var C *Conf

func ReplaceTpl(src string, tpl map[string]interface{}) (dst []byte) {
	oldnew := []string{}
	for k, v := range tpl {
		_v, _ := json.Marshal(v)
		oldnew = append(oldnew, "\"$"+k+"\"", string(_v))
	}

	r := strings.NewReplacer(oldnew...)
	dst = []byte(r.Replace(src))
	return
}

func init() {
	dir, _ := os.Getwd()
	fcontent, err := ioutil.ReadFile(filepath.Join(dir, "conf", "conf.json"))
	if err != nil {
		fmt.Println("conf.json not found")
		os.Exit(-1)
	}

	fcontent = utils.RemoveAnnotation(fcontent)
	if err := json.Unmarshal(fcontent, &CONFS); err != nil {
		fmt.Println("conf.json wrong format", err)
		os.Exit(-1)
	}

	fcontent = ReplaceTpl(string(fcontent), CONFS.Tpl)
	if err := json.Unmarshal(fcontent, &CONFS); err != nil {
		fmt.Println("conf.json wrong format", err)
		os.Exit(-1)
	}

	for env, CONF := range CONFS.Envs {
		if CONF == nil {
			continue
		}
		for k, proc := range CONF.Procs {
			var new_proc []*ProcAction
			_proc, _ := json.Marshal(proc)
			json.Unmarshal(_proc, &new_proc)
			CONF.Procs[k] = new_proc
		}
		CONFS.Envs[env] = CONF
	}

	Env = CONFS.Env
	C = CONFS.Envs[Env]
	if C == nil {
		fmt.Println("env not right", Env)
		os.Exit(-1)
	}
}
