package service

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// CheckinService 签到服务
type CheckinService struct {
	db *gorm.DB
}

// NewCheckinService 创建签到服务
func NewCheckinService() *CheckinService {
	return &CheckinService{
		db: client.GetDB(),
	}
}

// GenerateQrcodeToken 生成二维码token
// token格式: checkinId_timestamp_random
func (s *CheckinService) GenerateQrcodeToken(checkinID uint64, refreshInterval int) (string, error) {
	logger := utils.GetLogger()

	db := client.GetDB()

	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到失败", zap.Error(err), zap.Uint64("checkin_id", checkinID))
		return "", fmt.Errorf("签到不存在")
	}

	// 生成token
	timestamp := time.Now().UnixNano()
	rand.Seed(timestamp)
	random := fmt.Sprintf("%08x", rand.Uint32())
	token := fmt.Sprintf("%d_%d_%s", checkinID, timestamp/1e6, random)

	// 设置过期时间
	interval := refreshInterval
	if interval == 0 {
		interval = 15 // 默认15秒
	}

	expiresAt := time.Now().Add(time.Duration(interval) * time.Second)

	// 更新签到记录
	checkin.QrcodeToken = token
	checkin.QrcodeExpiresAt = &expiresAt

	if err := db.Save(&checkin).Error; err != nil {
		logger.Error("更新二维码token失败", zap.Error(err))
		return "", fmt.Errorf("更新二维码token失败")
	}

	return token, nil
}

// GetQrcodeInfo 获取当前二维码信息
func (s *CheckinService) GetQrcodeInfo(checkinID uint64) (*model.QrcodeCheckinVO, error) {
	logger := utils.GetLogger()

	db := client.GetDB()

	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到失败", zap.Error(err), zap.Uint64("checkin_id", checkinID))
		return nil, fmt.Errorf("签到不存在")
	}

	// 如果还没有二维码token，生成一个
	if checkin.QrcodeToken == "" || checkin.QrcodeExpiresAt == nil || time.Now().After(*checkin.QrcodeExpiresAt) {
		_, err := s.GenerateQrcodeToken(checkinID, checkin.QrcodeRefreshInterval)
		if err != nil {
			return nil, err
		}
		// 重新查询
		db.Where("id = ?", checkinID).First(&checkin)
	}

	// 生成二维码URL
	qrcodeURL := fmt.Sprintf("/api/classroom/checkin/%d/qrcode/scan?token=%s", checkinID, checkin.QrcodeToken)

	// 计算剩余秒数
	var refreshIn int64
	if checkin.QrcodeExpiresAt != nil {
		refreshIn = int64(checkin.QrcodeExpiresAt.Sub(time.Now()).Seconds())
		if refreshIn < 0 {
			refreshIn = 0
		}
	}

	vo := &model.QrcodeCheckinVO{
		QrcodeToken: checkin.QrcodeToken,
		QrcodeUrl:   qrcodeURL,
		ExpiresAt:   *checkin.QrcodeExpiresAt,
		RefreshIn:   refreshIn,
	}

	return vo, nil
}

// RefreshQrcode 刷新二维码
func (s *CheckinService) RefreshQrcode(checkinID uint64) (*model.QrcodeCheckinVO, error) {
	logger := utils.GetLogger()

	db := client.GetDB()

	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到失败", zap.Error(err), zap.Uint64("checkin_id", checkinID))
		return nil, fmt.Errorf("签到不存在")
	}

	// 生成新的token
	token, err := s.GenerateQrcodeToken(checkinID, checkin.QrcodeRefreshInterval)
	if err != nil {
		return nil, err
	}

	// 重新查询获取更新后的数据
	db.Where("id = ?", checkinID).First(&checkin)

	// 生成二维码URL
	qrcodeURL := fmt.Sprintf("/api/classroom/checkin/%d/qrcode/scan?token=%s", checkinID, token)

	// 计算剩余秒数
	refreshIn := int64(checkin.QrcodeRefreshInterval)

	vo := &model.QrcodeCheckinVO{
		QrcodeToken: token,
		QrcodeUrl:   qrcodeURL,
		ExpiresAt:   *checkin.QrcodeExpiresAt,
		RefreshIn:   refreshIn,
	}

	return vo, nil
}

// ValidateQrcodeToken 验证二维码token
func (s *CheckinService) ValidateQrcodeToken(token string, checkinID uint64) (bool, error) {
	db := client.GetDB()

	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		return false, fmt.Errorf("签到不存在")
	}

	// 检查签到类型
	if checkin.CheckinType != "qrcode" {
		return false, nil
	}

	// 检查token是否匹配
	if checkin.QrcodeToken != token {
		return false, nil
	}

	// 检查是否过期
	if checkin.QrcodeExpiresAt == nil || time.Now().After(*checkin.QrcodeExpiresAt) {
		return false, nil
	}

	// 检查签到状态
	if checkin.Status != 1 { // 1=进行中
		return false, nil
	}

	// 检查签到时间范围
	now := time.Now()
	if !checkin.StartTime.IsZero() && now.Before(checkin.StartTime) {
		return false, nil
	}
	if checkin.EndTime != nil && now.After(*checkin.EndTime) {
		return false, nil
	}

	return true, nil
}

// SubmitQrcodeCheckin 学生通过二维码签到
func (s *CheckinService) SubmitQrcodeCheckin(token string, checkinID uint64, userID string) error {
	logger := utils.GetLogger()

	db := client.GetDB()

	// 验证token
	valid, err := s.ValidateQrcodeToken(token, checkinID)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("二维码无效或已过期")
	}

	// 检查是否已经签到过
	var count int64
	db.Model(&model.ClassroomCheckinRecord{}).
		Where("checkin_id = ? AND uid = ?", checkinID, userID).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("已经签到过了")
	}

	// 创建签到记录
	now := time.Now()
	record := &model.ClassroomCheckinRecord{
		CheckinID:  checkinID,
		UID:        userID,
		Status:     "present", // 签到成功
		CheckinTime: &now,
	}

	if err := db.Create(record).Error; err != nil {
		logger.Error("创建签到记录失败", zap.Error(err))
		return fmt.Errorf("签到失败")
	}

	logger.Info("二维码签到成功", zap.Uint64("checkin_id", checkinID), zap.String("uid", userID))

	return nil
}
