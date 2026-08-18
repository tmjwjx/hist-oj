package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;

@Data
public class ContestQuestionCreateDTO {

    @NotNull(message = "比赛 ID 不能为空")
    private Long contestId;

    @NotBlank(message = "问题标题不能为空")
    @Size(max = 200, message = "问题标题不能超过 200 个字符")
    private String title;

    @NotBlank(message = "问题内容不能为空")
    @Size(max = 5000, message = "问题内容不能超过 5000 个字符")
    private String content;
}
