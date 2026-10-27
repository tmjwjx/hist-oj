package top.hcode.hoj.judge;

import cn.hutool.core.io.FileUtil;
import cn.hutool.extra.spring.SpringUtil;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.CompileError;
import top.hcode.hoj.common.exception.SubmitError;
import top.hcode.hoj.common.exception.SystemError;
import top.hcode.hoj.judge.entity.LanguageConfig;
import top.hcode.hoj.util.Constants;
import top.hcode.hoj.util.JudgeUtils;

import java.io.File;
import java.util.HashMap;
import java.util.List;

/**
 * @Author: Himit_ZH
 * @Date: 2021/4/16 12:14
 * @Description: 判题流程解耦重构2.0，该类只负责编译
 */
public class Compiler {

    private static final Object[] EXTRA_PROGRAM_LOCKS = new Object[64];

    static {
        for (int i = 0; i < EXTRA_PROGRAM_LOCKS.length; i++) {
            EXTRA_PROGRAM_LOCKS[i] = new Object();
        }
    }

    static Object extraProgramLock(Long pid) {
        return pid == null ? new Object() : EXTRA_PROGRAM_LOCKS[Math.floorMod(pid.hashCode(), EXTRA_PROGRAM_LOCKS.length)];
    }

    public static String compile(LanguageConfig languageConfig, String code,
                                 String language, HashMap<String, String> extraFiles) throws SystemError, CompileError, SubmitError {

        if (languageConfig == null) {
            throw new RuntimeException("Unsupported language " + language);
        }

        // 调用安全沙箱进行编译
        JSONArray result = SandboxRun.compile(languageConfig.getMaxCpuTime(),
                languageConfig.getMaxRealTime(),
                languageConfig.getMaxMemory(),
                256 * 1024 * 1024L,
                languageConfig.getSrcName(),
                languageConfig.getExeName(),
                parseCompileCommand(languageConfig.getCompileCommand()),
                languageConfig.getCompileEnvs(),
                code,
                extraFiles,
                true,
                false,
                null
        );
        JSONObject compileResult = (JSONObject) result.get(0);
        if (compileResult.getInt("status").intValue() != Constants.Judge.STATUS_ACCEPTED.getStatus()) {
            throw new CompileError("Compile Error.", ((JSONObject) compileResult.get("files")).getStr("stdout"),
                    ((JSONObject) compileResult.get("files")).getStr("stderr"));
        }

        String fileId = ((JSONObject) compileResult.get("fileIds")).getStr(languageConfig.getExeName());
        if (StringUtils.isEmpty(fileId)) {
            throw new SubmitError("Executable file not found.", ((JSONObject) compileResult.get("files")).getStr("stdout"),
                    ((JSONObject) compileResult.get("files")).getStr("stderr"));
        }
        return fileId;
    }

    public static Boolean compileSpj(String code, Long pid, String language, HashMap<String, String> extraFiles) throws SystemError {
        return compileExtraProgram(code, pid, language, extraFiles, false);
    }

    public static Boolean compileInteractive(String code, Long pid, String language, HashMap<String, String> extraFiles) throws SystemError {
        return compileExtraProgram(code, pid, language, extraFiles, true);
    }

    private static Boolean compileExtraProgram(String code, Long pid, String language,
                                                HashMap<String, String> extraFiles, boolean interactive) throws SystemError {
        LanguageConfigLoader languageConfigLoader = SpringUtil.getBean(LanguageConfigLoader.class);
        String prefix = interactive ? "INTERACTIVE-" : "SPJ-";
        LanguageConfig languageConfig = languageConfigLoader.getLanguageConfigByName(prefix + language);
        if (languageConfig == null) {
            throw new SystemError("Unsupported " + prefix + "language: " + language, null, null);
        }
        String programDir = (interactive ? Constants.JudgeDir.INTERACTIVE_WORKPLACE_DIR
                : Constants.JudgeDir.SPJ_WORKPLACE_DIR).getContent() + File.separator + pid;
        synchronized (extraProgramLock(pid)) {
            // 后台的独立编译入口也会覆盖可执行文件，必须使旧缓存失效。
            if (pid != null) {
                FileUtil.del(programDir + File.separator + "version");
            }
            JSONArray result = SandboxRun.compile(languageConfig.getMaxCpuTime(),
                    languageConfig.getMaxRealTime(),
                    languageConfig.getMaxMemory(),
                    256 * 1024 * 1024L,
                    languageConfig.getSrcName(),
                    languageConfig.getExeName(),
                    parseCompileCommand(languageConfig.getCompileCommand()),
                    languageConfig.getCompileEnvs(),
                    code,
                    extraFiles,
                    false,
                    pid != null,
                    programDir);
            JSONObject compileResult = (JSONObject) result.get(0);
            if (compileResult.getInt("status").intValue() != Constants.Judge.STATUS_ACCEPTED.getStatus()) {
                throw new SystemError(interactive ? "Interactive Judge Code Compile Error." : "Special Judge Code Compile Error.",
                        ((JSONObject) compileResult.get("files")).getStr("stdout"),
                        ((JSONObject) compileResult.get("files")).getStr("stderr"));
            }
            return true;
        }
    }

    private static List<String> parseCompileCommand(String command) {
        return JudgeUtils.translateCommandline(command);
    }
}