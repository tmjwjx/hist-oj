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
import top.hcode.hoj.pojo.dto.ProblemAiChatDTO;
import top.hcode.hoj.pojo.dto.ProblemAiGenerateProgramDTO;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ProblemAiConfigVO;
import top.hcode.hoj.pojo.vo.ProblemAiGeneratedProgramVO;
import top.hcode.hoj.shiro.AccountProfile;

import javax.annotation.Resource;
import java.util.*;

@Component
public class ProblemAiManager {
    private static final long CONFIG_ID = 1L;
    @Resource private ProblemAiConfigEntityService configService;
    @Resource private ProblemAiRecordEntityService recordService;
    @Resource private ProblemEntityService problemService;
    @Resource private ProblemCaseEntityService caseService;
    @Resource private ProblemAiPromptFactory promptFactory;
    @Resource private ProblemAiGateway gateway;
    @Resource private ProblemAiValidationRunner validationRunner;

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
                .setTimeoutSeconds(input.getTimeoutSeconds() == null ? 120 : Math.max(1, Math.min(input.getTimeoutSeconds(), 900)));
        configService.saveOrUpdate(input);
    }

    public List<ProblemAiRecord> records(Long pid) throws StatusFailException, StatusForbiddenException {
        checkAdmin();
        requireProblem(pid);
        return recordService.list(new QueryWrapper<ProblemAiRecord>().eq("pid", pid).orderByAsc("gmt_create"));
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
            String prompt = promptFactory.standardProgramPrompt(config, problem, dto.getLanguage(), 0);
            String content = request(config, problem, prompt, session, record.getId(), false, true);
            ProblemAiGeneratedProgramVO program = parseProgram(content, dto.getLanguage()).setRecordId(record.getId());
            record.setResponse(JSONUtil.toJsonStr(program)).setStatus("success")
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            return program;
        } catch (Exception e) {
            record.setStatus("failed").setErrorMessage(trim(e.getMessage()))
                    .setDurationMs((int) (System.currentTimeMillis() - started));
            recordService.updateById(record);
            throw e;
        }
    }

    private String request(ProblemAiConfig config, Problem problem, String prompt,
                           String sessionId, Long currentRecordId, boolean includeHistory) throws Exception {
        return request(config, problem, prompt, sessionId, currentRecordId, includeHistory, false);
    }

    private String request(ProblemAiConfig config, Problem problem, String prompt,
                           String sessionId, Long currentRecordId, boolean includeHistory,
                           boolean generation) throws Exception {
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
        return gateway.complete(config, messages);
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
            return new ProblemAiGeneratedProgramVO().setLanguage(requestedLanguage).setCode(code)
                    .setAlgorithm(json.getStr("algorithm", "")).setWarnings(json.getStr("warnings", ""));
        } catch (StatusFailException e) {
            throw e;
        } catch (Exception e) {
            throw new StatusFailException("AI 标准程序返回格式错误，请重试");
        }
    }
    private List<ProblemCase> problemCases(Long pid) {
        return caseService.list(new QueryWrapper<ProblemCase>().eq("pid", pid).eq("status", 0).orderByAsc("id"));
    }
    private String currentUid() { AccountProfile profile = (AccountProfile) SecurityUtils.getSubject().getPrincipal(); return profile == null ? "system" : profile.getUid(); }
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
