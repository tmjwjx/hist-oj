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
@TableName("contest_question")
public class ContestQuestion implements Serializable {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long contestId;

    private String questionerId;

    private String title;

    private String content;

    private String status;

    private Integer priority;

    private Date createdAt;

    private Date updatedAt;
}
