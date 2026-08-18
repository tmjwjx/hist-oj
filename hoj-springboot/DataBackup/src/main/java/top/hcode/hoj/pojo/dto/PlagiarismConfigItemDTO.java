package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.constraints.Max;
import javax.validation.constraints.Min;
import javax.validation.constraints.NotNull;

@Data
public class PlagiarismConfigItemDTO {
    @NotNull
    private Long cpid;
    @Min(0)
    @Max(100)
    private Integer threshold = 50;
}
