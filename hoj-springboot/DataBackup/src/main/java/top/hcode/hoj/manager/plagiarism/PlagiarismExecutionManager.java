package top.hcode.hoj.manager.plagiarism;

import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import top.hcode.hoj.dao.plagiarism.PlagiarismCheckEntityService;
import top.hcode.hoj.dao.plagiarism.PlagiarismResultEntityService;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.contest.ContestRecord;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheckConfig;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismResult;

import javax.annotation.Resource;
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Slf4j
@Component
public class PlagiarismExecutionManager {
    @Resource private PlagiarismManager manager;
    @Resource private PlagiarismSimilarityEngine similarity;
    @Resource private PlagiarismCheckEntityService checkService;
    @Resource private PlagiarismResultEntityService resultService;

    public void start(Long checkId) {
        PlagiarismCheck check = checkService.getById(checkId);
        if (check == null) return;
        check.setStatus("running").setStartedAt(new Date()).setErrorMessage(null);
        checkService.updateById(check);
        CompletableFuture.runAsync(() -> execute(checkId));
    }

    private void execute(Long checkId) {
        ExecutorService pool = Executors.newFixedThreadPool(Math.max(2,
                Math.min(8, Runtime.getRuntime().availableProcessors())));
        try {
            PlagiarismCheck check = checkService.getById(checkId);
            Contest contest = manager.contests().getById(check.getCid());
            if (contest == null) throw new IllegalStateException("比赛不存在");
            List<PlagiarismCheckConfig> configs = manager.allConfigs(check.getCid());
            List<ContestProblem> problems = manager.allContestProblems(check.getCid());
            List<Judge> submissions = manager.acSubmissions(check.getCid(), contest.getStartTime(), contest.getEndTime());

            check.setTotalSubmissions(submissions.size());
            Map<Long, Integer> thresholds = configs.stream().collect(Collectors.toMap(
                    PlagiarismCheckConfig::getCpid, PlagiarismCheckConfig::getThreshold, (a, b) -> b));
            Map<Long, Integer> thresholdsByPid = configs.stream().collect(Collectors.toMap(
                    PlagiarismCheckConfig::getPid, PlagiarismCheckConfig::getThreshold, (a, b) -> b));
            Map<Long, ContestProblem> problemById = problems.stream().collect(Collectors.toMap(
                    ContestProblem::getId, p -> p, (a, b) -> a));
            Map<Long, ContestProblem> problemByPid = problems.stream().collect(Collectors.toMap(
                    ContestProblem::getPid, p -> p, (a, b) -> a));
            Map<String, Long> recordSubmitIds = recordSubmitIds(manager.records(check.getCid()));

            Map<String, List<Judge>> groups = submissions.stream().collect(Collectors.groupingBy(
                    submission -> groupKey(submission), LinkedHashMap::new, Collectors.toList()));
            List<PairTask> tasks = new ArrayList<>();
            for (List<Judge> group : groups.values()) {
                for (int i = 0; i < group.size(); i++) {
                    for (int j = i + 1; j < group.size(); j++) {
                        if (!Objects.equals(group.get(i).getUid(), group.get(j).getUid())) {
                            tasks.add(new PairTask(group.get(i), group.get(j), problemById, problemByPid,
                                    thresholds, thresholdsByPid, recordSubmitIds));
                        }
                    }
                }
            }
            check.setTotalPairs(tasks.size()).setCheckedPairs(0)
                    .setProgress(tasks.isEmpty() ? 100D : 0D);
            checkService.updateById(check);
            if (tasks.isEmpty()) {
                finish(checkId, check);
                return;
            }

            List<Future<PlagiarismResult>> futures = tasks.stream()
                    .map(task -> pool.submit(() -> compare(checkId, check.getCid(), task))).collect(Collectors.toList());
            List<PlagiarismResult> pending = new ArrayList<>(50);
            AtomicInteger completed = new AtomicInteger();
            for (Future<PlagiarismResult> future : futures) {
                PlagiarismResult result = future.get();
                if (result != null) pending.add(result);
                if (pending.size() >= 50) {
                    resultService.saveBatch(pending);
                    pending.clear();
                }
                int count = completed.incrementAndGet();
                updateProgress(checkId, count, tasks.size());
            }
            if (!pending.isEmpty()) resultService.saveBatch(pending);
            finish(checkId, checkService.getById(checkId));
        } catch (Exception ex) {
            log.error("代码查重任务执行失败, checkId={}", checkId, ex);
            checkService.update(new PlagiarismCheck().setStatus("failed")
                            .setErrorMessage(ex.getMessage()).setCompletedAt(new Date()),
                    new UpdateWrapper<PlagiarismCheck>().eq("id", checkId));
        } finally {
            pool.shutdownNow();
        }
    }

    private PlagiarismResult compare(Long checkId, Long cid, PairTask task) {
        Judge first = task.first;
        Judge second = task.second;
        int[] scores = similarity.compare(first.getCode(), second.getCode(), first.getLanguage());
        int max = Math.max(scores[0], scores[1]);
        ContestProblem problem = task.problemById.get(first.getCpid());
        if (problem == null) problem = task.problemByPid.get(first.getPid());
        int threshold = task.thresholds.getOrDefault(first.getCpid(),
                task.thresholdsByPid.getOrDefault(first.getPid(), 50));
        return new PlagiarismResult().setCheckId(checkId).setCid(cid)
                .setCpid(problem == null ? first.getCpid() : problem.getId()).setPid(first.getPid())
                .setDisplayId(problem == null ? first.getDisplayPid() : problem.getDisplayId())
                .setProblemTitle(problem == null ? first.getDisplayPid() : problem.getDisplayTitle())
                .setSubmitId1(first.getSubmitId()).setSubmitId2(second.getSubmitId())
                .setContestRecordId1(task.recordSubmitIds.getOrDefault(recordKey(first), 0L))
                .setContestRecordId2(task.recordSubmitIds.getOrDefault(recordKey(second), 0L))
                .setUid1(first.getUid()).setUid2(second.getUid()).setUsername1(first.getUsername())
                .setUsername2(second.getUsername()).setLanguage(first.getLanguage())
                .setSimilarity1to2(scores[0]).setSimilarity2to1(scores[1]).setMaxSimilarity(max)
                .setIsOverThreshold(max >= threshold).setGmtCreate(new Date());
    }

    private void updateProgress(Long checkId, int checked, int total) {
        double progress = Math.min(100D, checked * 100D / total);
        checkService.update(new PlagiarismCheck().setCheckedPairs(checked).setProgress(progress),
                new UpdateWrapper<PlagiarismCheck>().eq("id", checkId));
    }

    private void finish(Long checkId, PlagiarismCheck check) {
        if (check == null) return;
        check.setStatus("completed").setProgress(100D).setCompletedAt(new Date());
        checkService.updateById(check);
    }

    private Map<String, Long> recordSubmitIds(List<ContestRecord> records) {
        Map<String, Long> result = new HashMap<>();
        for (ContestRecord record : records) {
            result.put(recordKey(record.getUid(), record.getCpid()),
                    record.getSubmitId() == null ? 0L : record.getSubmitId());
        }
        return result;
    }

    private String groupKey(Judge judge) {
        return String.valueOf(judge.getCpid()) + ":" + String.valueOf(judge.getLanguage());
    }

    private String recordKey(Judge judge) {
        return recordKey(judge.getUid(), judge.getCpid());
    }

    private String recordKey(String uid, Long cpid) {
        return String.valueOf(uid) + ":" + String.valueOf(cpid);
    }

    private static class PairTask {
        private final Judge first;
        private final Judge second;
        private final Map<Long, ContestProblem> problemById;
        private final Map<Long, ContestProblem> problemByPid;
        private final Map<Long, Integer> thresholds;
        private final Map<Long, Integer> thresholdsByPid;
        private final Map<String, Long> recordSubmitIds;

        private PairTask(Judge first, Judge second, Map<Long, ContestProblem> problemById,
                         Map<Long, ContestProblem> problemByPid, Map<Long, Integer> thresholds,
                         Map<Long, Integer> thresholdsByPid, Map<String, Long> recordSubmitIds) {
            this.first = first;
            this.second = second;
            this.problemById = problemById;
            this.problemByPid = problemByPid;
            this.thresholds = thresholds;
            this.thresholdsByPid = thresholdsByPid;
            this.recordSubmitIds = recordSubmitIds;
        }
    }
}
