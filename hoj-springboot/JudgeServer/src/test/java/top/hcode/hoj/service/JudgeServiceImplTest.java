package top.hcode.hoj.service;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.dao.JudgeEntityService;
import top.hcode.hoj.dao.ProblemEntityService;
import top.hcode.hoj.judge.JudgeContext;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.service.impl.JudgeServiceImpl;

import java.util.Arrays;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.same;
import static org.mockito.Mockito.*;

class JudgeServiceImplTest {
    @Test
    void creatorVerificationLoadsStatementSamplesForFormalJudging() {
        JudgeEntityService judges = mock(JudgeEntityService.class);
        ProblemEntityService problems = mock(ProblemEntityService.class);
        JudgeContext context = mock(JudgeContext.class);
        JudgeServiceImpl service = new JudgeServiceImpl();
        ReflectionTestUtils.setField(service, "judgeEntityService", judges);
        ReflectionTestUtils.setField(service, "problemEntityService", problems);
        ReflectionTestUtils.setField(service, "judgeContext", context);

        String examples = "<input>5 2 1 1\n1 2\n3 5\n2 3</input><output>0</output>"
                + "<input>8 4 2 2\n1 3\n2 5\n4 7\n6 8\n2 6\n5 8</input><output>1\n1</output>";
        Judge judge = new Judge().setSubmitId(13527L).setPid(1763L)
                .setIsProblemVerification(true).setSubmissionType("creator_validation");
        when(judges.update(any())).thenReturn(true);
        when(problems.getOne(any())).thenAnswer(invocation -> {
            QueryWrapper<Problem> query = invocation.getArgument(0);
            // Model database projection: unselected fields are absent from the returned entity.
            return new Problem().setId(1763L).setExamples(
                    Arrays.asList(query.getSqlSelect().split(",")).contains("examples") ? examples : null);
        });
        when(context.Judge(any(), same(judge))).thenAnswer(invocation -> {
            Problem loaded = invocation.getArgument(0);
            assertEquals(examples, loaded.getExamples(), "Formal verification must receive saved statement samples");
            return judge;
        });

        service.judge(judge);

        verify(context).Judge(any(), same(judge));
        verify(judges).updateById(judge);
    }
}
