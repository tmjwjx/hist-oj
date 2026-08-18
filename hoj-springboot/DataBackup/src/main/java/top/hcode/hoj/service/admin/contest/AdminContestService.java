package top.hcode.hoj.service.admin.contest;

;
import com.baomidou.mybatisplus.core.metadata.IPage;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestRegister;
import top.hcode.hoj.pojo.dto.ContestRegistrationUpdateDTO;
import top.hcode.hoj.pojo.vo.AdminContestVO;
import top.hcode.hoj.pojo.vo.ContestVerificationVO;

import java.util.List;


public interface AdminContestService {

    public CommonResult<IPage<Contest>> getContestList(Integer limit, Integer currentPage, String keyword);

    public CommonResult<AdminContestVO> getContest(Long cid);

    public CommonResult<Void> deleteContest(Long cid);

    public CommonResult<Void> addContest(AdminContestVO adminContestVo);

    public CommonResult<Void> cloneContest(Long cid);

    public CommonResult<Void> updateContest(AdminContestVO adminContestVo);

    public CommonResult<Void> changeContestVisible(Long cid, String uid, Boolean visible);

    public CommonResult<List<ContestRegister>> getContestRegistrations(Long cid);

    public CommonResult<Void> updateContestRegistration(ContestRegistrationUpdateDTO dto);

    public CommonResult<ContestVerificationVO> getContestVerification(Long cid);

}
