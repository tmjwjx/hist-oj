package top.hcode.hoj.manager.learning;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.LearningMapException;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.problem.ProblemTagEntityService;
import top.hcode.hoj.dao.problem.TagEntityService;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemTag;
import top.hcode.hoj.pojo.entity.problem.Tag;
import top.hcode.hoj.pojo.vo.LearningProblemVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Component
public class LearningMapProblemManager {

    @Resource
    private ProblemEntityService problemService;

    @Resource
    private ProblemTagEntityService problemTagService;

    @Resource
    private TagEntityService tagService;

    public List<LearningProblemVO> search(String keyword) {
        QueryWrapper<Problem> query = new QueryWrapper<Problem>().eq("auth", 1);
        String value = LearningMapRules.text(keyword);
        if (!value.isEmpty()) {
            query.and(wrapper -> wrapper.like("problem_id", value).or().like("title", value));
        }
        List<Problem> problems = problemService.list(query.orderByDesc("id").last("LIMIT 20"));
        return toViews(problems);
    }

    public LearningProblemVO get(String identifier) {
        String value = LearningMapRules.text(identifier);
        if (value.isEmpty()) {
            throw new LearningMapException("题目标识不能为空");
        }
        Problem problem = null;
        if (value.matches("\\d+")) {
            problem = problemService.getById(Long.valueOf(value));
        }
        if (problem == null) {
            problem = problemService.getOne(new QueryWrapper<Problem>().eq("problem_id", value));
        }
        if (problem == null) {
            throw LearningMapException.notFound("题目不存在");
        }
        if (!Integer.valueOf(1).equals(problem.getAuth())) {
            throw new LearningMapException("题目存在但未公开");
        }
        return toViews(Collections.singletonList(problem)).get(0);
    }

    public Map<Long, LearningProblemVO> getByNodes(Collection<Long> ids, Collection<String> displayIds) {
        Set<Long> problemIds = ids.stream().filter(Objects::nonNull).collect(Collectors.toSet());
        Set<String> displays = displayIds.stream().map(LearningMapRules::text)
                .filter(value -> !value.isEmpty()).collect(Collectors.toSet());
        if (problemIds.isEmpty() && displays.isEmpty()) {
            return Collections.emptyMap();
        }
        QueryWrapper<Problem> query = new QueryWrapper<Problem>().eq("auth", 1);
        query.and(wrapper -> {
            if (!problemIds.isEmpty()) {
                wrapper.in("id", problemIds);
            }
            if (!displays.isEmpty()) {
                if (!problemIds.isEmpty()) {
                    wrapper.or();
                }
                wrapper.in("problem_id", displays);
            }
        });
        return toViews(problemService.list(query)).stream()
                .collect(Collectors.toMap(LearningProblemVO::getId, Function.identity()));
    }

    private List<LearningProblemVO> toViews(List<Problem> problems) {
        if (problems.isEmpty()) {
            return new ArrayList<>();
        }
        List<Long> pids = problems.stream().map(Problem::getId).collect(Collectors.toList());
        List<ProblemTag> links = problemTagService.list(new QueryWrapper<ProblemTag>().in("pid", pids));
        Set<Long> tids = links.stream().map(ProblemTag::getTid).collect(Collectors.toSet());
        Map<Long, String> tagNames = tids.isEmpty() ? Collections.emptyMap()
                : tagService.listByIds(tids).stream().collect(Collectors.toMap(Tag::getId, Tag::getName));
        Map<Long, List<String>> tags = new HashMap<>();
        for (ProblemTag link : links) {
            String name = tagNames.get(link.getTid());
            if (name != null && !name.trim().isEmpty()) {
                tags.computeIfAbsent(link.getPid(), key -> new ArrayList<>()).add(name);
            }
        }
        tags.values().forEach(Collections::sort);
        return problems.stream().map(problem -> new LearningProblemVO()
                .setId(problem.getId())
                .setProblemDisplayId(problem.getProblemId())
                .setTitle(problem.getTitle())
                .setDifficulty(problem.getDifficulty())
                .setTags(tags.getOrDefault(problem.getId(), new ArrayList<>())))
                .collect(Collectors.toList());
    }
}
