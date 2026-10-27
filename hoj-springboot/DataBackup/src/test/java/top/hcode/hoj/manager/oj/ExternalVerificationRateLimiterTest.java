package top.hcode.hoj.manager.oj;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.script.RedisScript;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ExternalVerificationRateLimiterTest {
    private final StringRedisTemplate redis = mock(StringRedisTemplate.class);
    private final ExternalVerificationRateLimiter limiter = new ExternalVerificationRateLimiter();

    @BeforeEach
    void setup() { ReflectionTestUtils.setField(limiter, "redis", redis); }

    @Test
    void noQuotaAndUnavailableRedisBothFailClosed() {
        when(redis.execute(any(RedisScript.class), anyList())).thenReturn(0L);
        assertThrows(StatusFailException.class, () -> limiter.checkSubmission("203.0.113.1", 10L));
        when(redis.execute(any(RedisScript.class), anyList())).thenReturn(null);
        assertThrows(StatusFailException.class, () -> limiter.checkSubmission("203.0.113.1", 10L));
        when(redis.execute(any(RedisScript.class), anyList())).thenThrow(new IllegalStateException("internal address"));
        StatusFailException error = assertThrows(StatusFailException.class, () -> limiter.checkSubmission("203.0.113.1", 10L));
        assertFalse(error.getMessage().contains("internal"));
    }

    @Test
    void changingProblemDoesNotResetIpQuotaAndChangingIpDoesNotResetProblemQuota() throws Exception {
        when(redis.execute(any(RedisScript.class), anyList())).thenReturn(1L);
        limiter.checkSubmission("203.0.113.1", 10L);
        limiter.checkSubmission("203.0.113.1", 11L);
        limiter.checkSubmission("203.0.113.2", 10L);
        ArgumentCaptor<List> keys = ArgumentCaptor.forClass(List.class);
        verify(redis, times(3)).execute(any(RedisScript.class), keys.capture());
        List<List> values = keys.getAllValues();
        assertEquals(values.get(0).subList(0, 2), values.get(1).subList(0, 2));
        assertEquals(values.get(0).subList(2, 4), values.get(2).subList(2, 4));
        assertNotEquals(values.get(0).get(0), values.get(2).get(0));
    }
}
