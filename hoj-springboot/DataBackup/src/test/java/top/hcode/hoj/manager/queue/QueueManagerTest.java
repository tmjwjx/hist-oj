package top.hcode.hoj.manager.queue;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.springframework.data.redis.RedisConnectionFailureException;
import org.springframework.data.redis.connection.RedisStandaloneConfiguration;
import org.springframework.data.redis.connection.jedis.JedisConnectionFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.script.RedisScript;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import top.hcode.hoj.controller.oj.QueueController;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import static org.junit.jupiter.api.Assertions.*;
import static org.junit.jupiter.api.Assumptions.assumeTrue;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

class QueueManagerTest {

    private static String ticket() {
        return UUID.randomUUID().toString();
    }

    private static long number(Map<String, Object> state, String key) {
        return ((Number) state.get(key)).longValue();
    }

    private static void counts(Map<String, Object> state, long waiting, long reserved, long loading) {
        assertEquals(waiting, number(state, "waiting"));
        assertEquals(reserved, number(state, "reserved"));
        assertEquals(loading, number(state, "loading"));
        assertEquals(reserved + loading, number(state, "active"));
        assertEquals(waiting + reserved + loading, number(state, "total"));
    }

    @Test
    void rejectsInvalidCapacityAndIdentityBeforeCallingRedis() {
        StringRedisTemplate unused = new StringRedisTemplate();
        assertThrows(IllegalArgumentException.class, () -> new QueueManager(unused, true, 0));
        assertThrows(IllegalArgumentException.class, () -> new QueueManager(unused, true, -1));
        QueueManager manager = new QueueManager(unused, true, 1);
        for (String id : new String[]{null, "", "1-1-1-1-1", "not-a-ticket", ticket() + " "}) {
            assertThrows(IllegalArgumentException.class, () -> manager.status(id));
            assertThrows(IllegalArgumentException.class, () -> manager.claim(id));
            assertThrows(IllegalArgumentException.class, () -> manager.release(id));
            assertThrows(IllegalArgumentException.class, () -> manager.complete(id));
            assertThrows(IllegalArgumentException.class, () -> manager.reuse(id));
            assertThrows(IllegalArgumentException.class, () -> manager.leave(id, true));
            assertThrows(IllegalArgumentException.class, () -> manager.heartbeat(id));
        }
    }

    @Test
    void disabledQueueKeepsIdentityWithoutRedis() {
        QueueManager manager = new QueueManager(new StringRedisTemplate(), false, 5);
        String id = ticket();
        Map<String, Object> result = manager.status(id);
        assertEquals(id, result.get("myId"));
        assertEquals("disabled", result.get("state"));
        assertEquals(1000, number(result, "retryAfterMs"));
        assertEquals(true, result.get("admitted"));
        assertEquals(true, manager.claim(id).get("claimed"));
        counts(result, 0, 0, 0);
    }

    @Test
    void httpErrorsNeverGrantAdmissionOrCacheState() throws Exception {
        StringRedisTemplate broken = new StringRedisTemplate() {
            @Override
            public <T> T execute(RedisScript<T> script, List<String> keys, Object... args) {
                throw new RedisConnectionFailureException("test Redis unavailable");
            }
        };
        MockMvc http = MockMvcBuilders.standaloneSetup(new QueueController(new QueueManager(broken, true, 1))).build();
        http.perform(get("/api/queue/status").param("qid", ticket()))
                .andExpect(status().isServiceUnavailable()).andExpect(header().string("Cache-Control", "no-store"))
                .andExpect(header().string("Retry-After", "1")).andExpect(jsonPath("$.status").value(503));
        http.perform(get("/api/queue/claim").header("X-Queue-Ticket", ticket()))
                .andExpect(status().isServiceUnavailable());
        http.perform(get("/api/queue/claim")).andExpect(status().isUnauthorized());
        http.perform(get("/api/queue/claim").header("X-Queue-Warm", ticket()))
                .andExpect(status().isServiceUnavailable());
        http.perform(get("/api/queue/claim").header("X-Queue-Warm", ticket()).header("X-Queue-Force", "1"))
                .andExpect(status().isUnauthorized());
        http.perform(get("/api/queue/status").param("qid", "invalid"))
                .andExpect(status().isBadRequest()).andExpect(header().string("Cache-Control", "no-store"));
        http.perform(post("/api/queue/leave").param("qid", ticket()).param("immediate", "invalid"))
                .andExpect(status().isBadRequest());
        http.perform(post("/api/queue/release").param("qid", ticket()).param("completed", "invalid"))
                .andExpect(status().isBadRequest());
        http.perform(get("/api/queue/status-batch").param("qids", ticket())).andExpect(status().isNotFound());
        http.perform(get("/api/queue/release").param("qid", ticket())).andExpect(status().isMethodNotAllowed());
    }

    /** Uses the actual production Lua against an isolated Redis, with no reimplementation of its transitions.
     * Run with -Dqueue.test.redis.port=<isolated Redis port>; host defaults to 127.0.0.1.
     */
    @Nested
    class AtomicRedisTransitions {
        private JedisConnectionFactory factory;
        private StringRedisTemplate redis;
        private QueueManager manager;

        @BeforeEach
        void connect() {
            String port = System.getProperty("queue.test.redis.port");
            assumeTrue(port != null, "An isolated Redis port is required for Lua integration tests");
            factory = new JedisConnectionFactory(new RedisStandaloneConfiguration(
                    System.getProperty("queue.test.redis.host", "127.0.0.1"), Integer.parseInt(port)));
            factory.afterPropertiesSet();
            redis = new StringRedisTemplate(factory);
            redis.delete(QueueManager.KEYS);
            manager = new QueueManager(redis, true, 2);
        }

        @AfterEach
        void disconnect() {
            if (redis != null) {
                redis.delete(QueueManager.KEYS);
            }
            if (factory != null) {
                factory.destroy();
            }
        }

        private <T> List<T> concurrent(List<Callable<T>> operations) throws Exception {
            ExecutorService pool = Executors.newFixedThreadPool(12);
            try {
                List<T> results = new ArrayList<>();
                for (Future<T> result : pool.invokeAll(operations)) {
                    results.add(result.get());
                }
                return results;
            } finally {
                pool.shutdownNow();
            }
        }

        @Test
        void countsStayExactAcrossReservationClaimAndRelease() {
            String a = ticket(), b = ticket(), c = ticket();
            counts(manager.status(a), 0, 1, 0);
            counts(manager.status(b), 0, 2, 0);
            Map<String, Object> waiting = manager.status(c);
            counts(waiting, 1, 2, 0);
            assertEquals(1, number(waiting, "position"));
            counts(manager.claim(a), 1, 1, 1);
            Map<String, Object> released = manager.release(a);
            counts(released, 0, 2, 0);
            assertEquals("closed", released.get("state"));
            assertEquals("reserved", manager.status(c).get("state"));
            assertTrue(number(released, "revision") > number(waiting, "revision"));
        }

        @Test
        void completedLoadCreatesANewTokenAndReuseNeverOccupiesOrRenewsASlot() {
            manager = new QueueManager(redis, true, 1);
            String current = ticket(), next = ticket();
            manager.status(current);
            manager.claim(current);
            manager.status(next);
            Map<String, Object> completed = manager.complete(current);
            assertEquals(true, completed.get("completed"));
            assertEquals("closed", completed.get("state"));
            counts(completed, 0, 1, 0);
            String token = (String) completed.get("completionToken");
            assertNotNull(token);
            assertNotEquals(current, token);
            assertFalse(manager.reuse(current));
            Double expires = redis.opsForZSet().score(QueueManager.KEYS.get(8), token);
            assertTrue(expires > System.currentTimeMillis() + 86300000);
            assertTrue(expires <= System.currentTimeMillis() + 86400000);
            String revision = redis.opsForValue().get(QueueManager.KEYS.get(7));
            for (int i = 0; i < 3; i++) assertTrue(manager.reuse(token));
            assertEquals(expires, redis.opsForZSet().score(QueueManager.KEYS.get(8), token));
            assertEquals(revision, redis.opsForValue().get(QueueManager.KEYS.get(7)));
            counts(manager.status(next), 0, 1, 0);
            manager.release(current); // A delayed pagehide must not revoke successful completion.
            assertTrue(manager.reuse(token));
        }

        @Test
        void completionCannotGrantTokensToUnknownWaitingReservedReleasedOrExpiredPages() {
            manager = new QueueManager(redis, true, 1);
            String reserved = ticket(), waiting = ticket();
            manager.status(reserved);
            manager.status(waiting);
            for (String id : new String[]{ticket(), waiting, reserved}) {
                Map<String, Object> result = manager.complete(id);
                assertEquals(false, result.get("completed"));
                assertFalse(result.containsKey("completionToken"));
            }
            String released = ticket();
            manager.status(released);
            manager.claim(released);
            manager.release(released);
            assertEquals(false, manager.complete(released).get("completed"));
            String expired = ticket();
            manager.status(expired);
            manager.claim(expired);
            redis.opsForZSet().add(QueueManager.KEYS.get(3), expired, 0);
            assertEquals(false, manager.complete(expired).get("completed"));
            assertEquals(0L, redis.opsForZSet().zCard(QueueManager.KEYS.get(8)));
        }

        @Test
        void concurrentCompletionAndLateReplaysIssueExactlyOneToken() throws Exception {
            String id = ticket();
            manager.status(id);
            manager.claim(id);
            List<Callable<Map<String, Object>>> completions = new ArrayList<>();
            for (int i = 0; i < 32; i++) completions.add(() -> manager.complete(id));
            List<Map<String, Object>> results = concurrent(completions);
            assertEquals(1L, results.stream().filter(r -> Boolean.TRUE.equals(r.get("completed"))).count());
            Map<String, Object> accepted = results.stream().filter(r -> r.containsKey("completionToken"))
                    .findFirst().get();
            assertTrue(manager.reuse((String) accepted.get("completionToken")));
            assertEquals(1L, redis.opsForZSet().zCard(QueueManager.KEYS.get(8)));
            redis.opsForZSet().add(QueueManager.KEYS.get(5), id, 0);
            manager.cleanup();
            assertEquals(false, manager.complete(id).get("completed"));
            assertEquals(1L, redis.opsForZSet().zCard(QueueManager.KEYS.get(8)));
        }

        @Test
        void expiredCompletionTokensAreRemovedAndNeverRecreatedByReuse() {
            String id = ticket();
            manager.status(id);
            manager.claim(id);
            String token = (String) manager.complete(id).get("completionToken");
            redis.opsForZSet().add(QueueManager.KEYS.get(8), token, 0);
            assertFalse(manager.reuse(token));
            assertFalse(manager.reuse(token));
            assertFalse(manager.reuse(ticket()));
            assertEquals(0L, redis.opsForZSet().zCard(QueueManager.KEYS.get(8)));
        }

        @Test
        void completionHttpIssuesPrivateCookieAndHintButNormalReleaseAndReplayDoNot() throws Exception {
            MockMvc http = MockMvcBuilders.standaloneSetup(new QueueController(manager)).build();
            String id = ticket();
            manager.status(id);
            manager.claim(id);
            Collection<String> cookies = http.perform(post("/api/queue/release").param("qid", id)
                            .param("completed", "true"))
                    .andExpect(status().isOk()).andExpect(header().string("Cache-Control", "no-store"))
                    .andExpect(jsonPath("$.data.completed").value(true))
                    .andExpect(jsonPath("$.data.completionToken").doesNotExist())
                    .andReturn().getResponse().getHeaders("Set-Cookie");
            assertEquals(2, cookies.size());
            String cookie = cookies.stream().filter(value -> value.startsWith("hoj_loaded=")).findFirst().get();
            String hint = cookies.stream().filter(value -> value.startsWith("hoj_loaded_hint=1;")).findFirst().get();
            for (String value : cookies) {
                assertTrue(value.contains("Path=/"));
                assertTrue(value.contains("Max-Age=86400"));
                assertTrue(value.contains("Secure"));
                assertTrue(value.contains("SameSite=Lax"));
            }
            assertTrue(cookie.contains("HttpOnly"));
            assertFalse(hint.contains("HttpOnly"));
            String token = cookie.substring("hoj_loaded=".length(), cookie.indexOf(';'));
            assertTrue(manager.reuse(token));
            http.perform(post("/api/queue/release").param("qid", id).param("completed", "true"))
                    .andExpect(status().isOk()).andExpect(header().doesNotExist("Set-Cookie"))
                    .andExpect(jsonPath("$.data.completed").value(false));
            String failed = ticket();
            manager.status(failed);
            manager.claim(failed);
            http.perform(post("/api/queue/release").param("qid", failed))
                    .andExpect(status().isOk()).andExpect(header().doesNotExist("Set-Cookie"))
                    .andExpect(jsonPath("$.data.completed").value(false));
            http.perform(post("/api/queue/release").param("qid", failed).param("completed", "true"))
                    .andExpect(status().isOk()).andExpect(header().doesNotExist("Set-Cookie"))
                    .andExpect(jsonPath("$.data.completed").value(false));
        }

        @Test
        void claimValidatesCompletionCookieAndHonorsForcedReloadAndExplicitTicketFirst() throws Exception {
            MockMvc http = MockMvcBuilders.standaloneSetup(new QueueController(manager)).build();
            String id = ticket();
            manager.status(id);
            manager.claim(id);
            String token = (String) manager.complete(id).get("completionToken");
            http.perform(get("/api/queue/claim").header("X-Queue-Warm", token))
                    .andExpect(status().isNoContent()).andExpect(header().string("Cache-Control", "no-store"));
            http.perform(get("/api/queue/claim").header("X-Queue-Warm", token).header("X-Queue-Force", "1"))
                    .andExpect(status().isUnauthorized());
            for (String invalid : new String[]{"", "forged", ticket(), token + " "}) {
                http.perform(get("/api/queue/claim").header("X-Queue-Warm", invalid))
                        .andExpect(status().isUnauthorized());
            }
            http.perform(get("/api/queue/claim").header("X-Queue-Warm", token).header("X-Queue-Ticket", id))
                    .andExpect(status().isUnauthorized());
            http.perform(get("/api/queue/claim").header("X-Queue-Warm", token).header("X-Queue-Ticket", "forged"))
                    .andExpect(status().isUnauthorized());
            String fresh = ticket();
            manager.status(fresh);
            http.perform(get("/api/queue/claim").header("X-Queue-Warm", token).header("X-Queue-Ticket", fresh)
                            .header("X-Queue-Force", "1"))
                    .andExpect(status().isNoContent());
            counts(manager.status(fresh), 0, 0, 1);
            redis.opsForZSet().add(QueueManager.KEYS.get(8), token, 0);
            http.perform(get("/api/queue/claim").header("X-Queue-Warm", token))
                    .andExpect(status().isUnauthorized());
        }

        @Test
        void arrivalSequenceWinsOverUuidAndPollingOrder() {
            manager = new QueueManager(redis, true, 1);
            String holder = ticket();
            String first = "ffffffff-ffff-4fff-bfff-ffffffffffff";
            String second = "00000000-0000-4000-8000-000000000000";
            manager.status(holder);
            manager.status(first);
            assertEquals(2, number(manager.status(second), "position"));
            manager.release(holder);
            assertEquals("waiting", manager.status(second).get("state"));
            assertEquals("reserved", manager.status(first).get("state"));
            manager.release(first);
            assertEquals("reserved", manager.status(second).get("state"));
        }

        @Test
        void pollingIntervalTracksQueuePositionAndSpeedsUpAsItAdvances() {
            manager = new QueueManager(redis, true, 1);
            String holder = ticket();
            assertEquals(1000, number(manager.status(holder), "retryAfterMs"));
            assertEquals(1000, number(manager.claim(holder), "retryAfterMs"));
            List<String> waiters = new ArrayList<>();
            for (int position = 1; position <= 101; position++) {
                String id = ticket();
                waiters.add(id);
                Map<String, Object> result = manager.status(id);
                assertEquals(position, number(result, "position"));
                assertEquals(position > 100 ? 5000 : position > 20 ? 3000 : 1000,
                        number(result, "retryAfterMs"));
            }
            String last = waiters.get(100);
            manager.leave(waiters.get(0), true);
            assertEquals(3000, number(manager.status(last), "retryAfterMs"));
            for (int i = 1; i <= 80; i++) manager.leave(waiters.get(i), true);
            Map<String, Object> advanced = manager.status(last);
            assertEquals(20, number(advanced, "position"));
            assertEquals(1000, number(advanced, "retryAfterMs"));
            counts(advanced, 20, 0, 1);
        }

        @Test
        void reservedOwnerPollingRenewsAnOfferWhoseResponseWasLost() {
            String id = ticket();
            manager.status(id);
            redis.opsForZSet().incrementScore(QueueManager.KEYS.get(1), id, -10000);
            redis.opsForZSet().incrementScore(QueueManager.KEYS.get(2), id, -10000);
            Double previousExpiry = redis.opsForZSet().score(QueueManager.KEYS.get(2), id);
            Map<String, Object> renewed = manager.status(id);
            Double seenExpiry = redis.opsForZSet().score(QueueManager.KEYS.get(1), id);
            Double reservedExpiry = redis.opsForZSet().score(QueueManager.KEYS.get(2), id);
            assertEquals("reserved", renewed.get("state"));
            counts(renewed, 0, 1, 0);
            assertTrue(reservedExpiry >= previousExpiry + 10000);
            assertEquals(5000.0, seenExpiry - reservedExpiry);
            assertNull(redis.opsForZSet().score(QueueManager.KEYS.get(0), id));
        }

        @Test
        void promotionRetainsHeartbeatDeadlineAndClaimStartsLoadingLease() {
            manager = new QueueManager(redis, true, 1);
            String holder = ticket(), next = ticket();
            manager.status(holder);
            manager.status(next);
            Double deadline = redis.opsForZSet().incrementScore(QueueManager.KEYS.get(1), next, -12000);
            counts(manager.release(holder), 0, 1, 0);
            assertEquals(deadline, redis.opsForZSet().score(QueueManager.KEYS.get(1), next));
            assertEquals(deadline, redis.opsForZSet().score(QueueManager.KEYS.get(2), next));
            Map<String, Object> claimed = manager.claim(next);
            assertEquals(true, claimed.get("claimed"));
            counts(claimed, 0, 0, 1);
            assertNull(redis.opsForZSet().score(QueueManager.KEYS.get(1), next));
            assertTrue(redis.opsForZSet().score(QueueManager.KEYS.get(3), next) > deadline);
        }

        @Test
        void absentPromotedPageExpiresAtItsOriginalHeartbeatDeadline() {
            manager = new QueueManager(redis, true, 1);
            String holder = ticket(), absent = ticket(), next = ticket();
            manager.status(holder);
            manager.status(absent);
            manager.status(next);
            Double deadline = redis.opsForZSet().incrementScore(QueueManager.KEYS.get(1), absent, -12000);
            manager.release(holder);
            assertEquals(deadline, redis.opsForZSet().score(QueueManager.KEYS.get(2), absent));
            // Simulate the original heartbeat deadline passing while the reservation still exists.
            redis.opsForZSet().add(QueueManager.KEYS.get(1), absent, 0);
            manager.cleanup();
            assertEquals("closed", manager.status(absent).get("state"));
            assertEquals(false, manager.claim(absent).get("claimed"));
            assertEquals("reserved", manager.status(next).get("state"));
            counts(manager.heartbeat(absent), 0, 1, 0);
        }

        @Test
        void otherPagesCannotRenewAnUnclaimedOfferAndLatePollsCannotReviveIt() {
            String absent = ticket(), other = ticket();
            manager.status(absent);
            Double seenExpiry = redis.opsForZSet().score(QueueManager.KEYS.get(1), absent);
            Double reservedExpiry = redis.opsForZSet().score(QueueManager.KEYS.get(2), absent);
            manager.status(other);
            assertEquals(seenExpiry, redis.opsForZSet().score(QueueManager.KEYS.get(1), absent));
            assertEquals(reservedExpiry, redis.opsForZSet().score(QueueManager.KEYS.get(2), absent));
            redis.opsForZSet().add(QueueManager.KEYS.get(2), absent, 0);
            assertEquals("closed", manager.status(absent).get("state"));
            counts(manager.heartbeat(absent), 0, 1, 0);
            assertNull(redis.opsForZSet().score(QueueManager.KEYS.get(1), absent));
        }

        @Test
        void multipleManagersNeverOverbookAndDuplicateJoinsCountOnce() throws Exception {
            manager = new QueueManager(redis, true, 5);
            QueueManager other = new QueueManager(redis, true, 5);
            List<Callable<Map<String, Object>>> joins = new ArrayList<>();
            String duplicate = ticket();
            for (int i = 0; i < 100; i++) {
                String id = i < 20 ? duplicate : ticket();
                QueueManager target = i % 2 == 0 ? manager : other;
                joins.add(() -> target.status(id));
            }
            for (Map<String, Object> result : concurrent(joins)) {
                assertTrue(number(result, "active") <= 5);
            }
            counts(manager.status(duplicate), 76, 5, 0);
        }

        @Test
        void onlyOneConcurrentClaimConsumesTheTicket() throws Exception {
            String id = ticket();
            manager.status(id);
            List<Callable<Map<String, Object>>> claims = new ArrayList<>();
            for (int i = 0; i < 32; i++) {
                claims.add(() -> manager.claim(id));
            }
            long accepted = concurrent(claims).stream().filter(r -> Boolean.TRUE.equals(r.get("claimed"))).count();
            assertEquals(1, accepted);
            counts(manager.status(id), 0, 0, 1);
        }

        @Test
        void releaseRacingHeartbeatsAndStatusCannotResurrectLoading() throws Exception {
            manager = new QueueManager(redis, true, 1);
            String a = ticket(), b = ticket();
            manager.status(a);
            manager.claim(a);
            manager.status(b);
            List<Callable<Map<String, Object>>> operations = new ArrayList<>();
            for (int i = 0; i < 30; i++) {
                operations.add(() -> manager.heartbeat(a));
                operations.add(() -> manager.status(a));
                if (i == 10) operations.add(() -> manager.release(a));
            }
            concurrent(operations);
            assertEquals("closed", manager.status(a).get("state"));
            assertEquals("reserved", manager.status(b).get("state"));
            counts(manager.heartbeat(a), 0, 1, 0);
        }

        @Test
        void cancellationBeforeLateJoinLeavesATombstone() {
            String id = ticket();
            manager.release(id);
            assertEquals("closed", manager.status(id).get("state"));
            assertEquals(false, manager.claim(id).get("claimed"));
            counts(manager.heartbeat(id), 0, 0, 0);
            String cancelled = ticket();
            manager.leave(cancelled, true);
            assertEquals("closed", manager.status(cancelled).get("state"));
        }

        @Test
        void refreshCancelsGraceWithoutChangingFifoPosition() {
            manager = new QueueManager(redis, true, 1);
            manager.status(ticket());
            String first = ticket(), second = ticket();
            manager.status(first);
            manager.status(second);
            manager.leave(first, false);
            Double deadline = redis.opsForZSet().score(QueueManager.KEYS.get(4), first);
            assertNotNull(deadline);
            manager.leave(first, false);
            assertEquals(deadline, redis.opsForZSet().score(QueueManager.KEYS.get(4), first));
            assertEquals(1, number(manager.status(first), "position"));
            assertNull(redis.opsForZSet().score(QueueManager.KEYS.get(4), first));
            assertEquals(2, number(manager.status(second), "position"));
        }

        @Test
        void expiredGraceClosesWaitingAndReservedButLateUnloadCannotReleaseLoading() {
            String waiting = ticket(), reserved = ticket();
            manager.status(reserved);
            manager.status(ticket());
            manager.status(waiting);
            manager.leave(waiting, false);
            redis.opsForZSet().add(QueueManager.KEYS.get(4), waiting, 0);
            manager.cleanup();
            assertEquals("closed", manager.status(waiting).get("state"));
            manager.leave(reserved, false);
            redis.opsForZSet().add(QueueManager.KEYS.get(4), reserved, 0);
            manager.cleanup();
            assertEquals("closed", manager.status(reserved).get("state"));
            String loading = ticket();
            manager.status(loading);
            manager.claim(loading);
            manager.leave(loading, false);
            assertEquals("loading", manager.status(loading).get("state"));
            assertNull(redis.opsForZSet().score(QueueManager.KEYS.get(4), loading));
        }

        @Test
        void allExpiredStatesCloseAndNextLiveWaiterIsReserved() {
            manager = new QueueManager(redis, true, 1);
            String holder = ticket(), departed = ticket(), next = ticket();
            manager.status(holder);
            manager.status(departed);
            manager.status(next);
            redis.opsForZSet().add(QueueManager.KEYS.get(1), departed, 0);
            redis.opsForZSet().add(QueueManager.KEYS.get(2), holder, 0);
            manager.cleanup();
            assertEquals("closed", manager.status(holder).get("state"));
            assertEquals("closed", manager.status(departed).get("state"));
            assertEquals("reserved", manager.status(next).get("state"));
            manager.claim(next);
            String successor = ticket();
            manager.status(successor);
            redis.opsForZSet().add(QueueManager.KEYS.get(3), next, 0);
            assertEquals("closed", manager.heartbeat(next).get("state"));
            assertEquals("reserved", manager.status(successor).get("state"));
            counts(manager.status(next), 0, 1, 0);
        }

        @Test
        void heartbeatsOnlyExtendExistingLoadingAndClosedEntriesExpire() {
            String id = ticket();
            manager.status(id);
            Double reservedExpiry = redis.opsForZSet().score(QueueManager.KEYS.get(2), id);
            manager.heartbeat(id);
            assertEquals(reservedExpiry, redis.opsForZSet().score(QueueManager.KEYS.get(2), id));
            manager.claim(id);
            redis.opsForZSet().incrementScore(QueueManager.KEYS.get(3), id, -10000);
            Double prior = redis.opsForZSet().score(QueueManager.KEYS.get(3), id);
            manager.heartbeat(id);
            assertTrue(redis.opsForZSet().score(QueueManager.KEYS.get(3), id) > prior);
            manager.release(id);
            redis.opsForZSet().add(QueueManager.KEYS.get(5), id, 0);
            manager.cleanup();
            assertNull(redis.opsForZSet().score(QueueManager.KEYS.get(5), id));
        }

        @Test
        void deletingRevisionDoesNotMakeClientsRejectNewSnapshots() {
            String id = ticket();
            long before = number(manager.status(id), "revision");
            redis.delete(QueueManager.KEYS.get(7));
            long after = number(manager.status(id), "revision");
            assertTrue(after > before);
            assertTrue(after < 9007199254740991L);
        }
    }
}
