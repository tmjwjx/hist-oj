package top.hcode.hoj.pojo.dto;

import lombok.Data;

/** Request to recheck an existing AI validation report against the current problem. */
@Data
public class ProblemAiRecheckDTO {
    private Long recordId;
    private String requirements;
}
