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
@TableName("learning_map_node")
public class LearningMapNode implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long mapId;
    private String type;
    private String title;
    private String description;
    private String difficulty;
    private String tags;
    private Double x;
    private Double y;
    private Integer level;
    private String region;
    private Boolean published;
    private String knowledgeContent;
    private Long problemId;
    private String problemDisplayId;
    private String metadata;
    private Date createTime;
    private Date updateTime;
}
