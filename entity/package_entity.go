package entity

import "time"

type PackageEntity struct {
	ID           int64     `json:"id" xorm:"id,pk,autoincr"`
	OriginId     string    `json:"origin_id" xorm:"origin_id"`
	DisplayName  string    `json:"display_name" xorm:"display_name"`
	Comment      string    `json:"comment" xorm:"comment"`
	Platform     string    `json:"platform" xorm:"platform"`
	BucketName   string    `json:"bucket_name" xorm:"bucket_name"`
	KeyName      string    `json:"key_name" xorm:"key_name"`
	DataVersion  int64     `json:"data_version" xorm:"data_version"`
	RegisterTime time.Time `json:"register_time" xorm:"register_time"`
	CreateTime   time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime   time.Time `json:"update_time" xorm:"update_time"`
}

func (e *PackageEntity) TableName() string {
	return "package"
}
