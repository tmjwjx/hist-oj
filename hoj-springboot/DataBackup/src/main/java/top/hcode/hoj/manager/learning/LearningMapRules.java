package top.hcode.hoj.manager.learning;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;

import java.util.*;

public final class LearningMapRules {
    public static final String DRAFT = "draft";
    public static final String PUBLISHED = "published";
    public static final String ALL_OPEN = "all_open";
    public static final String ALL_CLOSED = "all_closed";
    public static final String KNOWLEDGE = "knowledge";
    public static final String PROBLEM = "problem";
    public static final String PREREQUISITE = "prerequisite";
    public static final String RELATED = "related";
    public static final String LOCKED = "locked";
    public static final String AVAILABLE = "available";
    public static final String IN_PROGRESS = "in_progress";
    public static final String COMPLETED = "completed";
    public static final String MASTERED = "mastered";

    private static final ObjectMapper JSON = new ObjectMapper();

    private LearningMapRules() {
    }

    public static String status(String value) {
        return PUBLISHED.equals(lower(value)) ? PUBLISHED : DRAFT;
    }

    public static String accessMode(String value) {
        return ALL_CLOSED.equals(lower(value)) ? ALL_CLOSED : ALL_OPEN;
    }

    public static String nodeType(String value) {
        String normalized = lower(value);
        return KNOWLEDGE.equals(normalized) || PROBLEM.equals(normalized) ? normalized : null;
    }

    public static String edgeType(String value) {
        String normalized = lower(value);
        if (normalized.isEmpty()) {
            return PREREQUISITE;
        }
        return PREREQUISITE.equals(normalized) || RELATED.equals(normalized) ? normalized : null;
    }

    public static String difficulty(String value) {
        String normalized = lower(value);
        return Arrays.asList("beginner", "easy", "medium", "hard", "expert").contains(normalized)
                ? normalized : "beginner";
    }

    public static String text(String value) {
        return value == null ? "" : value.trim();
    }

    public static String writeJson(Object value, String fallback) {
        try {
            return JSON.writeValueAsString(value);
        } catch (Exception ignored) {
            return fallback;
        }
    }

    public static List<String> tags(String raw) {
        if (text(raw).isEmpty()) {
            return new ArrayList<>();
        }
        try {
            return JSON.readValue(raw, new TypeReference<List<String>>() {
            });
        } catch (Exception ignored) {
            List<String> values = new ArrayList<>();
            for (String value : raw.split(",")) {
                if (!text(value).isEmpty()) {
                    values.add(text(value));
                }
            }
            return values;
        }
    }

    public static Map<String, Object> metadata(String raw) {
        if (text(raw).isEmpty()) {
            return new LinkedHashMap<>();
        }
        try {
            return JSON.readValue(raw, new TypeReference<Map<String, Object>>() {
            });
        } catch (Exception ignored) {
            return new LinkedHashMap<>();
        }
    }

    public static String remark(String raw) {
        Object remark = metadata(raw).get("remark");
        return remark instanceof String ? text((String) remark) : "";
    }

    public static boolean createsCycle(List<LearningMapEdge> edges, Long ignoredId,
                                       Long sourceId, Long targetId) {
        if (sourceId == null || targetId == null) {
            return false;
        }
        Map<Long, List<Long>> graph = new HashMap<>();
        for (LearningMapEdge edge : edges) {
            if (Objects.equals(edge.getId(), ignoredId) || !PREREQUISITE.equals(edge.getType())) {
                continue;
            }
            graph.computeIfAbsent(edge.getSourceNodeId(), key -> new ArrayList<>())
                    .add(edge.getTargetNodeId());
        }
        Deque<Long> queue = new ArrayDeque<>();
        Set<Long> visited = new HashSet<>();
        queue.add(targetId);
        visited.add(targetId);
        while (!queue.isEmpty()) {
            Long current = queue.removeFirst();
            if (Objects.equals(current, sourceId)) {
                return true;
            }
            for (Long next : graph.getOrDefault(current, Collections.emptyList())) {
                if (visited.add(next)) {
                    queue.addLast(next);
                }
            }
        }
        return false;
    }

    private static String lower(String value) {
        return text(value).toLowerCase(Locale.ROOT);
    }
}
