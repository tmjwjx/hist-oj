package top.hcode.hoj.service;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;
import top.hcode.hoj.util.Constants;

import java.io.BufferedInputStream;
import java.io.BufferedOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.regex.Pattern;
import java.util.stream.Stream;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

@Service
public class ProblemTestCaseSyncService {

    private static final long MAX_UNPACKED_BYTES = 2L * 1024 * 1024 * 1024;
    private static final int MAX_ENTRIES = 20000;
    private static final Pattern SHARED_TOKEN = Pattern.compile(
            "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}");
    private static final String[] FINGERPRINTS = {"inputMd5", "inputSize", "outputMd5", "outputSize",
            "allStrippedOutputMd5", "EOFStrippedOutputMd5"};
    private final Path root;

    public ProblemTestCaseSyncService() {
        this(Paths.get(Constants.JudgeDir.TEST_CASE_DIR.getContent()));
    }

    ProblemTestCaseSyncService(Path root) {
        this.root = root.toAbsolutePath().normalize();
    }

    /** A fresh marker proves shared storage; a matching info version alone does not. */
    public boolean confirmShared(Long pid, String version, String sharedToken) throws IOException {
        validateRequest(pid, version);
        if (sharedToken == null || !SHARED_TOKEN.matcher(sharedToken).matches()) {
            throw new IOException("共享测试目录校验凭据无效");
        }
        Path marker = root.resolve(".hoj-sync-" + sharedToken);
        BasicFileAttributes attributes;
        try {
            attributes = Files.readAttributes(marker, BasicFileAttributes.class, LinkOption.NOFOLLOW_LINKS);
        } catch (NoSuchFileException e) {
            return false;
        }
        if (!attributes.isRegularFile() || attributes.size() != sharedToken.length()
                || !sharedToken.equals(new String(Files.readAllBytes(marker), StandardCharsets.UTF_8))) {
            throw new IOException("共享测试目录校验凭据不匹配");
        }
        Path target = root.resolve("problem_" + pid);
        validateInfo(target, version, readInfo(target));
        return true;
    }

    public void delete(Long pid) throws IOException {
        Path target = root.resolve("problem_" + pid);
        deleteTree(target);
    }

    public void install(Long pid, String version, MultipartFile archive) throws IOException {
        validateRequest(pid, version);
        if (archive == null || archive.isEmpty()) {
            throw new IOException("测试数据压缩包为空");
        }
        Files.createDirectories(root);
        Path temporary = Files.createTempDirectory(root, "problem_" + pid + "_sync_");
        try {
            unpack(archive.getInputStream(), temporary);
            validateInfo(temporary, version, readInfo(temporary));
            replaceDirectory(root.resolve("problem_" + pid), temporary);
        } finally {
            deleteTree(temporary);
        }
    }

    public void installInfo(Long pid, String version, MultipartFile infoFile) throws IOException {
        validateRequest(pid, version);
        if (infoFile == null || infoFile.isEmpty()) throw new IOException("测试数据 info 文件为空");
        Path target = root.resolve("problem_" + pid);
        byte[] content = infoFile.getBytes();
        JSONObject updated = parseInfo(content);
        Map<String, Map<String, String>> files = validateInfo(target, version, updated);
        JSONObject previous = readInfo(target);
        if (!files.equals(validateInfo(target, previous.getStr("version"), previous))) {
            throw new IOException("测试数据文件清单或指纹已变化，需要完整同步");
        }
        Path temporary = Files.createTempFile(target, "info_", ".tmp");
        try {
            Files.write(temporary, content, StandardOpenOption.TRUNCATE_EXISTING);
            replaceFile(temporary, target.resolve("info"));
        } finally {
            Files.deleteIfExists(temporary);
        }
    }

    private void validateRequest(Long pid, String version) throws IOException {
        if (pid == null || pid <= 0 || version == null || version.trim().isEmpty()) {
            throw new IOException("题目编号和测试数据版本无效");
        }
    }

    private JSONObject readInfo(Path directory) throws IOException {
        return parseInfo(Files.readAllBytes(requireFile(directory, "info")));
    }

    private JSONObject parseInfo(byte[] content) throws IOException {
        try {
            return JSONUtil.parseObj(new String(content, StandardCharsets.UTF_8));
        } catch (RuntimeException e) {
            throw new IOException("测试数据 info 格式无效", e);
        }
    }

    /** Validate the actual files before acknowledging or replacing any published version. */
    private Map<String, Map<String, String>> validateInfo(Path directory, String version, JSONObject info)
            throws IOException {
        try {
            if (version == null || version.trim().isEmpty() || !version.equals(info.getStr("version"))) {
                throw new IOException("测试数据版本不一致");
            }
            JSONArray cases = info.getJSONArray("testCases");
            Object count = info.get("testCasesSize");
            if (cases == null || cases.isEmpty() || cases.size() > MAX_ENTRIES
                    || !(count instanceof Number) || ((Number) count).doubleValue() != cases.size()) {
                throw new IOException("测试数据清单数量无效");
            }
            Map<String, Map<String, String>> files = new LinkedHashMap<>();
            for (int i = 0; i < cases.size(); i++) {
                JSONObject item = cases.getJSONObject(i);
                if (item == null) throw new IOException("测试数据清单条目无效");
                String input = item.getStr("inputName"), output = item.getStr("outputName");
                validateSize(requireFile(directory, input), item.get("inputSize"));
                validateSize(requireFile(directory, output), item.get("outputSize"));
                Map<String, String> fingerprints = new LinkedHashMap<>();
                for (String field : FINGERPRINTS) fingerprints.put(field, item.getStr(field));
                if (files.put(input + "\0" + output, fingerprints) != null) {
                    throw new IOException("测试数据清单包含重复文件对");
                }
            }
            return files;
        } catch (RuntimeException e) {
            throw new IOException("测试数据 info 格式无效", e);
        }
    }

    private void validateSize(Path file, Object expected) throws IOException {
        if (expected == null) return;
        if (!(expected instanceof Number) || ((Number) expected).longValue() < 0
                || ((Number) expected).doubleValue() != ((Number) expected).longValue()
                || Files.size(file) != ((Number) expected).longValue()) {
            throw new IOException("测试数据文件大小不匹配：" + file.getFileName());
        }
    }

    private Path requireFile(Path directory, String name) throws IOException {
        if (!Files.isDirectory(directory, LinkOption.NOFOLLOW_LINKS)) {
            throw new IOException("测试数据目录不存在或不是普通目录");
        }
        Path file = safePath(directory, name);
        Path current = directory;
        for (Path part : directory.relativize(file)) {
            current = current.resolve(part);
            if (Files.isSymbolicLink(current)) throw new IOException("测试数据不允许符号链接");
        }
        if (!Files.isRegularFile(file, LinkOption.NOFOLLOW_LINKS)) {
            throw new IOException("测试数据文件不存在：" + name);
        }
        return file;
    }

    private Path safePath(Path directory, String name) throws IOException {
        try {
            if (name == null || name.trim().isEmpty() || name.contains("\\") || name.indexOf('\0') >= 0) {
                throw new IOException("测试数据包含非法路径");
            }
            Path relative = Paths.get(name);
            if (relative.isAbsolute()) throw new IOException("测试数据包含非法路径");
            for (Path part : relative) {
                if ("..".equals(part.toString()) || ".".equals(part.toString())) {
                    throw new IOException("测试数据包含非法路径");
                }
            }
            Path file = directory.resolve(relative).normalize();
            if (!file.startsWith(directory) || file.equals(directory)) throw new IOException("测试数据包含非法路径");
            return file;
        } catch (InvalidPathException e) {
            throw new IOException("测试数据包含非法路径", e);
        }
    }

    private void unpack(InputStream input, Path target) throws IOException {
        long totalBytes = 0;
        int entries = 0;
        try (ZipInputStream zip = new ZipInputStream(new BufferedInputStream(input))) {
            ZipEntry entry;
            byte[] buffer = new byte[8192];
            while ((entry = zip.getNextEntry()) != null) {
                if (++entries > MAX_ENTRIES) {
                    throw new IOException("测试数据文件数量超出限制");
                }
                Path destination = safePath(target, entry.getName());
                if (entry.isDirectory()) {
                    Files.createDirectories(destination);
                    continue;
                }
                Files.createDirectories(destination.getParent());
                try (OutputStream output = new BufferedOutputStream(Files.newOutputStream(destination))) {
                    int length;
                    while ((length = zip.read(buffer)) >= 0) {
                        totalBytes += length;
                        if (totalBytes > MAX_UNPACKED_BYTES) {
                            throw new IOException("测试数据解压后超过大小限制");
                        }
                        output.write(buffer, 0, length);
                    }
                }
            }
        }
    }

    private void replaceDirectory(Path target, Path temporary) throws IOException {
        Path backup = target.resolveSibling(target.getFileName() + ".bak");
        deleteTree(backup);
        if (Files.exists(target)) {
            move(target, backup);
        }
        try {
            move(temporary, target);
            deleteTree(backup);
        } catch (IOException e) {
            deleteTree(target);
            if (Files.exists(backup)) {
                move(backup, target);
            }
            throw e;
        }
    }

    private void move(Path source, Path target) throws IOException {
        try {
            Files.move(source, target, StandardCopyOption.ATOMIC_MOVE);
        } catch (AtomicMoveNotSupportedException e) {
            Files.move(source, target);
        }
    }

    private void replaceFile(Path source, Path target) throws IOException {
        try {
            Files.move(source, target, StandardCopyOption.ATOMIC_MOVE, StandardCopyOption.REPLACE_EXISTING);
        } catch (AtomicMoveNotSupportedException e) {
            Files.move(source, target, StandardCopyOption.REPLACE_EXISTING);
        }
    }

    private void deleteTree(Path path) throws IOException {
        if (!Files.exists(path)) {
            return;
        }
        try (Stream<Path> paths = Files.walk(path)) {
            try {
                paths.sorted(Comparator.reverseOrder()).forEach(item -> {
                    try {
                        Files.deleteIfExists(item);
                    } catch (IOException e) {
                        throw new DeleteException(e);
                    }
                });
            } catch (DeleteException e) {
                throw (IOException) e.getCause();
            }
        }
    }

    private static class DeleteException extends RuntimeException {
        private DeleteException(IOException cause) {
            super(cause);
        }
    }
}
