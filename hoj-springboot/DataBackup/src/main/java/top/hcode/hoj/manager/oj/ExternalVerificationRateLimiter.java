package top.hcode.hoj.manager.oj;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.script.DefaultRedisScript;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusFailException;
import cn.hutool.crypto.digest.DigestUtil;
import javax.annotation.Resource;
import java.util.Arrays;

@Component
public class ExternalVerificationRateLimiter {
    @Resource private StringRedisTemplate redis;

    // Check every quota before incrementing any counter; Redis executes the entire script atomically.
    private static final DefaultRedisScript<Long> SUBMISSION_LIMIT = new DefaultRedisScript<>(
            "local limits = {10, 120, 60, 600}; local ttl = {60, 3600, 60, 3600}; "
            + "for i = 1, #KEYS do if tonumber(redis.call('GET', KEYS[i]) or '0') >= limits[i] then return 0 end end; "
            + "for i = 1, #KEYS do local n = redis.call('INCR', KEYS[i]); "
            + "if n == 1 then redis.call('EXPIRE', KEYS[i], ttl[i]) end end; return 1", Long.class);

    public void checkSubmission(String ip, Long pid) throws StatusFailException {
        String prefix = "external-verification:limit:";
        String address = DigestUtil.sha256Hex(ip == null ? "unknown" : ip);
        Long accepted;
        try {
            accepted = redis.execute(SUBMISSION_LIMIT, Arrays.asList(
                    prefix + "ip:minute:" + address, prefix + "ip:hour:" + address,
                    prefix + "problem:minute:" + pid, prefix + "problem:hour:" + pid));
        } catch (RuntimeException e) {
            throw new StatusFailException("提交服务暂不可用，请稍后重试");
        }
        if (!Long.valueOf(1).equals(accepted)) {
            throw new StatusFailException("外部验题提交过于频繁，请稍后重试");
        }
    }
}
