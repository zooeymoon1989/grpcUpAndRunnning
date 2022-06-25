package bootstrap

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase() *gorm.DB {
	dsn := "liwenqiang:Liwenqi@ng10@tcp(p6ilmbnht7p3q0xzku2de5js4g91fowr.mysql.qingcloud.link:3306)/cdp?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}
