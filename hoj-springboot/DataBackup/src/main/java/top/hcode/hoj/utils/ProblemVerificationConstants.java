package top.hcode.hoj.utils;

public final class ProblemVerificationConstants {

    public static final int SYNC_PENDING = 0;
    public static final int SYNC_SUCCESS = 1;
    public static final int SYNC_FAILED = 2;

    public static final int REQUIRED = 0;
    public static final int JUDGING = 1;
    public static final int PASSED = 2;
    public static final int FAILED = 3;

    public static final String AI_VALIDATION = "ai_validation";
    public static final String CREATOR_VALIDATION = "creator_validation";
    public static final String USER_SUBMISSION = "user_submission";

    private ProblemVerificationConstants() {
    }

    public static boolean isJudgeRunning(Integer status) {
        return status != null && (status == Constants.Judge.STATUS_PENDING.getStatus()
                || status == Constants.Judge.STATUS_COMPILING.getStatus()
                || status == Constants.Judge.STATUS_JUDGING.getStatus()
                || status == Constants.Judge.STATUS_SUBMITTING.getStatus());
    }
}
