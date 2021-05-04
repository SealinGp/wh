package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	c_log "github.com/SealinGp/go-lib/c-log"
	self_report "github.com/SealinGp/wh/self-report"
)

var (
	configPath string
	httpServer *http.Server
)

type SelfReportApp struct {
	cfg *self_report.ConfigOpt
}

func main() {
	flag.StringVar(&configPath, "c", "./config/self_report.yml", "config path")
	flag.Parse()

	app := &SelfReportApp{}

	//config init
	cfg, err := self_report.ConfigInit(configPath)
	if err != nil {
		log.Fatalf("[E] config init failed. err:%v", err)
		return
	}
	app.cfg = cfg

	//log init
	logCf := c_log.CLogInit(&c_log.CLogOptions{
		Path:     app.cfg.LogPath,
		Flag:     log.Lshortfile | log.Ltime,
		LogLevel: c_log.LEVEL_ERR,
	})
	defer logCf()

	//http server init
	errCh := make(chan error)
	safetyApi := self_report.NewSafetyApi(&self_report.SafetyApiOpt{
		HttpCli: &http.Client{
			Timeout: time.Second * 10,
		},
		Cfg: app.cfg,
	})

	httpServer = &http.Server{
		Addr: app.cfg.ListenAddr,
		Handler: &WebServer{
			safetyApi: safetyApi,
		},
	}
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err := httpServer.Shutdown(ctx)
		if err != nil {
			c_log.E("shutdown failed. err:%v", err)
		}
		c_log.I("recevied signal exit. sig:%v", sig)
	case <-errCh:
		signal.Stop(sigCh)
	}
}

type WebServer struct {
	safetyApi *self_report.SafetyApi
	reqMap    sync.Map
}

func (webServer *WebServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/report" {
		if req.Header.Get("token") != webServer.safetyApi.GetCfg().WebToken {
			http.Error(rw, "auth failed.", http.StatusForbidden)
			return
		}

		reqData, err := ioutil.ReadAll(req.Body)
		if err != nil {
			c_log.E("read req body failed. err:%v", err)
			http.Error(rw, "read req body failed.", http.StatusForbidden)
			return
		}

		reqDataMap := make(map[string]interface{})
		err = json.Unmarshal(reqData, &reqDataMap)
		if err == nil {
			for k, v := range reqDataMap {
				webServer.reqMap.Store(k, v)
			}
		}

		resp := webServer.StartReport()
		respData, _ := json.Marshal(resp)

		rw.Header().Set("Content-type", "application/json")
		rw.WriteHeader(resp.Code)
		rw.Write(respData)
		return
	}

	http.NotFound(rw, req)
}

type StartReportResp struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Msg  string `json:"msg"`
}

func (webServer *WebServer) StartReport() *StartReportResp {
	startReportResp := &StartReportResp{
		Code: 200,
		Data: "",
		Msg:  "",
	}
	webServer.safetyApi.Login(&webServer.reqMap)

	//查询隐患列表
	tabItems, err := webServer.safetyApi.QueryEntpType(&webServer.reqMap)
	if err != nil {
		startReportResp.Code = http.StatusInternalServerError
		startReportResp.Msg = fmt.Sprintf("QueryEntpType failed. err:%v", err)
		c_log.E("%v", startReportResp.Msg)
		return startReportResp
	}

	totalReportItems := make([]*self_report.SelfCheckItem, 0)
	for _, checkItem := range tabItems {
		pageNum := 1
		pageSize := 100

		tableContents, err := webServer.safetyApi.QueryCheckItemID(&webServer.reqMap, checkItem, pageNum, pageSize)
		if err != nil {
			c_log.E("QueryCheckItemID failed. checkItem:%+v, err:%v", checkItem, err)
			continue
		}

		for _, content := range tableContents.PageList {
			totalReportItems = append(totalReportItems, content)
		}
	}

	totalReportCount := len(totalReportItems)
	reportFailedCount := 0
	reportSuccessCount := 0

	//生成随机隐患项目
	reportFailedNum := 3
	if paramVal, ok := webServer.reqMap.Load("reportFailedNum"); ok {
		if paramInt, ok := paramVal.(int); ok {
			reportFailedNum = paramInt
		}
	}
	rand.Seed(time.Now().UnixNano())
	selfCheckIndexMap := make(map[int]bool, reportFailedNum)
	for i := 0; i < reportFailedNum; i++ {
		index := rand.Intn(totalReportCount)
		selfCheckIndexMap[index] = true
	}

	//开始自查(1s执行一次,防止被封ip)
	c_log.I("开始自查. 随机有隐患数量:%v", len(selfCheckIndexMap))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	//测试上报条数
	totalReportItems = totalReportItems[:1]
	for selfCheckIndex, reportItem := range totalReportItems {
		<-ticker.C

		reqProto := &self_report.SaveOrUpdateReq{
			SelfCheckMan: "黄春芳",
			Item:         reportItem,
			TroubleType:  0,
		}

		//如果该项目为有隐患项目
		if _, ok := selfCheckIndexMap[selfCheckIndex]; ok {
			//查询所有隐患列表
			queryCheckInspects, err := webServer.safetyApi.QueryCheckInspect(&webServer.reqMap, reportItem)

			//取一条隐患内容上报
			if err == nil {
				if len(queryCheckInspects) > 0 {
					reqProto.QueryCheckInspectItems = queryCheckInspects[:1]
					reqProto.TroubleType = 1
					queryCheckInspectsData, _ := json.Marshal(queryCheckInspects)
					c_log.I("有隐患内容 %s", queryCheckInspectsData)
				}
			}

			if err != nil {
				c_log.E("查询有隐患内容失败. err:%v", err)
			}
		}

		//点击保存,默认无隐患
		_, err := webServer.safetyApi.SaveOrUpdate(&webServer.reqMap, reqProto)
		if err != nil {
			reqProtoData, _ := json.Marshal(reqProto)
			c_log.E("自查保存失败. reqData:%s, err:%v", reqProtoData, err)
			reportFailedCount++
			continue
		}
		reportSuccessCount++
	}

	c_log.I("开始上报")
	reportResp, err := webServer.safetyApi.Report(&webServer.reqMap)
	reportDate, reports := "", ""
	if err != nil {
		startReportResp.Msg = "点击上报按钮失败"
		c_log.E("%v", startReportResp.Msg)
	} else {
		reportDate = reportResp.ReportDate
		reports = reportResp.Reports
	}

	startReportResp.Msg += fmt.Sprintf(" 总自查数量:%v, 自查成功数量:%v, 自查失败数量:%v, 上报日期:%v, 上报次数:%v",
		totalReportCount, reportSuccessCount, reportFailedCount, reportDate, reports)
	return startReportResp
}
