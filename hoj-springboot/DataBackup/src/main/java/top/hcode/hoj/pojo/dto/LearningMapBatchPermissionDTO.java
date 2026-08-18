package top.hcode.hoj.pojo.dto;

import lombok.Data;

import java.util.List;

@Data
public class LearningMapBatchPermissionDTO {
    private List<String> userIds;
    private Boolean enabled;
}
