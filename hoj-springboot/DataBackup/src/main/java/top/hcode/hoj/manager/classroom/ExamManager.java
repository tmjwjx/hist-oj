package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomHomeworkMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.ExamViolationLogMapper;
import top.hcode.hoj.mapper.classroom.HomeworkQuestionMapper;
import top.hcode.hoj.mapper.classroom.HomeworkSubmitMapper;
import top.hcode.hoj.mapper.classroom.QuestionBankMapper;
import top.hcode.hoj.pojo.entity.classroom.ClassroomHomework;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.classroom.ExamViolationLog;
import top.hcode.hoj.pojo.entity.classroom.HomeworkQuestion;
import top.hcode.hoj.pojo.entity.classroom.HomeworkSubmit;
import top.hcode.hoj.pojo.entity.classroom.QuestionBank;
import top.hcode.hoj.pojo.entity.user.UserInfo;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ExamManager {
    @Resource private ClassroomHomeworkMapper homeworkMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private HomeworkQuestionMapper questionMapper;
    @Resource private HomeworkSubmitMapper submitMapper;
    @Resource private QuestionBankMapper bankMapper;
    @Resource private ExamViolationLogMapper violationMapper;
    @Resource private UserInfoEntityService userService;
    @Resource private ClassroomAccessManager access;

    public Map<String, Object> start(Long homeworkId, Map<String, Object> request) throws Exception {
        ClassroomHomework homework = requireExam(homeworkId);
        if (!access.isStudent(homework.getClassroomId(), access.uid())) {
            throw new StatusForbiddenException("只有班级学生可以开始考试");
        }
        Date now = new Date();
        if (now.before(homework.getStartTime()) || now.after(homework.getEndTime())) {
            throw new StatusFailException("当前不在考试时间内");
        }
        HomeworkSubmit marker = marker(homeworkId, access.uid());
        if (marker == null) {
            marker = new HomeworkSubmit().setHomeworkId(homeworkId).setUid(access.uid())
                    .setExamStartTime(now).setExamEndTime(deadline(homework, now))
                    .setDeviceInfo(text(request.get("deviceInfo"))).setBrowserInfo(text(request.get("browserInfo")))
                    .setIsOfficiallySubmitted(0).setIsForcedSubmit(0).setCreateTime(now).setUpdateTime(now);
            submitMapper.insert(marker);
        }
        return state(homework, marker, submitted(homeworkId, access.uid(), marker));
    }

    public Map<String, Object> status(Long homeworkId) throws Exception {
        ClassroomHomework homework = requireExam(homeworkId);
        if (!access.isStudent(homework.getClassroomId(), access.uid())
                && !access.canManage(access.requireClassroom(homework.getClassroomId()), access.uid())) {
            throw new StatusForbiddenException("无权查看考试状态");
        }
        HomeworkSubmit record = examRecord(homeworkId, access.uid());
        return state(homework, record, submitted(homeworkId, access.uid(), record));
    }

    public void violation(Map<String, Object> request) throws Exception {
        Long homeworkId = number(request.get("homeworkId"));
        ClassroomHomework homework = requireExam(homeworkId);
        if (!access.isStudent(homework.getClassroomId(), access.uid())) {
            throw new StatusForbiddenException("无权记录考试行为");
        }
        String details = request.containsKey("description") ? text(request.get("description")) : text(request.get("details"));
        ExamViolationLog row = new ExamViolationLog().setHomeworkId(homeworkId).setUid(access.uid())
                .setViolationType(text(request.get("violationType"))).setDetails(details).setCreateTime(new Date());
        violationMapper.insert(row);
        String field = counterField(row.getViolationType());
        if (field != null) {
            submitMapper.update(null, new UpdateWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                    .eq("uid", access.uid()).isNull("question_id").isNull("problem_id")
                    .setSql(field + " = COALESCE(" + field + ", 0) + 1"));
        }
    }

    public Map<String, Object> monitoring(Long homeworkId) throws Exception {
        ClassroomHomework homework = requireExam(homeworkId);
        access.requireManager(homework.getClassroomId());
        List<ClassroomStudent> students = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", homework.getClassroomId()).eq("status", 1).orderByAsc("id"));
        List<HomeworkSubmit> rows = submitMapper.selectList(new QueryWrapper<HomeworkSubmit>()
                .eq("homework_id", homeworkId).orderByAsc("id"));
        Map<String, HomeworkSubmit> records = records(rows);
        Map<String, Boolean> submitted = submitted(rows, records);
        Map<String, UserInfo> users = users(students);
        Map<String, List<ExamViolationLog>> violations = violations(homeworkId);

        int startedCount = 0, submittedCount = 0, inProgressCount = 0, notStartedCount = 0;
        List<Map<String, Object>> studentList = new ArrayList<>();
        for (ClassroomStudent student : students) {
            HomeworkSubmit record = records.get(student.getUid());
            boolean hasStarted = record != null && record.getExamStartTime() != null;
            boolean hasSubmitted = hasStarted && Boolean.TRUE.equals(submitted.get(student.getUid()));
            if (!hasStarted) notStartedCount++;
            else {
                startedCount++;
                if (hasSubmitted) submittedCount++; else inProgressCount++;
            }
            studentList.add(studentStatus(homework, student, users.get(student.getUid()), record,
                    hasSubmitted, violations.getOrDefault(student.getUid(), Collections.emptyList())));
        }

        Map<String, Object> result = new LinkedHashMap<>();
        result.put("totalStudents", students.size());
        result.put("startedCount", startedCount);
        result.put("submittedCount", submittedCount);
        result.put("inProgressCount", inProgressCount);
        result.put("notStartedCount", notStartedCount);
        result.put("studentList", studentList);
        return result;
    }

    @Transactional(rollbackFor = Exception.class)
    public void force(Long homeworkId, String uid, String reason) throws Exception {
        ClassroomHomework homework = requireExam(homeworkId);
        access.requireManager(homework.getClassroomId());
        List<HomeworkSubmit> rows = rows(homeworkId, uid);
        HomeworkSubmit record = records(rows).get(uid);
        if (record == null) throw new StatusFailException("该学生尚未开始考试");
        if (submitted(homeworkId, uid, record)) throw new StatusFailException("该学生已经交卷");
        if (forceRows(homeworkId, uid) == 0) throw new StatusFailException("该学生没有未提交的答卷");
        logForced(homeworkId, uid, blank(reason) ? "教师强制收卷" : reason);
    }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> forceAll(Long homeworkId) throws Exception {
        ClassroomHomework homework = requireExam(homeworkId);
        access.requireManager(homework.getClassroomId());
        List<HomeworkSubmit> rows = submitMapper.selectList(new QueryWrapper<HomeworkSubmit>()
                .eq("homework_id", homeworkId).orderByAsc("id"));
        Map<String, HomeworkSubmit> records = records(rows);
        Map<String, Boolean> submitted = submitted(rows, records);
        int forcedCount = 0;
        for (Map.Entry<String, HomeworkSubmit> item : records.entrySet()) {
            if (Boolean.TRUE.equals(submitted.get(item.getKey()))) continue;
            if (forceRows(homeworkId, item.getKey()) > 0) {
                forcedCount++;
                logForced(homeworkId, item.getKey(), "教师强制收卷");
            }
        }
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("forcedCount", forcedCount);
        return result;
    }

    private Map<String, Object> studentStatus(ClassroomHomework homework, ClassroomStudent student,
                                               UserInfo user, HomeworkSubmit record, boolean submitted,
                                               List<ExamViolationLog> logs) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("uid", student.getUid());
        value.put("name", first(student.getRealName(), user == null ? null : user.getRealname(),
                user == null ? null : user.getUsername(), student.getUid()));
        value.put("username", first(user == null ? null : user.getUsername(), student.getUid()));
        value.put("examStartTime", record == null ? null : record.getExamStartTime());
        value.put("elapsedMinutes", elapsedMinutes(homework, record, submitted));
        value.put("remainingSeconds", record == null || submitted ? (record == null ? -1 : 0) : remainingSeconds(homework, record));
        value.put("status", status(record, submitted));
        value.put("violationCount", logs.size());
        value.put("violations", violationSummary(logs));
        return value;
    }

    private Map<String, Object> state(ClassroomHomework homework, HomeworkSubmit record, boolean submitted) {
        boolean started = record != null && record.getExamStartTime() != null;
        int remaining = started ? remainingSeconds(homework, record) : 0;
        boolean forced = record != null && Objects.equals(record.getIsForcedSubmit(), 1);
        long elapsed = started ? Math.max(0L, new Date().getTime() - record.getExamStartTime().getTime()) : 0L;
        int allowAfter = homework.getAllowSubmitAfterMinutes() == null ? 0 : homework.getAllowSubmitAfterMinutes();
        boolean canSubmit = started && !submitted && remaining > 0 && elapsed >= allowAfter * 60_000L;

        Map<String, Object> result = new LinkedHashMap<>();
        result.put("homeworkId", homework.getId());
        result.put("isExamMode", homework.getIsExamMode());
        result.put("started", started);
        result.put("hasStarted", started);
        result.put("submitted", submitted);
        result.put("isSubmitted", submitted);
        result.put("isForcedSubmit", forced);
        result.put("remainingSeconds", remaining);
        result.put("isOvertime", started && remaining <= 0);
        result.put("canSubmit", canSubmit);
        result.put("questions", started ? questions(homework.getId()) : Collections.emptyList());
        if (started) {
            result.put("examStartTime", record.getExamStartTime());
            result.put("examEndTime", submitted && record.getExamEndTime() != null
                    ? record.getExamEndTime() : deadline(homework, record.getExamStartTime()));
        }
        Map<String, Object> config = new LinkedHashMap<>();
        config.put("examDuration", homework.getExamDuration());
        config.put("allowSubmitAfterMinutes", homework.getAllowSubmitAfterMinutes());
        config.put("disableCopyPaste", homework.getDisableCopyPaste());
        config.put("requireFullscreen", homework.getRequireFullscreen());
        config.put("disallowTabSwitch", homework.getDisallowTabSwitch());
        result.put("examConfig", config);
        return result;
    }

    private List<Map<String, Object>> questions(Long homeworkId) {
        List<HomeworkQuestion> relations = questionMapper.selectList(new QueryWrapper<HomeworkQuestion>()
                .eq("homework_id", homeworkId).orderByAsc("question_order"));
        Set<Long> bankIds = relations.stream().map(HomeworkQuestion::getQuestionId)
                .filter(Objects::nonNull).collect(Collectors.toSet());
        Map<Long, QuestionBank> banks = bankIds.isEmpty() ? Collections.emptyMap()
                : bankMapper.selectBatchIds(bankIds).stream().collect(Collectors.toMap(
                        QuestionBank::getId, value -> value, (left, right) -> left, LinkedHashMap::new));
        List<Map<String, Object>> result = new ArrayList<>();
        for (HomeworkQuestion relation : relations) {
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("id", relation.getId()); item.put("homeworkQuestionId", relation.getId());
            item.put("questionId", relation.getQuestionId()); item.put("problemId", relation.getProblemId());
            item.put("questionOrder", relation.getQuestionOrder()); item.put("displayOrder", relation.getQuestionOrder());
            item.put("score", relation.getScore());
            QuestionBank bank = banks.get(relation.getQuestionId());
            if (bank != null) item.put("question", questionView(bank));
            result.add(item);
        }
        return result;
    }

    private Map<String, Object> questionView(QuestionBank question) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("id", question.getId()); value.put("title", question.getTitle());
        value.put("type", question.getType()); value.put("content", question.getContent());
        value.put("options", question.getOptions()); value.put("difficulty", question.getDifficulty());
        return value;
    }

    private Map<String, HomeworkSubmit> records(List<HomeworkSubmit> rows) {
        Map<String, HomeworkSubmit> result = new LinkedHashMap<>();
        for (HomeworkSubmit row : rows) {
            if (row.getExamStartTime() == null) continue;
            HomeworkSubmit current = result.get(row.getUid());
            if (current == null || isMarker(row) || !isMarker(current)) result.put(row.getUid(), row);
        }
        return result;
    }

    private Map<String, Boolean> submitted(List<HomeworkSubmit> rows, Map<String, HomeworkSubmit> records) {
        Map<String, Boolean> result = new HashMap<>();
        for (Map.Entry<String, HomeworkSubmit> item : records.entrySet()) {
            if (isMarker(item.getValue())) result.put(item.getKey(), Objects.equals(item.getValue().getIsOfficiallySubmitted(), 1));
        }
        for (HomeworkSubmit row : rows) {
            if (Objects.equals(row.getIsOfficiallySubmitted(), 1)) result.put(row.getUid(), true);
        }
        return result;
    }

    private boolean submitted(Long homeworkId, String uid, HomeworkSubmit record) {
        if (record != null && isMarker(record) && Objects.equals(record.getIsOfficiallySubmitted(), 1)) return true;
        return submitMapper.selectCount(new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                .eq("uid", uid).eq("is_officially_submitted", 1)) > 0;
    }

    private HomeworkSubmit examRecord(Long homeworkId, String uid) {
        HomeworkSubmit marker = marker(homeworkId, uid);
        if (marker != null) return marker;
        return submitMapper.selectOne(new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                .eq("uid", uid).isNotNull("exam_start_time").orderByDesc("id").last("limit 1"));
    }

    private HomeworkSubmit marker(Long homeworkId, String uid) {
        return submitMapper.selectOne(new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                .eq("uid", uid).isNull("question_id").isNull("problem_id").orderByDesc("id").last("limit 1"));
    }

    private List<HomeworkSubmit> rows(Long homeworkId, String uid) {
        return submitMapper.selectList(new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                .eq("uid", uid).orderByAsc("id"));
    }

    private int forceRows(Long homeworkId, String uid) {
        Date now = new Date();
        return submitMapper.update(null, new UpdateWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                .eq("uid", uid).eq("is_officially_submitted", 0).set("is_officially_submitted", 1)
                .set("is_forced_submit", 1).set("exam_end_time", now).set("update_time", now));
    }

    private void logForced(Long homeworkId, String uid, String reason) {
        violationMapper.insert(new ExamViolationLog().setHomeworkId(homeworkId).setUid(uid)
                .setViolationType("forced_submit").setDetails(reason).setCreateTime(new Date()));
    }

    private Map<String, List<ExamViolationLog>> violations(Long homeworkId) {
        return violationMapper.selectList(new QueryWrapper<ExamViolationLog>().eq("homework_id", homeworkId)
                .orderByAsc("id")).stream().collect(Collectors.groupingBy(ExamViolationLog::getUid, LinkedHashMap::new, Collectors.toList()));
    }

    private List<Map<String, Object>> violationSummary(List<ExamViolationLog> rows) {
        Map<String, Integer> counts = new LinkedHashMap<>();
        for (ExamViolationLog row : rows) counts.merge(row.getViolationType(), 1, Integer::sum);
        List<Map<String, Object>> result = new ArrayList<>();
        for (Map.Entry<String, Integer> item : counts.entrySet()) {
            Map<String, Object> value = new LinkedHashMap<>();
            value.put("type", item.getKey());
            value.put("count", item.getValue());
            result.add(value);
        }
        return result;
    }

    private Map<String, UserInfo> users(List<ClassroomStudent> students) {
        Set<String> ids = students.stream().map(ClassroomStudent::getUid).collect(Collectors.toSet());
        if (ids.isEmpty()) return Collections.emptyMap();
        return userService.listByIds(ids).stream().collect(Collectors.toMap(
                UserInfo::getUuid, row -> row, (left, right) -> left, LinkedHashMap::new));
    }

    private int remainingSeconds(ClassroomHomework homework, HomeworkSubmit record) {
        if (record == null || record.getExamStartTime() == null) return 0;
        long remaining = deadline(homework, record.getExamStartTime()).getTime() - System.currentTimeMillis();
        return remaining <= 0 ? 0 : (int) Math.min(Integer.MAX_VALUE, remaining / 1000L);
    }

    private int elapsedMinutes(ClassroomHomework homework, HomeworkSubmit record, boolean submitted) {
        if (record == null || record.getExamStartTime() == null) return -1;
        Date end = submitted && record.getExamEndTime() != null ? record.getExamEndTime() : new Date();
        Date limit = deadline(homework, record.getExamStartTime());
        if (end.after(limit)) end = limit;
        if (end.before(record.getExamStartTime())) return 0;
        return (int) ((end.getTime() - record.getExamStartTime().getTime()) / 60_000L);
    }

    private Date deadline(ClassroomHomework homework, Date startedAt) {
        int duration = homework.getExamDuration() == null ? 60 : Math.max(1, homework.getExamDuration());
        long byDuration = startedAt.getTime() + duration * 60_000L;
        return new Date(Math.min(homework.getEndTime().getTime(), byDuration));
    }

    private String status(HomeworkSubmit record, boolean submitted) {
        if (record == null || record.getExamStartTime() == null) return "not_started";
        if (submitted && Objects.equals(record.getIsForcedSubmit(), 1)) return "forced_submit";
        return submitted ? "submitted" : "in_progress";
    }

    private boolean isMarker(HomeworkSubmit row) {
        return row != null && row.getQuestionId() == null && blank(row.getProblemId());
    }

    private ClassroomHomework requireExam(Long id) throws Exception {
        ClassroomHomework row = require(id);
        if (!Objects.equals(row.getIsExamMode(), 1)) throw new StatusFailException("当前作业不是考试模式");
        return row;
    }

    private ClassroomHomework require(Long id) throws StatusNotFoundException {
        ClassroomHomework row = id == null ? null : homeworkMapper.selectById(id);
        if (row == null) throw new StatusNotFoundException("作业不存在");
        return row;
    }

    private String counterField(String type) {
        if ("tab_switch".equals(type)) return "tab_switch_count";
        if ("fullscreen_exit".equals(type)) return "fullscreen_exit_count";
        if ("copy_paste".equals(type) || "copy_attempt".equals(type) || "paste_attempt".equals(type)) return "copy_paste_attempt_count";
        return null;
    }

    private Long number(Object value) {
        if (value instanceof Number) return ((Number) value).longValue();
        try { return value == null ? null : Long.parseLong(String.valueOf(value)); }
        catch (Exception e) { return null; }
    }

    private String first(String... values) {
        for (String value : values) if (!blank(value)) return value;
        return "";
    }

    private boolean blank(String value) { return value == null || value.trim().isEmpty(); }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
}
