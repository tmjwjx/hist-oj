package top.hcode.hoj.manager.classroom;

import cn.hutool.json.JSONUtil;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomHomeworkMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.ExamViolationLogMapper;
import top.hcode.hoj.mapper.classroom.HomeworkQuestionMapper;
import top.hcode.hoj.mapper.classroom.HomeworkSubmitMapper;
import top.hcode.hoj.mapper.classroom.QuestionBankMapper;
import top.hcode.hoj.pojo.entity.classroom.*;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.user.UserInfo;

import javax.annotation.Resource;
import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class HomeworkManager {
    @Resource private ClassroomHomeworkMapper homeworkMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private HomeworkQuestionMapper questionMapper;
    @Resource private HomeworkSubmitMapper submitMapper;
    @Resource private ExamViolationLogMapper violationMapper;
    @Resource private QuestionBankMapper bankMapper;
    @Resource private ClassroomAccessManager access;
    @Resource private UserInfoEntityService userService;
    @Resource private JudgeEntityService judgeService;

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> create(Map<String, Object> request) throws Exception {
        Long classroomId = number(request.get("classroomId"));
        access.requireManager(classroomId);
        Date start = date(request.get("startTime")), end = date(request.get("endTime"));
        if (start == null || end == null || end.before(start)) throw new StatusFailException("作业时间格式错误");
        ClassroomHomework row = new ClassroomHomework().setClassroomId(classroomId).setTitle(text(request.get("title")))
                .setDescription(text(request.get("description"))).setStartTime(start).setEndTime(end)
                .setShowScore(integer(request.get("showScore"), 1)).setShowHomework(integer(request.get("showHomework"), 0))
                .setShowAnswer(integer(request.get("showAnswer"), 0)).setShowRank(integer(request.get("showRank"), 0))
                .setStatus(status(start, end)).setIsExamMode(integer(request.get("isExamMode"), 0))
                .setExamDuration(integer(request.get("examDuration"), 0)).setAllowSubmitAfterMinutes(integer(request.get("allowSubmitAfterMinutes"), 0))
                .setDisableCopyPaste(integer(request.get("disableCopyPaste"), 0)).setRequireFullscreen(integer(request.get("requireFullscreen"), 0))
                .setDisallowTabSwitch(integer(request.get("disallowTabSwitch"), 0)).setCreateTime(new Date()).setUpdateTime(new Date());
        homeworkMapper.insert(row); replaceQuestions(row.getId(), list(request.get("questions"))); return detailView(row, false);
    }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> update(Long id, Map<String, Object> request) throws Exception {
        ClassroomHomework row = manager(id);
        Date start = date(request.get("startTime")), end = date(request.get("endTime"));
        if (start == null || end == null || end.before(start)) throw new StatusFailException("作业时间格式错误");
        row.setTitle(text(request.get("title"))).setDescription(text(request.get("description"))).setStartTime(start).setEndTime(end)
                .setShowScore(integer(request.get("showScore"), row.getShowScore())).setShowHomework(integer(request.get("showHomework"), row.getShowHomework()))
                .setShowAnswer(integer(request.get("showAnswer"), row.getShowAnswer())).setShowRank(integer(request.get("showRank"), row.getShowRank()))
                .setStatus(status(start, end)).setIsExamMode(integer(request.get("isExamMode"), row.getIsExamMode()))
                .setExamDuration(integer(request.get("examDuration"), row.getExamDuration())).setAllowSubmitAfterMinutes(integer(request.get("allowSubmitAfterMinutes"), row.getAllowSubmitAfterMinutes()))
                .setDisableCopyPaste(integer(request.get("disableCopyPaste"), row.getDisableCopyPaste())).setRequireFullscreen(integer(request.get("requireFullscreen"), row.getRequireFullscreen()))
                .setDisallowTabSwitch(integer(request.get("disallowTabSwitch"), row.getDisallowTabSwitch())).setUpdateTime(new Date());
        homeworkMapper.updateById(row); replaceQuestions(id, list(request.get("questions"))); return detailView(row, false);
    }

    @Transactional(rollbackFor = Exception.class)
    public void delete(Long id) throws Exception {
        ClassroomHomework row = manager(id);
        homeworkMapper.update(null, new UpdateWrapper<ClassroomHomework>().eq("id", row.getId()).set("status", 0).set("update_time", new Date()));
    }

    public List<Map<String, Object>> list(Long classroomId) throws Exception {
        requireClassroomAccess(classroomId);
        List<ClassroomHomework> rows = homeworkMapper.selectList(new QueryWrapper<ClassroomHomework>()
                .eq("classroom_id", classroomId).ne("status", 0).orderByDesc("start_time"));
        String uid = access.uid();
        for (ClassroomHomework row : rows) row.setStatus(status(row.getStartTime(), row.getEndTime()));
        return rows.stream().map(row -> {
            Map<String, Object> value = base(row); value.put("isCompleted", submitMapper.selectCount(new QueryWrapper<HomeworkSubmit>().eq("homework_id", row.getId()).eq("uid", uid).eq("is_officially_submitted", 1)) > 0); return value;
        }).collect(Collectors.toList());
    }

    public Map<String, Object> detail(Long id) throws Exception {
        ClassroomHomework row = viewable(id); boolean student = !access.canManage(access.requireClassroom(row.getClassroomId()), access.uid());
        boolean submitted = submitMapper.selectCount(new QueryWrapper<HomeworkSubmit>().eq("homework_id", id).eq("uid", access.uid()).eq("is_officially_submitted", 1)) > 0;
        return detailView(row, student && (!submitted || row.getShowAnswer() != 1));
    }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> saveDraft(Map<String, Object> request) throws Exception { return save(request, false); }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> submit(Map<String, Object> request) throws Exception { return save(request, true); }

    public List<Map<String, Object>> submissions(Long homeworkId) throws Exception {
        ClassroomHomework homework = manager(homeworkId);
        List<HomeworkQuestion> questions = questionMapper.selectList(new QueryWrapper<HomeworkQuestion>()
                .eq("homework_id", homeworkId).orderByAsc("question_order"));
        List<ClassroomStudent> students = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", homework.getClassroomId()).eq("status", 1));
        Map<String, ClassroomStudent> studentByUid = students.stream().collect(Collectors.toMap(
                ClassroomStudent::getUid, value -> value, (left, right) -> left, LinkedHashMap::new));
        Map<String, UserInfo> users = users(studentByUid.keySet());

        List<HomeworkSubmit> submits = submitMapper.selectList(new QueryWrapper<HomeworkSubmit>()
                .eq("homework_id", homeworkId).eq("is_officially_submitted", 1)
                .orderByDesc("update_time").orderByDesc("id"));
        Map<String, HomeworkSubmit> latest = new LinkedHashMap<>();
        for (HomeworkSubmit submit : submits) {
            String identity = identity(submit);
            if (identity == null || !studentByUid.containsKey(submit.getUid())) continue;
            latest.putIfAbsent(submit.getUid() + "|" + identity, submit);
        }
        Map<String, HomeworkSubmit> examMarkers = new LinkedHashMap<>();
        for (HomeworkSubmit marker : submitMapper.selectList(new QueryWrapper<HomeworkSubmit>()
                .eq("homework_id", homeworkId).isNull("question_id").isNull("problem_id")
                .orderByDesc("id"))) {
            examMarkers.putIfAbsent(marker.getUid(), marker);
        }

        List<Map<String, Object>> result = new ArrayList<>();
        Set<String> countersAdded = new HashSet<>();
        for (HomeworkSubmit submit : latest.values()) {
            HomeworkQuestion relation = relation(questions, submit);
            Map<String, Object> value = submitView(submit);
            ClassroomStudent student = studentByUid.get(submit.getUid());
            UserInfo user = users.get(submit.getUid());
            value.put("homeworkQuestionId", relation == null ? null : relation.getId());
            value.put("realName", first(student == null ? null : student.getRealName(),
                    user == null ? null : user.getRealname(), ""));
            value.put("student", studentView(submit.getUid(), student, user));
            HomeworkSubmit marker = examMarkers.get(submit.getUid());
            if (marker != null && countersAdded.add(submit.getUid())) copyViolationCounters(value, marker);
            if (submit.getQuestionId() != null) value.put("question", questionView(bankMapper.selectById(submit.getQuestionId())));
            result.add(value);
        }
        result.sort(Comparator.comparing((Map<String, Object> item) -> text(item.get("uid")))
                .thenComparingInt(item -> relationOrder(questions, number(item.get("homeworkQuestionId")))));
        return result;
    }

    public Map<String, Object> status(Long homeworkId) throws Exception {
        ClassroomHomework row = viewable(homeworkId); List<HomeworkSubmit> submits = own(homeworkId);
        Map<String, Object> data = base(row); data.put("submissions", submits.stream().map(this::submitView).collect(Collectors.toList()));
        data.put("submitted", submits.stream().anyMatch(s -> Objects.equals(s.getIsOfficiallySubmitted(), 1))); return data;
    }

    public Map<String, Object> myDetail(Long homeworkId) throws Exception {
        viewable(homeworkId);
        List<HomeworkSubmit> rows = own(homeworkId);
        Map<String, HomeworkSubmit> latest = new LinkedHashMap<>();
        for (HomeworkSubmit row : rows) {
            String identity = identity(row);
            if (identity != null) latest.put(identity, row);
        }

        Map<String, Object> answers = new LinkedHashMap<>();
        Map<String, Object> scores = new LinkedHashMap<>();
        Map<String, Object> scored = new LinkedHashMap<>();
        Map<String, Object> attachments = new LinkedHashMap<>();
        double totalScore = 0d, gradedScore = 0d;
        boolean hasUngraded = false;
        Date submitTime = null;
        for (HomeworkSubmit row : latest.values()) {
            String key = answerKey(row);
            answers.put(key, row.getAnswer() == null ? "" : row.getAnswer());
            scores.put(key, row.getScore() == null ? 0d : row.getScore());
            scored.put(key, Objects.equals(row.getIsScored(), 1));
            if (row.getAttachment() != null && !row.getAttachment().trim().isEmpty()) {
                attachments.put(key, row.getAttachment());
            }
            totalScore += row.getScore() == null ? 0d : row.getScore();
            if (Objects.equals(row.getIsScored(), 1)) gradedScore += row.getScore() == null ? 0d : row.getScore();
            else hasUngraded = true;
            Date current = row.getUpdateTime() == null ? row.getCreateTime() : row.getUpdateTime();
            if (current != null && (submitTime == null || current.after(submitTime))) submitTime = current;
        }
        boolean official = rows.stream().anyMatch(row -> Objects.equals(row.getIsOfficiallySubmitted(), 1));
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("homeworkId", homeworkId);
        result.put("uid", access.uid());
        result.put("answers", JSONUtil.toJsonStr(answers));
        result.put("scores", JSONUtil.toJsonStr(scores));
        result.put("isScoredMap", JSONUtil.toJsonStr(scored));
        result.put("attachments", JSONUtil.toJsonStr(attachments));
        result.put("score", totalScore);
        result.put("gradedScore", gradedScore);
        result.put("hasUngraded", hasUngraded);
        result.put("submitTime", submitTime);
        result.put("isOfficiallySubmitted", official ? 1 : 0);
        return result;
    }

    @Transactional(rollbackFor = Exception.class)
    public void grade(Map<String, Object> request) throws Exception {
        manager(number(request.get("homeworkId")));
        Long submitId = number(request.get("submissionId")); if (submitId == null) submitId = number(request.get("id"));
        HomeworkSubmit row = submitMapper.selectById(submitId); if (row == null) throw new StatusNotFoundException("提交记录不存在");
        submitMapper.update(null, new UpdateWrapper<HomeworkSubmit>().eq("id", submitId).set("score", decimal(request.get("score"))).set("is_scored", 1).set("update_time", new Date()));
    }

    @Transactional(rollbackFor = Exception.class)
    public void recalculate(Map<String, Object> request) throws Exception {
        Long homeworkId = number(request.get("homeworkId")); manager(homeworkId);
        for (HomeworkSubmit row : submitMapper.selectList(new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId))) {
            if (row.getQuestionId() == null) continue; HomeworkQuestion q = questionMapper.selectOne(new QueryWrapper<HomeworkQuestion>().eq("homework_id", homeworkId).eq("question_id", row.getQuestionId())); QuestionBank bank = bankMapper.selectById(row.getQuestionId());
            if (q != null && bank != null) { Double score = objectiveScore(bank, row.getAnswer(), q.getScore()); if (score != null) submitMapper.update(null, new UpdateWrapper<HomeworkSubmit>().eq("id", row.getId()).set("score", score).set("is_scored", 1)); }
        }
    }

    public Map<String, Object> saveProgramming(Map<String, Object> request) throws Exception {
        Long homeworkId = number(request.get("homeworkId"));
        String problemId = text(request.get("problemId")).trim();
        viewable(homeworkId);
        HomeworkQuestion relation = questionMapper.selectOne(new QueryWrapper<HomeworkQuestion>()
                .eq("homework_id", homeworkId).eq("problem_id", problemId).last("limit 1"));
        if (relation == null) throw new StatusNotFoundException("作业中的编程题不存在");

        QueryWrapper<HomeworkSubmit> query = new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId)
                .eq("uid", access.uid()).eq("problem_id", problemId).orderByDesc("id").last("limit 1");
        HomeworkSubmit row = submitMapper.selectOne(query);
        Date now = new Date();
        Map<String, Object> answer = new LinkedHashMap<>();
        answer.put("submitId", number(request.get("submitId")));
        answer.put("code", text(request.get("code")));
        answer.put("language", text(request.get("language")));
        if (row == null) {
            row = new HomeworkSubmit().setHomeworkId(homeworkId).setProblemId(problemId).setUid(access.uid())
                    .setIsOfficiallySubmitted(0).setScore(0d).setCreateTime(now);
        }
        row.setSubmitId(number(request.get("submitId"))).setAnswer(JSONUtil.toJsonStr(answer))
                .setIsScored(0).setJudgeResult(null).setUpdateTime(now);
        if (row.getId() == null) submitMapper.insert(row); else submitMapper.updateById(row);
        return submitView(row);
    }

    public List<Map<String, Object>> programmingSubmissions(Map<String, String> params) throws Exception {
        Long homeworkId = number(params.get("homeworkId")); viewable(homeworkId);
        Long relationId = number(params.get("homeworkQuestionId"));
        HomeworkQuestion relation = relationId == null ? null : questionMapper.selectById(relationId);
        if (relation == null || !Objects.equals(relation.getHomeworkId(), homeworkId)) {
            throw new StatusNotFoundException("作业题目不存在");
        }
        QueryWrapper<HomeworkSubmit> q = new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId).eq("uid", access.uid()).isNotNull("submit_id").orderByDesc("create_time");
        if (relation != null && relation.getProblemId() != null) q.eq("problem_id", relation.getProblemId());
        List<Map<String, Object>> result = new ArrayList<>();
        for (HomeworkSubmit row : submitMapper.selectList(q)) {
            Judge judge = row.getSubmitId() == null ? null : judgeService.getById(row.getSubmitId());
            if (judge != null) syncProgrammingResult(row, relation, judge);
            Map<String, Object> value = submitView(row);
            try {
                cn.hutool.json.JSONObject answer = JSONUtil.parseObj(row.getAnswer());
                value.put("code", answer.getStr("code", ""));
                value.put("language", answer.getStr("language", ""));
            } catch (Exception ignored) { }
            value.put("submitTime", row.getUpdateTime() == null ? row.getCreateTime() : row.getUpdateTime());
            if (judge != null) { value.put("result", judgeResult(judge.getStatus())); value.put("status", judge.getStatus()); value.put("code", judge.getCode()); value.put("language", judge.getLanguage()); }
            else value.put("result", row.getJudgeResult() == null ? "PENDING" : row.getJudgeResult());
            result.add(value);
        }
        return result;
    }

    private Map<String, Object> save(Map<String, Object> request, boolean official) throws Exception {
        Long homeworkId = number(request.get("homeworkId")); ClassroomHomework homework = viewable(homeworkId);
        Date now = new Date();
        if (official && (now.before(homework.getStartTime()) || now.after(homework.getEndTime()))) throw new StatusFailException("当前不在作业提交时间内");
        if (official && Objects.equals(homework.getIsExamMode(), 1) && homework.getAllowSubmitAfterMinutes() != null
                && now.getTime() - homework.getStartTime().getTime() < homework.getAllowSubmitAfterMinutes() * 60_000L) {
            throw new StatusFailException("考试开始后暂不允许交卷");
        }
        Map<String, Object> answers = request.get("answers") instanceof Map ? (Map<String, Object>) request.get("answers") : Collections.emptyMap();
        Map<String, Object> attachments = request.get("attachments") instanceof Map ? (Map<String, Object>) request.get("attachments") : Collections.emptyMap();
        for (HomeworkQuestion question : questionMapper.selectList(new QueryWrapper<HomeworkQuestion>().eq("homework_id", homeworkId))) {
            String key = question.getQuestionId() != null ? String.valueOf(question.getQuestionId()) : question.getProblemId(); if (!answers.containsKey(key) && !attachments.containsKey(key)) continue;
            HomeworkSubmit row = ownQuestion(homeworkId, question); String answer = text(answers.get(key)); QuestionBank bank = question.getQuestionId() == null ? null : bankMapper.selectById(question.getQuestionId()); Double score = bank == null ? null : objectiveScore(bank, answer, question.getScore());
            if (row == null) row = new HomeworkSubmit().setHomeworkId(homeworkId).setQuestionId(question.getQuestionId()).setProblemId(question.getProblemId()).setUid(access.uid()).setCreateTime(new Date());
            row.setAnswer(answer).setAttachment(text(attachments.get(key))).setIsOfficiallySubmitted(official ? 1 : 0).setIsScored(score == null ? 0 : 1).setScore(score == null ? 0d : score).setUpdateTime(new Date()); if (row.getId() == null) submitMapper.insert(row); else submitMapper.updateById(row);
        }
        if (official) {
            UpdateWrapper<HomeworkSubmit> finalise = new UpdateWrapper<HomeworkSubmit>()
                    .eq("homework_id", homeworkId).eq("uid", access.uid()).eq("is_officially_submitted", 0)
                    .set("is_officially_submitted", 1).set("update_time", now);
            if (Objects.equals(homework.getIsExamMode(), 1)) finalise.set("exam_end_time", now);
            submitMapper.update(null, finalise);
        }
        Map<String, Object> result = new LinkedHashMap<>(); result.put("homeworkId", homeworkId); result.put("submitted", official); return result;
    }

    private Map<String, Object> detailView(ClassroomHomework row, boolean hideAnswers) {
        Map<String, Object> data = base(row); List<Map<String, Object>> questions = new ArrayList<>();
        for (HomeworkQuestion relation : questionMapper.selectList(new QueryWrapper<HomeworkQuestion>().eq("homework_id", row.getId()).orderByAsc("question_order"))) {
            Map<String, Object> item = new LinkedHashMap<>(); item.put("id", relation.getId()); item.put("questionId", relation.getQuestionId()); item.put("problemId", relation.getProblemId()); item.put("questionOrder", relation.getQuestionOrder()); item.put("score", relation.getScore());
            if (relation.getQuestionId() != null) { QuestionBank q = bankMapper.selectById(relation.getQuestionId()); if (q != null) { Map<String, Object> v = new LinkedHashMap<>(); v.put("id", q.getId()); v.put("title", q.getTitle()); v.put("type", q.getType()); v.put("content", q.getContent()); v.put("options", q.getOptions()); v.put("difficulty", q.getDifficulty()); if (!hideAnswers) { v.put("answer", q.getAnswer()); v.put("analysis", q.getAnalysis()); } item.put("question", v); } }
            questions.add(item);
        }
        data.put("questions", questions); return data;
    }
    private Map<String, Object> base(ClassroomHomework row) { Map<String, Object> r = new LinkedHashMap<>(); r.put("id", row.getId()); r.put("classroomId", row.getClassroomId()); r.put("title", row.getTitle()); r.put("description", row.getDescription()); r.put("startTime", row.getStartTime()); r.put("endTime", row.getEndTime()); r.put("showScore", row.getShowScore()); r.put("showHomework", row.getShowHomework()); r.put("showAnswer", row.getShowAnswer()); r.put("showRank", row.getShowRank()); r.put("status", status(row.getStartTime(), row.getEndTime())); r.put("isExamMode", row.getIsExamMode()); r.put("examDuration", row.getExamDuration()); r.put("allowSubmitAfterMinutes", row.getAllowSubmitAfterMinutes()); r.put("disableCopyPaste", row.getDisableCopyPaste()); r.put("requireFullscreen", row.getRequireFullscreen()); r.put("disallowTabSwitch", row.getDisallowTabSwitch()); return r; }
    private Map<String, Object> submitView(HomeworkSubmit row) {
        Map<String, Object> r = new LinkedHashMap<>();
        r.put("id", row.getId()); r.put("homeworkId", row.getHomeworkId());
        r.put("questionId", row.getQuestionId()); r.put("problemId", row.getProblemId());
        r.put("uid", row.getUid()); r.put("answer", row.getAnswer()); r.put("attachment", row.getAttachment());
        r.put("submitId", row.getSubmitId()); r.put("score", row.getScore()); r.put("isScored", row.getIsScored());
        r.put("isOfficiallySubmitted", row.getIsOfficiallySubmitted()); r.put("judgeResult", row.getJudgeResult());
        copyViolationCounters(r, row);
        r.put("createTime", row.getCreateTime()); r.put("createdAt", row.getCreateTime());
        r.put("updateTime", row.getUpdateTime()); r.put("updatedAt", row.getUpdateTime());
        return r;
    }

    private void copyViolationCounters(Map<String, Object> target, HomeworkSubmit row) {
        target.put("tabSwitchCount", row.getTabSwitchCount() == null ? 0 : row.getTabSwitchCount());
        target.put("fullscreenExitCount", row.getFullscreenExitCount() == null ? 0 : row.getFullscreenExitCount());
        target.put("copyPasteAttemptCount", row.getCopyPasteAttemptCount() == null ? 0 : row.getCopyPasteAttemptCount());
    }

    private Map<String, Object> studentView(String uid, ClassroomStudent student, UserInfo user) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("uid", uid);
        value.put("username", first(user == null ? null : user.getUsername(), uid));
        value.put("realName", first(student == null ? null : student.getRealName(),
                user == null ? null : user.getRealname(), ""));
        value.put("nickname", first(user == null ? null : user.getNickname(), ""));
        return value;
    }

    private Map<String, Object> questionView(QuestionBank question) {
        if (question == null) return null;
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("id", question.getId()); value.put("title", question.getTitle());
        value.put("type", question.getType()); value.put("content", question.getContent());
        value.put("options", question.getOptions()); value.put("answer", question.getAnswer());
        value.put("analysis", question.getAnalysis()); value.put("difficulty", question.getDifficulty());
        return value;
    }

    private Map<String, UserInfo> users(Collection<String> ids) {
        if (ids.isEmpty()) return Collections.emptyMap();
        return userService.listByIds(ids).stream().collect(Collectors.toMap(
                UserInfo::getUuid, value -> value, (left, right) -> left, LinkedHashMap::new));
    }

    private HomeworkQuestion relation(List<HomeworkQuestion> relations, HomeworkSubmit submit) {
        return relations.stream().filter(value -> matches(value, submit)).findFirst().orElse(null);
    }

    private boolean matches(HomeworkQuestion relation, HomeworkSubmit submit) {
        if (relation.getQuestionId() != null) return Objects.equals(relation.getQuestionId(), submit.getQuestionId());
        return relation.getProblemId() != null && Objects.equals(relation.getProblemId(), submit.getProblemId());
    }

    private String identity(HomeworkSubmit row) {
        if (row.getProblemId() != null && !row.getProblemId().trim().isEmpty()) return "p:" + row.getProblemId().trim();
        return row.getQuestionId() == null ? null : "q:" + row.getQuestionId();
    }

    private String answerKey(HomeworkSubmit row) {
        return row.getQuestionId() == null ? row.getProblemId() : String.valueOf(row.getQuestionId());
    }

    private int relationOrder(List<HomeworkQuestion> questions, Long id) {
        if (id == null) return Integer.MAX_VALUE;
        for (HomeworkQuestion question : questions) {
            if (Objects.equals(question.getId(), id)) return question.getQuestionOrder() == null ? Integer.MAX_VALUE : question.getQuestionOrder();
        }
        return Integer.MAX_VALUE;
    }
    private void syncProgrammingResult(HomeworkSubmit row, HomeworkQuestion relation, Judge judge) { if (judge.getStatus() == null || (judge.getStatus() >= 5 && judge.getStatus() != 10) || relation == null) return; double score = judge.getStatus() == 0 ? relation.getScore() : judge.getStatus() == 8 && judge.getScore() != null ? relation.getScore() * Math.max(0, Math.min(100, judge.getScore())) / 100d : 0d; row.setScore(score).setJudgeResult(judgeResult(judge.getStatus())).setIsScored(1); submitMapper.updateById(row); }
    private String judgeResult(Integer status) { if (status == null || (status >= 5 && status != 10)) return "PENDING"; switch (status) { case 0: return "AC"; case -2: return "CE"; case -3: return "PE"; case -1: return "WA"; case 1: return "TLE"; case 2: return "MLE"; case 3: return "RE"; case 4: return "SE"; case 8: return "PAC"; case -4: return "CA"; case -5: return "SNR"; case 10: return "SE"; default: return "UNKNOWN(" + status + ")"; } }
    private ClassroomHomework viewable(Long homeworkId) throws Exception { ClassroomHomework row = require(homeworkId); if (!access.canManage(access.requireClassroom(row.getClassroomId()), access.uid()) && !access.isStudent(row.getClassroomId(), access.uid())) throw new StatusForbiddenException("无权访问该作业"); return row; }
    private void requireClassroomAccess(Long classroomId) throws Exception {
        String uid = access.uid();
        if (!access.canManage(access.requireClassroom(classroomId), uid)
                && !access.isStudent(classroomId, uid)) {
            throw new StatusForbiddenException("无权访问该班级");
        }
    }
    private ClassroomHomework manager(Long homeworkId) throws Exception { ClassroomHomework row = require(homeworkId); access.requireManager(row.getClassroomId()); return row; }
    private ClassroomHomework require(Long id) throws StatusNotFoundException { ClassroomHomework row = id == null ? null : homeworkMapper.selectById(id); if (row == null) throw new StatusNotFoundException("作业不存在"); return row; }
    private void replaceQuestions(Long homeworkId, List<Map<String, Object>> rows) throws StatusFailException { questionMapper.delete(new QueryWrapper<HomeworkQuestion>().eq("homework_id", homeworkId)); int order = 1; for (Map<String, Object> item : rows) { Long questionId = number(item.get("questionId")); String problemId = text(item.get("problemId")); if (questionId == null && problemId.trim().isEmpty()) throw new StatusFailException("题目必须提供题库题目或编程题编号"); String type = text(item.get("questionType")); int score = integer(item.get("score"), defaultScore(type)); questionMapper.insert(new HomeworkQuestion().setHomeworkId(homeworkId).setQuestionId(questionId).setProblemId(problemId.trim().isEmpty() ? null : problemId).setQuestionOrder(order++).setScore(score).setCreateTime(new Date())); } }
    private HomeworkSubmit ownQuestion(Long homeworkId, HomeworkQuestion q) { QueryWrapper<HomeworkSubmit> w = new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId).eq("uid", access.uid()); if (q.getQuestionId() == null) w.isNull("question_id").eq("problem_id", q.getProblemId()); else w.eq("question_id", q.getQuestionId()); return submitMapper.selectOne(w); }
    private List<HomeworkSubmit> own(Long homeworkId) { return submitMapper.selectList(new QueryWrapper<HomeworkSubmit>().eq("homework_id", homeworkId).eq("uid", access.uid()).orderByAsc("id")); }
    private Double objectiveScore(QuestionBank q, String answer, Integer fullScore) {
        if (q == null || "subjective".equals(q.getType()) || "programming".equals(q.getType())) return null;
        String expected = normalize(q.getAnswer(), q.getType()), actual = normalize(answer, q.getType());
        return expected.equals(actual) ? (fullScore == null ? 0d : fullScore.doubleValue()) : 0d;
    }
    private String normalize(String value, String type) {
        List<String> answers = HomeworkAnalysisSupport.answerList(value, type);
        if ("multiple_choice".equals(type)) {
            return answers.stream().map(this::choice).filter(item -> !item.isEmpty()).distinct().sorted()
                    .collect(Collectors.joining(","));
        }
        if ("fill_blank".equals(type)) {
            return answers.stream().map(this::plain).collect(Collectors.joining("|"));
        }
        String scalar = answers.isEmpty() ? "" : answers.get(0);
        if ("judge".equals(type)) {
            String judge = plain(scalar);
            if (Arrays.asList("true", "1", "是", "对", "正确").contains(judge)) return "true";
            if (Arrays.asList("false", "0", "否", "错", "错误").contains(judge)) return "false";
        }
        return plain(scalar);
    }
    private String choice(String value) { return text(value).replaceAll("^[\\[\\(\\{\\\"']+|[\\]\\)\\}\\\"']+$", "").trim().toUpperCase(); }
    private String plain(String value) { return text(value).replace("\r", "").replace("\n", " ").trim().toLowerCase(); }
    private int status(Date start, Date end) { Date now = new Date(); return now.before(start) ? 1 : now.after(end) ? 3 : 2; }
    private int defaultScore(String type) { if ("multiple_choice".equals(type)) return 5; if ("judge".equals(type) || "single_choice".equals(type)) return 1; if ("programming".equals(type)) return 20; return 2; }
    private Date date(Object value) { if (value instanceof Date) return (Date) value; try { return Date.from(OffsetDateTime.parse(text(value), DateTimeFormatter.ISO_DATE_TIME).toInstant()); } catch (Exception e) { try { return Date.from(java.time.Instant.parse(text(value))); } catch (Exception ignored) { return null; } } }
    @SuppressWarnings("unchecked") private List<Map<String, Object>> list(Object value) { return value instanceof List ? (List<Map<String, Object>>) value : Collections.emptyList(); }
    private Long number(Object value) { if (value instanceof Number) return ((Number) value).longValue(); try { return value == null ? null : Long.valueOf(String.valueOf(value)); } catch (Exception e) { return null; } }
    private int integer(Object value, int fallback) { if (value instanceof Number) return ((Number) value).intValue(); try { return value == null ? fallback : Integer.parseInt(String.valueOf(value)); } catch (Exception e) { return fallback; } }
    private double decimal(Object value) { if (value instanceof Number) return ((Number) value).doubleValue(); try { return Double.parseDouble(String.valueOf(value)); } catch (Exception e) { return 0; } }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private String first(String... values) { for (String value : values) if (value != null && !value.trim().isEmpty()) return value; return ""; }
}
