package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
public class ContestQuestionReplyVO {
    private Long id;
    private Long questionId;
    private String senderId;
    private String content;
    private Date createdAt;
    private ContestQuestionUserVO sender;
}
