package top.hcode.hoj.pojo.entity.problem;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;
import java.util.Date;

@Data
@Accessors(chain = true)
public class ProblemExternalVerification {
    @TableId(value = "pid", type = IdType.INPUT)
    private Long pid;
    private String token;
    private Boolean enabled;
    private String ownerUid;
    private String ownerUsername;
    private Date gmtCreate;
    private Date gmtModified;
}
