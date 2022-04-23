package dao

import (
	"gorm.io/gorm"
	"time"
)

type CommitInfo struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	FormJson  string    `json:"form" gorm:"column:form" description:"压缩后的JSON信息"`
	Out       int       `json:"out" gorm:"column:out" description:"外键"`
}

func (ci *CommitInfo) TableName() string {
	return "commit_json"
}

// RecordJson 把压缩的Json进行记录
func (ci *CommitInfo) RecordJson(tx *gorm.DB, newStruct *CommitInfo) (*CommitInfo, error) {
	err := tx.Table(ci.TableName()).Create(&newStruct).Error
	if err != nil {
		return nil, err
	}
	return newStruct, nil
}
