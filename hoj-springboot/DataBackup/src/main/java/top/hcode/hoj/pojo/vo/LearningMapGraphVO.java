package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.entity.learning.LearningMap;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;

import java.util.List;

@Data
@Accessors(chain = true)
public class LearningMapGraphVO {
    private LearningMap map;
    private List<LearningMapNodeVO> nodes;
    private List<LearningMapEdge> edges;
}
