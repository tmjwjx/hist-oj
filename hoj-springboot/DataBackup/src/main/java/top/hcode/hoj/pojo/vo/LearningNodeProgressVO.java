package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;
import java.util.List;

@Data
@Accessors(chain = true)
public class LearningNodeProgressVO {
    private Long nodeId;
    private String status;
    private Date completedAt;
    private Date masteredAt;
    private List<Long> missingPrerequisiteIds;
}
