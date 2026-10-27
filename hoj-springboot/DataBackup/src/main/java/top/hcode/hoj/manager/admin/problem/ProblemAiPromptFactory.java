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
                + "对具体正式测试点必须以正式判题结果为第一证据，并检查标准程序是否按题面输入格式执行 testlib 的"
                + "registerValidation 强检验（仅普通/SPJ题；交互题逐轮校验协议回复，不使用静态输入验证器）：Accepted 且无 Runtime Error（RE）只能标记 PASS，不能因为文件内容未展开、"
                + "只显示文件名、内容截断或 stderr 为空而标记 WARN；明确 RE、非 Accepted、未执行或一体化标准程序校验失败才标记 FAIL。"
                + "证据不足导致的 WARN 仅适用于无法由正式测试点结果覆盖的题面或程序分析。"
                + "标准程序 Accepted 不等于 SPJ/交互判题程序正确；判题程序必须独立审查接受合法解、拒绝错误解和协议终止行为。";
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
                + "普通/SPJ正式测试点必须按输入说明写出 testlib/registerValidation 风格的严格检查思路；交互题只能逐轮核对交互回复和协议，"
                + "但不得把无法展开测试点文件当成测试点 WARN；正式状态 Accepted 且无 RE 即视为该点 PASS。"
                + "最终 PASS 仅允许在正式判题全部测试点 Accepted 且没有确定性题面、样例、数据或判题程序错误时给出。";
    }

    public String systemPrompt(String custom) {
        return systemPrompt(custom, null);
    }

    public String systemPrompt(String custom, Problem problem) {
        return appendCustom(defaultSystemPrompt(), custom, defaultSystemPrompt())
                + "\n当前执行约束：" + (isInteractive(problem)
                ? "当前为交互题，std 是协议客户端，只能逐轮校验实际收到的回复范围、语义与终止信号，不能读取隐藏输入，不能调用 registerValidation 或 inf.readEof 等等待静态输入结束的接口。"
                : "普通/SPJ题的输入格式和范围校验必须由主模型 std 自身执行；若用户指定交互题，则改为逐轮校验交互回复，不能使用 registerValidation 或 inf.readEof。")
                + "不存在独立 validator 的提交或运行。"
                + "源码结构检查与正式判题 Accepted 都不能证明全部约束已覆盖，必须逐项审查实际读取及范围/语义检查逻辑；"
                + "发现缺失约束时保留风险或失败结论，不能由通过测试点覆盖掉。";
    }

    public String programSystemPrompt(String custom) {
        return programSystemPrompt(custom, null);
    }

    public String programSystemPrompt(String custom, Problem problem) {
        String builtIn = "你是 HistOJ 的 ACM/OI 标准程序生成器。只根据题目定义独立推导算法，"
                + "在内部比较至少三类候选方向并复核正确性、复杂度、边界、溢出和语言兼容性。"
                + "严格按用户指定的源码类型返回一个 JSON 对象，只生成一份完整标准程序，不生成独立 validator。"
                + "源码必须计算题目要求的答案并输出到 stdout。algorithm、warnings 和 difficulty.reason 使用简体中文。"
                + "warnings 仅记录源码本身的真实风险，不因生成阶段尚未运行而虚构验证结论或泛化警告。";
        return appendCustom(builtIn, custom, defaultSystemPrompt())
                + "\n当前生成约束优先：只生成用户本次指定的一份求解源码。" + (isInteractive(problem)
                ? "当前为交互题，主模型 std 生成普通交互协议客户端，逐轮校验回复、及时 flush，遵守查询上限和终止信号；不能使用 registerValidation 或 inf.readEof，不能读取隐藏输入。"
                : "普通/SPJ题的主模型 std 内置 testlib 严格校验；若用户指定交互题，生成普通协议客户端并逐轮校验回复，禁止 registerValidation 和 inf.readEof，")
                + "辅助模型和多语言实现使用普通求解代码；均须输出题目答案。禁止生成分离校验器或 validator 字段。";
    }

    public String validationPrompt(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                                   ProblemAiValidateDTO input,
                                   ProblemAiValidationContext.Snapshot snapshot) {
        String rules = appendCustom(defaultValidationPrompt(), config.getValidationPrompt(), defaultValidationPrompt());
        return rules + "\n\n" + problemOverview(problem, snapshot.getPoints().size())
                + "\n标准程序语言：" + safe(input.getLanguage())
                + "\n标准程序源码：\n" + limit(input.getStandardProgram(), 30000)
                + "\n" + (isInteractive(problem)
                ? "交互 std 必须逐轮校验实际收到的回复和终止信号，检查查询计数、flush 与提前 EOF；禁止读取隐藏输入或使用 registerValidation/inf.readEof 等静态文件验证方式。"
                : "输入校验与求解在同一份标准程序中执行；必须检查 registerValidation、严格读取和语义约束是否完整。")
                + "\n正式判题提交 ID：" + snapshot.getSubmitId()
                + "\n正式判题最终状态：" + snapshot.getStatusText()
                + "\n正式判题错误摘要：" + safe(snapshot.getJudgeError())
                + "\n测试点总数：" + snapshot.getPoints().size()
                + "\n逐测试点会由系统后续批次完整提供并合并到最终结果；本阶段只分析题面、样例、程序和判题模式，"
                + "不得因为当前消息未展开逐点内容而判定测试点证据缺失。"
                + judgeProgramReviewPrompt(problem);
    }

    public String standardProgramPrompt(ProblemAiConfig config, Problem problem, String language, int caseCount) {
        if (isInteractive(problem)) {
            return "请独立生成一份普通 C++17 交互 std.cpp，作为参赛程序与交互判题程序通信。"
                    + "只能读取题面协议允许发送给选手的初始信息和逐轮回复，不能读取测试文件、隐藏答案、inf 或 ans。"
                    + "禁止 registerValidation、inf.readEof 或等待完整输入结束；不得把交互记录或隐藏输入当作静态数据求解。"
                    + "读取每次回复后立即校验范围、语义、回复类型与终止信号；收到错误信号、提前 EOF 或非法回复应停止通信并非零退出，诊断写入 stderr。"
                    + "使用普通标准输入输出；每次查询和最终答案后及时 flush，显式控制查询上限，不得在协议结束后继续读写。"
                    + "内部独立推导算法并验证正确性、最坏查询次数、复杂度、溢出和边界；不得臆造协议或硬编码样例。"
                    + "只生成一份求解源码，不生成独立 validator。严格只返回 JSON："
                    + "{\"language\":\"C++ 17\",\"code\":\"完整交互标准程序\","
                    + "\"algorithm\":\"算法、正确性、复杂度、查询上限与逐轮回复校验摘要\",\"warnings\":\"真实风险，无则空字符串\","
                    + difficultySchema() + "}。\n" + difficultyRules() + "\n" + programOverview(problem);
        }
        return "请独立生成一份 C++17 std.cpp，将 testlib 严格输入校验与标准答案计算写在同一程序中。"
                + "不能只校验不输出答案，也不能生成独立 validator 或额外源码。禁止硬编码样例。"
                + "必须 #include \"testlib.h\"，入口 int main(int argc, char* argv[])，最开始调用 registerValidation(argc, argv);。"
                + "所有输入只能使用 inf.readInt/readLong/readToken/readString/readChar/readSpace/readEoln 等 testlib 接口，"
                + "禁止 cin、scanf 或重新读取输入；相邻字段显式 readSpace()、行末 readEoln()，末尾必须 inf.readEof()。"
                + "每个数值范围、字符串长度/字符集、数组长度、T、下标和跨组总和限制都显式校验；"
                + "使用 ensuref 校验题面承诺的互异、排序、排列、无自环/重边、连通性、DAG、区间合法、答案存在等语义条件。"
                + "不得臆造题面没有的限制；格式不明确时采用同一行 token 间单个空格、行尾换行并在 warnings 说明假设。"
                + "readToken(pattern,name) 第一参数是正则，不是变量名；禁止 inf.readToken(\"row\") 这类误用。"
                + "报错须定位测试编号、行或元素；合法输入计算后仅向 stdout 输出规定答案，最后 return 0，禁止 quitf(_ok,...)。"
                + "非法格式或语义约束必须通过 testlib 报错到 stderr 并非零退出。避免溢出、越界、递归爆栈和不匹配数据范围的复杂度。"
                + "SPJ 输出任意合法解。"
                + "严格只返回 JSON：{\"language\":\"C++ 17\",\"code\":\"完整一体化 std.cpp\","
                + "\"algorithm\":\"算法、正确性、复杂度与严格校验摘要\",\"warnings\":\"真实风险，无则空字符串\","
                + difficultySchema() + "}。\n" + difficultyRules() + "\n" + programOverview(problem);
    }

    public String auxiliaryProgramPrompt(ProblemAiConfig config, Problem problem) {
        return "请仅根据以下题目独立推导并编写普通 C++17 标准程序，不使用其他模型的代码、结论或难度估计。"
                + "采用标准输入输出，不需要 testlib 或 validator；实现完整题意，禁止硬编码样例。"
                + "内部比较至少三类候选方向，证明正确性并检查复杂度、溢出、边界；SPJ/交互题遵循对应协议。"
                + (isInteractive(problem) ? interactiveClientRules() : "")
                + "严格只返回 JSON：{\"language\":\"C++ 17\",\"code\":\"完整普通标准程序\","
                + "\"algorithm\":\"算法、正确性与复杂度摘要\",\"warnings\":\"真实风险，无则空字符串\","
                + difficultySchema() + "}。所有自然语言使用简体中文。\n"
                + difficultyRules() + "\n" + programOverview(problem);
    }

    private String difficultySchema() {
        return "\"difficulty\":{\"rating\":1600,\"rangeLow\":1500,\"rangeHigh\":1700,"
                + "\"confidence\":\"high|medium|low\",\"reason\":\"建模、算法、实现、边界与修正理由\"}";
    }

    private String difficultyRules() {
        return "Codeforces 难度必须独立估计，不按历史通过人数或冷门程度线性映射。"
                + "先识别模型，M建模/A算法/I实现/E边界各0到5，O晦涩0到3；"
                + "Raw=800+180*M+220*A+120*I+100*E+80*O。经典裸模板或小数据暴力减100到300；"
                + "深藏观察或多算法组合加100到300；纯格式复杂最多加100到200，冷门只降低置信度。"
                + "普通 BFS、最短路、MST、LCS、背包、区间DP裸题一般不超过1600到1800；"
                + "隐藏建模、复杂状态、强证明或多算法组合才考虑1900以上。"
                + "最终 rating 四舍五入到100的倍数，给合理 rangeLow/rangeHigh 且包含 rating；"
                + "confidence 取 high/medium/low，题面或范围不完整必须 low；reason 说明 M/A/I/E/O、修正和核心瓶颈。";
    }

    public String multiLanguageProgramPrompt(ProblemAiConfig config, Problem problem,
                                             String language, String canonicalProgram) {
        String prompt = "请把下面题目的标准程序独立翻译成指定语言，生成可直接提交到 HistOJ 判题机的完整源码。"
                + "保持题意、输入输出格式、复杂度和整数语义完全一致，禁止硬编码样例，禁止依赖第三方库。"
                + "参考 C++17 中的 testlib 输入校验不需要翻译；生成普通标准输入输出求解程序，保留答案逻辑与边界处理。"
                + (isInteractive(problem) ? interactiveClientRules() : "")
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
                .append("stderr 为空本身不是错误，但必须明确记录已检查。对具体正式测试点，")
                .append(isInteractive(problem)
                        ? "只能校验交互协议中实际收到的回复和终止信号；不能要求 std 使用 registerValidation、inf.readEof 或读取隐藏输入；"
                        : "输入格式检查必须采用 testlib/registerValidation 的严格规则；")
                .append("正式判题 Accepted 且无 RE 时必须给 PASS，")
                .append("即使输入/输出只保留文件名、内容未展开或被截断，也禁止给 WARN。")
                .append("任一 judgeStatus 非 Accepted、Runtime Error（RE）或 executed=false 必须给 FAIL；不要用 WARN 代替 FAIL。\n")
                .append("严格只返回 JSON：{\"testPointResults\":[{\"index\":1,\"status\":\"PASS|WARN|FAIL\",\"")
                .append("detail\":\"输入/输出/运行结果/stderr 的具体检查过程与结论\"}]}。\n")
                .append(problemOverview(problem, total))
                .append("\n标准程序语言：").append(safe(input.getLanguage()))
                .append("\n标准程序源码：\n").append(limit(input.getStandardProgram(), 30000))
                .append(isInteractive(problem)
                        ? "\n本次 std 是交互协议客户端，只能逐轮校验回复，以下正式状态和 stderr 不能单独证明交互判题程序没有漏判。"
                        : "\n严格输入校验由这份标准程序自身执行，以下状态和 stderr 均来自同一次正式判题，不存在独立 validator 执行。");
        for (ProblemAiTestPoint point : points) {
            text.append("\n\n测试点 ").append(point.getIndex())
                    .append("\ncaseId：").append(safe(point.getCaseId()))
                    .append("\n分组：").append(safe(point.getGroupNum()))
                    .append("，分值：").append(safe(point.getScore()))
                    .append("\n是否执行：").append(point.getExecuted())
                    .append("\n正式判题状态：").append(safe(point.getJudgeStatusText()))
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
        appendJudgeProgram(text, p);
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
        appendJudgeProgram(text, p);
        return text.toString();
    }

    private boolean isInteractive(Problem problem) {
        return problem != null && "interactive".equals(problem.getJudgeMode());
    }

    private void appendJudgeProgram(StringBuilder text, Problem problem) {
        if (!"spj".equals(problem.getJudgeMode()) && !isInteractive(problem)) return;
        text.append("\n判题程序语言：").append(safe(problem.getSpjLanguage()))
                .append("\n判题程序完整源码：\n").append(safe(problem.getSpjCode()))
                .append("\n判题程序额外文件（JSON，文件名到完整内容）：\n").append(safe(problem.getJudgeExtraFile()))
                .append("\n必须联合审查上述源码与全部额外文件；缺失文件、内容超过模型可审查范围或无法确认关键路径时标记 WARN，不能宣称已完整审查。");
    }

    private String interactiveClientRules() {
        return "当前交互程序只能读取协议发给选手的信息，不能读取隐藏测试输入或答案；每轮校验回复和终止信号，"
                + "遵守查询上限，每次输出立即 flush，正确处理提前 EOF，不能等待输入全部结束或使用 registerValidation/inf.readEof。";
    }

    private String judgeProgramReviewPrompt(Problem problem) {
        boolean interactive = isInteractive(problem);
        if (!interactive && !"spj".equals(problem.getJudgeMode())) return "";
        return "\n\n必须在本次 validation JSON 顶层额外返回独立判题程序审查，不得以 std Accepted 替代："
                + "\"judgeProgramReview\":{\"status\":\"PASS|WARN|FAIL\",\"summary\":\"审查结论与证据范围\","
                + "\"checks\":[{\"id\":\"规定检查ID\",\"status\":\"PASS|WARN|FAIL\",\"detail\":\"源码位置、条件和推理证据\"}],"
                + "\"probes\":[{\"id\":\"规定探针ID\",\"category\":\"与id相同\",\"expectation\":\"ACCEPT|REJECT\","
                + "\"code\":\"完整独立C++17参赛程序\",\"reason\":\"题意依据、如何覆盖所有测试点以及预期结果理由\"}]}。"
                + "自然语言 summary、detail、reason 使用简体中文。不得遗漏或重复检查ID；每项必须引用判题程序或额外文件中的具体逻辑。"
                + (interactive
                ? "checks 必须且只包含 protocol、limits、termination、verdict 四项。"
                    + "protocol：核对 registerInteraction 与 argv/inf/ouf/ans 的对应关系、查询与回复格式、每次输出 flush、双方读写顺序、死锁风险及隐藏输入/答案泄露。"
                    + "limits：核对每轮参数范围、语义约束、查询次数上限及计数时机、越界、溢出和不合法查询处理。"
                    + "termination：核对提前 EOF、错误信号、选手提前退出、非法或多余命令、最终答案后继续输出与结束握手，不能因等待 EOF 或未 flush 死锁。"
                    + "verdict：核对正确答案接受、错误答案拒绝、各退出分支和退出码，不能把选手异常终止或交互器异常当作正确答案。"
                    + "probes 最多两份，完整覆盖时必须是 protocol_violation（expectation=REJECT，违反明确协议规则但不靠崩溃/无限循环）"
                    + "和 premature_exit（expectation=REJECT，读完必要初始信息后在必须完成的协议之前正常退出）。"
                : "checks 必须且只包含 acceptance、rejection、numeric、termination 四项。"
                    + "acceptance：核对 registerTestlibCmd 与 argv/inf/ouf/ans 对应关系，合法输出及合法多解应接受，不能错误地要求与 ans 文本相等。"
                    + "rejection：核对错误答案、不完整输出、违反构造约束或答案条件能被拒绝，不能信任选手自报值或漏读关键字段。"
                    + "numeric：核对 NaN/Inf、绝对/相对误差与临界容差、整数溢出、类型转换、数组越界、范围限制及畸形 token；不适用项说明源码依据。"
                    + "termination：核对退出码、quitf 路径、提前 EOF、无效或多余尾 token 以及检查失败后仍接受的分支；空白/尾 token 的处理须符合题面语义。"
                    + "probes 最多两份，完整覆盖时必须是 alternate_valid（expectation=ACCEPT，独立计算合法答案，有多解时尽量构造不同于参考答案的合法解）"
                    + "和 invalid_output（expectation=REJECT，输出对每个合法测试输入都确定违反题意的错误答案）。")
                + "所有 probe 都是提交到现有正式判题流程的候选参赛程序，不是替换 SPJ/交互器；禁止文件、网络、进程和系统调用，"
                + "禁止复制隐藏答案、故意崩溃、耗尽资源或无限循环。每份必须能按当前题目协议正常编译运行，只用标准输入输出。"
                + "ACCEPT 必须期望所有测试点被判题程序接受；REJECT 必须期望所有测试点都被判题程序拒绝，不能把编译错误、运行异常、超时或系统错误当作拒错成功。"
                + "必须逐份说明对所有合法测试输入成立的题意依据。无法可靠构造某类 probe 时，probes 返回能可靠构造的部分或空数组，"
                + "status 至少 WARN 并在 summary 说明缺失类别和原因，禁止编造代码或声称动态验证已完成。"
                + "本阶段只能给静态审查结论与待运行 probes；系统会执行 probes 并合并真实结果。发现确定性判题漏洞必须 FAIL，证据不足必须 WARN。";
    }

    private String appendCustom(String builtIn, String custom, String defaultValue) {
        String legacySystem = defaultSystemPrompt()
                .replace("并检查标准程序是否按题面输入格式执行 testlib 的", "并按题面输入格式执行等价于 testlib 的")
                .replace("一体化标准程序校验失败", "验证器失败");
        if (StringUtils.isEmpty(custom) || defaultValue.equals(custom.trim())
                || legacySystem.equals(custom.trim())) return builtIn;
        return "管理员补充规则：\n" + custom.trim()
                + "\n\n本次内置规则（与补充规则冲突时以本段为准）：\n" + builtIn;
    }

    private String safe(Object value) { return value == null ? "未提供" : String.valueOf(value); }

    private String limit(String value, int max) {
        if (StringUtils.isEmpty(value)) return "未提供";
        String text = value.trim();
        return text.length() <= max ? text : text.substring(0, max) + "\n[内容已截断]";
    }
}
