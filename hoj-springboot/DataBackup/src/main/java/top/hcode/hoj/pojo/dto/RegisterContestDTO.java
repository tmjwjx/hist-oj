package top.hcode.hoj.pojo.dto;

import lombok.Data;
import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * @Author: Himit_ZH
 * @Date: 2022/3/12 14:37
 * @Description:
 */
@Data
public class RegisterContestDTO {

    private Long cid;

    private String password;

    private String name;

    @JsonProperty("class")
    private String clazz;

    private String college;

    private String studentId;

    private String gender;

    private String qq;

    private String phone;
}
