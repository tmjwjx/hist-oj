package top.hcode.hoj.pojo.entity.plagiarism;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("plagiarism_check")
public class PlagiarismCheck implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long cid;
    private String status;
    private Integer totalPairs;
    private Integer checkedPairs;
    private Integer totalSubmissions;
    private Double progress;
    private String errorMessage;
    private String startedBy;
    private Date gmtCreate;
    private Date gmtModified;
    private Date startedAt;
    private Date completedAt;
}
