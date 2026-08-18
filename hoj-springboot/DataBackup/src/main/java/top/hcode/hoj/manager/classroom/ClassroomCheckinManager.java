package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.CheckinMapper;
import top.hcode.hoj.mapper.classroom.CheckinRecordMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.pojo.entity.classroom.Checkin;
import top.hcode.hoj.pojo.entity.classroom.CheckinRecord;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;

import javax.annotation.Resource;
import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ClassroomCheckinManager {
    @Resource private CheckinMapper checkinMapper;
    @Resource private CheckinRecordMapper recordMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private UserInfoEntityService userInfoService;
    @Resource private ClassroomAccessManager accessManager;

    @Transactional(rollbackFor = Exception.class)
    public Checkin create(Map<String, Object> request) throws Exception {
        Long classroomId = number(request.get("classroomId"));
        accessManager.requireManager(classroomId);
        Date start = parseDate(text(request.get("startTime")));
        Date end = request.get("endTime") == null || text(request.get("endTime")).isEmpty() ? null : parseDate(text(request.get("endTime")));
        Checkin row = new Checkin().setClassroomId(classroomId).setCheckinCode(code())
                .setCheckinName(text(request.get("checkinName"))).setCheckinType(text(request.get("checkinType")).isEmpty() ? "code" : text(request.get("checkinType")))
                .setQrcodeRefreshInterval(request.get("qrcodeRefreshInterval") == null ? 15 : number(request.get("qrcodeRefreshInterval")).intValue())
                .setStartTime(start).setEndTime(end).setStatus(1).setCreateTime(new Date());
        checkinMapper.insert(row);
        return row;
    }

    public List<Checkin> list(Long classroomId, boolean studentView) throws Exception {
        if (studentView && !accessManager.isStudent(classroomId, accessManager.uid()) && !accessManager.isAdmin())
            throw new top.hcode.hoj.common.exception.StatusForbiddenException("你不在该班级中");
        if (!studentView) accessManager.requireManager(classroomId);
        List<Checkin> rows = checkinMapper.selectList(new QueryWrapper<Checkin>()
                .eq("classroom_id", classroomId).orderByDesc("create_time"));
        if (!studentView) return rows;
        Set<Long> ids = rows.stream().map(Checkin::getId).collect(Collectors.toSet());
        Map<Long, String> statuses = ids.isEmpty() ? Collections.emptyMap() : recordMapper.selectList(new QueryWrapper<CheckinRecord>()
                .in("checkin_id", ids).eq("uid", accessManager.uid())).stream()
                .collect(Collectors.toMap(CheckinRecord::getCheckinId, CheckinRecord::getStatus));
        return rows.stream().map(row -> row.setCheckinCode(null).setQrcodeToken(null).setQrcodeExpiresAt(null)).collect(Collectors.toList());
    }

    public Long classroomId(Long checkinId) {
        Checkin row = checkinMapper.selectById(checkinId);
        if (row == null) throw new IllegalArgumentException("签到不存在");
        return row.getClassroomId();
    }

    public List<Map<String, Object>> studentList(Long classroomId) throws Exception {
        List<Checkin> rows = list(classroomId, true);
        Map<Long, String> statuses = rows.isEmpty() ? Collections.emptyMap() : recordMapper.selectList(new QueryWrapper<CheckinRecord>()
                .in("checkin_id", rows.stream().map(Checkin::getId).collect(Collectors.toList())).eq("uid", accessManager.uid())).stream()
                .collect(Collectors.toMap(CheckinRecord::getCheckinId, CheckinRecord::getStatus));
        return rows.stream().map(row -> {
            Map<String, Object> item = new LinkedHashMap<>(); item.put("id", row.getId()); item.put("classroomId", row.getClassroomId());
            item.put("checkinName", row.getCheckinName()); item.put("checkinType", row.getCheckinType()); item.put("qrcodeRefreshInterval", row.getQrcodeRefreshInterval());
            item.put("startTime", row.getStartTime()); item.put("endTime", row.getEndTime()); item.put("status", row.getStatus()); item.put("createdAt", row.getCreateTime());
            item.put("userCheckinStatus", statuses.get(row.getId())); return item;
        }).collect(Collectors.toList());
    }

    @Transactional(rollbackFor = Exception.class)
    public CheckinRecord checkin(Map<String, Object> request) throws Exception {
        Long id = number(request.get("checkinId"));
        Checkin row = checkinMapper.selectOne(new QueryWrapper<Checkin>().eq("id", id).eq("checkin_code", text(request.get("checkinCode"))));
        if (row == null) throw new StatusNotFoundException("签到码不存在");
        Date now = new Date(); if (!Objects.equals(row.getStatus(), 1)) throw new StatusFailException("签到已结束");
        if (now.before(row.getStartTime())) throw new StatusFailException("签到未开始");
        if (row.getEndTime() != null && now.after(row.getEndTime())) throw new StatusFailException("签到已结束");
        if (recordMapper.selectCount(new QueryWrapper<CheckinRecord>().eq("checkin_id", id).eq("uid", accessManager.uid())) > 0)
            throw new StatusFailException("已签到");
        CheckinRecord record = new CheckinRecord().setCheckinId(id).setUid(accessManager.uid()).setStatus("present")
                .setCheckinTime(now).setCreateTime(now).setUpdateTime(now); recordMapper.insert(record); return record;
    }

    public List<Map<String, Object>> records(Long checkinId) throws Exception {
        Checkin checkin = require(checkinId); accessManager.requireManager(checkin.getClassroomId());
        List<CheckinRecord> records = recordMapper.selectList(new QueryWrapper<CheckinRecord>().eq("checkin_id", checkinId));
        Set<String> uids = records.stream().map(CheckinRecord::getUid).collect(Collectors.toSet());
        Map<String, UserInfo> users = uids.isEmpty() ? Collections.emptyMap() : userInfoService.listByIds(uids).stream().collect(Collectors.toMap(UserInfo::getUuid, u -> u));
        Map<String, ClassroomStudent> students = uids.isEmpty() ? Collections.emptyMap() : studentMapper.selectList(new QueryWrapper<ClassroomStudent>().eq("classroom_id", checkin.getClassroomId()).in("uid", uids)).stream().collect(Collectors.toMap(ClassroomStudent::getUid, s -> s));
        return records.stream().map(row -> {
            Map<String, Object> item = new LinkedHashMap<>(); item.put("id", row.getId()); item.put("checkinId", row.getCheckinId()); item.put("uid", row.getUid()); item.put("status", row.getStatus()); item.put("checkinTime", row.getCheckinTime()); item.put("remark", row.getRemark()); item.put("createTime", row.getCreateTime()); item.put("updateTime", row.getUpdateTime());
            UserInfo user = users.get(row.getUid()); if (user != null) item.put("student", userView(user));
            ClassroomStudent student = students.get(row.getUid()); if (student != null) item.put("classStudent", student);
            return item;
        }).collect(Collectors.toList());
    }

    public void addRecord(Long checkinId, Map<String, Object> request) throws Exception {
        Checkin checkin = require(checkinId); accessManager.requireManager(checkin.getClassroomId());
        String uid = text(request.get("uid")); if (uid.isEmpty()) throw new StatusFailException("uid不能为空");
        if (recordMapper.selectCount(new QueryWrapper<CheckinRecord>().eq("checkin_id", checkinId).eq("uid", uid)) > 0)
            throw new StatusFailException("该用户已有签到记录");
        recordMapper.insert(new CheckinRecord().setCheckinId(checkinId).setUid(uid).setStatus(text(request.get("status")))
                .setRemark(text(request.get("remark"))).setCreateTime(new Date()).setUpdateTime(new Date()));
    }

    public void updateRecord(Map<String, Object> request) throws Exception {
        CheckinRecord row = recordMapper.selectById(number(request.get("recordId"))); if (row == null) throw new StatusNotFoundException("记录不存在");
        accessManager.requireManager(require(row.getCheckinId()).getClassroomId());
        recordMapper.update(null, new UpdateWrapper<CheckinRecord>().eq("id", row.getId()).set("status", text(request.get("status"))).set("remark", text(request.get("remark"))).set("update_time", new Date()));
    }

    public void end(Long checkinId) throws Exception { accessManager.requireManager(require(checkinId).getClassroomId()); checkinMapper.update(null, new UpdateWrapper<Checkin>().eq("id", checkinId).set("status", 2)); }

    public void update(Long checkinId, Map<String, Object> request) throws Exception {
        Checkin row = require(checkinId); accessManager.requireManager(row.getClassroomId());
        UpdateWrapper<Checkin> update = new UpdateWrapper<Checkin>().eq("id", checkinId).set("checkin_name", text(request.get("checkinName"))).set("start_time", parseDate(text(request.get("startTime")))).set("end_time", request.get("endTime") == null || text(request.get("endTime")).isEmpty() ? null : parseDate(text(request.get("endTime"))));
        checkinMapper.update(null, update);
    }

    @Transactional(rollbackFor = Exception.class)
    public void delete(Long checkinId) throws Exception { accessManager.requireManager(require(checkinId).getClassroomId()); recordMapper.delete(new QueryWrapper<CheckinRecord>().eq("checkin_id", checkinId)); checkinMapper.deleteById(checkinId); }

    private Checkin require(Long id) throws StatusNotFoundException { Checkin row = checkinMapper.selectById(id); if (row == null) throw new StatusNotFoundException("签到不存在"); return row; }
    private ClassroomUserVO userView(UserInfo user) { return new ClassroomUserVO().setUid(user.getUuid()).setUuid(user.getUuid()).setUsername(user.getUsername()).setNickname(user.getNickname()).setRealname(user.getRealname()).setEmail(user.getEmail()); }
    private String code() { return String.valueOf(100000 + new Random().nextInt(900000)); }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private Long number(Object value) throws StatusFailException { if (!(value instanceof Number)) throw new StatusFailException("参数不能为空"); return ((Number) value).longValue(); }
    private Date parseDate(String value) throws StatusFailException { try { return Date.from(OffsetDateTime.parse(value, DateTimeFormatter.ISO_DATE_TIME).toInstant()); } catch (Exception e) { throw new StatusFailException("时间格式错误"); } }
}
