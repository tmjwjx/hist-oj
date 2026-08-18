package top.hcode.hoj.dao.battle.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.battle.BattleRoomEntityService;
import top.hcode.hoj.mapper.BattleRoomMapper;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;

@Service
public class BattleRoomEntityServiceImpl
        extends ServiceImpl<BattleRoomMapper, BattleRoom>
        implements BattleRoomEntityService {
}
