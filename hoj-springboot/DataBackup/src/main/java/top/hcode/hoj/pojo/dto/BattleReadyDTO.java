package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
public class BattleReadyDTO {

    @NotBlank(message = "房间号不能为空")
    private String roomId;

    @NotNull(message = "准备状态不能为空")
    private Boolean ready;
}
