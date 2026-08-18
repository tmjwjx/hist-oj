package top.hcode.hoj.manager.plagiarism;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Slf4j
@Component
public class PlagiarismSimilarityEngine {
    private static final Pattern PERCENT = Pattern.compile("(\\d+)\\s*%");
    private static final Pattern TOKEN = Pattern.compile(
            "[A-Za-z_][A-Za-z0-9_]*|\\d+(?:\\.\\d+)?|==|!=|<=|>=|&&|\\|\\||\\+\\+|--|\\+=|-=|\\*=|/=|->|<<|>>|[{}()\\[\\];,.:?+\\-*/%<>=!&|^~]");
    private static final Set<String> KEYWORDS = new HashSet<>(Arrays.asList(
            "if", "else", "for", "while", "do", "switch", "case", "default", "break", "continue",
            "return", "class", "struct", "interface", "public", "private", "protected", "static",
            "final", "const", "void", "int", "long", "short", "float", "double", "char", "boolean",
            "bool", "new", "this", "true", "false", "null", "package", "import", "using", "namespace",
            "try", "catch", "finally", "throw", "throws", "extends", "implements", "template", "typename"));

    public int[] compare(String code1, String code2, String language) {
        if (!supported(language)) {
            return new int[]{0, 0};
        }
        int[] external = runSim(code1, code2, language);
        return external == null ? fallback(code1, code2) : external;
    }

    private int[] runSim(String code1, String code2, String language) {
        if (Boolean.getBoolean("hoj.sim.disable")) {
            return null;
        }
        String executable = executable(language);
        if (executable == null) {
            return null;
        }
        Path directory = null;
        try {
            directory = Files.createTempDirectory("hoj-sim-");
            Path file1 = Files.write(directory.resolve("a.tmp"), code1.getBytes(StandardCharsets.UTF_8));
            Path file2 = Files.write(directory.resolve("b.tmp"), code2.getBytes(StandardCharsets.UTF_8));
            String binDir = Optional.ofNullable(System.getenv("HOJ_SIM_BIN_DIR")).orElse("/usr/local/bin");
            Process process = new ProcessBuilder(Paths.get(binDir, executable).toString(), "-p", "-a", "-t", "1",
                    file1.toString(), file2.toString()).redirectErrorStream(true).start();
            if (!process.waitFor(5, TimeUnit.SECONDS)) {
                process.destroyForcibly();
                return null;
            }
            String output = new String(readAll(process), StandardCharsets.UTF_8);
            if (process.exitValue() != 0) {
                return null;
            }
            Matcher matcher = PERCENT.matcher(output);
            List<Integer> values = new ArrayList<>();
            while (matcher.find()) {
                values.add(Integer.valueOf(matcher.group(1)));
            }
            if (values.isEmpty()) {
                return new int[]{0, 0};
            }
            return new int[]{values.get(0), values.size() > 1 ? values.get(1) : values.get(0)};
        } catch (Exception ex) {
            log.debug("sim 工具不可用，使用内置查重算法", ex);
            return null;
        } finally {
            if (directory != null) {
                deleteDirectory(directory);
            }
        }
    }

    private byte[] readAll(Process process) throws IOException {
        java.io.ByteArrayOutputStream output = new java.io.ByteArrayOutputStream();
        byte[] buffer = new byte[4096];
        int count;
        while ((count = process.getInputStream().read(buffer)) >= 0) {
            output.write(buffer, 0, count);
        }
        return output.toByteArray();
    }

    private int[] fallback(String code1, String code2) {
        Map<String, Integer> first = ngrams(tokenize(code1));
        Map<String, Integer> second = ngrams(tokenize(code2));
        return new int[]{coverage(first, second), coverage(second, first)};
    }

    private int coverage(Map<String, Integer> source, Map<String, Integer> target) {
        int total = source.values().stream().mapToInt(Integer::intValue).sum();
        if (total == 0) {
            return 0;
        }
        int shared = 0;
        for (Map.Entry<String, Integer> entry : source.entrySet()) {
            shared += Math.min(entry.getValue(), target.getOrDefault(entry.getKey(), 0));
        }
        return Math.min(100, shared * 100 / total);
    }

    private Map<String, Integer> ngrams(List<String> tokens) {
        Map<String, Integer> result = new HashMap<>();
        int size = tokens.size() < 5 ? 1 : 5;
        for (int i = 0; i + size <= tokens.size(); i++) {
            String key = String.join(" ", tokens.subList(i, i + size));
            result.put(key, result.getOrDefault(key, 0) + 1);
        }
        return result;
    }

    private List<String> tokenize(String source) {
        String code = source == null ? "" : source;
        code = code.replaceAll("(?s)/\\*.*?\\*/", " ").replaceAll("(?m)//.*$", " ");
        code = code.replaceAll("\\\"(?:\\\\.|[^\\\"\\\\])*\\\"|'(?:\\\\.|[^'\\\\])*'", " LITERAL ");
        List<String> tokens = new ArrayList<>();
        Matcher matcher = TOKEN.matcher(code);
        while (matcher.find()) {
            String token = matcher.group();
            if (token.matches("[A-Za-z_][A-Za-z0-9_]*") && !KEYWORDS.contains(token)) {
                token = "ID";
            } else if (token.matches("\\d+(?:\\.\\d+)?")) {
                token = "NUM";
            }
            tokens.add(token);
        }
        return tokens;
    }

    private boolean supported(String language) {
        String value = language == null ? "" : language.toLowerCase(Locale.ROOT);
        return value.contains("c++") || value.contains("cpp") || value.contains("gcc") || value.contains("g++")
                || "c".equals(value) || value.startsWith("c ") || value.contains("java")
                || value.contains("pascal") || value.contains("pas");
    }

    private String executable(String language) {
        String value = language == null ? "" : language.toLowerCase(Locale.ROOT);
        if (value.contains("java")) return "sim_java";
        if (value.contains("pascal") || value.contains("pas")) return "sim_pasc";
        return "sim_c";
    }

    private void deleteDirectory(Path directory) {
        try {
            Files.walk(directory).sorted(Comparator.reverseOrder()).forEach(path -> {
                try { Files.deleteIfExists(path); } catch (IOException ignored) { }
            });
        } catch (IOException ignored) { }
    }
}
