package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.*;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

/**
 * 班级签到实体类
 */
@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true)
@TableName("classroom_checkin")
@ApiModel(value="Checkin对象", description="班级签到")
public class Checkin implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId(value = "id", type = IdType.AUTO)
    @ApiModelProperty(value = "主键ID")
    private Long id;

    @ApiModelProperty(value = "班级ID")
    private Long classroomId;

    @ApiModelProperty(value = "签到名称")
    private String checkinName;

    @ApiModelProperty(value = "签到码")
    private String checkinCode;

    // ==================== 二维码签到字段 ====================

    @ApiModelProperty(value = "签到类型: code-签到码, qrcode-二维码")
    private String checkinType;

    @ApiModelProperty(value = "二维码token")
    private String qrcodeToken;

    @ApiModelProperty(value = "二维码过期时间")
    private Date qrcodeExpiresAt;

    @ApiModelProperty(value = "二维码刷新间隔（秒）")
    private Integer qrcodeRefreshInterval;

    // ====================================================

    @ApiModelProperty(value = "签到开始时间")
    private Date startTime;

    @ApiModelProperty(value = "签到结束时间")
    private Date endTime;

    @ApiModelProperty(value = "签到状态: 0-未开始, 1-进行中, 2-已结束")
    private Integer status;

    @ApiModelProperty(value = "创建时间")
    private Date createTime;
}
