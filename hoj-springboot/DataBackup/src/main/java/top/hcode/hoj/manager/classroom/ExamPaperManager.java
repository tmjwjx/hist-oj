package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.mapper.classroom.ExamPaperMapper;
import top.hcode.hoj.mapper.classroom.ExamPaperQuestionMapper;
import top.hcode.hoj.mapper.classroom.QuestionBankMapper;
import top.hcode.hoj.pojo.entity.classroom.ExamPaper;
import top.hcode.hoj.pojo.entity.classroom.ExamPaperQuestion;
import top.hcode.hoj.pojo.entity.classroom.QuestionBank;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ExamPaperManager {
    @Resource private ExamPaperMapper paperMapper;
    @Resource private ExamPaperQuestionMapper paperQuestionMapper;
    @Resource private QuestionBankMapper questionMapper;
    @Resource private ClassroomAccessManager accessManager;

    @Transactional(rollbackFor = Exception.class)
    public ExamPaper create(Map<String, Object> request) throws Exception {
        List<Map<String, Object>> questions = questionList(request.get("questions"));
        ExamPaper paper = new ExamPaper().setTitle(text(request.get("title"))).setDescription(text(request.get("description")))
                .setCreatorId(accessManager.uid()).setIsShared(integer(request.get("isShared"))).setIsPublic(0)
                .setTotalScore(total(questions)).setQuestionCount(questions.size()).setStatus(1).setCreateTime(new Date()).setUpdateTime(new Date());
        paperMapper.insert(paper); replaceQuestions(paper.getId(), questions); return paper;
    }

    public Map<String, Object> list(Map<String, String> params, boolean admin) {
        int page = positive(params.get("page"), 1), limit = Math.min(100, positive(params.get("limit"), 20));
        QueryWrapper<ExamPaper> query = new QueryWrapper<ExamPaper>().eq("status", 1);
        if (!admin) query.and(q -> q.eq("creator_id", accessManager.uid()).or().eq("is_shared", 1));
        filter(query, params);
        IPage<ExamPaper> result = paperMapper.selectPage(new Page<>(page, limit), query.orderByDesc("create_time"));
        Map<String, Object> data = new LinkedHashMap<>(); data.put("total", result.getTotal()); data.put("page", page); data.put("limit", limit);
        data.put("papers", result.getRecords().stream().map(row -> paperView(row, true)).collect(Collectors.toList())); return data;
    }

    public Map<String, Object> detail(Long id, boolean publicOnly) throws StatusNotFoundException {
        ExamPaper paper = require(id, publicOnly);
        if (!publicOnly && !accessManager.isAdmin() && !accessManager.uid().equals(paper.getCreatorId())
                && !Objects.equals(paper.getIsShared(), 1)) {
            throw new StatusNotFoundException("试卷不存在");
        }
        return paperView(paper, !publicOnly);
    }

    public Map<String, Object> adminDetail(Long id) throws StatusNotFoundException {
        return paperView(require(id, false), true);
    }

    public Map<String, Object> publicList(Map<String, String> params) {
        int page = positive(params.get("page"), 1), limit = Math.min(100, positive(params.get("limit"), 20));
        QueryWrapper<ExamPaper> query = new QueryWrapper<ExamPaper>().eq("status", 1).eq("is_public", 1);
        filter(query, params);
        IPage<ExamPaper> result = paperMapper.selectPage(new Page<>(page, limit), query.orderByDesc("update_time"));
        Map<String, Object> data = new LinkedHashMap<>();
        data.put("total", result.getTotal()); data.put("page", page); data.put("limit", limit);
        data.put("papers", result.getRecords().stream().map(row -> paperView(row, false)).collect(Collectors.toList()));
        return data;
    }

    public Map<String, Object> questionAnswer(Long paperId, Long questionId) throws StatusNotFoundException {
        require(paperId, true);
        if (paperQuestionMapper.selectCount(new QueryWrapper<ExamPaperQuestion>().eq("exam_paper_id", paperId).eq("question_id", questionId)) == 0)
            throw new StatusNotFoundException("题目不在当前试卷中");
        QuestionBank question = questionMapper.selectOne(new QueryWrapper<QuestionBank>().eq("id", questionId).eq("status", 1));
        if (question == null) throw new StatusNotFoundException("题目不存在");
        Map<String, Object> result = new LinkedHashMap<>(); result.put("questionId", question.getId()); result.put("answer", question.getAnswer()); result.put("analysis", question.getAnalysis()); return result;
    }

    @Transactional(rollbackFor = Exception.class)
    public void update(Long id, Map<String, Object> request, boolean admin) throws Exception {
        ExamPaper paper = require(id, false); if (!admin && !accessManager.uid().equals(paper.getCreatorId())) throw new StatusForbiddenException("无权修改此试卷");
        UpdateWrapper<ExamPaper> update = new UpdateWrapper<ExamPaper>().eq("id", id).set("update_time", new Date());
        set(update, request, "title", "title"); set(update, request, "description", "description"); set(update, request, "isShared", "is_shared");
        if (admin) set(update, request, "isPublic", "is_public");
        if (request.containsKey("questions")) {
            List<Map<String, Object>> questions = questionList(request.get("questions"));
            update.set("total_score", total(questions)).set("question_count", questions.size());
            paperMapper.update(null, update); replaceQuestions(id, questions);
        } else paperMapper.update(null, update);
    }

    public void delete(Long id, boolean admin) throws Exception {
        ExamPaper paper = require(id, false); if (!admin && !accessManager.uid().equals(paper.getCreatorId())) throw new StatusForbiddenException("无权删除此试卷");
        paperMapper.update(null, new UpdateWrapper<ExamPaper>().eq("id", id).set("status", 0).set("update_time", new Date()));
    }

    public Map<String, Object> importToHomework(Long paperId, Long classroomId) throws Exception {
        ExamPaper paper = require(paperId, false);
        if (!accessManager.isAdmin() && !accessManager.uid().equals(paper.getCreatorId()) && !Objects.equals(paper.getIsShared(), 1))
            throw new StatusForbiddenException("无权导入此试卷");
        List<Map<String, Object>> questions = new ArrayList<>();
        for (ExamPaperQuestion row : paperQuestions(paperId)) {
            Map<String, Object> item = new LinkedHashMap<>(); item.put("questionOrder", row.getQuestionOrder()); item.put("questionType", row.getQuestionType()); item.put("score", row.getScore());
            if ("programming".equals(row.getQuestionType())) { item.put("problemId", row.getProblemId()); item.put("title", "BingOJ 编程题 - " + row.getProblemId()); item.put("type", "programming"); item.put("difficulty", 5); }
            else { QuestionBank question = row.getQuestionId() == null ? null : questionMapper.selectById(row.getQuestionId()); if (question != null) { item.put("questionId", question.getId()); item.put("id", question.getId()); item.put("type", question.getType()); item.put("title", question.getTitle()); item.put("difficulty", question.getDifficulty()); } }
            questions.add(item);
        }
        Map<String, Object> result = new LinkedHashMap<>(); result.put("questions", questions); result.put("totalScore", paper.getTotalScore()); result.put("questionCount", paper.getQuestionCount()); return result;
    }

    private ExamPaper require(Long id, boolean publicOnly) throws StatusNotFoundException {
        QueryWrapper<ExamPaper> query = new QueryWrapper<ExamPaper>().eq("id", id).eq("status", 1); if (publicOnly) query.eq("is_public", 1);
        ExamPaper row = paperMapper.selectOne(query); if (row == null) throw new StatusNotFoundException(publicOnly ? "试卷不存在或未公开" : "试卷不存在"); return row;
    }
    private Map<String, Object> paperView(ExamPaper paper, boolean includeAnswers) {
        Map<String, Object> result = new LinkedHashMap<>(); result.put("id", paper.getId()); result.put("title", paper.getTitle()); result.put("creatorId", paper.getCreatorId()); result.put("isShared", paper.getIsShared()); result.put("isPublic", paper.getIsPublic()); result.put("totalScore", paper.getTotalScore()); result.put("questionCount", paper.getQuestionCount()); result.put("description", paper.getDescription()); result.put("status", paper.getStatus()); result.put("createdAt", paper.getCreateTime()); result.put("updatedAt", paper.getUpdateTime());
        result.put("questions", paperQuestions(paper.getId()).stream().map(row -> questionView(row, includeAnswers)).collect(Collectors.toList())); return result;
    }
    private Map<String, Object> questionView(ExamPaperQuestion row, boolean includeAnswers) {
        Map<String, Object> result = new LinkedHashMap<>(); result.put("id", row.getId()); result.put("examPaperId", row.getExamPaperId()); result.put("questionId", row.getQuestionId()); result.put("problemId", row.getProblemId()); result.put("questionOrder", row.getQuestionOrder()); result.put("questionType", row.getQuestionType()); result.put("score", row.getScore());
        if (row.getQuestionId() != null) {
            QuestionBank question = questionMapper.selectById(row.getQuestionId());
            if (question != null) {
                Map<String, Object> questionData = new LinkedHashMap<>();
                questionData.put("id", question.getId()); questionData.put("title", question.getTitle());
                questionData.put("type", question.getType()); questionData.put("content", question.getContent());
                questionData.put("options", question.getOptions()); questionData.put("difficulty", question.getDifficulty());
                questionData.put("score", question.getScore()); questionData.put("problemId", question.getProblemId());
                if (includeAnswers) { questionData.put("answer", question.getAnswer()); questionData.put("analysis", question.getAnalysis()); }
                result.put("question", questionData);
            }
        }
        return result;
    }
    private List<ExamPaperQuestion> paperQuestions(Long id) { return paperQuestionMapper.selectList(new QueryWrapper<ExamPaperQuestion>().eq("exam_paper_id", id).orderByAsc("question_order")); }
    private void replaceQuestions(Long paperId, List<Map<String, Object>> questions) { paperQuestionMapper.delete(new QueryWrapper<ExamPaperQuestion>().eq("exam_paper_id", paperId)); int order = 1; for (Map<String, Object> item : questions) { paperQuestionMapper.insert(new ExamPaperQuestion().setExamPaperId(paperId).setQuestionId(longValue(item.get("questionId"))).setProblemId(text(item.get("problemId"))).setQuestionOrder(order++).setQuestionType(text(item.get("questionType"))).setScore(integer(item.get("score")))); } }
    private void filter(QueryWrapper<ExamPaper> q, Map<String, String> p) { if (notBlank(p.get("paperId"))) q.eq("id", p.get("paperId")); if (notBlank(p.get("keyword"))) q.and(x -> x.like("title", p.get("keyword"))); if (notBlank(p.get("isShared"))) q.eq("is_shared", p.get("isShared")); if (notBlank(p.get("isPublic"))) q.eq("is_public", p.get("isPublic")); }
    @SuppressWarnings("unchecked") private List<Map<String, Object>> questionList(Object value) { return value instanceof List ? (List<Map<String, Object>>) value : Collections.emptyList(); }
    private int total(List<Map<String, Object>> rows) { return rows.stream().mapToInt(row -> integer(row.get("score"))).sum(); }
    private void set(UpdateWrapper<ExamPaper> update, Map<String, Object> request, String key, String column) { if (request.containsKey(key)) update.set(column, request.get(key)); }
    private int integer(Object value) { return value instanceof Number ? ((Number) value).intValue() : 0; }
    private Long longValue(Object value) { if (value instanceof Number) return ((Number) value).longValue(); try { return value == null ? null : Long.valueOf(String.valueOf(value)); } catch (Exception e) { return null; } }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private int integer(Object value, int fallback) { int result = integer(value); return result == 0 ? fallback : result; }
    private int positive(String value, int fallback) { try { int result = Integer.parseInt(value); return result < 1 ? fallback : result; } catch (Exception e) { return fallback; } }
    private boolean notBlank(String value) { return value != null && !value.trim().isEmpty(); }
}
