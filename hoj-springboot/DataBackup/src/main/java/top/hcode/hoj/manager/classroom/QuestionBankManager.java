package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.QuestionBankMapper;
import top.hcode.hoj.pojo.entity.classroom.QuestionBank;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;
import top.hcode.hoj.pojo.vo.classroom.QuestionBankVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class QuestionBankManager {
    @Resource private QuestionBankMapper questionMapper;
    @Resource private UserInfoEntityService userInfoService;
    @Resource private ClassroomAccessManager accessManager;
    @Resource private ClassroomViewBuilder viewBuilder;

    public QuestionBankVO create(Map<String, Object> request) throws Exception {
        String type = text(request.get("type"));
        String[] normalized = QuestionAnswerNormalizer.normalize(type, text(request.get("answer")), text(request.get("options")));
        int score = integer(request.get("score"), "programming".equals(type) ? 20 : 2);
        QuestionBank row = new QuestionBank().setTitle(text(request.get("title"))).setType(type).setContent(text(request.get("content")))
                .setOptions(normalized[1]).setAnswer(normalized[0]).setAnalysis(text(request.get("analysis"))).setTags(text(request.get("tags")))
                .setCourse(text(request.get("course"))).setDifficulty(integer(request.get("difficulty"), 1)).setScore(score)
                .setCreatorId(accessManager.uid()).setIsShared(integer(request.get("isShared"), 0)).setProblemId("programming".equals(type) ? text(request.get("problemId")) : null)
                .setStatus(1).setCreateTime(new Date()).setUpdateTime(new Date());
        questionMapper.insert(row); return view(row);
    }

    public Map<String, Object> list(Map<String, String> params, boolean admin) {
        int page = positive(params.get("page"), 1), limit = Math.min(100, positive(params.get("limit"), 20));
        QueryWrapper<QuestionBank> query = new QueryWrapper<QuestionBank>().eq("status", 1);
        if (!admin) query.and(q -> q.eq("creator_id", accessManager.uid()).or().eq("is_shared", 1));
        filter(query, params);
        IPage<QuestionBank> result = questionMapper.selectPage(new Page<>(page, limit), query.orderByDesc("create_time"));
        Map<String, Object> data = new LinkedHashMap<>(); data.put("total", result.getTotal()); data.put("page", page); data.put("limit", limit);
        data.put("questions", result.getRecords().stream().map(this::view).collect(Collectors.toList())); return data;
    }

    public QuestionBankVO detail(Long id) throws StatusNotFoundException {
        return view(require(id));
    }

    public void update(Long id, Map<String, Object> request, boolean admin) throws Exception {
        QuestionBank row = require(id); checkOwner(row, admin);
        String type = request.containsKey("type") ? text(request.get("type")) : row.getType();
        String answer = request.containsKey("answer") ? text(request.get("answer")) : row.getAnswer();
        String options = request.containsKey("options") ? text(request.get("options")) : row.getOptions();
        String[] normalized = QuestionAnswerNormalizer.normalize(type, answer, options);
        UpdateWrapper<QuestionBank> update = new UpdateWrapper<QuestionBank>().eq("id", id).set("answer", normalized[0]).set("options", normalized[1]).set("update_time", new Date());
        set(update, request, "title", "title"); set(update, request, "content", "content"); set(update, request, "analysis", "analysis");
        set(update, request, "tags", "tags"); set(update, request, "course", "course"); set(update, request, "difficulty", "difficulty");
        set(update, request, "score", "score"); set(update, request, "isShared", "is_shared"); if (request.containsKey("type")) update.set("type", type);
        questionMapper.update(null, update);
    }

    @Transactional(rollbackFor = Exception.class)
    public void delete(Long id, boolean admin) throws Exception {
        QuestionBank row = require(id); checkOwner(row, admin);
        questionMapper.update(null, new UpdateWrapper<QuestionBank>().eq("id", id).set("status", 0).set("update_time", new Date()));
    }

    private QuestionBank require(Long id) throws StatusNotFoundException {
        QuestionBank row = questionMapper.selectOne(new QueryWrapper<QuestionBank>().eq("id", id).eq("status", 1));
        if (row == null) throw new StatusNotFoundException("题目不存在"); return row;
    }
    private void checkOwner(QuestionBank row, boolean admin) throws StatusForbiddenException { if (!admin && !accessManager.uid().equals(row.getCreatorId())) throw new StatusForbiddenException("无权操作该题目"); }
    private void filter(QueryWrapper<QuestionBank> q, Map<String, String> p) {
        if (notBlank(p.get("type"))) q.eq("type", p.get("type")); if (notBlank(p.get("isShared"))) q.eq("is_shared", p.get("isShared"));
        if (notBlank(p.get("course"))) q.eq("course", p.get("course")); if (notBlank(p.get("difficulty"))) q.eq("difficulty", p.get("difficulty"));
        if (notBlank(p.get("keyword"))) q.like("title", p.get("keyword")); if (notBlank(p.get("questionId"))) q.eq("id", p.get("questionId"));
        if (notBlank(p.get("tag"))) q.apply("JSON_CONTAINS(tags, {0})", "\"" + p.get("tag") + "\"");
    }
    private QuestionBankVO view(QuestionBank row) {
        UserInfo user = userInfoService.getById(row.getCreatorId()); ClassroomUserVO creator = viewBuilder.user(user);
        return new QuestionBankVO().setId(row.getId()).setTitle(row.getTitle()).setType(row.getType()).setContent(row.getContent()).setOptions(row.getOptions()).setAnswer(row.getAnswer()).setAnalysis(row.getAnalysis()).setTags(row.getTags()).setCourse(row.getCourse()).setDifficulty(row.getDifficulty()).setScore(row.getScore()).setCreatorId(row.getCreatorId()).setIsShared(row.getIsShared()).setProblemId(row.getProblemId()).setStatus(row.getStatus()).setCreatedAt(row.getCreateTime()).setUpdatedAt(row.getUpdateTime()).setCreator(creator);
    }
    private void set(UpdateWrapper<QuestionBank> update, Map<String, Object> request, String key, String column) { if (request.containsKey(key)) update.set(column, request.get(key)); }
    private int integer(Object value, int fallback) { return value instanceof Number && ((Number) value).intValue() != 0 ? ((Number) value).intValue() : fallback; }
    private int positive(String value, int fallback) { try { return Integer.parseInt(value) < 1 ? fallback : Integer.parseInt(value); } catch (Exception e) { return fallback; } }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private boolean notBlank(String value) { return value != null && !value.trim().isEmpty(); }
}
