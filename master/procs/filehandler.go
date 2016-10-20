package procs

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

const (
	ROOT_DIR = "logs"
)

func FileHandler(cate, subcate string, content []byte, params map[string]interface{}) {
	now := time.Now()
	dir := path.Join(ROOT_DIR, cate, strconv.Itoa(now.Day()))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("FileHandler() mkdir error %v\n", err)
			return
		}
	}

	filename := "log" + subcate
	f, err := os.OpenFile(
		path.Join(dir, filename),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Printf("FileHandler() openfile error %v\n", err)
		return
	}
	defer f.Close()

	_, err = f.Write(append(content, fmt.Sprintf("[%02d:%02d:%02d]\n", now.Hour(), now.Minute(), now.Second())...))
	if err != nil {
		log.Printf("FileHandler() writefile error %v\n", err)
		return
	}
	return
}

func timerproc() {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("timerproc() recover %v\n", err)
				}
			}()

			tick := time.Tick(time.Hour * 24)
			for {
				select {
				case <-tick:
					day := strconv.Itoa(time.Now().Add(time.Hour * 24).Day())
					filepath.Walk(ROOT_DIR, func(path string, info os.FileInfo, err error) (reterr error) {
						if err != nil {
							return
						}
						if !info.IsDir() {
							return
						}
						if info.Name() != day {
							return
						}
						os.RemoveAll(path)
						return
					})
				}
			}
		}()
		time.Sleep(time.Second * 5)
	}
}

func init() {
	RegisterHandler("filehandler", FileHandler)
	go timerproc()
}
