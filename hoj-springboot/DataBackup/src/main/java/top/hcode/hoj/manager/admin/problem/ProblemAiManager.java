package top.hcode.hoj.manager.admin.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.problem.ProblemAiConfigEntityService;
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
import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;

import javax.annotation.Resource;
import java.util.*;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Component
public class ProblemAiManager {
    private static final long CONFIG_ID = 1L;
    private static final int DEFAULT_AI_TIMEOUT_SECONDS = 120;
    private static final int MAX_AI_TIMEOUT_SECONDS = 3600;
    /** AI 验题的权威标准程序固定使用判题机支持的 C++17。 */
    private static final String CANONICAL_STANDARD_LANGUAGE = "C++ 17";
    private static final long JUDGE_WAIT_TIMEOUT_MS = 20 * 60 * 1000L;
    private static final ExecutorService FULL_VALIDATION_POOL = Executors.newFixedThreadPool(2, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-full-validation");
        thread.setDaemon(false);
        return thread;
    });
    @Resource private ProblemAiConfigEntityService configService;
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

    public ProblemAiConfigVO getConfig() throws StatusForbiddenException {
        checkAdmin();
        ProblemAiConfig config = configService.getById(CONFIG_ID);
        if (config == null) config = defaults();
        applyPromptDefaults(config);
        return toVO(config);
    }

    public void saveConfig(ProblemAiConfig input) throws StatusFailException, StatusForbiddenException {
        checkAdmin();
        if (input == null || (Boolean.TRUE.equals(input.getEnabled()) && StringUtils.isEmpty(input.getApiUrl()))) throw new StatusFailException("启用 AI 时接口地址不能为空");
        ProblemAiConfig old = configService.getById(CONFIG_ID);
        if (old != null && (input.getApiKey() == null || input.getApiKey().startsWith("***"))) input.setApiKey(old.getApiKey());
        if (input.getApiUrl() == null) input.setApiUrl("");
        if (StringUtils.isEmpty(input.getModel())) input.setModel("gpt-4o-mini");
        applyPromptDefaults(input);
        input.setId(CONFIG_ID).setEnabled(!Boolean.FALSE.equals(input.getEnabled()))
                .setTimeoutSeconds(input.getTimeoutSeconds() == null ? DEFAULT_AI_TIMEOUT_SECONDS
                        : Math.max(1, Math.min(input.getTimeoutSeconds(), MAX_AI_TIMEOUT_SECONDS)));
        configService.saveOrUpdate(input);
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
        if (StringUtils.isEmpty(dto.getValidatorCode())) {
            ProblemAiGeneratedProgramVO generated = new ProblemAiGeneratedProgramVO()
                    .setLanguage(dto.getLanguage()).setCode(dto.getStandardProgram());
            String validatorContent = request(config, problem,
                    promptFactory.validatorPrompt(config, problem, dto.getLanguage(), dto.getStandardProgram()),
                    session, null, false, true);
            applyValidator(generated, validatorContent);
            dto.setValidatorLanguage(generated.getValidatorLanguage());
            dto.setValidatorCode(generated.getValidatorCode());
        } else {
            // Also repair validators supplied by an older generation record.
            dto.setValidatorCode(normalizeValidatorCode(dto.getValidatorCode()));
        }
        ProblemAiRecord record = new ProblemAiRecord().setPid(dto.getPid()).setSessionId(session)
                .setUid(currentUid()).setQuestion("AI 一键验题").setStatus("running").setGmtCreate(new Date());
        recordService.save(record);
        long started = System.currentTimeMillis();
        try {
            String content = validationRunner.run(config, problem, problemCases(problem.getId()), dto);
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
        if (dto == null || dto.getPid() == null || StringUtils.isEmpty(dto.getLanguage())) {
            throw new StatusFailException("题目和标准程序语言不能为空");
        }
        Problem problem = requireProblem(dto.getPid());
        ProblemAiConfig config = requireConfig();
        String session = UUID.randomUUID().toString();
        ProblemAiRecord record = new ProblemAiRecord().setPid(dto.getPid()).setSessionId(session)
                .setUid(currentUid()).setQuestion("AI 生成标准程序").setStatus("running").setGmtCreate(new Date());
        recordService.save(record);
        long started = System.currentTimeMillis();
        try {
            String prompt = promptFactory.standardProgramPrompt(config, problem, CANONICAL_STANDARD_LANGUAGE, 0);
            String content = request(config, problem, prompt, session, record.getId(), false, true);
            ProblemAiGeneratedProgramVO program = parseProgram(content, CANONICAL_STANDARD_LANGUAGE).setRecordId(record.getId());
            if (StringUtils.isEmpty(program.getValidatorCode())) {
                String validatorContent = request(config, problem,
                        // The validator is compiled by the C++ testlib toolchain
                        // regardless of the UI's language selector.  Keeping it
                        // on the canonical C++17 path prevents a Java/PyPy3
                        // selection from producing a validator that the
                        // strict testlib runner cannot compile.
                        promptFactory.validatorPrompt(config, problem, CANONICAL_STANDARD_LANGUAGE, program.getCode()),
                        session, record.getId(), false, true);
                applyValidator(program, validatorContent);
            }
            record.setResponse(JSONUtil.toJsonStr(program)).setStatus("success")
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);

            // 生成成功后立即创建后台全流程任务。后续正式判题、等待结果和 AI
            // 综合分析均由服务端完成，不再依赖前端弹窗或页面轮询是否存在。
            ProblemAiRecord validationRecord = new ProblemAiRecord().setPid(problem.getId())
                    .setSessionId(UUID.randomUUID().toString()).setUid(currentUid())
                    .setQuestion("AI 一键验题").setStatus("running").setGmtCreate(new Date());
            recordService.save(validationRecord);
            program.setValidationRecordId(validationRecord.getId());
            submitFullValidationTask(config, problem, program, validationRecord,
                    currentUid(), currentUsername());
            return program;
        } catch (Exception e) {
            record.setStatus("failed").setErrorMessage(trim(e.getMessage()))
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            throw e;
        }
    }

    private void submitFullValidationTask(ProblemAiConfig config, Problem problem,
                                           ProblemAiGeneratedProgramVO program,
                                           ProblemAiRecord validationRecord,
                                           String uid, String username) {
        try {
            FULL_VALIDATION_POOL.submit(() -> runFullValidation(
                    config, problem, program, validationRecord, uid, username));
        } catch (RuntimeException e) {
            validationRecord.setStatus("failed").setErrorMessage(trim(e.getMessage()));
            recordService.updateById(validationRecord);
        }
    }

    private void runFullValidation(ProblemAiConfig config, Problem problem,
                                   ProblemAiGeneratedProgramVO program,
                                   ProblemAiRecord validationRecord,
                                   String uid, String username) {
        long started = System.currentTimeMillis();
        try {
            // C++17 is the canonical submission. Generate the two portability
            // implementations in the background so closing the browser cannot
            // interrupt their generation or the subsequent test runs.
            // Keep the submitted C++17 source in the same result set as the
            // translated Java/PyPy3 ports.  This makes the UI's language
            // summary an actual three-language all-test-point report rather
            // than a report that omits the canonical submission.
            List<ProblemAiLanguageProgramVO> multiLanguagePrograms = new ArrayList<>();
            multiLanguagePrograms.add(new ProblemAiLanguageProgramVO()
                    .setLanguage(CANONICAL_STANDARD_LANGUAGE)
                    .setCode(program.getCode())
                    .setAlgorithm(program.getAlgorithm())
                    .setWarnings(program.getWarnings()));
            multiLanguagePrograms.addAll(generateMultiLanguagePrograms(
                    config, problem, program, validationRecord.getId()));
            program.setMultiLanguagePrograms(multiLanguagePrograms);
            recordService.updateById(new ProblemAiRecord().setId(program.getRecordId())
                    .setResponse(JSONUtil.toJsonStr(program)));
            // 如果该题刚刚有其它标准程序在判题，先等它结束，避免覆盖正在执行的验题提交。
            verificationManager.waitForAiJudge(problem.getId(), JUDGE_WAIT_TIMEOUT_MS);
            ProblemVerificationSubmitDTO submit = new ProblemVerificationSubmitDTO();
            submit.setPid(problem.getId());
            submit.setLanguage(program.getLanguage());
            submit.setCode(program.getCode());
            submit.setVerificationType("ai_validation");
            Long submitId = verificationManager.submitForAi(submit, uid, username);
            verificationManager.waitForAiJudge(submit.getPid(), submitId, JUDGE_WAIT_TIMEOUT_MS);

            ProblemAiValidateDTO input = new ProblemAiValidateDTO();
            input.setPid(problem.getId());
            input.setLanguage(program.getLanguage());
            input.setStandardProgram(program.getCode());
            input.setAiGenerated(true);
            input.setAlgorithmSummary(program.getAlgorithm());
            input.setValidatorLanguage(program.getValidatorLanguage());
            input.setValidatorCode(program.getValidatorCode());
            input.setMultiLanguagePrograms(multiLanguagePrograms);
            String content = validationRunner.run(config, problem, problemCases(problem.getId()), input);
            validationRecord.setResponse(content).setStatus("success")
                    .setDurationMs((int) (System.currentTimeMillis() - started));
        } catch (Exception e) {
            validationRecord.setStatus("failed").setErrorMessage(trim(e.getMessage()))
                    .setDurationMs((int) (System.currentTimeMillis() - started));
        }
        recordService.updateById(validationRecord);
    }

    private List<ProblemAiLanguageProgramVO> generateMultiLanguagePrograms(ProblemAiConfig config,
                                                                            Problem problem,
                                                                            ProblemAiGeneratedProgramVO canonical,
                                                                            Long recordId) {
        List<ProblemAiLanguageProgramVO> result = new ArrayList<>();
        for (String language : Arrays.asList("Java", "PyPy3")) {
            ProblemAiLanguageProgramVO translated = new ProblemAiLanguageProgramVO().setLanguage(language);
            try {
                String content = request(config, problem,
                        promptFactory.multiLanguageProgramPrompt(config, problem, language, canonical.getCode()),
                        UUID.randomUUID().toString(), recordId, false, true);
                JSONObject json = parseJson(content);
                String code = json.getStr("code", "").trim();
                if (code.isEmpty() || code.length() > 65535) {
                    translated.setWarnings("AI 未生成有效的" + language + "源码");
                } else {
                    translated.setCode(code).setAlgorithm(json.getStr("algorithm", ""))
                            .setWarnings(sanitizeWarnings(json.getStr("warnings", "")));
                }
            } catch (Exception e) {
                translated.setWarnings("生成" + language + "源码失败：" + trim(e.getMessage()));
            }
            result.add(translated);
        }
        return result;
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
        FULL_VALIDATION_POOL.shutdown();
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
                ? promptFactory.programSystemPrompt(config.getSystemPrompt())
                : promptFactory.systemPrompt(config.getSystemPrompt())));
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
        ProblemAiConfig config = configService.getById(CONFIG_ID);
        if (config == null || !Boolean.TRUE.equals(config.getEnabled()) || StringUtils.isEmpty(config.getApiKey()))
            throw new StatusFailException("尚未配置可用的 AI 验题服务");
        applyPromptDefaults(config);
        return config;
    }
    private ProblemAiGeneratedProgramVO parseProgram(String content, String requestedLanguage) throws StatusFailException {
        String value = content == null ? "" : content.trim();
        if (value.startsWith("```")) value = value.replaceFirst("^```(?:json)?\\s*", "").replaceFirst("\\s*```$", "");
        int start = value.indexOf('{'), end = value.lastIndexOf('}');
        if (start >= 0 && end > start) value = value.substring(start, end + 1);
        try {
            JSONObject json = JSONUtil.parseObj(value);
            String code = json.getStr("code", "").trim();
            if (code.isEmpty() || code.length() > 65535) throw new StatusFailException("AI 未生成有效标准程序或源码过长");
            ProblemAiGeneratedProgramVO program = new ProblemAiGeneratedProgramVO().setLanguage(requestedLanguage).setCode(code)
                    .setAlgorithm(json.getStr("algorithm", "")).setWarnings(sanitizeWarnings(json.getStr("warnings", "")));
            JSONObject validator = json.getJSONObject("validator");
            if (validator != null) {
                String validatorCode = normalizeValidatorCode(validator.getStr("code", "").trim());
                if (!StringUtils.isEmpty(validatorCode)
                        && (validatorCode.length() > 65535 || !validatorCode.contains("registerValidation"))) {
                    throw new StatusFailException("AI 生成的 testlib 输入校验器源码无效或过长");
                }
                program.setValidatorLanguage(validator.getStr("language", "C++"))
                        .setValidatorCode(validatorCode)
                        .setValidatorAlgorithm(validator.getStr("algorithm", ""))
                        .setValidatorWarnings(sanitizeWarnings(validator.getStr("warnings", "")));
            }
            if (StringUtils.isEmpty(program.getValidatorCode())) {
                String validatorCode = normalizeValidatorCode(json.getStr("validatorCode", "").trim());
                if (!StringUtils.isEmpty(validatorCode)) {
                    program.setValidatorLanguage(json.getStr("validatorLanguage", "C++"))
                            .setValidatorCode(validatorCode)
                            .setValidatorAlgorithm(json.getStr("validatorAlgorithm", ""))
                            .setValidatorWarnings(sanitizeWarnings(json.getStr("validatorWarnings", "")));
                }
            }
            return program;
        } catch (StatusFailException e) {
            throw e;
        } catch (Exception e) {
            throw new StatusFailException("AI 标准程序返回格式错误，请重试");
        }
    }

    private void applyValidator(ProblemAiGeneratedProgramVO program, String content) throws StatusFailException {
        String value = content == null ? "" : content.trim();
        if (value.startsWith("```")) value = value.replaceFirst("^```(?:json)?\\s*", "").replaceFirst("\\s*```$", "");
        int start = value.indexOf('{'), end = value.lastIndexOf('}');
        if (start >= 0 && end > start) value = value.substring(start, end + 1);
        try {
            JSONObject json = JSONUtil.parseObj(value);
            String code = normalizeValidatorCode(json.getStr("code", "").trim());
            if (code.length() > 65535 || !code.contains("registerValidation")) {
                throw new StatusFailException("AI 未生成有效的 testlib 输入校验器");
            }
            program.setValidatorLanguage(json.getStr("language", "C++"))
                    .setValidatorCode(code)
                    .setValidatorAlgorithm(json.getStr("algorithm", ""))
                    .setValidatorWarnings(sanitizeWarnings(json.getStr("warnings", "")));
        } catch (StatusFailException e) {
            throw e;
        } catch (Exception e) {
            throw new StatusFailException("AI testlib 输入校验器返回格式错误，请重试");
        }
    }

    private String sanitizeWarnings(String value) {
        if (StringUtils.isEmpty(value)) return "";
        String text = value.trim();
        if (text.contains("题目未提供正式测试点") || text.contains("无法给出实测验证")
                || text.contains("没有正式测试点")) return "";
        return text;
    }

    /**
     * Normalize common model mistakes in validators before they reach the
     * judge. In particular, readToken(pattern, name) treats its first
     * argument as a regular expression, and validators must return normally
     * on success instead of calling quitf(_ok, ...).
     */
    private String normalizeValidatorCode(String code) {
        if (StringUtils.isEmpty(code)) return code;
        String normalized = code;
        for (String name : new String[]{"row", "line", "token", "field", "value", "str", "s"}) {
            normalized = normalized.replace("inf.readToken(\"" + name + "\")", "inf.readToken()")
                    .replace("inf.readToken( '" + name + "' )", "inf.readToken()");
        }
        // Common generated form for a first-line `n m`: readInt skips
        // whitespace by itself, so make the required separator explicit.
        normalized = normalized.replaceAll(
                "(inf\\.readInt\\([^;\\n]*\\\"n\\\"\\)\\s*;\\s*)(int\\s+m\\s*=\\s*inf\\.readInt\\([^;\\n]*\\\"m\\\"\\)\\s*;)",
                "$1inf.readSpace();\\n    $2");
        // In validation mode testlib only accepts quitf(_fail, ...). A
        // successful validator must finish normally after readEof(); using
        // quitf(_ok, ...) is converted by testlib to FAIL (exit code 3).
        normalized = normalized.replaceAll(
                "(?m)^([\\t ]*)quitf\\s*\\(\\s*_ok\\s*,[^\\r\\n]*\\);\\s*$",
                "$1return 0;");
        normalized = normalized.replaceAll(
                "quitf\\s*\\(\\s*_ok\\s*,\\s*\\\"(?:\\\\.|[^\\\"\\\\])*\\\"\\s*\\)\\s*;",
                "return 0;");
        return normalized;
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
    private ProblemAiConfig defaults() { return new ProblemAiConfig().setEnabled(false).setModel("gpt-4o-mini").setTimeoutSeconds(120).setSystemPrompt(promptFactory.defaultSystemPrompt()).setValidationPrompt(promptFactory.defaultValidationPrompt()); }
    private void applyPromptDefaults(ProblemAiConfig config) {
        if (StringUtils.isEmpty(config.getSystemPrompt()) || config.getSystemPrompt().trim().length() < 100)
            config.setSystemPrompt(promptFactory.defaultSystemPrompt());
        if (StringUtils.isEmpty(config.getValidationPrompt()) || config.getValidationPrompt().trim().length() < 100)
            config.setValidationPrompt(promptFactory.defaultValidationPrompt());
    }
    private ProblemAiConfigVO toVO(ProblemAiConfig c) { return new ProblemAiConfigVO().setEnabled(c.getEnabled()).setApiUrl(c.getApiUrl()).setApiKey(c.getApiKey() == null ? "" : "***已配置***").setApiKeyConfigured(!StringUtils.isEmpty(c.getApiKey())).setModel(c.getModel()).setTimeoutSeconds(c.getTimeoutSeconds()).setSystemPrompt(c.getSystemPrompt()).setValidationPrompt(c.getValidationPrompt()); }
    private String trim(String value) { if (value == null) return ""; value = value.trim(); return value.length() > 12000 ? value.substring(0, 12000) : value; }
}
