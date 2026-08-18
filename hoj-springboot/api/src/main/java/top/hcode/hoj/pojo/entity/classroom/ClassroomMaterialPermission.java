package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;
import java.io.Serializable;
import java.util.Date;

@Data @Accessors(chain = true) @TableName("classroom_material_permission")
public class ClassroomMaterialPermission implements Serializable {
    @TableId(type = IdType.AUTO) private Long id; private Long materialId; private String studentUid; private Integer canPreview; private Integer canDownload; private Date createTime; private Date updateTime;
}
