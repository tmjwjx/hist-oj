package top.hcode.hoj.pojo.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

/** 管理员修改比赛报名信息。用户身份由已有报名记录确定，不接受前端修改。 */
@Data
public class ContestRegistrationUpdateDTO {

    private Long id;
    private Long cid;
    private String name;

    @JsonProperty("class")
    private String clazz;

    private String college;
    private String studentId;
    private String gender;
    private String qq;
    private String phone;
}
