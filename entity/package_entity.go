package entity

import "time"

type PackageEntity struct {
	ID          string    `json:"id" xorm:"id,pk,autoincr"`
	Name        string    `json:"name" xorm:"name"`
	Description string    `json:"description" xorm:"description"`
	Url         string    `json:"url" xorm:"url"`
	Platform    string    `json:"platform" xorm:"platform"`
	CreateTime  time.Time `xorm:"create_time"`
	UpdateTime  time.Time `xorm:"update_time"`
}
