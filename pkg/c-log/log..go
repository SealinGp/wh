package c_log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/**
c_log: customize log , rotate by day, filter by level

example:
xx.log -> xx.log.2021-04-11
xx.log.2021-04-11
xx.log.2021-04-10
*/

type LogLevel uint8

const (
	DateLayout        = "2006-01-02"
	HourMinuteMLayout = "15:04:05"
)

const (
	LogLevelNone LogLevel = iota
	LogLevelInfo
	LogLevelErr
)

var (
	logLevelErrPrefix = []byte("[E]")
	clog              *CLog
)

type CLogOptions struct {
	Flag     int
	Path     string
	LogLevel LogLevel
}

type CLog struct {
	*CLogOptions

	w io.WriteCloser
	sync.RWMutex

	closed  atomic.Value
	closeCh chan struct{}
}

func CLogInit(opt *CLogOptions) func() error {
	if opt.LogLevel < LogLevelNone {
		opt.LogLevel = LogLevelNone
	}

	clog = &CLog{
		closeCh: make(chan struct{}),
	}
	clog.closed.Store(false)
	clog.CLogOptions = opt

	log.SetFlags(opt.Flag)
	log.SetOutput(clog)

	//假如没有设置日志,说明不需要日志文件分割功能
	if opt.Path == "" {
		return nil
	}

	go clog.serve()
	return func() error {
		if closedBool, ok := clog.closed.Load().(bool); !ok || closedBool {
			return nil
		}

		close(clog.closeCh)
		clog.closed.Store(true)
		return nil
	}
}

//set out put
func (CLog *CLog) setOutput(w io.WriteCloser) {
	CLog.Lock()
	defer CLog.Unlock()

	//close old writer
	if CLog.w != nil {
		olaWriter := CLog.w
		defer olaWriter.Close()
	}

	//new writer
	CLog.w = w
}

//filter contents not belong to assigned log level
func (CLog *CLog) filterLevelLog(p []byte) []byte {
	//only print err log
	if CLog.LogLevel == LogLevelErr && bytes.Contains(p, logLevelErrPrefix) {
		return p
	}

	//<= info print all
	if CLog.LogLevel <= LogLevelInfo {
		return p
	}

	//drop
	return nil
}

//write
func (CLog *CLog) Write(p []byte) (n int, err error) {
	p = CLog.filterLevelLog(p)

	//drop
	if len(p) <= 0 {
		return 0, nil
	}

	//write to default
	if CLog.Path == "" {
		return os.Stderr.Write(p)
	}

	//writer nil, write to default
	CLog.RLock()
	if CLog.w == nil {
		CLog.RUnlock()
		return os.Stderr.Write(p)
	}
	CLog.RUnlock()

	//write to writer(file...)
	CLog.Lock()
	defer CLog.Unlock()
	return CLog.w.Write(p)
}

/**
1.打开新文件
2.删除软链接
3.新建软链接
4.设置日志新输出
5.关闭旧日志文件描述符,设置新日志文件描述符
*/
func (CLog *CLog) separateFile(now time.Time) {
	//打开文件
	dateFilePath := fmt.Sprintf("%v.%v", CLog.Path, now.Format(DateLayout))
	dateFileDesc, err := os.OpenFile(dateFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	//删除软链接
	err = os.Remove(CLog.Path)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	//指向软链接
	err = os.Symlink(dateFilePath, CLog.Path)
	if err != nil {
		panic(err)
	}

	//设置输出
	CLog.setOutput(dateFileDesc)
}

func (CLog *CLog) serve() {
	now := time.Now()
	CLog.separateFile(now)

	todayEndDur := getTodayEndSubNow(now)
	timer := time.NewTimer(todayEndDur)
	defer timer.Stop()

	for {
		select {
		case <-clog.closeCh:
			return
		default:
		}

		now = <-timer.C
		CLog.separateFile(now)
		timer = time.NewTimer(getTodayEndSubNow(now))
	}
}

/*
calculate the duration of now sub end of today
duration = now - Y-M-D 23:59:59

example:
now = 2021-04-11 09:55:00
duration = (2021-04-11 23:99:99 - 2021-04-11 09:55:00) + 1s
*/
func getTodayEndSubNow(now time.Time) time.Duration {
	nowStr := now.Format("2006-01-02 15:04:05")
	spaceIndex := strings.Index(nowStr, " ")
	nowStr = nowStr[:spaceIndex]

	nowStr = fmt.Sprintf("%v 23:59:59", nowStr)
	todayEnd, _ := time.Parse("2006-01-02 15:04:05", nowStr)
	return todayEnd.Sub(time.Now()) + 1*time.Second
}
