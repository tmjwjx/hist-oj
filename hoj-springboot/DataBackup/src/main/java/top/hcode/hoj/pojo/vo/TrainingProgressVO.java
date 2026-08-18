package top.hcode.hoj.pojo.vo;

import lombok.Data;

@Data
public class TrainingProgressVO {
    private Long trainingId;
    private Long solvedCount;
    private Long totalCount;
    private String status;
}
