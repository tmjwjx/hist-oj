package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;
import java.io.Serializable;
import java.util.Date;

@Data @Accessors(chain = true) @TableName("classroom_material")
public class ClassroomMaterial implements Serializable {
    @TableId(type = IdType.AUTO) private Long id; private Long classroomId; private Long folderId; private String fileName; private String fileType; private String filePath; private Long fileSize; private String creatorId; private Integer isShared; private Integer downloadCount; private Integer status; private Date createTime; private Date updateTime;
}
