package top.hcode.hoj.pojo.vo.plagiarism;

import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;

import java.io.Serializable;
import java.util.Date;
import java.util.List;

/**
 * 比赛设备异常筛查结果。
 * 仅基于 judge.device_id 聚合，不使用IP（校园网NAT会导致同出口大量误判）。
 */
@Data
@ApiModel(value = "DeviceAnomalyVO", description = "比赛设备异常筛查结果")
public class DeviceAnomalyVO implements Serializable {

    private static final long serialVersionUID = 1L;

    @ApiModelProperty(value = "比赛id")
    private Long cid;

    @ApiModelProperty(value = "比赛时间窗内的提交总数")
    private Long totalSubmissions;

    @ApiModelProperty(value = "未采集到设备ID的提交数（历史数据），不参与聚合")
    private Long unknownDeviceSubmissions;

    @ApiModelProperty(value = "作弊嫌疑账号数（命中任一规则的账号去重数）")
    private Integer suspiciousAccountCount;

    @ApiModelProperty(value = "嫌疑账号清单（两种规则命中的并集）")
    private List<SuspiciousAccount> suspiciousAccounts;

    @ApiModelProperty(value = "作弊团伙：嫌疑账号通过共用设备两两连通形成的分组（传递闭包）")
    private List<CheatCluster> clusters;

    @ApiModelProperty(value = "被多个账号使用的设备清单")
    private List<SharedDevice> sharedDevices;

    @ApiModelProperty(value = "生成时间")
    private Date generatedAt;

    @Data
    @ApiModel(value = "SuspiciousAccount", description = "命中作弊规则的账号")
    public static class SuspiciousAccount implements Serializable {

        private static final long serialVersionUID = 1L;

        @ApiModelProperty(value = "用户id")
        private String uid;

        @ApiModelProperty(value = "用户名")
        private String username;

        @ApiModelProperty(value = "是否命中：一个账号使用多台设备提交")
        private Boolean multiDevice;

        @ApiModelProperty(value = "是否命中：该账号使用了被多个账号共用的设备")
        private Boolean deviceSharing;

        @ApiModelProperty(value = "使用的不同设备数量")
        private Integer deviceCount;

        @ApiModelProperty(value = "比赛内提交总数")
        private Integer submissionCount;

        @ApiModelProperty(value = "首次提交时间")
        private Date firstSubmitTime;

        @ApiModelProperty(value = "最后提交时间")
        private Date lastSubmitTime;

        @ApiModelProperty(value = "使用的设备明细")
        private List<DeviceUsage> devices;

        @ApiModelProperty(value = "所属团伙编号")
        private Integer clusterId;

        @ApiModelProperty(value = "团伙内其它关联账号（沿共用设备逐层推进的传递闭包）")
        private List<String> relatedUsernames;
    }

    @Data
    @ApiModel(value = "DeviceUsage", description = "账号在某设备上的使用明细")
    public static class DeviceUsage implements Serializable {

        private static final long serialVersionUID = 1L;

        @ApiModelProperty(value = "设备ID")
        private String deviceId;

        @ApiModelProperty(value = "该设备的User-Agent")
        private String userAgent;

        @ApiModelProperty(value = "该账号在该设备上的提交数")
        private Integer submissionCount;

        @ApiModelProperty(value = "首次提交时间")
        private Date firstSubmitTime;

        @ApiModelProperty(value = "最后提交时间")
        private Date lastSubmitTime;

        @ApiModelProperty(value = "该设备上同时使用过的其它账号（仅共用设备时非空）")
        private List<String> otherUsernames;
    }

    @Data
    @ApiModel(value = "SharedDevice", description = "被多个账号使用的设备")
    public static class SharedDevice implements Serializable {

        private static final long serialVersionUID = 1L;

        @ApiModelProperty(value = "设备ID")
        private String deviceId;

        @ApiModelProperty(value = "该设备的User-Agent")
        private String userAgent;

        @ApiModelProperty(value = "使用过该设备的账号数")
        private Integer userCount;

        @ApiModelProperty(value = "该设备的提交总数")
        private Integer submissionCount;

        @ApiModelProperty(value = "首次提交时间")
        private Date firstSubmitTime;

        @ApiModelProperty(value = "最后提交时间")
        private Date lastSubmitTime;

        @ApiModelProperty(value = "使用该设备的账号明细")
        private List<UserUsage> users;
    }

    @Data
    @ApiModel(value = "UserUsage", description = "账号在某设备上的提交概况")
    public static class UserUsage implements Serializable {

        private static final long serialVersionUID = 1L;

        @ApiModelProperty(value = "用户id")
        private String uid;

        @ApiModelProperty(value = "用户名")
        private String username;

        @ApiModelProperty(value = "该账号在该设备上的提交数")
        private Integer submissionCount;

        @ApiModelProperty(value = "首次提交时间")
        private Date firstSubmitTime;

        @ApiModelProperty(value = "最后提交时间")
        private Date lastSubmitTime;
    }

    @Data
    @ApiModel(value = "CheatCluster", description = "作弊团伙：账号经共用设备传递连通的分组")
    public static class CheatCluster implements Serializable {

        private static final long serialVersionUID = 1L;

        @ApiModelProperty(value = "团伙编号")
        private Integer clusterId;

        @ApiModelProperty(value = "团伙成员账号（按用户名排序）")
        private List<String> usernames;

        @ApiModelProperty(value = "团伙内的关联关系（每个共用设备一条）")
        private List<ClusterLink> links;
    }

    @Data
    @ApiModel(value = "ClusterLink", description = "团伙内一条共用设备关联")
    public static class ClusterLink implements Serializable {

        private static final long serialVersionUID = 1L;

        @ApiModelProperty(value = "共用设备ID")
        private String deviceId;

        @ApiModelProperty(value = "该设备上的团伙账号（按用户名排序，≥2个）")
        private List<String> usernames;
    }
}
