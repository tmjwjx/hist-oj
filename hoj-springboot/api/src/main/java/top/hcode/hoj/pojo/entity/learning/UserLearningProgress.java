package top.hcode.hoj.pojo.entity.learning;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("user_learning_progress")
public class UserLearningProgress implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String userId;
    private Long mapId;
    private Long nodeId;
    private String status;
    private Date completedAt;
    private Date masteredAt;
    private Date createTime;
    private Date updateTime;
}
