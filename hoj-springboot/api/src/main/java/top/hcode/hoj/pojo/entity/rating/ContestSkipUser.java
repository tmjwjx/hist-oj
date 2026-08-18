package top.hcode.hoj.pojo.entity.rating;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("contest_skip_users")
public class ContestSkipUser {
    @TableId(type = IdType.AUTO) private Long id;
    private Long contestId;
    private String uid;
    private String username;
    private String reason;
    private String operatorUid;
    private String operatorUsername;
    private Boolean isApplied;
    private Date createdAt;
    private Date updatedAt;
}
