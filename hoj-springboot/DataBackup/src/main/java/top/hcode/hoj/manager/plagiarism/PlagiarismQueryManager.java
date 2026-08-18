package top.hcode.hoj.manager.plagiarism;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismResult;
import top.hcode.hoj.pojo.vo.PlagiarismResultPageVO;

import javax.annotation.Resource;
import java.nio.charset.StandardCharsets;
import java.util.List;

@Component
public class PlagiarismQueryManager {
    @Resource private PlagiarismManager manager;

    public PlagiarismResultPageVO results(Long checkId, String displayId)
            throws StatusNotFoundException, StatusForbiddenException {
        manager.check(checkId);
        QueryWrapper<PlagiarismResult> query = new QueryWrapper<PlagiarismResult>()
                .eq("check_id", checkId).eq("is_over_threshold", true);
        if (displayId != null && !displayId.trim().isEmpty()) {
            query.eq("display_id", displayId.trim());
        }
        List<PlagiarismResult> results = manager.results().list(query.orderByDesc("max_similarity"));
        return new PlagiarismResultPageVO().setResults(results).setTotalCount(results.size());
    }

    public byte[] export(Long checkId) throws StatusNotFoundException, StatusForbiddenException {
        manager.check(checkId);
        List<PlagiarismResult> results = manager.results().list(new QueryWrapper<PlagiarismResult>()
                .eq("check_id", checkId).orderByDesc("max_similarity"));
        StringBuilder csv = new StringBuilder("\uFEFF题目编号,题目标题,用户1,用户2,语言,相似度1to2(%),相似度2to1(%),最大相似度(%),超过阈值\n");
        for (PlagiarismResult result : results) {
            csv.append(row(result.getDisplayId(), result.getProblemTitle(), result.getUsername1(),
                    result.getUsername2(), result.getLanguage(), result.getSimilarity1to2(),
                    result.getSimilarity2to1(), result.getMaxSimilarity(),
                    Boolean.TRUE.equals(result.getIsOverThreshold()) ? "是" : "否"));
        }
        return csv.toString().getBytes(StandardCharsets.UTF_8);
    }

    public Judge submission(Long submitId)
            throws StatusNotFoundException, StatusForbiddenException {
        return manager.submission(submitId);
    }

    private String row(String displayId, String title, String user1, String user2, String language,
                       Integer first, Integer second, Integer max, String overThreshold) {
        return String.join(",", java.util.Arrays.asList(csv(displayId), csv(title), csv(user1), csv(user2),
                csv(language), String.valueOf(first == null ? 0 : first), String.valueOf(second == null ? 0 : second),
                String.valueOf(max == null ? 0 : max), csv(overThreshold))) + "\n";
    }

    private String csv(String value) {
        if (value == null || value.isEmpty()) return "";
        if (!value.matches(".*[\\\",\\n\\r].*")) return value;
        return "\"" + value.replace("\"", "\"\"") + "\"";
    }
}
