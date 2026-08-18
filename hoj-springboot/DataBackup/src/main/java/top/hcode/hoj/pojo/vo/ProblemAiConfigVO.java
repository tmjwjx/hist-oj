package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ProblemAiConfigVO {
    private Boolean enabled;
    private String apiUrl;
    private String apiKey;
    private Boolean apiKeyConfigured;
    private String model;
    private Integer timeoutSeconds;
    private String systemPrompt;
    private String validationPrompt;
}
