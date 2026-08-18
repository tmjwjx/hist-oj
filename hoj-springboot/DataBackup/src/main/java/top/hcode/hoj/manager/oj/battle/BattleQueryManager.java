package top.hcode.hoj.manager.oj.battle;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.battle.BattleRecordEntityService;
import top.hcode.hoj.mapper.BattleRecordMapper;
import top.hcode.hoj.pojo.entity.battle.BattleRecord;
import top.hcode.hoj.pojo.entity.user.UserRecord;
import top.hcode.hoj.pojo.vo.BattleRankPageVO;
import top.hcode.hoj.pojo.vo.BattleRankVO;
import top.hcode.hoj.pojo.vo.BattleRecordPageVO;
import top.hcode.hoj.utils.ShiroUtils;

import javax.annotation.Resource;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Component
public class BattleQueryManager {

    @Resource
    private BattleRecordEntityService recordService;

    @Resource
    private BattleRecordMapper recordMapper;

    @Resource
    private top.hcode.hoj.dao.user.UserRecordEntityService userRecordService;

    public BattleRecordPageVO myRecords(Integer limit, Integer currentPage) {
        Page<BattleRecord> page = new Page<>(safePage(currentPage), safeLimit(limit, 10));
        IPage<BattleRecord> result = recordService.page(page, new QueryWrapper<BattleRecord>()
                .eq("user_id", uid())
                .orderByDesc("gmt_create"));
        fillRatings(result.getRecords());
        return new BattleRecordPageVO().setTotal(result.getTotal()).setRecords(result.getRecords());
    }

    public BattleRecordPageVO allRecords(Integer limit, Integer currentPage,
                                         String username, String roomId, String isWinner) {
        Page<BattleRecord> page = new Page<>(safePage(currentPage), safeLimit(limit, 20));
        QueryWrapper<BattleRecord> query = new QueryWrapper<BattleRecord>().orderByDesc("gmt_create");
        if (username != null && !username.trim().isEmpty()) {
            query.and(w -> w.like("username", username).or().like("opponent_username", username));
        }
        if (roomId != null && !roomId.trim().isEmpty()) {
            query.like("room_id", roomId.trim());
        }
        if ("true".equalsIgnoreCase(isWinner) || "false".equalsIgnoreCase(isWinner)) {
            query.eq("is_winner", Boolean.parseBoolean(isWinner));
        }
        IPage<BattleRecord> result = recordService.page(page, query);
        fillRatings(result.getRecords());
        return new BattleRecordPageVO().setTotal(result.getTotal()).setRecords(result.getRecords());
    }

    public BattleRankPageVO rank(Integer limit, Integer currentPage, String username) {
        Page<BattleRankVO> page = new Page<>(safePage(currentPage), safeLimit(limit, 50));
        IPage<BattleRankVO> result = recordMapper.selectRank(page, username);
        int start = (int) ((page.getCurrent() - 1) * page.getSize());
        for (int i = 0; i < result.getRecords().size(); i++) {
            result.getRecords().get(i).setRank(start + i + 1);
        }
        return new BattleRankPageVO().setTotal(result.getTotal()).setRankList(result.getRecords());
    }

    @Transactional(rollbackFor = Exception.class)
    public void exclude(Long recordId, boolean excluded)
            throws StatusNotFoundException, StatusFailException {
        BattleRecord record = recordService.getById(recordId);
        if (record == null) {
            throw new StatusNotFoundException("对战记录不存在");
        }
        if (Boolean.valueOf(excluded).equals(record.getIsExcluded())) {
            return;
        }
        UpdateWrapper<BattleRecord> update = new UpdateWrapper<BattleRecord>()
                .eq("battle_pair_id", record.getBattlePairId())
                .set("is_excluded", excluded);
        if (record.getBattlePairId() == null) {
            update = new UpdateWrapper<BattleRecord>()
                    .eq("id", record.getId())
                    .set("is_excluded", excluded);
        }
        if (recordService.update(update)) {
            return;
        }
        if (record.getBattlePairId() == null || recordService.count(
                new QueryWrapper<BattleRecord>().eq("battle_pair_id", record.getBattlePairId())) == 0) {
            throw new StatusFailException("更新对战记录失败");
        }
    }

    private void fillRatings(List<BattleRecord> records) {
        if (records == null || records.isEmpty()) {
            return;
        }
        List<String> uids = records.stream().map(BattleRecord::getUserId).distinct().collect(Collectors.toList());
        List<UserRecord> userRecords = userRecordService.list(new QueryWrapper<UserRecord>().in("uid", uids));
        Map<String, Integer> ratings = new HashMap<>();
        for (UserRecord userRecord : userRecords) {
            ratings.put(userRecord.getUid(), userRecord.getHistRating());
        }
        for (BattleRecord record : records) {
            record.setUserRating(ratings.get(record.getUserId()));
        }
    }

    private String uid() {
        return ShiroUtils.getProfile().getUid();
    }

    private long safePage(Integer page) {
        return page == null || page < 1 ? 1 : page;
    }

    private long safeLimit(Integer limit, int defaultLimit) {
        if (limit == null || limit < 1) {
            return defaultLimit;
        }
        return Math.min(limit, 10000);
    }
}
