package top.hcode.hoj.controller.oj;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.CacheControl;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseCookie;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.method.annotation.MethodArgumentTypeMismatchException;
import top.hcode.hoj.annotation.AnonApi;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.queue.QueueManager;

import java.util.Map;

@RestController
@RequestMapping("/api/queue")
@AnonApi
public class QueueController {

    private static final Logger log = LoggerFactory.getLogger(QueueController.class);
    private final QueueManager queueManager;

    public QueueController(QueueManager queueManager) {
        this.queueManager = queueManager;
    }

    @GetMapping("/status")
    public ResponseEntity<CommonResult<Map<String, Object>>> status(
            @RequestParam(value = "qid", required = false) String id) {
        return success(queueManager.status(id));
    }

    @PostMapping("/heartbeat")
    public ResponseEntity<CommonResult<Map<String, Object>>> heartbeat(
            @RequestParam(value = "qid", required = false) String id) {
        return success(queueManager.heartbeat(id));
    }

    @PostMapping("/release")
    public ResponseEntity<CommonResult<Map<String, Object>>> release(
            @RequestParam(value = "qid", required = false) String id,
            @RequestParam(value = "completed", defaultValue = "false") boolean completed) {
        Map<String, Object> result = completed ? queueManager.complete(id) : queueManager.release(id);
        String token = (String) result.remove("completionToken");
        ResponseEntity.BodyBuilder response = ResponseEntity.ok().cacheControl(CacheControl.noStore());
        if (token != null) {
            response.header(HttpHeaders.SET_COOKIE, ResponseCookie.from("hoj_loaded", token)
                    .httpOnly(true).secure(true).sameSite("Lax").path("/")
                    .maxAge(QueueManager.COMPLETION_MAX_AGE_SECONDS).build().toString());
            // This readable hint is never trusted for admission; the private token is verified in Redis.
            response.header(HttpHeaders.SET_COOKIE, ResponseCookie.from("hoj_loaded_hint", "1")
                    .secure(true).sameSite("Lax").path("/")
                    .maxAge(QueueManager.COMPLETION_MAX_AGE_SECONDS).build().toString());
        }
        return response.body(CommonResult.successResponse(result));
    }

    @PostMapping("/leave")
    public ResponseEntity<CommonResult<Map<String, Object>>> leave(
            @RequestParam(value = "qid", required = false) String id,
            @RequestParam(value = "immediate", defaultValue = "false") boolean immediate) {
        return success(queueManager.leave(id, immediate));
    }

    /** Nginx consumes a reservation, or verifies a previous successful load before forwarding HTML. */
    @GetMapping("/claim")
    public ResponseEntity<Void> claim(
            @RequestHeader(value = "X-Queue-Ticket", required = false) String id,
            @RequestHeader(value = "X-Queue-Warm", required = false) String warmToken,
            @RequestHeader(value = "X-Queue-Force", required = false) String force) {
        Map<String, Object> result;
        try {
            if (id == null || id.isEmpty()) {
                return ResponseEntity.status(!"1".equals(force) && queueManager.reuse(warmToken) ? 204 : 401)
                        .cacheControl(CacheControl.noStore()).build();
            }
            result = queueManager.claim(id);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(401).cacheControl(CacheControl.noStore()).build();
        }
        ResponseEntity.BodyBuilder response = ResponseEntity
                .status(Boolean.TRUE.equals(result.get("claimed")) ? 204 : 401)
                .cacheControl(CacheControl.noStore());
        if (Boolean.TRUE.equals(result.get("disabled"))) {
            response.header("X-Queue-Disabled", "true");
        }
        return response.build();
    }

    @ExceptionHandler({IllegalArgumentException.class, MethodArgumentTypeMismatchException.class})
    public ResponseEntity<CommonResult<Map<String, Object>>> invalidRequest(Exception exception) {
        return ResponseEntity.badRequest().cacheControl(CacheControl.noStore())
                .body(CommonResult.errorResponse("无效的排队请求，请重新进入排队页", 400));
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<CommonResult<Map<String, Object>>> unavailable(Exception exception) {
        log.error("Queue request failed", exception);
        return ResponseEntity.status(503).cacheControl(CacheControl.noStore()).header("Retry-After", "1")
                .body(CommonResult.errorResponse("排队服务暂时不可用，请稍后重试", 503));
    }

    private ResponseEntity<CommonResult<Map<String, Object>>> success(Map<String, Object> data) {
        return ResponseEntity.ok().cacheControl(CacheControl.noStore()).body(CommonResult.successResponse(data));
    }
}
