package top.hcode.hoj.pojo.vo;

import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Data;
import lombok.experimental.Accessors;
import java.util.Date;
import java.util.List;

@Data
@Accessors(chain = true)
@JsonInclude(JsonInclude.Include.NON_NULL)
public class ExternalVerificationSubmissionVO {
    private Long submitId;
    private String nickname;
    private String language;
    private Integer status;
    private Integer time;
    private Integer memory;
    private Integer score;
    private Date submitTime;
    private String statusText;
    private Integer currentTest;
    private String code;
    private String errorMessage;
    private List<CaseResult> caseResults;

    @Data
    @Accessors(chain = true)
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public static class CaseResult {
        private Integer seq;
        private Integer status;
        private Integer time;
        private Integer memory;
        private Integer score;
        private Integer groupNum;
        private String mode;
        // Populated exclusively for authorized administrators / problem authors.
        private String inputData;
        private String outputData;
        private String userOutput;
        private String stderr;
    }
}
