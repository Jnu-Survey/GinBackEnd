package services

import (
	"fmt"
	"testing"
)

func TestGetHomeCache(t *testing.T) {
	res, err := GetHomeCache("hello")
	fmt.Println(res)
	fmt.Println(err)
}
