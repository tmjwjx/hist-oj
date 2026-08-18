package top.hcode.hoj.manager.rating;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.rating.*;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.dao.user.UserRecordEntityService;
import top.hcode.hoj.manager.oj.ContestCalculateRankManager;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.rating.*;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.entity.user.UserRecord;
import top.hcode.hoj.pojo.vo.ACMContestRankVO;
import top.hcode.hoj.pojo.vo.OIContestRankVO;
import top.hcode.hoj.utils.ShiroUtils;
import top.hcode.hoj.utils.Constants;

import javax.annotation.Resource;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

@Component
public class RatingManager {
    private static final int INITIAL_RATING = 0;
    private static final int[] NEWBIE_BONUS = {500, 350, 250, 150, 100, 50};

    @Resource private ContestEntityService contestService;
    @Resource(name = "contestCalculateRankManager")
    private ContestCalculateRankManager contestCalculateRankManager;
    @Resource private UserInfoEntityService userInfoService;
    @Resource private UserRecordEntityService userRecordService;
    @Resource private RatingHistoryEntityService historyService;
    @Resource private ContestRatingStatusEntityService statusService;
    @Resource private ContestSkipUserEntityService skipService;
    @Resource private RatingRecalculateQueueEntityService queueService;
    @Resource private RatingOperationLogEntityService logService;

    public Contest requireContest(Long cid) throws StatusNotFoundException {
        Contest contest = contestService.getById(cid);
        if (contest == null) throw new StatusNotFoundException("比赛不存在");
        return contest;
    }

    public void requireOwnerOrAdmin(Contest contest) throws StatusForbiddenException {
        if (isAdmin()) return;
        if (ShiroUtils.getProfile() == null || !contest.getUid().equals(ShiroUtils.getProfile().getUid())) {
            throw new StatusForbiddenException("无权限管理 Rating");
        }
    }

    public boolean isAdmin() {
        return SecurityUtils.getSubject().hasRole("root") || SecurityUtils.getSubject().hasRole("admin");
    }

    @Transactional(rollbackFor = Exception.class)
    public void setRatingType(Long cid, boolean rated) throws Exception {
        Contest contest = requireContest(cid); requireOwnerOrAdmin(contest);
        contest.setIsRating(rated); contestService.updateById(contest);
        ContestRatingStatus status = statusService.getById(cid);
        if (status == null) statusService.save(new ContestRatingStatus().setContestId(cid).setIsRated(rated)
                .setRatingCalculated(false).setSkipCount(0).setHasPendingSkip(false).setRecalculateLock(false));
        else statusService.updateById(status.setIsRated(rated));
    }

    public void initialize(String uid, int rating) {
        UserRecord record = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", uid));
        if (record == null) userRecordService.save(new UserRecord().setUid(uid).setHistRating(rating));
    }

    public boolean canCalculate(Long cid) {
        ContestRatingStatus status = statusService.getById(cid);
        return status != null && Boolean.TRUE.equals(status.getIsRated())
                && !Boolean.TRUE.equals(status.getRatingCalculated());
    }

    /** 查找已结束且尚未计算的计分比赛，供定时任务和管理员手动触发使用。 */
    public int calculatePending() {
        Date now = new Date();
        List<Contest> contests = contestService.list(new QueryWrapper<Contest>()
                .eq("is_rating", true).le("end_time", now).orderByAsc("end_time"));
        int calculated = 0;
        for (Contest contest : contests) {
            ContestRatingStatus status = statusService.getById(contest.getId());
            if (status == null) {
                status = new ContestRatingStatus().setContestId(contest.getId()).setIsRated(true)
                        .setRatingCalculated(false).setSkipCount(0).setHasPendingSkip(false)
                        .setRecalculateLock(false);
                statusService.save(status);
            } else if (!Boolean.TRUE.equals(status.getIsRated())) {
                statusService.updateById(status.setIsRated(true));
            }
            if (!canCalculate(contest.getId())) continue;
            try {
                calculate(contest.getId());
                calculated++;
            } catch (Exception ignored) {
                // 人数不足等业务错误不应中断其他比赛，下一轮任务会再次检查。
            }
        }
        return calculated;
    }

    /** 计算成绩、写入历史并更新 user_record.hist_rating。 */
    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> calculate(Long cid) throws Exception {
        Contest contest = requireContest(cid);
        if (ShiroUtils.getProfile() != null) requireOwnerOrAdmin(contest);
        ContestRatingStatus status = statusService.getOne(new QueryWrapper<ContestRatingStatus>()
                .eq("contest_id", cid).last("FOR UPDATE"));
        if (status == null || !Boolean.TRUE.equals(status.getIsRated())) throw new StatusFailException("比赛未开启 Rating");
        if (Boolean.TRUE.equals(status.getRatingCalculated())) throw new StatusFailException("该比赛已经计算过 Rating");
        status.setRatingCalculated(true).setCalculatedAt(new Date()).setRecalculateStatus("calculating");
        statusService.updateById(status);
        List<Score> ranked = leaderboardScores(contest);
        List<ContestSkipUser> skipRows = skipService.list(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid));
        Set<String> skipped = skipRows.stream().map(ContestSkipUser::getUid).collect(Collectors.toSet());
        List<Score> active = ranked.stream().filter(s -> !skipped.contains(s.uid)).collect(Collectors.toList());
        if (active.size() < 3) throw new StatusFailException("有效参赛人数不足 3 人");

        Set<String> allUids = ranked.stream().map(s -> s.uid).collect(Collectors.toSet());
        allUids.addAll(skipped);
        Map<String, Integer> current = currentRatings(allUids);
        Map<String, Long> contestCounts = contestCounts(current.keySet());
        List<RatingCalculator.Participant> inputs = new ArrayList<>();
        for (Score score : active) {
            int old = current.getOrDefault(score.uid, INITIAL_RATING);
            long count = contestCounts.getOrDefault(score.uid, 0L);
            inputs.add(new RatingCalculator.Participant(score.uid, active.indexOf(score) + 1,
                    old + remainingBonus(count)));
        }
        Map<String, Integer> changes = RatingCalculator.changes(inputs);
        List<RatingHistory> histories = new ArrayList<>();
        for (Score score : active) {
            int old = current.getOrDefault(score.uid, INITIAL_RATING);
            int change = changes.getOrDefault(score.uid, 0) + displayBonus(contestCounts.getOrDefault(score.uid, 0L));
            int next = RatingCalculator.newRating(old, change);
            saveUserRating(score.uid, next);
            histories.add(new RatingHistory().setUid(score.uid).setContestId(cid).setOldRating(old).setNewRating(next)
                    .setRatingChange(change).setRank(score.rank).setParticipants(active.size()).setIsManual(false)
                    .setIsSkip(false).setReason(displayBonus(contestCounts.getOrDefault(score.uid, 0L)) > 0
                            ? "cf_newbie_bonus:+" + displayBonus(contestCounts.getOrDefault(score.uid, 0L)) : "")
                    .setCreatedAt(new Date()));
        }
        for (Score score : ranked) if (skipped.contains(score.uid)) {
            int old = current.getOrDefault(score.uid, INITIAL_RATING);
            ContestSkipUser skip = skipService.getOne(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid).eq("uid", score.uid));
            histories.add(new RatingHistory().setUid(score.uid).setContestId(cid).setOldRating(old).setNewRating(old)
                    .setRatingChange(0).setRank(score.rank).setParticipants(active.size()).setIsManual(false).setIsSkip(true)
                    .setReason(skip == null ? "skip" : skip.getReason()).setSkipReason(skip == null ? "skip" : skip.getReason()).setCreatedAt(new Date()));
            saveUserRating(score.uid, old);
        }
        Set<String> rankedUids = ranked.stream().map(s -> s.uid).collect(Collectors.toSet());
        for (ContestSkipUser skip : skipRows) if (!rankedUids.contains(skip.getUid())) {
            int old = current.getOrDefault(skip.getUid(), INITIAL_RATING);
            histories.add(new RatingHistory().setUid(skip.getUid()).setContestId(cid).setOldRating(old).setNewRating(old)
                    .setRatingChange(0).setRank(0).setParticipants(active.size()).setIsManual(false).setIsSkip(true)
                    .setReason(skip.getReason()).setSkipReason(skip.getReason()).setCreatedAt(new Date()));
            saveUserRating(skip.getUid(), old);
        }
        historyService.saveBatch(histories);
        applyRelatedManualAdjustments(cid);
        if (!skipRows.isEmpty()) {
            skipRows.forEach(row -> row.setIsApplied(true));
            skipService.updateBatchById(skipRows);
        }
        status.setRecalculateStatus("completed");
        statusService.updateById(status);
        return new LinkedHashMap<String, Object>() {{ put("contestId", cid); put("participants", histories.size()); }};
    }

    private List<Score> leaderboardScores(Contest contest) {
        List<Score> result = new ArrayList<>();
        if (Objects.equals(contest.getType(), Constants.Contest.TYPE_OI.getCode())) {
            List<OIContestRankVO> ranks = contestCalculateRankManager.calcOIRank(false, true, contest,
                    null, new ArrayList<>(), Collections.emptyList(), false);
            for (OIContestRankVO rank : ranks) {
                result.add(new Score(rank.getUid(), rank.getUsername(), rank.getRank()));
            }
        } else {
            List<ACMContestRankVO> ranks = contestCalculateRankManager.calcACMRank(false, true, contest,
                    null, new ArrayList<>(), Collections.emptyList(), false);
            for (ACMContestRankVO rank : ranks) {
                result.add(new Score(rank.getUid(), rank.getUsername(), rank.getRank()));
            }
        }
        return result;
    }

    private Map<String, Integer> currentRatings(Set<String> uids) {
        Map<String, Integer> result = new HashMap<>();
        if (uids.isEmpty()) return result;
        for (UserRecord record : userRecordService.list(new QueryWrapper<UserRecord>().in("uid", uids)))
            result.put(record.getUid(), record.getHistRating() == null ? INITIAL_RATING : record.getHistRating());
        return result;
    }

    private Map<String, Long> contestCounts(Set<String> uids) {
        Map<String, Long> result = new HashMap<>();
        for (RatingHistory history : historyService.list(new QueryWrapper<RatingHistory>().in("uid", uids)
                .isNotNull("contest_id").eq("is_manual", false).eq("is_skip", false)))
            result.merge(history.getUid(), 1L, Long::sum);
        return result;
    }

    private void saveUserRating(String uid, int rating) {
        UserRecord record = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", uid));
        if (record == null) userRecordService.save(new UserRecord().setUid(uid).setHistRating(rating));
        else userRecordService.updateById(record.setHistRating(rating));
    }

    private void applyRelatedManualAdjustments(Long cid) {
        List<RatingHistory> adjustments = historyService.list(new QueryWrapper<RatingHistory>()
                .eq("is_manual", true).eq("related_contest_id", cid).orderByAsc("created_at", "id"));
        for (RatingHistory adjustment : adjustments) {
            UserRecord record = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", adjustment.getUid()));
            int old = record == null || record.getHistRating() == null ? INITIAL_RATING : record.getHistRating();
            int next = RatingCalculator.newRating(old, adjustment.getRatingChange());
            saveUserRating(adjustment.getUid(), next);
            historyService.updateById(adjustment.setOldRating(old).setNewRating(next));
        }
    }

    private int displayBonus(long completed) { return completed < NEWBIE_BONUS.length ? NEWBIE_BONUS[(int) completed] : 0; }
    private int remainingBonus(long completed) { int total = 0; for (int i = (int) Math.max(0, completed); i < NEWBIE_BONUS.length; i++) total += NEWBIE_BONUS[i]; return total; }

    public Map<String, Object> adjust(String username, int delta, String reason, Long relatedContestId) throws Exception {
        if (delta == 0 || reason == null || reason.trim().isEmpty()) throw new StatusFailException("变化值和原因不能为空");
        UserInfo user = userInfoService.getOne(new QueryWrapper<UserInfo>().eq("username", username));
        if (user == null) throw new StatusNotFoundException("用户不存在");
        if (relatedContestId != null) {
            ContestRatingStatus status = statusService.getById(relatedContestId);
            if (status == null || !Boolean.TRUE.equals(status.getIsRated()) || !Boolean.TRUE.equals(status.getRatingCalculated()))
                throw new StatusFailException("关联比赛尚未完成 Rating 计算");
        }
        UserRecord record = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", user.getUuid()));
        int old = record == null || record.getHistRating() == null ? INITIAL_RATING : record.getHistRating();
        int next = RatingCalculator.newRating(old, delta); saveUserRating(user.getUuid(), next);
        RatingHistory history = new RatingHistory().setUid(user.getUuid()).setOldRating(old).setNewRating(next)
                .setRatingChange(delta).setRank(0).setParticipants(0).setReason(reason).setIsManual(true)
                .setIsSkip(false).setRelatedContestId(relatedContestId).setOperatorUid(operatorUid()).setCreatedAt(new Date());
        historyService.save(history); log("personal_adjust", "user", username, reason);
        return new LinkedHashMap<String, Object>() {{ put("username", username); put("oldRating", old); put("newRating", next); put("ratingChange", delta); put("reason", reason); put("relatedContestId", relatedContestId); }};
    }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> cancelAdjustment(Long id) throws Exception {
        RatingHistory history = historyService.getById(id);
        if (history == null || !Boolean.TRUE.equals(history.getIsManual())) throw new StatusNotFoundException("手动调整记录不存在");
        UserRecord record = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", history.getUid()));
        int current = record == null || record.getHistRating() == null ? INITIAL_RATING : record.getHistRating();
        saveUserRating(history.getUid(), RatingCalculator.newRating(current, -history.getRatingChange()));
        historyService.removeById(id); log("cancel_personal_adjust", "manual_adjustment", String.valueOf(id), "撤销手动调整");
        return new LinkedHashMap<String, Object>() {{ put("adjustmentId", id); put("deleted", true); put("ratingChange", -history.getRatingChange()); }};
    }

    public List<ContestSkipUser> skipUsers(Long cid) throws Exception { Contest c = requireContest(cid); requireOwnerOrAdmin(c); return skipService.list(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid).orderByDesc("created_at")); }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> batchSkip(Long cid, List<String> usernames, String reason, boolean autoRecalc) throws Exception {
        Contest c = requireContest(cid); requireOwnerOrAdmin(c);
        Map<String, Object> result = new LinkedHashMap<>(); List<String> success = new ArrayList<>(), failed = new ArrayList<>(), duplicate = new ArrayList<>();
        for (String username : usernames) {
            UserInfo user = userInfoService.getOne(new QueryWrapper<UserInfo>().eq("username", username));
            if (user == null) { failed.add(username); continue; }
            if (skipService.count(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid).eq("uid", user.getUuid())) > 0) { duplicate.add(username); continue; }
            skipService.save(new ContestSkipUser().setContestId(cid).setUid(user.getUuid()).setUsername(username).setReason(reason)
                    .setOperatorUid(operatorUid()).setOperatorUsername(operatorName()).setIsApplied(false).setCreatedAt(new Date())); success.add(username);
        }
        ContestRatingStatus status = statusService.getById(cid); if (status != null) statusService.updateById(status.setHasPendingSkip(!success.isEmpty())
                .setSkipCount((status.getSkipCount() == null ? 0 : status.getSkipCount()) + success.size()).setSkipDataChangedAt(new Date()));
        result.put("successUsers", success); result.put("failedUsers", failed); result.put("duplicatedUsers", duplicate); result.put("pendingRecalculate", !success.isEmpty());
        if (autoRecalc && !success.isEmpty()) result.put("taskId", recalculate(cid));
        return result;
    }

    @Transactional(rollbackFor = Exception.class)
    public void cancelSkip(Long cid, List<String> uids) throws Exception {
        Contest c = requireContest(cid); requireOwnerOrAdmin(c);
        skipService.remove(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid).in("uid", uids));
        ContestRatingStatus status = statusService.getById(cid); if (status != null) statusService.updateById(status.setHasPendingSkip(false).setSkipDataChangedAt(new Date()));
    }

    public Long recalculate(Long fromCid) throws Exception {
        requireOwnerOrAdmin(requireContest(fromCid));
        List<Contest> contests = contestService.list(new QueryWrapper<Contest>().ge("id", fromCid).eq("is_rating", true).eq("status", 1).orderByAsc("id"));
        if (contests.isEmpty()) throw new StatusFailException("没有需要重算的比赛");
        RatingRecalculateQueue task = new RatingRecalculateQueue().setContestId(fromCid).setStatus("pending").setTotalContests(contests.size()).setProcessedContests(0).setCreatedBy(operatorUid()).setCreatedAt(new Date());
        queueService.save(task); CompletableFuture.runAsync(() -> runRecalculate(task.getId(), contests)); return task.getId();
    }

    private void runRecalculate(Long taskId, List<Contest> contests) {
        RatingRecalculateQueue task = queueService.getById(taskId); if (task == null) return;
        task.setStatus("running").setStartedAt(new Date()); queueService.updateById(task);
        try {
            resetForRecalculation(contests);
            for (int i = 0; i < contests.size(); i++) {
                Contest contest = contests.get(i);
                calculate(contest.getId()); task.setProcessedContests(i + 1); queueService.updateById(task);
            }
            task.setStatus("completed").setCompletedAt(new Date()); queueService.updateById(task);
        } catch (Exception ex) { task.setStatus("failed").setErrorMessage(ex.getMessage()).setCompletedAt(new Date()); queueService.updateById(task); }
    }

    private void resetForRecalculation(List<Contest> contests) {
        Long startCid = contests.get(0).getId();
        List<Long> ids = contests.stream().map(Contest::getId).collect(Collectors.toList());
        Set<String> affected = historyService.list(new QueryWrapper<RatingHistory>().in("contest_id", ids))
                .stream().map(RatingHistory::getUid).collect(Collectors.toSet());
        affected.addAll(skipService.list(new QueryWrapper<ContestSkipUser>().in("contest_id", ids))
                .stream().map(ContestSkipUser::getUid).collect(Collectors.toSet()));
        historyService.remove(new QueryWrapper<RatingHistory>().in("contest_id", ids));
        for (String uid : affected) {
            RatingHistory previous = historyService.getOne(new QueryWrapper<RatingHistory>().eq("uid", uid)
                    .eq("is_manual", false).isNotNull("contest_id").lt("contest_id", startCid)
                    .orderByDesc("contest_id", "id").last("LIMIT 1"));
            int rating = previous == null ? INITIAL_RATING : previous.getNewRating();
            List<RatingHistory> adjustments = historyService.list(new QueryWrapper<RatingHistory>().eq("uid", uid)
                    .eq("is_manual", true).isNotNull("related_contest_id").lt("related_contest_id", startCid)
                    .orderByAsc("related_contest_id", "created_at", "id"));
            for (RatingHistory adjustment : adjustments) rating = RatingCalculator.newRating(rating, adjustment.getRatingChange());
            saveUserRating(uid, rating);
        }
        for (Contest contest : contests) {
            ContestRatingStatus status = statusService.getById(contest.getId());
            if (status != null) statusService.updateById(status.setRatingCalculated(false).setRecalculateStatus("pending"));
        }
    }

    public RatingRecalculateQueue progress(Long taskId) { return queueService.getById(taskId); }
    public void resetLock(Long cid) throws Exception {
        Contest contest = requireContest(cid); requireOwnerOrAdmin(contest);
        ContestRatingStatus status = statusService.getById(cid);
        if (status != null) statusService.updateById(status.setRecalculateLock(false).setRecalculateStatus("none"));
    }
    public Map<String, Object> syncSkip(Long cid) throws Exception {
        Contest contest = requireContest(cid); requireOwnerOrAdmin(contest);
        List<ContestSkipUser> skips = skipService.list(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid).eq("is_applied", true));
        int updated = 0;
        for (ContestSkipUser skip : skips) {
            updated += historyService.update(new RatingHistory().setIsSkip(true).setSkipReason(skip.getReason()),
                    new QueryWrapper<RatingHistory>().eq("contest_id", cid).eq("uid", skip.getUid())) ? 1 : 0;
        }
        QueryWrapper<RatingHistory> clear = new QueryWrapper<RatingHistory>().eq("contest_id", cid).eq("is_skip", true);
        if (!skips.isEmpty()) clear.notIn("uid", skips.stream().map(ContestSkipUser::getUid).collect(Collectors.toList()));
        historyService.update(new RatingHistory().setIsSkip(false).setSkipReason(""), clear);
        Map<String, Object> result = new LinkedHashMap<>(); result.put("contestId", cid); result.put("updatedCount", updated); return result;
    }
    public List<RatingOperationLog> logs(int page, int limit, String type, String range) {
        QueryWrapper<RatingOperationLog> q = logQuery(type, range);
        return logService.page(new com.baomidou.mybatisplus.extension.plugins.pagination.Page<>(page, limit), q.orderByDesc("created_at")).getRecords();
    }
    public long logCount(String type, String range) { return logService.count(logQuery(type, range)); }
    private QueryWrapper<RatingOperationLog> logQuery(String type, String range) {
        QueryWrapper<RatingOperationLog> q = new QueryWrapper<>(); if (type != null && !type.isEmpty()) q.eq("operation_type", type);
        if ("7d".equals(range)) q.ge("created_at", new Date(System.currentTimeMillis() - 7L * 86400000));
        else if ("30d".equals(range)) q.ge("created_at", new Date(System.currentTimeMillis() - 30L * 86400000));
        return q;
    }
    private void log(String type, String targetType, String targetId, String detail) { logService.save(new RatingOperationLog().setOperatorUid(operatorUid()).setOperatorUsername(operatorName()).setOperationType(type).setTargetType(targetType).setTargetId(targetId).setOperationDetail(detail).setCreatedAt(new Date())); }
    private String operatorUid() { return ShiroUtils.getProfile() == null ? "system" : ShiroUtils.getProfile().getUid(); }
    private String operatorName() { return ShiroUtils.getProfile() == null ? "system" : ShiroUtils.getProfile().getUsername(); }

    private static class Score {
        private final String uid;
        private final String username;
        private final int rank;

        Score(String uid, String username, Integer rank) {
            this.uid = uid;
            this.username = username;
            this.rank = rank == null ? 0 : rank;
        }
    }
}
