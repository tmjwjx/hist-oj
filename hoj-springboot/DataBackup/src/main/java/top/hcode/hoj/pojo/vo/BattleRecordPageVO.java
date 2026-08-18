package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;
import top.hcode.hoj.pojo.entity.battle.BattleRecord;

import java.util.List;

@Data
@Accessors(chain = true)
public class BattleRecordPageVO {

    private Long total;

    private List<BattleRecord> records;
}
