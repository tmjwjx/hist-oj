package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.entity.problem.ProblemAiModelConfig;

import java.util.List;

@Data
@Accessors(chain = true)
public class ProblemAiConfigVO {
    private Boolean enabled;
    private String apiUrl;
    private String apiKey;
    private Boolean apiKeyConfigured;
    private String model;
    private String protocol;
    private List<ProblemAiModelConfig> models;
    private Integer timeoutSeconds;
    private String systemPrompt;
    private String validationPrompt;
}
