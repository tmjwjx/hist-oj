package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.springframework.http.*;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestClientResponseException;
import org.springframework.web.client.RestTemplate;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.*;

@Component
public class ProblemAiGateway {
    private static final Logger log = LoggerFactory.getLogger(ProblemAiGateway.class);
    private static final String REASONING_EFFORT = "xhigh";
    private static final int CONNECT_TIMEOUT_MS = 15_000;
    private static final int MAX_READ_TIMEOUT_SECONDS = 3_600;
    private static final int MAX_OUTPUT_TOKENS = 16_384;
    private static final int MAX_ATTEMPTS = 2;
    private static final Set<Integer> RETRYABLE_STATUS =
            new HashSet<>(Arrays.asList(502, 503, 504, 524));

    public String complete(ProblemAiConfig config, List<Map<String, String>> messages) {
        int attempt = 1;
        boolean stream = true;
        boolean fallbackUsed = false;
        while (attempt <= MAX_ATTEMPTS) {
            try {
                return execute(config, messages, stream);
            } catch (RestClientResponseException e) {
                // Some OpenAI-compatible gateways reject stream=true while still
                // supporting the regular JSON response. Keep this compatibility
                // fallback within the same retry attempt and do not repeat it.
                if (stream && !fallbackUsed && streamUnsupported(e)) {
                    stream = false;
                    fallbackUsed = true;
                    continue;
                }
                int status = e.getRawStatusCode();
                // A few CDN/proxy layers return 502/503/504/524 after waiting for
                // a streamed response. Retry the same request as regular JSON
                // first; this avoids repeating another long-lived SSE request.
                if (stream && !fallbackUsed && RETRYABLE_STATUS.contains(status)) {
                    stream = false;
                    fallbackUsed = true;
                    continue;
                }
                if (RETRYABLE_STATUS.contains(status) && attempt < MAX_ATTEMPTS) {
                    pauseBeforeRetry();
                    attempt++;
                    continue;
                }
                throw upstreamFailure(status);
            } catch (ResourceAccessException e) {
                if (stream && !fallbackUsed) {
                    stream = false;
                    fallbackUsed = true;
                    continue;
                }
                throw new IllegalStateException(
                        "AI 服务连接或读取超时，请稍后重试；若持续失败，请检查管理员配置的 AI 直连地址", e);
            } catch (IllegalStateException e) {
                // A successful HTTP response can still contain an empty or
                // provider-specific stream. Try one non-stream request before
                // exposing the compatibility error to the caller.
                if (stream && !fallbackUsed && responseCompatibilityFailure(e)) {
                    stream = false;
                    fallbackUsed = true;
                    continue;
                }
                throw e;
            }
        }
        throw new IllegalStateException("AI 服务暂时不可用，请稍后重试");
    }

    /**
     * Report-only operations should have a bounded, single upstream request.
     * The normal streaming/retry path can otherwise spend the configured read
     * timeout once for SSE and once again for the non-stream fallback.
     */
    public String completeOnce(ProblemAiConfig config, List<Map<String, String>> messages,
                               int timeoutSeconds) {
        int bounded = Math.max(30, Math.min(timeoutSeconds, MAX_READ_TIMEOUT_SECONDS));
        try {
            return execute(config, messages, false, bounded, "medium");
        } catch (RestClientResponseException e) {
            throw upstreamFailure(e.getRawStatusCode());
        } catch (ResourceAccessException e) {
            throw new IllegalStateException("AI 复检请求超时（超过 " + bounded + " 秒），请稍后重试", e);
        }
    }

    private String execute(ProblemAiConfig config, List<Map<String, String>> messages, boolean stream) {
        int readSeconds = config.getTimeoutSeconds() == null ? 120 : config.getTimeoutSeconds();
        return execute(config, messages, stream, readSeconds);
    }

    private String execute(ProblemAiConfig config, List<Map<String, String>> messages, boolean stream,
                           int timeoutSeconds) {
        return execute(config, messages, stream, timeoutSeconds, REASONING_EFFORT);
    }

    private String execute(ProblemAiConfig config, List<Map<String, String>> messages, boolean stream,
                           int timeoutSeconds, String reasoningEffort) {
        byte[] payload = JSONUtil.toJsonStr(requestBody(config, messages, stream, reasoningEffort))
                .getBytes(StandardCharsets.UTF_8);
        RestTemplate client = new RestTemplate(requestFactory(timeoutSeconds));
        return client.execute(chatCompletionsUrl(config.getApiUrl()), HttpMethod.POST, request -> {
            HttpHeaders headers = request.getHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.setAccept(Arrays.asList(MediaType.TEXT_EVENT_STREAM, MediaType.APPLICATION_JSON));
            headers.setBearerAuth(config.getApiKey());
            headers.setCacheControl(CacheControl.noCache());
            headers.set(HttpHeaders.ACCEPT_ENCODING, "identity");
            headers.set(HttpHeaders.USER_AGENT, "Mozilla/5.0 (compatible; HistOJ-AIValidator/1.0)");
            request.getBody().write(payload);
        }, this::readCompletion);
    }

    private Map<String, Object> requestBody(ProblemAiConfig config, List<Map<String, String>> messages,
                                            boolean stream) {
        return requestBody(config, messages, stream, REASONING_EFFORT);
    }

    private Map<String, Object> requestBody(ProblemAiConfig config, List<Map<String, String>> messages,
                                            boolean stream, String reasoningEffort) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("model", config.getModel());
        body.put("messages", messages);
        if (!isReasoningModel(config.getModel())) body.put("temperature", 0.1);
        body.put("reasoning_effort", reasoningEffort);
        body.put(completionTokenField(config.getModel()), MAX_OUTPUT_TOKENS);
        body.put("stream", stream);
        return body;
    }

    private boolean streamUnsupported(RestClientResponseException error) {
        int status = error.getRawStatusCode();
        if (status != 400 && status != 422) return false;
        String body = error.getResponseBodyAsString();
        if (body == null || body.length() > 4096) return false;
        String value = body.toLowerCase(Locale.ROOT);
        return value.contains("stream")
                && (value.contains("unsupported") || value.contains("not support")
                || value.contains("invalid") || value.contains("unknown"));
    }

    private SimpleClientHttpRequestFactory requestFactory(int timeoutSeconds) {
        int readSeconds = Math.max(30, Math.min(timeoutSeconds, MAX_READ_TIMEOUT_SECONDS));
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(CONNECT_TIMEOUT_MS);
        factory.setReadTimeout(readSeconds * 1000);
        return factory;
    }

    private String readCompletion(org.springframework.http.client.ClientHttpResponse response)
            throws java.io.IOException {
        StringBuilder raw = new StringBuilder();
        StringBuilder streamed = new StringBuilder();
        boolean sse = false;
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(response.getBody(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                // Strip an optional UTF-8 BOM and tolerate the whitespace used
                // by a few SSE proxies before the field name.
                String eventLine = line;
                if (!eventLine.isEmpty() && eventLine.charAt(0) == '\uFEFF') {
                    eventLine = eventLine.substring(1);
                }
                eventLine = eventLine.trim();
                if (eventLine.startsWith("data:")) {
                    sse = true;
                    appendStreamEvent(streamed, eventLine.substring(5).trim());
                } else if (!sse) {
                    raw.append(line).append('\n');
                }
            }
        }
        String content = sse ? streamed.toString() : extractMessageContent(raw.toString());
        if (StringUtils.isEmpty(content)) {
            log.warn("AI response contained no text: status={}, contentType={}, sse={}, rawChars={}, streamChars={}",
                    response.getRawStatusCode(), response.getHeaders().getContentType(), sse,
                    raw.length(), streamed.length());
            throw new IllegalStateException("AI 接口未返回有效内容，请检查模型和接口兼容性");
        }
        return content;
    }

    private void appendStreamEvent(StringBuilder content, String data) {
        if (StringUtils.isEmpty(data) || "[DONE]".equals(data)) return;
        try {
            JSONObject event = JSONUtil.parseObj(data);
            appendText(content, firstText(event,
                    "choices[0].delta.content",
                    "choices[0].message.content",
                    "choices[0].text",
                    "output_text",
                    "output[0].content[0].text"));
        } catch (Exception ignored) {
            // 心跳或非标准扩展事件不应中断后续 SSE 内容读取。
        }
    }

    private String extractMessageContent(String responseBody) {
        try {
            JSONObject response = JSONUtil.parseObj(responseBody);
            return firstText(response,
                    "choices[0].message.content",
                    "choices[0].text",
                    "output_text",
                    "output[0].content[0].text");
        } catch (Exception e) {
            throw new IllegalStateException("AI 接口返回格式不兼容，请检查管理员配置的接口地址和模型");
        }
    }

    private String firstText(JSONObject value, String... paths) {
        for (String path : paths) {
            try {
                String text = textValue(value.getByPath(path));
                if (!StringUtils.isEmpty(text)) return text;
            } catch (Exception ignored) {
                // Continue with the next compatible response shape.
            }
        }
        return "";
    }

    /**
     * OpenAI-compatible providers normally return a String, but some return
     * content parts ({@code [{"type":"text","text":"..."}]}). Extract
     * only text parts so their JSON envelope is not appended to generated code.
     */
    private String textValue(Object value) {
        if (value == null) return "";
        if (value instanceof CharSequence) return value.toString();
        if (value instanceof JSONObject) {
            JSONObject object = (JSONObject) value;
            String text = textValue(object.get("text"));
            if (!StringUtils.isEmpty(text)) return text;
            return textValue(object.get("content"));
        }
        if (value instanceof Map) {
            Map<?, ?> object = (Map<?, ?>) value;
            String text = textValue(object.get("text"));
            if (!StringUtils.isEmpty(text)) return text;
            return textValue(object.get("content"));
        }
        if (value instanceof JSONArray) {
            StringBuilder result = new StringBuilder();
            for (Object item : (JSONArray) value) result.append(textValue(item));
            return result.toString();
        }
        if (value instanceof Iterable) {
            StringBuilder result = new StringBuilder();
            for (Object item : (Iterable<?>) value) result.append(textValue(item));
            return result.toString();
        }
        return String.valueOf(value);
    }

    private void appendText(StringBuilder target, String text) {
        if (!StringUtils.isEmpty(text)) target.append(text);
    }

    private boolean responseCompatibilityFailure(IllegalStateException error) {
        String message = error.getMessage();
        return message != null && (message.contains("未返回有效内容") || message.contains("返回格式不兼容"));
    }

    private IllegalStateException upstreamFailure(int status) {
        if (status == 524 || status == 504) {
            return new IllegalStateException(
                    "AI 上游网关生成超时（HTTP " + status + "），系统已自动重试一次；请稍后重试或改用 AI 直连地址");
        }
        if (status == 502 || status == 503) {
            return new IllegalStateException(
                    "AI 上游服务暂时不可用（HTTP " + status + "），系统已自动重试一次，请稍后再试");
        }
        if (status == 401 || status == 403) {
            return new IllegalStateException("AI 服务认证失败，请检查管理员配置的 API Key");
        }
        if (status == 429) {
            return new IllegalStateException("AI 服务请求过于频繁或额度不足，请稍后重试");
        }
        return new IllegalStateException("AI 服务请求失败（HTTP " + status + "），请检查接口地址、模型和服务状态");
    }

    private void pauseBeforeRetry() {
        try {
            Thread.sleep(800L);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new IllegalStateException("AI 请求已取消", e);
        }
    }

    private String completionTokenField(String model) {
        String value = model == null ? "" : model.toLowerCase(Locale.ROOT);
        return isReasoningModel(value) ? "max_completion_tokens" : "max_tokens";
    }

    private boolean isReasoningModel(String model) {
        String value = model == null ? "" : model.toLowerCase(Locale.ROOT);
        return value.startsWith("o1") || value.startsWith("o3") || value.startsWith("o4")
                || value.startsWith("gpt-5");
    }

    private String chatCompletionsUrl(String apiUrl) {
        String value = apiUrl == null ? "" : apiUrl.trim().replaceAll("/+$", "");
        return value.endsWith("/v1") ? value + "/chat/completions" : value;
    }
}
