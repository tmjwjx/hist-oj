package top.hcode.hoj.common.exception;

import lombok.Getter;
import top.hcode.hoj.common.result.ResultStatus;

@Getter
public class LearningMapException extends RuntimeException {
    private final ResultStatus status;

    public LearningMapException(String message) {
        this(message, ResultStatus.FAIL);
    }

    public LearningMapException(String message, ResultStatus status) {
        super(message);
        this.status = status;
    }

    public static LearningMapException notFound(String message) {
        return new LearningMapException(message, ResultStatus.NOT_FOUND);
    }

    public static LearningMapException forbidden(String message) {
        return new LearningMapException(message, ResultStatus.FORBIDDEN);
    }
}
