package top.hcode.hoj.pojo.dto;

import lombok.Data;

@Data
public class ExternalVerificationSubmitDTO {
    private String visitorToken;
    private String nickname;
    private String language;
    private String code;
}
