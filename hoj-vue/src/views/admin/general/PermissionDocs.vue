<template>
  <div class="permission-docs-container">
    <el-card class="docs-card">
      <div slot="header" class="docs-header">
        <i class="fa fa-key" style="margin-right: 10px;"></i>
        <span>管理员权限说明</span>
      </div>
      <div class="markdown-content" v-html="renderedMarkdown"></div>
    </el-card>
  </div>
</template>

<script>
import { marked } from 'marked';

// 配置 marked 选项
marked.setOptions({
  breaks: true,
  gfm: true
});

export default {
  name: 'PermissionDocs',
  mounted() {
    // 设置页面标题为"权限说明",避免显示国际化key
    document.title = '权限说明';
  },
  data() {
    return {
      markdownContent: `#### 通用管理

| 权限项 | 超级管理员 | 题目管理员 | 普通管理员 |
| ---- | ---- | ---- | ---- |
| 仪表盘 | ✔ | ✔ | ✔ |
| 用户管理 | ✔ | ❌ | ❌ |
| 系统公告管理 | ✔ | ❌ | ❌ |
| 系统通知推送管理 | ✔ | ❌ | ❌ |
| 系统配置 | ✔ | ❌ | ❌ |
| 系统开关管理 | ✔ | ❌ | ❌ |
| 全局 Rejudge | ✔ | ❌ | ❌ |
| Rating 管理 | ✔ | ❌ | ❌ |
| 对战记录查询 | ✔ | ❌ | ❌ |

#### 题库与题目

| 权限项 | 超级管理员 | 题目管理员 | 普通管理员 |
| ---- | ---- | ---- | ---- |
| 全部题目增加 | ✔ | ✔ | ✔ |
| 自己创建的题目查看 / 修改 | ✔ | ✔ | ✔ |
| 其他人创建的题目查看 / 修改 | ✔ | ✔ | ❌ |
| 全部题目删除 | ✔ | ✔ | ❌ |
| 题目权限修改（隐藏、比赛） | ✔ | ✔ | ✔ |
| 题目权限修改（公开） | ✔ | ✔ | ❌ |
| 全部题目评测数据下载 | ✔ | ✔ | ❌（题目作者可） |
| 导入远程 OJ 题目 | ✔ | ✔ | ✔ |
| FPS / QDUOJ / Hydro 题目包导入导出 | ✔ | ❌ | ❌ |
| 标签管理 | ✔ | ✔ | ❌ |
| 团队题目审批 | ✔ | ✔ | ❌ |
| AI 验题与 AI 助手 | ✔ | ✔ | ✔ |
| 客观题题库管理 | ✔ | ❌ | ❌ |
| 试卷库管理 | ✔ | ✔ | ✔ |
| 算法航海图管理 | ✔ | ✔ | ✔ |

#### 比赛管理

| 权限项 | 超级管理员 | 题目管理员 | 普通管理员 |
| ---- | ---- | ---- | ---- |
| 比赛删除 | ✔ | ❌ | ❌ |
| 比赛增加 / 修改自己创建的比赛 | ✔ | ✔ | ✔ |
| 修改他人创建的比赛 | ✔ | ❌ | ❌ |
| 比赛答疑（查看 / 答复） | ✔ | 仅自己创建的比赛 | 仅自己创建的比赛 |
| 比赛代码查重 / 设备异常检测 | ✔ | 仅自己创建的比赛 | 仅自己创建的比赛 |
| 赛事报名管理（查看 / 修改报名） | ✔ | 仅自己创建的比赛 | 仅自己创建的比赛 |
| 比赛内 Rejudge | ✔ | ❌ | ❌ |
| 比赛工具（打印管理、AC 信息确认） | ✔ | 仅自己创建的比赛 | 仅自己创建的比赛 |
| 自己创建的比赛的题目（查看、增加、修改） | ✔ | ✔ | ✔ |
| 他人创建的比赛的题目（查看、修改） | ✔ | ✔ | ❌（题目作者可） |
| 比赛题目移除 / 删除 | ✔ | ✔ | ❌ |
| 比赛题目评测数据下载 | ✔ | ✔ | ❌（题目作者可） |
| 比赛题目权限修改（隐藏、比赛） | ✔ | ✔ | ✔ |
| 比赛题目权限修改（公开） | ✔ | ✔ | ❌ |
| 比赛公告管理 | ✔ | ✔ | ✔ |

#### 训练与讨论

| 权限项 | 超级管理员 | 题目管理员 | 普通管理员 |
| ---- | ---- | ---- | ---- |
| 训练删除 | ✔ | ❌ | ❌ |
| 训练增加 / 修改自己创建的训练 | ✔ | ✔ | ✔ |
| 修改他人创建的训练 | ✔ | ❌ | ❌ |
| 训练题目更新 / 移除 | ✔ | ✔ | ❌ |
| 训练参与者查看 | ✔ | 仅自己创建的训练 | ✔ |
| 讨论管理 | ✔ | ✔ | ✔ |

#### 班级与教学

| 权限项 | 超级管理员 | 题目管理员 | 普通管理员 |
| ---- | ---- | ---- | ---- |
| 班级管理（后台） | ✔ | ❌ | ❌ |
| 设置教师权限 | ✔ | ❌ | ❌ |
| 设置学生权限 | ✔ | ❌ | ✔ |
| 作业学情分析 / 考试监控（后台） | ✔ | ❌ | ❌（班级教师可在班级内使用） |

#### 说明

- 「仅自己创建的」表示该角色仅当本人是该资源的创建者时才拥有此权限。
- 教师 / 学生是班级内部角色，与系统管理员角色（超级管理员、题目管理员、普通管理员）相互独立。
- 「（题目作者可）」表示除表格标注的角色外，该题目的创建者本人始终拥有此权限。
- 个别接口后端的放行范围略宽于页面入口的可见性控制，实际以管理后台的菜单 / 按钮可见性为准。`
    };
  },
  computed: {
    renderedMarkdown() {
      try {
        return marked(this.markdownContent);
      } catch (error) {
        console.error('Markdown渲染失败:', error);
        return this.markdownContent;
      }
    }
  }
};
</script>

<style scoped>
.permission-docs-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.docs-card {
  border-radius: 8px;
}

.docs-header {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
}

.markdown-content {
  padding: 20px;
  line-height: 1.8;
}

.markdown-content >>> h4 {
  margin: 28px 0 10px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.markdown-content >>> ul {
  padding-left: 20px;
  color: #606266;
  font-size: 14px;
}

.markdown-content >>> li {
  margin: 4px 0;
}

.markdown-content >>> table {
  width: 100%;
  border-collapse: collapse;
  margin: 20px 0;
  font-size: 14px;
}

.markdown-content >>> table th {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
  text-align: left;
  padding: 12px;
  border: 1px solid #ebeef5;
}

.markdown-content >>> table td {
  padding: 12px;
  border: 1px solid #ebeef5;
  color: #606266;
}

.markdown-content >>> table tr:hover {
  background-color: #f5f7fa;
}

.markdown-content >>> table td:first-child,
.markdown-content >>> table th:first-child {
  font-weight: 600;
  color: #303133;
}
</style>
