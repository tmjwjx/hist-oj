package top.hcode.hoj.pojo.entity.learning;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("learning_map_edge")
public class LearningMapEdge implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long mapId;
    private Long sourceNodeId;
    private Long targetNodeId;
    private String type;
    private Date createTime;
    private Date updateTime;
}
