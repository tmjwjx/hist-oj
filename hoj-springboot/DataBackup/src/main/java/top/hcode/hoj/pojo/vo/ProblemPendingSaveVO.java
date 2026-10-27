package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;

@Data
@Accessors(chain = true)
public class ProblemPendingSaveVO {
    private Long pendingId;
    private Long pid;
    private Long targetPid;
    private Boolean requiresVerification;
    private Problem problem;
    private ProblemDTO payload;
}
