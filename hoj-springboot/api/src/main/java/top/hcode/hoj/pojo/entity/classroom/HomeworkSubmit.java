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
@TableName("homework_submit")
public class HomeworkSubmit implements Serializable {
    @TableId(type = IdType.AUTO) private Long id;
    private Long homeworkId;
    private Long questionId;
    private String problemId;
    private String uid;
    private String answer;
    private String attachment;
    private Long submitId;
    private Double score;
    private Integer isScored;
    private Integer isOfficiallySubmitted;
    private String judgeResult;
    private Date examStartTime;
    private Date examEndTime;
    private Integer isForcedSubmit;
    private Integer tabSwitchCount;
    private Integer fullscreenExitCount;
    private Integer copyPasteAttemptCount;
    private String deviceInfo;
    private String browserInfo;
    private Date createTime;
    private Date updateTime;
}
