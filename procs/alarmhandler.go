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
	LastTime  time.Time
	LastTexts []string
}

type AlarmParam struct {
	Sender    string
	Receivers []string
	Excludes  []string
}

func AlarmSplitWord(body string) (m map[string]bool) {
	m = map[string]bool{}

	for _, word := range strings.FieldsFunc(body, func(r rune) bool {
		switch r {
		case ' ', ',', ':', '{', '}', '"', '&':
			return true
		}
		return false
	}) {
		m[word] = true
	}

	return
}

func AlarmIsSimilar(src, dst string) bool {
	if src == dst {
		return true
	}

	shortM, longM := AlarmSplitWord(src), AlarmSplitWord(dst)
	if len(shortM) > len(longM) {
		shortM, longM = longM, shortM
	}

	if float64(len(shortM))/float64(len(longM)) < 0.8 {
		return false
	}

	l := 0
	for word := range shortM {
		if longM[word] {
			l++
		}
	}

	if l == 0 {
		return false
	}

	if float64(l)/float64(len(shortM)) > 0.8 {
		return true
	}

	return false
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
	if alarmstatLc, ok := lc.Get(tube); ok && alarmstatLc != nil {
		alarmstat = alarmstatLc.(*AlarmStat)
	} else {
		alarmstat = &AlarmStat{}
		lc.Set(tube, alarmstat, time.Hour)
	}

	diff := time.Since(alarmstat.LastTime)
	if diff < time.Second*30 {
		return
	}

	if diff < time.Minute*5 {
		for _, lastText := range alarmstat.LastTexts {
			if AlarmIsSimilar(lastText, body) {
				return
			}
		}
	}

	alarmstat.LastTime = time.Now()
	alarmstat.LastTexts = append(alarmstat.LastTexts, body)
	if curNum, maxNum := len(alarmstat.LastTexts), 5; curNum > maxNum {
		alarmstat.LastTexts = alarmstat.LastTexts[curNum-maxNum:]
	}

	AlarmFunc(alarmParam.Sender, alarmParam.Receivers, fmt.Sprintf("%s:%s", tube, body))
	return
}

func init() {
	RegisterHandler("alarmhandler", AlarmHandler)
}
