package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;
import java.util.List;

@Data
@Accessors(chain = true)
public class ContestQuestionVO {
    private Long id;
    private Long contestId;
    private String questionerId;
    private String title;
    private String content;
    private String status;
    private Integer priority;
    private Date createdAt;
    private Date updatedAt;
    private ContestQuestionUserVO questioner;
    private List<ContestQuestionReplyVO> replies;
}
