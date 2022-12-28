package xdb

import (
	"fmt"
	"gorm.io/gorm"
	"testing"
	"time"
)

type UserFeedback struct {
	Id       int64  `gorm:"primaryKey;autoIncrement" json:"id" form:"id"`
	UserID   int64  `json:"user_id" gorm:"index"`
	ShowID   int64  `json:"show_id" gorm:"index"`
	UserName string `json:"user_name" gorm:"size:128;"`
	NickName string `json:"nick_name" gorm:"size:128;"`
	// 1兑换/置换   2售后/发货   3建议/投诉   4充值   5其他
	Type    int    `json:"type" gorm:"index"`
	Channel string `json:"channel" gorm:"size:64;"`
	Version string `json:"version"  gorm:"size:16;"`
	Phone   string `json:"phone"  gorm:"size:32;"`
	Content string `json:"content"  gorm:"size:1024;"`
	Time    string `json:"time"  gorm:"size:32;"`
}

func TestGetSliceDBAuto(t *testing.T) {
	if _, err := OpenMysqlDB(MYSQL, "root:123456@tcp(127.0.0.1:3306)/zxmh_db?charset=utf8mb4&parseTime=True&loc=Local", &gorm.Config{}, 10, 20,
		[]interface{}{}...,
	); err != nil {
		panic(err)
	}
	InitQpsDB(10, time.Second)
	go func() {
		// 按月分表
		y := time.Now().Year()
		for i := 1; i < 13; i++ {
			byMoth := fmt.Sprintf("%d-%02d", y, i)
			// UserFeedback 自动分成20表
			if sdb := GetAutoSliceDB(QqsDB(), int64(i), SP{&UserFeedback{}, PartZero, byMoth}); sdb == nil {
				panic("get auto slice db nil")
			} else {
				sdb.Debug().Save(&UserFeedback{
					UserID:   int64(i),
					ShowID:   int64(i),
					UserName: "dddd",
					NickName: "ddd",
					Type:     0,
					Channel:  "ddd",
					Version:  "ddd",
					Phone:    "dd",
					Content:  "dd",
					Time:     time.Now().Format("2006-01-02 15:04:05"),
				})
			}
		}
	}()
	for i := 0; i < 100; i++ {
		// 通过用户的ID（i），UserFeedback 自动分成20表
		if sdb := GetAutoSliceDB(QqsDB(), int64(i), SP{&UserFeedback{}, 20, EmptyVal}); sdb == nil {
			panic("get auto slice db nil")
		} else {
			sdb.Debug().Save(&UserFeedback{
				UserID:   int64(i),
				ShowID:   int64(i),
				UserName: "dddd",
				NickName: "ddd",
				Type:     0,
				Channel:  "ddd",
				Version:  "ddd",
				Phone:    "dd",
				Content:  "dd",
				Time:     time.Now().Format("2006-01-02 15:04:05"),
			})
		}
	}
}
