package top.hcode.hoj.service.classroom;

import com.baomidou.mybatisplus.extension.service.IService;
import top.hcode.hoj.pojo.entity.classroom.Checkin;
import top.hcode.hoj.pojo.vo.classroom.QrcodeCheckinVO;

/**
 * 签到服务接口
 */
public interface CheckinService extends IService<Checkin> {

    /**
     * 生成二维码token
     * @param checkinId 签到ID
     * @return 二维码token
     */
    String generateQrcodeToken(Long checkinId);

    /**
     * 获取当前二维码信息
     * @param checkinId 签到ID
     * @return 二维码信息
     */
    QrcodeCheckinVO getQrcodeInfo(Long checkinId);

    /**
     * 刷新二维码
     * @param checkinId 签到ID
     * @return 新的二维码信息
     */
    QrcodeCheckinVO refreshQrcode(Long checkinId);

    /**
     * 验证二维码token
     * @param token 二维码token
     * @param checkinId 签到ID
     * @return 是否有效
     */
    boolean validateQrcodeToken(String token, Long checkinId);

    /**
     * 学生通过二维码签到
     * @param token 二维码token
     * @param checkinId 签到ID
     * @param userId 学生ID
     * @return 签到结果
     */
    boolean submitQrcodeCheckin(String token, Long checkinId, String userId);
}
