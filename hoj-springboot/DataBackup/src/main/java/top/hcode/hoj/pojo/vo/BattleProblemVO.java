package top.hcode.hoj.pojo.vo;

import lombok.Data;
import lombok.experimental.Accessors;

@Data
@Accessors(chain = true)
public class BattleProblemVO {

    private Long pid;

    private String displayId;

    private String problemId;

    private String title;

    private Integer difficulty;

    private Integer auth;
}
