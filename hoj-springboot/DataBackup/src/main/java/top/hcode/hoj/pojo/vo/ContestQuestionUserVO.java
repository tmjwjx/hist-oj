package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class ContestQuestionUserVO {
    private String uid;
    private String username;
    private String nickname;
    private String avatar;
    private String titleName;
    private String titleColor;
    private Integer histRating;
}
