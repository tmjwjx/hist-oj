package top.hcode.hoj.service.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.ProblemVerificationConstants;

@Component
public class ProblemVerificationLifecycle {

    @Autowired
    private ProblemVerificationEntityService verificationService;

    @Autowired
    private ProblemTestCaseSyncManager syncManager;

    public void markTestCaseChanged(Long pid, String version, String judgeMode) {
        ProblemVerification record = find(pid);
        if (sameSyncedVersion(record, version)) {
            resetVerification(record, judgeMode, "测试数据已导入，无需重复同步");
            verificationService.updateById(record);
            return;
        }
        record = prepare(record, pid, version, judgeMode, "正在导入并同步测试数据");
        verificationService.saveOrUpdate(record);
        try {
            int count = syncManager.sync(pid, version);
            record.setSyncStatus(ProblemVerificationConstants.SYNC_SUCCESS)
                    .setSyncMessage("测试数据已导入并同步到 " + count + " 台判题服务器");
        } catch (Exception e) {
            record.setSyncStatus(ProblemVerificationConstants.SYNC_FAILED)
                    .setSyncMessage(message(e));
        }
        verificationService.updateById(record);
    }

    /** 仅刷新 info 元数据；测试文件本身不重复传输。 */
    public void markTestCaseMetadataChanged(Long pid, String version, String judgeMode) {
        ProblemVerification record = prepare(find(pid), pid, version, judgeMode, "正在更新判题配置");
        verificationService.saveOrUpdate(record);
        try {
            int count = syncManager.syncInfo(pid, version);
            record.setSyncStatus(ProblemVerificationConstants.SYNC_SUCCESS)
                    .setSyncMessage("判题配置已更新到 " + count + " 台判题服务器，测试数据未重复传输");
        } catch (Exception metadataError) {
            try {
                int count = syncManager.sync(pid, version);
                record.setSyncStatus(ProblemVerificationConstants.SYNC_SUCCESS)
                        .setSyncMessage("判题服务器缺少原数据，已补充同步到 " + count + " 台服务器");
            } catch (Exception fullSyncError) {
                record.setSyncStatus(ProblemVerificationConstants.SYNC_FAILED)
                        .setSyncMessage(message(fullSyncError));
            }
        }
        verificationService.updateById(record);
    }

    /** 题面、时限等变化只要求重新验题，不传输已同步的测试包。 */
    public void markVerificationChanged(Long pid, String version, String judgeMode) {
        ProblemVerification record = find(pid);
        if (!sameSyncedVersion(record, version)) {
            markTestCaseChanged(pid, version, judgeMode);
            return;
        }
        resetVerification(record, judgeMode, "测试数据未变化，无需重复同步；请重新提交标准程序验题");
        verificationService.updateById(record);
    }

    private ProblemVerification find(Long pid) {
        return verificationService.getOne(new QueryWrapper<ProblemVerification>().eq("pid", pid), false);
    }

    private boolean sameSyncedVersion(ProblemVerification record, String version) {
        return record != null && java.util.Objects.equals(record.getCaseVersion(), version)
                && java.util.Objects.equals(record.getSyncStatus(), ProblemVerificationConstants.SYNC_SUCCESS);
    }

    private ProblemVerification prepare(ProblemVerification record, Long pid, String version,
                                        String judgeMode, String message) {
        if (record == null) record = new ProblemVerification().setPid(pid);
        record.setCaseVersion(version).setSyncStatus(ProblemVerificationConstants.SYNC_PENDING);
        resetVerification(record, judgeMode, message);
        return record;
    }

    private void resetVerification(ProblemVerification record, String judgeMode, String message) {
        record.setJudgeMode(StringUtils.isEmpty(judgeMode) ? Constants.JudgeMode.DEFAULT.getMode() : judgeMode)
                .setSyncMessage(message).setVerificationStatus(ProblemVerificationConstants.REQUIRED)
                .setSubmitId(null).setVerifiedUid(null).setSampleVerified(false);
    }

    private String message(Exception e) {
        String message = e.getMessage();
        if (message == null || message.trim().isEmpty()) {
            return "未知错误";
        }
        message = message.trim();
        return message.length() > 900 ? message.substring(0, 900) : message;
    }
}
