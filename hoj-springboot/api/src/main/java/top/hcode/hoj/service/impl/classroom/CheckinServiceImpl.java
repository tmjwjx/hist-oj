package top.hcode.hoj.service.impl.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.mapper.classroom.CheckinMapper;
import top.hcode.hoj.mapper.classroom.CheckinRecordMapper;
import top.hcode.hoj.pojo.entity.classroom.Checkin;
import top.hcode.hoj.pojo.entity.classroom.CheckinRecord;
import top.hcode.hoj.pojo.vo.classroom.QrcodeCheckinVO;
import top.hcode.hoj.service.classroom.CheckinService;

import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.UUID;

/**
 * 签到服务实现类
 */
@Service
public class CheckinServiceImpl extends ServiceImpl<CheckinMapper, Checkin> implements CheckinService {

    @Autowired
    private CheckinMapper checkinMapper;

    @Autowired
    private CheckinRecordMapper checkinRecordMapper;

    private static final SimpleDateFormat DATE_FORMAT = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

    @Override
    @Transactional(rollbackFor = Exception.class)
    public String generateQrcodeToken(Long checkinId) {
        Checkin checkin = checkinMapper.selectById(checkinId);
        if (checkin == null) {
            throw new RuntimeException("签到不存在");
        }

        // token格式: checkinId_timestamp_random
        String timestamp = String.valueOf(System.currentTimeMillis());
        String random = UUID.randomUUID().toString().substring(0, 8);
        String token = checkinId + "_" + timestamp + "_" + random;

        // 设置过期时间
        Integer interval = checkin.getQrcodeRefreshInterval() != null
            ? checkin.getQrcodeRefreshInterval()
            : 15;

        Date expiresAt = new Date(System.currentTimeMillis() + interval * 1000);

        // 更新签到记录
        checkin.setQrcodeToken(token);
        checkin.setQrcodeExpiresAt(expiresAt);
        checkinMapper.updateById(checkin);

        return token;
    }

    @Override
    public QrcodeCheckinVO getQrcodeInfo(Long checkinId) {
        Checkin checkin = checkinMapper.selectById(checkinId);
        if (checkin == null) {
            throw new RuntimeException("签到不存在");
        }

        // 如果还没有二维码token，生成一个
        if (checkin.getQrcodeToken() == null ||
            checkin.getQrcodeExpiresAt() == null ||
            new Date().after(checkin.getQrcodeExpiresAt())) {
            generateQrcodeToken(checkinId);
            checkin = checkinMapper.selectById(checkinId);
        }

        QrcodeCheckinVO vo = new QrcodeCheckinVO();
        vo.setQrcodeToken(checkin.getQrcodeToken());

        // 生成二维码URL（前端使用这个URL生成二维码）
        // TODO: 替换为实际的域名
        String qrcodeUrl = "/api/classroom/checkin/scan?token=" + checkin.getQrcodeToken();
        vo.setQrcodeUrl(qrcodeUrl);
        vo.setExpiresAt(DATE_FORMAT.format(checkin.getQrcodeExpiresAt()));

        // 计算剩余秒数
        long remaining = (checkin.getQrcodeExpiresAt().getTime() - System.currentTimeMillis()) / 1000;
        vo.setRefreshIn(Math.max(0, remaining));

        return vo;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public QrcodeCheckinVO refreshQrcode(Long checkinId) {
        // 生成新的token
        generateQrcodeToken(checkinId);

        // 返回新的二维码信息
        return getQrcodeInfo(checkinId);
    }

    @Override
    public boolean validateQrcodeToken(String token, Long checkinId) {
        Checkin checkin = checkinMapper.selectById(checkinId);
        if (checkin == null) {
            return false;
        }

        // 检查签到类型
        if (!"qrcode".equals(checkin.getCheckinType())) {
            return false;
        }

        // 检查token是否匹配
        if (!token.equals(checkin.getQrcodeToken())) {
            return false;
        }

        // 检查是否过期
        if (checkin.getQrcodeExpiresAt() == null || new Date().after(checkin.getQrcodeExpiresAt())) {
            return false;
        }

        // 检查签到状态
        if (checkin.getStatus() != 1) { // 1=进行中
            return false;
        }

        // 检查签到时间范围
        Date now = new Date();
        if (checkin.getStartTime() != null && now.before(checkin.getStartTime())) {
            return false;
        }
        if (checkin.getEndTime() != null && now.after(checkin.getEndTime())) {
            return false;
        }

        return true;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean submitQrcodeCheckin(String token, Long checkinId, String userId) {
        // 验证token
        if (!validateQrcodeToken(token, checkinId)) {
            throw new RuntimeException("二维码无效或已过期");
        }

        // 检查是否已经签到过
        QueryWrapper<CheckinRecord> wrapper = new QueryWrapper<>();
        wrapper.eq("checkin_id", checkinId).eq("uid", userId);
        CheckinRecord existingRecord = checkinRecordMapper.selectOne(wrapper);
        if (existingRecord != null) {
            throw new RuntimeException("已经签到过了");
        }

        // 创建签到记录
        CheckinRecord record = new CheckinRecord();
        record.setCheckinId(checkinId);
        record.setUid(userId);
        record.setStatus("present"); // 签到成功
        record.setCheckinTime(new Date());

        int result = checkinRecordMapper.insert(record);

        return result > 0;
    }
}
