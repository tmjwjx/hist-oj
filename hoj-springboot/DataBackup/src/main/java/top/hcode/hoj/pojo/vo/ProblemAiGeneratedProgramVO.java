package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ProblemAiGeneratedProgramVO {
    private Long recordId;
    private String language;
    private String code;
    private String algorithm;
    private String warnings;
}
