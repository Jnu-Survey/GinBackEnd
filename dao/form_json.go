package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type FormInfo struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	FormJson  string    `json:"form" gorm:"column:form" description:"压缩后的JSON信息"`
	Title     string    `json:"title" gorm:"column:title" description:"标题"`
	Tip       string    `json:"tip" gorm:"column:tip" description:"提示"`
	Out       int       `json:"out" gorm:"column:out" description:"外键"`
}

func (f *FormInfo) TableName() string {
	return "form_done_json"
}

// RecordJson 把压缩的Json进行记录
func (f *FormInfo) RecordJson(c *gin.Context, tx *gorm.DB, newStruct *FormInfo) (*FormInfo, error) {
	err := tx.WithContext(c).Table(f.TableName()).Create(&newStruct).Error
	if err != nil {
		return nil, err
	}
	return newStruct, nil
}

func (f *FormInfo) GetTitleAndTips(c *gin.Context, tx *gorm.DB, inWhere []int64) ([]FormInfo, error) {
	var infos []FormInfo
	result := tx.WithContext(c).Table(f.TableName()).Where(inWhere).Find(&infos)
	err := result.Error
	if err != nil {
		return nil, err
	}
	return infos, nil
}
