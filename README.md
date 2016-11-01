# [clog](http://github.com/simplejia/clog) (集中式日志收集服务)
## 实现初衷
* 实际项目中，服务会部署到多台服务器上去，机器本地日志不方便查看，通过集中收集日志到一台或两台机器上，日志以文件形式存在，按服务名，ip，日期，日志类型分别存储，这样查看日志时就方便多了
* 我们做服务时，经常需要添加一些跟业务逻辑无关的功能，比如按错误日志报警，上报数据用于统计等等，这些功能和业务逻辑混在一起，实在没有必要，有了clog，我们只需要发送有效的数据，然后就可把数据处理的工作留给clog去做

## 功能
* 发送日志至远程server主机，server可以配多台机器，api目前提供golang，c支持
* 根据配置(server/conf/conf.json)运行相关日志分析程序，目前已实现：日志输出，报警
* 输出日志文件按server/logs/{模块名}/log{dbg|err|info|war}/{day}/log{ip}{+}{sub}规则命名，最多保存30天日志

## 使用方法
* server机器
> 布署server服务：server/server，配置文件：server/conf/conf.json

* server服务建议用[cmonitor](http://github.com/simplejia/cmonitor)启动管理

## 注意
* api.go文件里定义了获取server服务addr方法
* server/conf/conf.json文件里，tpl定义模板，然后通过`$xxx`方式引用，目前支持的handler有：filehandler和alarmhandler，filehandler用来记录本地日志，alarmhandler用来发报警
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


## demo
* [api_test.go](http://github.com/simplejia/clog/tree/master/api_test.go)
* [demo](http://github.com/simplejia/wsp/tree/master/demo) (demo项目里有clog的使用例子)

## LICENSE
clog is licensed under the Apache Licence, Version 2.0
(http://www.apache.org/licenses/LICENSE-2.0.html)
