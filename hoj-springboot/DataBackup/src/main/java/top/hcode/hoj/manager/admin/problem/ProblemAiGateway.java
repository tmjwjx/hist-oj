package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.springframework.http.*;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
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
    private static final String REASONING_EFFORT = "xhigh";
    private static final int CONNECT_TIMEOUT_MS = 15_000;
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
                if (RETRYABLE_STATUS.contains(status) && attempt < MAX_ATTEMPTS) {
                    pauseBeforeRetry();
                    attempt++;
                    continue;
                }
                throw upstreamFailure(status);
            } catch (ResourceAccessException e) {
                throw new IllegalStateException(
                        "AI 服务连接或读取超时，请稍后重试；若持续失败，请检查管理员配置的 AI 直连地址", e);
            }
        }
        throw new IllegalStateException("AI 服务暂时不可用，请稍后重试");
    }

    private String execute(ProblemAiConfig config, List<Map<String, String>> messages, boolean stream) {
        byte[] payload = JSONUtil.toJsonStr(requestBody(config, messages, stream)).getBytes(StandardCharsets.UTF_8);
        RestTemplate client = new RestTemplate(requestFactory(config));
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
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("model", config.getModel());
        body.put("messages", messages);
        if (!isReasoningModel(config.getModel())) body.put("temperature", 0.1);
        body.put("reasoning_effort", REASONING_EFFORT);
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

    private SimpleClientHttpRequestFactory requestFactory(ProblemAiConfig config) {
        int readSeconds = config.getTimeoutSeconds() == null ? 120 : config.getTimeoutSeconds();
        readSeconds = Math.max(30, Math.min(readSeconds, 900));
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
                if (line.startsWith("data:")) {
                    sse = true;
                    appendStreamEvent(streamed, line.substring(5).trim());
                } else if (!sse) {
                    raw.append(line).append('\n');
                }
            }
        }
        String content = sse ? streamed.toString() : extractMessageContent(raw.toString());
        if (StringUtils.isEmpty(content)) {
            throw new IllegalStateException("AI 接口未返回有效内容，请检查模型和接口兼容性");
        }
        return content;
    }

    private void appendStreamEvent(StringBuilder content, String data) {
        if (StringUtils.isEmpty(data) || "[DONE]".equals(data)) return;
        try {
            JSONObject event = JSONUtil.parseObj(data);
            String delta = event.getByPath("choices[0].delta.content", String.class);
            if (!StringUtils.isEmpty(delta)) {
                content.append(delta);
                return;
            }
            if (content.length() == 0) {
                String message = event.getByPath("choices[0].message.content", String.class);
                if (StringUtils.isEmpty(message)) {
                    message = event.getByPath("choices[0].text", String.class);
                }
                if (!StringUtils.isEmpty(message)) content.append(message);
            }
        } catch (Exception ignored) {
            // 心跳或非标准扩展事件不应中断后续 SSE 内容读取。
        }
    }

    private String extractMessageContent(String responseBody) {
        try {
            return JSONUtil.parseObj(responseBody)
                    .getByPath("choices[0].message.content", String.class);
        } catch (Exception e) {
            throw new IllegalStateException("AI 接口返回格式不兼容，请检查管理员配置的接口地址和模型");
        }
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
