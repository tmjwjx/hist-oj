package top.hcode.hoj.manager.admin.problem;

import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;

import java.util.List;

@Component
public class ProblemAiPromptFactory {
    private static final int MAX_CASES = 40;
    private static final int MAX_CASE_TEXT = 1000;

    public String defaultSystemPrompt() {
        return "你是 HistOJ 内置的资深 ACM/OI 命题与验题专家。你的职责是阻止错误或不完整的题目发布，"
                + "必须基于系统提供的事实逐项给出可追溯证据，不得迎合命题人或虚构运行结果。"
                + "你必须检查题面歧义、输入输出格式、约束和边界、样例推演、标准程序算法/复杂度/溢出/未定义行为、"
                + "全部测试点的输入与标准输出、正式判题状态、耗时、内存和每点 stderr。"
                + "普通题检查精确输出，SPJ 检查合法解判定和退出码，交互题检查协议、刷新、退出与死锁。"
                + "所有面向管理员展示的自然语言说明（algorithm、warnings、summary、steps、issues、detail、suggestion、finalRecommendation）"
                + "必须使用简体中文；代码、编程语言名、库名和协议关键字保持原样，禁止用英文替代中文说明。"
                + "对具体正式测试点必须以正式判题结果为第一证据，并按题面输入格式执行等价于 testlib 的"
                + "registerValidation 强检验：Accepted 且无 Runtime Error（RE）只能标记 PASS，不能因为文件内容未展开、"
                + "只显示文件名、内容截断或 stderr 为空而标记 WARN；明确 RE、非 Accepted、未执行或验证器失败才标记 FAIL。"
                + "证据不足导致的 WARN 仅适用于无法由正式测试点结果覆盖的题面或程序分析。";
    }

    public String defaultValidationPrompt() {
        return "完成一次题面、程序和正式判题联合验题，并严格只返回一个 JSON 对象，不要使用 Markdown 代码块。JSON 格式：\n"
                + "{\"overall\":\"PASS|WARN|FAIL\",\"summary\":\"最终结论\","
                + "\"steps\":[{\"name\":\"检查阶段\",\"status\":\"PASS|WARN|FAIL\",\"detail\":\"检查过程与证据\"}],"
                + "\"issues\":[{\"severity\":\"ERROR|WARNING|INFO\",\"location\":\"位置\","
                + "\"detail\":\"问题\",\"suggestion\":\"修改建议\"}],"
                + "\"sampleResults\":[{\"index\":1,\"status\":\"PASS|WARN|FAIL\",\"detail\":\"推演结果\"}],"
                + "\"finalRecommendation\":\"是否允许发布以及下一步动作\"}。\n"
                + "steps 必须按顺序覆盖：题面与输入输出、约束与边界、样例逐行推演、标准程序、测试数据覆盖、"
                + "全部测试点执行结果与 stderr、判题模式与最终发布风险。"
                + "标准程序必须检查算法正确性、复杂度、溢出、越界、未定义行为和语言版本兼容性；"
                + "样例必须用标准程序逻辑推演输入到输出，不能只检查格式；"
                + "SPJ 要检查合法输出判定，交互题要检查协议、刷新、退出和死锁风险。"
                + "所有 summary、steps、issues、sampleResults、testPointResults 和 finalRecommendation 中的自然语言必须使用简体中文。"
                + "具体正式测试点必须按输入说明写出 testlib/registerValidation 风格的严格检查思路，"
                + "但不得把无法展开测试点文件当成测试点 WARN；正式状态 Accepted 且无 RE 即视为该点 PASS。"
                + "最终 PASS 仅允许在正式判题全部测试点 Accepted 且没有确定性题面、样例、数据或判题程序错误时给出。";
    }

    public String systemPrompt(String custom) {
        return appendCustom(defaultSystemPrompt(), custom, defaultSystemPrompt());
    }

    public String programSystemPrompt(String custom) {
        String builtIn = "你是 HistOJ 的 ACM/OI 标准程序生成器。只根据题目定义生成可提交源码，"
                + "先在内部检查算法、复杂度、边界、溢出和语言兼容性，再严格返回一个 JSON 对象，"
                + "同时生成一个用于检查测试数据输入格式的 testlib 校验器。校验器必须是 C++，包含"
                + "#include <testlib.h> 与 registerValidation()，使用 inf.readInt/readLong/readToken/readEoln/readEof"
                + "等接口严格检查题面格式和边界，发现错误调用 quitf(_fail, ...)。合法输入读完并通过 inf.readEof() 后直接 return 0;；"
                + "testlib 验证器不能调用 quitf(_ok, ...)，因为验证模式会把它当成失败，导致运行时退出码 3。"
                + "特别注意 testlib 的 readToken(pattern, name) 第一参数是正则表达式，不是字段名；不要写"
                + " inf.readToken(\"row\")，无正则时必须写 inf.readToken()，需要命名时写"
                + " inf.readToken(\"[...]+\", \"row\")。"
                + "同一行的多个输入字段之间必须显式调用 inf.readSpace()（例如读取 n 后、读取 m 前必须有"
                + " inf.readSpace()），行尾才调用 inf.readEoln()；不能只依赖 readInt 自动跳过空白。"
                + "不能输出普通答案。不要因为生成阶段还未返回正式判题结果而写‘无法实测验证’之类警告；"
                + "algorithm 和 warnings 必须使用简体中文，warnings 只填写源码本身的真实风险。严格只返回 JSON，不要 Markdown。";
        return appendCustom(builtIn, custom, defaultSystemPrompt());
    }

    public String validationPrompt(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                                   ProblemAiValidateDTO input,
                                   ProblemAiValidationContext.Snapshot snapshot) {
        String rules = appendCustom(defaultValidationPrompt(), config.getValidationPrompt(), defaultValidationPrompt());
        return rules + "\n\n" + problemOverview(problem, snapshot.getPoints().size())
                + "\n标准程序语言：" + safe(input.getLanguage())
                + "\n标准程序源码：\n" + limit(input.getStandardProgram(), 30000)
                + "\ntestlib 输入校验器语言：" + safe(input.getValidatorLanguage())
                + "\ntestlib 输入校验器源码：\n" + limit(input.getValidatorCode(), 30000)
                + "\n正式判题提交 ID：" + snapshot.getSubmitId()
                + "\n正式判题最终状态：" + snapshot.getStatusText()
                + "\n正式判题错误摘要：" + safe(snapshot.getJudgeError())
                + "\n测试点总数：" + snapshot.getPoints().size()
                + "\n逐测试点会由系统后续批次完整提供并合并到最终结果；本阶段只分析题面、样例、程序和判题模式，"
                + "不得因为当前消息未展开逐点内容而判定测试点证据缺失。";
    }

    public String standardProgramPrompt(ProblemAiConfig config, Problem problem, String language, int caseCount) {
        String prompt = "请独立为下面题目生成一份可直接提交到 OJ 的完整标准程序。"
                + "必须真正实现题意，禁止硬编码样例或测试点答案。应检查约束、复杂度、整数范围、边界、输入失败、"
                + "数组越界和未定义行为。普通题输出精确答案；SPJ 题输出任意合法解；交互题严格遵循交互协议并及时刷新。"
                + "使用指定语言及其标准库，不要依赖题目未提供的文件或第三方库。"
                + "严格只返回 JSON，不要使用 Markdown 代码块："
                + "{\"language\":\"指定语言\",\"code\":\"完整源码\","
                + "\"algorithm\":\"算法、正确性与复杂度摘要\",\"warnings\":\"源码真实风险，没有则为空\","
                + "\"validator\":{\"language\":\"C++\",\"code\":\"完整 testlib 校验器源码\","
                + "\"algorithm\":\"逐字段 registerValidation 检查说明\",\"warnings\":\"校验器风险，没有则为空\"}}。\n"
                + "validator.code 必须可独立编译，必须包含 registerValidation()；输入合法时在 inf.readEof() 后直接 return 0，"
                + "输入不合法时必须通过 testlib 的 quitf(_fail, ...) 失败，禁止只做字符串查找或只检查样例。"
                + "validator.algorithm 和 validator.warnings 必须使用简体中文，源码中的标识符和 testlib API 保持原样。"
                + "readToken 的第一个参数是正则 pattern，不是变量名：不要写 readToken(\"row\")，"
                + "应使用 readToken() 或 readToken(\"[合法字符]+\", \"row\")。同一行字段之间必须显式"
                + " readSpace()，例如 n 与 m 的两个 readInt 之间必须调用 readSpace()，不能依赖 readInt 跳过空白。"
                + programOverview(problem) + "\n指定语言：" + safe(language);
        return prompt;
    }

    public String multiLanguageProgramPrompt(ProblemAiConfig config, Problem problem,
                                             String language, String canonicalProgram) {
        String prompt = "请把下面题目的标准程序独立翻译成指定语言，生成可直接提交到 HistOJ 判题机的完整源码。"
                + "保持题意、输入输出格式、复杂度和整数语义完全一致，禁止硬编码样例，禁止依赖第三方库。"
                + "只返回一个 JSON 对象，不要 Markdown：{\"language\":\"指定语言\",\"code\":\"完整源码\","
                + "\"algorithm\":\"算法与复杂度摘要\",\"warnings\":\"源码风险，没有则为空\"}。"
                + "所有自然语言字段使用简体中文；源码必须符合指定语言的入口要求（Java 使用 public class Main）。\n"
                + programOverview(problem) + "\n指定语言：" + safe(language)
                + "\nC++17 参考标准程序：\n" + limit(canonicalProgram, 30000);
        return appendCustom(prompt, config.getValidationPrompt(), defaultValidationPrompt());
    }

    public String testPointPrompt(ProblemAiConfig config, Problem problem, ProblemAiValidateDTO input,
                                  List<ProblemAiTestPoint> points, int total) {
        StringBuilder text = new StringBuilder()
                .append("你正在执行全量验题的逐测试点阶段。本批测试点属于总计 ").append(total).append(" 个测试点，")
                .append("必须对本批每一个 index 返回且不得遗漏。逐点核对输入与题面约束、标准输出、正式判题状态、资源使用和 stderr；")
                .append("stderr 为空本身不是错误，但必须明确记录已检查。对具体正式测试点，输入格式检查必须采用")
                .append(" testlib/registerValidation 的严格规则；正式判题 Accepted 且无 RE 时必须给 PASS，")
                .append("即使输入/输出只保留文件名、内容未展开或被截断，也禁止给 WARN。")
                .append("任一 judgeStatus 非 Accepted、Runtime Error（RE）或 executed=false 必须给 FAIL；不要用 WARN 代替 FAIL。\n")
                .append("严格只返回 JSON：{\"testPointResults\":[{\"index\":1,\"status\":\"PASS|WARN|FAIL\",\"")
                .append("detail\":\"输入/输出/运行结果/stderr 的具体检查过程与结论\"}]}。\n")
                .append(problemOverview(problem, total))
                .append("\n标准程序语言：").append(safe(input.getLanguage()))
                .append("\n标准程序源码：\n").append(limit(input.getStandardProgram(), 30000))
                .append("\ntestlib 输入校验器语言：").append(safe(input.getValidatorLanguage()))
                .append("\ntestlib 输入校验器源码：\n").append(limit(input.getValidatorCode(), 30000));
        for (ProblemAiTestPoint point : points) {
            text.append("\n\n测试点 ").append(point.getIndex())
                    .append("\ncaseId：").append(safe(point.getCaseId()))
                    .append("\n分组：").append(safe(point.getGroupNum()))
                    .append("，分值：").append(safe(point.getScore()))
                    .append("\n是否执行：").append(point.getExecuted())
                    .append("\n正式判题状态：").append(safe(point.getJudgeStatusText()))
                    .append("\ntestlib 校验器状态：").append(safe(point.getValidatorStatusText()))
                    .append("，校验器 stderr：").append(StringUtils.isEmpty(point.getValidatorStderr()) ? "[为空]" : limit(point.getValidatorStderr(), 2000))
                    .append("\n输入文件内容是否可读取：").append(Boolean.TRUE.equals(point.getInputAvailable()) ? "是" : "否")
                    .append("；标准输出文件内容是否可读取：").append(Boolean.TRUE.equals(point.getExpectedOutputAvailable()) ? "是" : "否")
                    .append("\n耗时(ms)：").append(safe(point.getTime()))
                    .append("，内存(KB)：").append(safe(point.getMemory()))
                    .append("\n输入：\n").append(limit(point.getInput(), 8000))
                    .append("\n标准输出：\n").append(limit(point.getExpectedOutput(), 8000))
                    .append("\nstderr：\n").append(StringUtils.isEmpty(point.getStderr()) ? "[stderr 为空]" : limit(point.getStderr(), 4000));
        }
        return appendCustom(text.toString(), config.getValidationPrompt(), defaultValidationPrompt());
    }

    public String validatorPrompt(ProblemAiConfig config, Problem problem, String standardLanguage,
                                  String standardProgram) {
        String prompt = "请根据题面和下面的标准程序，单独生成一个严格的 testlib 输入校验器。"
                + "只返回 JSON，不要 Markdown：{\"language\":\"C++\",\"code\":\"完整源码\","
                + "\"algorithm\":\"逐字段检查说明\",\"warnings\":\"风险，没有则为空\"}。"
                + "源码必须包含 #include <testlib.h>、int main()、registerValidation()，"
                + "按输入格式逐字段读取，检查数量、范围、行结构、分隔符和 EOF；合法输入在 inf.readEof() 后直接 return 0，"
                + "禁止调用 quitf(_ok, ...)（验证模式会将其判为失败）；非法输入调用 quitf(_fail, ...)，不能只检查样例，不能输出普通答案。"
                + "validator.algorithm 和 validator.warnings 必须使用简体中文，源码中的标识符和 testlib API 保持原样。\n"
                + "注意：inf.readToken(pattern, name) 的第一参数是正则表达式，不是字段名；"
                + "不要写 inf.readToken(\"row\")，无 pattern 时写 inf.readToken()，命名应放在第二参数。"
                + "题面规定同一行的字段之间必须显式调用 inf.readSpace()，再在行尾调用 inf.readEoln()；"
                + "不要只依赖 readInt 自动跳过空白。\n"
                + problemOverview(problem, 0)
                + "\n标准程序语言：" + safe(standardLanguage)
                + "\n标准程序源码：\n" + limit(standardProgram, 30000);
        return appendCustom(prompt, config.getValidationPrompt(), defaultValidationPrompt());
    }

    public String chatPrompt(ProblemAiConfig config, Problem problem, List<ProblemCase> cases, String question) {
        String rules = appendCustom("根据管理员问题复核题目，并明确给出依据、风险等级和修改建议。",
                config.getValidationPrompt(), defaultValidationPrompt());
        return rules + "\n\n" + problemContext(problem, cases) + "\n管理员问题：" + safe(question);
    }

    public String recheckPrompt(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                                String currentLanguage, String currentStandardProgram,
                                String previousReport, String requirements,
                                String currentCaseVersion, ProblemVerification verification,
                                String currentTestcaseEvidence,
                                String historicalExecutionEvidence) {
        String rules = appendCustom(defaultValidationPrompt(), config.getValidationPrompt(), defaultValidationPrompt());
        return rules + "\n\n你正在复检一份已有的 AI 验题报告。必须以当前题目内容为准，不能默认旧报告仍然正确。"
                + "重点比较管理员修改前后的差异，只针对复检要求和旧报告中提出的风险逐项复核；"
                + "如果当前版本已经修复，明确说明修复依据；如果仍有问题，给出当前版本的证据、风险等级和修改建议。"
                + "这是复检，不要重新生成标准程序，也不要要求重新跑全量测试点；旧报告中的正式判题结果只能作为历史证据，"
                + "不能冒充当前版本的运行结果。若当前题目、标准程序、测试数据内容和测试数据版本都与历史判题绑定一致，"
                + "应明确说明‘未发现管理员修改’，不要仅因复检没有重跑而把结论降为 WARN；只有确实存在版本不一致或缺少关键证据时才标记 WARN。"
                + "caseVersion、测试点模式或测试点总数的变化本身不等于输入/输出内容被修改：应结合当前实际测试数据内容、哈希、"
                + "历史报告中的具体测试点和正式提交绑定判断；仅有动态/样例占位点数量差异时记为 INFO，不要给出‘当前数据未验证’的泛化警告。"
                + "当前测试数据内容已由系统直接读取时，应以这些内容和哈希为准，不得把文件名本身误判为证据缺失。严格只返回 validation JSON 对象。\n\n"
                + problemContext(problem, cases)
                + "\n当前标准程序/测试数据版本信息："
                + "\n当前题目 caseVersion：" + safe(currentCaseVersion)
                + "\n当前正式判题绑定 caseVersion：" + safe(verification == null ? null : verification.getCaseVersion())
                + "\n当前正式判题 submitId：" + safe(verification == null ? null : verification.getSubmitId())
                + "\n当前正式判题状态：" + safe(verification == null ? null : verification.getVerificationStatus())
                + "\n当前测试数据证据：\n" + limit(currentTestcaseEvidence, 12000)
                + "\n历史报告中的执行绑定摘要：\n" + limit(historicalExecutionEvidence, 2000)
                + "\n管理员当前标准程序语言：" + safe(currentLanguage)
                + "\n管理员当前标准程序源码：\n" + limit(currentStandardProgram, 30000)
                + "\n复检要求：\n" + limit(requirements, 4000)
                + "\n历史 AI 验题报告：\n" + limit(previousReport, 30000);
    }

    private String problemContext(Problem p, List<ProblemCase> cases) {
        StringBuilder text = new StringBuilder(problemOverview(p, cases.size()));
        int count = Math.min(cases.size(), MAX_CASES);
        for (int i = 0; i < count; i++) {
            ProblemCase item = cases.get(i);
            text.append("\n测试点 ").append(i + 1).append("（分组 ").append(item.getGroupNum())
                    .append("，分值 ").append(item.getScore()).append("）")
                    .append("\n输入：").append(limit(item.getInput(), MAX_CASE_TEXT))
                    .append("\n标准输出：").append(limit(item.getOutput(), MAX_CASE_TEXT));
        }
        if (cases.size() > count) text.append("\n其余 ").append(cases.size() - count).append(" 个测试点内容未在当前消息展开。");
        return text.toString();
    }

    private String problemOverview(Problem p, int caseCount) {
        StringBuilder text = new StringBuilder("题目信息：")
                .append("\n内部ID：").append(p.getId())
                .append("\n展示ID：").append(safe(p.getProblemId()))
                .append("\n标题：").append(safe(p.getTitle()))
                .append("\n类型：").append(p.getType())
                .append("\n判题模式：").append(safe(p.getJudgeMode()))
                .append("\n测试点模式：").append(safe(p.getJudgeCaseMode()))
                .append("\n时间限制(ms)：").append(p.getTimeLimit())
                .append("\n内存限制(MB)：").append(p.getMemoryLimit())
                .append("\n题面：").append(limit(p.getDescription(), 10000))
                .append("\n输入说明：").append(limit(p.getInput(), 5000))
                .append("\n输出说明：").append(limit(p.getOutput(), 5000))
                .append("\n样例：").append(limit(p.getExamples(), 8000))
                .append("\n测试点总数：").append(caseCount);
        if (!"default".equals(p.getJudgeMode())) {
            text.append("\n判题程序语言：").append(safe(p.getSpjLanguage()))
                    .append("\n判题程序源码：\n").append(limit(p.getSpjCode(), 12000));
        }
        return text.toString();
    }

    private String programOverview(Problem p) {
        StringBuilder text = new StringBuilder("题目定义：")
                .append("\n标题：").append(safe(p.getTitle()))
                .append("\n类型：").append(p.getType())
                .append("\n判题模式：").append(safe(p.getJudgeMode()))
                .append("\n时间限制(ms)：").append(p.getTimeLimit())
                .append("\n内存限制(MB)：").append(p.getMemoryLimit())
                .append("\n题面：").append(limit(p.getDescription(), 8000))
                .append("\n输入说明：").append(limit(p.getInput(), 4000))
                .append("\n输出说明：").append(limit(p.getOutput(), 4000))
                .append("\n样例：").append(limit(p.getExamples(), 6000));
        if (!"default".equals(p.getJudgeMode())) {
            text.append("\n判题程序语言：").append(safe(p.getSpjLanguage()))
                    .append("\n判题程序源码：\n").append(limit(p.getSpjCode(), 8000));
        }
        return text.toString();
    }

    private String appendCustom(String builtIn, String custom, String defaultValue) {
        if (StringUtils.isEmpty(custom) || defaultValue.equals(custom.trim())) return builtIn;
        return builtIn + "\n\n管理员补充规则：\n" + custom.trim();
    }

    private String safe(Object value) { return value == null ? "未提供" : String.valueOf(value); }

    private String limit(String value, int max) {
        if (StringUtils.isEmpty(value)) return "未提供";
        String text = value.trim();
        return text.length() <= max ? text : text.substring(0, max) + "\n[内容已截断]";
    }
}
