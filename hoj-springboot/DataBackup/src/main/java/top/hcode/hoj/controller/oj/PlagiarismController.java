package top.hcode.hoj.controller.oj;

import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.pojo.dto.PlagiarismConfigDTO;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheckConfig;
import top.hcode.hoj.pojo.vo.PlagiarismResultPageVO;
import top.hcode.hoj.service.plagiarism.PlagiarismService;

import javax.annotation.Resource;
import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/api/plagiarism")
@RequiresAuthentication
public class PlagiarismController {
    @Resource private PlagiarismService service;

    @GetMapping("/ping")
    public CommonResult<Void> ping() {
        return CommonResult.successResponse("查重路由已注册");
    }

    @GetMapping("/contest/{cid}/config")
    public CommonResult<List<PlagiarismCheckConfig>> configs(@PathVariable Long cid) {
        return service.configs(cid);
    }

    @PostMapping("/contest/{cid}/config")
    public CommonResult<Void> saveConfig(@PathVariable Long cid, @Valid @RequestBody PlagiarismConfigDTO dto) {
        return service.saveConfig(cid, dto);
    }

    @PostMapping("/contest/{cid}/check/start")
    public CommonResult<PlagiarismCheck> start(@PathVariable Long cid) {
        return service.start(cid);
    }

    @GetMapping("/contest/{cid}/check/progress")
    public CommonResult<PlagiarismCheck> progress(@PathVariable Long cid) {
        return service.progress(cid);
    }

    @GetMapping("/contest/{cid}/check/latest")
    public CommonResult<PlagiarismCheck> latest(@PathVariable Long cid) {
        return service.latest(cid);
    }

    @GetMapping("/check/{checkId}/results")
    public CommonResult<PlagiarismResultPageVO> results(@PathVariable Long checkId,
                                                        @RequestParam(required = false) String displayId) {
        return service.results(checkId, displayId);
    }

    @GetMapping("/check/{checkId}/results/export")
    public ResponseEntity<?> export(@PathVariable Long checkId) {
        try {
            byte[] content = service.export(checkId);
            return ResponseEntity.ok().contentType(MediaType.parseMediaType("text/csv; charset=UTF-8"))
                    .header(HttpHeaders.CONTENT_DISPOSITION, "attachment; filename=plagiarism_results_" + checkId + ".csv")
                    .header("Cache-Control", "no-cache, no-store, must-revalidate")
                    .body(content);
        } catch (Exception ex) {
            return ResponseEntity.status(ResultStatus.NOT_FOUND.getStatus())
                    .body(CommonResult.errorResponse(ex.getMessage(), ResultStatus.NOT_FOUND));
        }
    }

    @GetMapping("/submission/{submitId}")
    public CommonResult<Judge> submission(@PathVariable Long submitId) {
        return service.submission(submitId);
    }
}
