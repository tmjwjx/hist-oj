package top.hcode.hoj.pojo.dto;

import top.hcode.hoj.pojo.vo.ProblemAiLanguageProgramVO;

import lombok.Data;

import java.util.List;

@Data
public class ProblemAiValidateDTO {
    private Long pid;
    private String language;
    private String standardProgram;
    private Boolean aiGenerated;
    private String algorithmSummary;
    private String validatorLanguage;
    private String validatorCode;
    private List<ProblemAiLanguageProgramVO> multiLanguagePrograms;
}
