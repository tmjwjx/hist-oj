package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;

@Data
@Accessors(chain = true)
@TableName("classroom_exam_paper_question")
public class ExamPaperQuestion implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long examPaperId;
    private Long questionId;
    private String problemId;
    private Integer questionOrder;
    private String questionType;
    private Integer score;
}
