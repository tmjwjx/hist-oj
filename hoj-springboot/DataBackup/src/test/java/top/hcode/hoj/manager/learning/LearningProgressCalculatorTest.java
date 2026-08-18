package top.hcode.hoj.manager.learning;

import org.junit.Test;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.entity.learning.UserLearningProgress;

import java.util.*;

import static org.junit.Assert.assertEquals;

public class LearningProgressCalculatorTest {
    @Test
    public void prerequisiteUnlocksOnlyAfterCompletion() {
        LearningMapNode first = new LearningMapNode().setId(1L).setType(LearningMapRules.KNOWLEDGE);
        LearningMapNode second = new LearningMapNode().setId(2L).setType(LearningMapRules.PROBLEM).setProblemId(100L);
        LearningMapEdge edge = new LearningMapEdge().setSourceNodeId(1L).setTargetNodeId(2L)
                .setType(LearningMapRules.PREREQUISITE);
        Map<Long, UserLearningProgress> rows = new HashMap<>();
        rows.put(1L, new UserLearningProgress().setNodeId(1L).setStatus(LearningMapRules.COMPLETED));

        Map<Long, top.hcode.hoj.pojo.vo.LearningNodeProgressVO> result =
                LearningProgressCalculator.calculate(Arrays.asList(first, second),
                        Collections.singletonList(edge), rows, Collections.singleton(100L));

        assertEquals(LearningMapRules.COMPLETED, result.get(1L).getStatus());
        assertEquals(LearningMapRules.COMPLETED, result.get(2L).getStatus());
    }

    @Test
    public void unsolvedProblemDoesNotUnlockNextNode() {
        LearningMapNode problem = new LearningMapNode().setId(1L).setType(LearningMapRules.PROBLEM).setProblemId(100L);
        LearningMapNode next = new LearningMapNode().setId(2L).setType(LearningMapRules.KNOWLEDGE);
        LearningMapEdge edge = new LearningMapEdge().setSourceNodeId(1L).setTargetNodeId(2L)
                .setType(LearningMapRules.PREREQUISITE);

        Map<Long, top.hcode.hoj.pojo.vo.LearningNodeProgressVO> result =
                LearningProgressCalculator.calculate(Arrays.asList(problem, next),
                        Collections.singletonList(edge), Collections.emptyMap(), Collections.emptySet());

        assertEquals(LearningMapRules.AVAILABLE, result.get(1L).getStatus());
        assertEquals(LearningMapRules.LOCKED, result.get(2L).getStatus());
    }
}
