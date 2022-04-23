package dto

type EachDoneInfo struct {
	Id           string `json:"id" form:"id" comment:"id"`                                  // id
	JsonInfo     string `json:"json_info" form:"json_info" comment:"json_info"`             // json_info
	Create       string `json:"create" form:"create" comment:"create"`                      // create
	FormNickname string `json:"form_nickname" form:"form_nickname" comment:"form_nickname"` // form_nickname
}

type AllDoneInfoOutput struct {
	InitForm string         `json:"initForm" form:"initForm" comment:"initForm"` // initForm
	Infos    []EachDoneInfo `json:"infos" form:"infos" comment:"infos"`          // infos
}

type ShareTempInfo struct {
	JsonInfo string `json:"json_info" form:"json_info" comment:"json_info"` // json_info
	Title    string `json:"title" form:"title" comment:"title"`             // title
}

type CodeOutput struct {
	Path        string `json:"path" form:"path" comment:"path"`                         // path
	Params      string `json:"params" form:"params" comment:"params"`                   // params
	Base64Image string `json:"base64_image" form:"base64_image" comment:"base64_image"` // base64_image
}

type ExcelHeader struct {
	HeaderName  []string
	HeaderField []string
}

type JsonOne struct {
	Flag int
	Info []string
}

type EachData struct {
	Name string
	Num  int
}

type ParseEach struct {
	Kind  string
	Title string
	Data  []EachData
}

type ParseFormOutput struct {
	Header []string  `json:"header" form:"header" comment:"header"` // header
	Body   []JsonOne `json:"body" form:"body" comment:"body"`       // body
}
