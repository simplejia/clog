package procs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/simplejia/lc"

	"time"
)

// 请赋值成自己的报警处理函数
var AlarmFunc = func(sender string, receivers []string, text string) {
	params := map[string]interface{}{
		"Sender":    sender,
		"Receivers": receivers,
		"Text":      text,
	}
	json.NewEncoder(os.Stdout).Encode(params)
}

type AlarmStat struct {
	LastTime time.Time
	LastText string
}

type AlarmParam struct {
	Sender    string
	Receivers []string
	Excludes  []string
}

func AlarmHandler(cate, subcate, body string, params map[string]interface{}) {
	var alarmParam *AlarmParam
	bs, _ := json.Marshal(params)
	json.Unmarshal(bs, &alarmParam)
	if alarmParam == nil {
		log.Printf("AlarmHandler() params not right: %v\n", params)
		return
	}

	for _, exclude := range alarmParam.Excludes {
		if strings.Contains(body, exclude) {
			return
		}
	}

	tube := cate + "|" + subcate
	var alarmstat *AlarmStat
	if alarmstat_lc, ok := lc.Get(tube); ok && alarmstat_lc != nil {
		alarmstat = alarmstat_lc.(*AlarmStat)
	} else {
		alarmstat = &AlarmStat{}
		lc.Set(tube, alarmstat, time.Hour)
	}

	if time.Since(alarmstat.LastTime) < time.Second*30 ||
		(time.Since(alarmstat.LastTime) < time.Minute && strings.Compare(alarmstat.LastText, body) == 0) {
		return
	} else {
		alarmstat.LastTime = time.Now()
		alarmstat.LastText = body
	}

	AlarmFunc(alarmParam.Sender, alarmParam.Receivers, fmt.Sprintf("%s:%s", tube, body))
	return
}

func init() {
	RegisterHandler("alarmhandler", AlarmHandler)
}
