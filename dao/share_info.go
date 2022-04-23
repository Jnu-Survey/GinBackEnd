package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
	"wechatGin/dto"
	"wechatGin/public"
)

type ShareInfo struct {
	Id         int       `json:"id" gorm:"primary_key" description:"自增主键"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	ShareId    string    `json:"share_id" gorm:"column:share_id" description:"分享号"`
	Out        int       `json:"out" gorm:"column:out" description:"外键"`
	Parent     int       `json:"parent" gorm:"column:parent" description:"外键"`
	FromInfo   Form      `gorm:"foreignkey:Out"`
	ParentInfo Login     `gorm:"foreignkey:Parent"`
}

func (s *ShareInfo) TableName() string {
	return "share_info"
}

// IsExistShareInfo 判断是否存在信息
func (s *ShareInfo) IsExistShareInfo(c *gin.Context, tx *gorm.DB, shareId string) (string, error) {
	var temp ShareInfo
	res := tx.WithContext(c).Table(s.TableName()).
		Where("share_id = ?", shareId).
		First(&temp)
	if errors.Is(gorm.ErrRecordNotFound, res.Error) {
		return "", nil
	}
	if res.Error != nil {
		return "", errors.New("数据库错误")
	}
	return temp.ShareId, nil
}

// MakeRecord 记录
func (s *ShareInfo) MakeRecord(c *gin.Context, tx *gorm.DB, shareInfo *ShareInfo) error {
	err := tx.WithContext(c).Table(s.TableName()).Create(&shareInfo).Error
	if err != nil {
		return err
	}
	return nil
}

// GetParentOrderJsonInfo 拿到对应的表单的Json字符串
func (s *ShareInfo) GetParentOrderJsonInfo(c *gin.Context, tx *gorm.DB, shareId string) (dto.ShareTempInfo, error) {
	var finalRes dto.ShareTempInfo
	var temp ShareInfo
	res := tx.WithContext(c).Table(s.TableName()).
		Where("share_id = ?", shareId).
		First(&temp)
	if res.RowsAffected == 0 {
		return finalRes, errors.New("不存在分享号")
	}
	if res.Error != nil {
		return finalRes, errors.New("服务器错误")
	}
	err := tx.Model(&temp).Association("FromInfo").Find(&temp.FromInfo) // 联合查询
	if err != nil {
		return finalRes, errors.New("服务器错误")
	}
	var tempOrderInfo *Form
	tempOrderInfo, err = tempOrderInfo.GetFormDetailByOrderId(c, tx, temp.FromInfo.RandomId, true)
	if err != nil {
		return finalRes, errors.New("服务器错误")
	}
	decompress, err := public.JsonDeTool(tempOrderInfo.FormInfos.FormJson)
	if err != nil {
		return finalRes, errors.New("解压错误")
	}
	finalRes.JsonInfo = string(decompress)
	finalRes.Title = tempOrderInfo.FormInfos.Title
	return finalRes, nil
}
