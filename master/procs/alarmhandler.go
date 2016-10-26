package procs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

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

// go1.6之后，map在并发情况下会exit掉，但基于实际并发不高，exit的可能很小
var AlarmRegexps = make(map[string]*struct {
	IncludesComp []*regexp.Regexp
	ExcludesComp []*regexp.Regexp
})

// go1.6之后，map在并发情况下会exit掉，但基于实际并发不高，exit的可能很小
var AlarmStats = make(map[string]*struct {
	LastTime time.Time
	LastText []byte
})

// go1.6之后，map在并发情况下会exit掉，但基于实际并发不高，exit的可能很小
var AlarmParams = make(map[string]*struct {
	Sender    string
	Receivers []string
	Includes  []string
	Excludes  []string
})

func AlarmHandler(cate, subcate string, content []byte, params map[string]interface{}) {
	paramsT, ok := AlarmParams[cate]
	if !ok {
		bs, _ := json.Marshal(params)
		json.Unmarshal(bs, &paramsT)
		AlarmParams[cate] = paramsT
	}

	var includesComp, excludesComp []*regexp.Regexp
	if v, ok := AlarmRegexps[cate]; !ok {
		includesComp = make([]*regexp.Regexp, 0)
		for _, vv := range paramsT.Includes {
			includesComp = append(includesComp, regexp.MustCompile(vv))
		}
		excludesComp = make([]*regexp.Regexp, 0)
		for _, vv := range paramsT.Excludes {
			excludesComp = append(excludesComp, regexp.MustCompile(vv))
		}
		AlarmRegexps[cate] = &struct {
			IncludesComp []*regexp.Regexp
			ExcludesComp []*regexp.Regexp
		}{includesComp, excludesComp}
	} else {
		includesComp = v.IncludesComp
		excludesComp = v.ExcludesComp
	}

	result := false
	for _, excludeComp := range excludesComp {
		result = excludeComp.Match(content)
		if result {
			break
		}
	}
	if result {
		return
	}

	for _, includeComp := range includesComp {
		result = includeComp.Match(content)
		if result {
			break
		}
	}
	if !result {
		return
	}

	tube := cate + "|" + subcate
	alarmstat, ok := AlarmStats[tube]
	if !ok {
		alarmstat = &struct {
			LastTime time.Time
			LastText []byte
		}{}
		AlarmStats[tube] = alarmstat
	}

	if time.Since(alarmstat.LastTime) < time.Second*30 ||
		(time.Since(alarmstat.LastTime) < time.Minute && bytes.Compare(alarmstat.LastText, content) == 0) {
		return
	} else {
		alarmstat.LastTime = time.Now()
		alarmstat.LastText = content
	}

	AlarmFunc(paramsT.Sender, paramsT.Receivers, fmt.Sprintf("%s:%s", tube, content))
	return
}

func init() {
	RegisterHandler("alarmhandler", AlarmHandler)
}
