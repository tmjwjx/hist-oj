package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ExternalVerificationLinkVO {
    private Long pid;
    private String token;
    private Boolean enabled;
}
