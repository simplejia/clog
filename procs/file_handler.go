package procs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ROOT_DIR = "logs"
)

type FileParam struct {
	Excludes []string
}

var (
	fileMutex sync.RWMutex
	hourUses  = map[string]int64{}
	logfps    = map[string]*os.File{}
	loggers   = map[string]*log.Logger{}
)

func FileHandler(cate, subcate, body string, params map[string]interface{}) {
	var fileParam *FileParam
	bs, _ := json.Marshal(params)
	json.Unmarshal(bs, &fileParam)
	if fileParam != nil {
		for _, exclude := range fileParam.Excludes {
			if strings.Contains(body, exclude) {
				return
			}
		}
	}

	key := cate + "," + subcate

	fileMutex.RLock()
	logger := loggers[key]
	hourUse := hourUses[key]

	now := time.Now()
	if now.Unix() >= hourUse {
		fileMutex.RUnlock()
		fileMutex.Lock()

		logfp := logfps[key]
		if logfp != nil {
			logfp.Close()
			delete(logfps, key)
		}

		nx := now.Unix() + 3600
		hourUse = time.Unix(nx-nx%3600, 0).Unix()
		hourUses[key] = hourUse

		dir := path.Join(ROOT_DIR, cate, strconv.Itoa(now.Day()))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Printf("FileHandler() mkdir error %v\n", err)
				fileMutex.Unlock()
				return
			}
		}

		filename := fmt.Sprintf("log%s_%d", subcate, now.Hour())
		logfp, err := os.OpenFile(
			path.Join(dir, filename),
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0644,
		)
		if err != nil {
			log.Printf("FileHandler() openfile error %v\n", err)
			fileMutex.Unlock()
			return
		}

		logger = log.New(logfp, "", log.Ldate|log.Ltime|log.Lmicroseconds)
		logfps[key] = logfp
		loggers[key] = logger
		fileMutex.Unlock()
		fileMutex.RLock()
	}

	logger.Println(body)
	fileMutex.RUnlock()
	return
}

func init() {
	RegisterHandler("file_handler", FileHandler)
}
