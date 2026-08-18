package top.hcode.hoj.pojo.entity.rating;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("rating_recalculate_queue")
public class RatingRecalculateQueue {
    @TableId(type = IdType.AUTO) private Long id;
    private Long contestId;
    private String status;
    private Integer totalContests;
    private Integer processedContests;
    private String errorMessage;
    private String createdBy;
    private Date createdAt;
    private Date startedAt;
    private Date completedAt;
}
