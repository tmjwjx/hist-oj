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
@TableName("plagiarism_check_config")
public class PlagiarismCheckConfig implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long cid;
    private Long pid;
    private Long cpid;
    private Integer threshold;
    private String createdBy;
    private Date gmtCreate;
    private Date gmtModified;
}
