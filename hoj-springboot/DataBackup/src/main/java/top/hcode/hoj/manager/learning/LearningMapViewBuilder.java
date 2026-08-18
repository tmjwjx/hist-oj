package top.hcode.hoj.manager.learning;

import cn.hutool.core.bean.BeanUtil;
import org.springframework.stereotype.Component;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.vo.LearningMapNodeVO;
import top.hcode.hoj.pojo.vo.LearningProblemVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Component
public class LearningMapViewBuilder {

    @Resource(name = "learningMapProblemManager")
    private LearningMapProblemManager problemManager;

    public List<LearningMapNodeVO> build(List<LearningMapNode> nodes) {
        List<Long> ids = nodes.stream().filter(this::isProblem).map(LearningMapNode::getProblemId)
                .filter(Objects::nonNull).collect(Collectors.toList());
        List<String> displays = nodes.stream().filter(this::isProblem)
                .map(LearningMapNode::getProblemDisplayId).collect(Collectors.toList());
        Map<Long, LearningProblemVO> problems = problemManager.getByNodes(ids, displays);
        Map<String, LearningProblemVO> byDisplay = problems.values().stream()
                .collect(Collectors.toMap(LearningProblemVO::getProblemDisplayId, Function.identity()));
        List<LearningMapNodeVO> result = new ArrayList<>();
        for (LearningMapNode node : nodes) {
            LearningMapNodeVO view = BeanUtil.copyProperties(node, LearningMapNodeVO.class);
            view.setTagsList(LearningMapRules.tags(node.getTags()));
            if (isProblem(node)) {
                LearningProblemVO problem = node.getProblemId() == null
                        ? byDisplay.get(node.getProblemDisplayId()) : problems.get(node.getProblemId());
                view.setProblemInfo(problem).setRemark(LearningMapRules.remark(node.getMetadata()));
                if (problem != null && view.getProblemId() == null) {
                    view.setProblemId(problem.getId()).setProblemDisplayId(problem.getProblemDisplayId());
                }
            }
            result.add(view);
        }
        return result;
    }

    private boolean isProblem(LearningMapNode node) {
        return LearningMapRules.PROBLEM.equals(node.getType());
    }
}
