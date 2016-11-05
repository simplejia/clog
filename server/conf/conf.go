package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/simplejia/utils"
)

type ProcAction struct {
	Handler string
	Params  map[string]interface{}
}

type Conf struct {
	Port  int
	Tpl   map[string]interface{}
	Procs map[string]interface{}
	Clog  *struct {
		Mode  int
		Level int
	}
}

var (
	Envs map[string]*Conf
	Env  string
	C    *Conf
)

func replaceTpl(src string, tpl map[string]interface{}) (dst []byte) {
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
	flag.StringVar(&Env, "env", "prod", "set env")
	var conf string
	flag.StringVar(&conf, "conf", "", "set custom conf")
	flag.Parse()

	dir, _ := os.Getwd()
	fcontent, err := ioutil.ReadFile(filepath.Join(dir, "conf", "conf.json"))
	if err != nil {
		fmt.Println("conf.json not found")
		os.Exit(-1)
	}

	fcontent = utils.RemoveAnnotation(fcontent)
	if err := json.Unmarshal(fcontent, &Envs); err != nil {
		fmt.Println("conf.json wrong format:", err)
		os.Exit(-1)
	}

	C = Envs[Env]
	if C == nil {
		fmt.Println("env not right:", Env)
		os.Exit(-1)
	}

	cs, _ := json.Marshal(C)
	cs = replaceTpl(string(cs), C.Tpl)
	if err := json.Unmarshal(cs, &C); err != nil {
		fmt.Println("conf.json wrong format:", err)
		os.Exit(-1)
	}

	for k, proc := range C.Procs {
		var new_proc []*ProcAction
		_proc, _ := json.Marshal(proc)
		json.Unmarshal(_proc, &new_proc)
		C.Procs[k] = new_proc
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("conf not right:", err)
				os.Exit(-1)
			}
		}()
		ccs := strings.Split(conf, "::")
		for _, cs := range ccs {
			pos := strings.Index(cs, "=")
			if pos == -1 {
				continue
			}
			name, value := strings.TrimSpace(cs[:pos]), strings.TrimSpace(cs[pos+1:])

			rv := reflect.Indirect(reflect.ValueOf(C))
			for _, field := range strings.Split(name, ".") {
				rv = reflect.Indirect(rv.FieldByName(strings.Title(field)))
			}
			switch rv.Kind() {
			case reflect.String:
				rv.SetString(value)
			case reflect.Bool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					panic(err)
				}
				rv.SetBool(b)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					panic(err)
				}
				rv.SetInt(i)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				u, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					panic(err)
				}
				rv.SetUint(u)
			case reflect.Float32, reflect.Float64:
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					panic(err)
				}
				rv.SetFloat(f)
			}
		}
	}()

	fmt.Printf("Env: %s\nC: %s\n", Env, utils.Iprint(C))

	return
}
