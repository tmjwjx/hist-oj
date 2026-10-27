package top.hcode.hoj.manager.admin.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONUtil;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.problem.ProblemAiRecordEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.pojo.dto.LastAcceptedCodeVO;
import top.hcode.hoj.pojo.dto.ProblemAiChatDTO;
import top.hcode.hoj.pojo.dto.ProblemAiGenerateProgramDTO;
import top.hcode.hoj.pojo.dto.ProblemAiRecheckDTO;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ProblemAiConfigVO;
import top.hcode.hoj.pojo.vo.ProblemAiGeneratedProgramVO;
import top.hcode.hoj.pojo.vo.ProblemAiLanguageProgramVO;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.service.problem.ProblemVerificationChangeDetector;
import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;

import javax.annotation.Resource;
import java.util.*;
import java.util.concurrent.*;

@Component
public class ProblemAiManager {
    private static final int DEFAULT_AI_TIMEOUT_SECONDS = 120;
    /** AI 验题的权威标准程序固定使用判题机支持的 C++17。 */
    private static final String CANONICAL_STANDARD_LANGUAGE = "C++ 17";
    private static final long JUDGE_WAIT_TIMEOUT_MS = 20 * 60 * 1000L;
    private static final ExecutorService FULL_VALIDATION_POOL = Executors.newFixedThreadPool(2, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-full-validation");
        thread.setDaemon(false);
        return thread;
    });
    private static final ExecutorService PROGRAM_POOL = Executors.newFixedThreadPool(3, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-program");
        thread.setDaemon(true);
        return thread;
    });
    private static final ExecutorService COMPARISON_POOL = Executors.newFixedThreadPool(2, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-comparison");
        thread.setDaemon(true);
        return thread;
    });
    @Resource private ProblemAiConfigManager problemAiConfigManager;
    @Resource private ProblemAiRecordEntityService recordService;
    @Resource private ProblemEntityService problemService;
    @Resource private ProblemCaseEntityService caseService;
    @Resource private ProblemAiPromptFactory promptFactory;
    @Resource private ProblemAiGateway gateway;
    @Resource private ProblemAiValidationRunner validationRunner;
    @Resource private ProblemAiValidationContext validationContext;
    @Resource private ProblemVerificationManager verificationManager;

    /** A process restart cannot resume an in-flight provider request. */
    @PostConstruct
    public void recoverInterruptedRecords() {
        List<ProblemAiRecord> running = recordService.list(new QueryWrapper<ProblemAiRecord>()
                .eq("status", "running"));
        if (running == null || running.isEmpty()) return;
        Date now = new Date();
        for (ProblemAiRecord record : running) {
            record.setStatus("failed")
                    .setErrorMessage("服务重启导致本次 AI 验题中断，请重新发起")
                    .setDurationMs(record.getGmtCreate() == null ? null
                            : (int) Math.min(Integer.MAX_VALUE, now.getTime() - record.getGmtCreate().getTime()));
        }
        recordService.updateBatchById(running);
    }

    public ProblemAiConfigVO getConfig() throws StatusForbiddenException, StatusFailException {
        return problemAiConfigManager.getConfig();
    }

    public void saveConfig(ProblemAiConfig input) throws StatusFailException, StatusForbiddenException {
        problemAiConfigManager.saveConfig(input);
    }

    public List<ProblemAiRecord> records(Long pid) throws StatusFailException, StatusForbiddenException {
        checkAdmin();
        requireProblem(pid);
        return recordService.list(new QueryWrapper<ProblemAiRecord>().eq("pid", pid).orderByAsc("gmt_create"));
    }

    public ProblemAiRecord record(Long id) throws StatusFailException, StatusForbiddenException {
        checkAdmin();
        ProblemAiRecord record = recordService.getById(id);
        if (record == null) throw new StatusFailException("AI 验题记录不存在");
        requireProblem(record.getPid());
        return record;
    }

    /**
     * Recheck one completed report using the current problem content. This intentionally
     * does not submit a new standard program or rerun the full judge; the old report is
     * supplied as historical context and the result is stored as a separate record.
     */
    public ProblemAiRecord recheck(ProblemAiRecheckDTO dto) throws Exception {
        checkAdmin();
        if (dto == null || dto.getRecordId() == null || StringUtils.isEmpty(dto.getRequirements())) {
            throw new StatusFailException("复检报告和复检要求不能为空");
        }
        ProblemAiRecord source = recordService.getById(dto.getRecordId());
        if (source == null) throw new StatusFailException("复检来源报告不存在");
        if (!("AI 一键验题".equals(source.getQuestion()) || "AI 验题复检".equals(source.getQuestion()))
                || !"success".equals(source.getStatus())) {
            throw new StatusFailException("请选择一份已完成的综合验题或复检报告");
        }
        String requirements = dto.getRequirements().trim();
        if (requirements.length() > 4000) throw new StatusFailException("复检要求不能超过 4000 个字符");
        Problem problem = requireProblem(source.getPid());
        ProblemAiConfig config = requireConfig();
        ProblemVerificationDraft draft = verificationManager.getDraft(problem.getId());
        ProblemVerification verification = verificationManager.getVerification(problem.getId());
        List<ProblemCase> currentCases = problemCases(problem.getId());
        String currentLanguage = draft == null ? "" : draft.getLanguage();
        String currentStandardProgram = draft == null ? "" : draft.getCode();
        if (StringUtils.isEmpty(currentStandardProgram)) {
            LastAcceptedCodeVO passed = verificationManager.getLastPassedCode(problem.getId());
            currentLanguage = passed == null ? "" : passed.getLanguage();
            currentStandardProgram = passed == null ? "" : passed.getCode();
        }
        String session = UUID.randomUUID().toString();
        ProblemAiRecord result = new ProblemAiRecord().setPid(problem.getId()).setSessionId(session)
                .setUid(currentUid()).setQuestion("AI 验题复检").setStatus("running").setGmtCreate(new Date());
        recordService.save(result);
        long started = System.currentTimeMillis();
        try {
            String sourceReport = source.getResponse();
            if (StringUtils.isEmpty(sourceReport)) {
                throw new StatusFailException("复检来源报告没有可用内容");
            }
            JSONObject previous = parseJson(sourceReport);
            JSONObject historicalExecution = previous.getJSONObject("execution");
            String prompt = promptFactory.recheckPrompt(config, problem, currentCases,
                    currentLanguage, currentStandardProgram, sourceReport, requirements,
                    problem.getCaseVersion(), verification,
                    validationContext.currentTestcaseEvidence(problem.getId(), currentCases,
                            verification == null ? null : verification.getSubmitId()),
                    historicalExecution == null ? "未提供" : historicalExecution.toString());
            String content = request(config, problem, prompt, session, result.getId(), false, false,
                    config.getTimeoutSeconds() == null ? DEFAULT_AI_TIMEOUT_SECONDS : config.getTimeoutSeconds());
            JSONObject checked = parseJson(content);
            if (!checked.containsKey("overall")) checked.set("overall", "WARN");
            if (!checked.containsKey("steps")) checked.set("steps", new cn.hutool.json.JSONArray());
            if (!checked.containsKey("issues")) checked.set("issues", new cn.hutool.json.JSONArray());
            if (!checked.containsKey("sampleResults")) checked.set("sampleResults", new cn.hutool.json.JSONArray());
            // Keep the reviewed program visible in the new composite report, but do not
            // copy old execution results into the current conclusion as if they were rerun.
            if (!StringUtils.isEmpty(currentStandardProgram)) {
                checked.set("standardProgram", new JSONObject()
                        .set("language", currentLanguage)
                        .set("code", currentStandardProgram)
                        .set("source", draft == null ? "last_accepted" : "current_admin_draft"));
            } else if (previous.containsKey("standardProgram")) {
                checked.set("standardProgram", previous.get("standardProgram"));
            }
            checked.set("recheck", new JSONObject()
                    .set("sourceRecordId", source.getId())
                    .set("requirements", requirements)
                    .set("mode", "report_only")
                    .set("fullJudgeRerun", false)
                    .set("historicalExecution", historicalExecution == null ? null : historicalExecution)
                    .set("currentCaseVersion", problem.getCaseVersion())
                    .set("currentVerificationCaseVersion", verification == null ? null : verification.getCaseVersion())
                    .set("currentSubmitId", verification == null ? null : verification.getSubmitId())
                    .set("currentTestcaseCount", currentCases.size())
                    .set("currentTestcaseEvidenceResolved", true)
                    .set("note", "旧报告的判题与测试点结果仅作为历史证据，本次未重新执行全量判题"));
            result.setResponse(checked.toString()).setStatus("success")
                    .setDurationMs((int) (System.currentTimeMillis() - started));
        } catch (Exception e) {
            result.setStatus("failed").setErrorMessage(trim(e.getMessage()))
                    .setDurationMs((int) (System.currentTimeMillis() - started));
        }
        recordService.updateById(result);
        if ("failed".equals(result.getStatus())) throw new StatusFailException(result.getErrorMessage());
        return result;
    }

    public ProblemAiRecord chat(ProblemAiChatDTO dto) throws Exception {
        checkAdmin();
        Problem problem = requireProblem(dto.getPid());
        if (StringUtils.isEmpty(dto.getMessage())) throw new StatusFailException("请输入要让 AI 检查的内容");
        ProblemAiConfig config = requireConfig();
        String session = StringUtils.isEmpty(dto.getSessionId()) ? UUID.randomUUID().toString() : dto.getSessionId();
        ProblemAiRecord record = new ProblemAiRecord().setPid(dto.getPid()).setSessionId(session)
                .setUid(currentUid()).setQuestion(dto.getMessage()).setStatus("running").setGmtCreate(new Date());
        recordService.save(record);
        long started = System.currentTimeMillis();
        try {
            String prompt = promptFactory.chatPrompt(config, problem, problemCases(problem.getId()), dto.getMessage());
            String content = request(config, problem, prompt, session, record.getId(), true);
            record.setResponse(content).setStatus("success").setDurationMs((int) (System.currentTimeMillis() - started));
        } catch (Exception e) {
            record.setStatus("failed").setErrorMessage(trim(e.getMessage())).setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            throw e;
        }
        recordService.updateById(record);
        return record;
    }

    public ProblemAiRecord validate(ProblemAiValidateDTO dto) throws Exception {
        checkAdmin();
        if (dto == null || dto.getPid() == null || StringUtils.isEmpty(dto.getLanguage())
                || StringUtils.isEmpty(dto.getStandardProgram()) || dto.getStandardProgram().length() > 65535) {
            throw new StatusFailException("题目、语言和标准程序不能为空，代码不能超过 65535 个字符");
        }
        Problem problem = requireProblem(dto.getPid());
        ProblemAiConfig config = requireConfig();
        String session = UUID.randomUUID().toString();
        if (!"interactive".equals(problem.getJudgeMode())) requireIntegratedProgram(dto.getStandardProgram());
        ProblemAiRecord record = new ProblemAiRecord().setPid(dto.getPid()).setSessionId(session)
                .setUid(currentUid()).setQuestion("AI 一键验题").setStatus("running").setGmtCreate(new Date());
        recordService.save(record);
        long started = System.currentTimeMillis();
        try {
            String content = validationRunner.run(config, problem, problemCases(problem.getId()), dto, currentUid(), currentUsername());
            record.setResponse(content).setStatus("success")
                    .setDurationMs((int) (System.currentTimeMillis() - started));
        } catch (Exception e) {
            record.setStatus("failed").setErrorMessage(trim(e.getMessage()))
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            throw e;
        }
        recordService.updateById(record);
        return record;
    }

    public ProblemAiGeneratedProgramVO generateProgram(ProblemAiGenerateProgramDTO dto) throws Exception {
        checkAdmin();
        if (dto == null || dto.getPid() == null) throw new StatusFailException("题目不能为空");
        Problem problem = requireProblem(dto.getPid());
        ProblemAiConfig config = requireConfig();
        List<ProblemAiModelConfig> selected = problemAiConfigManager.selectModels(config, dto.getModelIds());
        ProblemAiModelConfig primary = selected.stream().filter(model -> Boolean.TRUE.equals(model.getPrimary()))
                .findFirst().orElseThrow(() -> new StatusFailException("验题必须包含主模型"));
        String session = UUID.randomUUID().toString();
        ProblemAiRecord record = new ProblemAiRecord().setPid(dto.getPid()).setSessionId(session)
                .setUid(currentUid()).setQuestion("AI 生成标准程序").setStatus("running").setGmtCreate(new Date());
        recordService.save(record);
        long started = System.currentTimeMillis();
        try {
            ProblemAiConfig primaryConfig = problemAiConfigManager.forModel(config, primary);
            String content = request(primaryConfig, problem,
                    promptFactory.standardProgramPrompt(primaryConfig, problem, CANONICAL_STANDARD_LANGUAGE, 0),
                    session, record.getId(), false, true);
            ProblemAiGeneratedProgramVO program = parseProgram(content, !"interactive".equals(problem.getJudgeMode())).setRecordId(record.getId())
                    .setModelId(primary.getId()).setModelName(primary.getName()).setModel(primary.getModel());
            ProblemAiRecord validationRecord = new ProblemAiRecord().setPid(problem.getId())
                    .setSessionId(UUID.randomUUID().toString()).setUid(currentUid())
                    .setQuestion("AI 一键验题").setStatus("running").setGmtCreate(new Date());
            recordService.save(validationRecord);
            program.setValidationRecordId(validationRecord.getId());
            record.setResponse(JSONUtil.toJsonStr(program)).setStatus("success")
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            submitFullValidationTask(config, selected, problem, program, validationRecord,
                    currentUid(), currentUsername());
            return program;
        } catch (Exception e) {
            record.setStatus("failed").setErrorMessage(trim(e.getMessage()))
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            throw e;
        }
    }

    private void submitFullValidationTask(ProblemAiConfig config, List<ProblemAiModelConfig> models,
                                          Problem problem, ProblemAiGeneratedProgramVO program,
                                          ProblemAiRecord record, String uid, String username) {
        try {
            FULL_VALIDATION_POOL.submit(() -> {
                try {
                    runFullValidation(config, models, problem, program, record, uid, username);
                } catch (Exception e) {
                    record.setStatus("failed").setErrorMessage("后台验题任务异常终止，请查看已保存的阶段结果后重试");
                    recordService.updateById(record);
                }
            });
        } catch (RuntimeException e) {
            record.setStatus("failed").setErrorMessage("后台验题任务无法启动，请重试");
            recordService.updateById(record);
        }
    }

    private void runFullValidation(ProblemAiConfig config, List<ProblemAiModelConfig> models,
                                   Problem problem, ProblemAiGeneratedProgramVO program,
                                   ProblemAiRecord record, String uid, String username) throws Exception {
        long started = System.currentTimeMillis();
        JSONArray results = new JSONArray();
        for (ProblemAiModelConfig model : models) results.add(modelResult(model));
        JSONObject report = new JSONObject().set("overall", "WARN").set("summary", "正在并行执行主模型全流程与辅助模型验题")
                .set("standardProgram", JSONUtil.parseObj(program)).set("modelResults", results)
                .set("difficulty", program.getDifficulty()).set("caseVersion", problem.getCaseVersion());
        persistProgress(record, report, started);
        Map<Integer, Future<JSONObject>> comparisons = new LinkedHashMap<>();
        List<Future<ProblemAiLanguageProgramVO>> translations = new ArrayList<>();
        CountDownLatch primaryJudged = new CountDownLatch(1);
        try {
            try {
                ensureCaseVersion(problem);
            } catch (StatusFailException e) {
                for (Object value : results) ((JSONObject) value).set("status", "ERROR").set("errorMessage", e.getMessage());
                report.set("overall", "FAIL").set("summary", e.getMessage());
                record.setStatus("success");
                persistProgress(record, report, started);
                return;
            }
            // Independent generations use separate bounded pools: no task waits for its own pool.
            for (int index = 0; index < models.size(); index++) {
                ProblemAiModelConfig model = models.get(index);
                if (!Boolean.TRUE.equals(model.getPrimary())) comparisons.put(index,
                        COMPARISON_POOL.submit(() -> runComparison(config, model, problem, record.getId(), uid, username, primaryJudged)));
            }
            for (String language : Arrays.asList("Java", "PyPy3")) translations.add(PROGRAM_POOL.submit(() ->
                    generateLanguageProgram(config, problem, program, record.getId(), language)));

            for (Object value : results) {
                JSONObject primaryResult = (JSONObject) value;
                if (!primaryResult.getBool("primary", false)) continue;
                long primaryStarted = System.currentTimeMillis();
                addProgram(primaryResult, program);
                try {
                    ensureCaseVersion(problem);
                    verificationManager.waitForAiJudge(problem.getId(), JUDGE_WAIT_TIMEOUT_MS);
                    ensureCaseVersion(problem);
                    Long submitId = verificationManager.submitForAi(submission(problem, program), uid, username);
                    primaryResult.set("submitId", submitId);
                    collectComparisons(comparisons, results, false);
                    persistProgress(record, report, started);
                    verificationManager.waitForAiJudge(problem.getId(), submitId, JUDGE_WAIT_TIMEOUT_MS);
                    ensureCaseVersion(problem);
                    primaryResult.putAll(verificationManager.comparisonResult(problem.getId(), submitId));
                    primaryJudged.countDown();
                    collectComparisons(comparisons, results, false);
                    persistProgress(record, report, started);
                    List<ProblemAiLanguageProgramVO> programs = new ArrayList<>();
                    for (Future<ProblemAiLanguageProgramVO> translation : translations) programs.add(translation.get());
                    ensureCaseVersion(problem);
                    program.setMultiLanguagePrograms(programs);
                    ProblemAiValidateDTO input = new ProblemAiValidateDTO();
                    input.setPid(problem.getId()); input.setSubmitId(submitId);
                    input.setLanguage(program.getLanguage()); input.setStandardProgram(program.getCode());
                    input.setAiGenerated(true); input.setAlgorithmSummary(program.getAlgorithm());
                    input.setMultiLanguagePrograms(programs);
                    collectComparisons(comparisons, results, false);
                    persistProgress(record, report, started);
                    JSONObject fullReport = parseJson(validationRunner.run(config, problem,
                            problemCases(problem.getId()), input, uid, username));
                    report.putAll(fullReport);
                    report.set("modelResults", results).set("difficulty", program.getDifficulty());
                    primaryResult.set("fullValidationStatus", report.getStr("overall", "WARN"));
                    ensureCaseVersion(problem);
                } catch (Exception e) {
                    if (e instanceof InterruptedException) Thread.currentThread().interrupt();
                    if (Thread.currentThread().isInterrupted()) throw e;
                    primaryResult.set("status", "ERROR").set("errorMessage", trim(e.getMessage()));
                    report.set("overall", "FAIL").set("summary", "主模型全流程验题未完成：" + trim(e.getMessage()));
                    report.set("issues", new JSONArray().put(new JSONObject().set("severity", "ERROR")
                            .set("location", "主模型验题").set("detail", trim(e.getMessage()))));
                }
                primaryJudged.countDown();
                primaryResult.set("durationMs", System.currentTimeMillis() - primaryStarted);
                collectComparisons(comparisons, results, false);
                persistProgress(record, report, started);
            }
            // Only the orchestration thread merges JSON and persists progress. No late worker writes.
            collectComparisons(comparisons, results, true);
            try { ensureCaseVersion(problem); }
            catch (StatusFailException e) {
                report.set("overall", "FAIL").set("summary", e.getMessage());
            }
            record.setStatus("success");
            persistProgress(record, report, started);
        } finally {
            for (Future<?> future : translations) if (!future.isDone()) future.cancel(true);
            for (Future<?> future : comparisons.values()) if (!future.isDone()) future.cancel(true);
        }
    }

    private JSONObject runComparison(ProblemAiConfig config, ProblemAiModelConfig model, Problem problem,
                                     Long recordId, String uid, String username, CountDownLatch primaryJudged) throws Exception {
        long started = System.currentTimeMillis();
        JSONObject item = modelResult(model);
        try {
            ensureCaseVersion(problem);
            ProblemAiConfig modelConfig = problemAiConfigManager.forModel(config, model);
            String content = request(modelConfig, problem, promptFactory.auxiliaryProgramPrompt(modelConfig, problem),
                    UUID.randomUUID().toString(), recordId, false, true);
            ProblemAiGeneratedProgramVO auxiliary = parseProgram(content, false);
            addProgram(item, auxiliary);
            // Warm SPJ/interactive compiler caches with the primary before comparisons enter the judge.
            primaryJudged.await();
            ensureCaseVersion(problem);
            Long submitId = verificationManager.submitComparisonForAi(submission(problem, auxiliary), uid, username);
            item.set("submitId", submitId);
            verificationManager.waitForAiJudge(problem.getId(), submitId, JUDGE_WAIT_TIMEOUT_MS);
            ensureCaseVersion(problem);
            item.putAll(verificationManager.comparisonResult(problem.getId(), submitId));
        } catch (Exception e) {
            if (e instanceof InterruptedException) Thread.currentThread().interrupt();
            if (Thread.currentThread().isInterrupted()) throw e;
            item.set("status", "ERROR").set("errorMessage", trim(e.getMessage()));
        }
        return item.set("durationMs", System.currentTimeMillis() - started);
    }

    private void collectComparisons(Map<Integer, Future<JSONObject>> futures, JSONArray results, boolean wait)
            throws InterruptedException, ExecutionException {
        Iterator<Map.Entry<Integer, Future<JSONObject>>> iterator = futures.entrySet().iterator();
        while (iterator.hasNext()) {
            Map.Entry<Integer, Future<JSONObject>> entry = iterator.next();
            if (wait || entry.getValue().isDone()) {
                results.set(entry.getKey(), entry.getValue().get());
                iterator.remove();
            }
        }
    }

    private JSONObject modelResult(ProblemAiModelConfig model) {
        return new JSONObject().set("modelId", model.getId()).set("name", model.getName())
                .set("model", model.getModel()).set("primary", model.getPrimary()).set("status", "RUNNING");
    }

    private void addProgram(JSONObject result, ProblemAiGeneratedProgramVO program) {
        result.set("language", program.getLanguage()).set("code", program.getCode())
                .set("algorithm", program.getAlgorithm()).set("warnings", program.getWarnings())
                .set("difficulty", program.getDifficulty());
    }

    private ProblemVerificationSubmitDTO submission(Problem problem, ProblemAiGeneratedProgramVO program) {
        ProblemVerificationSubmitDTO input = new ProblemVerificationSubmitDTO();
        input.setPid(problem.getId()); input.setLanguage(program.getLanguage()); input.setCode(program.getCode());
        input.setVerificationType("ai_validation");
        return input;
    }

    private void persistProgress(ProblemAiRecord record, JSONObject report, long started) {
        record.setResponse(report.toString()).setDurationMs((int) Math.min(Integer.MAX_VALUE, System.currentTimeMillis() - started));
        recordService.updateById(record);
    }

    private void ensureCaseVersion(Problem snapshot) throws StatusFailException {
        if (Thread.currentThread().isInterrupted()) throw new StatusFailException("后台 AI 验题任务被中断");
        Problem current = requireProblem(snapshot.getId());
        if (ProblemVerificationChangeDetector.changed(snapshot, current)
                || !Objects.equals(snapshot.getTitle(), current.getTitle())
                || !Objects.equals(snapshot.getDescription(), current.getDescription())
                || !Objects.equals(snapshot.getInput(), current.getInput())
                || !Objects.equals(snapshot.getOutput(), current.getOutput())) {
            throw new StatusFailException("验题期间题目或测试数据已变化，本次结果不能作为新版本的验题结论，请重新验题");
        }
    }

    private ProblemAiLanguageProgramVO generateLanguageProgram(ProblemAiConfig config, Problem problem,
                                                               ProblemAiGeneratedProgramVO canonical,
                                                               Long recordId, String language) throws Exception {
        ProblemAiLanguageProgramVO translated = new ProblemAiLanguageProgramVO().setLanguage(language);
        try {
            ensureCaseVersion(problem);
            String content = request(config, problem,
                    promptFactory.multiLanguageProgramPrompt(config, problem, language, canonical.getCode()),
                    UUID.randomUUID().toString(), recordId, false, true);
            ensureCaseVersion(problem);
            JSONObject json = parseJson(content);
            String code = json.getStr("code", "").trim();
            if (code.isEmpty() || code.length() > 65535) {
                translated.setWarnings("AI 未生成有效的" + language + "源码");
            } else {
                translated.setCode(code).setAlgorithm(json.getStr("algorithm", ""))
                        .setWarnings(sanitizeWarnings(json.getStr("warnings", "")));
            }
        } catch (Exception e) {
            if (e instanceof InterruptedException) Thread.currentThread().interrupt();
            if (Thread.currentThread().isInterrupted()) throw e;
            translated.setWarnings("生成" + language + "源码失败：" + trim(e.getMessage()));
        }
        return translated;
    }

    private JSONObject parseJson(String content) throws StatusFailException {
        String value = content == null ? "" : content.trim();
        if (value.startsWith("```")) value = value.replaceFirst("^```(?:json)?\\s*", "").replaceFirst("\\s*```$", "");
        int start = value.indexOf('{'), end = value.lastIndexOf('}');
        if (start >= 0 && end > start) value = value.substring(start, end + 1);
        try {
            return JSONUtil.parseObj(value);
        } catch (Exception e) {
            throw new StatusFailException("AI 返回的 JSON 格式错误，请重试");
        }
    }

    @PreDestroy
    public void shutdownFullValidationPool() {
        FULL_VALIDATION_POOL.shutdownNow();
        PROGRAM_POOL.shutdownNow();
        COMPARISON_POOL.shutdownNow();
    }

    private String request(ProblemAiConfig config, Problem problem, String prompt,
                           String sessionId, Long currentRecordId, boolean includeHistory) throws Exception {
        return request(config, problem, prompt, sessionId, currentRecordId, includeHistory, false);
    }

    private String request(ProblemAiConfig config, Problem problem, String prompt,
                           String sessionId, Long currentRecordId, boolean includeHistory,
                           boolean generation) throws Exception {
        return request(config, problem, prompt, sessionId, currentRecordId, includeHistory, generation, null);
    }

    private String request(ProblemAiConfig config, Problem problem, String prompt,
                           String sessionId, Long currentRecordId, boolean includeHistory,
                           boolean generation, Integer timeoutSeconds) throws Exception {
        List<Map<String, String>> messages = new ArrayList<>();
        messages.add(message("system", generation
                ? promptFactory.programSystemPrompt(config.getSystemPrompt(), problem)
                : promptFactory.systemPrompt(config.getSystemPrompt(), problem)));
        if (includeHistory) {
            List<ProblemAiRecord> history = recordService.list(new QueryWrapper<ProblemAiRecord>()
                    .eq("pid", problem.getId()).eq("session_id", sessionId).eq("status", "success")
                    .ne("id", currentRecordId).orderByDesc("id").last("LIMIT 10"));
            Collections.reverse(history);
            for (ProblemAiRecord item : history) {
                messages.add(message("user", trim(item.getQuestion())));
                messages.add(message("assistant", trim(item.getResponse())));
            }
        }
        messages.add(message("user", prompt));
        return timeoutSeconds == null ? gateway.complete(config, messages)
                : gateway.completeOnce(config, messages, timeoutSeconds);
    }

    private Map<String, String> message(String role, String content) {
        Map<String, String> value = new HashMap<>(); value.put("role", role); value.put("content", content); return value;
    }
    private Problem requireProblem(Long pid) throws StatusFailException {
        Problem p = problemService.getById(pid); if (p == null) throw new StatusFailException("题目不存在"); return p;
    }
    private ProblemAiConfig requireConfig() throws StatusFailException {
        return problemAiConfigManager.requireConfig();
    }

    private ProblemAiGeneratedProgramVO parseProgram(String content, boolean integrated) throws StatusFailException {
        JSONObject json = parseJson(content);
        String code = json.getStr("code", "").trim();
        if (code.isEmpty() || code.length() > 65535) throw new StatusFailException("AI 未生成有效标准程序或源码过长");
        if (integrated) requireIntegratedProgram(code);
        JSONObject difficulty = validatedDifficulty(json.getJSONObject("difficulty"));
        String warnings = sanitizeWarnings(json.getStr("warnings", ""));
        if (difficulty == null) warnings = (warnings + "\n模型未返回有效的 Codeforces 难度估计").trim();
        return new ProblemAiGeneratedProgramVO().setLanguage(CANONICAL_STANDARD_LANGUAGE).setCode(code)
                .setAlgorithm(json.getStr("algorithm", "")).setWarnings(warnings).setDifficulty(difficulty);
    }

    private void requireIntegratedProgram(String code) throws StatusFailException {
        if (code == null || !code.matches("(?s).*#\\s*include\\s*[\"<]testlib\\.h[\">].*")
                || !code.matches("(?s).*\\bregisterValidation\\s*\\(.*")
                || !code.matches("(?s).*\\binf\\s*\\.\\s*readEof\\s*\\(.*")) {
            throw new StatusFailException("主模型 std 必须内置 testlib/registerValidation 输入校验并调用 inf.readEof()，不能使用分离校验器");
        }
    }

    private JSONObject validatedDifficulty(JSONObject value) {
        if (value == null) return null;
        try {
            Integer rating = value.getInt("rating"), low = value.getInt("rangeLow"), high = value.getInt("rangeHigh");
            String confidence = value.getStr("confidence", "").toLowerCase(Locale.ROOT);
            String reason = value.getStr("reason", "").trim();
            if (rating == null || low == null || high == null || low < 800 || high > 4000
                    || low > rating || high < rating || rating % 100 != 0
                    || !Arrays.asList("high", "medium", "low").contains(confidence) || reason.isEmpty()) return null;
            return new JSONObject().set("rating", rating).set("rangeLow", low).set("rangeHigh", high)
                    .set("confidence", confidence).set("reason", trim(reason));
        } catch (RuntimeException e) { return null; }
    }

    private String sanitizeWarnings(String value) {
        if (StringUtils.isEmpty(value)) return "";
        String text = value.trim();
        if (text.contains("题目未提供正式测试点") || text.contains("无法给出实测验证")
                || text.contains("没有正式测试点")) return "";
        return text;
    }

    private List<ProblemCase> problemCases(Long pid) {
        return caseService.list(new QueryWrapper<ProblemCase>().eq("pid", pid).eq("status", 0).orderByAsc("id"));
    }
    private String currentUid() { AccountProfile profile = (AccountProfile) SecurityUtils.getSubject().getPrincipal(); return profile == null ? "system" : profile.getUid(); }
    private String currentUsername() { AccountProfile profile = (AccountProfile) SecurityUtils.getSubject().getPrincipal(); return profile == null ? "system" : profile.getUsername(); }
    private void checkAdmin() throws StatusForbiddenException {
        org.apache.shiro.subject.Subject subject = SecurityUtils.getSubject();
        if (subject == null || !subject.isAuthenticated()
                || !(subject.hasRole("root") || subject.hasRole("admin") || subject.hasRole("problem_admin"))) {
            throw new StatusForbiddenException("请先登录并确认拥有题目管理员权限");
        }
    }
    private String trim(String value) { if (value == null) return ""; value = value.trim(); return value.length() > 12000 ? value.substring(0, 12000) : value; }
}
