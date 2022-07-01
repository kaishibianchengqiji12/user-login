package library

import (
	"github.com/kataras/iris/v12"
	"github.com/opfsun/user-login/v1/pkg/logiclog"
	"gorm.io/gorm"
	"time"
)

type LoginRecord struct {
	ID         int       `gorm:"column:id"`
	UserId     string    `gorm:"column:user_id"`
	LoginTime  time.Time `gorm:"column:login_time"`
	UserNameCn string    `gorm:"column:user_name_cn"`
}

//记录用户登录系统次数
func GetLoginRecord(ctx iris.Context, db *gorm.DB, userId string, name string) error {
	logiclog.CtxLogger(ctx).Info("调用登录统计信息GetLoginRecord开始")
	loginRecord := LoginRecord{
		UserId:     userId,
		LoginTime:  time.Now(),
		UserNameCn: name,
	}
	err := db.Table("user_login_record").Create(&loginRecord)
	if err != nil {
		logiclog.CtxLogger(ctx).Warnf("GetLoginRecord", err)
	}
	logiclog.CtxLogger(ctx).Info("调用登录统计信息GetLoginRecord结束")
	return nil

}
