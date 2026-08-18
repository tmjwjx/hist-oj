package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismResult;

import java.util.List;

@Data
@Accessors(chain = true)
public class PlagiarismResultPageVO {
    private List<PlagiarismResult> results;
    private long totalCount;
}
