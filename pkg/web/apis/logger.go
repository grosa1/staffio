package apis

import (
	zlog "github.com/liut/staffio/pkg/log"
)

func logger() zlog.Logger {
	return zlog.GetLogger()
}
