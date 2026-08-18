package top.hcode.hoj.pojo.dto;

import lombok.Data;

@Data
public class ProblemAiValidateDTO {
    private Long pid;
    private String language;
    private String standardProgram;
    private Boolean aiGenerated;
    private String algorithmSummary;
}
