package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONUtil;
import org.apache.shiro.SecurityUtils;
import org.apache.shiro.subject.Subject;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.problem.ProblemAiConfigEntityService;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemAiModelConfig;
import top.hcode.hoj.pojo.vo.ProblemAiConfigVO;

import javax.annotation.Resource;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ProblemAiConfigManager {
    private static final long CONFIG_ID = 1L;
    private static final String LEGACY_MODEL_ID = "00000000-0000-4000-8000-000000000001";
    private static final String MASKED_KEY = "***已配置***";
    private static final String DEFAULT_PROTOCOL = "openai_chat";
    private static final Set<String> PROTOCOLS = new HashSet<>(Arrays.asList(
            DEFAULT_PROTOCOL, "openai_responses", "gemini", "anthropic"));
    @Resource private ProblemAiConfigEntityService configService;
    @Resource private ProblemAiPromptFactory promptFactory;

    public ProblemAiConfigVO getConfig() throws StatusForbiddenException, StatusFailException {
        checkAdmin();
        ProblemAiConfig config = load();
        return new ProblemAiConfigVO().setEnabled(config.getEnabled()).setApiUrl(config.getApiUrl())
                .setApiKey("").setApiKeyConfigured(!blank(config.getApiKey())).setModel(config.getModel())
                .setProtocol(config.getProtocol())
                .setModels(config.getModels().stream().map(model -> copyModel(model).setApiKey("")
                        .setApiKeyConfigured(!blank(model.getApiKey()))).collect(Collectors.toList()))
                .setTimeoutSeconds(config.getTimeoutSeconds()).setSystemPrompt(config.getSystemPrompt())
                .setValidationPrompt(config.getValidationPrompt());
    }

    public void saveConfig(ProblemAiConfig input) throws StatusForbiddenException, StatusFailException {
        checkAdmin();
        if (input == null) throw new StatusFailException("AI 配置不能为空");
        ProblemAiConfig old = load();
        ProblemAiConfig config = copyConfig(input);
        List<ProblemAiModelConfig> models;
        if (input.getModels() == null) {
            // Keep old single-model clients usable without silently deleting additional models.
            if (old.getModels().size() > 1) throw new StatusFailException("请刷新配置页面后编辑多模型配置");
            models = Collections.singletonList(new ProblemAiModelConfig()
                    .setId(old.getModels().get(0).getId()).setName(old.getModels().get(0).getName())
                    .setApiUrl(input.getApiUrl()).setApiKey(input.getApiKey()).setProtocol(input.getProtocol())
                    .setModel(blank(input.getModel()) ? "gpt-4o-mini" : input.getModel()).setPrimary(true));
        } else {
            models = input.getModels().stream().map(model -> model == null ? null : copyModel(model))
                    .collect(Collectors.toList());
        }
        Map<String, ProblemAiModelConfig> previous = old.getModels().stream()
                .collect(Collectors.toMap(ProblemAiModelConfig::getId, model -> model));
        for (ProblemAiModelConfig model : models) {
            if (model == null) throw new StatusFailException("模型配置不能为空");
            ProblemAiModelConfig matching = previous.get(clean(model.getId()).toLowerCase(Locale.ROOT));
            // Older clients omit protocol; preserve native protocols by stable model ID.
            if (model.getProtocol() == null)
                model.setProtocol(matching == null ? DEFAULT_PROTOCOL : matching.getProtocol());
            String key = clean(model.getApiKey());
            if (key.isEmpty() || MASKED_KEY.equals(key)) {
                model.setApiKey(matching == null ? "" : matching.getApiKey());
            } else {
                model.setApiKey(key);
            }
            model.setApiKeyConfigured(null);
        }
        config.setId(CONFIG_ID).setEnabled(Boolean.TRUE.equals(input.getEnabled())).setModels(models);
        validateModels(config);
        applyDefaults(config);
        applyPrimary(config);
        config.setModelsJson(JSONUtil.toJsonStr(models));
        try {
            if (!configService.saveOrUpdate(config)) throw new StatusFailException("保存 AI 配置失败，请重试");
        } catch (RuntimeException e) {
            // Database exceptions may include the JSON and API keys in bound parameters.
            throw new StatusFailException("保存 AI 配置失败，请重试");
        }
    }

    public ProblemAiConfig requireConfig() throws StatusFailException {
        ProblemAiConfig config = load();
        if (!Boolean.TRUE.equals(config.getEnabled())) throw new StatusFailException("尚未配置可用的 AI 验题服务");
        validateModels(config);
        return config;
    }

    /** Missing selection preserves the old single-provider behavior. */
    public List<ProblemAiModelConfig> selectModels(ProblemAiConfig config, List<String> ids) throws StatusFailException {
        if (config == null || config.getModels() == null) throw new StatusFailException("AI 模型配置不存在");
        ProblemAiModelConfig primary = primary(config);
        if (ids == null) return Collections.singletonList(copyModel(primary));
        if (ids.isEmpty()) throw new StatusFailException("请至少选择主模型参与验题");
        Set<String> selected = new LinkedHashSet<>();
        Map<String, ProblemAiModelConfig> available = config.getModels().stream()
                .collect(Collectors.toMap(ProblemAiModelConfig::getId, model -> model));
        for (String id : ids) {
            if (!available.containsKey(id)) throw new StatusFailException("所选 AI 模型不存在，请刷新配置");
            if (!selected.add(id)) throw new StatusFailException("所选 AI 模型不能重复");
        }
        if (!selected.contains(primary.getId())) throw new StatusFailException("验题必须包含主模型");
        return config.getModels().stream().filter(model -> selected.contains(model.getId()))
                .map(this::copyModel).collect(Collectors.toList());
    }

    public ProblemAiConfig forModel(ProblemAiConfig config, ProblemAiModelConfig model) {
        return copyConfig(config).setApiUrl(model.getApiUrl()).setApiKey(model.getApiKey()).setModel(model.getModel())
                .setProtocol(model.getProtocol() == null ? DEFAULT_PROTOCOL : model.getProtocol());
    }

    private ProblemAiConfig load() throws StatusFailException {
        ProblemAiConfig stored;
        try {
            stored = configService.getById(CONFIG_ID);
        } catch (RuntimeException e) {
            throw new StatusFailException("读取 AI 配置失败，请重试");
        }
        ProblemAiConfig config = stored == null ? new ProblemAiConfig().setEnabled(false) : copyConfig(stored);
        if (blank(config.getModelsJson())) {
            config.setModels(Collections.singletonList(new ProblemAiModelConfig().setId(LEGACY_MODEL_ID)
                    .setName("默认模型").setApiUrl(clean(config.getApiUrl())).setApiKey(clean(config.getApiKey()))
                    .setModel(blank(config.getModel()) ? "gpt-4o-mini" : clean(config.getModel())).setPrimary(true)));
        } else {
            try {
                String json = config.getModelsJson().trim();
                if (!json.startsWith("[")) throw new IllegalArgumentException();
                config.setModels(JSONUtil.toList(JSONUtil.parseArray(json), ProblemAiModelConfig.class));
            } catch (RuntimeException e) {
                // Never attach the JSON or parser error: either may contain credentials.
                throw new StatusFailException("AI 模型配置格式错误，请联系管理员修复");
            }
        }
        // A disabled fresh configuration is an editable draft; no writes occur here.
        validateModels(config);
        applyDefaults(config);
        applyPrimary(config);
        return config;
    }

    private void validateModels(ProblemAiConfig config) throws StatusFailException {
        List<ProblemAiModelConfig> models = config.getModels();
        if (models == null || models.isEmpty()) throw new StatusFailException("至少配置一个 AI 模型");
        Set<String> ids = new HashSet<>();
        int primaries = 0;
        for (ProblemAiModelConfig model : models) {
            if (model == null) throw new StatusFailException("模型配置不能为空");
            String id = clean(model.getId()).toLowerCase(Locale.ROOT);
            try {
                if (!UUID.fromString(id).toString().equals(id)) throw new IllegalArgumentException();
            } catch (IllegalArgumentException e) {
                throw new StatusFailException("模型 ID 必须为 UUID");
            }
            if (!ids.add(id)) throw new StatusFailException("模型 ID 不能重复");
            model.setId(id).setName(clean(model.getName())).setApiUrl(clean(model.getApiUrl()))
                    .setModel(clean(model.getModel())).setApiKey(clean(model.getApiKey()))
                    .setProtocol(model.getProtocol() == null ? DEFAULT_PROTOCOL : clean(model.getProtocol()))
                    .setPrimary(Boolean.TRUE.equals(model.getPrimary()));
            if (!PROTOCOLS.contains(model.getProtocol()))
                throw new StatusFailException("接口协议仅支持 OpenAI Chat Completions、OpenAI Responses、Gemini 或 Anthropic");
            if (model.getName().isEmpty() || model.getName().length() > 100)
                throw new StatusFailException("模型名称不能为空且不能超过 100 个字符");
            if (model.getModel().isEmpty() || model.getModel().length() > 100)
                throw new StatusFailException("模型标识不能为空且不能超过 100 个字符");
            if (model.getApiKey().length() > 500 || model.getApiKey().chars().anyMatch(Character::isISOControl))
                throw new StatusFailException("模型 API Key 不能超过 500 个字符或包含控制字符");
            if (!model.getApiUrl().isEmpty()) validateUrl(model.getApiUrl());
            if (Boolean.TRUE.equals(config.getEnabled()) && (model.getApiUrl().isEmpty() || model.getApiKey().isEmpty()))
                throw new StatusFailException("启用 AI 时每个模型都必须配置接口地址和 API Key");
            if (model.getPrimary()) primaries++;
        }
        if (primaries != 1) throw new StatusFailException("必须且只能指定一个主模型");
        Integer timeout = config.getTimeoutSeconds();
        if (timeout != null && (timeout < 1 || timeout > 3600)) throw new StatusFailException("AI 超时必须为 1 至 3600 秒");
    }

    private void validateUrl(String value) throws StatusFailException {
        try {
            URI uri = new URI(value);
            if (value.length() > 500 || !("http".equalsIgnoreCase(uri.getScheme()) || "https".equalsIgnoreCase(uri.getScheme()))
                    || uri.getHost() == null || uri.getRawUserInfo() != null || uri.getFragment() != null
                    || uri.getPort() < -1 || uri.getPort() > 65535) throw new URISyntaxException("", "invalid");
        } catch (URISyntaxException e) {
            throw new StatusFailException("模型接口地址必须为有效的 HTTP 或 HTTPS URL，且不能包含用户凭据或片段");
        }
    }

    private void applyDefaults(ProblemAiConfig config) {
        if (config.getTimeoutSeconds() == null) config.setTimeoutSeconds(120);
        if (blank(config.getSystemPrompt()) || config.getSystemPrompt().trim().length() < 100)
            config.setSystemPrompt(promptFactory.defaultSystemPrompt());
        if (blank(config.getValidationPrompt()) || config.getValidationPrompt().trim().length() < 100)
            config.setValidationPrompt(promptFactory.defaultValidationPrompt());
    }

    private ProblemAiModelConfig primary(ProblemAiConfig config) throws StatusFailException {
        return config.getModels().stream().filter(model -> Boolean.TRUE.equals(model.getPrimary())).findFirst()
                .orElseThrow(() -> new StatusFailException("请先指定主模型"));
    }

    private void applyPrimary(ProblemAiConfig config) throws StatusFailException {
        ProblemAiModelConfig model = primary(config);
        config.setApiUrl(model.getApiUrl()).setApiKey(model.getApiKey()).setModel(model.getModel())
                .setProtocol(model.getProtocol());
    }

    private ProblemAiConfig copyConfig(ProblemAiConfig source) {
        ProblemAiConfig copy = new ProblemAiConfig();
        BeanUtils.copyProperties(source, copy);
        if (source.getModels() != null) copy.setModels(source.getModels().stream()
                .map(model -> model == null ? null : copyModel(model)).collect(Collectors.toList()));
        return copy;
    }

    private ProblemAiModelConfig copyModel(ProblemAiModelConfig source) {
        ProblemAiModelConfig copy = new ProblemAiModelConfig();
        BeanUtils.copyProperties(source, copy);
        return copy;
    }

    private void checkAdmin() throws StatusForbiddenException {
        Subject subject = SecurityUtils.getSubject();
        if (!subject.isAuthenticated() || !(subject.hasRole("root") || subject.hasRole("admin") || subject.hasRole("problem_admin")))
            throw new StatusForbiddenException("请先登录并确认拥有题目管理员权限");
    }

    private boolean blank(String value) { return value == null || value.trim().isEmpty(); }
    private String clean(String value) { return value == null ? "" : value.trim(); }
}
