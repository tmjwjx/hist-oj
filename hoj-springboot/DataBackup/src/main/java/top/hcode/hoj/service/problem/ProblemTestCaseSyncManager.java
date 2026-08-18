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
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.dao.judge.JudgeServerEntityService;
import top.hcode.hoj.pojo.entity.judge.JudgeServer;
import top.hcode.hoj.utils.Constants;

import java.io.File;
import java.util.List;

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

    @Value("${hoj.judge.token:no_judge_token}")
    private String judgeToken;

    public int sync(Long pid, String version) throws Exception {
        List<JudgeServer> servers = judgeServerEntityService.list(
                new QueryWrapper<JudgeServer>().eq("status", 0));
        servers.removeIf(server -> Boolean.TRUE.equals(server.getIsRemote()));
        if (servers.isEmpty()) {
            throw new IllegalStateException("没有可用的本地判题服务器");
        }

        File archive = archiveService.createArchive(pid);
        try {
            for (JudgeServer server : servers) {
                syncToServer(server, pid, version, archive);
            }
            return servers.size();
        } finally {
            if (!archive.delete()) {
                archive.deleteOnExit();
            }
        }
    }

    public int syncInfo(Long pid, String version) {
        List<JudgeServer> servers = localServers();
        if (servers.isEmpty()) throw new IllegalStateException("没有可用的本地判题服务器");
        File info = new File(Constants.File.TESTCASE_BASE_FOLDER.getPath(),
                "problem_" + pid + File.separator + "info");
        if (!info.isFile()) throw new IllegalStateException("测试数据 info 文件不存在");
        for (JudgeServer server : servers) syncInfoToServer(server, pid, version, info);
        return servers.size();
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
}
