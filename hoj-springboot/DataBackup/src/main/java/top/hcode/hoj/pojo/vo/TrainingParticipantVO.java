package top.hcode.hoj.pojo.vo;

import lombok.Data;

import java.util.Date;

@Data
public class TrainingParticipantVO {
    private Long id;
    private Long trainingId;
    private String uid;
    private String status;
    private Date joinTime;
    private Date gmtCreate;
    private Date gmtModified;
    private TrainingParticipantUserVO user;
    private TrainingProgressVO progress;
}
