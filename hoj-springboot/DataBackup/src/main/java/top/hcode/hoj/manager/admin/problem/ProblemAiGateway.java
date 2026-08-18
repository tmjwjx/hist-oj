package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONUtil;
import org.springframework.http.*;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.web.client.RestTemplate;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;

import java.nio.charset.StandardCharsets;
import java.util.*;

@Component
public class ProblemAiGateway {
    private static final String REASONING_EFFORT = "xhigh";

    public String complete(ProblemAiConfig config, List<Map<String, String>> messages) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("model", config.getModel());
        body.put("messages", messages);
        body.put("temperature", 0.1);
        body.put("reasoning_effort", REASONING_EFFORT);

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.setAccept(Collections.singletonList(MediaType.APPLICATION_JSON));
        headers.setBearerAuth(config.getApiKey());
        headers.set(HttpHeaders.USER_AGENT, "Mozilla/5.0 (compatible; HistOJ-AIValidator/1.0)");

        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        int timeout = (config.getTimeoutSeconds() == null ? 120 : config.getTimeoutSeconds()) * 1000;
        factory.setConnectTimeout(timeout);
        factory.setReadTimeout(timeout);
        ResponseEntity<byte[]> response = new RestTemplate(factory).postForEntity(
                chatCompletionsUrl(config.getApiUrl()), new HttpEntity<>(body, headers), byte[].class);
        if (!response.getStatusCode().is2xxSuccessful()) {
            throw new IllegalStateException("AI 接口返回 HTTP " + response.getStatusCodeValue());
        }
        String responseBody = response.getBody() == null
                ? "" : new String(response.getBody(), StandardCharsets.UTF_8);
        String content = JSONUtil.parseObj(responseBody)
                .getByPath("choices[0].message.content", String.class);
        if (StringUtils.isEmpty(content)) throw new IllegalStateException("AI 接口未返回有效内容");
        return content;
    }

    private String chatCompletionsUrl(String apiUrl) {
        String value = apiUrl == null ? "" : apiUrl.trim().replaceAll("/+$", "");
        return value.endsWith("/v1") ? value + "/chat/completions" : value;
    }
}
