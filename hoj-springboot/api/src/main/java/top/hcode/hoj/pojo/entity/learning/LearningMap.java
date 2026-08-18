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
@TableName("learning_map")
public class LearningMap implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String title;
    private String description;
    private String status;
    private String accessMode;
    private Date createTime;
    private Date updateTime;
}
