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
 * 签到记录实体类
 */
@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true)
@TableName("classroom_checkin_record")
@ApiModel(value="CheckinRecord对象", description="签到记录")
public class CheckinRecord implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId(value = "id", type = IdType.AUTO)
    @ApiModelProperty(value = "主键ID")
    private Long id;

    @ApiModelProperty(value = "签到ID")
    private Long checkinId;

    @ApiModelProperty(value = "学生用户ID")
    private String uid;

    @ApiModelProperty(value = "签到状态: present-签到, absent-缺勤, sick_leave-病假, personal_leave-事假")
    private String status;

    @ApiModelProperty(value = "签到时间")
    private Date checkinTime;

    @ApiModelProperty(value = "备注")
    private String remark;

    @ApiModelProperty(value = "创建时间")
    private Date createTime;

    @ApiModelProperty(value = "更新时间")
    private Date updateTime;
}
