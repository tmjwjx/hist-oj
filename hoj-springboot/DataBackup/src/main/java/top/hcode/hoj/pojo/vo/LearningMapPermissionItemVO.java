package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
public class LearningMapPermissionItemVO {
    private String userId;
    private String username;
    private String nickname;
    private String realname;
    private Boolean enabled;
    private Date updatedAt;
}
