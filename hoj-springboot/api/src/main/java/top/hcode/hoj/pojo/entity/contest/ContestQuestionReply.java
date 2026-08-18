package top.hcode.hoj.pojo.entity.contest;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("contest_question_reply")
public class ContestQuestionReply implements Serializable {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long questionId;

    private String senderId;

    private String content;

    private Date createdAt;
}
