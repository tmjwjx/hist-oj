package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.List;

@Data
@Accessors(chain = true)
public class LearningProblemVO {
    private Long id;
    private String problemDisplayId;
    private String title;
    private Integer difficulty;
    private List<String> tags;
}
