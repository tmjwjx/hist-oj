package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ProblemVerificationVO {
    private Long pid;
    private String problemId;
    private String title;
    private String caseVersion;
    private String judgeMode;
    private Integer syncStatus;
    private String syncMessage;
    private Integer verificationStatus;
    private Long submitId;
    private Integer judgeStatus;
    private String judgeStatusText;
    private Integer currentTest;
    private String judgeMessage;
    private Boolean canSubmit;
    private Boolean verified;
    private Boolean sampleVerified;
}
