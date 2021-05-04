package self_report

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	c_log "github.com/SealinGp/go-lib/c-log"

	uuid "github.com/satori/go.uuid"
)

type SafetyApi struct {
	sync.RWMutex
	httpCli   *http.Client
	cfg       *ConfigOpt
	loginInfo *LoginResp
}

type SafetyApiOpt struct {
	HttpCli *http.Client
	Cfg     *ConfigOpt
}

func NewSafetyApi(opt *SafetyApiOpt) *SafetyApi {
	api := &SafetyApi{
		httpCli: opt.HttpCli,
		cfg:     opt.Cfg,
	}
	return api
}

func (safetyApi *SafetyApi) GetCfg() *ConfigOpt {
	return safetyApi.cfg
}

//登录
func (safetyApi *SafetyApi) Login(paramMap *sync.Map) *LoginResp {
	safetyApi.reqRandCodeUrl()
	tokenID := safetyApi.getTokenID()
	randCode := GetRandCode()
	if tokenID == "" {
		return nil
	}

	body := map[string]string{
		"userName":     StrEnc(base64.StdEncoding.EncodeToString([]byte(safetyApi.cfg.User)), "1", "2", "3"),
		"password":     StrEnc(base64.StdEncoding.EncodeToString([]byte(safetyApi.cfg.Pass)), "1", "2", "3"),
		"randCode":     fmt.Sprintf("%d", randCode),
		"xzxk":         "",
		"tooken":       tokenID,
		"initAreaCode": "4306",
	}

	if initAreaCode, ok := paramMap.Load("initAreaCode"); ok {
		if initAreaCodeStr, ok := initAreaCode.(string); ok && initAreaCodeStr != "" {
			body["initAreaCode"] = initAreaCodeStr
		}
	}

	reqBodyData, err := json.Marshal(body)
	if err != nil {
		c_log.E("Marshal body failed. err:%v", err)
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.LoginUrl, bytes.NewReader(reqBodyData))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("randCode_", "4e7d960b1fe8481f9a6fe0185921f873")

	if paramValue, ok := paramMap.Load("randCode_"); ok {
		if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
			req.Header.Set("randCode_", paramValueStr)
		}
	}

	//开始登录
	resp, err := safetyApi.httpCli.Do(req)
	if err != nil {
		c_log.E("post failed. url:%v, err:%v", safetyApi.cfg.LoginUrl, err)
		return nil
	}
	defer resp.Body.Close()

	respBodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c_log.E("read body failed. err:%v", err)
		return nil
	}

	loginResp := &LoginResp{}
	err = json.Unmarshal(respBodyData, loginResp)
	if err != nil {
		c_log.E("Unmarshal loginResp failed. err:%v", err)
		return nil
	}

	if loginResp.Code != "10" {
		c_log.E("login req failed. data:%s", respBodyData)
		loginResp.InnerId = "700EF8026D9AE6D6E0530100007F4FE8"
		loginResp.UnitType = "21"
		loginResp.UnitId = "700EF8026D9AE6D6E0530100007F4FE8"
		loginResp.UserId = "700EF8026D9CE6D6E0530100007F4FE8"

		loginResp.SelfcheckDefineID = "7595255C359F7018E0530100007FE0E2"
		loginResp.IsMajor = ""
		loginResp.DeptId = "700EF8026D9BE6D6E0530100007F4FE8"

		if paramValue, ok := paramMap.Load("innerId"); ok {
			if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
				loginResp.InnerId = paramValueStr
			}
		}
		if paramValue, ok := paramMap.Load("unitType"); ok {
			if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
				loginResp.UnitType = paramValueStr
			}
		}
		if paramValue, ok := paramMap.Load("unitId"); ok {
			if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
				loginResp.UnitId = paramValueStr
			}
		}
		if paramValue, ok := paramMap.Load("userId"); ok {
			if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
				loginResp.UserId = paramValueStr
			}
		}

		if paramValue, ok := paramMap.Load("selfcheckDefineID"); ok {
			if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
				loginResp.SelfcheckDefineID = paramValueStr
			}
		}
		if paramValue, ok := paramMap.Load("isMajor"); ok {
			if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
				loginResp.IsMajor = paramValueStr
			}
		}

	}

	safetyApi.Lock()
	safetyApi.loginInfo = loginResp
	safetyApi.Unlock()
	return loginResp
}

//获取tokenid
func (safetyApi *SafetyApi) getTokenID() string {
	resp, err := safetyApi.httpCli.Get(safetyApi.cfg.TokenUrl)
	if err != nil {
		c_log.E("getToken failed. err:%v", err)
		return ""
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c_log.E("read body failed. err:%v", err)
		return ""
	}

	getTokenResp := &GetTokenResp{}
	err = json.Unmarshal(data, getTokenResp)
	if err != nil {
		c_log.E("Unmarshal failed. data:%s, err:%v", data, err)
		return ""
	}

	c_log.I("getToken response:%s", data)
	return getTokenResp.ID
}

//验证码
func (safetyApi *SafetyApi) reqRandCodeUrl() {
	randCodeUrl := fmt.Sprintf("%v?version=%v", safetyApi.cfg.RandCodeUrl, MakeTimestamp())
	resp, err := safetyApi.httpCli.Get(randCodeUrl)
	if err != nil {
		c_log.E("reqRandCodeUrl failed. err:%v", err)
		return
	}
	defer resp.Body.Close()

	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		c_log.E("jpeg decode failed. err:%v", err)
		return
	}

	file, err := os.OpenFile(safetyApi.cfg.RandCodeImg, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		c_log.E("OpenFile failed. file:%v, err:%v", safetyApi.cfg.RandCodeImg, err)
		return
	}
	defer file.Close()

	jpegOpt := &jpeg.Options{
		Quality: 100,
	}
	err = jpeg.Encode(file, img, jpegOpt)
	if err != nil {
		c_log.E("jpeg encode failed. err:%v", err)
		return
	}
}

//cookie
func (safetyApi *SafetyApi) newCookie(paramMap *sync.Map) *http.Cookie {
	newCookie := &http.Cookie{
		Name:    "3B4A770C3AF55A22884CD9C5F462DF3E",
		Value:   "700EF8026D9CE6D6E0530100007F4FE81617438380137",
		Path:    "/",
		Domain:  "safety.yjgl.sz.gov.cn",
		Expires: time.Now().Add(time.Hour * 12),
	}

	if paramValue, ok := paramMap.Load("cookie-name"); ok {
		if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
			newCookie.Name = paramValueStr
		}
	}
	if paramValue, ok := paramMap.Load("cookie-value"); ok {
		if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
			newCookie.Value = paramValueStr
		}
	}
	if paramValue, ok := paramMap.Load("cookie-path"); ok {
		if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
			newCookie.Path = paramValueStr
		}
	}
	if paramValue, ok := paramMap.Load("cookie-domain"); ok {
		if paramValueStr, ok := paramValue.(string); ok && paramValueStr != "" {
			newCookie.Domain = paramValueStr
		}
	}

	return newCookie
}

//查询目录
func (safetyApi *SafetyApi) QueryEntpType(paramMap *sync.Map) ([]*CheckItem, error) {
	val := &url2.Values{}
	val.Set("innerIdId", safetyApi.loginInfo.InnerId)
	val.Set("selfcheckDefineID", safetyApi.loginInfo.SelfcheckDefineID)
	//0: 编制计划
	//"":隐患自查
	if paramValue, ok := paramMap.Load("isCusotm"); ok {
		if paramStr, ok := paramValue.(string); ok {
			val.Set("isCusotm", paramStr)
		}
	}

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.QueryEntpTypeUrl, strings.NewReader(val.Encode()))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil, err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.AddCookie(safetyApi.newCookie(paramMap))

	resp := make([]*CheckItem, 0)
	err = safetyApi.doRequest(req, &resp)
	if err != nil {
		c_log.E("doRequest failed. err:%v", err)
		return nil, err
	}

	return resp, nil
}

//查询章节
func (safetyApi *SafetyApi) QueryCheckItem(paramMap *sync.Map, item *CheckItem) ([]*CheckItem, error) {
	val := &url2.Values{}
	val.Set("innerId", safetyApi.loginInfo.InnerId)
	val.Set("selfcheckDefineID", safetyApi.loginInfo.SelfcheckDefineID)
	val.Set("entpTypeId", item.Id)
	val.Set("isMajor", safetyApi.loginInfo.IsMajor)

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.QueryCheckItemUrl, strings.NewReader(val.Encode()))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil, err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.AddCookie(safetyApi.newCookie(paramMap))

	resp := make([]*CheckItem, 0)
	err = safetyApi.doRequest(req, &resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

//查询某章的某小节
func (safetyApi *SafetyApi) QueryCheckItemID(paramMap *sync.Map, checkItem *CheckItem, pageNum int, pageSize int) (*QueryCheckItemIDResp, error) {
	val := &url2.Values{}
	val.Set("innerId", safetyApi.loginInfo.InnerId)
	val.Set("selfcheckDefineID", safetyApi.loginInfo.SelfcheckDefineID)
	val.Set("major", safetyApi.loginInfo.IsMajor)
	val.Set("reportYear", fmt.Sprintf("%d", time.Now().Year()))
	val.Set("reportMonth", fmt.Sprintf("%d", time.Now().Month()))
	val.Set("deptId", safetyApi.loginInfo.DeptId)
	val.Set("id", checkItem.Id)
	val.Set("major", safetyApi.loginInfo.IsMajor)
	val.Set("pageNo", fmt.Sprintf("%d", pageNum))
	val.Set("pageSize", fmt.Sprintf("%d", pageSize))
	level, _ := strconv.ParseInt(checkItem.Levels, 10, 64)
	val.Set("levle", fmt.Sprintf("%d", level))

	reportQuarter := 1
	if time.Now().Month() >= 4 && time.Now().Month() <= 6 {
		reportQuarter = 2
	}
	if time.Now().Month() >= 7 && time.Now().Month() <= 9 {
		reportQuarter = 3
	}
	if time.Now().Month() >= 10 && time.Now().Month() <= 12 {
		reportQuarter = 4
	}
	val.Set("reportQuarter", fmt.Sprintf("%d", reportQuarter))

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.QueryCheckItemIDUrl, strings.NewReader(val.Encode()))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil, err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.AddCookie(safetyApi.newCookie(paramMap))

	resp := &QueryCheckItemIDResp{}
	err = safetyApi.doRequest(req, resp)
	if err != nil {
		c_log.E("doRequest QueryCheckItemID failed. err:%v", err)
		return nil, err
	}

	return resp, nil
}

/**
上报
req:
innerId: 700EF8026D9AE6D6E0530100007F4FE8
schinspectdefineID: 7595255C359F7018E0530100007FE0E2
reportDate: 2021-04-06
deptId: 700EF8026D9BE6D6E0530100007F4FE8

resp:
{"reports":"1","reportDate":"2021-04-06 21:28:47"}
*/
func (safetyApi *SafetyApi) Report(paramMap *sync.Map) (*ReportResp, error) {
	val := &url2.Values{}
	val.Set("innerIdId", safetyApi.loginInfo.InnerId)
	val.Set("schinspectdefineID", safetyApi.loginInfo.SelfcheckDefineID)
	val.Set("reportDate", time.Now().Format("2006-01-02"))
	val.Set("deptId", safetyApi.loginInfo.DeptId)

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.ReportUrl, strings.NewReader(val.Encode()))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil, err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.AddCookie(safetyApi.newCookie(paramMap))

	resp := &ReportResp{}
	err = safetyApi.doRequest(req, resp)
	if err != nil {
		c_log.E("doRequest Report failed. err:%v", err)
		return nil, err
	}

	return resp, nil
}

//点击无隐患|有隐患
func (safetyApi *SafetyApi) SaveOrUpdate(paramMap *sync.Map, reqProto *SaveOrUpdateReq) (map[string]interface{}, error) {
	val := &url2.Values{}
	val.Set("innerId", safetyApi.loginInfo.InnerId)
	val.Set("createBy", reqProto.SelfCheckMan)
	val.Set("updateBy", reqProto.SelfCheckMan)
	val.Set("selfcheckDefineID", safetyApi.loginInfo.SelfcheckDefineID)
	val.Set("checkFrequency", reqProto.Item.CheckFrequency)
	val.Set("entpTypeID", reqProto.Item.EntpTypeID)
	val.Set("checkItemPid", reqProto.Item.CheckItemPid)
	val.Set("checkItemID", reqProto.Item.CheckItemID)
	val.Set("deptId", safetyApi.loginInfo.DeptId)

	val.Set("troubleType", fmt.Sprintf("%d", reqProto.TroubleType)) //0: 无隐患 1:一般隐患 2:重大隐患
	//有隐患时以下字段必填
	if reqProto.TroubleType > 0 {
		guuid := uuid.NewV1()
		val.Set("guuid", guuid.String())
		//有隐患,则设置公司名为创建人
		val.Set("createBy", reqProto.Item.EntpName)
		val.Set("updateBy", reqProto.Item.EntpName)
		val.Set("baseCase", reqProto.Item.CheckContent)
		val.Set("position", "隐患部位")
		val.Set("inspectContent", "隐患基本情况")
		val.Set("formationCause", reqProto.FormationCause)              //形成原因
		val.Set("consequence", reqProto.Consequence)                    //结果
		val.Set("measuresControl", reqProto.MeasuresControl)            //衡量控制
		val.Set("improveStep", fmt.Sprintf("%d", reqProto.ImproveStep)) //改进策略
		val.Set("planFinishDate", time.Now().Format("2006-01-02"))      //计划改进时间
		val.Set("checkMan", reqProto.SelfCheckMan)                      //改进人
		val.Set("checkContentId", reqProto.Item.DocId)

		//?
		val.Set("supervisionLevel", "")
		val.Set("supervisionNum", "")
		val.Set("supervisionUnit", "")
		val.Set("summaryReport", "")
		val.Set("summaryReport", "")
		val.Set("responsibilityPlace", "")
		val.Set("responsibilityDate", "")
		val.Set("measuresControlInfo", "")
		val.Set("measuresControlDate", "")
		val.Set("rectificationFundInfo", "")
		val.Set("rectificationFund", "")
		val.Set("rectificationFundDate", "")
		val.Set("improveContent", "")
		val.Set("closeFlag", "0")
		val.Set("closeDate", "")
		val.Set("improveManager", "")
		val.Set("improverTel", "")
		val.Set("id", "")
		val.Set("troubleNo", "")

		if len(reqProto.QueryCheckInspectItems) > 0 {
			for _, queryCheckInspectItem := range reqProto.QueryCheckInspectItems {
				val.Set("checkNegative[]", queryCheckInspectItem.CheckContent)
			}
		}

	} else {
		val.Set("selfCheckMan", reqProto.SelfCheckMan)
		val.Set("checkDate", time.Now().Format("2006-01-02"))
		val.Set("inspectType", reqProto.Item.CheckContent)
		val.Set("docId", reqProto.Item.DocId)
		val.Set("checkContentIds", "")
	}

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.SaveOrUpdateUrl, strings.NewReader(val.Encode()))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil, err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.AddCookie(safetyApi.newCookie(paramMap))

	resp := make(map[string]interface{})
	err = safetyApi.doRequest(req, &resp)
	if err != nil {
		c_log.E("SaveOrUpdate doRequest failed. err:%v", err)
		return nil, err
	}

	return resp, nil
}

//有隐患的内容接口
func (safetyApi *SafetyApi) QueryCheckInspect(paramMap *sync.Map, checkItem *SelfCheckItem) ([]*QueryCheckInspectItem, error) {
	val := &url2.Values{}
	val.Set("checkItemId", checkItem.CheckItemID)

	req, err := http.NewRequest(http.MethodPost, safetyApi.cfg.QueryCheckInspectUrl, strings.NewReader(val.Encode()))
	if err != nil {
		c_log.E("NewRequest failed. err:%v", err)
		return nil, err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.AddCookie(safetyApi.newCookie(paramMap))

	resp := make([]*QueryCheckInspectItem, 0)
	err = safetyApi.doRequest(req, &resp)
	if err != nil {
		c_log.E("doRequest Report failed. err:%v", err)
		return nil, err
	}

	return resp, nil
}

func (safetyApi *SafetyApi) doRequest(req *http.Request, respStru interface{}) error {
	resp, err := safetyApi.httpCli.Do(req)
	if err != nil {
		c_log.E("httpCli do failed. err:%v", err)
		return err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c_log.E("read doRequest resp failed. err:%v", err)
		return err
	}

	err = json.Unmarshal(respData, respStru)
	if err != nil {
		c_log.E("Unmarshal failed. respData:%s", respData)
		if bytes.Contains(respData, []byte("timg.jpg")) {
			return errors.New("404")
		}
		return err
	}

	return nil
}
