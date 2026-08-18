package top.hcode.hoj.pojo.dto;

import lombok.Data;

import java.util.List;
import java.util.Map;

@Data
public class LearningMapNodeDTO {
    private String type;
    private String title;
    private String description;
    private String difficulty;
    private List<String> tags;
    private Double x;
    private Double y;
    private Integer level;
    private String region;
    private Boolean published;
    private String knowledgeContent;
    private Long problemId;
    private String problemDisplayId;
    private Map<String, Object> metadata;
}
