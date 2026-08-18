package top.hcode.hoj.manager.admin.problem;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ProblemAiTestPoint {
    private Integer index;
    private Long caseId;
    private String input;
    private String expectedOutput;
    private Integer groupNum;
    private Integer score;
    private Boolean executed;
    private Integer judgeStatus;
    private String judgeStatusText;
    private Integer time;
    private Integer memory;
    private String stderr;
}
