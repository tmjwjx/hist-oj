package top.hcode.hoj.pojo.vo.classroom;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
public class QuestionBankVO {
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
    private Date createdAt;
    private Date updatedAt;
    private ClassroomUserVO creator;
}
