package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class LearningMapRecommendVO {
    private LearningMapNodeVO next;
    private LearningMapSummaryVO summary;
}
