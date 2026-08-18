package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class BattleResultVO {

    private String winner;

    private String winnerId;

    private String endReason;

    private Boolean isWinner;
}
