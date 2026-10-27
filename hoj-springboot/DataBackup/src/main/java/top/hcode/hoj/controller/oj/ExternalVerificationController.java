package top.hcode.hoj.controller.oj;

import com.baomidou.mybatisplus.core.metadata.IPage;
import org.springframework.http.HttpStatus;
import org.springframework.http.converter.HttpMessageNotReadableException;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.annotation.AnonApi;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.oj.ExternalVerificationManager;
import top.hcode.hoj.pojo.dto.ExternalVerificationSubmitDTO;
import top.hcode.hoj.pojo.vo.ExternalVerificationProblemVO;
import top.hcode.hoj.pojo.vo.ExternalVerificationSubmissionVO;
import javax.annotation.Resource;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.net.InetAddress;

@RestController
@RequestMapping("/api/external-verification/{token}")
@AnonApi
public class ExternalVerificationController {
    @Resource private ExternalVerificationManager manager;

    @ModelAttribute
    public void privateResponse(HttpServletResponse response) {
        response.setHeader("Cache-Control", "no-store");
        response.setHeader("Referrer-Policy", "no-referrer");
        response.setHeader("X-Robots-Tag", "noindex, nofollow");
    }

    @GetMapping
    public CommonResult<ExternalVerificationProblemVO> problem(@PathVariable String token)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.problem(token));
    }

    @PostMapping("/submit")
    public CommonResult<ExternalVerificationSubmissionVO> submit(@PathVariable String token,
            @RequestBody ExternalVerificationSubmitDTO dto, HttpServletRequest request)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.submit(token, dto, clientIp(request)));
    }

    @GetMapping("/submissions")
    public CommonResult<IPage<ExternalVerificationSubmissionVO>> submissions(@PathVariable String token,
            @RequestHeader("X-External-Visitor-Token") String visitorToken, @RequestParam(required = false) Integer currentPage,
            @RequestParam(required = false) Integer limit) throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.submissions(token, visitorToken, currentPage, limit));
    }

    @GetMapping("/submissions/{submitId}")
    public CommonResult<ExternalVerificationSubmissionVO> submission(@PathVariable String token,
            @PathVariable Long submitId, @RequestHeader("X-External-Visitor-Token") String visitorToken)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.submission(token, submitId, visitorToken));
    }

    @ExceptionHandler(StatusForbiddenException.class)
    @ResponseStatus(HttpStatus.FORBIDDEN)
    public CommonResult<Void> forbidden(StatusForbiddenException exception) {
        return CommonResult.errorResponse(exception.getMessage(), 403);
    }

    @ExceptionHandler(StatusFailException.class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    public CommonResult<Void> invalid(StatusFailException exception) {
        return CommonResult.errorResponse(exception.getMessage());
    }

    @ExceptionHandler(Exception.class)
    @ResponseStatus(HttpStatus.INTERNAL_SERVER_ERROR)
    public CommonResult<Void> unavailable(Exception exception) {
        return CommonResult.errorResponse("外部验题服务暂不可用，请稍后重试");
    }

    @ExceptionHandler(HttpMessageNotReadableException.class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    public CommonResult<Void> invalidBody(HttpMessageNotReadableException exception) {
        return CommonResult.errorResponse("请求内容无效，JSON 请求体不能超过 512 KiB");
    }

    // Each deployed Nginx appends its peer; walk from the trusted end, never the spoofable prefix.
    static String clientIp(HttpServletRequest request) {
        String peer = request.getRemoteAddr();
        try {
            InetAddress address = InetAddress.getByName(peer);
            if (!address.isLoopbackAddress() && !address.isSiteLocalAddress()) return peer;
            InetAddress real = literalAddress(request.getHeader("X-Real-IP"));
            if (real != null && !real.isLoopbackAddress() && !real.isSiteLocalAddress()) return real.getHostAddress();
            String chain = request.getHeader("X-Forwarded-For");
            if (chain != null && chain.length() <= 2048) {
                String[] hops = chain.split(",", -1);
                for (int i = hops.length - 1; i >= 0; i--) {
                    InetAddress hop = literalAddress(hops[i].trim());
                    if (hop == null) return peer;
                    if (!hop.isLoopbackAddress() && !hop.isSiteLocalAddress()) return hop.getHostAddress();
                }
            }
        } catch (Exception ignored) {
            // Unparseable headers cannot create additional rate-limit buckets.
        }
        return peer;
    }

    private static InetAddress literalAddress(String value) throws java.net.UnknownHostException {
        if (value == null || !(value.matches("[0-9.]{7,15}")
                || (value.contains(":") && value.matches("[0-9a-fA-F:]{2,45}")))) return null;
        return InetAddress.getByName(value);
    }
}
