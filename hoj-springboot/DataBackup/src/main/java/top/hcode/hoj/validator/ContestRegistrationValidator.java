package top.hcode.hoj.validator;

import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.pojo.dto.RegisterContestDTO;
import top.hcode.hoj.pojo.entity.contest.ContestRegister;
import top.hcode.hoj.pojo.vo.AdminContestVO;
import top.hcode.hoj.utils.ContestRegistrationUtils;

import java.util.List;

@Component
public class ContestRegistrationValidator {

    public void validateConfig(AdminContestVO contest) throws StatusFailException {
        if (!Boolean.TRUE.equals(contest.getOpenRegistration())) {
            contest.setUseRegistrationName(false);
            contest.setRegistrationNameFields(null);
            return;
        }

        List<String> fields = ContestRegistrationUtils.normalize(contest.getRegistrationFields());
        if (fields.isEmpty()) {
            throw new StatusFailException("开启比赛报名后，至少选择一个报名字段！");
        }
        contest.setRegistrationFields(fields);

        if (Boolean.TRUE.equals(contest.getUseRegistrationName())) {
            List<String> nameFields = ContestRegistrationUtils.normalize(contest.getRegistrationNameFields());
            if (nameFields.isEmpty() || !fields.containsAll(nameFields)) {
                throw new StatusFailException("比赛内名称只能由已启用的报名字段组成！");
            }
            contest.setRegistrationNameFields(nameFields);
        }
    }

    public void validateForm(List<String> fields, RegisterContestDTO dto) throws StatusFailException {
        for (String field : fields) {
            String value = ContestRegistrationUtils.getValue(dto, field);
            validateValue(field, value);
        }
    }

    public void validateForm(List<String> fields, ContestRegister register) throws StatusFailException {
        for (String field : fields) {
            validateValue(field, ContestRegistrationUtils.getValue(register, field));
        }
    }

    private void validateValue(String field, String value) throws StatusFailException {
        if (StringUtils.isEmpty(value)) {
            throw new StatusFailException("报名字段“" + fieldLabel(field) + "”不能为空！");
        }
        if (value.trim().length() > maxLength(field)) {
            throw new StatusFailException("报名字段“" + fieldLabel(field) + "”内容过长！");
        }
    }

    private int maxLength(String field) {
        if ("studentId".equals(field)) return 50;
        if ("gender".equals(field)) return 10;
        if ("qq".equals(field)) return 20;
        if ("phone".equals(field)) return 30;
        return 100;
    }

    private String fieldLabel(String field) {
        switch (field) {
            case "name": return "姓名";
            case "class": return "班级";
            case "college": return "学院";
            case "studentId": return "学号";
            case "gender": return "性别";
            case "qq": return "QQ";
            case "phone": return "电话号码";
            default: return field;
        }
    }
}
