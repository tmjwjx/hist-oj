package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemAiModelConfig;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.*;

class ProblemAiGatewayTest {
    private final ProblemAiGateway gateway = new ProblemAiGateway();

    @Test
    void ordinaryAndOtherProvidersReceiveStandardChatParameters() {
        for (String model : Arrays.asList("gpt-4o-mini", "deepseek-chat", "gemini-2.5-pro", "claude-sonnet-4", null)) {
            Map<String, Object> body = body(model, "xhigh");
            assertFalse(body.containsKey("reasoning_effort"));
            assertFalse(body.containsKey("max_completion_tokens"));
            assertEquals(16384, body.get("max_tokens"));
            assertEquals(0.1, body.get("temperature"));
            assertEquals(model, body.get("model"));
            assertEquals(false, body.get("stream"));
        }
    }

    @Test
    void recognizedReasoningModelsUseReasoningParametersIncludingGpt6() {
        for (String model : Arrays.asList("o1", "o3-mini", "o4-mini", "gpt-5.2", "gpt-6-astra", "GPT-6-SOL")) {
            Map<String, Object> body = body(model, "xhigh");
            assertEquals("xhigh", body.get("reasoning_effort"));
            assertFalse(body.containsKey("temperature"));
            assertFalse(body.containsKey("max_tokens"));
            assertEquals(16384, body.get("max_completion_tokens"));
        }
    }

    @Test
    void boundedRecheckEffortIsOnlySentToRecognizedReasoningModels() {
        assertEquals("medium", body("gpt-6-astra", "medium").get("reasoning_effort"));
        assertFalse(body("deepseek-chat", "medium").containsKey("reasoning_effort"));
    }

    @Test
    void concurrentProvidersKeepTheirEndpointCredentialAndModelTogether() throws Exception {
        List<HttpServer> servers = new ArrayList<>();
        List<ConcurrentLinkedQueue<JSONObject>> requests = Arrays.asList(
                new ConcurrentLinkedQueue<>(), new ConcurrentLinkedQueue<>());
        List<ProblemAiModelConfig> models = new ArrayList<>();
        ExecutorService callers = Executors.newFixedThreadPool(2);
        try {
            for (int index = 0; index < 2; index++) {
                final int provider = index;
                String path = index == 0 ? "/v1/chat/completions" : "/custom/completions";
                HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
                servers.add(server);
                server.createContext(path, exchange -> {
                    try {
                        ByteArrayOutputStream bytes = new ByteArrayOutputStream();
                        try (InputStream input = exchange.getRequestBody()) {
                            byte[] buffer = new byte[4096];
                            int count;
                            while ((count = input.read(buffer)) != -1) bytes.write(buffer, 0, count);
                        }
                        requests.get(provider).add(new JSONObject()
                                .set("method", exchange.getRequestMethod())
                                .set("path", exchange.getRequestURI().getPath())
                                .set("authorization", exchange.getRequestHeaders().getFirst("Authorization"))
                                .set("body", JSONUtil.parseObj(new String(bytes.toByteArray(), StandardCharsets.UTF_8))));
                        byte[] response = ("{\"choices\":[{\"message\":{\"content\":\"provider-"
                                + provider + "\"}}]}").getBytes(StandardCharsets.UTF_8);
                        exchange.getResponseHeaders().set("Content-Type", "application/json");
                        exchange.sendResponseHeaders(200, response.length);
                        exchange.getResponseBody().write(response);
                    } finally {
                        exchange.close();
                    }
                });
                server.start();
                models.add(new ProblemAiModelConfig()
                        .setApiUrl("http://127.0.0.1:" + server.getAddress().getPort()
                                + (index == 0 ? "/v1" : path))
                        .setApiKey("test-only-key-" + index)
                        .setModel(index == 0 ? "gpt-6-astra" : "deepseek-chat"));
            }
            ProblemAiConfig shared = new ProblemAiConfig().setModels(models).setTimeoutSeconds(30);
            ProblemAiConfigManager configs = new ProblemAiConfigManager();
            List<Future<String>> completions = new ArrayList<>();
            for (ProblemAiModelConfig model : models) {
                completions.add(callers.submit(() -> gateway.complete(configs.forModel(shared, model),
                        Collections.singletonList(Collections.singletonMap("content", "question")))));
            }
            assertEquals("provider-0", completions.get(0).get(10, TimeUnit.SECONDS));
            assertEquals("provider-1", completions.get(1).get(10, TimeUnit.SECONDS));
            assertEquals("provider-0", gateway.complete(configs.forModel(shared, models.get(0)),
                    Collections.singletonList(Collections.singletonMap("content", "question again"))));
            assertEquals(2, requests.get(0).size());
            assertEquals(1, requests.get(1).size());
            for (int index = 0; index < 2; index++) {
                for (JSONObject request : requests.get(index)) {
                    assertEquals("POST", request.getStr("method"));
                    assertEquals(index == 0 ? "/v1/chat/completions" : "/custom/completions", request.getStr("path"));
                    assertEquals("Bearer test-only-key-" + index, request.getStr("authorization"));
                    JSONObject body = request.getJSONObject("body");
                    assertEquals(models.get(index).getModel(), body.getStr("model"));
                    assertEquals(index == 0, body.containsKey("reasoning_effort"));
                    assertTrue(body.getBool("stream"));
                    assertFalse(body.toString().contains("test-only-key-"));
                    assertFalse(body.containsKey("models"));
                }
            }
            assertNull(shared.getApiKey());
            assertNull(shared.getApiUrl());
            assertNull(shared.getModel());
        } finally {
            callers.shutdownNow();
            for (HttpServer server : servers) server.stop(0);
        }
    }

    private Map<String, Object> body(String model, String effort) {
        return ReflectionTestUtils.invokeMethod(gateway, "requestBody", new ProblemAiConfig().setModel(model),
                Collections.singletonList(Collections.singletonMap("content", "question")), false, effort);
    }
}
