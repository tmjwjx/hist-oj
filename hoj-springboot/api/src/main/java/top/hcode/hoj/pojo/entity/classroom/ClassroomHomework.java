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
@TableName("classroom_homework")
public class ClassroomHomework implements Serializable {
    @TableId(type = IdType.AUTO) private Long id;
    private Long classroomId;
    private String title;
    private String description;
    private Date startTime;
    private Date endTime;
    private Integer showScore;
    private Integer showHomework;
    private Integer showAnswer;
    private Integer showRank;
    private Integer status;
    private Integer isExamMode;
    private Integer examDuration;
    private Integer allowSubmitAfterMinutes;
    private Integer disableCopyPaste;
    private Integer requireFullscreen;
    private Integer disallowTabSwitch;
    private Date createTime;
    private Date updateTime;
}
