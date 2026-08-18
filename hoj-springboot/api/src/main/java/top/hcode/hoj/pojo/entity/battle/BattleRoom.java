package top.hcode.hoj.pojo.entity.battle;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("battle_room")
public class BattleRoom implements Serializable {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String roomId;

    private String hostId;

    private String hostUsername;

    @TableField(exist = false)
    private Integer hostRating;

    private String challengerId;

    private String challengerUsername;

    @TableField(exist = false)
    private Integer challengerRating;

    private Boolean challengerReady;

    private String problemId;

    private Integer status;

    private String winnerId;

    private String endReason;

    private Date startTime;

    private Date endTime;

    private Date gmtCreate;

    private Date gmtModified;
}
