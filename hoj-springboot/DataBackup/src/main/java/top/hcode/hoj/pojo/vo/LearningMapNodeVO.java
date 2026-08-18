package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;

import java.util.List;

@Data
@Accessors(chain = true)
@EqualsAndHashCode(callSuper = true)
public class LearningMapNodeVO extends LearningMapNode {
    private List<String> tagsList;
    private String remark;
    private LearningProblemVO problemInfo;
}
