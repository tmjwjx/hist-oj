package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("classroom_user_role")
public class ClassroomUserRole implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String uid;
    private String role;
    private Date createTime;
    private Date updateTime;
}
