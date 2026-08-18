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
@TableName("question_bank")
public class QuestionBank implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String title;
    private String type;
    private String content;
    private String options;
    private String answer;
    private String analysis;
    private String tags;
    private String course;
    private Integer difficulty;
    private Integer score;
    private String creatorId;
    private Integer isShared;
    private String problemId;
    private Integer status;
    private Date createTime;
    private Date updateTime;
}
