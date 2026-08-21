package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ProblemAiGeneratedProgramVO {
    private Long recordId;
    /** 后台全流程 AI 验题记录 ID，前端可据此查询任务状态。 */
    private Long validationRecordId;
    private String language;
    private String code;
    private String algorithm;
    private String warnings;
    private String validatorLanguage;
    private String validatorCode;
    private String validatorAlgorithm;
    private String validatorWarnings;
    /** C++17 is canonical; Java and PyPy3 are checked as secondary ports. */
    private java.util.List<ProblemAiLanguageProgramVO> multiLanguagePrograms;
}
