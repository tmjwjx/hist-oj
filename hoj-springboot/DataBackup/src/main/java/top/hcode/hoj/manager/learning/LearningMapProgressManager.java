package top.hcode.hoj.manager.learning;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.LearningMapException;
import top.hcode.hoj.dao.learning.UserLearningProgressEntityService;
import top.hcode.hoj.dao.user.UserAcproblemEntityService;
import top.hcode.hoj.pojo.entity.learning.UserLearningProgress;
import top.hcode.hoj.pojo.entity.user.UserAcproblem;
import top.hcode.hoj.pojo.vo.*;

import javax.annotation.Resource;
import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Component
public class LearningMapProgressManager {

    @Resource
    private LearningMapCatalogManager catalogManager;
    @Resource
    private LearningMapPermissionManager permissionManager;
    @Resource
    private UserLearningProgressEntityService progressService;
    @Resource
    private UserAcproblemEntityService acProblemService;

    public LearningMapGraphVO graph(Long mapId, String uid) {
        permissionManager.ensureAccess(mapId, uid);
        return catalogManager.graph(mapId, true);
    }

    @Transactional(rollbackFor = Exception.class)
    public LearningMapProgressVO progress(Long mapId, String uid) {
        LearningMapGraphVO graph = graph(mapId, uid);
        Map<Long, LearningNodeProgressVO> progress = sync(mapId, uid, graph);
        return new LearningMapProgressVO().setProgress(sorted(progress.values()))
                .setSummary(LearningProgressCalculator.summary(progress.values()));
    }

    @Transactional(rollbackFor = Exception.class)
    public LearningMapFullVO full(Long mapId, String uid) {
        LearningMapGraphVO graph = graph(mapId, uid);
        Map<Long, LearningNodeProgressVO> progress = sync(mapId, uid, graph);
        return new LearningMapFullVO().setMap(graph.getMap()).setNodes(graph.getNodes()).setEdges(graph.getEdges())
                .setProgress(sorted(progress.values()))
                .setSummary(LearningProgressCalculator.summary(progress.values()))
                .setNextRecommended(next(graph.getNodes(), progress));
    }

    @Transactional(rollbackFor = Exception.class)
    public void start(Long mapId, Long nodeId, String uid) {
        LearningMapGraphVO graph = graph(mapId, uid);
        LearningMapNodeVO node = findNode(graph, nodeId);
        requireKnowledge(node);
        Map<Long, LearningNodeProgressVO> progress = sync(mapId, uid, graph);
        requireUnlocked(progress.get(nodeId));
        saveManual(mapId, nodeId, uid, LearningMapRules.IN_PROGRESS, false);
    }

    @Transactional(rollbackFor = Exception.class)
    public void complete(Long mapId, Long nodeId, String uid) {
        LearningMapGraphVO graph = graph(mapId, uid);
        LearningMapNodeVO node = findNode(graph, nodeId);
        requireKnowledge(node);
        Map<Long, LearningNodeProgressVO> progress = sync(mapId, uid, graph);
        requireUnlocked(progress.get(nodeId));
        saveManual(mapId, nodeId, uid, LearningMapRules.COMPLETED, true);
        sync(mapId, uid, graph);
    }

    @Transactional(rollbackFor = Exception.class)
    public List<LearningMapNodeVO> search(Long mapId, String uid, String keyword) {
        LearningMapGraphVO graph = graph(mapId, uid);
        Map<Long, LearningNodeProgressVO> progress = sync(mapId, uid, graph);
        String value = LearningMapRules.text(keyword).toLowerCase(Locale.ROOT);
        List<LearningMapNodeVO> result = new ArrayList<>();
        for (LearningMapNodeVO node : graph.getNodes()) {
            if (!value.isEmpty() && !matches(node, value)) {
                continue;
            }
            Map<String, Object> metadata = LearningMapRules.metadata(node.getMetadata());
            LearningNodeProgressVO item = progress.get(node.getId());
            metadata.put("status", item.getStatus());
            if (item.getMissingPrerequisiteIds() != null && !item.getMissingPrerequisiteIds().isEmpty()) {
                metadata.put("missingPrerequisiteIds", item.getMissingPrerequisiteIds());
            }
            node.setMetadata(LearningMapRules.writeJson(metadata, "{}"));
            result.add(node);
        }
        return result;
    }

    @Transactional(rollbackFor = Exception.class)
    public LearningMapRecommendVO recommend(Long mapId, String uid) {
        LearningMapGraphVO graph = graph(mapId, uid);
        Map<Long, LearningNodeProgressVO> progress = sync(mapId, uid, graph);
        return new LearningMapRecommendVO().setNext(next(graph.getNodes(), progress))
                .setSummary(LearningProgressCalculator.summary(progress.values()));
    }

    private Map<Long, LearningNodeProgressVO> sync(Long mapId, String uid, LearningMapGraphVO graph) {
        List<Long> problemIds = graph.getNodes().stream().filter(node -> LearningMapRules.PROBLEM.equals(node.getType()))
                .map(LearningMapNodeVO::getProblemId).filter(Objects::nonNull).collect(Collectors.toList());
        Set<Long> solved = problemIds.isEmpty() ? Collections.emptySet()
                : acProblemService.list(new QueryWrapper<UserAcproblem>().eq("uid", uid).in("pid", problemIds))
                .stream().map(UserAcproblem::getPid).collect(Collectors.toSet());
        List<UserLearningProgress> rows = progressService.list(new QueryWrapper<UserLearningProgress>()
                .eq("user_id", uid).eq("map_id", mapId));
        Map<Long, UserLearningProgress> existing = rows.stream()
                .collect(Collectors.toMap(UserLearningProgress::getNodeId, Function.identity()));
        Map<Long, LearningNodeProgressVO> calculated = LearningProgressCalculator.calculate(
                graph.getNodes(), graph.getEdges(), existing, solved);
        persist(mapId, uid, calculated, existing);
        return calculated;
    }

    private void persist(Long mapId, String uid, Map<Long, LearningNodeProgressVO> calculated,
                         Map<Long, UserLearningProgress> existing) {
        Date now = new Date();
        for (LearningNodeProgressVO item : calculated.values()) {
            UserLearningProgress row = existing.get(item.getNodeId());
            boolean insert = row == null;
            if (insert) {
                row = new UserLearningProgress().setUserId(uid).setMapId(mapId).setNodeId(item.getNodeId())
                        .setCreateTime(now);
            }
            row.setStatus(item.getStatus()).setUpdateTime(now);
            if (LearningMapRules.COMPLETED.equals(item.getStatus()) || LearningMapRules.MASTERED.equals(item.getStatus())) {
                row.setCompletedAt(row.getCompletedAt() == null ? now : row.getCompletedAt());
            } else {
                row.setCompletedAt(null);
            }
            if (LearningMapRules.MASTERED.equals(item.getStatus())) {
                row.setMasteredAt(row.getMasteredAt() == null ? now : row.getMasteredAt());
            } else {
                row.setMasteredAt(null);
            }
            if (insert) {
                progressService.save(row);
            } else {
                progressService.updateById(row);
            }
        }
    }

    private void saveManual(Long mapId, Long nodeId, String uid, String status, boolean completed) {
        UserLearningProgress row = progressService.getOne(new QueryWrapper<UserLearningProgress>()
                .eq("user_id", uid).eq("map_id", mapId).eq("node_id", nodeId));
        Date now = new Date();
        boolean insert = row == null;
        if (insert) {
            row = new UserLearningProgress().setUserId(uid).setMapId(mapId).setNodeId(nodeId).setCreateTime(now);
        }
        row.setStatus(status).setUpdateTime(now);
        if (completed) {
            row.setCompletedAt(now);
        }
        if (insert) {
            progressService.save(row);
        } else {
            progressService.updateById(row);
        }
    }

    private LearningMapNodeVO next(List<LearningMapNodeVO> nodes, Map<Long, LearningNodeProgressVO> progress) {
        return nodes.stream().filter(node -> {
            String status = progress.get(node.getId()).getStatus();
            return LearningMapRules.AVAILABLE.equals(status) || LearningMapRules.IN_PROGRESS.equals(status);
        }).min(Comparator.comparing((LearningMapNodeVO node) -> node.getLevel() == null ? 0 : node.getLevel())
                .thenComparing(node -> LearningMapRules.KNOWLEDGE.equals(node.getType()) ? 0 : 1)
                .thenComparing(LearningMapNodeVO::getId)).orElse(null);
    }

    private List<LearningNodeProgressVO> sorted(Collection<LearningNodeProgressVO> progress) {
        return progress.stream().sorted(Comparator.comparing(LearningNodeProgressVO::getNodeId))
                .collect(Collectors.toList());
    }

    private LearningMapNodeVO findNode(LearningMapGraphVO graph, Long nodeId) {
        return graph.getNodes().stream().filter(node -> Objects.equals(node.getId(), nodeId)).findFirst()
                .orElseThrow(() -> LearningMapException.notFound("航海点不存在"));
    }

    private void requireKnowledge(LearningMapNodeVO node) {
        if (!LearningMapRules.KNOWLEDGE.equals(node.getType())) {
            throw new LearningMapException("航海点类型不匹配");
        }
    }

    private void requireUnlocked(LearningNodeProgressVO progress) {
        if (progress == null || LearningMapRules.LOCKED.equals(progress.getStatus())) {
            throw LearningMapException.forbidden("当前航海点尚未解锁");
        }
    }

    private boolean matches(LearningMapNodeVO node, String keyword) {
        if (contains(node.getTitle(), keyword) || contains(node.getDescription(), keyword)
                || contains(node.getRemark(), keyword)) {
            return true;
        }
        return node.getProblemInfo() != null && contains(node.getProblemInfo().getProblemDisplayId(), keyword);
    }

    private boolean contains(String text, String keyword) {
        return text != null && text.toLowerCase(Locale.ROOT).contains(keyword);
    }
}
