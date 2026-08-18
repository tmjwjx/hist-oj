package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;

@Data
@Accessors(chain = true)
public class BattleRoomInfoVO {

    private BattleRoom room;

    private BattleProblemVO problem;
}
