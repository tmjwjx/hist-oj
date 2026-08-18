package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class LearningMapPermissionUserVO {
    private String userId;
    private String username;
    private String nickname;
    private String realname;
}
