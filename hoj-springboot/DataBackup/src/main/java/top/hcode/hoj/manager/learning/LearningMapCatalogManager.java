package top.hcode.hoj.manager.learning;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.LearningMapException;
import top.hcode.hoj.dao.learning.LearningMapEdgeEntityService;
import top.hcode.hoj.dao.learning.LearningMapEntityService;
import top.hcode.hoj.dao.learning.LearningMapNodeEntityService;
import top.hcode.hoj.dao.learning.LearningMapPermissionEntityService;
import top.hcode.hoj.dao.learning.UserLearningProgressEntityService;
import top.hcode.hoj.pojo.dto.LearningMapDTO;
import top.hcode.hoj.pojo.dto.LearningMapEdgeDTO;
import top.hcode.hoj.pojo.dto.LearningMapNodeDTO;
import top.hcode.hoj.pojo.entity.learning.LearningMap;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;
import top.hcode.hoj.pojo.vo.LearningMapGraphVO;
import top.hcode.hoj.pojo.vo.LearningMapNodeVO;
import top.hcode.hoj.pojo.vo.LearningMapValidationErrorVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class LearningMapCatalogManager {

    @Resource
    private LearningMapEntityService mapService;
    @Resource
    private LearningMapNodeEntityService nodeService;
    @Resource
    private LearningMapEdgeEntityService edgeService;
    @Resource
    private UserLearningProgressEntityService progressService;
    @Resource
    private LearningMapPermissionEntityService permissionService;
    @Resource
    private LearningMapViewBuilder viewBuilder;
    @Resource(name = "learningMapProblemManager")
    private LearningMapProblemManager problemManager;

    public List<LearningMap> list(boolean published) {
        QueryWrapper<LearningMap> query = new QueryWrapper<>();
        if (published) {
            query.eq("status", LearningMapRules.PUBLISHED);
        }
        return mapService.list(query.orderByDesc("id"));
    }

    public LearningMap get(Long mapId) {
        LearningMap map = mapService.getById(mapId);
        if (map == null) {
            throw LearningMapException.notFound("航海图不存在");
        }
        return map;
    }

    public LearningMapGraphVO graph(Long mapId, boolean onlyPublished) {
        LearningMap map = get(mapId);
        if (onlyPublished && !LearningMapRules.PUBLISHED.equals(map.getStatus())) {
            throw new LearningMapException("航海图尚未发布", top.hcode.hoj.common.result.ResultStatus.FORBIDDEN);
        }
        List<LearningMapNode> nodes = nodeService.list(new QueryWrapper<LearningMapNode>()
                .eq("map_id", mapId).orderByAsc("id"));
        if (onlyPublished) {
            nodes = nodes.stream().filter(node -> Boolean.TRUE.equals(node.getPublished())).collect(Collectors.toList());
        }
        List<LearningMapEdge> edges = edgeService.list(new QueryWrapper<LearningMapEdge>()
                .eq("map_id", mapId).orderByAsc("id"));
        Set<Long> nodeIds = nodes.stream().map(LearningMapNode::getId).collect(Collectors.toSet());
        if (onlyPublished) {
            edges = edges.stream().filter(edge -> nodeIds.contains(edge.getSourceNodeId())
                    && nodeIds.contains(edge.getTargetNodeId())).collect(Collectors.toList());
        }
        return new LearningMapGraphVO().setMap(map)
                .setNodes(viewBuilder.build(nodes)).setEdges(edges);
    }

    public LearningMap create(LearningMapDTO dto) {
        LearningMap map = new LearningMap()
                .setTitle(LearningMapRules.text(dto.getTitle()))
                .setDescription(LearningMapRules.text(dto.getDescription()))
                .setStatus(LearningMapRules.status(dto.getStatus()))
                .setAccessMode(LearningMapRules.accessMode(dto.getAccessMode()))
                .setCreateTime(new Date()).setUpdateTime(new Date());
        requireTitle(map.getTitle());
        mapService.save(map);
        return map;
    }

    public LearningMap update(Long mapId, LearningMapDTO dto) {
        LearningMap map = get(mapId);
        String title = LearningMapRules.text(dto.getTitle());
        requireTitle(title);
        map.setTitle(title).setDescription(LearningMapRules.text(dto.getDescription()))
                .setStatus(LearningMapRules.status(dto.getStatus()))
                .setAccessMode(LearningMapRules.text(dto.getAccessMode()).isEmpty()
                        ? LearningMapRules.accessMode(map.getAccessMode()) : LearningMapRules.accessMode(dto.getAccessMode()))
                .setUpdateTime(new Date());
        mapService.updateById(map);
        return map;
    }

    @Transactional(rollbackFor = Exception.class)
    public void delete(Long mapId) {
        get(mapId);
        progressService.remove(new QueryWrapper<top.hcode.hoj.pojo.entity.learning.UserLearningProgress>().eq("map_id", mapId));
        permissionService.remove(new QueryWrapper<top.hcode.hoj.pojo.entity.learning.LearningMapPermission>().eq("map_id", mapId));
        edgeService.remove(new QueryWrapper<LearningMapEdge>().eq("map_id", mapId));
        nodeService.remove(new QueryWrapper<LearningMapNode>().eq("map_id", mapId));
        mapService.removeById(mapId);
    }

    public LearningMapNode createNode(Long mapId, LearningMapNodeDTO dto) {
        get(mapId);
        LearningMapNode node = buildNode(mapId, dto);
        nodeService.save(node);
        return node;
    }

    public LearningMapNode updateNode(Long mapId, Long nodeId, LearningMapNodeDTO dto) {
        get(mapId);
        LearningMapNode node = nodeService.getOne(new QueryWrapper<LearningMapNode>()
                .eq("map_id", mapId).eq("id", nodeId));
        if (node == null) {
            throw LearningMapException.notFound("航海点不存在");
        }
        LearningMapNode updated = buildNode(mapId, dto).setId(nodeId)
                .setCreateTime(node.getCreateTime()).setUpdateTime(new Date());
        nodeService.updateById(updated);
        return nodeService.getById(nodeId);
    }

    @Transactional(rollbackFor = Exception.class)
    public void deleteNode(Long mapId, Long nodeId) {
        get(mapId);
        LearningMapNode node = nodeService.getOne(new QueryWrapper<LearningMapNode>()
                .eq("map_id", mapId).eq("id", nodeId));
        if (node == null) {
            throw LearningMapException.notFound("航海点不存在");
        }
        edgeService.remove(new QueryWrapper<LearningMapEdge>().eq("map_id", mapId)
                .and(wrapper -> wrapper.eq("source_node_id", nodeId).or().eq("target_node_id", nodeId)));
        progressService.remove(new QueryWrapper<top.hcode.hoj.pojo.entity.learning.UserLearningProgress>()
                .eq("map_id", mapId).eq("node_id", nodeId));
        nodeService.removeById(nodeId);
    }

    public LearningMapEdge createEdge(Long mapId, LearningMapEdgeDTO dto) {
        get(mapId);
        LearningMapEdge edge = buildEdge(mapId, dto);
        validateNodes(mapId, edge);
        validateCycle(mapId, edge, null);
        edgeService.save(edge);
        return edge;
    }

    public LearningMapEdge updateEdge(Long mapId, Long edgeId, LearningMapEdgeDTO dto) {
        get(mapId);
        LearningMapEdge old = edgeService.getOne(new QueryWrapper<LearningMapEdge>()
                .eq("map_id", mapId).eq("id", edgeId));
        if (old == null) {
            throw LearningMapException.notFound("连线不存在");
        }
        LearningMapEdge edge = buildEdge(mapId, dto).setId(edgeId)
                .setCreateTime(old.getCreateTime()).setUpdateTime(new Date());
        validateNodes(mapId, edge);
        validateCycle(mapId, edge, edgeId);
        edgeService.updateById(edge);
        return edgeService.getById(edgeId);
    }

    public void deleteEdge(Long mapId, Long edgeId) {
        get(mapId);
        if (!edgeService.remove(new QueryWrapper<LearningMapEdge>().eq("map_id", mapId).eq("id", edgeId))) {
            throw LearningMapException.notFound("连线不存在");
        }
    }

    public List<LearningMapValidationErrorVO> validate(Long mapId) {
        get(mapId);
        List<LearningMapNode> nodes = nodeService.list(new QueryWrapper<LearningMapNode>().eq("map_id", mapId));
        List<LearningMapEdge> edges = edgeService.list(new QueryWrapper<LearningMapEdge>().eq("map_id", mapId));
        List<LearningMapValidationErrorVO> errors = new ArrayList<>();
        if (nodes.isEmpty()) {
            errors.add(new LearningMapValidationErrorVO("EMPTY_MAP", "至少需要一个航海点"));
            return errors;
        }
        Set<Long> ids = nodes.stream().map(LearningMapNode::getId).collect(Collectors.toSet());
        for (LearningMapNode node : nodes) {
            String type = LearningMapRules.nodeType(node.getType());
            if (type == null) {
                errors.add(new LearningMapValidationErrorVO("INVALID_NODE_TYPE", "节点 " + node.getId() + " 类型无效"));
                continue;
            }
            if (LearningMapRules.text(node.getTitle()).isEmpty()) {
                errors.add(new LearningMapValidationErrorVO("EMPTY_TITLE", "节点 " + node.getId() + " 标题不能为空"));
            }
            if (LearningMapRules.KNOWLEDGE.equals(type) && LearningMapRules.text(node.getKnowledgeContent()).isEmpty()) {
                errors.add(new LearningMapValidationErrorVO("EMPTY_KNOWLEDGE_CONTENT", "知识点节点 " + node.getId() + " 需要填写学习内容"));
            }
            if (LearningMapRules.PROBLEM.equals(type)) {
                try {
                    if (node.getProblemId() == null && LearningMapRules.text(node.getProblemDisplayId()).isEmpty()) {
                        throw new LearningMapException("未绑定主 OJ 题目");
                    }
                    if (node.getProblemId() != null) {
                        problemManager.get(String.valueOf(node.getProblemId()));
                    } else {
                        problemManager.get(node.getProblemDisplayId());
                    }
                } catch (LearningMapException ex) {
                    errors.add(new LearningMapValidationErrorVO("PROBLEM_NOT_FOUND", "题目节点 " + node.getId() + " 绑定题目无效: " + ex.getMessage()));
                }
            }
        }
        for (LearningMapEdge edge : edges) {
            if (!ids.contains(edge.getSourceNodeId()) || !ids.contains(edge.getTargetNodeId())) {
                errors.add(new LearningMapValidationErrorVO("INVALID_EDGE", "连线 " + edge.getId() + " 的节点不存在"));
            }
            if (LearningMapRules.edgeType(edge.getType()) == null) {
                errors.add(new LearningMapValidationErrorVO("INVALID_EDGE_TYPE", "连线 " + edge.getId() + " 的类型无效"));
            }
            if (LearningMapRules.PREREQUISITE.equals(edge.getType())
                    && Objects.equals(edge.getSourceNodeId(), edge.getTargetNodeId())) {
                errors.add(new LearningMapValidationErrorVO("INVALID_EDGE", "前置连线 " + edge.getId() + " 不能自环"));
            }
        }
        if (errors.isEmpty() && hasCycle(nodes, edges)) {
            errors.add(new LearningMapValidationErrorVO("CYCLE_DETECTED", "检测到前置依赖环，请修正"));
        }
        return errors;
    }

    public void publish(Long mapId) {
        List<LearningMapValidationErrorVO> errors = validate(mapId);
        if (!errors.isEmpty()) {
            throw new LearningMapException("发布校验失败：" + errors.stream().map(LearningMapValidationErrorVO::getMessage)
                    .collect(Collectors.joining("; ")));
        }
        LearningMap map = get(mapId).setStatus(LearningMapRules.PUBLISHED).setUpdateTime(new Date());
        mapService.updateById(map);
    }

    private LearningMapNode buildNode(Long mapId, LearningMapNodeDTO dto) {
        String type = LearningMapRules.nodeType(dto.getType());
        if (type == null) {
            throw new LearningMapException("节点类型不合法");
        }
        String title = LearningMapRules.text(dto.getTitle());
        String knowledge = LearningMapRules.text(dto.getKnowledgeContent());
        Long problemId = dto.getProblemId();
        String problemDisplayId = LearningMapRules.text(dto.getProblemDisplayId());
        if (LearningMapRules.KNOWLEDGE.equals(type)) {
            if (knowledge.isEmpty()) {
                throw new LearningMapException("知识点节点必须填写学习内容");
            }
            problemId = null;
            problemDisplayId = "";
        } else {
            if (problemId == null && problemDisplayId.isEmpty()) {
                throw new LearningMapException("题目节点必须填写主站题目ID或展示题号");
            }
            top.hcode.hoj.pojo.vo.LearningProblemVO problem = problemManager.get(problemId == null
                    ? problemDisplayId : String.valueOf(problemId));
            problemId = problem.getId();
            problemDisplayId = problem.getProblemDisplayId();
            if (title.isEmpty()) {
                title = problem.getTitle();
            }
            knowledge = "";
        }
        if (title.isEmpty()) {
            throw new LearningMapException("节点标题不能为空");
        }
        List<String> tags = dto.getTags() == null ? new ArrayList<>() : dto.getTags().stream()
                .map(LearningMapRules::text).filter(value -> !value.isEmpty()).collect(Collectors.toList());
        return new LearningMapNode().setMapId(mapId).setType(type).setTitle(title)
                .setDescription(LearningMapRules.text(dto.getDescription()))
                .setDifficulty(LearningMapRules.difficulty(dto.getDifficulty()))
                .setTags(LearningMapRules.writeJson(tags, "[]"))
                .setX(dto.getX() == null ? 0D : dto.getX()).setY(dto.getY() == null ? 0D : dto.getY())
                .setLevel(dto.getLevel() == null ? 0 : dto.getLevel())
                .setRegion(LearningMapRules.text(dto.getRegion()))
                .setPublished(dto.getPublished() == null || dto.getPublished())
                .setKnowledgeContent(knowledge).setProblemId(problemId).setProblemDisplayId(problemDisplayId)
                .setMetadata(LearningMapRules.writeJson(dto.getMetadata() == null ? Collections.emptyMap() : dto.getMetadata(), "{}"))
                .setCreateTime(new Date()).setUpdateTime(new Date());
    }

    private LearningMapEdge buildEdge(Long mapId, LearningMapEdgeDTO dto) {
        String type = LearningMapRules.edgeType(dto.getType());
        if (type == null) {
            throw new LearningMapException("连线类型不合法");
        }
        if (dto.getSourceNodeId() == null || dto.getTargetNodeId() == null) {
            throw new LearningMapException("起点和终点不能为空");
        }
        if (LearningMapRules.PREREQUISITE.equals(type) && Objects.equals(dto.getSourceNodeId(), dto.getTargetNodeId())) {
            throw new LearningMapException("前置依赖连线不能自环");
        }
        return new LearningMapEdge().setMapId(mapId).setSourceNodeId(dto.getSourceNodeId())
                .setTargetNodeId(dto.getTargetNodeId()).setType(type).setCreateTime(new Date()).setUpdateTime(new Date());
    }

    private void validateNodes(Long mapId, LearningMapEdge edge) {
        long count = nodeService.count(new QueryWrapper<LearningMapNode>().eq("map_id", mapId)
                .in("id", Arrays.asList(edge.getSourceNodeId(), edge.getTargetNodeId())));
        if (count != 2) {
            throw new LearningMapException("起点或终点不存在");
        }
    }

    private void validateCycle(Long mapId, LearningMapEdge edge, Long ignoredId) {
        if (LearningMapRules.PREREQUISITE.equals(edge.getType())) {
            List<LearningMapEdge> edges = edgeService.list(new QueryWrapper<LearningMapEdge>()
                    .eq("map_id", mapId).eq("type", LearningMapRules.PREREQUISITE));
            if (LearningMapRules.createsCycle(edges, ignoredId, edge.getSourceNodeId(), edge.getTargetNodeId())) {
                throw new LearningMapException("前置依赖会形成环");
            }
        }
    }

    private boolean hasCycle(List<LearningMapNode> nodes, List<LearningMapEdge> edges) {
        Set<Long> ids = nodes.stream().map(LearningMapNode::getId).collect(Collectors.toSet());
        Map<Long, Integer> degree = new HashMap<>();
        Map<Long, List<Long>> graph = new HashMap<>();
        ids.forEach(id -> degree.put(id, 0));
        for (LearningMapEdge edge : edges) {
            if (!LearningMapRules.PREREQUISITE.equals(edge.getType()) || !ids.contains(edge.getSourceNodeId())
                    || !ids.contains(edge.getTargetNodeId())) {
                continue;
            }
            graph.computeIfAbsent(edge.getSourceNodeId(), key -> new ArrayList<>()).add(edge.getTargetNodeId());
            degree.put(edge.getTargetNodeId(), degree.get(edge.getTargetNodeId()) + 1);
        }
        Deque<Long> queue = degree.entrySet().stream().filter(entry -> entry.getValue() == 0)
                .map(Map.Entry::getKey).collect(Collectors.toCollection(ArrayDeque::new));
        int visited = 0;
        while (!queue.isEmpty()) {
            Long current = queue.removeFirst();
            visited++;
            for (Long next : graph.getOrDefault(current, Collections.emptyList())) {
                degree.put(next, degree.get(next) - 1);
                if (degree.get(next) == 0) {
                    queue.addLast(next);
                }
            }
        }
        return visited != ids.size();
    }

    private void requireTitle(String title) {
        if (title.isEmpty()) {
            throw new LearningMapException("标题不能为空");
        }
    }
}
