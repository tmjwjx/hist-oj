package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;
import top.hcode.hoj.pojo.entity.battle.BattleRecord;
import top.hcode.hoj.pojo.vo.BattleRankVO;

@Mapper
@Repository
public interface BattleRecordMapper extends BaseMapper<BattleRecord> {

    IPage<BattleRankVO> selectRank(Page<BattleRankVO> page,
                                   @Param("username") String username);
}
