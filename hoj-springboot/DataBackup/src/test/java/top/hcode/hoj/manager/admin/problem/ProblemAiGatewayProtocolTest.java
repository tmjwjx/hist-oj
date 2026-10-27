package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.net.InetSocketAddress;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.jupiter.api.Assertions.*;

/** Native protocol contracts use local HTTP only, never bill a model provider. */
class ProblemAiGatewayProtocolTest {
    private final ProblemAiGateway gateway = new ProblemAiGateway();
    private static final List<String> PROTOCOLS = Arrays.asList(
            "openai_chat", "openai_responses", "gemini", "anthropic");

    @Test
    void nativeJsonRequestsUseTheirOwnHeadersRouteBodyAndText() throws Exception {
        for (String protocol : PROTOCOLS) {
            try (Server server = new Server(json(protocol), false)) {
                ProblemAiConfig config = config(server.base(), protocol);
                assertEquals("answer", gateway.completeOnce(config, messages(), 30), protocol);
                JSONObject request = server.requests.remove();
                assertEquals("POST", request.getStr("method"));
                assertEquals(expectedPath(protocol, false), request.getStr("path"));
                assertAuthentication(request, protocol);
                JSONObject body = request.getJSONObject("body");
                assertFalse(body.toString().contains("test-key"));
                if ("gemini".equals(protocol)) {
                    assertNull(body.get("stream"));
                    assertNull(body.get("model"));
                    assertEquals("rules\n\nextra rules", body.getByPath("systemInstruction.parts[0].text"));
                    assertEquals("user", body.getByPath("contents[0].role"));
                    assertEquals("model", body.getByPath("contents[1].role"));
                    assertEquals("previous", body.getByPath("contents[1].parts[0].text"));
                    assertEquals(16384, body.getByPath("generationConfig.maxOutputTokens"));
                } else {
                    assertEquals("some-model", body.getStr("model"));
                    assertFalse(body.getBool("stream"));
                    if ("openai_responses".equals(protocol)) {
                        assertFalse(body.getBool("store"));
                        assertEquals(16384, body.getInt("max_output_tokens"));
                        assertEquals("rules\n\nextra rules", body.getStr("instructions"));
                        assertEquals("user", body.getByPath("input[0].role"));
                        assertEquals("assistant", body.getByPath("input[1].role"));
                        assertFalse(body.containsKey("messages"));
                    } else if ("anthropic".equals(protocol)) {
                        assertEquals("rules\n\nextra rules", body.getStr("system"));
                        assertEquals("user", body.getByPath("messages[0].role"));
                        assertEquals(16384, body.getInt("max_tokens"));
                        assertFalse(body.containsKey("reasoning_effort"));
                        assertFalse(body.containsKey("temperature"));
                    } else {
                        assertEquals("system", body.getByPath("messages[0].role"));
                        assertEquals(16384, body.getInt("max_tokens"));
                    }
                }
            }
        }
    }

    @Test
    void nativeStreamsIgnoreReasoningAndDoNotDuplicateTerminalSnapshots() throws Exception {
        for (String protocol : PROTOCOLS) {
            try (Server server = new Server(stream(protocol), true)) {
                assertEquals("answer", gateway.complete(config(server.base(), protocol), messages()), protocol);
                assertEquals(1, server.requests.size());
                JSONObject request = server.requests.remove();
                assertEquals(expectedPath(protocol, true), request.getStr("path"));
                assertAuthentication(request, protocol);
                if ("gemini".equals(protocol)) assertEquals("alt=sse", request.getStr("query"));
                else assertTrue(request.getJSONObject("body").getBool("stream"));
            }
        }
    }

    @Test
    void errorsWithinSuccessfulHttpResponsesNeverBecomeGeneratedCode() throws Exception {
        for (String protocol : PROTOCOLS) {
            for (boolean sse : Arrays.asList(false, true)) {
                String body = "{\"error\":{\"message\":\"test-key secret prompt\"}}";
                if (sse) body = "event: error\ndata: " + body + "\n\n";
                try (Server server = new Server(body, sse)) {
                    IllegalStateException error = assertThrows(IllegalStateException.class,
                            () -> gateway.complete(config(server.base(), protocol), messages()));
                    assertFalse(error.toString().contains("test-key"));
                    assertFalse(error.toString().contains("secret prompt"));
                    assertEquals(1, server.requests.size());
                }
            }
        }
    }

    @Test
    void partialStreamsRequireProviderCompletionMarkers() throws Exception {
        Map<String, String> partial = new LinkedHashMap<>();
        partial.put("openai_chat", "{\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}");
        partial.put("openai_responses", "{\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}");
        partial.put("gemini", "{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"partial\"}]}}]}");
        partial.put("anthropic", "{\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"partial\"}}");
        for (Map.Entry<String, String> entry : partial.entrySet()) {
            try (Server server = new Server(event(entry.getValue()), true)) {
                IllegalStateException error = assertThrows(IllegalStateException.class,
                        () -> gateway.complete(config(server.base(), entry.getKey()), messages()));
                assertTrue(error.getMessage().contains("中断"));
                assertEquals(1, server.requests.size());
            }
        }
    }

    @Test
    void malformedResponseStructuresNeverLeakProviderContent() throws Exception {
        Map<String, String> malformed = new LinkedHashMap<>();
        malformed.put("openai_chat", "{\"choices\":\"sensitive-provider-secret\"}");
        malformed.put("openai_responses", "{\"status\":\"completed\",\"output\":[\"sensitive-provider-secret\"]}");
        malformed.put("gemini", "{\"candidates\":\"sensitive-provider-secret\"}");
        malformed.put("anthropic", "{\"stop_reason\":\"end_turn\",\"content\":123}");
        for (Map.Entry<String, String> entry : malformed.entrySet()) {
            try (Server server = new Server(entry.getValue(), false)) {
                IllegalStateException error = assertThrows(IllegalStateException.class,
                        () -> gateway.completeOnce(config(server.base(), entry.getKey()), messages(), 30));
                assertFalse(error.toString().contains("sensitive-provider-secret"));
                assertNull(error.getCause());
            }
        }
    }

    @Test
    void tokenLimitsSafetyRefusalsAndIncompleteResponsesAreFailures() throws Exception {
        Map<String, List<String>> errors = new LinkedHashMap<>();
        errors.put("openai_chat", Arrays.asList(
                "{\"choices\":[{\"finish_reason\":\"length\",\"message\":{\"content\":\"partial\"}}]}",
                "{\"choices\":[{\"finish_reason\":\"content_filter\",\"message\":{\"content\":\"partial\"}}]}",
                "{\"choices\":[{\"message\":{\"refusal\":\"no\",\"content\":\"partial\"}}]}"));
        errors.put("openai_responses", Arrays.asList(
                "{\"status\":\"incomplete\",\"output_text\":\"partial\"}",
                "{\"status\":\"failed\",\"output_text\":\"partial\"}",
                "{\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"content\":[{\"type\":\"refusal\",\"refusal\":\"no\"}]}]}"));
        errors.put("gemini", Arrays.asList(
                "{\"candidates\":[{\"finishReason\":\"MAX_TOKENS\",\"content\":{\"parts\":[{\"text\":\"partial\"}]}}]}",
                "{\"promptFeedback\":{\"blockReason\":\"SAFETY\"}}"));
        errors.put("anthropic", Arrays.asList(
                "{\"stop_reason\":\"max_tokens\",\"content\":[{\"type\":\"text\",\"text\":\"partial\"}]}",
                "{\"stop_reason\":\"tool_use\",\"content\":[{\"type\":\"text\",\"text\":\"partial\"}]}"));
        for (Map.Entry<String, List<String>> entry : errors.entrySet()) {
            for (String body : entry.getValue()) {
                try (Server server = new Server(body, false)) {
                    assertThrows(IllegalStateException.class,
                            () -> gateway.completeOnce(config(server.base(), entry.getKey()), messages(), 30));
                    assertEquals(1, server.requests.size());
                }
            }
        }
    }

    @Test
    void streamLimitErrorsCannotBeHiddenByDoneMarkers() throws Exception {
        Map<String, String> errors = new LinkedHashMap<>();
        errors.put("openai_chat", event("{\"choices\":[{\"finish_reason\":\"length\",\"delta\":{\"content\":\"partial\"}}]}") + event("[DONE]"));
        errors.put("openai_responses", event("{\"type\":\"response.incomplete\",\"response\":{\"status\":\"incomplete\"}}"));
        errors.put("gemini", event("{\"candidates\":[{\"finishReason\":\"MAX_TOKENS\",\"content\":{\"parts\":[{\"text\":\"partial\"}]}}]}"));
        errors.put("anthropic", event("{\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"max_tokens\"}}") + event("{\"type\":\"message_stop\"}"));
        for (Map.Entry<String, String> entry : errors.entrySet()) {
            try (Server server = new Server(entry.getValue(), true)) {
                assertThrows(IllegalStateException.class,
                        () -> gateway.complete(config(server.base(), entry.getKey()), messages()));
                assertEquals(1, server.requests.size());
            }
        }
    }

    @Test
    void streamingFallbackPreservesProtocolAndChangesGeminiAction() throws Exception {
        for (String protocol : PROTOCOLS) {
            try (Server server = new Server(json(protocol), false)) {
                server.rejectFirstStream = true;
                assertEquals("answer", gateway.complete(config(server.base(), protocol), messages()));
                assertEquals(2, server.requests.size());
                JSONObject first = server.requests.remove();
                JSONObject second = server.requests.remove();
                assertEquals(expectedPath(protocol, true), first.getStr("path"));
                assertEquals(expectedPath(protocol, false), second.getStr("path"));
                assertAuthentication(first, protocol);
                assertAuthentication(second, protocol);
                if ("gemini".equals(protocol)) assertNull(second.getStr("query"));
                else assertFalse(second.getJSONObject("body").getBool("stream"));
            }
        }
    }

    @Test
    void endpointsAcceptBaseVersionAndFullPathsWithoutLeakingCredentials() {
        assertEndpoint("https://example.com", "openai_chat", false, "https://example.com/v1/chat/completions");
        assertEndpoint("https://example.com/v1/", "openai_chat", true, "https://example.com/v1/chat/completions");
        assertEndpoint("https://example.com/custom/completions", "openai_chat", true, "https://example.com/custom/completions");
        assertEndpoint("https://example.com/v1/responses", "openai_responses", true, "https://example.com/v1/responses");
        assertEndpoint("https://example.com/proxy", "openai_responses", false, "https://example.com/proxy/v1/responses");
        assertEndpoint("https://example.com/v1", "anthropic", false, "https://example.com/v1/messages");
        assertEndpoint("https://example.com/proxy/v1/messages?custom=yes", "anthropic", true,
                "https://example.com/proxy/v1/messages?custom=yes");
        assertEndpoint("https://example.com/v1beta", "gemini", true,
                "https://example.com/v1beta/models/some-model:streamGenerateContent?alt=sse");
        assertEndpoint("https://example.com/v1/models/some-model:streamGenerateContent?alt=sse&custom=yes", "gemini", false,
                "https://example.com/v1/models/some-model:generateContent?custom=yes");
        assertEndpoint("https://example.com/v1beta/models/some-model:generateContent", "gemini", true,
                "https://example.com/v1beta/models/some-model:streamGenerateContent?alt=sse");
        for (String url : Arrays.asList("https://user:test-key@example.com", "https://example.com?key=test-key",
                "https://example.com?api_key=test-key", "https://example.com?%61piKey=test-key", "file:///tmp/local")) {
            IllegalStateException error = assertThrows(IllegalStateException.class,
                    () -> gateway.completeOnce(config(url, "gemini"), messages(), 30));
            assertFalse(error.toString().contains("test-key"));
        }
        assertThrows(IllegalStateException.class,
                () -> gateway.completeOnce(config("http://127.0.0.1:1", "unknown"), messages(), 30));
        IllegalStateException invalidKey = assertThrows(IllegalStateException.class,
                () -> gateway.completeOnce(config("http://127.0.0.1:1", "anthropic")
                        .setApiKey("test-key\r\nInjected: header"), messages(), 30));
        assertFalse(invalidKey.toString().contains("test-key"));
        assertNull(invalidKey.getCause());
    }

    @Test
    void redirectsNeverForwardProviderKeys() throws Exception {
        try (Server destination = new Server(json("openai_chat"), false);
             Server redirect = new Server("", false)) {
            redirect.redirectTo = destination.base() + "/v1/chat/completions";
            assertThrows(IllegalStateException.class,
                    () -> gateway.completeOnce(config(redirect.base(), "openai_chat"), messages(), 30));
            assertEquals(1, redirect.requests.size());
            assertEquals(0, destination.requests.size());
        }
    }

    @Test
    void simultaneousProtocolsKeepCredentialAndEndpointTogether() throws Exception {
        ExecutorService pool = Executors.newFixedThreadPool(4);
        List<Server> servers = new ArrayList<>();
        try {
            List<Future<String>> requests = new ArrayList<>();
            for (String protocol : PROTOCOLS) {
                Server server = new Server(json(protocol), false);
                servers.add(server);
                requests.add(pool.submit(() -> gateway.completeOnce(config(server.base(), protocol), messages(), 30)));
            }
            for (int i = 0; i < requests.size(); i++) {
                assertEquals("answer", requests.get(i).get(10, TimeUnit.SECONDS));
                assertAuthentication(servers.get(i).requests.remove(), PROTOCOLS.get(i));
            }
        } finally {
            pool.shutdownNow();
            for (Server server : servers) server.close();
        }
    }

    @Test
    void responsesReasoningConfigurationKeepsEffortOutOfOtherProtocols() {
        ProblemAiConfig config = config("https://example.com", "openai_responses").setModel("gpt-6-astra");
        Map<String, Object> body = ReflectionTestUtils.invokeMethod(gateway, "requestBody", config, messages(), false, "medium");
        assertEquals("medium", ((Map<?, ?>) body.get("reasoning")).get("effort"));
        assertFalse(body.containsKey("temperature"));
        assertFalse(body.containsKey("reasoning_effort"));
        config.setProtocol("anthropic");
        body = ReflectionTestUtils.invokeMethod(gateway, "requestBody", config, messages(), false, "medium");
        assertFalse(body.containsKey("reasoning"));
        assertFalse(body.containsKey("reasoning_effort"));
    }

    private void assertEndpoint(String url, String protocol, boolean stream, String expected) {
        URI actual = ReflectionTestUtils.invokeMethod(gateway, "endpoint", config(url, protocol), protocol, stream);
        assertEquals(expected, actual.toString());
    }

    private static ProblemAiConfig config(String url, String protocol) {
        return new ProblemAiConfig().setProtocol(protocol).setApiUrl(url).setApiKey("test-key-" + protocol)
                .setModel("some-model").setTimeoutSeconds(30);
    }

    private static List<Map<String, String>> messages() {
        List<Map<String, String>> result = new ArrayList<>();
        String[] roles = {"system", "developer", "user", "assistant", "user"};
        String[] contents = {"rules", "extra rules", "question", "previous", "continue"};
        for (int i = 0; i < roles.length; i++) {
            Map<String, String> message = new LinkedHashMap<>();
            message.put("role", roles[i]);
            message.put("content", contents[i]);
            result.add(message);
        }
        return result;
    }

    private static void assertAuthentication(JSONObject request, String protocol) {
        assertEquals("gemini".equals(protocol) ? "test-key-gemini" : null, request.getStr("google-key"));
        assertEquals("anthropic".equals(protocol) ? "test-key-anthropic" : null, request.getStr("anthropic-key"));
        assertEquals("anthropic".equals(protocol) ? "2023-06-01" : null, request.getStr("anthropic-version"));
        assertEquals(protocol.startsWith("openai_") ? "Bearer test-key-" + protocol : null, request.getStr("authorization"));
        assertFalse(request.getStr("path").contains("test-key"));
        assertFalse(request.getStr("query", "").contains("test-key"));
    }

    private static String expectedPath(String protocol, boolean stream) {
        switch (protocol) {
            case "openai_chat": return "/v1/chat/completions";
            case "openai_responses": return "/v1/responses";
            case "anthropic": return "/v1/messages";
            default: return "/v1beta/models/some-model:" + (stream ? "streamGenerateContent" : "generateContent");
        }
    }

    private static String json(String protocol) {
        switch (protocol) {
            case "openai_chat": return "{\"choices\":[{\"finish_reason\":\"stop\",\"message\":{\"reasoning_content\":\"secret thought\",\"content\":[{\"type\":\"reasoning\",\"text\":\"secret thought\"},{\"type\":\"text\",\"text\":\"answer\"}]}}]}";
            case "openai_responses": return "{\"status\":\"completed\",\"error\":null,\"output\":[{\"type\":\"reasoning\",\"content\":[{\"type\":\"text\",\"text\":\"secret thought\"}]},{\"type\":\"message\",\"status\":\"completed\",\"content\":[{\"type\":\"output_text\",\"text\":\"answer\"}]}]}";
            case "gemini": return "{\"candidates\":[{\"finishReason\":\"STOP\",\"content\":{\"parts\":[{\"thought\":true,\"text\":\"secret thought\"},{\"text\":\"answer\"}]}}]}";
            default: return "{\"type\":\"message\",\"stop_reason\":\"end_turn\",\"content\":[{\"type\":\"thinking\",\"thinking\":\"secret thought\"},{\"type\":\"text\",\"text\":\"answer\"}]}";
        }
    }

    private static String event(String json) { return "data: " + json + "\n\n"; }

    private static String stream(String protocol) {
        switch (protocol) {
            case "openai_chat": return ": heartbeat\n\n"
                    + event("{\"choices\":[{\"delta\":{\"reasoning_content\":\"secret thought\"}}]}")
                    + event("{\"choices\":[{\"delta\":{\"content\":\"answer\"},\"finish_reason\":null}]}")
                    + event("{\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}") + event("[DONE]");
            case "openai_responses": return event("{\"type\":\"response.reasoning_summary_text.delta\",\"delta\":\"secret thought\"}")
                    + event("{\"type\":\"response.output_text.delta\",\"delta\":\"answer\"}")
                    + event("{\"type\":\"response.output_text.done\",\"text\":\"answer\"}")
                    + event("{\"type\":\"response.completed\",\"response\":" + json(protocol) + "}");
            case "gemini": return event("{\"candidates\":[{\"content\":{\"parts\":[{\"thought\":true,\"text\":\"secret thought\"},{\"text\":\"answer\"}]}}]}")
                    + event("{\"candidates\":[{\"finishReason\":\"STOP\",\"content\":{\"parts\":[]}}]}");
            default: return event("{\"type\":\"message_start\",\"message\":{\"stop_reason\":null}}")
                    + event("{\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\",\"text\":\"\"}}")
                    + event("{\"type\":\"content_block_delta\",\"delta\":{\"type\":\"thinking_delta\",\"thinking\":\"secret thought\"}}")
                    + event("{\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"answer\"}}")
                    + event("{\"type\":\"message_delta\",\"delta\":{\"stop_reason\":null}}")
                    + event("{\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}")
                    + event("{\"type\":\"message_stop\"}");
        }
    }

    private static final class Server implements AutoCloseable {
        final HttpServer server;
        final Queue<JSONObject> requests = new ConcurrentLinkedQueue<>();
        volatile boolean rejectFirstStream;
        volatile String redirectTo;
        final AtomicInteger count = new AtomicInteger();

        Server(String response, boolean sse) throws Exception {
            server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
            server.createContext("/", exchange -> {
                try {
                    ByteArrayOutputStream bytes = new ByteArrayOutputStream();
                    try (InputStream input = exchange.getRequestBody()) {
                        byte[] buffer = new byte[4096];
                        int length;
                        while ((length = input.read(buffer)) != -1) bytes.write(buffer, 0, length);
                    }
                    requests.add(new JSONObject().set("method", exchange.getRequestMethod())
                            .set("path", exchange.getRequestURI().getPath()).set("query", exchange.getRequestURI().getRawQuery())
                            .set("authorization", exchange.getRequestHeaders().getFirst("Authorization"))
                            .set("google-key", exchange.getRequestHeaders().getFirst("x-goog-api-key"))
                            .set("anthropic-key", exchange.getRequestHeaders().getFirst("x-api-key"))
                            .set("anthropic-version", exchange.getRequestHeaders().getFirst("anthropic-version"))
                            .set("body", JSONUtil.parseObj(new String(bytes.toByteArray(), StandardCharsets.UTF_8))));
                    boolean reject = rejectFirstStream && count.getAndIncrement() == 0;
                    byte[] output = (reject ? "{\"error\":\"stream unsupported\"}" : response).getBytes(StandardCharsets.UTF_8);
                    exchange.getResponseHeaders().set("Content-Type", sse ? "text/event-stream" : "application/json");
                    if (redirectTo != null) exchange.getResponseHeaders().set("Location", redirectTo);
                    exchange.sendResponseHeaders(redirectTo != null ? 307 : reject ? 400 : 200, output.length);
                    exchange.getResponseBody().write(output);
                } finally { exchange.close(); }
            });
            server.start();
        }

        String base() { return "http://127.0.0.1:" + server.getAddress().getPort(); }
        public void close() { server.stop(0); }
    }
}
