package top.hcode.hoj.manager.plagiarism;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;

class PlagiarismDolosWorkerClientTest {
    @Test
    void normalizesJudgeLanguagesIntoDolosFamilies() {
        PlagiarismDolosWorkerClient client = new PlagiarismDolosWorkerClient();
        assertEquals("python", client.normalizeLanguage("PyPy3"));
        assertEquals("javascript", client.normalizeLanguage("JavaScript Node"));
        assertEquals("cpp", client.normalizeLanguage("GNU C++17"));
        assertEquals("go", client.normalizeLanguage("Golang 1.22"));
    }

    @Test
    void parsesDirectionalScoresAndKeepsIdOrientationStable() throws Exception {
        PlagiarismDolosWorkerClient client = new PlagiarismDolosWorkerClient();
        Map<String, int[]> scores = client.parsePairs("{\"pairs\":["
                + "{\"submissionId1\":\"20\",\"submissionId2\":\"10\","
                + "\"similarity1to2\":81,\"similarity2to1\":73}]}" );
        assertArrayEquals(new int[]{73, 81}, scores.get(PlagiarismDolosWorkerClient.pairKey("10", "20")));
    }

    @Test
    void acceptsSingleSymmetricSimilarityAndNestedPairs() throws Exception {
        PlagiarismDolosWorkerClient client = new PlagiarismDolosWorkerClient();
        Map<String, int[]> scores = client.parsePairs("{\"data\":{\"pairs\":["
                + "{\"left_id\":\"a\",\"right_id\":\"b\",\"similarity\":88}]}}" );
        assertArrayEquals(new int[]{88, 88}, scores.get(PlagiarismDolosWorkerClient.pairKey("a", "b")));
    }

    @Test
    void parsesHistOjWorkerResponseWithDolosCoverageFields() throws Exception {
        PlagiarismDolosWorkerClient client = new PlagiarismDolosWorkerClient();
        Map<String, int[]> scores = client.parsePairs("{\"pairs\":["
                + "{\"leftSubmissionId\":\"submit-1\",\"rightSubmissionId\":\"submit-2\","
                + "\"similarity\":0.74,\"similarityPercent\":74,"
                + "\"similarityLeftToRight\":0.92,\"similarityRightToLeft\":0.81}]}" );
        assertArrayEquals(new int[]{92, 81},
                scores.get(PlagiarismDolosWorkerClient.pairKey("submit-1", "submit-2")));
    }
}
