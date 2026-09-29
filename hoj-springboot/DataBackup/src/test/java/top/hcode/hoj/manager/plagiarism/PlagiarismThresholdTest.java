package top.hcode.hoj.manager.plagiarism;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class PlagiarismThresholdTest {
    @Test
    void requiresBothDirectionalScoresToReachThreshold() {
        assertFalse(PlagiarismExecutionManager.isOverThreshold(79, 100, 80));
        assertTrue(PlagiarismExecutionManager.isOverThreshold(80, 80, 80));
        assertFalse(PlagiarismExecutionManager.isOverThreshold(100, 79, 80));
    }
}
