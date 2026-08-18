package top.hcode.hoj.manager.learning;

import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.entity.learning.UserLearningProgress;
import top.hcode.hoj.pojo.vo.LearningMapSummaryVO;
import top.hcode.hoj.pojo.vo.LearningNodeProgressVO;

import java.util.*;

public final class LearningProgressCalculator {
    private LearningProgressCalculator() {
    }

    public static Map<Long, LearningNodeProgressVO> calculate(
            List<? extends LearningMapNode> nodes,
            List<LearningMapEdge> edges,
            Map<Long, UserLearningProgress> existing,
            Set<Long> solvedProblems) {
        Map<Long, List<Long>> incoming = incoming(edges);
        Set<Long> completed = new HashSet<>();
        Set<Long> mastered = new HashSet<>();
        Set<Long> inProgress = new HashSet<>();
        for (LearningMapNode node : nodes) {
            UserLearningProgress row = existing.get(node.getId());
            if (LearningMapRules.PROBLEM.equals(node.getType())) {
                if (node.getProblemId() != null && solvedProblems.contains(node.getProblemId())) {
                    completed.add(node.getId());
                    if (row != null && LearningMapRules.MASTERED.equals(row.getStatus())) {
                        mastered.add(node.getId());
                    }
                }
            } else if (row != null) {
                if (LearningMapRules.MASTERED.equals(row.getStatus())) {
                    completed.add(node.getId());
                    mastered.add(node.getId());
                } else if (LearningMapRules.COMPLETED.equals(row.getStatus())) {
                    completed.add(node.getId());
                } else if (LearningMapRules.IN_PROGRESS.equals(row.getStatus())) {
                    inProgress.add(node.getId());
                }
            }
        }
        Map<Long, Boolean> effective = new HashMap<>();
        for (LearningMapNode node : nodes) {
            resolve(node.getId(), incoming, completed, effective, new HashSet<>());
        }
        Map<Long, LearningNodeProgressVO> result = new LinkedHashMap<>();
        for (LearningMapNode node : nodes) {
            List<Long> missing = new ArrayList<>();
            for (Long prerequisite : incoming.getOrDefault(node.getId(), Collections.emptyList())) {
                if (!Boolean.TRUE.equals(effective.get(prerequisite))) {
                    missing.add(prerequisite);
                }
            }
            UserLearningProgress row = existing.get(node.getId());
            LearningNodeProgressVO progress = new LearningNodeProgressVO().setNodeId(node.getId());
            if (!missing.isEmpty()) {
                progress.setStatus(LearningMapRules.LOCKED).setMissingPrerequisiteIds(missing);
            } else if (mastered.contains(node.getId()) && Boolean.TRUE.equals(effective.get(node.getId()))) {
                progress.setStatus(LearningMapRules.MASTERED)
                        .setCompletedAt(row == null ? null : row.getCompletedAt())
                        .setMasteredAt(row == null ? null : row.getMasteredAt());
            } else if (Boolean.TRUE.equals(effective.get(node.getId()))) {
                progress.setStatus(LearningMapRules.COMPLETED)
                        .setCompletedAt(row == null ? null : row.getCompletedAt());
            } else if (inProgress.contains(node.getId())) {
                progress.setStatus(LearningMapRules.IN_PROGRESS);
            } else {
                progress.setStatus(LearningMapRules.AVAILABLE);
            }
            result.put(node.getId(), progress);
        }
        return result;
    }

    public static LearningMapSummaryVO summary(Collection<LearningNodeProgressVO> progress) {
        int locked = 0, available = 0, inProgress = 0, completed = 0, mastered = 0;
        for (LearningNodeProgressVO item : progress) {
            switch (item.getStatus()) {
                case LearningMapRules.LOCKED:
                    locked++;
                    break;
                case LearningMapRules.AVAILABLE:
                    available++;
                    break;
                case LearningMapRules.IN_PROGRESS:
                    inProgress++;
                    break;
                case LearningMapRules.COMPLETED:
                    completed++;
                    break;
                case LearningMapRules.MASTERED:
                    mastered++;
                    break;
                default:
                    break;
            }
        }
        int total = progress.size();
        return new LearningMapSummaryVO().setTotal(total).setLocked(locked).setAvailable(available)
                .setInProgress(inProgress).setCompleted(completed).setMastered(mastered)
                .setCompletionRate(total == 0 ? 0 : (completed + mastered) * 100 / total);
    }

    private static boolean resolve(Long nodeId, Map<Long, List<Long>> incoming, Set<Long> completed,
                                   Map<Long, Boolean> memo, Set<Long> visiting) {
        if (memo.containsKey(nodeId)) {
            return memo.get(nodeId);
        }
        if (!completed.contains(nodeId) || !visiting.add(nodeId)) {
            memo.put(nodeId, false);
            return false;
        }
        for (Long prerequisite : incoming.getOrDefault(nodeId, Collections.emptyList())) {
            if (!resolve(prerequisite, incoming, completed, memo, visiting)) {
                visiting.remove(nodeId);
                memo.put(nodeId, false);
                return false;
            }
        }
        visiting.remove(nodeId);
        memo.put(nodeId, true);
        return true;
    }

    private static Map<Long, List<Long>> incoming(List<LearningMapEdge> edges) {
        Map<Long, List<Long>> result = new HashMap<>();
        for (LearningMapEdge edge : edges) {
            if (LearningMapRules.PREREQUISITE.equals(edge.getType())) {
                result.computeIfAbsent(edge.getTargetNodeId(), key -> new ArrayList<>())
                        .add(edge.getSourceNodeId());
            }
        }
        return result;
    }
}
