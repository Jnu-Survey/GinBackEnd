package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Commit struct {
	Id           int        `json:"id" gorm:"primary_key" description:"自增主键"`
	FromUid      int        `json:"from_uid" gorm:"column:from_uid" description:"提交者的ID"`
	ToUid        int        `json:"to_uid" gorm:"column:to_uid" description:"发布者的ID"`
	OrderId      string     `json:"order_id" gorm:"column:order_id" description:"表单号"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	IsDelete     int        `json:"is_delete" gorm:"column:is_delete" description:"0:正常 1:无效(双方都可修改)"`
	Status       int        `json:"status" gorm:"column:status" description:"0: 我要填 1:我填好了"`
	HexId        string     `json:"hex_id" gorm:"column:hex_id" description:"mongoId"`
	FromUserInfo Login      `gorm:"foreignkey:FromUid"` // 填写者外键
	ToUserInfo   Login      `gorm:"foreignkey:ToUid"`   // 发布者外键
	CommitInfos  CommitInfo `gorm:"foreignkey:Id"`      // 详细信息外键
	OrderInfos   Form       `gorm:"foreignkey:OrderId"` // 订单信息外键
}

func (co *Commit) TableName() string {
	return "commit_info"
}

// IsExistDoingInfo 检查填写信息
func (co *Commit) IsExistDoingInfo(c *gin.Context, tx *gorm.DB, order, uid string) (int, string, error) {
	var temp Commit
	res := tx.WithContext(c).Table(co.TableName()).
		Where("order_id = ?", order).
		Where("from_uid = ?", uid).
		First(&temp)
	if errors.Is(gorm.ErrRecordNotFound, res.Error) {
		return 0, "", nil
	}
	if res.Error != nil { // 错误
		return -1, "", res.Error
	}
	if temp.Status == 1 { // 已经搞定了
		err := tx.Model(&temp).Association("CommitInfos").Find(&temp.CommitInfos)
		if err != nil {
			return -1, "", err
		}
		return 2, temp.CommitInfos.FormJson, nil
	} else { // 处于正在填写的状态
		return 1, "", nil
	}
}

// AddFrom2To 创建双方关系
func (co *Commit) AddFrom2To(tx *gorm.DB, newStruct *Commit) (*Commit, error) {
	err := tx.Table(co.TableName()).Create(&newStruct).Error
	if err != nil {
		return nil, err
	}
	return newStruct, nil
}

// JudgeStatusIsBeWrittenAndCreated 判断提交记录是不是已经被写了/是否存在
func (co *Commit) JudgeStatusIsBeWrittenAndCreated(tx *gorm.DB, order, uid string) error {
	var temp Commit
	res := tx.Table(co.TableName()).
		Where("from_uid = ?", uid).
		Where("order_id = ?", order).
		First(&temp)
	if errors.Is(gorm.ErrRecordNotFound, res.Error) {
		return errors.New("请勿直接提交")
	}
	if res.Error != nil { // 错误
		return errors.New("数据库错误")
	}
	if temp.Status == 1 {
		return errors.New("已提交过该表单")
	} else {
		return nil
	}
}

// RewriteCommit 重新修改关系
func (co *Commit) RewriteCommit(tx *gorm.DB, order, uid, hexId string) (*Commit, error) {
	var temp Commit
	res := tx.Table(co.TableName()).
		Where("from_uid = ?", uid).
		Where("order_id = ?", order).
		First(&temp)
	if res.Error != nil { // 错误
		return nil, res.Error
	}
	temp.Status = 1
	temp.HexId = hexId
	err := tx.Table(co.TableName()).Save(&temp).Error
	if err != nil {
		return nil, err
	}
	return &temp, nil
}

// GetAllDoingOrDone 获取我已经填写好了的/正在制作的
func (co *Commit) GetAllDoingOrDone(c *gin.Context, tx *gorm.DB, fromUid string, pageIndex int, status int) ([]Commit, error) {
	var commitInfos []Commit
	result := tx.WithContext(c).Table(co.TableName()).
		Where("from_uid = ?", fromUid).
		Where("status = ?", status).            // 已经填写好了的
		Limit(10).Offset(10 * (pageIndex - 1)). // 实现分页查询
		Find(&commitInfos)
	err := result.Error
	if result.RowsAffected == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for k, _ := range commitInfos {
		err := tx.Model(&commitInfos[k]).Association("ToUserInfo").Find(&commitInfos[k].ToUserInfo) // 联合查询对方的信息
		if err != nil {
			return nil, err
		}
	}
	for k, _ := range commitInfos {
		var tempOrderInfo *Form
		tempOrderInfo, err = tempOrderInfo.GetFormDetailByOrderId(c, tx, commitInfos[k].OrderId, true) // 不关心关不关闭
		if err != nil {
			return nil, err
		}
		commitInfos[k].OrderInfos = *tempOrderInfo
	}
	return commitInfos, nil
}

// GetAllFillFormForMe 获取别人已经对我填写好了的
func (co *Commit) GetAllFillFormForMe(c *gin.Context, tx *gorm.DB, toUid string, pageIndex int) ([]Commit, error) {
	var commitInfos []Commit
	result := tx.WithContext(c).Table(co.TableName()).
		Where("to_uid = ?", toUid).
		Where("status = ?", 1).                 // 已经填写好了的
		Limit(10).Offset(10 * (pageIndex - 1)). // 实现分页查询
		Find(&commitInfos)
	err := result.Error
	if result.RowsAffected == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for k, _ := range commitInfos {
		err := tx.Model(&commitInfos[k]).Association("FromUserInfo").Find(&commitInfos[k].FromUserInfo) // 联合查询对方的信息
		if err != nil {
			return nil, err
		}
	}
	for k, _ := range commitInfos {
		var tempOrderInfo *Form
		tempOrderInfo, err = tempOrderInfo.GetFormDetailByOrderId(c, tx, commitInfos[k].OrderId, true) // 不关心关不关闭
		if err != nil {
			return nil, err
		}
		commitInfos[k].OrderInfos = *tempOrderInfo
	}
	return commitInfos, nil
}

// GetInfoToMeByOrder 获取对我填写某一个表单的信息
func (co *Commit) GetInfoToMeByOrder(c *gin.Context, tx *gorm.DB, order, Uid string) ([]Commit, error) {
	var commitInfos []Commit
	result := tx.WithContext(c).Table(co.TableName()).
		Where("to_uid = ?", Uid).
		Where("order_id = ?", order).
		Where("status = ?", 1). // 已经填写好了的
		Find(&commitInfos)
	err := result.Error
	if err != nil {
		return nil, err
	}
	for k, _ := range commitInfos {
		err := tx.Model(&commitInfos[k]).Association("FromUserInfo").Find(&commitInfos[k].FromUserInfo) // 联合查询对方的信息
		if err != nil {
			return nil, err
		}
	}
	return commitInfos, nil
}

// HandleValid 处理设置填写记录为无效
func (co *Commit) HandleValid(tx *gorm.DB, fromUid, toUid, order string) error {
	var commitInfo Commit
	result := tx.Table(co.TableName()).
		Where("from_uid = ?", fromUid).
		Where("to_uid = ?", toUid).
		Where("order_id = ?", order).
		Where("status = ?", 1).
		Find(&commitInfo)
	err := result.Error
	if err != nil {
		return errors.New("数据库错误")
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到有效数据")
	}
	if commitInfo.IsDelete != 0 {
		return errors.New("已设置为无效")
	}
	commitInfo.IsDelete = 1
	err = tx.Table(co.TableName()).Save(&commitInfo).Error
	if err != nil {
		return errors.New("数据库错误")
	}
	return nil
}

func (co *Commit) GetValidFromUId(tx *gorm.DB, toUid, order string) (map[string]bool, error) {
	var commitInfos []Commit
	result := tx.Table(co.TableName()).
		Where("to_uid = ?", toUid).
		Where("order_id = ?", order).
		Where("status = ?", 1).
		Where("is_delete = ?", 1).
		Find(&commitInfos)
	err := result.Error
	var res map[string]bool
	if err != nil {
		return res, errors.New("数据库错误")
	}
	for _, v := range commitInfos {
		res[strconv.Itoa(v.FromUid)] = true
	}
	return res, nil
}

// GetHowManyIGet 获取我拿到了多少表单数据
func (co *Commit) GetHowManyIGet(tx *gorm.DB, toUid string) (int, error) {
	var commitInfos []Commit
	result := tx.Table(co.TableName()).
		Where("to_uid = ?", toUid).
		Where("status = ?", 1).
		Where("is_delete = ?", 0).
		Find(&commitInfos)
	if result.Error != nil {
		return -1, errors.New("数据库错误")
	}
	return int(result.RowsAffected), nil
}
