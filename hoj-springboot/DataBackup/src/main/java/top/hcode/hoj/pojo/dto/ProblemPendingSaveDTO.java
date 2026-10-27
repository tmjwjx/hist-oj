package top.hcode.hoj.pojo.dto;

import lombok.Data;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;

@Data
public class ProblemPendingSaveDTO {
    private ProblemDTO problemDTO;
    private Long pendingId;
    private Long contestId;
    private ContestProblem contestProblem;
}
