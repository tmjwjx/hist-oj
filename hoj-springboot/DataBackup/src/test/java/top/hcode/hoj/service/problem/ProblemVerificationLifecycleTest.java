package top.hcode.hoj.service.problem;

import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;
import top.hcode.hoj.utils.ProblemVerificationConstants;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemVerificationLifecycleTest {
    private final ProblemVerificationLifecycle lifecycle = new ProblemVerificationLifecycle();
    private final ProblemVerificationEntityService verifications = mock(ProblemVerificationEntityService.class);
    private final ProblemTestCaseSyncManager sync = mock(ProblemTestCaseSyncManager.class);

    @BeforeEach
    void setUp() {
        ReflectionTestUtils.setField(lifecycle, "verificationService", verifications);
        ReflectionTestUtils.setField(lifecycle, "syncManager", sync);
    }

    @Test
    void changedConfigurationExplicitlyClearsPreviousAcceptedSubmissionWithoutResyncingData() {
        ProblemVerification record = new ProblemVerification().setId(7L).setPid(12L)
                .setCaseVersion("version").setSyncStatus(ProblemVerificationConstants.SYNC_SUCCESS)
                .setVerificationStatus(ProblemVerificationConstants.PASSED)
                .setSubmitId(99L).setVerifiedUid("author").setSampleVerified(true);
        when(verifications.getOne(any(), eq(false))).thenReturn(record);

        lifecycle.markVerificationChanged(12L, "version", "default");

        assertEquals(ProblemVerificationConstants.REQUIRED, record.getVerificationStatus());
        assertNull(record.getSubmitId());
        assertNull(record.getVerifiedUid());
        assertFalse(record.getSampleVerified());
        ArgumentCaptor<UpdateWrapper<ProblemVerification>> wrapper = ArgumentCaptor.forClass(UpdateWrapper.class);
        verify(verifications).update(eq(record), wrapper.capture());
        assertTrue(wrapper.getValue().getSqlSet().contains("submit_id="));
        assertTrue(wrapper.getValue().getSqlSet().contains("verified_uid="));
        assertTrue(wrapper.getValue().getSqlSegment().contains("id ="));
        assertEquals(2, Collections.frequency(wrapper.getValue().getParamNameValuePairs().values(), null));
        assertTrue(wrapper.getValue().getParamNameValuePairs().containsValue(7L));
        verifyNoInteractions(sync);
    }

    @Test
    void firstSynchronizationInsertsBeforeAnyIdScopedUpdate() throws Exception {
        when(sync.sync(12L, "version")).thenReturn(1);
        when(verifications.save(any())).thenAnswer(invocation -> {
            ProblemVerification record = invocation.getArgument(0);
            assertNull(record.getId());
            record.setId(7L);
            return true;
        });

        lifecycle.markTestCaseChanged(12L, "version", "default");

        verify(verifications).save(any());
        ArgumentCaptor<UpdateWrapper<ProblemVerification>> wrapper = ArgumentCaptor.forClass(UpdateWrapper.class);
        verify(verifications).update(any(), wrapper.capture());
        assertTrue(wrapper.getValue().getSqlSegment().contains("id ="));
        assertTrue(wrapper.getValue().getParamNameValuePairs().containsValue(7L));
    }
}
