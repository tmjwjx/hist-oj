package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomHomeworkMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.HomeworkQuestionMapper;
import top.hcode.hoj.mapper.classroom.HomeworkSubmitMapper;
import top.hcode.hoj.pojo.entity.classroom.ClassroomHomework;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.classroom.HomeworkQuestion;
import top.hcode.hoj.pojo.entity.classroom.HomeworkSubmit;
import top.hcode.hoj.pojo.entity.user.UserInfo;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class HomeworkRankingManager {
    @Resource private ClassroomHomeworkMapper homeworkMapper;
    @Resource private HomeworkQuestionMapper questionMapper;
    @Resource private HomeworkSubmitMapper submitMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private UserInfoEntityService userService;
    @Resource private ClassroomAccessManager access;

    public Map<String, Object> ranking(Long homeworkId) throws Exception {
        ClassroomHomework homework = require(homeworkId);
        checkAccess(homework);
        List<HomeworkQuestion> questions = questionMapper.selectList(new QueryWrapper<HomeworkQuestion>()
                .eq("homework_id", homeworkId).orderByAsc("question_order"));
        Map<String, String> columnByQuestion = new HashMap<>();
        List<Map<String, Object>> columns = columns(questions, columnByQuestion);

        List<ClassroomStudent> students = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", homework.getClassroomId()).eq("status", 1).orderByAsc("id"));
        Map<String, UserInfo> users = users(students);
        Set<String> activeUids = students.stream().map(ClassroomStudent::getUid).collect(Collectors.toSet());
        List<HomeworkSubmit> submissions = submitMapper.selectList(new QueryWrapper<HomeworkSubmit>()
                .eq("homework_id", homeworkId).eq("is_officially_submitted", 1)
                .orderByDesc("update_time").orderByDesc("id"));

        Map<String, HomeworkSubmit> latest = new LinkedHashMap<>();
        for (HomeworkSubmit row : submissions) {
            String identity = identity(row);
            if (identity == null || !activeUids.contains(row.getUid()) || !columnByQuestion.containsKey(identity)) continue;
            latest.putIfAbsent(row.getUid() + "|" + identity, row);
        }

        Map<String, Map<String, Double>> questionScores = new HashMap<>();
        Map<String, Double> totalScores = new HashMap<>();
        for (HomeworkSubmit row : latest.values()) {
            String column = columnByQuestion.get(identity(row));
            double score = row.getScore() == null ? 0d : row.getScore();
            questionScores.computeIfAbsent(row.getUid(), key -> new LinkedHashMap<>()).put(column, score);
            totalScores.merge(row.getUid(), score, Double::sum);
        }

        List<Map<String, Object>> rows = new ArrayList<>();
        for (ClassroomStudent student : students) {
            UserInfo user = users.get(student.getUid());
            Map<String, Object> row = new LinkedHashMap<>();
            row.put("uid", student.getUid());
            row.put("realName", first(student.getRealName(), user == null ? null : user.getRealname(), student.getUid()));
            row.put("username", first(user == null ? null : user.getUsername(), student.getUid()));
            row.put("totalScore", totalScores.getOrDefault(student.getUid(), 0d));
            row.put("score", totalScores.getOrDefault(student.getUid(), 0d));
            row.put("hasScore", questionScores.containsKey(student.getUid()));
            row.put("questionScores", questionScores.getOrDefault(student.getUid(), Collections.emptyMap()));
            rows.add(row);
        }
        sortAndRank(rows);

        Map<String, Object> result = new LinkedHashMap<>();
        result.put("questionColumns", columns);
        result.put("rankings", rows);
        result.put("totalStudentCount", rows.size());
        return result;
    }

    private List<Map<String, Object>> columns(List<HomeworkQuestion> questions, Map<String, String> index) {
        List<Map<String, Object>> result = new ArrayList<>();
        for (int i = 0; i < questions.size(); i++) {
            HomeworkQuestion question = questions.get(i);
            String key = String.valueOf(i + 1);
            String identity = identity(question);
            if (identity != null) index.put(identity, key);
            Map<String, Object> column = new LinkedHashMap<>();
            column.put("key", key);
            column.put("label", key);
            column.put("questionOrder", question.getQuestionOrder());
            result.add(column);
        }
        return result;
    }

    private void sortAndRank(List<Map<String, Object>> rows) {
        rows.sort((left, right) -> {
            boolean leftHas = Boolean.TRUE.equals(left.get("hasScore"));
            boolean rightHas = Boolean.TRUE.equals(right.get("hasScore"));
            if (leftHas != rightHas) return leftHas ? -1 : 1;
            int score = Double.compare(number(right.get("totalScore")), number(left.get("totalScore")));
            if (leftHas && score != 0) return score;
            int username = text(left.get("username")).compareTo(text(right.get("username")));
            return username != 0 ? username : text(left.get("uid")).compareTo(text(right.get("uid")));
        });
        int rank = 0;
        double previous = Double.NaN;
        boolean previousHasScore = false;
        for (int i = 0; i < rows.size(); i++) {
            boolean hasScore = Boolean.TRUE.equals(rows.get(i).get("hasScore"));
            double score = number(rows.get(i).get("totalScore"));
            if (i == 0 || hasScore != previousHasScore || !hasScore || Math.abs(score - previous) > 1e-9) rank = i + 1;
            rows.get(i).put("rank", rank);
            previous = score;
            previousHasScore = hasScore;
        }
    }

    private void checkAccess(ClassroomHomework homework) throws Exception {
        String uid = access.uid();
        boolean manager = access.canManage(access.requireClassroom(homework.getClassroomId()), uid);
        if (manager) return;
        if (!access.isStudent(homework.getClassroomId(), uid)) throw new StatusForbiddenException("无权查看该作业排行榜");
        if (!Objects.equals(homework.getShowRank(), 1)) throw new StatusForbiddenException("教师未开放排行榜");
        if (new Date().before(homework.getEndTime())) throw new StatusForbiddenException("作业未结束，排行榜暂未开放");
    }

    private Map<String, UserInfo> users(List<ClassroomStudent> students) {
        Set<String> ids = students.stream().map(ClassroomStudent::getUid).collect(Collectors.toSet());
        if (ids.isEmpty()) return Collections.emptyMap();
        return userService.listByIds(ids).stream().collect(Collectors.toMap(
                UserInfo::getUuid, row -> row, (left, right) -> left, LinkedHashMap::new));
    }

    private String identity(HomeworkQuestion row) {
        if (row.getProblemId() != null && !row.getProblemId().trim().isEmpty()) return "p:" + row.getProblemId().trim();
        return row.getQuestionId() == null ? null : "q:" + row.getQuestionId();
    }

    private String identity(HomeworkSubmit row) {
        if (row.getProblemId() != null && !row.getProblemId().trim().isEmpty()) return "p:" + row.getProblemId().trim();
        return row.getQuestionId() == null ? null : "q:" + row.getQuestionId();
    }

    private ClassroomHomework require(Long id) throws StatusNotFoundException {
        ClassroomHomework row = id == null ? null : homeworkMapper.selectById(id);
        if (row == null) throw new StatusNotFoundException("作业不存在");
        return row;
    }

    private double number(Object value) { return value instanceof Number ? ((Number) value).doubleValue() : 0d; }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private String first(String... values) { for (String value : values) if (value != null && !value.trim().isEmpty()) return value; return ""; }
}
