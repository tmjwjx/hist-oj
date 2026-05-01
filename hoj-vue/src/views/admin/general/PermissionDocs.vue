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
      markdownContent: `| 权限项 | 超级管理员 | 题目管理员 | 普通管理员 |
| ---- | ---- | ---- | ---- |
| 比赛答疑 | ✔ | ❌ | ❌ |
| 比赛代码查重 | ✔ | ❌ | ❌ |
| Rating 管理 | ✔ | ❌ | ❌ |
| 赛事报名系统管理 | ✔ | ✔ | ✔ |
| 班级管理 | ✔ | ❌ | ❌ |
| 客观题题库管理 | ✔ | ❌ | ❌ |
| 试卷库管理 | ✔ | ✔ | ✔ |
| 算法航海图管理 | ✔ | ✔ | ✔ |
| 设置教师权限 | ✔ | ❌ | ❌ |
| 设置学生权限 | ✔ | ✔ | ✔ |
| 判题终端 | ✔ | ✔ | ❌ |
| 对战记录查询 | ✔ | ❌ | ❌ |
| 系统公告管理 | ✔ | ❌ | ❌ |
| 系统通知推送管理 | ✔ | ❌ | ❌ |
| 系统配置 | ✔ | ❌ | ❌ |
| 用户管理 | ✔ | ❌ | ❌ |
| 全部题目增加 | ✔ | ✔ | ✔ |
| 其他人创建的题目查看 | ✔ | ✔ | ❌ |
| 自己创建的题目查看 | ✔ | ✔ | ✔ |
| 其他人创建的题目修改 | ✔ | ✔ | ❌ |
| 自己创建的题目修改 | ✔ | ✔ | ✔ |
| 全部题目删除 | ✔ | ✔ | ❌ |
| 全部题目权限修改（公开、隐藏、比赛） | ✔ | ✔ | ❌ |
| 全部题目评测数据下载 | ✔ | ✔ | ❌ |
| 导入远程OJ题目 | ✔ | ✔ | ✔ |
| 全部比赛权限(增加、删除、修改) | ✔ | ❌ | ❌ |
| 自己创建的比赛（增加、修改） | ✔ | ✔ | ✔ |
| 自己创建的比赛的题目（查看、增加，修改，移除） | ✔ | ✔ | ✔ |
| 其他人创建的比赛的题目（查看、增加，修改，移除） | ✔ | ❌ | ❌ |
| 自己创建的比赛的题目（评测数据下载、删除） | ✔ | ✔ | ❌ |
| 自己创建的比赛的题目权限修改（隐藏、删除） | ✔ | ✔ | ✔ |
| 自己创建的比赛的题目权限修改为公开题目 | ✔ | ✔ | ❌ |
| 讨论管理 | ✔ | ✔ | ✔ |
| 全部训练权限(增加、删除、修改) | ✔ | ❌ | ❌ |
| 自己创建的训练（增加、修改） | ✔ | ✔ | ✔ |`
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
