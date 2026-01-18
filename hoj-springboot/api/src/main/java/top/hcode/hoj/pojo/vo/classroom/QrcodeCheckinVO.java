package top.hcode.hoj.pojo.vo.classroom;

import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;

/**
 * 二维码签到响应VO
 */
@Data
@ApiModel(value = "二维码签到信息", description = "二维码签到相关信息")
public class QrcodeCheckinVO {

    @ApiModelProperty(value = "二维码token")
    private String qrcodeToken;

    @ApiModelProperty(value = "二维码URL（用于生成二维码图片）")
    private String qrcodeUrl;

    @ApiModelProperty(value = "过期时间")
    private String expiresAt;

    @ApiModelProperty(value = "剩余秒数")
    private Long refreshIn;
}
