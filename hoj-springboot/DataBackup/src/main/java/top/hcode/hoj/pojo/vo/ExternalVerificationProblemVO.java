package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import java.util.List;

/** Explicit public projection: never serialize the Problem or Language entities. */
@Data
@Accessors(chain = true)
public class ExternalVerificationProblemVO {
    private Statement problem;
    private List<AllowedLanguage> languages;

    @Data
    @Accessors(chain = true)
    public static class Statement {
        private Long id;
        private String problemId;
        private String title;
        private String description;
        private String input;
        private String output;
        private String examples;
        private String hint;
        private Integer timeLimit;
        private Integer memoryLimit;
        private Integer stackLimit;
        private Integer type;
        private String judgeMode;
        private Boolean isFileIO;
        private String ioReadFileName;
        private String ioWriteFileName;
    }

    @Data
    @Accessors(chain = true)
    public static class AllowedLanguage {
        private String name;
        private String contentType;
    }
}
