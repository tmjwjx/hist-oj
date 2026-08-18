package top.hcode.hoj.schedule;

import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import top.hcode.hoj.manager.rating.RatingManager;

import javax.annotation.Resource;

/** 统一由 Java 主后端处理比赛结束后的站内 Rating 计算。 */
@Slf4j
@Component
public class RatingScheduler {
    @Resource
    private RatingManager ratingManager;

    @Scheduled(fixedDelay = 60000, initialDelay = 30000)
    public void calculatePending() {
        try {
            int count = ratingManager.calculatePending();
            if (count > 0) log.info("自动完成 {} 场比赛的 Rating 计算", count);
        } catch (Exception e) {
            log.error("自动计算比赛 Rating 失败", e);
        }
    }
}
