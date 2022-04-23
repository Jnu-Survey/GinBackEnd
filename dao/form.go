package dao

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
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

// IsExistAndJudgeStatus 检查order是否已经写入库了
func (f *Form) IsExistAndJudgeStatus(c *gin.Context, tx *gorm.DB, order, uid string, temp *Form) (*Form, error) {
	res := tx.WithContext(c).Table(f.TableName()).
		Where("random_id = ?", order).
		Where("is_delete = ?", 0). // 没有删除
		Where("is_ban = ?", 0).    // 没有禁止填写
		First(&temp)
	if errors.Is(gorm.ErrRecordNotFound, res.Error) {
		return nil, errors.New("没有找到数据")
	}
	if res.Error != nil { // 错误
		return nil, res.Error
	}
	if strconv.Itoa(temp.Uid) != uid {
		return nil, errors.New("禁止横向越权")
	}
	if temp.Status == 1 {
		return nil, errors.New("该表单已提交")
	}
	return temp, nil
}

// IsDoingAndJudge 判断是否是正在填写的表单
func (f *Form) IsDoingAndJudge(c *gin.Context, tx *gorm.DB, order, uid string, temp *Form) error {
	res := tx.WithContext(c).Table(f.TableName()).
		Where("random_id = ?", order).
		Where("is_delete = ?", 0). // 没有删除
		Where("is_ban = ?", 0).    // 没有禁止填写
		Where("status = ?", 0).    // 正在填写
		Where("uid = ?", uid).     // 是我本人的才行
		First(&temp)
	if errors.Is(gorm.ErrRecordNotFound, res.Error) {
		return errors.New("未找到表单处于正在填写状态")
	}
	if res.Error != nil { // 错误
		return errors.New("数据库错误")
	}
	return nil
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
func (f *Form) GetAllDoneInfo(c *gin.Context, tx *gorm.DB, uid string, pageIndex int) ([]Form, error) {
	var more []Form
	res := tx.Table(f.TableName()).
		Where("uid = ?", uid).
		Where("status = ?", 1).
		//Where("is_delete = ?", 0).              // 没有删除
		Limit(10).Offset(10 * (pageIndex - 1)). // 实现分页查询
		Find(&more)
	err := res.Error
	if res.RowsAffected == 0 { // 不存在
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	for k, _ := range more { // 进行联合查询
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
		Where("uid = ?", uid).
		Where("status = ?", 0).    // 正在填写
		Where("is_delete = ?", 0). // 没有删除
		Find(&forms)
	err := result.Error
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return forms, nil
}

// GetFormDetailByOrderId 通过表单ID查询订单详情
func (f *Form) GetFormDetailByOrderId(c *gin.Context, tx *gorm.DB, orderId string, manager bool) (*Form, error) {
	var form Form
	result := tx.WithContext(c).Table(f.TableName()).
		Where("random_id = ?", orderId).
		Where("is_delete = ?", 0). // 没有删除
		Where("status = ?", 1).    // 已经定型了
		Find(&form)
	if result.RowsAffected == 0 { // 没找到
		return nil, errors.New("没有该条记录")
	}
	if result.Error != nil { // 存在错误
		return nil, errors.New("数据库错误")
	}
	if !manager && form.IsBan == 1 { // 如果是管理路径那么不用关心是不是关闭了
		return nil, errors.New("该订单已关闭填写通道")
	}
	// 进行联合查询
	err := tx.Model(&form).Association("FormInfos").Find(&form.FormInfos)
	if err != nil { // 存在错误
		return nil, errors.New("数据库错误")
	}
	return &form, nil
}

// GetDetailDone 找到制作好了的
func (f *Form) GetDetailDone(c *gin.Context, tx *gorm.DB, orderId, uid string, temp *Form) (*Form, error) {
	res := tx.WithContext(c).Table(f.TableName()).
		Where("random_id = ?", orderId).
		Where("is_delete = ?", 0). // 没有删除
		Where("is_ban = ?", 0).    // 没有禁止填写
		Where("status = ?", 1).    // 正在填写
		Where("uid = ?", uid).     // 是我本人的才行
		First(&temp)
	if errors.Is(gorm.ErrRecordNotFound, res.Error) {
		return nil, errors.New("该表单不存在")
	}
	if res.Error != nil { // 错误
		return nil, errors.New("数据库错误")
	}
	err := tx.Model(&temp).Association("FormInfos").Find(&temp.FormInfos)
	if err != nil { // 存在错误
		return nil, errors.New("数据库错误")
	}
	return temp, nil
}

// GetUidFromOrder 通过订单找uid
func (f *Form) GetUidFromOrder(c *gin.Context, tx *gorm.DB, orderId string) (int, error) {
	var form Form
	result := tx.WithContext(c).Table(f.TableName()).
		Where("random_id = ?", orderId).
		Find(&form)
	if result.RowsAffected == 0 { // 没找到
		return -1, errors.New("没有该条记录")
	}
	if result.Error != nil { // 存在错误
		return -1, errors.New("数据库错误")
	}
	return form.Uid, nil
}

// ChangeBanStatus 修改状态
func (f *Form) ChangeBanStatus(c *gin.Context, tx *gorm.DB, form *Form) error {
	err := tx.WithContext(c).Table(f.TableName()).Save(&form).Error
	if err != nil {
		return err
	}
	return nil
}

// GetHowManyIDone 查询我创建了多少表单
func (f *Form) GetHowManyIDone(c *gin.Context, tx *gorm.DB, uid string) (int, error) {
	var form []Form
	result := tx.WithContext(c).Table(f.TableName()).
		Where("uid = ?", uid).
		Where("is_delete = ?", 0).
		Find(&form)
	if result.Error != nil {
		return -1, errors.New("数据库错误")
	}
	return int(result.RowsAffected), nil
}
