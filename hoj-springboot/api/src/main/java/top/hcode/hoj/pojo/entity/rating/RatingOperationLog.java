package top.hcode.hoj.pojo.entity.rating;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("rating_operation_logs")
public class RatingOperationLog {
    @TableId(type = IdType.AUTO) private Long id;
    private String operatorUid;
    private String operatorUsername;
    private String operationType;
    private String targetType;
    private String targetId;
    private String operationDetail;
    private String ip;
    private Date createdAt;
}
