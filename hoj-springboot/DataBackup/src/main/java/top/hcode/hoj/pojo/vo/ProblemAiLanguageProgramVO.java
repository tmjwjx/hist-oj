package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

/**
 * An additional implementation generated for cross-language compatibility
 * checking.  The C++17 implementation remains the canonical standard program;
 * these implementations are only executed by the AI validation pipeline.
 */
@Data
@Accessors(chain = true)
public class ProblemAiLanguageProgramVO {
    private String language;
    private String code;
    private String algorithm;
    private String warnings;
}
