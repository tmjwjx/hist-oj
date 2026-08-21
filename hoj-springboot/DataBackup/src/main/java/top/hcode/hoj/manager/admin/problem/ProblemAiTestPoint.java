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
    /**
     * Whether the input/output fields were resolved from the concrete testcase
     * files.  A testcase is still a formal judge point when these are false;
     * this flag must never be used as a reason to downgrade an Accepted point
     * to WARN.
     */
    private Boolean inputAvailable;
    private Boolean expectedOutputAvailable;
    /** 是否对应一个有独立输入文件的正式测试点。动态/样例占位点为 false。 */
    private Boolean concrete;
    private Integer groupNum;
    private Integer score;
    private Boolean executed;
    private Integer judgeStatus;
    private String judgeStatusText;
    private Integer time;
    private Integer memory;
    private String stderr;
    private String validatorStatus;
    private String validatorStatusText;
    private String validatorStderr;
    private Long validatorTime;
    private Long validatorMemory;
}
