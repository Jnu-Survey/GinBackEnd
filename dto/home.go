package dto

type HomeOutput struct {
	BackgroundColor string `json:"backgroundColor" form:"backgroundColor" comment:"backgroundColor"` // backgroundColor
	Title           string `json:"title" form:"title" comment:"title"`                               // title
	Img             string `json:"img" form:"img" comment:"img"`                                     // img
}
