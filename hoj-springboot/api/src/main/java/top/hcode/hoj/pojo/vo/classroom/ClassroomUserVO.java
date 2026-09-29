package top.hcode.hoj.pojo.vo.classroom;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ClassroomUserVO {
    private String uid;
    private String uuid;
    private String username;
    private String nickname;
    private String realname;
    private String email;
    private String avatar;
    private Integer status;
    // 班级角色（teacher/student，空表示无角色），供管理端搜索下拉展示
    private String role;
}
