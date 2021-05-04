package self_report

type GetTokenResp struct {
	ID      string `json:"id"`
	Result  string `json:"result"`
	Message string `json:"message"`
}

type ConfigOpt struct {
	LogPath     string
	TokenUrl    string
	RandCodeUrl string
	RandCodeImg string
	LoginUrl    string

	QueryEntpTypeUrl     string
	QueryCheckItemUrl    string
	QueryCheckItemIDUrl  string
	ReportUrl            string
	QueryCheckInspectUrl string
	SaveOrUpdateUrl      string

	ListenAddr string
	WebToken   string

	User string
	Pass string
}

type LoginResp struct {
	AreaCode    string `json:"areaCode"`
	Code        string `json:"code"`
	DeptId      string `json:"deptId"`
	DeptName    string `json:"deptName"`
	FreeCount   int    `json:"freeCount"`
	GroupId     string `json:"groupId"`
	GroupName   string `json:"groupName"`
	Id          string `json:"id"`
	InnerId     string `json:"innerId"`
	Message     string `json:"message"`
	OldUserId   string `json:"oldUserId"`
	Qx          string `json:"qx"`
	RealName    string `json:"realName"`
	RegionName  string `json:"regionName"`
	ReturnStr   string `json:"returnStr"`
	SafetyDept  string `json:"safetyDept"`
	SessionFlag string `json:"sessionFlag"`
	UnitName    string `json:"unitName"`
	UnitType    string `json:"unitType"`
	UnitId      string `json:"unitId"`
	UserId      string `json:"userId"`
	UserName    string `json:"userName"`

	SelfcheckDefineID string `json:"selfcheckDefineID"`
	IsMajor           string `json:"isMajor"`
}

type CheckItems []*CheckItem

type CheckItem struct {
	CheckDocCount      string `json:"checkDocCount"`
	CreateBy           string `json:"createBy"`
	CreateTime         string `json:"createTime"`
	EntpType           string `json:"entpType"` //一级标题
	Id                 string `json:"id"`
	IsNext             int    `json:"isNext"`
	JcType             string `json:"jcType"`
	JcTypeName         string `json:"jcTypeName"`
	Levels             string `json:"levels"`
	MajorCheckDocCount string `json:"majorCheckDocCount"`
	ParentId           string `json:"parentId"`
	Type               int    `json:"type"`
	UpdateBy           string `json:"updateBy"`
	UpdateTime         string `json:"UpdateTime"`

	ChildList []*CheckItem `json:"childList"`
}

type QueryCheckItemIDResp struct {
	PageList  []*SelfCheckItem `json:"pageList"`
	PageNo    int              `json:"pageNo"`
	PageSize  int              `json:"pageSize"`
	TotalNum  int              `json:"totalNum"`
	TotalPage int              `json:"totalPage"`
}

type SelfCheckItem struct {
	Accord              string `json:"accord"`
	CheckContent        string `json:"checkContent"`
	CheckContentName    string `json:"checkContentName"`
	CheckContentNo      string `json:"checkContentNo"`
	CheckDate           string `json:"checkDate"`
	CheckEntry          string `json:"checkEntry"`
	CheckFrequency      string `json:"checkFrequency"`
	CheckFrequencyName  string `json:"checkFrequencyName"`
	CheckItemID         string `json:"checkItemID"`
	CheckItemNo         string `json:"checkItemNo"`
	CheckItemPid        string `json:"checkItemPid"`
	CheckMan            string `json:"checkMan"`
	CheckMethod         string `json:"checkMethod"`
	ClassifyId          string `json:"classifyId"`
	ClassifyName        string `json:"classifyName"`
	ContentNo           string `json:"contentNo"`
	CreateBy            string `json:"createBy"`
	CreateTime          string `json:"createTime"`
	DeptName            string `json:"deptName"`
	DocId               string `json:"docId"`
	EntpName            string `json:"entpName"`
	EntpTypeID          string `json:"entpTypeID"`
	EntpTypeNo          string `json:"entpTypeNo"`
	HazardId            string `json:"hazardId"`
	HazardName          string `json:"hazardName"`
	Id                  string `json:"id"`
	InnerId             string `json:"innerId"`
	InspectType         string `json:"inspectType"`
	IsCustom            string `json:"isCustom"`
	IsMajor             string `json:"isMajor"`
	IsValid             int    `json:"isValid"`
	Items               string `json:"items"`
	LastCheckDate       string `json:"lastCheckDate"`
	LastNeedCheckDate   string `json:"lastNeedCheckDate"`
	NextNeedCheckDate   string `json:"nextNeedCheckDate"`
	Num                 string `json:"num"`
	Position            string `json:"position"`
	Result              string `json:"result"`
	SelfcheckDefineID   string `json:"selfcheckDefineID"`
	SeqId               string `json:"seqId"`
	Stand               int    `json:"stand"`
	TroubleFrequency    string `json:"troubleFrequency"`
	TroubleFrequencyStr string `json:"troubleFrequencyStr"`
	UpdateBy            string `json:"updateBy"`
	UpdateTime          string `json:"updateTime"`
}

type ReportResp struct {
	ReportDate string `json:"reportDate"` //上报日期
	Reports    string `json:"reports"`    //上报次数
}

//某个整改项目,如果有隐患,则该隐患的整改内容
type QueryCheckInspectItem struct {
	CheckContent   string `json:"checkContent"`
	CheckContentNo int    `json:"checkContentNo"`
	CheckItemID    string `json:"checkItemID"`
	CreateTime     string `json:"createTime"`
	Createby       string `json:"createby"`
	DocId          string `json:"docId"`
	EntpTypeID     string `json:"entpTypeID"`
	Id             string `json:"id"`
	UpdateTime     string `json:"updateTime"`
	Updateby       string `json:"updateby"`
}

type SaveOrUpdateReq struct {
	Item *SelfCheckItem //自查项

	SelfCheckMan string `json:"selfCheckMan"` //检查人

	TroubleType            int                      `json:"troubleType"`            //是否有隐患,0:无隐患 1:一般隐患 2:重大隐患
	QueryCheckInspectItems []*QueryCheckInspectItem `json:"queryCheckInspectItems"` //如果有隐患,隐患内容,通过接口QueryCheckInspect查询得知
	FormationCause         string                   `json:"FormationCause"`         //隐患结果
	Consequence            string                   `json:"consequence"`            //隐患形成原因
	MeasuresControl        string                   `json:"measuresControl"`        //衡量控制
	ImproveStep            int                      `json:"improveStep"`            //改进策略,1:立即改进
}
