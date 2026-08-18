package top.hcode.hoj.service.rating.impl;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.manager.rating.RatingManager;
import top.hcode.hoj.manager.rating.RatingQueryManager;
import top.hcode.hoj.service.rating.RatingService;

import javax.annotation.Resource;
import java.util.List;
import java.util.Map;

@Slf4j
@Service
public class RatingServiceImpl implements RatingService {
    @Resource private RatingManager manager;
    @Resource private RatingQueryManager query;
    @Override public CommonResult<?> user(String uid) { return run(() -> query.user(uid)); }
    @Override public CommonResult<?> history(String uid, int page, int limit) { return run(() -> query.history(uid, page, limit)); }
    @Override public CommonResult<?> color(int rating) { return run(() -> query.color(rating)); }
    @Override public CommonResult<?> contest(Long cid) { return run(() -> query.contest(cid)); }
    @Override public CommonResult<?> batchUsers(List<String> uids) { return run(() -> query.batchUsers(uids)); }
    @Override public CommonResult<?> contestInfo(Long cid) { return run(() -> query.contestInfo(cid)); }
    @Override public CommonResult<?> contestBatch(List<Long> ids) { return run(() -> query.contestBatch(ids)); }
    @Override public CommonResult<?> rank(int page, int limit, String keyword) { return run(() -> query.rank(page, limit, keyword)); }
    @Override public CommonResult<?> setRatingType(Long cid, boolean rated) { return runVoid(() -> manager.setRatingType(cid, rated)); }
    @Override public CommonResult<?> calculate(Long cid) { return run(() -> manager.calculate(cid)); }
    @Override public CommonResult<?> triggerScheduler() { return run(() -> java.util.Collections.singletonMap("calculated", manager.calculatePending())); }
    @Override public CommonResult<?> initialize(String uid, int rating) { return runVoid(() -> manager.initialize(uid, rating)); }
    @Override public CommonResult<?> adjust(Map<String, Object> request) { return run(() -> manager.adjust(string(request, "username"), integer(request, "ratingChange"), string(request, "reason"), longValue(request.get("relatedContestId")))); }
    @Override public CommonResult<?> cancelAdjustment(Long id) { return run(() -> manager.cancelAdjustment(id)); }
    @Override public CommonResult<?> manualHistory(int page, int limit) { return run(() -> query.manualHistory(page, limit)); }
    @Override public CommonResult<?> skipUsers(Map<String, Object> request) { return run(() -> manager.batchSkip(longValue(request.get("contestId")), (List<String>) request.get("usernames"), string(request, "reason"), Boolean.TRUE.equals(request.get("autoRecalc")))); }
    @Override public CommonResult<?> skipList(Long cid) { return run(() -> manager.skipUsers(cid)); }
    @Override public CommonResult<?> cancelSkip(Map<String, Object> request) { return runVoid(() -> manager.cancelSkip(longValue(request.get("contestId")), (List<String>) request.get("uids"))); }
    @Override public CommonResult<?> recalculate(Long cid) { return run(() -> { Map<String, Object> result = new java.util.LinkedHashMap<>(); result.put("taskId", manager.recalculate(cid)); result.put("message", "重算任务已创建"); return result; }); }
    @Override public CommonResult<?> progress(Long taskId) { return run(() -> manager.progress(taskId)); }
    @Override public CommonResult<?> operationLogs(int page, int limit, String type, String range) { return run(() -> { Map<String, Object> data = new java.util.LinkedHashMap<>(); data.put("records", manager.logs(page, limit, type, range)); data.put("total", manager.logCount(type, range)); data.put("page", page); data.put("limit", limit); return data; }); }
    @Override public CommonResult<?> resetLock(Long cid) { return runVoid(() -> manager.resetLock(cid)); }
    @Override public CommonResult<?> syncSkip(Long cid) { return run(() -> manager.syncSkip(cid)); }

    private <T> CommonResult<T> run(CheckedSupplier<T> action) { try { return CommonResult.successResponse(action.get()); } catch (StatusNotFoundException e) { return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND); } catch (StatusForbiddenException e) { return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN); } catch (StatusFailException e) { return CommonResult.errorResponse(e.getMessage(), ResultStatus.FAIL); } catch (Exception e) { log.error("rating request failed", e); return CommonResult.errorResponse(e.getMessage() == null ? "Rating 服务暂时不可用" : e.getMessage(), ResultStatus.SYSTEM_ERROR); } }
    private CommonResult<Void> runVoid(CheckedRunnable action) { return run(() -> { action.run(); return null; }); }
    private static String string(Map<String, Object> map, String key) { return map.get(key) == null ? "" : String.valueOf(map.get(key)); }
    private static int integer(Map<String, Object> map, String key) { return map.get(key) == null ? 0 : ((Number) map.get(key)).intValue(); }
    private static Long longValue(Object value) { return value == null ? null : ((Number) value).longValue(); }
    @FunctionalInterface private interface CheckedSupplier<T> { T get() throws Exception; }
    @FunctionalInterface private interface CheckedRunnable { void run() throws Exception; }
}
