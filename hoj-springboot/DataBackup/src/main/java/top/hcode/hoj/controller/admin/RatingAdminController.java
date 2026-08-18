package top.hcode.hoj.controller.admin;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.service.rating.RatingService;

import javax.annotation.Resource;
import java.util.Map;

@RestController
@RequestMapping("/api/rating")
@RequiresAuthentication
@RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
public class RatingAdminController {
    @Resource private RatingService service;

    @PostMapping("/initialize") public CommonResult<?> initialize(@RequestBody Map<String, Object> r) { return service.initialize(String.valueOf(r.get("uid")), number(r.get("initialRating"), 0)); }
    @PostMapping("/contest/set-rating-type") public CommonResult<?> setType(@RequestBody Map<String, Object> r) { return service.setRatingType(number(r.get("contestId"), 0).longValue(), Boolean.TRUE.equals(r.get("isRating"))); }
    @PostMapping("/calculate/{cid}") public CommonResult<?> calculate(@PathVariable Long cid) { return service.calculate(cid); }
    @PostMapping("/trigger-scheduler") public CommonResult<?> scheduler() { return service.triggerScheduler(); }
    @PostMapping("/admin/adjust") public CommonResult<?> adjust(@RequestBody Map<String, Object> r) { return service.adjust(r); }
    @PostMapping("/admin/adjust/cancel") public CommonResult<?> cancelAdjust(@RequestBody Map<String, Object> r) { return service.cancelAdjustment(number(r.get("adjustmentId"), 0).longValue()); }
    @GetMapping("/admin/history") public CommonResult<?> history(@RequestParam(defaultValue = "1") int page, @RequestParam(defaultValue = "20") int limit) { return service.manualHistory(page, limit); }
    @PostMapping("/admin/contest/skip-users") public CommonResult<?> skip(@RequestBody Map<String, Object> r) { return service.skipUsers(r); }
    @GetMapping("/admin/contest/{cid}/skip-users") public CommonResult<?> skipList(@PathVariable Long cid) { return service.skipList(cid); }
    @DeleteMapping("/admin/contest/skip-users") public CommonResult<?> cancelSkip(@RequestBody Map<String, Object> r) { return service.cancelSkip(r); }
    @PostMapping("/admin/contest/{cid}/recalculate") public CommonResult<?> recalculate(@PathVariable Long cid) { return service.recalculate(cid); }
    @GetMapping("/admin/recalculate-progress/{taskId}") public CommonResult<?> progress(@PathVariable Long taskId) { return service.progress(taskId); }
    @GetMapping("/admin/operation-logs") public CommonResult<?> logs(@RequestParam(defaultValue = "1") int page, @RequestParam(defaultValue = "20") int limit, @RequestParam(defaultValue = "") String type, @RequestParam(defaultValue = "7d") String timeRange) { return service.operationLogs(page, limit, type, timeRange); }
    @PostMapping("/admin/contest/{cid}/sync-skip") public CommonResult<?> sync(@PathVariable Long cid) { return service.syncSkip(cid); }
    @PostMapping("/admin/contest/{cid}/reset-recalculate-lock") public CommonResult<?> reset(@PathVariable Long cid) { return service.resetLock(cid); }
    @PostMapping("/admin/migrate-logs") public CommonResult<?> migrate() { return CommonResult.successResponse(java.util.Collections.singletonMap("migrated", 0)); }
    @PostMapping("/admin/fix-logs-username") public CommonResult<?> fix() { return CommonResult.successResponse(); }

    private static Integer number(Object value, int fallback) { return value instanceof Number ? ((Number) value).intValue() : fallback; }
}
