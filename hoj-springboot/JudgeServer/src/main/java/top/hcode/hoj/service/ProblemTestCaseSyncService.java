package top.hcode.hoj.service;

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
import java.util.Comparator;
import java.util.stream.Stream;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

@Service
public class ProblemTestCaseSyncService {

    private static final long MAX_UNPACKED_BYTES = 2L * 1024 * 1024 * 1024;
    private static final int MAX_ENTRIES = 20000;

    public void delete(Long pid) throws IOException {
        Path target = Paths.get(Constants.JudgeDir.TEST_CASE_DIR.getContent(), "problem_" + pid);
        deleteTree(target);
    }

    public void install(Long pid, String version, MultipartFile archive) throws IOException {
        if (archive == null || archive.isEmpty()) {
            throw new IOException("测试数据压缩包为空");
        }
        Path root = Paths.get(Constants.JudgeDir.TEST_CASE_DIR.getContent());
        Files.createDirectories(root);
        Path temporary = Files.createTempDirectory(root, "problem_" + pid + "_sync_");
        try {
            unpack(archive.getInputStream(), temporary);
            Path info = temporary.resolve("info");
            if (!Files.isRegularFile(info)) {
                throw new IOException("测试数据压缩包缺少 info 文件");
            }
            String infoJson = new String(Files.readAllBytes(info), StandardCharsets.UTF_8);
            String actualVersion = JSONUtil.parseObj(infoJson).getStr("version");
            if (!version.equals(actualVersion)) {
                throw new IOException("测试数据版本不一致");
            }
            replaceDirectory(root.resolve("problem_" + pid), temporary);
        } finally {
            deleteTree(temporary);
        }
    }

    public void installInfo(Long pid, String version, MultipartFile infoFile) throws IOException {
        if (infoFile == null || infoFile.isEmpty()) throw new IOException("测试数据 info 文件为空");
        Path target = Paths.get(Constants.JudgeDir.TEST_CASE_DIR.getContent(), "problem_" + pid);
        if (!Files.isDirectory(target)) throw new IOException("判题服务器尚未同步该题测试数据");
        byte[] content = infoFile.getBytes();
        String actualVersion = JSONUtil.parseObj(new String(content, StandardCharsets.UTF_8)).getStr("version");
        if (!version.equals(actualVersion)) throw new IOException("测试数据版本不一致");
        Path temporary = Files.createTempFile(target, "info_", ".tmp");
        try {
            Files.write(temporary, content, StandardOpenOption.TRUNCATE_EXISTING);
            replaceFile(temporary, target.resolve("info"));
        } finally {
            Files.deleteIfExists(temporary);
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
                Path destination = target.resolve(entry.getName()).normalize();
                if (!destination.startsWith(target) || entry.getName().startsWith("/")) {
                    throw new IOException("测试数据压缩包包含非法路径");
                }
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
