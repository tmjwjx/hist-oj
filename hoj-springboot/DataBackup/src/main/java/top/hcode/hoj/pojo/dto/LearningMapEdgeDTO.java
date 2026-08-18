package top.hcode.hoj.pojo.dto;

import lombok.Data;

@Data
public class LearningMapEdgeDTO {
    private Long sourceNodeId;
    private Long targetNodeId;
    private String type;
}
