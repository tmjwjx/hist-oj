package top.hcode.hoj.pojo.entity.rating;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("rating_history")
public class RatingHistory {
    @TableId(type = IdType.AUTO) private Long id;
    private String uid;
    @JsonProperty("contest_id") @TableField("contest_id") private Long contestId;
    @TableField("related_contest_id") private Long relatedContestId;
    @JsonProperty("old_rating") @TableField("old_rating") private Integer oldRating;
    @JsonProperty("new_rating") @TableField("new_rating") private Integer newRating;
    @JsonProperty("rating_change") @TableField("rating_change") private Integer ratingChange;
    /** MySQL 8 treats RANK as a reserved keyword; keep the legacy column name quoted. */
    @TableField("`rank`")
    private Integer rank;
    private Integer participants;
    private String reason;
    @JsonProperty("is_manual") @TableField("is_manual") private Boolean isManual;
    @TableField("is_skip") private Boolean isSkip;
    @TableField("skip_reason") private String skipReason;
    @JsonProperty("operator_uid") @TableField("operator_uid") private String operatorUid;
    @JsonProperty("created_at") @TableField("created_at") private Date createdAt;

    @JsonProperty("contest_title") @TableField(exist = false) private String contestTitle;
    @JsonProperty("contest_time") @TableField(exist = false) private Date contestTime;
    @TableField(exist = false) private String manualAdjustReason;
    @TableField(exist = false) private Integer manualAdjustDelta;
    @TableField(exist = false) private String username;
    @TableField(exist = false) private Boolean canceled;
    @TableField(exist = false) private Date canceledAt;
}
