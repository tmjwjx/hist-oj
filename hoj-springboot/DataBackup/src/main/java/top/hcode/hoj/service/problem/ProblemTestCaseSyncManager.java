package top.hcode.hoj.service.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.cloud.context.config.annotation.RefreshScope;
import org.springframework.core.io.FileSystemResource;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.client.HttpClientErrorException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.dao.judge.JudgeServerEntityService;
import top.hcode.hoj.pojo.entity.judge.JudgeServer;
import top.hcode.hoj.utils.Constants;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.util.List;
import java.util.UUID;

@Component
@RefreshScope
public class ProblemTestCaseSyncManager {

    @Autowired
    private JudgeServerEntityService judgeServerEntityService;

    @Autowired
    private ProblemTestCaseArchiveService archiveService;

    @Autowired
    private RestTemplate restTemplate;

    private final RestTemplate archiveRestTemplate = createArchiveRestTemplate();
    private final RestTemplate probeRestTemplate = createProbeRestTemplate();

    @Value("${hoj.judge.token:no_judge_token}")
    private String judgeToken;

    public int sync(Long pid, String version) throws Exception {
        List<JudgeServer> servers = localServers();
        if (servers.isEmpty()) {
            throw new IllegalStateException("没有可用的本地判题服务器");
        }

        Path marker = createSharedMarker(pid, version);
        File archive = null;
        try {
            for (JudgeServer server : servers) {
                if (usesSharedTestCases(server, pid, version, marker)) continue;
                if (archive == null) archive = archiveService.createArchive(pid);
                syncToServer(server, pid, version, archive);
            }
            return servers.size();
        } finally {
            Files.deleteIfExists(marker);
            if (archive != null && !archive.delete()) {
                archive.deleteOnExit();
            }
        }
    }

    public int syncInfo(Long pid, String version) throws Exception {
        List<JudgeServer> servers = localServers();
        if (servers.isEmpty()) throw new IllegalStateException("没有可用的本地判题服务器");
        File info = new File(Constants.File.TESTCASE_BASE_FOLDER.getPath(),
                "problem_" + pid + File.separator + "info");
        if (!info.isFile()) throw new IllegalStateException("测试数据 info 文件不存在");
        Path marker = createSharedMarker(pid, version);
        try {
            for (JudgeServer server : servers) {
                if (!usesSharedTestCases(server, pid, version, marker)) {
                    syncInfoToServer(server, pid, version, info);
                }
            }
        } finally {
            Files.deleteIfExists(marker);
        }
        return servers.size();
    }

    private Path createSharedMarker(Long pid, String version) throws Exception {
        if (pid == null || pid <= 0 || version == null || version.trim().isEmpty()) {
            throw new IllegalArgumentException("题目 ID 和测试数据版本无效");
        }
        Path root = Paths.get(Constants.File.TESTCASE_BASE_FOLDER.getPath());
        if (!Files.isRegularFile(root.resolve("problem_" + pid).resolve("info"))) {
            throw new IllegalStateException("测试数据 info 文件不存在");
        }
        String token = UUID.randomUUID().toString();
        return Files.write(root.resolve(".hoj-sync-" + token), token.getBytes(StandardCharsets.UTF_8),
                StandardOpenOption.CREATE_NEW, StandardOpenOption.WRITE);
    }

    private boolean usesSharedTestCases(JudgeServer server, Long pid, String version, Path marker) {
        String url = server.getUrl();
        if (url == null || url.trim().isEmpty()) url = server.getIp() + ":" + server.getPort();
        MultiValueMap<String, Object> body = new LinkedMultiValueMap<>();
        body.add("token", judgeToken);
        body.add("pid", pid.toString());
        body.add("version", version);
        body.add("sharedToken", marker.getFileName().toString().substring(".hoj-sync-".length()));
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.MULTIPART_FORM_DATA);
        try {
            CommonResult result = probeRestTemplate.postForEntity("http://" + url + "/sync-testcase",
                    new HttpEntity<>(body, headers), CommonResult.class).getBody();
            if (result == null || result.getStatus() == null || result.getStatus() != 200
                    || !(result.getData() instanceof Boolean)) {
                throw new IllegalStateException("确认 " + url + " 测试数据失败："
                        + (result == null ? "判题服务器没有返回结果" : result.getMsg()));
            }
            return Boolean.TRUE.equals(result.getData());
        } catch (HttpClientErrorException e) {
            // 旧判题服务要求 archive，尚不支持只读共享目录确认。
            int status = e.getRawStatusCode();
            if (status == 400 || status == 404 || status == 405) return false;
            throw e;
        }
    }

    public int delete(Long pid) throws Exception {
        List<JudgeServer> servers = judgeServerEntityService.list(
                new QueryWrapper<JudgeServer>().eq("status", 0));
        servers.removeIf(server -> Boolean.TRUE.equals(server.getIsRemote()));
        if (servers.isEmpty()) return 0;
        for (JudgeServer server : servers) {
            String url = server.getUrl();
            if (url == null || url.trim().isEmpty()) url = server.getIp() + ":" + server.getPort();
            ResponseEntity<CommonResult> response = restTemplate.postForEntity(
                    "http://" + url + "/delete-testcase?token={token}&pid={pid}",
                    null, CommonResult.class, judgeToken, pid);
            CommonResult result = response.getBody();
            if (result == null || result.getStatus() == null || result.getStatus() != 200) {
                throw new IllegalStateException("清理 " + url + " 的测试数据失败");
            }
        }
        return servers.size();
    }

    private void syncToServer(JudgeServer server, Long pid, String version, File archive) {
        String url = server.getUrl();
        if (url == null || url.trim().isEmpty()) {
            url = server.getIp() + ":" + server.getPort();
        }
        MultiValueMap<String, Object> body = new LinkedMultiValueMap<>();
        body.add("token", judgeToken);
        body.add("pid", pid.toString());
        body.add("version", version);
        body.add("archive", new FileSystemResource(archive));

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.MULTIPART_FORM_DATA);
        ResponseEntity<CommonResult> response = archiveRestTemplate.postForEntity(
                "http://" + url + "/sync-testcase", new HttpEntity<>(body, headers), CommonResult.class);
        CommonResult result = response.getBody();
        if (result == null || result.getStatus() == null || result.getStatus() != 200) {
            String message = result == null ? "判题服务器没有返回结果" : result.getMsg();
            throw new IllegalStateException("同步到 " + url + " 失败：" + message);
        }
    }

    private void syncInfoToServer(JudgeServer server, Long pid, String version, File info) {
        String url = server.getUrl();
        if (url == null || url.trim().isEmpty()) url = server.getIp() + ":" + server.getPort();
        MultiValueMap<String, Object> body = new LinkedMultiValueMap<>();
        body.add("token", judgeToken);
        body.add("pid", pid.toString());
        body.add("version", version);
        body.add("info", new FileSystemResource(info));
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.MULTIPART_FORM_DATA);
        ResponseEntity<CommonResult> response = restTemplate.postForEntity(
                "http://" + url + "/sync-testcase-info", new HttpEntity<>(body, headers), CommonResult.class);
        CommonResult result = response.getBody();
        if (result == null || result.getStatus() == null || result.getStatus() != 200) {
            throw new IllegalStateException(result == null ? "判题服务器没有返回结果" : result.getMsg());
        }
    }

    private List<JudgeServer> localServers() {
        List<JudgeServer> servers = judgeServerEntityService.list(
                new QueryWrapper<JudgeServer>().eq("status", 0));
        servers.removeIf(server -> Boolean.TRUE.equals(server.getIsRemote()));
        return servers;
    }

    private static RestTemplate createArchiveRestTemplate() {
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setBufferRequestBody(false);
        factory.setConnectTimeout(50000);
        factory.setReadTimeout(1800000);
        return new RestTemplate(factory);
    }

    private static RestTemplate createProbeRestTemplate() {
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(5000);
        factory.setReadTimeout(15000);
        return new RestTemplate(factory);
    }
}
