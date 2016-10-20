package clog

import (
	"github.com/simplejia/clog"
	"github.com/simplejia/clog/master/conf"
)

func init() {
	clog.Init("clog", "", conf.C.Log.Level, conf.C.Log.Mode)
}
