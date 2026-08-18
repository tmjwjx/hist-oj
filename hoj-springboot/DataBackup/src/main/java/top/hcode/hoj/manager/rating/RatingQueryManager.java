package top.hcode.hoj.manager.rating;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.stereotype.Component;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.rating.*;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.dao.user.UserRecordEntityService;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.rating.*;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.entity.user.UserRecord;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class RatingQueryManager {
    @Resource private UserInfoEntityService userInfoService;
    @Resource private UserRecordEntityService userRecordService;
    @Resource private RatingHistoryEntityService historyService;
    @Resource private ContestEntityService contestService;
    @Resource private ContestRatingStatusEntityService statusService;
    @Resource private ContestSkipUserEntityService skipService;
    @Resource private RatingOperationLogEntityService logService;

    public Map<String, Object> user(String uidOrUsername) {
        String uid = resolveUid(uidOrUsername); if (uid == null) return null;
        UserRecord record = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", uid));
        int rating = record == null || record.getHistRating() == null ? 0 : record.getHistRating();
        Integer max = historyService.getBaseMapper().selectObjs(new QueryWrapper<RatingHistory>().select("MAX(new_rating)").eq("uid", uid))
                .stream().filter(Objects::nonNull).map(v -> ((Number) v).intValue()).findFirst().orElse(rating);
        Map<String, Object> result = RatingColor.of(rating); result.put("uid", uid); result.put("rating", rating); result.put("maxRating", Math.max(rating, max)); return result;
    }

    public Map<String, Object> history(String uidOrUsername, int page, int limit) {
        String uid = resolveUid(uidOrUsername); if (uid == null) return pageData(page, limit, Collections.emptyList(), 0);
        List<RatingHistory> rows = historyService.list(new QueryWrapper<RatingHistory>().eq("uid", uid).orderByDesc("created_at"));
        Set<Long> ids = rows.stream().map(RatingHistory::getContestId).filter(Objects::nonNull).collect(Collectors.toSet());
        Map<Long, Contest> contests = contestService.listByIds(ids).stream().collect(Collectors.toMap(Contest::getId, c -> c));
        for (RatingHistory row : rows) if (row.getContestId() != null && contests.containsKey(row.getContestId())) {
            Contest c = contests.get(row.getContestId()); row.setContestTitle(c.getTitle()).setContestTime(c.getEndTime());
        }
        Set<Long> historyContests = ids;
        for (ContestSkipUser skip : skipService.list(new QueryWrapper<ContestSkipUser>().eq("uid", uid))) if (!historyContests.contains(skip.getContestId())) {
            Contest c = contestService.getById(skip.getContestId()); Map<String, Object> current = user(uid); int rating = current != null && current.get("rating") instanceof Integer ? (Integer) current.get("rating") : 0;
            rows.add(new RatingHistory().setUid(uid).setContestId(skip.getContestId()).setOldRating(rating).setNewRating(rating).setRatingChange(0).setRank(0).setParticipants(0).setReason(skip.getReason()).setSkipReason(skip.getReason()).setIsSkip(true).setCreatedAt(skip.getCreatedAt()).setContestTitle(c == null ? "未知比赛" : c.getTitle()).setContestTime(c == null ? null : c.getEndTime()));
        }
        rows = mergeRelatedManual(rows);
        rows.sort(Comparator.comparing(RatingHistory::getCreatedAt, Comparator.nullsLast(Date::compareTo)).reversed());
        int from = Math.min((page - 1) * limit, rows.size()), to = Math.min(from + limit, rows.size());
        return pageData(page, limit, rows.subList(from, to), rows.size());
    }

    public Map<String, Object> color(int rating) { return RatingColor.of(rating); }

    public Map<String, Object> contest(Long cid) {
        List<RatingHistory> histories = historyService.list(new QueryWrapper<RatingHistory>().eq("contest_id", cid).orderByAsc("`rank`"));
        Map<String, RatingHistory> byUser = histories.stream().collect(Collectors.toMap(RatingHistory::getUid, h -> h, (a, b) -> a));
        List<Map<String, Object>> records = new ArrayList<>();
        for (RatingHistory h : histories) records.add(participant(h.getUid(), h.getRank(), h.getOldRating(), h.getNewRating(), h.getRatingChange(), h.getIsSkip(), h.getSkipReason()));
        for (ContestSkipUser skip : skipService.list(new QueryWrapper<ContestSkipUser>().eq("contest_id", cid))) if (!byUser.containsKey(skip.getUid())) {
            UserRecord user = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", skip.getUid())); int rating = user == null || user.getHistRating() == null ? 0 : user.getHistRating();
            records.add(participant(skip.getUid(), 0, rating, rating, 0, true, skip.getReason()));
        }
        Map<String, Object> result = new LinkedHashMap<>(); result.put("contestId", cid); result.put("participants", records.size()); result.put("records", records); return result;
    }

    private Map<String, Object> participant(String uid, Integer rank, Integer oldRating, Integer newRating, Integer change, Boolean skip, String reason) {
        Map<String, Object> result = new LinkedHashMap<>(RatingColor.of(newRating == null ? 0 : newRating)); result.put("uid", uid); result.put("rank", rank); result.put("oldRating", oldRating); result.put("newRating", newRating); result.put("ratingChange", change); if (Boolean.TRUE.equals(skip)) { result.put("isSkip", true); result.put("skipReason", reason); } return result;
    }

    public Map<String, Object> batchUsers(List<String> uids) { Map<String, Object> result = new LinkedHashMap<>(); for (String uid : uids) { Map<String, Object> user = user(uid); result.put(uid, user == null ? participant(uid, 0, 0, 0, 0, false, "") : user); } return result; }

    public Map<String, Object> contestInfo(Long cid) {
        Contest contest = contestService.getById(cid); if (contest == null) return null; ContestRatingStatus status = statusService.getById(cid);
        Map<String, Object> result = new LinkedHashMap<>(); result.put("id", contest.getId()); result.put("title", contest.getTitle()); result.put("type", contest.getType()); result.put("isRating", Boolean.TRUE.equals(contest.getIsRating())); result.put("startTime", contest.getStartTime()); result.put("endTime", contest.getEndTime()); result.put("status", contest.getStatus()); result.put("calculatedAt", status == null ? null : status.getCalculatedAt()); result.put("skipDataChangedAt", status == null ? null : status.getSkipDataChangedAt()); return result;
    }

    public Map<String, Object> contestBatch(List<Long> ids) { Map<String, Object> result = new LinkedHashMap<>(); for (Long id : ids) { Map<String, Object> info = contestInfo(id); result.put(String.valueOf(id), info == null ? Collections.singletonMap("isRating", false) : info); } return result; }

    public Map<String, Object> rank(int page, int limit, String keyword) {
        List<UserInfo> users = userInfoService.list(); Map<String, Integer> ratings = new HashMap<>();
        for (UserRecord record : userRecordService.list()) ratings.put(record.getUid(), record.getHistRating() == null ? 0 : record.getHistRating());
        List<UserInfo> filtered = users.stream().filter(u -> keyword == null || keyword.isEmpty() || contains(u.getUsername(), keyword) || contains(u.getNickname(), keyword)).sorted((a, b) -> { int c = Integer.compare(ratings.getOrDefault(b.getUuid(), 0), ratings.getOrDefault(a.getUuid(), 0)); return c != 0 ? c : a.getUsername().compareTo(b.getUsername()); }).collect(Collectors.toList());
        List<Map<String, Object>> rows = new ArrayList<>(); for (UserInfo u : filtered) { int rating = ratings.getOrDefault(u.getUuid(), 0); Map<String, Object> row = new LinkedHashMap<>(RatingColor.of(rating)); row.put("uid", u.getUuid()); row.put("username", u.getUsername()); row.put("avatar", u.getAvatar()); row.put("nickname", u.getNickname()); row.put("school", u.getSchool()); row.put("titleName", u.getTitleName()); row.put("titleColor", u.getTitleColor()); rows.add(row); }
        int from = Math.min((page - 1) * limit, rows.size()), to = Math.min(from + limit, rows.size()); return pageData(page, limit, rows.subList(from, to), rows.size());
    }

    public Map<String, Object> manualHistory(int page, int limit) {
        Page<RatingHistory> p = new Page<>(page, limit); List<RatingHistory> rows = historyService.page(p, new QueryWrapper<RatingHistory>().eq("is_manual", true).orderByDesc("created_at")).getRecords();
        for (RatingHistory row : rows) { UserInfo u = userInfoService.getById(row.getUid()); row.setUsername(u == null ? row.getUid() : u.getUsername()); }
        return pageData(page, limit, rows, p.getTotal());
    }

    private List<RatingHistory> mergeRelatedManual(List<RatingHistory> rows) {
        Map<Long, RatingHistory> contestRows = rows.stream().filter(r -> r.getContestId() != null && !Boolean.TRUE.equals(r.getIsManual()))
                .collect(Collectors.toMap(RatingHistory::getContestId, r -> r, (a, b) -> a));
        Set<Long> merged = new HashSet<>();
        List<RatingHistory> manual = rows.stream().filter(r -> Boolean.TRUE.equals(r.getIsManual()) && r.getRelatedContestId() != null)
                .sorted(Comparator.comparing(RatingHistory::getCreatedAt, Comparator.nullsLast(Date::compareTo)).thenComparing(RatingHistory::getId)).collect(Collectors.toList());
        for (RatingHistory adjustment : manual) {
            RatingHistory target = contestRows.get(adjustment.getRelatedContestId()); if (target == null) continue;
            target.setNewRating(RatingCalculator.newRating(target.getNewRating(), adjustment.getRatingChange()));
            target.setManualAdjustDelta((target.getManualAdjustDelta() == null ? 0 : target.getManualAdjustDelta()) + adjustment.getRatingChange());
            if (adjustment.getReason() != null && !adjustment.getReason().trim().isEmpty()) target.setManualAdjustReason(target.getManualAdjustReason() == null ? adjustment.getReason() : target.getManualAdjustReason() + "；" + adjustment.getReason());
            if (target.getOldRating() != null) target.setRatingChange(target.getNewRating() - target.getOldRating()); merged.add(adjustment.getId());
        }
        return rows.stream().filter(r -> !merged.contains(r.getId())).collect(Collectors.toList());
    }

    private String resolveUid(String value) { UserInfo user = userInfoService.getOne(new QueryWrapper<UserInfo>().eq("uuid", value).or().eq("username", value)); return user == null ? null : user.getUuid(); }
    private boolean contains(String value, String keyword) { return value != null && value.toLowerCase().contains(keyword.toLowerCase()); }
    private Map<String, Object> pageData(int page, int limit, Object rows, long total) { Map<String, Object> result = new LinkedHashMap<>(); result.put("total", total); result.put("page", page); result.put("limit", limit); result.put("records", rows); return result; }
}
