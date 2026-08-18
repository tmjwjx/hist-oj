package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.contest.ContestRegisterEntityService;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.pojo.dto.ContestRegistrationUpdateDTO;
import top.hcode.hoj.pojo.dto.RegisterContestDTO;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestRegister;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.ContestRegistrationUtils;
import top.hcode.hoj.validator.ContestRegistrationValidator;
import top.hcode.hoj.validator.ContestValidator;
import top.hcode.hoj.validator.GroupValidator;

import javax.annotation.Resource;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Component
public class ContestRegistrationManager {

    @Resource
    private ContestEntityService contestEntityService;

    @Resource
    private ContestRegisterEntityService contestRegisterEntityService;

    @Resource
    private UserInfoEntityService userInfoEntityService;

    @Resource
    private ContestValidator contestValidator;

    @Resource
    private ContestRegistrationValidator registrationValidator;

    @Resource
    private GroupValidator groupValidator;

    public void register(RegisterContestDTO dto) throws StatusFailException, StatusForbiddenException {
        if (dto == null || dto.getCid() == null) {
            throw new StatusFailException("cid不能为空！");
        }

        AccountProfile user = currentUser();
        Contest contest = contestEntityService.getById(dto.getCid());
        validateContestAccess(contest, user, dto.getPassword());
        if (new Date().after(contest.getEndTime())) {
            throw new StatusFailException("比赛已经结束，无法报名！");
        }

        QueryWrapper<ContestRegister> query = new QueryWrapper<ContestRegister>()
                .eq("cid", dto.getCid()).eq("uid", user.getUid());
        if (contestRegisterEntityService.getOne(query, false) != null) {
            throw new StatusFailException("您已注册过该比赛，请勿重复注册！");
        }

        List<String> fields = ContestRegistrationUtils.fromJson(contest.getRegistrationFields());
        if (Boolean.TRUE.equals(contest.getOpenRegistration())) {
            registrationValidator.validateForm(fields, dto);
        }

        ContestRegister register = new ContestRegister()
                .setCid(dto.getCid())
                .setUid(user.getUid())
                .setStatus(0)
                .setName(dto.getName())
                .setClazz(dto.getClazz())
                .setCollege(dto.getCollege())
                .setStudentId(dto.getStudentId())
                .setGender(dto.getGender())
                .setQq(dto.getQq())
                .setPhone(dto.getPhone());
        if (!contestRegisterEntityService.save(register)) {
            throw new StatusFailException("比赛报名失败，请稍后重试！");
        }
    }

    public ContestRegister getCurrentRegistration(Long cid) throws StatusFailException {
        if (cid == null) {
            throw new StatusFailException("cid不能为空！");
        }
        AccountProfile user = currentUser();
        ContestRegister registration = contestRegisterEntityService.getOne(new QueryWrapper<ContestRegister>()
                .eq("cid", cid).eq("uid", user.getUid()), false);
        if (registration != null) {
            registration.setUsername(user.getUsername());
        }
        return registration;
    }

    public List<ContestRegister> getContestRegistrations(Long cid)
            throws StatusFailException, StatusForbiddenException {
        Contest contest = contestEntityService.getById(cid);
        if (contest == null) {
            throw new StatusFailException("该比赛不存在！");
        }
        AccountProfile user = currentUser();
        if (!SecurityUtils.getSubject().hasRole("root") && !user.getUid().equals(contest.getUid())) {
            throw new StatusForbiddenException("对不起，你无权限查看该比赛的报名信息！");
        }
        List<ContestRegister> registrations = contestRegisterEntityService.list(new QueryWrapper<ContestRegister>()
                .eq("cid", cid).orderByAsc("gmt_create"));
        if (registrations.isEmpty()) {
            return registrations;
        }
        List<String> uids = registrations.stream().map(ContestRegister::getUid).collect(Collectors.toList());
        Map<String, UserInfo> users = new HashMap<>();
        for (UserInfo info : userInfoEntityService.listByIds(uids)) {
            users.put(info.getUuid(), info);
        }
        for (ContestRegister registration : registrations) {
            UserInfo info = users.get(registration.getUid());
            registration.setUsername(info == null ? registration.getUid() : info.getUsername());
        }
        return registrations;
    }

    public void updateContestRegistration(ContestRegistrationUpdateDTO dto)
            throws StatusFailException, StatusForbiddenException {
        if (dto == null || dto.getId() == null || dto.getCid() == null) {
            throw new StatusFailException("报名记录参数不完整！");
        }
        Contest contest = contestEntityService.getById(dto.getCid());
        assertCanManage(contest);
        ContestRegister registration = contestRegisterEntityService.getOne(new QueryWrapper<ContestRegister>()
                .eq("id", dto.getId()).eq("cid", dto.getCid()), false);
        if (registration == null) {
            throw new StatusFailException("报名记录不存在！");
        }
        registration.setName(dto.getName())
                .setClazz(dto.getClazz())
                .setCollege(dto.getCollege())
                .setStudentId(dto.getStudentId())
                .setGender(dto.getGender())
                .setQq(dto.getQq())
                .setPhone(dto.getPhone());
        if (Boolean.TRUE.equals(contest.getOpenRegistration())) {
            registrationValidator.validateForm(
                    ContestRegistrationUtils.fromJson(contest.getRegistrationFields()), registration);
        }
        if (!contestRegisterEntityService.updateById(registration)) {
            throw new StatusFailException("报名信息更新失败，请稍后重试！");
        }
    }

    private void assertCanManage(Contest contest) throws StatusFailException, StatusForbiddenException {
        if (contest == null) {
            throw new StatusFailException("该比赛不存在！");
        }
        AccountProfile user = currentUser();
        if (!SecurityUtils.getSubject().hasRole("root")
                && !user.getUid().equals(contest.getUid())) {
            throw new StatusForbiddenException("对不起，你无权限修改该比赛的报名信息！");
        }
    }

    private void validateContestAccess(Contest contest, AccountProfile user, String password)
            throws StatusFailException, StatusForbiddenException {
        if (contest == null || !Boolean.TRUE.equals(contest.getVisible())) {
            throw new StatusFailException("对不起，该比赛不存在！");
        }
        if (Boolean.TRUE.equals(contest.getIsGroup())
                && !SecurityUtils.getSubject().hasRole("root")
                && !groupValidator.isGroupMember(user.getUid(), contest.getGid())) {
            throw new StatusForbiddenException("对不起，您并非团队内成员！");
        }
        if (contest.getAuth() != Constants.Contest.AUTH_PUBLIC.getCode()
                && (StringUtils.isEmpty(password) || !password.equals(contest.getPwd()))) {
            throw new StatusFailException("比赛密码错误，请重新输入！");
        }
        if (Boolean.TRUE.equals(contest.getOpenAccountLimit())
                && !contestValidator.validateAccountRule(contest.getAccountLimitRule(), user.getUsername())) {
            throw new StatusFailException("对不起！本次比赛只允许特定账号规则的用户参赛！");
        }
    }

    private AccountProfile currentUser() {
        return (AccountProfile) SecurityUtils.getSubject().getPrincipal();
    }
}
