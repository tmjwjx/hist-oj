package top.hcode.hoj.manager.admin.problem;

import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;

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
                + "证据不足或内容被截断时标记 WARN，确定错误或任一正式测试点未通过时标记 FAIL。";
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
                + "最终 PASS 仅允许在正式判题全部测试点 Accepted 且没有确定性题面、样例、数据或判题程序错误时给出。";
    }

    public String systemPrompt(String custom) {
        return appendCustom(defaultSystemPrompt(), custom, defaultSystemPrompt());
    }

    public String validationPrompt(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                                   ProblemAiValidateDTO input,
                                   ProblemAiValidationContext.Snapshot snapshot) {
        String rules = appendCustom(defaultValidationPrompt(), config.getValidationPrompt(), defaultValidationPrompt());
        return rules + "\n\n" + problemOverview(problem, snapshot.getPoints().size())
                + "\n标准程序语言：" + safe(input.getLanguage())
                + "\n标准程序源码：\n" + limit(input.getStandardProgram(), 30000)
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
                + "\"algorithm\":\"算法、正确性与复杂度摘要\",\"warnings\":\"仍需人工注意的事项，没有则为空\"}。\n"
                + problemOverview(problem, caseCount) + "\n指定语言：" + safe(language);
        return appendCustom(prompt, config.getValidationPrompt(), defaultValidationPrompt());
    }

    public String testPointPrompt(ProblemAiConfig config, Problem problem, ProblemAiValidateDTO input,
                                  List<ProblemAiTestPoint> points, int total) {
        StringBuilder text = new StringBuilder()
                .append("你正在执行全量验题的逐测试点阶段。本批测试点属于总计 ").append(total).append(" 个测试点，")
                .append("必须对本批每一个 index 返回且不得遗漏。逐点核对输入与题面约束、标准输出、正式判题状态、资源使用和 stderr；")
                .append("stderr 为空本身不是错误，但必须明确记录已检查。若内容标记为截断，必须给 WARN，不能假定未展示部分正确。")
                .append("任一 judgeStatus 非 Accepted 或 executed=false 必须给 FAIL。\n")
                .append("严格只返回 JSON：{\"testPointResults\":[{\"index\":1,\"status\":\"PASS|WARN|FAIL\",\"")
                .append("detail\":\"输入/输出/运行结果/stderr 的具体检查过程与结论\"}]}。\n")
                .append(problemOverview(problem, total))
                .append("\n标准程序语言：").append(safe(input.getLanguage()))
                .append("\n标准程序源码：\n").append(limit(input.getStandardProgram(), 30000));
        for (ProblemAiTestPoint point : points) {
            text.append("\n\n测试点 ").append(point.getIndex())
                    .append("\ncaseId：").append(safe(point.getCaseId()))
                    .append("\n分组：").append(safe(point.getGroupNum()))
                    .append("，分值：").append(safe(point.getScore()))
                    .append("\n是否执行：").append(point.getExecuted())
                    .append("\n正式判题状态：").append(safe(point.getJudgeStatusText()))
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
        if (cases.size() > count) text.append("\n其余 ").append(cases.size() - count).append(" 个测试点请使用一键 AI 验题逐点检查。");
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
