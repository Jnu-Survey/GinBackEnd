package services

import (
	"fmt"
	"testing"
)

func TestCode2Session(t *testing.T) {
	token := "rNf0+Ud5YABfVT7lihNMrT+1Q3ctl+A0HRu4b7aA6lgKz/61Usu2+N35YNZ5cSrc"
	openId, err := GetOpenIdFormToken(token)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(openId)
}
