package top.hcode.hoj.controller.oj;

import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.service.rating.RatingService;

import javax.annotation.Resource;
import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/rating")
public class RatingController {
    @Resource private RatingService service;

    @GetMapping("/user/{uid}") public CommonResult<?> user(@PathVariable String uid) { return service.user(uid); }
    @GetMapping("/history/{uid}") public CommonResult<?> history(@PathVariable String uid, @RequestParam(defaultValue = "1") int page, @RequestParam(defaultValue = "20") int limit) { return service.history(uid, page, limit); }
    @GetMapping("/color/{rating}") public CommonResult<?> color(@PathVariable int rating) { return service.color(rating); }
    @GetMapping("/contest/{cid}") public CommonResult<?> contest(@PathVariable Long cid) { return service.contest(cid); }
    @PostMapping("/batch") public CommonResult<?> batch(@RequestBody Map<String, Object> request) { return service.batchUsers(strings(request.get("uids"))); }
    @GetMapping("/contest/info/{cid}") public CommonResult<?> contestInfo(@PathVariable Long cid) { return service.contestInfo(cid); }
    @PostMapping("/contest/batch") public CommonResult<?> contestBatch(@RequestBody Map<String, Object> request) { return service.contestBatch(longs(request.get("contestIds"))); }
    @GetMapping("/rank") public CommonResult<?> rank(@RequestParam(defaultValue = "1") int page, @RequestParam(defaultValue = "30") int limit, @RequestParam(defaultValue = "") String keyword) { return service.rank(page, limit, keyword); }

    private static List<String> strings(Object value) { return value instanceof List ? (List<String>) value : Collections.emptyList(); }
    private static List<Long> longs(Object value) { if (!(value instanceof List)) return Collections.emptyList(); java.util.ArrayList<Long> result = new java.util.ArrayList<>(); for (Object item : (List<?>) value) result.add(number(item, 0).longValue()); return result; }
    private static Integer number(Object value, int fallback) { return value instanceof Number ? ((Number) value).intValue() : fallback; }
}
