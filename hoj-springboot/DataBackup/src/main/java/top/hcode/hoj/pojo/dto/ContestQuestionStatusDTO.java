package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;

@Data
public class ContestQuestionStatusDTO {

    @NotBlank(message = "问题状态不能为空")
    private String status;
}
