package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;

@Data
public class BattleRoomDTO {

    @NotBlank(message = "房间号不能为空")
    private String roomId;
}
