package top.hcode.hoj.pojo.vo;

import lombok.Data;

@Data
public class BattleRankVO {

    private String userId;

    private String username;

    private Integer totalBattles;

    private Integer winCount;

    private Integer loseCount;

    private Double winRate;

    private Integer totalSubmitCount;

    private Integer avgBattleTime;

    private Integer rating;

    private Integer rank;
}
