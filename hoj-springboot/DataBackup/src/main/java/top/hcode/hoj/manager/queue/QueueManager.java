package top.hcode.hoj.manager.queue;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.io.ClassPathResource;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.script.DefaultRedisScript;
import org.springframework.stereotype.Component;

import java.util.Arrays;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.regex.Pattern;

/** Page-load tickets; every transition and returned count is one atomic Redis operation. */
@Component
public class QueueManager {

    public static final long COMPLETION_MAX_AGE_SECONDS = 86400;

    private static final Pattern UUID_PATTERN = Pattern.compile(
            "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}");
    static final List<String> KEYS = Collections.unmodifiableList(Arrays.asList(
            "hoj:q:v2:{queue}:waiting", "hoj:q:v2:{queue}:seen",
            "hoj:q:v2:{queue}:reserved", "hoj:q:v2:{queue}:loading",
            "hoj:q:v2:{queue}:departing", "hoj:q:v2:{queue}:closed",
            "hoj:q:v2:{queue}:sequence", "hoj:q:v2:{queue}:revision",
            "hoj:q:v2:{queue}:warm"));
    private static final DefaultRedisScript<List> SCRIPT = new DefaultRedisScript<>();

    static {
        SCRIPT.setLocation(new ClassPathResource("queue.lua"));
        SCRIPT.setResultType(List.class);
    }

    private final StringRedisTemplate redis;
    private final boolean enabled;
    private final int maxSlots;

    public QueueManager(StringRedisTemplate redis,
                        @Value("${hoj.queue.enabled:true}") boolean enabled,
                        @Value("${hoj.queue.max-slots:5}") int maxSlots) {
        if (maxSlots <= 0) {
            throw new IllegalArgumentException("hoj.queue.max-slots must be positive");
        }
        this.redis = redis;
        this.enabled = enabled;
        this.maxSlots = maxSlots;
    }

    public Map<String, Object> status(String id) {
        return execute("status", identity(id), false);
    }

    public Map<String, Object> heartbeat(String id) {
        return execute("heartbeat", identity(id), false);
    }

    public Map<String, Object> release(String id) {
        return execute("release", identity(id), false);
    }

    public Map<String, Object> complete(String id) {
        return execute("complete", identity(id), false);
    }

    public boolean reuse(String token) {
        String id = identity(token);
        if (!enabled) return true;
        List<?> values = redis.execute(SCRIPT, KEYS, "reuse", id,
                Integer.toString(maxSlots), "0", "");
        if (values == null || values.size() != 1) {
            throw new IllegalStateException("Invalid queue reuse response");
        }
        return ((Number) values.get(0)).longValue() == 1;
    }

    public Map<String, Object> leave(String id, boolean immediate) {
        return execute("leave", identity(id), immediate);
    }

    public Map<String, Object> claim(String id) {
        return execute("claim", identity(id), false);
    }

    public void cleanup() {
        execute("cleanup", "", false);
    }

    private String identity(String id) {
        if (id == null || !UUID_PATTERN.matcher(id).matches()) {
            throw new IllegalArgumentException("qid must be a canonical UUID");
        }
        return UUID.fromString(id).toString();
    }

    private Map<String, Object> execute(String operation, String id, boolean immediate) {
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("myId", id);
        result.put("maxSlots", maxSlots);
        result.put("enabled", enabled);
        result.put("retryAfterMs", 1000);
        if (!enabled) {
            result.put("state", "disabled");
            result.put("disabled", true);
            result.put("admitted", true);
            result.put("claimed", "claim".equals(operation));
            result.put("position", 0);
            result.put("waiting", 0);
            result.put("reserved", 0);
            result.put("loading", 0);
            result.put("active", 0);
            result.put("total", 0);
            result.put("revision", System.currentTimeMillis() * 1000);
            return result;
        }
        String completionToken = "complete".equals(operation) ? UUID.randomUUID().toString() : "";
        List<?> values = redis.execute(SCRIPT, KEYS, operation, id,
                Integer.toString(maxSlots), immediate ? "1" : "0", completionToken);
        if (values == null || values.size() != 8) {
            throw new IllegalStateException("Invalid queue script response");
        }
        String state = (String) values.get(0);
        long waiting = ((Number) values.get(2)).longValue();
        long reserved = ((Number) values.get(3)).longValue();
        long loading = ((Number) values.get(4)).longValue();
        long position = ((Number) values.get(1)).longValue();
        result.put("state", state);
        result.put("admitted", "reserved".equals(state) || "loading".equals(state));
        result.put("position", position);
        result.put("retryAfterMs", position > 100 ? 5000 : position > 20 ? 3000 : 1000);
        result.put("waiting", waiting);
        result.put("reserved", reserved);
        result.put("loading", loading);
        result.put("active", reserved + loading);
        result.put("total", waiting + reserved + loading);
        result.put("revision", ((Number) values.get(5)).longValue());
        result.put("claimed", ((Number) values.get(6)).longValue() == 1);
        boolean completed = ((Number) values.get(7)).longValue() == 1;
        result.put("completed", completed);
        if (completed) result.put("completionToken", completionToken);
        return result;
    }
}
