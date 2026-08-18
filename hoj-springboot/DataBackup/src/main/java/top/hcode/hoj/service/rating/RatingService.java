package top.hcode.hoj.service.rating;

import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.entity.rating.*;

import java.util.List;
import java.util.Map;

public interface RatingService {
    CommonResult<?> user(String uid);
    CommonResult<?> history(String uid, int page, int limit);
    CommonResult<?> color(int rating);
    CommonResult<?> contest(Long cid);
    CommonResult<?> batchUsers(List<String> uids);
    CommonResult<?> contestInfo(Long cid);
    CommonResult<?> contestBatch(List<Long> ids);
    CommonResult<?> rank(int page, int limit, String keyword);
    CommonResult<?> setRatingType(Long cid, boolean rated);
    CommonResult<?> calculate(Long cid);
    CommonResult<?> triggerScheduler();
    CommonResult<?> initialize(String uid, int rating);
    CommonResult<?> adjust(Map<String, Object> request);
    CommonResult<?> cancelAdjustment(Long id);
    CommonResult<?> manualHistory(int page, int limit);
    CommonResult<?> skipUsers(Map<String, Object> request);
    CommonResult<?> skipList(Long cid);
    CommonResult<?> cancelSkip(Map<String, Object> request);
    CommonResult<?> recalculate(Long cid);
    CommonResult<?> progress(Long taskId);
    CommonResult<?> operationLogs(int page, int limit, String type, String range);
    CommonResult<?> resetLock(Long cid);
    CommonResult<?> syncSkip(Long cid);
}
