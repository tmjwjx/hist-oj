package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.List;

@Data
@Accessors(chain = true)
public class ContestVerificationVO {
    private Boolean ready;
    private Integer totalProblems;
    private List<ProblemVerificationVO> problems;
}
