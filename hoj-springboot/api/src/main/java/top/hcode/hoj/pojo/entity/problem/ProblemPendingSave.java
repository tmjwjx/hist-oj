package top.hcode.hoj.pojo.entity.problem;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;
import lombok.experimental.Accessors;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("problem_pending_save")
public class ProblemPendingSave {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String uid;
    private Long targetPid;
    private Long candidatePid;
    private String payload;
    private String requestFingerprint;
    private String baseFingerprint;
    private String candidateFingerprint;
    private Long publishedPid;
    private Boolean superseded;
    @TableField(fill = FieldFill.INSERT)
    private Date gmtCreate;
    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Date gmtModified;
}
