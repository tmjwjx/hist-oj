package top.hcode.hoj.service;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.springframework.mock.web.MockMultipartFile;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import top.hcode.hoj.controller.JudgeController;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.UUID;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

import static org.junit.jupiter.api.Assertions.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.multipart;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

class ProblemTestCaseSyncServiceTest {
    @TempDir
    Path root;
    private ProblemTestCaseSyncService service;
    private String token;

    @BeforeEach
    void setup() {
        service = new ProblemTestCaseSyncService(root);
        token = UUID.randomUUID().toString();
    }

    private JSONObject info(String version) {
        JSONObject entry = new JSONObject().set("inputName", "1.in").set("outputName", "1.out")
                .set("caseId", 1).set("score", 100).set("outputMd5", "b026324c6904b2a9cb4b88d6d61c81d1")
                .set("outputSize", 2);
        return new JSONObject().set("version", version).set("mode", "default")
                .set("testCasesSize", 1).set("testCases", new JSONArray().put(entry));
    }

    private byte[] bytes(JSONObject info) {
        return JSONUtil.toJsonStr(info).getBytes(StandardCharsets.UTF_8);
    }

    private Path installed(String version) throws IOException {
        Path directory = root.resolve("problem_1");
        Files.createDirectories(directory);
        Files.write(directory.resolve("info"), bytes(info(version)));
        Files.write(directory.resolve("1.in"), "1\n".getBytes(StandardCharsets.UTF_8));
        Files.write(directory.resolve("1.out"), "1\n".getBytes(StandardCharsets.UTF_8));
        return directory;
    }

    private Path marker() throws IOException {
        return Files.write(root.resolve(".hoj-sync-" + token), token.getBytes(StandardCharsets.UTF_8));
    }

    private MockMultipartFile metadata(JSONObject info) {
        return new MockMultipartFile("info", "info", "application/json", bytes(info));
    }

    private MockMultipartFile archive(JSONObject info, boolean data, String extraPath) throws IOException {
        Map<String, byte[]> entries = new LinkedHashMap<>();
        entries.put("info", bytes(info));
        if (data) {
            entries.put("1.in", "2\n".getBytes(StandardCharsets.UTF_8));
            entries.put("1.out", "1\n".getBytes(StandardCharsets.UTF_8));
        }
        if (extraPath != null) entries.put(extraPath, new byte[]{1});
        ByteArrayOutputStream buffer = new ByteArrayOutputStream();
        try (ZipOutputStream zip = new ZipOutputStream(buffer)) {
            for (Map.Entry<String, byte[]> entry : entries.entrySet()) {
                zip.putNextEntry(new ZipEntry(entry.getKey()));
                zip.write(entry.getValue());
                zip.closeEntry();
            }
        }
        return new MockMultipartFile("archive", "cases.zip", "application/zip", buffer.toByteArray());
    }

    @Test
    void freshSharedMarkerConfirmsCompleteVersionWithoutWriting() throws Exception {
        Path directory = installed("v1");
        byte[] original = Files.readAllBytes(directory.resolve("info"));
        marker();
        assertTrue(service.confirmShared(1L, "v1", token));
        assertArrayEquals(original, Files.readAllBytes(directory.resolve("info")));
        assertEquals(token, new String(Files.readAllBytes(root.resolve(".hoj-sync-" + token)), StandardCharsets.UTF_8));
    }

    @Test
    void missingMarkerMeansSeparateStorageEvenWithMatchingVersion() throws Exception {
        assertFalse(service.confirmShared(1L, "v1", token));
        installed("v1");
        assertFalse(service.confirmShared(1L, "v1", token));
    }

    @Test
    void invalidRequestCannotBeMistakenForSeparateStorage() {
        for (String invalid : new String[]{null, "", "1-1-1-1-1", "../escape", token + "/info", token + " "}) {
            assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", invalid));
        }
        assertThrows(IOException.class, () -> service.confirmShared(0L, "v1", token));
        assertThrows(IOException.class, () -> service.confirmShared(null, "v1", token));
        assertThrows(IOException.class, () -> service.confirmShared(1L, " ", token));
        assertThrows(IOException.class, () -> service.install(-1L, "v1", archive(info("v1"), true, null)));
        assertThrows(IOException.class, () -> service.installInfo(1L, "", metadata(info("v1"))));
    }

    @Test
    void mismatchedMarkerIsAnError() throws Exception {
        Path marker = marker();
        Files.write(marker, UUID.randomUUID().toString().getBytes(StandardCharsets.UTF_8));
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
    }

    @Test
    void directoryOrSymlinkMarkerIsAnError() throws Exception {
        Path marker = root.resolve(".hoj-sync-" + token);
        Files.createDirectory(marker);
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
        Files.delete(marker);
        Path value = Files.write(root.resolve("marker-value"), token.getBytes(StandardCharsets.UTF_8));
        Files.createSymbolicLink(marker, value);
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
    }

    @Test
    void sharedVersionMismatchAndMissingDataMustFail() throws Exception {
        marker();
        Path directory = installed("v1");
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v2", token));
        Files.delete(directory.resolve("1.in"));
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
    }

    @Test
    void sharedProbeRejectsUnsafeManifestPathsAndSymbolicLinks() throws Exception {
        marker();
        Path directory = installed("v1");
        for (String name : new String[]{"../outside.in", "/tmp/outside.in", "sub/../../outside.in", "sub\\file.in"}) {
            JSONObject malformed = info("v1");
            malformed.getJSONArray("testCases").getJSONObject(0).set("inputName", name);
            Files.write(directory.resolve("info"), bytes(malformed));
            assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
        }
        Files.write(directory.resolve("info"), bytes(info("v1")));
        Files.delete(directory.resolve("1.in"));
        Files.createSymbolicLink(directory.resolve("1.in"), directory.resolve("1.out"));
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
    }

    @Test
    void sharedProbeRejectsSymlinkInfoAndDirectory() throws Exception {
        marker();
        Path directory = installed("v1");
        Files.move(directory.resolve("info"), root.resolve("real-info"));
        Files.createSymbolicLink(directory.resolve("info"), root.resolve("real-info"));
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
        Files.move(directory, root.resolve("real-problem"));
        Files.createSymbolicLink(directory, root.resolve("real-problem"));
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
    }

    @Test
    void sharedManifestMustHaveMatchingNonzeroCount() throws Exception {
        marker();
        Path directory = installed("v1");
        for (Object count : new Object[]{0, 2, "1", 1.5}) {
            JSONObject malformed = info("v1").set("testCasesSize", count);
            Files.write(directory.resolve("info"), bytes(malformed));
            assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
        }
        Files.write(directory.resolve("info"), bytes(info("v1").set("testCasesSize", 0).set("testCases", new JSONArray())));
        assertThrows(IOException.class, () -> service.confirmShared(1L, "v1", token));
    }

    @Test
    void fullInstallReplacesOnlyAfterCompleteManifestValidation() throws Exception {
        Path directory = installed("old");
        Files.write(directory.resolve("obsolete"), new byte[]{1});
        service.install(1L, "new", archive(info("new"), true, null));
        assertEquals("new", JSONUtil.parseObj(new String(Files.readAllBytes(directory.resolve("info")), StandardCharsets.UTF_8)).getStr("version"));
        assertFalse(Files.exists(directory.resolve("obsolete")));
        assertEquals("2\n", new String(Files.readAllBytes(directory.resolve("1.in")), StandardCharsets.UTF_8));
    }

    @Test
    void incompleteOrWrongVersionArchiveNeverReplacesPublishedData() throws Exception {
        Path directory = installed("old");
        byte[] original = Files.readAllBytes(directory.resolve("info"));
        assertThrows(IOException.class, () -> service.install(1L, "new", archive(info("new"), false, null)));
        assertThrows(IOException.class, () -> service.install(1L, "new", archive(info("other"), true, null)));
        assertThrows(IOException.class, () -> service.install(1L, "new", archive(info("new").set("testCasesSize", 2), true, null)));
        assertArrayEquals(original, Files.readAllBytes(directory.resolve("info")));
        assertEquals("1\n", new String(Files.readAllBytes(directory.resolve("1.in")), StandardCharsets.UTF_8));
        try (java.util.stream.Stream<Path> paths = Files.list(root)) {
            assertFalse(paths.anyMatch(p -> p.getFileName().toString().contains("_sync_")));
        }
    }

    @Test
    void zipTraversalNeverWritesOutsideOrReplacesPublishedData() throws Exception {
        Path directory = installed("old");
        byte[] original = Files.readAllBytes(directory.resolve("info"));
        assertThrows(IOException.class, () -> service.install(1L, "new", archive(info("new"), true, "../escaped")));
        assertFalse(Files.exists(root.resolve("escaped")));
        assertArrayEquals(original, Files.readAllBytes(directory.resolve("info")));
    }

    @Test
    void metadataOnlyUpdatesScoresGroupingAndModeWithoutChangingTestFiles() throws Exception {
        Path directory = installed("old");
        byte[] input = Files.readAllBytes(directory.resolve("1.in"));
        JSONObject updated = info("new").set("judgeCaseMode", "subtask_lowest");
        updated.getJSONArray("testCases").getJSONObject(0).set("score", 50).set("groupNum", 2);
        service.installInfo(1L, "new", metadata(updated));
        assertArrayEquals(bytes(updated), Files.readAllBytes(directory.resolve("info")));
        assertArrayEquals(input, Files.readAllBytes(directory.resolve("1.in")));
    }

    @Test
    void metadataCannotBlessMissingDataOrMissingOldInfo() throws Exception {
        Path directory = installed("old");
        byte[] original = Files.readAllBytes(directory.resolve("info"));
        Files.delete(directory.resolve("1.out"));
        assertThrows(IOException.class, () -> service.installInfo(1L, "new", metadata(info("new"))));
        assertArrayEquals(original, Files.readAllBytes(directory.resolve("info")));
        Files.write(directory.resolve("1.out"), new byte[]{1});
        Files.delete(directory.resolve("info"));
        assertThrows(IOException.class, () -> service.installInfo(1L, "new", metadata(info("new"))));
        assertFalse(Files.exists(directory.resolve("info")));
    }

    @Test
    void metadataChangedFileNamesOrFingerprintsRequireFullSync() throws Exception {
        Path directory = installed("old");
        byte[] original = Files.readAllBytes(directory.resolve("info"));
        Files.write(directory.resolve("2.in"), new byte[]{2});
        JSONObject renamed = info("new");
        renamed.getJSONArray("testCases").getJSONObject(0).set("inputName", "2.in");
        assertThrows(IOException.class, () -> service.installInfo(1L, "new", metadata(renamed)));
        JSONObject changed = info("new");
        changed.getJSONArray("testCases").getJSONObject(0).set("outputMd5", "different");
        assertThrows(IOException.class, () -> service.installInfo(1L, "new", metadata(changed)));
        assertArrayEquals(original, Files.readAllBytes(directory.resolve("info")));
    }

    @Test
    void truncatedFilesCannotPassSharedProbeOrMetadataUpdate() throws Exception {
        Path directory = installed("old");
        marker();
        byte[] original = Files.readAllBytes(directory.resolve("info"));
        Files.write(directory.resolve("1.out"), new byte[]{1});
        assertThrows(IOException.class, () -> service.confirmShared(1L, "old", token));
        assertThrows(IOException.class, () -> service.installInfo(1L, "new", metadata(info("new"))));
        assertArrayEquals(original, Files.readAllBytes(directory.resolve("info")));
    }

    @Test
    void declaredSizesMustBeNonnegativeIntegersAndMatchActualFiles() throws Exception {
        installed("old");
        for (Object invalid : new Object[]{-1, 1.5, "2", 3}) {
            JSONObject manifest = info("new");
            manifest.getJSONArray("testCases").getJSONObject(0).set("outputSize", invalid);
            assertThrows(IOException.class, () -> service.install(1L, "new", archive(manifest, true, null)));
        }
        JSONObject manifest = info("new");
        manifest.getJSONArray("testCases").getJSONObject(0).set("inputSize", 3);
        assertThrows(IOException.class, () -> service.install(1L, "new", archive(manifest, true, null)));
    }

    @Test
    void controllerSupportsOptionalArchiveAndNeverReportsProbeErrorsAsSuccess() throws Exception {
        JudgeController controller = new JudgeController();
        ReflectionTestUtils.setField(controller, "judgeToken", "test-secret");
        ReflectionTestUtils.setField(controller, "problemTestCaseSyncService", service);
        MockMvc http = MockMvcBuilders.standaloneSetup(controller).build();
        http.perform(post("/sync-testcase").param("token", "test-secret").param("pid", "1")
                .param("version", "v1").param("sharedToken", token))
                .andExpect(status().isOk()).andExpect(jsonPath("$.status").value(200)).andExpect(jsonPath("$.data").value(false));
        installed("v1");
        marker();
        http.perform(post("/sync-testcase").param("token", "test-secret").param("pid", "1")
                .param("version", "v1").param("sharedToken", token))
                .andExpect(jsonPath("$.status").value(200)).andExpect(jsonPath("$.data").value(true));
        http.perform(post("/sync-testcase").param("token", "test-secret").param("pid", "1")
                .param("version", "wrong").param("sharedToken", token))
                .andExpect(jsonPath("$.status").value(400)).andExpect(jsonPath("$.data").doesNotExist());
        http.perform(post("/sync-testcase").param("token", "wrong").param("pid", "1")
                .param("version", "v1").param("sharedToken", token)).andExpect(jsonPath("$.status").value(401));
        http.perform(multipart("/sync-testcase").file(archive(info("new"), true, null))
                .param("token", "test-secret").param("pid", "1").param("version", "new"))
                .andExpect(jsonPath("$.status").value(200)).andExpect(jsonPath("$.data").value(true));
        http.perform(post("/sync-testcase").param("token", "test-secret").param("pid", "1").param("version", "new"))
                .andExpect(jsonPath("$.status").value(400));
    }
}
