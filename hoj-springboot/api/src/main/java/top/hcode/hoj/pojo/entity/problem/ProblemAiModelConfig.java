package top.hcode.hoj.pojo.entity.problem;

import lombok.Data;
import lombok.ToString;
import lombok.experimental.Accessors;

import java.io.Serializable;

@Data
@ToString(exclude = "apiKey")
@Accessors(chain = true)
public class ProblemAiModelConfig implements Serializable {
    private String id;
    private String name;
    private String apiUrl;
    private String apiKey;
    private Boolean apiKeyConfigured;
    private String model;
    private String protocol;
    private Boolean primary;
}
