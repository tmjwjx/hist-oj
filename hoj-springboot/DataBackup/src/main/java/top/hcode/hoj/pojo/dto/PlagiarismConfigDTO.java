package top.hcode.hoj.pojo.dto;

import lombok.Data;

import javax.validation.Valid;
import java.util.List;

@Data
public class PlagiarismConfigDTO {
    @Valid
    private List<PlagiarismConfigItemDTO> configs;
}
