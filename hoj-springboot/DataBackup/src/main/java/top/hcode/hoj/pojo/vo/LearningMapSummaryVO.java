package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class LearningMapSummaryVO {
    private Integer total;
    private Integer locked;
    private Integer available;
    private Integer inProgress;
    private Integer completed;
    private Integer mastered;
    private Integer completionRate;
}
