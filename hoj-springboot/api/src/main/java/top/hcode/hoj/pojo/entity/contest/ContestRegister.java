package top.hcode.hoj.pojo.entity.contest;

import com.baomidou.mybatisplus.annotation.FieldFill;
import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.io.Serializable;
import java.util.Date;

/**
 * <p>
 * 
 * </p>
 *
 * @author Himit_ZH
 * @since 2020-10-23
 */
@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true)
@ApiModel(value="ContestRegister对象", description="")
public class ContestRegister implements Serializable {

    private static final long serialVersionUID = 1L;

    @TableId(value = "id", type = IdType.AUTO)
    private Long id;

    @ApiModelProperty(value = "比赛id")
    private Long cid;

    @ApiModelProperty(value = "用户id")
    private String uid;

    @TableField(exist = false)
    @ApiModelProperty(value = "OJ用户名")
    private String username;

    @ApiModelProperty(value = "默认为0表示正常，1为失效。")
    private Integer status;

    @ApiModelProperty(value = "姓名")
    private String name;

    @TableField("class")
    @JsonProperty("class")
    @ApiModelProperty(value = "班级")
    private String clazz;

    @ApiModelProperty(value = "学院")
    private String college;

    @ApiModelProperty(value = "学号")
    private String studentId;

    @ApiModelProperty(value = "性别")
    private String gender;

    @ApiModelProperty(value = "QQ")
    private String qq;

    @ApiModelProperty(value = "电话号码")
    private String phone;

    @TableField(fill = FieldFill.INSERT)
    private Date gmtCreate;

    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Date gmtModified;


}
