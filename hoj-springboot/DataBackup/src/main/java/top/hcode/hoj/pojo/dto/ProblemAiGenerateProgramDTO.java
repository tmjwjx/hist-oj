package top.hcode.hoj.pojo.dto;

import lombok.Data;

@Data
public class ProblemAiGenerateProgramDTO {
    private Long pid;
    private String language;
    private java.util.List<String> modelIds;
}
