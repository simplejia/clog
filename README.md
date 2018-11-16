[中文 README](#中文)


# [clog](http://github.com/simplejia/clog) (Centralized log collection service)
## Original Intention
* In actual projects, the service will be deplored to many servers and it is not convenient to view local logs. But it is convenient to view logs via collecting logs to one or two computers. The logs exist in the form of files and are respectively stored according to their service names, ip addresses, dates and types.
* In our service, we often need add some features that have nothing to do with our business logic, such as alarm logs, reporting data for statistics and so on. It is really not necessary mix these features and business logic. While having clog, we just need send effective data and leave clog to deal with the data.

## Functions
* Sending logs to remote host which can be equipped with one or two servers. 
* Api currently offers golang, c, php.
* According to the configuration (server/conf/conf.json), clog will execute relevant log handlers.  And log output and alarm process have been implemented so far.
* The output log files is named according to server/logs/{module name}/log{dbg|err|info|war}/{day}/log{ip}{+}{sub}. And the logs can be saved at most 30 days.

## Usage
* server
> Deploring server service: server/server, configuring files: server/conf/conf.json

* Recommending using [cmonitor](http://github.com/simplejia/cmonitor) to manage

## Notice
* api.go has defined AddrFunc variable to get server address, you can redefine it.
* In server/conf/conf.json files, tpl defines template and then is cited via $xxx. Currently supported handlers are filehandler and alarmhandler. Filehandler is used to record local logs and alarmhandler is used to send alarms. you can redefine the parameters in configuration files via passing custom env and conf parameter, for example: ./server -env dev -conf='port=8080;clog.mode=1'(multiple parameters are seperated by `;`
* For alarmhandler, relevant parameter configuration can be referred to params field. In current alarms, it just provides output logs. In actual use, we should replace it with our own alarm logic by redefine procs.AlarmFunc. We can create a new go file under the directory of server/procs, as following:
```
package procs

import (
	"encoding/json"
	"os"
)

func init() {
	// replace it with your own alarm function
	AlarmFunc = func(sender string, receivers []string, text string) {
		params := map[string]interface{}{
			"Sender":    sender,
			"Receivers": receivers,
			"Text":      text,
		}
		json.NewEncoder(os.Stdout).Encode(params)
	}
}
```
* Alarmhandler has anti disturbance control logic, the same content will not be alarmed again within one minute. The interval between two alarms is not less than 30 seconds.
* If you want to add a new handler, you just need create a new go file under the directory of server/procs, as following:
```
package procs

func XxxHandler(cate, subcate string, content []byte, params map[string]interface{}) {
}

func init() {
	RegisterHandler("xxxhandler", XxxHandler)
}
```

> Handler used in product is as following:(having implemented the function of sending data to multiple subscribers after receiving it)

```
package procs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/utils"
)

type TransParam struct {
	Nodes []*struct {
		Addr     string
		AddrType string
		Retry    int
		Host     string
		Cgi      string
		Params   string
		Method   string
		Timeout  string
	}
}

func TransHandler(cate, subcate, body string, params map[string]interface{}) {
	clog.Info("TransHandler() Begin Trans: %s, %s, %s", cate, subcate, body)

	var transParam *TransParam
	bs, _ := json.Marshal(params)
	json.Unmarshal(bs, &transParam)
	if transParam == nil {
		clog.Error("TransHandler() params not right: %v", params)
		return
	}

	arrs := []string{body}
	json.Unmarshal([]byte(body), &arrs)
	for pos, str := range arrs {
		arrs[pos] = url.QueryEscape(str)
	}

	for _, node := range transParam.Nodes {
		addr := node.Addr
		ps := map[string]string{}
		values, _ := url.ParseQuery(fmt.Sprintf(node.Params, utils.Slice2Interface(arrs)...))
		for k, vs := range values {
			ps[k] = vs[0]
		}

		timeout, _ := time.ParseDuration(node.Timeout)

		headers := map[string]string{
			"Host": node.Host,
		}

        gpp := &utils.GPP{
            Uri: fmt.Sprintf("http://%s/%s", addr, strings.TrimPrefix(node.Cgi, "/")),
            Timeout: timeout,
            Headers: headers,
            Params: ps,
        }

		for step := -1; step < node.Retry; step++ {
			var (
				body []byte
				err  error
			)
			switch node.Method {
			case "get":
				body, err = utils.Get(gpp)
			case "post":
				body, err = utils.Post(gpp)
			}

			if err != nil {
				clog.Error("TransHandler() http error, err: %v, body: %s, gpp: %v, step: %d", err, body, gpp, step)
				continue
			} else {
				clog.Info("TransHandler() http success, body: %s, gpp: %v", body, gpp)
				break
			}
		}
	}

	return
}

func init() {
	RegisterHandler("transhandler", TransHandler)
}
```

> Corresponding configuration is as following（configuring a template field in server/conf/conf.json）：

```
"tpl": { 
    "trans": [
        {   
            "handler": "transhandler",
            "params": {
                "nodes": [
                    {   
                        "addr": "127.0.0.1:80",
                        "addrType": "ip",
                        "host": "xx.xx.com",
                        "cgi": "/c/a",
                        "params": "a=1&b=%s&c=%s",
                        "method": "post",
                        "retry": 2,
                        "timeout": "50ms"
                    }   
                ]   
            }   
        }   
    ]   
},
"procs": {
    "demo/logbusi_trans": "$trans"
}
```

> the clog.Busi() function in api.go will be used: clog.Busi("trans", "xxx", "xxx", ...)

## demo
* [api_test.go](http://github.com/simplejia/clog/tree/master/api_test.go)
* [demo](http://github.com/simplejia/wsp/tree/master/demo) (has examples for using clog)

---
中文
===

# [clog](http://github.com/simplejia/clog) (集中式日志收集服务)
## 实现初衷
* 实际项目中，服务会部署到多台服务器上去，机器本地日志不方便查看，通过集中收集日志到一台或两台机器上，日志以文件形式存在，按服务名，ip，日期，日志类型分别存储，这样查看日志时就方便多了
* 我们做服务时，经常需要添加一些跟业务逻辑无关的功能，比如按错误日志报警，上报数据用于统计等等，这些功能和业务逻辑混在一起，实在没有必要，有了clog，我们只需要发送有效的数据，然后就可把数据处理的工作留给clog去做

## 功能
* 发送日志至远程server主机，server可以配多台机器，api目前提供golang，c，php支持
* 根据配置(server/conf/conf.json)运行相关日志分析程序，目前已实现：日志输出，报警
* 输出日志文件按server/logs/{模块名}/log{dbg|err|info|war}/{day}/log{ip}{+}{sub}规则命名，最多保存30天日志

## 使用方法
* server机器
> 布署server服务：server/server，配置文件：server/conf/conf.json

* server服务建议用[cmonitor](http://github.com/simplejia/cmonitor)启动管理

## 注意
* api.go文件里定义了获取server服务addr方法
* server/conf/conf.json文件里，tpl定义模板，然后通过`$xxx`方式引用，目前支持的handler有：filehandler和alarmhandler，filehandler用来记录本地日志，alarmhandler用来发报警，可以通过传入自定义的env及conf参数来重定义配置文件里的参数，如：./server -env dev -conf='port=8080;clog.mode=1'，多个参数用`;`分隔
* 对于alarmhandler，相关参数配置见params，目前的报警只是打印日志，实际实用，应替换成自己的报警处理逻辑，重新赋值procs.AlarmFunc就可以了，可以在server/procs目录下新建一个go文件，如下示例：
```
package procs

import (
	"encoding/json"
	"os"
)

func init() {
	// 请替换成你自己的报警处理函数
	AlarmFunc = func(sender string, receivers []string, text string) {
		params := map[string]interface{}{
			"Sender":    sender,
			"Receivers": receivers,
			"Text":      text,
		}
		json.NewEncoder(os.Stdout).Encode(params)
	}
}
```
* alarmhandler有防骚扰控制逻辑，相同内容，一分钟内不再报，两次报警不少于30秒，以上限制和日志文件一一对应
* 如果想添加新的handler，只需在server/procs目录下新建一个go文件，如下示例：
```
package procs

func XxxHandler(cate, subcate string, content []byte, params map[string]interface{}) {
}

func init() {
	RegisterHandler("xxxhandler", XxxHandler)
}
```

> 一个实际生产环境使用到的handler如下：（实现了接收数据后分发给多个订阅者的功能）

```
package procs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/utils"
)

type TransParam struct {
	Nodes []*struct {
		Addr     string
		AddrType string
		Retry    int
		Host     string
		Cgi      string
		Params   string
		Method   string
		Timeout  string
	}
}

func TransHandler(cate, subcate, body string, params map[string]interface{}) {
	clog.Info("TransHandler() Begin Trans: %s, %s, %s", cate, subcate, body)

	var transParam *TransParam
	bs, _ := json.Marshal(params)
	json.Unmarshal(bs, &transParam)
	if transParam == nil {
		clog.Error("TransHandler() params not right: %v", params)
		return
	}

	arrs := []string{body}
	json.Unmarshal([]byte(body), &arrs)
	for pos, str := range arrs {
		arrs[pos] = url.QueryEscape(str)
	}

	for _, node := range transParam.Nodes {
		addr := node.Addr
		ps := map[string]string{}
		values, _ := url.ParseQuery(fmt.Sprintf(node.Params, utils.Slice2Interface(arrs)...))
		for k, vs := range values {
			ps[k] = vs[0]
		}

		timeout, _ := time.ParseDuration(node.Timeout)

		headers := map[string]string{
			"Host": node.Host,
		}

        gpp := &utils.GPP{
            Uri: fmt.Sprintf("http://%s/%s", addr, strings.TrimPrefix(node.Cgi, "/")),
            Timeout: timeout,
            Headers: headers,
            Params: ps,
        }

		for step := -1; step < node.Retry; step++ {
			var (
				body []byte
				err  error
			)
			switch node.Method {
			case "get":
				body, err = utils.Get(gpp)
			case "post":
				body, err = utils.Post(gpp)
			}

			if err != nil {
				clog.Error("TransHandler() http error, err: %v, body: %s, gpp: %v, step: %d", err, body, gpp, step)
				continue
			} else {
				clog.Info("TransHandler() http success, body: %s, gpp: %v", body, gpp)
				break
			}
		}
	}

	return
}

func init() {
	RegisterHandler("transhandler", TransHandler)
}
```

> 相应配置如下（server/conf/conf.json里配上一个模板）：

```
"tpl": { // 模板定义
    "trans": [
        {   
            "handler": "transhandler",
            "params": {
                "nodes": [
                    {   
                        "addr": "127.0.0.1:80",
                        "addrType": "ip",
                        "host": "xx.xx.com",
                        "cgi": "/c/a",
                        "params": "a=1&b=%s&c=%s",
                        "method": "post",
                        "retry": 2,
                        "timeout": "50ms"
                    }   
                ]   
            }   
        }   
    ]   
},
"procs": {
    "demo/logbusi_trans": "$trans"
}
```

> 使用时调用api.go里提供的clog.Busi函数: clog.Busi("trans", "xxx", "xxx", ...)

## demo
* [api_test.go](http://github.com/simplejia/clog/tree/master/api_test.go)
* [demo](http://github.com/simplejia/wsp/tree/master/demo) (demo项目里有clog的使用例子)

