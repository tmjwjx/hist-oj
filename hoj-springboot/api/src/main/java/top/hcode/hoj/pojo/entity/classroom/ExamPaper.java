package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("classroom_exam_paper")
public class ExamPaper implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String title;
    private String creatorId;
    private Integer isShared;
    private Integer isPublic;
    private Integer totalScore;
    private Integer questionCount;
    private String description;
    private Integer status;
    private Date createTime;
    private Date updateTime;
}
