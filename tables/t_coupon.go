package tables

import "time"

type CouponType int

const (
	TableNameCoupon = "t_coupon"

	CouponType4byte CouponType = 1
	CouponType5byte CouponType = 2
)

type TableCoupon struct {
	Id         uint64     `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Code       string     `json:"code" gorm:"column:code;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT ''"`
	CouponType CouponType `json:"type" gorm:"column:type;type:tinyint NOT NULL DEFAULT '0' COMMENT '1:4 2:5'"`
	OrderId    string     `json:"order_id" gorm:"column:order_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''""`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UseAt      int64      `json:"use_at" gorm:"column:use_at;type:bigint(20) NOT NULL DEFAULT '0' COMMENT 'used time'"`
	ExpiredAt  time.Time  `json:"expired_at" gorm:"column:expired_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	StartAt    time.Time  `json:"start_at" gorm:"column:start_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`

	IsCheck uint8  `json:"is_check" gorm:"column:is_check;type:tinyint NOT NULL DEFAULT '0'"`
	Desc    string `json:"desc" gorm:"column:desc;type:text NOT NULL COMMENT 'coupon desc'"`
}

func (t *TableCoupon) TableName() string {
	return TableNameCoupon
}
