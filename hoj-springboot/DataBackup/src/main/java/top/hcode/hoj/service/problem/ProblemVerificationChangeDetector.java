package top.hcode.hoj.service.problem;

import top.hcode.hoj.pojo.entity.problem.Problem;

import java.util.Objects;

/** 判断题目改动是否会影响标准程序验题结果。 */
public final class ProblemVerificationChangeDetector {
    private ProblemVerificationChangeDetector() {
    }

    public static boolean changed(Problem before, Problem after) {
        if (before == null || after == null) return false;
        return !Objects.equals(before.getCaseVersion(), after.getCaseVersion())
                || !Objects.equals(before.getExamples(), after.getExamples())
                || !Objects.equals(before.getType(), after.getType())
                || !Objects.equals(before.getJudgeMode(), after.getJudgeMode())
                || !Objects.equals(before.getJudgeCaseMode(), after.getJudgeCaseMode())
                || !Objects.equals(before.getSpjCode(), after.getSpjCode())
                || !Objects.equals(before.getSpjLanguage(), after.getSpjLanguage())
                || !Objects.equals(before.getTimeLimit(), after.getTimeLimit())
                || !Objects.equals(before.getMemoryLimit(), after.getMemoryLimit())
                || !Objects.equals(before.getStackLimit(), after.getStackLimit())
                || !Objects.equals(before.getIsRemoveEndBlank(), after.getIsRemoveEndBlank())
                || !Objects.equals(before.getIsFileIO(), after.getIsFileIO())
                || !Objects.equals(before.getIoReadFileName(), after.getIoReadFileName())
                || !Objects.equals(before.getIoWriteFileName(), after.getIoWriteFileName())
                || !Objects.equals(before.getUserExtraFile(), after.getUserExtraFile())
                || !Objects.equals(before.getJudgeExtraFile(), after.getJudgeExtraFile());
    }
}
