package model

import "time"

// QrcodeCheckinVO 二维码签到响应VO
type QrcodeCheckinVO struct {
	QrcodeToken string     `json:"qrcodeToken"` // 二维码token
	QrcodeUrl   string     `json:"qrcodeUrl"`   // 二维码URL（用于生成二维码图片）
	ExpiresAt   time.Time  `json:"expiresAt"`   // 过期时间
	RefreshIn   int64      `json:"refreshIn"`   // 剩余秒数
}
