package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
public class ProblemVerificationSubmitDTO {

    @NotNull(message = "题目 ID 不能为空")
    private Long pid;

    @NotBlank(message = "语言不能为空")
    private String language;

    @NotBlank(message = "标准程序不能为空")
    private String code;

    /** ai_validation 或 creator_validation，省略时按创建者验题兼容处理。 */
    private String verificationType;
}
