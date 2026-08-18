package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;

@Data
public class BattleExcludeDTO {

    @NotNull(message = "记录ID不能为空")
    private Long recordId;

    @NotNull(message = "计入状态不能为空")
    private Boolean isExcluded;
}
