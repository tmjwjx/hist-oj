package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;
import top.hcode.hoj.dao.contest.ContestRegisterEntityService;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestRegister;
import top.hcode.hoj.utils.ContestRegistrationUtils;

import javax.annotation.Resource;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class ContestRegistrationNameManager {

    @Resource
    private ContestRegisterEntityService contestRegisterEntityService;

    public Map<String, String> getNames(Contest contest, List<String> uids) {
        if (!Boolean.TRUE.equals(contest.getUseRegistrationName()) || CollectionUtils.isEmpty(uids)) {
            return Collections.emptyMap();
        }

        List<String> fields = ContestRegistrationUtils.fromJson(contest.getRegistrationNameFields());
        if (fields.isEmpty()) {
            return Collections.emptyMap();
        }

        List<ContestRegister> registrations = contestRegisterEntityService.list(
                new QueryWrapper<ContestRegister>().eq("cid", contest.getId()).in("uid", uids));
        Map<String, String> names = new HashMap<>();
        for (ContestRegister registration : registrations) {
            String name = ContestRegistrationUtils.composeName(registration, fields);
            if (!StringUtils.isEmpty(name)) {
                names.put(registration.getUid(), name);
            }
        }
        return names;
    }
}
