package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.List;

@Data
@Accessors(chain = true)
public class BattleRankPageVO {

    private Long total;

    private List<BattleRankVO> rankList;
}
