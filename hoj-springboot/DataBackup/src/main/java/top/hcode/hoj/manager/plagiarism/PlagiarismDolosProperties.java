package top.hcode.hoj.manager.plagiarism;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * Configuration for the optional Dolos worker.
 *
 * <p>The worker is required for production plagiarism checks. An empty URL is
 * treated as a configuration error when a check is started.</p>
 */
@Component
public class PlagiarismDolosProperties {
    @Value("${hoj.plagiarism.dolos.worker-url:}")
    private String workerUrl;

    @Value("${hoj.plagiarism.dolos.token:}")
    private String token;

    @Value("${hoj.plagiarism.dolos.enabled:true}")
    private boolean enabled;

    @Value("${hoj.plagiarism.dolos.timeout-ms:900000}")
    private int timeoutMs;

    public String getWorkerUrl() {
        return workerUrl;
    }

    public String getToken() {
        return token;
    }

    public boolean isEnabled() {
        return enabled;
    }

    public boolean isConfigured() {
        return enabled && workerUrl != null && !workerUrl.trim().isEmpty();
    }

    public int getTimeoutMs() {
        return Math.max(1000, Math.min(timeoutMs, 600000));
    }
}
