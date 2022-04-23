package dao

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type Form struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	RandomId  string    `json:"random_id" gorm:"column:random_id" description:"表单编号"`
	Uid       int       `json:"uid" gorm:"uid" description:"用户id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	Status    int       `json:"status" gorm:"column:status" description:"0:创建 1:完成提交"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"0:正常 1:已经删除"`
	IsBan     int       `json:"is_ban" gorm:"column:is_ban" description:"0:正常填写 1:不能填写"`
	FormInfos FormInfo  `gorm:"FOREIGNKEY:Out;ASSOCIATION_FOREIGNKEY:Id"` // 外键
}

func (f *Form) TableName() string {
	return "form_uid"
}

// AddForm2Uid 将表单与id建立映射
func (f *Form) AddForm2Uid(tx *gorm.DB, newStruct *Form) (*Form, error) {
	err := tx.Table(f.TableName()).Create(&newStruct).Error
	if err != nil {
		return nil, err
	}
	return newStruct, nil
}

// IsExistAndJudgeStatus 检查order是不是在消息队列里面写入了以及判断状态
func (f *Form) IsExistAndJudgeStatus(c *gin.Context, tx *gorm.DB, order string, temp *Form) (*Form, error) {
	res := tx.WithContext(c).Table(f.TableName()).Where("random_id = ?", order).First(&temp)
	if res.Error != nil { // 错误
		return nil, res.Error
	}
	if temp.Status != 0 { // 防止别人直接怼接口自己修改
		return nil, errors.New("无法修改")
	}
	return temp, nil
}

// RewriteStatus 重写状态
func (f *Form) RewriteStatus(c *gin.Context, tx *gorm.DB, temp *Form) error {
	temp.Status = 1
	err := tx.WithContext(c).Table(f.TableName()).Save(&temp).Error
	if err != nil {
		return err
	}
	return nil
}

// GetAllDoneInfo 查询此用户下的所有已经确定了的
func (f *Form) GetAllDoneInfo(c *gin.Context, tx *gorm.DB, uid string) ([]Form, error) {
	var more []Form
	res := tx.WithContext(c).Table(f.TableName()).Where("uid = ?", uid).Where("status = ?", 1).Find(&more)
	err := res.Error
	if res.RowsAffected == 0 { // 不存在
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	for k, _ := range more {
		err = tx.Model(&more[k]).Association("FormInfos").Find(&more[k].FormInfos)
		if err != nil {
			return nil, err
		}
	}
	return more, nil
}

// GetAllDoing 拿到正在填写的信息
func (f *Form) GetAllDoing(c *gin.Context, tx *gorm.DB, uid string) ([]Form, error) {
	var forms []Form
	result := tx.WithContext(c).Table(f.TableName()).Select("random_id, created_at, is_delete, is_ban").
		Where("uid = ?", uid).Where("status = ?", 0).Find(&forms)
	err := result.Error
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return forms, nil
}
