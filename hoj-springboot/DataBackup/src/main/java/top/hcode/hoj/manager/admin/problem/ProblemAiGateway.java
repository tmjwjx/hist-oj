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
import java.net.HttpURLConnection;
import java.net.URI;
import java.net.URLDecoder;
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
                        "AI 服务连接或读取超时，请稍后重试；若持续失败，请检查管理员配置的 AI 直连地址");
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
            throw new IllegalStateException("AI 复检请求超时（超过 " + bounded + " 秒），请稍后重试");
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
        String protocol = protocol(config);
        String apiKey = config.getApiKey();
        if (apiKey == null || apiKey.trim().isEmpty()
                || apiKey.chars().anyMatch(value -> value < 32 || value == 127)) {
            throw new IllegalStateException("AI API Key 格式无效，请检查管理员配置");
        }
        URI url = endpoint(config, protocol, stream);
        byte[] payload = JSONUtil.toJsonStr(requestBody(config, messages, stream, reasoningEffort))
                .getBytes(StandardCharsets.UTF_8);
        RestTemplate client = new RestTemplate(requestFactory(timeoutSeconds));
        return client.execute(url, HttpMethod.POST, request -> {
            HttpHeaders headers = request.getHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.setAccept(Arrays.asList(MediaType.TEXT_EVENT_STREAM, MediaType.APPLICATION_JSON));
            if ("gemini".equals(protocol)) headers.set("x-goog-api-key", config.getApiKey());
            else if ("anthropic".equals(protocol)) {
                headers.set("x-api-key", config.getApiKey());
                headers.set("anthropic-version", "2023-06-01");
            } else headers.setBearerAuth(config.getApiKey());
            headers.setCacheControl(CacheControl.noCache());
            headers.set(HttpHeaders.ACCEPT_ENCODING, "identity");
            headers.set(HttpHeaders.USER_AGENT, "Mozilla/5.0 (compatible; HistOJ-AIValidator/1.0)");
            request.getBody().write(payload);
        }, response -> readCompletion(response, protocol));
    }

    private Map<String, Object> requestBody(ProblemAiConfig config, List<Map<String, String>> messages,
                                            boolean stream) {
        return requestBody(config, messages, stream, REASONING_EFFORT);
    }

    private Map<String, Object> requestBody(ProblemAiConfig config, List<Map<String, String>> messages,
                                            boolean stream, String reasoningEffort) {
        String protocol = protocol(config);
        Map<String, Object> body = new LinkedHashMap<>();
        if ("openai_chat".equals(protocol)) {
            body.put("model", config.getModel());
            body.put("messages", messages);
            if (isReasoningModel(config.getModel())) body.put("reasoning_effort", reasoningEffort);
            else body.put("temperature", 0.1);
            body.put(completionTokenField(config.getModel()), MAX_OUTPUT_TOKENS);
            body.put("stream", stream);
            return body;
        }
        StringBuilder system = new StringBuilder();
        List<Map<String, Object>> conversation = new ArrayList<>();
        for (Map<String, String> message : messages) {
            String role = message.getOrDefault("role", "user");
            String content = message.getOrDefault("content", "");
            if ("system".equals(role) || "developer".equals(role)) {
                if (system.length() > 0) system.append("\n\n");
                system.append(content);
                continue;
            }
            if (!"user".equals(role) && !"assistant".equals(role)) {
                throw new IllegalStateException("AI 验题仅支持文本对话消息");
            }
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("role", "gemini".equals(protocol) && "assistant".equals(role) ? "model" : role);
            if ("gemini".equals(protocol)) {
                item.put("parts", Collections.singletonList(Collections.singletonMap("text", content)));
            } else item.put("content", content);
            conversation.add(item);
        }
        if ("gemini".equals(protocol)) {
            body.put("contents", conversation);
            if (system.length() > 0) body.put("systemInstruction", Collections.singletonMap("parts",
                    Collections.singletonList(Collections.singletonMap("text", system.toString()))));
            Map<String, Object> generation = new LinkedHashMap<>();
            generation.put("maxOutputTokens", MAX_OUTPUT_TOKENS);
            generation.put("temperature", 0.1);
            body.put("generationConfig", generation);
        } else {
            body.put("model", config.getModel());
            body.put("stream", stream);
            if ("openai_responses".equals(protocol)) {
                body.put("input", conversation);
                if (system.length() > 0) body.put("instructions", system.toString());
                body.put("store", false);
                body.put("max_output_tokens", MAX_OUTPUT_TOKENS);
                if (isReasoningModel(config.getModel())) {
                    body.put("reasoning", Collections.singletonMap("effort", reasoningEffort));
                } else body.put("temperature", 0.1);
            } else {
                body.put("messages", conversation);
                if (system.length() > 0) body.put("system", system.toString());
                body.put("max_tokens", MAX_OUTPUT_TOKENS);
                // Native Claude models have differing temperature restrictions;
                // their documented default works without a model-name guess.
            }
        }
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
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory() {
            @Override
            protected void prepareConnection(HttpURLConnection connection, String method)
                    throws java.io.IOException {
                super.prepareConnection(connection, method);
                connection.setInstanceFollowRedirects(false);
            }
        };
        factory.setConnectTimeout(CONNECT_TIMEOUT_MS);
        factory.setReadTimeout(readSeconds * 1000);
        return factory;
    }

    private static class AiResponseException extends IllegalStateException {
        AiResponseException(String message) { super(message); }
    }

    private String readCompletion(org.springframework.http.client.ClientHttpResponse response, String protocol)
            throws java.io.IOException {
        if (response.getRawStatusCode() >= 300) throw upstreamFailure(response.getRawStatusCode());
        try {
            return readCompletionBody(response, protocol);
        } catch (AiResponseException error) {
            throw error;
        } catch (RuntimeException error) {
            // Parsing libraries may embed the raw response in their exceptions.
            throw new AiResponseException("AI 接口返回格式不兼容，请检查管理员配置的接口地址和模型");
        }
    }

    private String readCompletionBody(org.springframework.http.client.ClientHttpResponse response, String protocol)
            throws java.io.IOException {
        StringBuilder raw = new StringBuilder();
        StringBuilder eventData = new StringBuilder();
        StreamResult result = new StreamResult();
        String eventName = "";
        boolean sse = response.getHeaders().getContentType() != null
                && MediaType.TEXT_EVENT_STREAM.isCompatibleWith(response.getHeaders().getContentType());
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(response.getBody(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                if (!line.isEmpty() && line.charAt(0) == '\uFEFF') line = line.substring(1);
                String eventLine = line.trim();
                if (eventLine.startsWith("data:") || eventLine.startsWith("event:")
                        || eventLine.startsWith(":")) sse = true;
                if (!sse) {
                    raw.append(line).append('\n');
                    continue;
                }
                if (eventLine.isEmpty()) {
                    consumeEvent(result, eventName, eventData.toString(), protocol);
                    eventData.setLength(0);
                    eventName = "";
                } else if (eventLine.startsWith("event:")) {
                    eventName = eventLine.substring(6).trim();
                } else if (eventLine.startsWith("data:")) {
                    // Some compatible proxies omit the blank line between JSON events.
                    if (eventData.length() > 0 && (JSONUtil.isJsonObj(eventData.toString())
                            || "[DONE]".contentEquals(eventData))) {
                        consumeEvent(result, eventName, eventData.toString(), protocol);
                        eventData.setLength(0);
                    }
                    if (eventData.length() > 0) eventData.append('\n');
                    eventData.append(eventLine.substring(5).trim());
                }
            }
        }
        if (sse) {
            consumeEvent(result, eventName, eventData.toString(), protocol);
            if (!result.completed) throw new AiResponseException("AI 流式响应中断，未收到完整结束标记，请重试");
        }
        String content = sse ? result.text.toString() : extractMessageContent(raw.toString(), protocol);
        if (StringUtils.isEmpty(content)) {
            log.warn("AI response contained no text: status={}, protocol={}, sse={}, rawChars={}, streamChars={}",
                    response.getRawStatusCode(), protocol, sse, raw.length(), result.text.length());
            throw new AiResponseException("AI 接口未返回有效内容，请检查模型和接口兼容性");
        }
        return content;
    }

    private static class StreamResult {
        private final StringBuilder text = new StringBuilder();
        private boolean completed;
        private boolean stopReasonReceived;
    }

    private void consumeEvent(StreamResult result, String eventName, String data, String protocol) {
        if ("error".equals(eventName)) throw responseFailure();
        if (StringUtils.isEmpty(data)) return;
        if ("[DONE]".equals(data)) {
            if ("openai_chat".equals(protocol)) result.completed = true;
            return;
        }
        JSONObject event = parseResponse(data);
        rejectError(event);
        String type = event.getStr("type", eventName);
        if ("error".equals(type) || "response.failed".equals(type)
                || "response.incomplete".equals(type) || "response.refusal.delta".equals(type)
                || "response.refusal.done".equals(type)) throw responseFailure();
        switch (protocol) {
            case "openai_chat":
                JSONObject choice = firstObject(event.getJSONArray("choices"));
                if (choice != null) {
                    String finish = choice.getStr("finish_reason");
                    validateStop(finish, "stop", false);
                    rejectRefusal(choice.getJSONObject("delta"));
                    rejectRefusal(choice.getJSONObject("message"));
                    appendText(result.text, firstText(choice, "delta.content", "message.content", "text"));
                    if (finish != null) result.completed = true;
                } else if ("response.output_text.delta".equals(type)) {
                    appendText(result.text, event.getStr("delta"));
                } else if ("response.completed".equals(type)) {
                    String text = responsesText(event.getJSONObject("response"));
                    if (!StringUtils.isEmpty(text)) {
                        result.text.setLength(0);
                        result.text.append(text);
                    }
                    result.completed = true;
                } else {
                    appendText(result.text, firstText(event, "output_text", "output[0].content"));
                }
                break;
            case "openai_responses":
                if ("response.output_text.delta".equals(type)) appendText(result.text, event.getStr("delta"));
                else if ("response.completed".equals(type)) {
                    String text = responsesText(event.getJSONObject("response"));
                    if (!StringUtils.isEmpty(text)) {
                        result.text.setLength(0);
                        result.text.append(text);
                    }
                    result.completed = true;
                }
                break;
            case "gemini":
                JSONObject candidate = geminiCandidate(event);
                if (candidate != null) {
                    String finish = candidate.getStr("finishReason");
                    validateStop(finish, "STOP", false);
                    appendText(result.text, firstText(candidate, "content.parts"));
                    if (finish != null) result.completed = true;
                }
                break;
            case "anthropic":
                if ("content_block_start".equals(type)) {
                    appendText(result.text, textValue(event.get("content_block")));
                } else if ("content_block_delta".equals(type)) {
                    JSONObject delta = event.getJSONObject("delta");
                    if (delta != null && "text_delta".equals(delta.getStr("type"))) {
                        appendText(result.text, delta.getStr("text"));
                    }
                } else if ("message_delta".equals(type)) {
                    JSONObject delta = event.getJSONObject("delta");
                    String stop = delta == null ? null : delta.getStr("stop_reason");
                    validateAnthropicStop(stop, false);
                    if (stop != null) result.stopReasonReceived = true;
                } else if ("message_stop".equals(type)) {
                    if (!result.stopReasonReceived) throw responseFailure();
                    result.completed = true;
                }
                break;
            default:
                throw new AiResponseException("不支持的 AI 接口协议");
        }
    }

    private String extractMessageContent(String responseBody, String protocol) {
        JSONObject response = parseResponse(responseBody);
        rejectError(response);
        switch (protocol) {
            case "openai_responses":
                return responsesText(response);
            case "gemini":
                JSONObject candidate = geminiCandidate(response);
                if (candidate == null) return "";
                validateStop(candidate.getStr("finishReason"), "STOP", true);
                return firstText(candidate, "content.parts");
            case "anthropic":
                validateAnthropicStop(response.getStr("stop_reason"), true);
                return textValue(response.get("content"));
            default:
                JSONObject choice = firstObject(response.getJSONArray("choices"));
                if (choice != null) {
                    validateStop(choice.getStr("finish_reason"), "stop", false);
                    rejectRefusal(choice.getJSONObject("message"));
                    return firstText(choice, "message.content", "text");
                }
                if (response.containsKey("status")) return responsesText(response);
                return firstText(response, "output_text", "output[0].content");
        }
    }

    private JSONObject parseResponse(String data) {
        try {
            if (!JSONUtil.isJsonObj(data)) throw new IllegalArgumentException();
            return JSONUtil.parseObj(data);
        } catch (Exception e) {
            throw new AiResponseException("AI 接口返回格式不兼容，请检查管理员配置的接口地址和模型");
        }
    }

    private JSONObject firstObject(JSONArray array) {
        return array == null || array.isEmpty() ? null : array.getJSONObject(0);
    }

    private void rejectError(JSONObject object) {
        if (object == null || !JSONUtil.isNull(object.get("error")) || "error".equals(object.getStr("type"))) {
            throw responseFailure();
        }
    }

    private void rejectRefusal(JSONObject object) {
        if (object != null && !StringUtils.isEmpty(object.getStr("refusal"))) throw responseFailure();
    }

    private IllegalStateException responseFailure() {
        // Provider errors can echo credentials or prompts. Never expose their body.
        return new AiResponseException("AI 服务返回错误、拒绝作答或内容未完整生成，请检查模型后重试");
    }

    private void validateStop(String actual, String expected, boolean required) {
        if ((required && actual == null) || (actual != null && !expected.equals(actual))) throw responseFailure();
    }

    private void validateAnthropicStop(String stop, boolean required) {
        if ("stop_sequence".equals(stop)) return;
        validateStop(stop, "end_turn", required);
    }

    private String responsesText(JSONObject response) {
        rejectError(response);
        validateStop(response.getStr("status"), "completed", true);
        JSONArray output = response.getJSONArray("output");
        StringBuilder content = new StringBuilder();
        if (output != null) {
            for (int i = 0; i < output.size(); i++) {
                JSONObject item = output.getJSONObject(i);
                if ("message".equals(item.getStr("type"))) {
                    validateStop(item.getStr("status"), "completed", false);
                    appendText(content, textValue(item.get("content")));
                }
            }
        }
        return content.length() > 0 ? content.toString() : textValue(response.get("output_text"));
    }

    private JSONObject geminiCandidate(JSONObject response) {
        rejectError(response);
        JSONObject feedback = response.getJSONObject("promptFeedback");
        String blockReason = feedback == null ? null : feedback.getStr("blockReason");
        if (blockReason != null && !"BLOCK_REASON_UNSPECIFIED".equals(blockReason)) throw responseFailure();
        return firstObject(response.getJSONArray("candidates"));
    }

    private String firstText(JSONObject value, String... paths) {
        for (String path : paths) {
            String text = textValue(value.getByPath(path));
            if (!StringUtils.isEmpty(text)) return text;
        }
        return "";
    }

    private String textValue(Object value) {
        if (value == null) return "";
        if (value instanceof CharSequence) return value.toString();
        if (value instanceof Map) {
            Map<?, ?> object = (Map<?, ?>) value;
            String type = object.get("type") == null ? "" : object.get("type").toString();
            if ("refusal".equals(type)) throw responseFailure();
            if (Boolean.TRUE.equals(object.get("thought"))) return "";
            if (!type.isEmpty() && !"text".equals(type) && !"output_text".equals(type)) return "";
            Object text = object.get("text");
            return textValue(text == null && type.isEmpty() ? object.get("content") : text);
        }
        if (value instanceof Iterable) {
            StringBuilder result = new StringBuilder();
            for (Object item : (Iterable<?>) value) result.append(textValue(item));
            return result.toString();
        }
        return "";
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
                || value.startsWith("gpt-5") || value.startsWith("gpt-6");
    }

    private String protocol(ProblemAiConfig config) {
        String value = config.getProtocol();
        if (value == null || value.trim().isEmpty()) return "openai_chat";
        value = value.trim().toLowerCase(Locale.ROOT);
        if (!Arrays.asList("openai_chat", "openai_responses", "gemini", "anthropic").contains(value)) {
            throw new IllegalStateException("不支持的 AI 接口协议");
        }
        return value;
    }

    private URI endpoint(ProblemAiConfig config, String protocol, boolean stream) {
        try {
            URI base = URI.create(config.getApiUrl().trim());
            if ((!"http".equalsIgnoreCase(base.getScheme()) && !"https".equalsIgnoreCase(base.getScheme()))
                    || base.getHost() == null || base.getUserInfo() != null || base.getFragment() != null) {
                throw new IllegalArgumentException();
            }
            String path = base.getRawPath() == null ? "" : base.getRawPath().replaceAll("/+$", "");
            List<String> query = new ArrayList<>();
            if (base.getRawQuery() != null) {
                for (String part : base.getRawQuery().split("&")) {
                    String key = URLDecoder.decode(part.split("=", 2)[0], "UTF-8").toLowerCase(Locale.ROOT);
                    if (Arrays.asList("key", "api_key", "apikey", "access_token", "token").contains(key)) {
                        throw new IllegalArgumentException();
                    }
                    if (!"gemini".equals(protocol) || !"alt".equals(key)) query.add(part);
                }
            }
            switch (protocol) {
                case "openai_chat":
                    if (path.isEmpty()) path = "/v1/chat/completions";
                    else if (path.endsWith("/v1")) path += "/chat/completions";
                    break;
                case "openai_responses":
                case "anthropic":
                    String suffix = "anthropic".equals(protocol) ? "/messages" : "/responses";
                    if (!path.endsWith(suffix)) path += path.endsWith("/v1") ? suffix : "/v1" + suffix;
                    break;
                case "gemini":
                    if (path.endsWith(":generateContent") || path.endsWith(":streamGenerateContent")) {
                        path = path.substring(0, path.lastIndexOf(':'));
                    } else {
                        String model = config.getModel();
                        if (model != null && model.startsWith("models/")) model = model.substring(7);
                        if (model == null || !model.matches("[A-Za-z0-9._-]+")) throw new IllegalArgumentException();
                        if (!path.matches(".*/v1(?:beta|alpha)?")) path += "/v1beta";
                        path += "/models/" + model;
                    }
                    path += stream ? ":streamGenerateContent" : ":generateContent";
                    if (stream) query.add("alt=sse");
                    break;
                default:
                    throw new IllegalArgumentException();
            }
            return URI.create(base.getScheme() + "://" + base.getRawAuthority() + path
                    + (query.isEmpty() ? "" : "?" + String.join("&", query)));
        } catch (Exception e) {
            throw new IllegalStateException("AI 接口地址或模型格式无效；请使用 HTTP(S) 地址，密钥仅填写在 API Key 中");
        }
    }
}
