package top.hcode.hoj.manager.plagiarism;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import javax.annotation.Resource;
import java.util.*;

/**
 * Client for the self-hosted Dolos worker used by contest plagiarism checks.
 *
 * <p>Request contract:</p>
 * <pre>
 * POST {worker-url}/v1/analyze
 * {
 *   "language": "python",
 *   "submissions": [{"id": "1", "code": "..."}]
 * }
 * </pre>
 *
 * <p>The response must contain a {@code pairs} array. A pair contains the two
 * submission ids and either two directional fields ({@code similarity1to2},
 * {@code similarity2to1}) or one {@code similarity} field. A few snake_case
 * and coverage aliases are accepted to keep the adapter compatible with thin
 * worker wrappers around Dolos.</p>
 */
@Component
public class PlagiarismDolosWorkerClient {
    private static final ObjectMapper JSON = new ObjectMapper();
    private static final String ENDPOINT = "/v1/analyze";

    @Resource
    private PlagiarismDolosProperties properties;

    public boolean isConfigured() {
        return properties.isConfigured();
    }

    public Map<String, int[]> analyze(String language, List<Submission> submissions) {
        if (!properties.isConfigured()) {
            throw new IllegalStateException("Dolos worker 未配置");
        }
        if (submissions == null || submissions.size() < 2) {
            return Collections.emptyMap();
        }
        Map<String, Object> request = new LinkedHashMap<>();
        request.put("language", normalizeLanguage(language));
        request.put("submissions", submissions);
        // The current result schema persists pair scores only. Omitting
        // fragment coordinates keeps large contest responses bounded; the
        // existing detail view still loads the two complete submissions.
        request.put("fragmentLimit", 0);
        try {
            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.setAccept(Collections.singletonList(MediaType.APPLICATION_JSON));
            if (properties.getToken() != null && !properties.getToken().trim().isEmpty()) {
                headers.setBearerAuth(properties.getToken().trim());
            }
            HttpEntity<String> entity = new HttpEntity<>(JSON.writeValueAsString(request), headers);
            ResponseEntity<String> response = client().postForEntity(endpoint(), entity, String.class);
            if (!response.getStatusCode().is2xxSuccessful() || response.getBody() == null) {
                throw new IllegalStateException("Dolos worker returned HTTP " + response.getStatusCodeValue()
                        + ": " + responseMessage(response.getBody()));
            }
            return parsePairs(response.getBody());
        } catch (Exception ex) {
            throw new IllegalStateException("Dolos worker unavailable: " + ex.getMessage(), ex);
        }
    }

    private String responseMessage(String body) {
        if (body == null || body.trim().isEmpty()) return "empty response";
        try {
            JsonNode root = JSON.readTree(body);
            JsonNode error = root == null ? null : root.get("error");
            JsonNode message = error == null ? null : error.get("message");
            return message == null ? body : message.asText(body);
        } catch (Exception ignored) {
            return body.length() > 256 ? body.substring(0, 256) : body;
        }
    }

    Map<String, int[]> parsePairs(String body) throws Exception {
        JsonNode root = JSON.readTree(body);
        JsonNode pairs = root == null ? null : root.get("pairs");
        if (pairs == null && root != null) pairs = root.get("results");
        if (pairs == null && root != null && root.get("data") != null) {
            pairs = root.get("data").get("pairs");
        }
        if (pairs == null || !pairs.isArray()) {
            throw new IllegalArgumentException("Dolos worker response does not contain an array named pairs");
        }
        Map<String, int[]> result = new HashMap<>();
        for (JsonNode pair : pairs) {
            String first = text(pair, "submissionId1", "submission_id_1", "leftSubmissionId", "left_submission_id",
                    "leftId", "left_id", "sourceId", "source_id", "file1");
            String second = text(pair, "submissionId2", "submission_id_2", "rightSubmissionId", "right_submission_id",
                    "rightId", "right_id", "targetId", "target_id", "file2");
            if (first == null || second == null || first.equals(second)) continue;
            Integer firstToSecond = score(pair, "similarity1to2", "similarity_1_to_2",
                    "similarityLeftToRight", "similarity_left_to_right", "coverage1", "coverage_1");
            Integer secondToFirst = score(pair, "similarity2to1", "similarity_2_to_1",
                    "similarityRightToLeft", "similarity_right_to_left", "coverage2", "coverage_2");
            Integer symmetric = score(pair, "similarityPercent", "similarity", "score");
            if (firstToSecond == null) firstToSecond = symmetric;
            if (secondToFirst == null) secondToFirst = symmetric;
            if (firstToSecond == null || secondToFirst == null) continue;
            if (first.compareTo(second) <= 0) {
                result.put(pairKey(first, second), new int[]{firstToSecond, secondToFirst});
            } else {
                result.put(pairKey(first, second), new int[]{secondToFirst, firstToSecond});
            }
        }
        return result;
    }

    private String text(JsonNode node, String... names) {
        for (String name : names) {
            JsonNode value = node.get(name);
            if (value != null && !value.isNull() && value.asText() != null && !value.asText().trim().isEmpty()) {
                return value.asText();
            }
        }
        return null;
    }

    private Integer score(JsonNode node, String... names) {
        for (String name : names) {
            JsonNode value = node.get(name);
            if (value == null || value.isNull() || (!value.isNumber() && !value.isTextual())) continue;
            try {
                double number = value.asDouble();
                if (isFractionField(name) && number >= 0D && number <= 1D) number *= 100D;
                return Math.max(0, Math.min(100, (int) Math.round(number)));
            } catch (Exception ignored) {
                // Try the next compatible alias.
            }
        }
        return null;
    }

    private boolean isFractionField(String name) {
        return "similarity".equals(name)
                || "score".equals(name)
                || name.startsWith("coverage")
                || name.startsWith("similarityLeftToRight")
                || name.startsWith("similarityRightToLeft")
                || name.startsWith("similarity_left_to_right")
                || name.startsWith("similarity_right_to_left");
    }

    public String normalizeLanguage(String language) {
        String value = language == null ? "" : language.toLowerCase(Locale.ROOT);
        if (value.contains("python") || value.contains("pypy") || value.contains("py3")) return "python";
        if (value.contains("javascript") || value.contains("node") || value.contains("js")) return "javascript";
        if (value.contains("typescript") || value.contains("ts")) return "typescript";
        if (value.contains("golang") || value.equals("go") || value.startsWith("go ")) return "go";
        if (value.contains("rust") || value.equals("rs")) return "rust";
        if (value.contains("c++") || value.contains("cpp") || value.contains("g++") || value.contains("gcc")) return "cpp";
        if (value.equals("c") || value.startsWith("c ")) return "c";
        if (value.contains("c#") || value.contains("csharp") || value.contains("dotnet")) return "csharp";
        if (value.contains("java") || value.contains("kotlin")) return value.contains("kotlin") ? "kotlin" : "java";
        if (value.contains("php")) return "php";
        if (value.contains("ruby") || value.equals("rb")) return "ruby";
        if (value.contains("swift")) return "swift";
        if (value.contains("scala")) return "scala";
        return value.trim();
    }

    public static String pairKey(String first, String second) {
        return first.compareTo(second) <= 0 ? first + "|" + second : second + "|" + first;
    }

    private String endpoint() {
        String base = properties.getWorkerUrl().trim();
        while (base.endsWith("/")) base = base.substring(0, base.length() - 1);
        return base.endsWith(ENDPOINT) ? base : base + ENDPOINT;
    }

    private RestTemplate client() {
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(Math.min(properties.getTimeoutMs(), 15000));
        factory.setReadTimeout(properties.getTimeoutMs());
        return new RestTemplate(factory);
    }

    public static class Submission {
        private final String id;
        private final String code;

        public Submission(String id, String code) {
            this.id = id;
            this.code = code == null ? "" : code;
        }

        public String getId() {
            return id;
        }

        public String getCode() {
            return code;
        }
    }
}
