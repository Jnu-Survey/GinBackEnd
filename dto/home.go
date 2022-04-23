package dto

type HomeOutput struct {
	Jump string `json:"jump" form:"jump" comment:"jump"` // jump
	Img  string `json:"img" form:"img" comment:"img"`    // img
}

type HomePartOutput struct {
	Swapping []HomeOutput `json:"swapping" form:"swapping" comment:"swapping"` // swapping
	Button   []HomeOutput `json:"button" form:"button" comment:"button"`       // button
}
