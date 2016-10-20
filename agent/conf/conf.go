package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/simplejia/utils"
)

type Conf struct {
	Port    int
	Master  string
	Timeout string
}

var CONFS struct {
	Env  string
	Envs map[string]*Conf
}

var Env string
var C *Conf

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

	Env = CONFS.Env
	C = CONFS.Envs[Env]
	if C == nil {
		fmt.Println("env not right", Env)
		os.Exit(-1)
	}

	return
}
