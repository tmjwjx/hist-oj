package top.hcode.hoj.manager.plagiarism;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

class PlagiarismSimilarityEngineTest {
    @Test
    void ignoresCommentsAndRenamedIdentifiersInFallback() {
        System.setProperty("hoj.sim.disable", "true");
        PlagiarismSimilarityEngine engine = new PlagiarismSimilarityEngine();
        int[] score = engine.compare("// one\nint sum(int a){return a+1;}",
                "int value(int b){return b+1;}", "C++");
        assertEquals(100, score[0]);
        assertEquals(100, score[1]);
        System.clearProperty("hoj.sim.disable");
    }

    @Test
    void skipsLanguagesWithoutSimSupport() {
        PlagiarismSimilarityEngine engine = new PlagiarismSimilarityEngine();
        int[] score = engine.compare("print(1)", "print(1)", "Python3");
        assertEquals(0, score[0]);
        assertEquals(0, score[1]);
    }
}
