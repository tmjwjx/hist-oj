package top.hcode.hoj.manager.plagiarism;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO.DeviceUsage;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO.SharedDevice;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO.SuspiciousAccount;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO.UserUsage;

import javax.annotation.Resource;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.Date;
import java.util.Deque;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.stream.Collectors;

/**
 * 比赛设备异常筛查。
 * 仅依赖 judge.device_id（前端持久化的随机UUID）聚合，不使用IP：
 * 校园网NAT会让大量正常用户共享同一出口IP，参与规则必然误判。
 * device_id 为空的提交（迁移前历史数据）只计入 unknownDeviceSubmissions，不参与聚合。
 */
@Component
public class DeviceAnomalyManager {

    @Resource private JudgeEntityService judgeService;
    @Resource private PlagiarismManager plagiarismManager;

    public DeviceAnomalyVO deviceAnomalies(Long cid)
            throws StatusNotFoundException, StatusForbiddenException {
        Contest contest = plagiarismManager.requireAccess(cid);
        QueryWrapper<Judge> wrapper = new QueryWrapper<Judge>()
                .eq("cid", cid)
                .select("submit_id", "uid", "username", "pid", "device_id", "user_agent", "submit_time");
        if (contest.getStartTime() != null) {
            wrapper.ge("submit_time", contest.getStartTime());
        }
        if (contest.getEndTime() != null) {
            wrapper.le("submit_time", contest.getEndTime());
        }
        List<Judge> judges = judgeService.list(wrapper);
        return aggregate(cid, judges);
    }

    /**
     * 纯内存聚合，不依赖Spring上下文，便于单测。
     */
    static DeviceAnomalyVO aggregate(Long cid, List<Judge> judges) {
        DeviceAnomalyVO vo = new DeviceAnomalyVO();
        vo.setCid(cid);
        vo.setGeneratedAt(new Date());
        vo.setTotalSubmissions((long) judges.size());

        Map<String, DeviceRecord> deviceRecords = new LinkedHashMap<>();
        Map<String, AccountRecord> accountRecords = new TreeMap<>();
        long unknownDeviceSubmissions = 0;

        for (Judge judge : judges) {
            if (judge.getDeviceId() == null || judge.getDeviceId().trim().isEmpty()) {
                unknownDeviceSubmissions++;
                continue;
            }
            String username = judge.getUsername() == null ? "" : judge.getUsername();
            deviceRecords.computeIfAbsent(judge.getDeviceId(), k -> new DeviceRecord(k))
                    .add(judge, username);
            accountRecords.computeIfAbsent(judge.getUid(), k -> new AccountRecord(k, username))
                    .add(judge);
        }
        vo.setUnknownDeviceSubmissions(unknownDeviceSubmissions);

        // 一台设备多个账号
        Map<String, SharedDevice> sharedDevices = new LinkedHashMap<>();
        for (DeviceRecord device : deviceRecords.values()) {
            if (device.usernames.size() <= 1) {
                continue;
            }
            SharedDevice shared = new SharedDevice();
            shared.setDeviceId(device.deviceId);
            shared.setUserAgent(device.latestUserAgent);
            shared.setUserCount(device.usernames.size());
            shared.setSubmissionCount(device.submissionCount);
            shared.setFirstSubmitTime(device.firstSubmitTime);
            shared.setLastSubmitTime(device.lastSubmitTime);
            List<UserUsage> users = new ArrayList<>();
            for (String uid : device.uidOrder) {
                AccountRecord account = accountRecords.get(uid);
                UserUsage usage = new UserUsage();
                usage.setUid(uid);
                usage.setUsername(account == null ? "" : account.username);
                AccountRecord.DeviceStat stat = account == null ? null : device.stats.get(uid);
                usage.setSubmissionCount(stat == null ? 0 : stat.submissionCount);
                usage.setFirstSubmitTime(stat == null ? null : stat.firstSubmitTime);
                usage.setLastSubmitTime(stat == null ? null : stat.lastSubmitTime);
                users.add(usage);
            }
            users.sort(Comparator.comparingInt(UserUsage::getSubmissionCount).reversed());
            shared.setUsers(users);
            sharedDevices.put(device.deviceId, shared);
        }
        List<SharedDevice> sharedDeviceList = new ArrayList<>(sharedDevices.values());
        sharedDeviceList.sort(Comparator.comparingInt(SharedDevice::getUserCount).reversed()
                .thenComparingInt(SharedDevice::getSubmissionCount).reversed());
        vo.setSharedDevices(sharedDeviceList);

        // 一账号多设备，及账号使用了共用设备
        List<SuspiciousAccount> suspiciousAccounts = new ArrayList<>();
        for (AccountRecord account : accountRecords.values()) {
            boolean multiDevice = account.devices.size() > 1;
            boolean deviceSharing = account.devices.keySet().stream().anyMatch(sharedDevices::containsKey);
            if (!multiDevice && !deviceSharing) {
                continue;
            }
            SuspiciousAccount suspicious = new SuspiciousAccount();
            suspicious.setUid(account.uid);
            suspicious.setUsername(account.username);
            suspicious.setMultiDevice(multiDevice);
            suspicious.setDeviceSharing(deviceSharing);
            suspicious.setDeviceCount(account.devices.size());
            suspicious.setSubmissionCount(account.submissionCount);
            suspicious.setFirstSubmitTime(account.firstSubmitTime);
            suspicious.setLastSubmitTime(account.lastSubmitTime);
            List<DeviceUsage> deviceUsages = new ArrayList<>();
            for (Map.Entry<String, AccountRecord.DeviceStat> entry : account.devices.entrySet()) {
                DeviceRecord device = deviceRecords.get(entry.getKey());
                DeviceUsage usage = new DeviceUsage();
                usage.setDeviceId(entry.getKey());
                usage.setUserAgent(device == null ? null : device.latestUserAgent);
                usage.setSubmissionCount(entry.getValue().submissionCount);
                usage.setFirstSubmitTime(entry.getValue().firstSubmitTime);
                usage.setLastSubmitTime(entry.getValue().lastSubmitTime);
                if (device != null && device.usernames.size() > 1) {
                    Set<String> others = new TreeSet<>(device.usernames);
                    others.remove(account.username);
                    usage.setOtherUsernames(new ArrayList<>(others));
                } else {
                    usage.setOtherUsernames(new ArrayList<>());
                }
                deviceUsages.add(usage);
            }
            deviceUsages.sort(Comparator.comparing(DeviceUsage::getFirstSubmitTime,
                    Comparator.nullsLast(Comparator.naturalOrder())));
            suspicious.setDevices(deviceUsages);
            suspiciousAccounts.add(suspicious);
        }
        suspiciousAccounts.sort(Comparator
                .comparingInt((SuspiciousAccount account) ->
                        (account.getMultiDevice() && account.getDeviceSharing()) ? 0 : 1)
                .thenComparingInt(SuspiciousAccount::getDeviceCount).reversed()
                .thenComparingInt(SuspiciousAccount::getSubmissionCount).reversed());
        vo.setSuspiciousAccounts(suspiciousAccounts);
        vo.setSuspiciousAccountCount(suspiciousAccounts.size());
        vo.setClusters(buildClusters(suspiciousAccounts, sharedDeviceList));
        return vo;
    }

    /**
     * 团伙聚类：嫌疑账号之间通过共用设备连通，BFS 逐层推进到传递闭包。
     * 例如 alice 与 bob 共用设备A、bob 又与 carol 共用设备B，则三人同属一个团伙。
     * 单独命中多设备规则、未与任何账号共道的账号自成一个单人团伙。
     */
    static List<DeviceAnomalyVO.CheatCluster> buildClusters(
            List<SuspiciousAccount> suspiciousAccounts, List<SharedDevice> sharedDeviceList) {
        Map<String, SuspiciousAccount> accountsByUid = new LinkedHashMap<>();
        for (SuspiciousAccount account : suspiciousAccounts) {
            accountsByUid.put(account.getUid(), account);
        }
        // 邻接表：共用一台设备的账号两两相连
        Map<String, Set<String>> adjacency = new HashMap<>();
        for (SharedDevice device : sharedDeviceList) {
            List<String> uidsOnDevice = device.getUsers().stream()
                    .map(UserUsage::getUid).collect(Collectors.toList());
            for (String uid : uidsOnDevice) {
                Set<String> neighbors = adjacency.computeIfAbsent(uid, k -> new LinkedHashSet<>());
                for (String other : uidsOnDevice) {
                    if (!other.equals(uid)) {
                        neighbors.add(other);
                    }
                }
            }
        }
        // BFS 求连通分量
        List<List<String>> components = new ArrayList<>();
        Set<String> visited = new HashSet<>();
        for (String uid : accountsByUid.keySet()) {
            if (visited.contains(uid)) {
                continue;
            }
            List<String> component = new ArrayList<>();
            Deque<String> queue = new ArrayDeque<>();
            queue.add(uid);
            visited.add(uid);
            while (!queue.isEmpty()) {
                String current = queue.poll();
                component.add(current);
                for (String neighbor : adjacency.getOrDefault(current, Collections.emptySet())) {
                    if (visited.add(neighbor)) {
                        queue.add(neighbor);
                    }
                }
            }
            components.add(component);
        }
        // 每个团伙内部按用户名排序；团伙之间成员多的在前
        List<List<String>> componentsByNames = new ArrayList<>();
        for (List<String> component : components) {
            List<String> memberNames = component.stream()
                    .map(uid -> accountsByUid.get(uid).getUsername())
                    .sorted()
                    .collect(Collectors.toList());
            componentsByNames.add(memberNames);
        }
        List<Integer> order = new ArrayList<>();
        for (int i = 0; i < components.size(); i++) {
            order.add(i);
        }
        order.sort(Comparator
                .comparingInt((Integer i) -> components.get(i).size()).reversed()
                .thenComparing(i -> componentsByNames.get(i).get(0)));
        // 生成团伙与关联边
        List<DeviceAnomalyVO.CheatCluster> clusters = new ArrayList<>();
        Map<String, Integer> uidToClusterId = new HashMap<>();
        int clusterId = 1;
        for (int index : order) {
            List<String> component = components.get(index);
            DeviceAnomalyVO.CheatCluster cluster = new DeviceAnomalyVO.CheatCluster();
            cluster.setClusterId(clusterId);
            cluster.setUsernames(componentsByNames.get(index));
            List<DeviceAnomalyVO.ClusterLink> links = new ArrayList<>();
            for (SharedDevice device : sharedDeviceList) {
                List<String> membersOnDevice = device.getUsers().stream()
                        .map(UserUsage::getUid)
                        .filter(component::contains)
                        .map(uid -> accountsByUid.get(uid).getUsername())
                        .sorted()
                        .collect(Collectors.toList());
                if (membersOnDevice.size() >= 2) {
                    DeviceAnomalyVO.ClusterLink link = new DeviceAnomalyVO.ClusterLink();
                    link.setDeviceId(device.getDeviceId());
                    link.setUsernames(membersOnDevice);
                    links.add(link);
                }
            }
            cluster.setLinks(links);
            for (String uid : component) {
                uidToClusterId.put(uid, clusterId);
            }
            clusters.add(cluster);
            clusterId++;
        }
        for (SuspiciousAccount account : suspiciousAccounts) {
            Integer accountClusterId = uidToClusterId.get(account.getUid());
            account.setClusterId(accountClusterId);
            DeviceAnomalyVO.CheatCluster cluster = clusters.get(accountClusterId - 1);
            account.setRelatedUsernames(cluster.getUsernames().stream()
                    .filter(name -> !name.equals(account.getUsername()))
                    .collect(Collectors.toList()));
        }
        return clusters;
    }

    private static class DeviceRecord {
        private final String deviceId;
        private final Map<String, AccountRecord.DeviceStat> stats = new LinkedHashMap<>();
        private final Set<String> usernames = new LinkedHashSet<>();
        private final Set<String> uidOrder = new LinkedHashSet<>();
        private String latestUserAgent;
        private int submissionCount;
        private Date firstSubmitTime;
        private Date lastSubmitTime;

        private DeviceRecord(String deviceId) {
            this.deviceId = deviceId;
        }

        private void add(Judge judge, String username) {
            Date submitTime = judge.getSubmitTime();
            submissionCount++;
            if (firstSubmitTime == null || (submitTime != null && submitTime.before(firstSubmitTime))) {
                firstSubmitTime = submitTime;
            }
            if (lastSubmitTime == null || (submitTime != null && submitTime.after(lastSubmitTime))) {
                lastSubmitTime = submitTime;
            }
            if (judge.getUserAgent() != null && !judge.getUserAgent().trim().isEmpty()) {
                latestUserAgent = judge.getUserAgent().trim();
            }
            if (!usernames.contains(username)) {
                usernames.add(username);
                uidOrder.add(judge.getUid());
            }
            AccountRecord.DeviceStat stat = stats.computeIfAbsent(judge.getUid(),
                    k -> new AccountRecord.DeviceStat());
            stat.add(submitTime);
        }
    }

    private static class AccountRecord {
        private final String uid;
        private final String username;
        private final Map<String, DeviceStat> devices = new LinkedHashMap<>();
        private int submissionCount;
        private Date firstSubmitTime;
        private Date lastSubmitTime;

        private AccountRecord(String uid, String username) {
            this.uid = uid;
            this.username = username;
        }

        private void add(Judge judge) {
            Date submitTime = judge.getSubmitTime();
            submissionCount++;
            if (firstSubmitTime == null || (submitTime != null && submitTime.before(firstSubmitTime))) {
                firstSubmitTime = submitTime;
            }
            if (lastSubmitTime == null || (submitTime != null && submitTime.after(lastSubmitTime))) {
                lastSubmitTime = submitTime;
            }
            devices.computeIfAbsent(judge.getDeviceId(), k -> new DeviceStat()).add(submitTime);
        }

        private static class DeviceStat {
            private int submissionCount;
            private Date firstSubmitTime;
            private Date lastSubmitTime;

            private void add(Date submitTime) {
                submissionCount++;
                if (firstSubmitTime == null || (submitTime != null && submitTime.before(firstSubmitTime))) {
                    firstSubmitTime = submitTime;
                }
                if (lastSubmitTime == null || (submitTime != null && submitTime.after(lastSubmitTime))) {
                    lastSubmitTime = submitTime;
                }
            }
        }
    }
}
