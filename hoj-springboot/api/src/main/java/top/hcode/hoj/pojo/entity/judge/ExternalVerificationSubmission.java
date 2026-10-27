package top.hcode.hoj.pojo.entity.judge;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;
import java.util.Date;

@Data
@Accessors(chain = true)
public class ExternalVerificationSubmission {
    @TableId(value = "submit_id", type = IdType.INPUT)
    private Long submitId;
    private Long pid;
    private String visitorTokenHash;
    private String nickname;
    private Date gmtCreate;
}
