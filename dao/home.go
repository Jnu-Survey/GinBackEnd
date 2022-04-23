package dao

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type Home struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	Jump      string    `json:"jump" gorm:"column:jump" description:"跳转链接"`
	Img       string    `json:"img" gorm:"column:img" description:"图片链接"`
}

func (h *Home) TableName() string {
	return "home"
}

// GetAllInfo 查询所有信息
func (h *Home) GetAllInfo(c *gin.Context, tx *gorm.DB) ([]Home, error) {
	var homes []Home
	result := tx.WithContext(c).Table(h.TableName()).Find(&homes)
	err := result.Error
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("没有查找到相关数据")
	}
	return homes, nil
}
