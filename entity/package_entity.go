package entity

import "time"

type PackageEntity struct {
	ID           int64     `json:"id" xorm:"id,pk,autoincr"`
	Name         string    `json:"name" xorm:"name"`
	Description  string    `json:"description" xorm:"description"`
	Url          string    `json:"url" xorm:"url"`
	Platform     string    `json:"platform" xorm:"platform"`
	RegisterTime time.Time `json:"register_time" xorm:"register_time"`
	CreateTime   time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime   time.Time `json:"update_time" xorm:"update_time"`
}
