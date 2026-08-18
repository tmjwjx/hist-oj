package top.hcode.hoj.pojo.entity.rating;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("contest_rating_status")
public class ContestRatingStatus {
    @TableId private Long contestId;
    private Boolean isRated;
    private Boolean ratingCalculated;
    private Date calculatedAt;
    private Integer skipCount;
    private Boolean hasPendingSkip;
    private String recalculateStatus;
    private Date lastRecalculateAt;
    private Boolean recalculateLock;
    private Date skipDataChangedAt;
    private Date createdAt;
    private Date updatedAt;
}
