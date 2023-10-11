package models

import (
	"time"
)

// Users [...]
type Users struct {
	UID      int64     `gorm:"autoIncrement:true;primaryKey;column:uid;type:bigint;not null;comment:'用户唯一id'" json:"uid"`
	Uname    string    `gorm:"unique;column:uname;type:varchar(64);not null;comment:'用户名'" json:"uname"`
	Passwd   string    `gorm:"column:passwd;type:varchar(64);not null;default:'';comment:'密码'" json:"passwd"`
	Nickname string    `gorm:"column:nickname;type:varchar(64);not null;default:'';comment:'昵称'" json:"nickname"`
	Avatar   string    `gorm:"column:avatar;type:varchar(1024);not null;default:'';comment:'头像'" json:"avatar"`
	Gender   int8      `gorm:"column:gender;type:tinyint;not null;default:0;comment:'性别'" json:"gender"`
	Phone    string    `gorm:"column:phone;type:varchar(32);not null;default:'';comment:'电话号码'" json:"phone"`
	Email    string    `gorm:"column:email;type:varchar(64);not null;default:'';comment:'电子邮箱'" json:"email"`
	Stat     int8      `gorm:"column:stat;type:tinyint;not null;default:0;comment:'状态码'" json:"stat"`
	CreateAt time.Time `gorm:"column:create_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_at"`
	UpdateAt time.Time `gorm:"column:update_at;type:datetime;default:null;comment:'修改时间'" json:"update_at"`
}

// TableName get sql table name.获取数据库表名
func (m *Users) TableName() string {
	return "users"
}

// UsersColumns get sql column name.获取数据库列名
var UsersColumns = struct {
	UID      string
	Uname    string
	Passwd   string
	Nickname string
	Avatar   string
	Gender   string
	Phone    string
	Email    string
	Stat     string
	CreateAt string
	UpdateAt string
}{
	UID:      "uid",
	Uname:    "uname",
	Passwd:   "passwd",
	Nickname: "nickname",
	Avatar:   "avatar",
	Gender:   "gender",
	Phone:    "phone",
	Email:    "email",
	Stat:     "stat",
	CreateAt: "create_at",
	UpdateAt: "update_at",
}
