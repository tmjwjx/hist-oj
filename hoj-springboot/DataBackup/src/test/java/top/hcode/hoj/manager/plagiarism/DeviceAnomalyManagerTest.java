package top.hcode.hoj.manager.plagiarism;

import org.junit.jupiter.api.Test;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO.SharedDevice;
import top.hcode.hoj.pojo.vo.plagiarism.DeviceAnomalyVO.SuspiciousAccount;

import java.util.Arrays;
import java.util.Date;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

class DeviceAnomalyManagerTest {

    private Judge submission(String uid, String username, String deviceId, Date submitTime) {
        return new Judge().setUid(uid).setUsername(username).setDeviceId(deviceId)
                .setUserAgent("UA-" + deviceId).setSubmitTime(submitTime).setCid(1L);
    }

    @Test
    void flagsAccountSubmittingFromMultipleDevices() {
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                submission("u1", "alice", "dev-a", new Date(1000)),
                submission("u1", "alice", "dev-b", new Date(2000))));

        assertEquals(1, vo.getSuspiciousAccountCount());
        SuspiciousAccount alice = vo.getSuspiciousAccounts().get(0);
        assertEquals("u1", alice.getUid());
        assertTrue(alice.getMultiDevice());
        assertFalse(alice.getDeviceSharing());
        assertEquals(2, alice.getDeviceCount());
        assertEquals(2, alice.getDevices().size());
        assertEquals(0, vo.getSharedDevices().size());
    }

    @Test
    void flagsDeviceSharedByMultipleAccounts() {
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                submission("u1", "alice", "dev-a", new Date(1000)),
                submission("u2", "bob", "dev-a", new Date(2000)),
                submission("u2", "bob", "dev-c", new Date(3000))));

        // 设备维度：dev-a 被两个账号使用
        assertEquals(1, vo.getSharedDevices().size());
        SharedDevice shared = vo.getSharedDevices().get(0);
        assertEquals("dev-a", shared.getDeviceId());
        assertEquals(2, shared.getUserCount());
        // 两人在该设备上提交数相同，按提交顺序排列
        assertEquals("alice", shared.getUsers().get(0).getUsername());
        assertEquals("bob", shared.getUsers().get(1).getUsername());

        // 账号维度：两人都命中共用设备，bob 额外命中多设备
        assertEquals(2, vo.getSuspiciousAccountCount());
        SuspiciousAccount bob = vo.getSuspiciousAccounts().stream()
                .filter(a -> "u2".equals(a.getUid())).findFirst().get();
        assertTrue(bob.getMultiDevice());
        assertTrue(bob.getDeviceSharing());
        SuspiciousAccount alice = vo.getSuspiciousAccounts().stream()
                .filter(a -> "u1".equals(a.getUid())).findFirst().get();
        assertFalse(alice.getMultiDevice());
        assertTrue(alice.getDeviceSharing());
        assertEquals("bob", alice.getDevices().get(0).getOtherUsernames().get(0));
    }

    @Test
    void unknownDeviceSubmissionsStayOutOfAggregation() {
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                submission("u1", "alice", null, new Date(1000)),
                submission("u2", "bob", null, new Date(2000)),
                submission("u3", "carol", "dev-a", new Date(3000))));

        assertEquals(3L, vo.getTotalSubmissions());
        assertEquals(2L, vo.getUnknownDeviceSubmissions());
        assertEquals(0, vo.getSuspiciousAccountCount());
        assertEquals(0, vo.getSharedDevices().size());
    }

    @Test
    void normalSingleDeviceSingleAccountIsNotFlagged() {
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                submission("u1", "alice", "dev-a", new Date(1000)),
                submission("u1", "alice", "dev-a", new Date(2000)),
                submission("u2", "bob", "dev-b", new Date(1500))));

        assertEquals(0, vo.getSuspiciousAccountCount());
        assertEquals(0, vo.getSharedDevices().size());
    }

    @Test
    void missingUserAgentStaysNull() {
        Judge silent = submission("u1", "alice", "dev-a", new Date(1000));
        silent.setUserAgent(null);
        // 让账号命中"多设备"规则进入嫌疑清单，再校验UA为空时不被填充
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                silent,
                submission("u1", "alice", "dev-b", new Date(2000))));
        assertEquals(1, vo.getSuspiciousAccountCount());
        assertNull(vo.getSuspiciousAccounts().get(0).getDevices().get(0).getUserAgent());
        assertEquals("UA-dev-b", vo.getSuspiciousAccounts().get(0).getDevices().get(1).getUserAgent());
    }

    @Test
    void clustersExpandTransitivelyThroughSharedDevices() {
        // alice 与 bob 共用 dev-a，bob 又与 carol 共用 dev-c：三人应串成一个团伙
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                submission("u1", "alice", "dev-a", new Date(1000)),
                submission("u2", "bob", "dev-a", new Date(2000)),
                submission("u2", "bob", "dev-c", new Date(3000)),
                submission("u3", "carol", "dev-c", new Date(4000))));

        assertEquals(3, vo.getSuspiciousAccountCount());
        assertEquals(1, vo.getClusters().size());
        DeviceAnomalyVO.CheatCluster cluster = vo.getClusters().get(0);
        assertEquals(1, cluster.getClusterId());
        assertEquals(Arrays.asList("alice", "bob", "carol"), cluster.getUsernames());
        // 两条关联边：dev-a 上 alice/bob，dev-c 上 bob/carol
        assertEquals(2, cluster.getLinks().size());
        assertEquals("dev-a", cluster.getLinks().get(0).getDeviceId());
        assertEquals(Arrays.asList("alice", "bob"), cluster.getLinks().get(0).getUsernames());
        assertEquals("dev-c", cluster.getLinks().get(1).getDeviceId());
        assertEquals(Arrays.asList("bob", "carol"), cluster.getLinks().get(1).getUsernames());
        // 每个账号都能看到团伙内其它账号（传递闭包，不只一层）
        for (SuspiciousAccount account : vo.getSuspiciousAccounts()) {
            assertEquals(Integer.valueOf(1), account.getClusterId());
            assertEquals(Arrays.asList("alice", "bob", "carol").stream()
                    .filter(name -> !name.equals(account.getUsername()))
                    .collect(java.util.stream.Collectors.toList()), account.getRelatedUsernames());
        }
    }

    @Test
    void soloMultiDeviceAccountFormsSingleMemberCluster() {
        DeviceAnomalyVO vo = DeviceAnomalyManager.aggregate(1L, Arrays.asList(
                submission("u1", "alice", "dev-a", new Date(1000)),
                submission("u1", "alice", "dev-b", new Date(2000))));
        assertEquals(1, vo.getClusters().size());
        DeviceAnomalyVO.CheatCluster cluster = vo.getClusters().get(0);
        assertEquals(Arrays.asList("alice"), cluster.getUsernames());
        assertTrue(cluster.getLinks().isEmpty());
        assertEquals(Integer.valueOf(1), vo.getSuspiciousAccounts().get(0).getClusterId());
        assertTrue(vo.getSuspiciousAccounts().get(0).getRelatedUsernames().isEmpty());
    }
}
