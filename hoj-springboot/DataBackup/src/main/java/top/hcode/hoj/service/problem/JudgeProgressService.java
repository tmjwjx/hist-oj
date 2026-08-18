package top.hcode.hoj.service.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import lombok.AllArgsConstructor;
import lombok.Getter;
import org.springframework.stereotype.Component;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.utils.Constants;

import javax.annotation.Resource;
import java.util.List;
import java.util.Objects;

@Component
public class JudgeProgressService {
    @Resource private JudgeCaseEntityService judgeCaseService;

    public Progress resolve(Long submitId, Integer status, String judgeMessage) {
        List<JudgeCase> cases = judgeCaseService.list(new QueryWrapper<JudgeCase>()
                .select("seq", "status", "user_output")
                .eq("submit_id", submitId).orderByAsc("seq"));
        if (isRunning(status)) {
            JudgeCase current = cases.stream().filter(this::isPendingCase).findFirst().orElse(null);
            int test = current == null ? Math.max(1, cases.size() + 1) : current.getSeq();
            return new Progress("Running on test " + test, test, "正在评测第 " + test + " 个测试点");
        }
        JudgeCase failed = cases.stream().filter(this::isFailedCase).findFirst().orElse(null);
        if (failed != null) {
            String text = statusName(failed.getStatus()) + " on test " + failed.getSeq();
            return new Progress(text, failed.getSeq(), firstNonBlank(judgeMessage, failed.getUserOutput(), text));
        }
        String text = statusName(status);
        return new Progress(text, null, firstNonBlank(judgeMessage, text));
    }

    private boolean isRunning(Integer status) {
        return Objects.equals(status, Constants.Judge.STATUS_JUDGING.getStatus());
    }

    private boolean isPendingCase(JudgeCase item) {
        return Objects.equals(item.getStatus(), Constants.Judge.STATUS_PENDING.getStatus())
                || Objects.equals(item.getStatus(), Constants.Judge.STATUS_COMPILING.getStatus())
                || Objects.equals(item.getStatus(), Constants.Judge.STATUS_JUDGING.getStatus());
    }

    private boolean isFailedCase(JudgeCase item) {
        return item.getStatus() != null
                && !Objects.equals(item.getStatus(), Constants.Judge.STATUS_ACCEPTED.getStatus())
                && !Objects.equals(item.getStatus(), Constants.Judge.STATUS_CANCELLED.getStatus())
                && !isPendingCase(item);
    }

    private String statusName(Integer status) {
        for (Constants.Judge value : Constants.Judge.values()) {
            if (Objects.equals(value.getStatus(), status)) {
                String name = value.getName();
                return name.length() < 2 ? name : name.substring(0, 1) + name.substring(1).toLowerCase();
            }
        }
        return "Unknown";
    }

    private String firstNonBlank(String... values) {
        for (String value : values) if (value != null && !value.trim().isEmpty()) return value.trim();
        return "Unknown";
    }

    @Getter
    @AllArgsConstructor
    public static class Progress {
        private final String text;
        private final Integer currentTest;
        private final String message;
    }
}
