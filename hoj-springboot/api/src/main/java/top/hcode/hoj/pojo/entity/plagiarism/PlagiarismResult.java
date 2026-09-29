package top.hcode.hoj.pojo.entity.plagiarism;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
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
    @TableField("submit_id_1")
    private Long submitId1;
    @TableField("submit_id_2")
    private Long submitId2;
    @TableField("contest_record_id_1")
    private Long contestRecordId1;
    @TableField("contest_record_id_2")
    private Long contestRecordId2;
    @TableField("uid_1")
    private String uid1;
    @TableField("uid_2")
    private String uid2;
    @TableField("username_1")
    private String username1;
    @TableField("username_2")
    private String username2;
    private String language;
    @TableField("similarity_1_to_2")
    private Integer similarity1to2;
    @TableField("similarity_2_to_1")
    private Integer similarity2to1;
    private Integer maxSimilarity;
    private Boolean isOverThreshold;
    private Date gmtCreate;
}
