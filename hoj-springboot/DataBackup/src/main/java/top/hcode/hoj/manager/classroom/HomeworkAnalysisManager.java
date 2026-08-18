package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomHomeworkMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.HomeworkQuestionMapper;
import top.hcode.hoj.mapper.classroom.HomeworkSubmitMapper;
import top.hcode.hoj.mapper.classroom.QuestionBankMapper;
import top.hcode.hoj.pojo.entity.classroom.ClassroomHomework;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.classroom.HomeworkQuestion;
import top.hcode.hoj.pojo.entity.classroom.HomeworkSubmit;
import top.hcode.hoj.pojo.entity.classroom.QuestionBank;
import top.hcode.hoj.pojo.entity.user.UserInfo;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class HomeworkAnalysisManager {
    @Resource private ClassroomHomeworkMapper homeworkMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private HomeworkQuestionMapper questionMapper;
    @Resource private HomeworkSubmitMapper submitMapper;
    @Resource private QuestionBankMapper bankMapper;
    @Resource private UserInfoEntityService userService;
    @Resource private ClassroomAccessManager access;

    public Map<String, Object> analysis(Long homeworkId) throws Exception {
        ClassroomHomework homework = require(homeworkId);
        access.requireManager(homework.getClassroomId());

        List<ClassroomStudent> students = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", homework.getClassroomId()).eq("status", 1).orderByAsc("id"));
        Map<String, UserInfo> users = users(students);
        Map<String, ClassroomStudent> studentRows = students.stream()
                .collect(Collectors.toMap(ClassroomStudent::getUid, row -> row, (left, right) -> left, LinkedHashMap::new));

        List<HomeworkSubmit> officialRows = submitMapper.selectList(new QueryWrapper<HomeworkSubmit>()
                .eq("homework_id", homeworkId).eq("is_officially_submitted", 1).orderByAsc("id"));
        officialRows = officialRows.stream().filter(row -> studentRows.containsKey(row.getUid())).collect(Collectors.toList());
        Set<String> submittedUids = officialRows.stream().map(HomeworkSubmit::getUid).collect(Collectors.toSet());
        List<HomeworkSubmit> finalRows = finalQuestionRows(officialRows);

        Map<String, Double> totals = new HashMap<>();
        for (HomeworkSubmit row : finalRows) totals.merge(row.getUid(), score(row), Double::sum);
        List<Map<String, Object>> rankings = rankings(students, users, totals, submittedUids);
        Map<String, Map<String, Object>> profiles = rankings.stream().collect(Collectors.toMap(
                row -> String.valueOf(row.get("uid")), row -> row, (left, right) -> left, LinkedHashMap::new));

        List<Map<String, Object>> submittedStudents = filter(rankings, submittedUids, true);
        List<Map<String, Object>> unsubmittedStudents = filter(rankings, submittedUids, false);
        List<HomeworkQuestion> relations = questionMapper.selectList(new QueryWrapper<HomeworkQuestion>()
                .eq("homework_id", homeworkId).orderByAsc("question_order"));
        Map<Long, QuestionBank> banks = banks(relations);

        List<Map<String, Object>> questionAnalysis = new ArrayList<>();
        for (HomeworkQuestion relation : relations) {
            QuestionBank bank = relation.getQuestionId() == null ? null : banks.get(relation.getQuestionId());
            List<HomeworkSubmit> rows = finalRows.stream().filter(row -> matches(relation, row)).collect(Collectors.toList());
            questionAnalysis.add(question(relation, bank, rows, students, profiles));
        }

        double totalScore = submittedStudents.stream().mapToDouble(row -> number(row.get("score"))).sum();
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("totalStudentCount", students.size());
        result.put("submittedCount", submittedStudents.size());
        result.put("unsubmittedCount", unsubmittedStudents.size());
        result.put("submittedStudents", submittedStudents);
        result.put("unsubmittedStudents", unsubmittedStudents);
        result.put("studentRankings", rankings);
        result.put("questionAnalysis", questionAnalysis);
        result.put("averageScore", submittedStudents.isEmpty() ? 0d : totalScore / submittedStudents.size());
        result.put("submissionCount", finalRows.size());
        result.put("studentCount", submittedStudents.size());
        return result;
    }

    private Map<String, Object> question(HomeworkQuestion relation, QuestionBank bank,
                                         List<HomeworkSubmit> rows, List<ClassroomStudent> allStudents,
                                         Map<String, Map<String, Object>> profiles) {
        String type = bank == null ? "programming" : text(bank.getType());
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("homeworkQuestionId", relation.getId());
        result.put("questionId", relation.getQuestionId());
        result.put("problemId", relation.getProblemId());
        result.put("questionOrder", relation.getQuestionOrder());
        result.put("score", relation.getScore());
        result.put("type", type);
        result.put("title", bank == null ? "编程题 - " + text(relation.getProblemId()) : bank.getTitle());
        result.put("answer", bank == null ? "" : HomeworkAnalysisSupport.displayAnswer(bank.getAnswer(), type));
        result.put("submittedCount", rows.size());
        result.put("avgScore", rows.isEmpty() ? 0d : rows.stream().mapToDouble(this::score).average().orElse(0d));

        Set<String> submitted = rows.stream().map(HomeworkSubmit::getUid).collect(Collectors.toSet());
        List<Map<String, Object>> submittedBy = new ArrayList<>();
        for (HomeworkSubmit row : rows) {
            Map<String, Object> base = profiles.get(row.getUid());
            if (base == null) continue;
            Map<String, Object> student = new LinkedHashMap<>(base);
            student.put("score", score(row));
            student.put("answer", HomeworkAnalysisSupport.answerList(row.getAnswer(), type));
            submittedBy.add(student);
        }
        List<Map<String, Object>> unsubmittedBy = new ArrayList<>();
        for (ClassroomStudent student : allStudents) {
            if (!submitted.contains(student.getUid()) && profiles.containsKey(student.getUid())) {
                unsubmittedBy.add(new LinkedHashMap<>(profiles.get(student.getUid())));
            }
        }
        result.put("submittedBy", submittedBy);
        result.put("unsubmittedBy", unsubmittedBy);
        result.put("unsubmittedCount", unsubmittedBy.size());
        result.put("options", HomeworkAnalysisSupport.distributions(bank, relation, rows, profiles));
        result.put("scoreSegments", !Arrays.asList("single_choice", "multiple_choice", "judge", "fill_blank").contains(type));
        return result;
    }

    private List<Map<String, Object>> rankings(List<ClassroomStudent> students, Map<String, UserInfo> users,
                                                Map<String, Double> totals, Set<String> submitted) {
        List<Map<String, Object>> result = new ArrayList<>();
        for (ClassroomStudent row : students) {
            UserInfo user = users.get(row.getUid());
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("uid", row.getUid());
            item.put("realName", first(row.getRealName(), user == null ? null : user.getRealname(), row.getUid()));
            item.put("username", first(user == null ? null : user.getUsername(), row.getUid()));
            item.put("nickname", first(user == null ? null : user.getNickname(), ""));
            item.put("score", totals.getOrDefault(row.getUid(), 0d));
            item.put("submitted", submitted.contains(row.getUid()));
            result.add(item);
        }
        result.sort((left, right) -> {
            int score = Double.compare(number(right.get("score")), number(left.get("score")));
            if (score != 0) return score;
            int username = text(left.get("username")).compareTo(text(right.get("username")));
            if (username != 0) return username;
            return text(left.get("uid")).compareTo(text(right.get("uid")));
        });
        double previous = Double.NaN;
        int rank = 0;
        for (int index = 0; index < result.size(); index++) {
            double current = number(result.get(index).get("score"));
            if (index == 0 || Math.abs(current - previous) > 1e-9) rank = index + 1;
            result.get(index).put("rank", rank);
            previous = current;
        }
        return result;
    }

    private List<Map<String, Object>> filter(List<Map<String, Object>> rankings, Set<String> submitted, boolean expected) {
        return rankings.stream().filter(row -> submitted.contains(String.valueOf(row.get("uid"))) == expected)
                .map(LinkedHashMap::new).collect(Collectors.toList());
    }

    private List<HomeworkSubmit> finalQuestionRows(List<HomeworkSubmit> rows) {
        Map<String, HomeworkSubmit> latest = new LinkedHashMap<>();
        for (HomeworkSubmit row : rows) {
            String key = questionKey(row);
            if (key == null) continue;
            HomeworkSubmit current = latest.get(key);
            if (current == null || prefer(row, current)) latest.put(key, row);
        }
        return new ArrayList<>(latest.values());
    }

    private boolean prefer(HomeworkSubmit candidate, HomeworkSubmit current) {
        if (candidate.getProblemId() != null && !candidate.getProblemId().trim().isEmpty()) {
            int score = Double.compare(score(candidate), score(current));
            if (score != 0) return score > 0;
        }
        return candidate.getId() != null && (current.getId() == null || candidate.getId() > current.getId());
    }

    private String questionKey(HomeworkSubmit row) {
        if (row.getQuestionId() != null) return row.getUid() + "|q:" + row.getQuestionId();
        if (row.getProblemId() != null && !row.getProblemId().trim().isEmpty()) return row.getUid() + "|p:" + row.getProblemId();
        return null;
    }

    private boolean matches(HomeworkQuestion relation, HomeworkSubmit row) {
        if (relation.getQuestionId() != null) return Objects.equals(relation.getQuestionId(), row.getQuestionId());
        return relation.getProblemId() != null && Objects.equals(relation.getProblemId(), row.getProblemId());
    }

    private Map<String, UserInfo> users(List<ClassroomStudent> students) {
        Set<String> ids = students.stream().map(ClassroomStudent::getUid).collect(Collectors.toSet());
        if (ids.isEmpty()) return Collections.emptyMap();
        return userService.listByIds(ids).stream().collect(Collectors.toMap(
                UserInfo::getUuid, row -> row, (left, right) -> left, LinkedHashMap::new));
    }

    private Map<Long, QuestionBank> banks(List<HomeworkQuestion> relations) {
        Set<Long> ids = relations.stream().map(HomeworkQuestion::getQuestionId).filter(Objects::nonNull).collect(Collectors.toSet());
        if (ids.isEmpty()) return Collections.emptyMap();
        return bankMapper.selectBatchIds(ids).stream().collect(Collectors.toMap(
                QuestionBank::getId, row -> row, (left, right) -> left, LinkedHashMap::new));
    }

    private ClassroomHomework require(Long id) throws StatusNotFoundException {
        ClassroomHomework row = id == null ? null : homeworkMapper.selectById(id);
        if (row == null) throw new StatusNotFoundException("作业不存在");
        return row;
    }

    private double score(HomeworkSubmit row) { return row.getScore() == null ? 0d : row.getScore(); }
    private double number(Object value) { return value instanceof Number ? ((Number) value).doubleValue() : 0d; }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private String first(String... values) { for (String value : values) if (value != null && !value.trim().isEmpty()) return value; return ""; }
}
