package top.hcode.hoj.service.problem;

import org.springframework.stereotype.Service;
import top.hcode.hoj.utils.Constants;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.stream.Stream;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

@Service
public class ProblemTestCaseArchiveService {

    public File createArchive(Long pid) throws IOException {
        Path source = Paths.get(Constants.File.TESTCASE_BASE_FOLDER.getPath(), "problem_" + pid);
        if (!Files.isDirectory(source) || !Files.exists(source.resolve("info"))) {
            throw new IOException("题目测试数据目录或 info 文件不存在");
        }

        File archive = File.createTempFile("problem-" + pid + "-", ".zip");
        try (ZipOutputStream output = new ZipOutputStream(
                new BufferedOutputStream(new java.io.FileOutputStream(archive)));
             Stream<Path> paths = Files.walk(source)) {
            paths.filter(Files::isRegularFile).forEach(path -> {
                String entryName = source.relativize(path).toString().replace(File.separatorChar, '/');
                try {
                    output.putNextEntry(new ZipEntry(entryName));
                    try (FileInputStream input = new FileInputStream(path.toFile())) {
                        byte[] buffer = new byte[8192];
                        int length;
                        while ((length = input.read(buffer)) >= 0) {
                            output.write(buffer, 0, length);
                        }
                    }
                    output.closeEntry();
                } catch (IOException e) {
                    throw new ArchiveException(e);
                }
            });
        } catch (ArchiveException e) {
            Files.deleteIfExists(archive.toPath());
            throw e.ioCause;
        } catch (IOException e) {
            Files.deleteIfExists(archive.toPath());
            throw e;
        }
        return archive;
    }

    private static class ArchiveException extends RuntimeException {
        private final IOException ioCause;

        private ArchiveException(IOException cause) {
            super(cause);
            this.ioCause = cause;
        }
    }
}
