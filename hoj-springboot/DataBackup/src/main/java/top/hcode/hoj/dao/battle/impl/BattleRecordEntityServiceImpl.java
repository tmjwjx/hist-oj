package top.hcode.hoj.dao.battle.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.battle.BattleRecordEntityService;
import top.hcode.hoj.mapper.BattleRecordMapper;
import top.hcode.hoj.pojo.entity.battle.BattleRecord;

@Service
public class BattleRecordEntityServiceImpl
        extends ServiceImpl<BattleRecordMapper, BattleRecord>
        implements BattleRecordEntityService {
}
