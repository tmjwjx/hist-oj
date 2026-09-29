package top.hcode.hoj.service.problem;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.client.ClientHttpRequest;
import org.springframework.mock.http.client.MockClientHttpRequest;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.test.web.client.MockRestServiceServer;
import org.springframework.test.web.client.ResponseActions;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;
import top.hcode.hoj.dao.judge.JudgeServerEntityService;
import top.hcode.hoj.pojo.entity.judge.JudgeServer;
import top.hcode.hoj.utils.Constants;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;
import java.util.LinkedHashSet;
import java.util.Set;
import java.util.UUID;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;
import static org.junit.jupiter.api.Assumptions.assumeTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.client.match.MockRestRequestMatchers.method;
import static org.springframework.test.web.client.match.MockRestRequestMatchers.requestTo;
import static org.springframework.test.web.client.response.MockRestResponseCreators.withStatus;
import static org.springframework.test.web.client.response.MockRestResponseCreators.withSuccess;

class ProblemTestCaseSyncManagerTest {
    private static final String VERSION = "fixture-version";
    private static final String TOKEN = "fixture-judge-token";
    private final Path root = Paths.get(Constants.File.TESTCASE_BASE_FOLDER.getPath());
    private final Set<Path> markers = new LinkedHashSet<>();
    private Path problemDirectory;
    private Path archive;
    private Long pid;
    private ProblemTestCaseSyncManager manager;
    private JudgeServerEntityService servers;
    private ProblemTestCaseArchiveService archiveService;
    private MockRestServiceServer probeRequests;
    private MockRestServiceServer archiveRequests;
    private MockRestServiceServer metadataRequests;

    @BeforeEach
    void setUp() throws Exception {
        // Run in the JDK container with /hoj/testcase prepared; never create a host root directory.
        assumeTrue(Files.isDirectory(root) && Files.isWritable(root),
                "Run in a container with a writable /hoj/testcase directory");
        pid = (UUID.randomUUID().getMostSignificantBits() & Long.MAX_VALUE) | 1L;
        problemDirectory = Files.createDirectory(root.resolve("problem_" + pid));
        Files.write(problemDirectory.resolve("info"), "fixture-info".getBytes(StandardCharsets.UTF_8));
        manager = new ProblemTestCaseSyncManager();
        servers = mock(JudgeServerEntityService.class);
        archiveService = mock(ProblemTestCaseArchiveService.class);
        RestTemplate metadataTemplate = new RestTemplate();
        ReflectionTestUtils.setField(manager, "judgeServerEntityService", servers);
        ReflectionTestUtils.setField(manager, "archiveService", archiveService);
        ReflectionTestUtils.setField(manager, "restTemplate", metadataTemplate);
        ReflectionTestUtils.setField(manager, "judgeToken", TOKEN);
        probeRequests = MockRestServiceServer.bindTo(
                (RestTemplate) ReflectionTestUtils.getField(manager, "probeRestTemplate")).build();
        archiveRequests = MockRestServiceServer.bindTo(
                (RestTemplate) ReflectionTestUtils.getField(manager, "archiveRestTemplate")).build();
        metadataRequests = MockRestServiceServer.bindTo(metadataTemplate).build();
        when(archiveService.createArchive(pid)).thenAnswer(invocation -> {
            archive = Files.createTempFile(problemDirectory, "archive-", ".zip");
            Files.write(archive, "fixture-archive".getBytes(StandardCharsets.UTF_8));
            return archive.toFile();
        });
    }

    @AfterEach
    void tearDown() throws Exception {
        try {
            if (probeRequests != null) probeRequests.verify();
            if (archiveRequests != null) archiveRequests.verify();
            if (metadataRequests != null) metadataRequests.verify();
        } finally {
            for (Path marker : markers) Files.deleteIfExists(marker);
            if (problemDirectory != null) {
                try (Stream<Path> paths = Files.walk(problemDirectory)) {
                    for (Path path : paths.sorted(Comparator.reverseOrder()).collect(Collectors.toList())) {
                        Files.deleteIfExists(path);
                    }
                }
            }
        }
    }

    @Test
    void sharedNodesNeedNeitherArchiveNorUpload() throws Exception {
        localServers("shared-a:8080", "shared-b:8080");
        probe("shared-a:8080").andRespond(result(true));
        probe("shared-b:8080").andRespond(result(true));

        assertEquals(2, manager.sync(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertNull(archive);
        assertMarkersCleaned();
    }

    @Test
    void mixedStorageArchivesOnceAndUploadsOnlyToSeparatedNode() throws Exception {
        localServers("shared:8080", "separated:8080");
        probe("shared:8080").andRespond(result(true));
        probe("separated:8080").andRespond(result(false));
        expectArchive("separated:8080");

        assertEquals(2, manager.sync(pid, VERSION));
        verify(archiveService, times(1)).createArchive(pid);
        assertFalse(Files.exists(archive));
        assertMarkersCleaned();
    }

    @Test
    void separatedNodesReuseOneArchive() throws Exception {
        localServers("separated-a:8080", "separated-b:8080");
        probe("separated-a:8080").andRespond(result(false));
        probe("separated-b:8080").andRespond(result(false));
        expectArchive("separated-a:8080");
        expectArchive("separated-b:8080");

        assertEquals(2, manager.sync(pid, VERSION));
        verify(archiveService, times(1)).createArchive(pid);
        assertFalse(Files.exists(archive));
        assertMarkersCleaned();
    }

    @Test
    void probeBusinessFailureDoesNotFallBackToUploading() throws Exception {
        localServers("failed:8080");
        probe("failed:8080").andRespond(withSuccess(
                "{\"status\":500,\"msg\":\"invalid version\",\"data\":false}", MediaType.APPLICATION_JSON));

        assertThrows(IllegalStateException.class, () -> manager.sync(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertMarkersCleaned();
    }

    @Test
    void unauthorizedProbeDoesNotFallBackToUploading() throws Exception {
        localServers("unauthorized:8080");
        probe("unauthorized:8080").andRespond(withStatus(HttpStatus.UNAUTHORIZED));

        assertThrows(HttpClientErrorException.class, () -> manager.sync(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertMarkersCleaned();
    }

    @Test
    void transportFailureDoesNotFallBackToUploading() throws Exception {
        localServers("unreachable:8080");
        probe("unreachable:8080").andRespond(request -> { throw new IOException("fixture disconnect"); });

        assertThrows(ResourceAccessException.class, () -> manager.sync(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertMarkersCleaned();
    }

    @Test
    void malformedProbeDataDoesNotFallBackToUploading() throws Exception {
        localServers("invalid:8080");
        probe("invalid:8080").andRespond(withSuccess(
                "{\"status\":200,\"data\":\"true\"}", MediaType.APPLICATION_JSON));

        assertThrows(IllegalStateException.class, () -> manager.sync(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertMarkersCleaned();
    }

    @Test
    void oldServerBadRequestFallsBackToOneArchiveUpload() throws Exception {
        localServers("old:8080");
        probe("old:8080").andRespond(withStatus(HttpStatus.BAD_REQUEST));
        expectArchive("old:8080");

        assertEquals(1, manager.sync(pid, VERSION));
        verify(archiveService, times(1)).createArchive(pid);
        assertFalse(Files.exists(archive));
        assertMarkersCleaned();
    }

    @Test
    void failedUploadStillRemovesMarkerAndArchive() throws Exception {
        localServers("separated:8080");
        probe("separated:8080").andRespond(result(false));
        archiveRequests.expect(requestTo("http://separated:8080/sync-testcase"))
                .andRespond(withSuccess("{\"status\":500,\"msg\":\"fixture failure\"}", MediaType.APPLICATION_JSON));

        assertThrows(IllegalStateException.class, () -> manager.sync(pid, VERSION));
        assertFalse(Files.exists(archive));
        assertMarkersCleaned();
    }

    @Test
    void sharedInfoNeedsNoMetadataUpload() throws Exception {
        localServers("shared:8080");
        probe("shared:8080").andRespond(result(true));

        assertEquals(1, manager.syncInfo(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertMarkersCleaned();
    }

    @Test
    void separatedInfoUploadsMetadataWithoutArchive() throws Exception {
        localServers("separated:8080");
        probe("separated:8080").andRespond(result(false));
        metadataRequests.expect(requestTo("http://separated:8080/sync-testcase-info"))
                .andExpect(method(HttpMethod.POST))
                .andExpect(request -> {
                    String body = body(request);
                    assertTrue(body.contains("name=\"info\""));
                    assertTrue(body.contains("fixture-info"));
                    assertFalse(body.contains("name=\"archive\""));
                    assertCommonParts(body);
                }).andRespond(result(true));

        assertEquals(1, manager.syncInfo(pid, VERSION));
        verify(archiveService, never()).createArchive(anyLong());
        assertMarkersCleaned();
    }

    private void localServers(String... urls) {
        when(servers.list(any())).thenReturn(new ArrayList<>(Arrays.stream(urls)
                .map(url -> new JudgeServer().setUrl(url).setIsRemote(false))
                .collect(Collectors.toList())));
    }

    private ResponseActions probe(String url) {
        return probeRequests.expect(requestTo("http://" + url + "/sync-testcase"))
                .andExpect(method(HttpMethod.POST))
                .andExpect(request -> {
                    String body = body(request);
                    assertCommonParts(body);
                    assertFalse(body.contains("name=\"archive\""));
                    assertFalse(body.contains("name=\"info\""));
                    Matcher token = Pattern.compile("name=\"sharedToken\"[\\s\\S]*?\\r\\n\\r\\n([0-9a-f-]{36})").matcher(body);
                    assertTrue(token.find(), "sharedToken must be included in the probe");
                    Path marker = root.resolve(".hoj-sync-" + token.group(1));
                    markers.add(marker);
                    assertTrue(Files.isRegularFile(marker));
                    assertEquals(token.group(1), new String(Files.readAllBytes(marker), StandardCharsets.UTF_8));
                });
    }

    private void expectArchive(String url) {
        archiveRequests.expect(requestTo("http://" + url + "/sync-testcase"))
                .andExpect(method(HttpMethod.POST))
                .andExpect(request -> {
                    String body = body(request);
                    assertCommonParts(body);
                    assertTrue(body.contains("name=\"archive\""));
                    assertTrue(body.contains("fixture-archive"));
                    assertTrue(Files.isRegularFile(archive));
                }).andRespond(result(true));
    }

    private void assertCommonParts(String body) {
        for (String[] part : new String[][]{{"token", TOKEN}, {"pid", pid.toString()}, {"version", VERSION}}) {
            assertTrue(Pattern.compile("name=\"" + part[0] + "\"[\\s\\S]*?\\r\\n\\r\\n"
                    + Pattern.quote(part[1]) + "\\r\\n").matcher(body).find(), "Missing form part " + part[0]);
        }
    }

    private void assertMarkersCleaned() {
        assertEquals(1, markers.size(), "One marker must be reused for every node in one sync");
        for (Path marker : markers) assertFalse(Files.exists(marker));
    }

    private static String body(ClientHttpRequest request) {
        return ((MockClientHttpRequest) request).getBodyAsString(StandardCharsets.UTF_8);
    }

    private static org.springframework.test.web.client.ResponseCreator result(boolean shared) {
        return withSuccess("{\"status\":200,\"data\":" + shared + "}", MediaType.APPLICATION_JSON);
    }
}
