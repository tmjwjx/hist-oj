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
@TableName("plagiarism_result")
public class PlagiarismResult implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long checkId;
    private Long cid;
    private Long cpid;
    private Long pid;
    private String displayId;
    private String problemTitle;
    private Long submitId1;
    private Long submitId2;
    private Long contestRecordId1;
    private Long contestRecordId2;
    private String uid1;
    private String uid2;
    private String username1;
    private String username2;
    private String language;
    private Integer similarity1to2;
    private Integer similarity2to1;
    private Integer maxSimilarity;
    private Boolean isOverThreshold;
    private Date gmtCreate;
}
