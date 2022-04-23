package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type Login struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	AvatarUrl string    `json:"avatar_url" gorm:"column:avatar_url" description:"头像地址"`
	City      string    `json:"city" gorm:"column:city" description:"城市"`
	Country   string    `json:"country" gorm:"column:country" description:"国家"`
	Gender    int       `json:"gender" gorm:"column:gender" description:"性别"`
	NickName  string    `json:"nick_name" gorm:"column:nick_name" description:"昵称"`
	Province  string    `json:"province" gorm:"column:province" description:"省会"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	IsBan     int       `json:"is_ban" gorm:"column:is_ban" description:"是否在黑名单中 0:正常 1:禁止"`
	Identity  int       `json:"identity" gorm:"column:identity" description:"身份 0:普通 1:超级"`
	OpenId    string    `json:"open_id" gorm:"column:open_id" description:"固定的openId"` // 建立了索引
}

func (t *Login) TableName() string {
	return "user_info"
}

// GetInfoByOpenId 根据openId查询信息
func (t *Login) GetInfoByOpenId(c *gin.Context, tx *gorm.DB, openId string) (*Login, error) {
	login := &Login{}
	err := tx.WithContext(c).Table(t.TableName()).Where("open_id = ?", openId).First(login).Error
	if err != nil {
		return nil, err
	}
	return login, nil
}

// RegisterOne 进行注册
func (t *Login) RegisterOne(c *gin.Context, tx *gorm.DB, newStruct *Login) (*Login, error) {
	err := tx.WithContext(c).Table(t.TableName()).Create(&newStruct).Error
	if err != nil {
		return nil, err
	}
	return newStruct, nil
}

// UpdateStatus 更新状态
func (t *Login) UpdateStatus(c *gin.Context, tx *gorm.DB, newStruct *Login) (*Login, error) {
	err := tx.WithContext(c).Table(t.TableName()).Save(newStruct).Error
	if err != nil {
		return nil, err
	}
	return newStruct, nil
}

// GetUidBaseInfo 获取基本信息
func (t *Login) GetUidBaseInfo(c *gin.Context, tx *gorm.DB, uid string) (Login, error) {
	var info Login
	result := tx.WithContext(c).Table(t.TableName()).Where("id = ?", uid).First(&info)
	err := result.Error
	if err != nil {
		return info, err
	}
	return info, nil
}
