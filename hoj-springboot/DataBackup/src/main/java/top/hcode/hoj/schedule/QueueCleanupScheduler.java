package top.hcode.hoj.schedule;

import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import top.hcode.hoj.manager.queue.QueueManager;

import javax.annotation.Resource;

/**
 * 排队房定时清扫：释放超时的活跃槽位与失效的等待项。
 */
@Slf4j(topic = "hoj")
@Component
public class QueueCleanupScheduler {

    @Resource private QueueManager queueManager;

    @Scheduled(fixedDelay = 1000, initialDelay = 1000)
    public void cleanup() {
        try {
            queueManager.cleanup();
        } catch (Exception e) {
            log.error("queue cleanup scheduler failed", e);
        }
    }
}
