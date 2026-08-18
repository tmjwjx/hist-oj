package top.hcode.hoj.utils;

import cn.hutool.json.JSONUtil;
import org.springframework.util.StringUtils;
import top.hcode.hoj.pojo.dto.RegisterContestDTO;
import top.hcode.hoj.pojo.entity.contest.ContestRegister;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Set;

public final class ContestRegistrationUtils {

    public static final List<String> SUPPORTED_FIELDS = Arrays.asList(
            "name", "class", "college", "studentId", "gender", "qq", "phone");

    private ContestRegistrationUtils() {
    }

    public static List<String> normalize(List<String> fields) {
        Set<String> normalized = new LinkedHashSet<>();
        if (fields != null) {
            for (String field : fields) {
                if (SUPPORTED_FIELDS.contains(field)) {
                    normalized.add(field);
                }
            }
        }
        return new ArrayList<>(normalized);
    }

    public static String toJson(List<String> fields) {
        return JSONUtil.toJsonStr(normalize(fields));
    }

    public static List<String> fromJson(String json) {
        if (StringUtils.isEmpty(json)) {
            return new ArrayList<>();
        }
        try {
            return normalize(JSONUtil.toList(JSONUtil.parseArray(json), String.class));
        } catch (Exception ignored) {
            return new ArrayList<>();
        }
    }

    public static String getValue(RegisterContestDTO dto, String field) {
        switch (field) {
            case "name": return dto.getName();
            case "class": return dto.getClazz();
            case "college": return dto.getCollege();
            case "studentId": return dto.getStudentId();
            case "gender": return dto.getGender();
            case "qq": return dto.getQq();
            case "phone": return dto.getPhone();
            default: return null;
        }
    }

    public static String getValue(ContestRegister register, String field) {
        switch (field) {
            case "name": return register.getName();
            case "class": return register.getClazz();
            case "college": return register.getCollege();
            case "studentId": return register.getStudentId();
            case "gender": return register.getGender();
            case "qq": return register.getQq();
            case "phone": return register.getPhone();
            default: return null;
        }
    }

    public static String composeName(ContestRegister register, List<String> fields) {
        List<String> values = new ArrayList<>();
        for (String field : normalize(fields)) {
            String value = getValue(register, field);
            if (!StringUtils.isEmpty(value)) {
                values.add(value.trim());
            }
        }
        return String.join(" ", values);
    }
}
