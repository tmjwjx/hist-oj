package top.hcode.hoj.pojo.entity.battle;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("battle_record")
public class BattleRecord implements Serializable {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String battlePairId;

    private String roomId;

    private String userId;

    private String username;

    @TableField(exist = false)
    private Integer userRating;

    private String opponentId;

    private String opponentUsername;

    private Integer opponentRating;

    private String problemId;

    private String problemTitle;

    private Boolean isWinner;

    private String endReason;

    @JsonProperty("SubmitCount")
    private Integer submitCount;

    private Integer battleTime;

    private Boolean isExcluded;

    private Date gmtCreate;
}
