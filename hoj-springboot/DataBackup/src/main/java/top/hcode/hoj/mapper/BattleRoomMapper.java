package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.springframework.stereotype.Repository;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;

@Mapper
@Repository
public interface BattleRoomMapper extends BaseMapper<BattleRoom> {

    @Select("SELECT * FROM battle_room WHERE room_id = #{roomId} FOR UPDATE")
    BattleRoom selectByRoomIdForUpdate(@Param("roomId") String roomId);

    @Select("SELECT * FROM battle_room "
            + "WHERE (host_id = #{uid} OR challenger_id = #{uid}) "
            + "AND status IN (0, 1) LIMIT 1 FOR UPDATE")
    BattleRoom selectActiveRoomForUpdate(@Param("uid") String uid);
}
