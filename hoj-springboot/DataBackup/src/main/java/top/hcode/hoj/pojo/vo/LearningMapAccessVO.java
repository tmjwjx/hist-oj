package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.List;

@Data
@Accessors(chain = true)
public class LearningMapAccessVO {
    private String accessMode;
    private List<LearningMapPermissionItemVO> permissions;
}
