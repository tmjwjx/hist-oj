package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Size;

@Data
public class ContestQuestionReplyDTO {

    @NotBlank(message = "回复内容不能为空")
    @Size(max = 10000, message = "回复内容不能超过 10000 个字符")
    private String content;
}
