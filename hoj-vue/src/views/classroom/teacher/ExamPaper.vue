<template>
  <div>
    <!-- 编程题详情对话框 - 移到最外层 -->
    <el-dialog
      title="编程题详情"
      :visible.sync="showProblemDetailDialog"
      width="900px"
      :close-on-click-modal="false"
      class="problem-detail-dialog-wrapper"
      append-to-body
    >
      <div v-if="fetchingViewProblem" v-loading="fetchingViewProblem" style="min-height: 200px;"></div>
      <div v-else-if="currentViewProblem && currentViewProblem.problem">
        <div class="problem-detail-view">
          <h3 style="margin-top: 0;" v-html="renderMarkdown(currentViewProblem.problem.title)"></h3>

          <div class="problem-meta-info" style="margin: 15px 0;">
            <el-tag size="small" type="info">
              <span v-html="renderMarkdown('**时间限制:** ' + currentViewProblem.problem.timeLimit + 'ms')"></span>
            </el-tag>
            <el-tag size="small" type="warning">
              <span v-html="renderMarkdown('**内存限制:** ' + currentViewProblem.problem.memoryLimit + 'MB')"></span>
            </el-tag>
            <el-tag size="small" type="primary">
              <span v-html="renderMarkdown('**判题模式:** ' + getJudgeModeText(currentViewProblem.problem.judgeMode))"></span>
            </el-tag>
            <el-tag size="small" type="success">
              <span v-html="renderMarkdown('**难度:** ' + (currentViewProblem.problem.difficulty || '未知'))"></span>
            </el-tag>
          </div>

          <el-divider></el-divider>

          <div class="problem-section">
            <h4>题目描述</h4>
            <div class="problem-description" v-html="renderMarkdown(currentViewProblem.problem.description)"></div>
          </div>

          <div class="problem-section" v-if="currentViewProblem.problem.input">
            <h4>输入格式</h4>
            <div class="problem-io" v-html="renderMarkdown(currentViewProblem.problem.input)"></div>
          </div>

          <div class="problem-section" v-if="currentViewProblem.problem.output">
            <h4>输出格式</h4>
            <div class="problem-io" v-html="renderMarkdown(currentViewProblem.problem.output)"></div>
          </div>

          <div class="problem-section" v-if="currentViewProblem.problem.hint">
            <h4>提示</h4>
            <div class="problem-hint" v-html="renderMarkdown(currentViewProblem.problem.hint)"></div>
          </div>

          <div class="problem-section" v-if="parseExamples(currentViewProblem.problem.examples).length > 0">
            <h4>样例</h4>
            <div v-for="(example, index) in parseExamples(currentViewProblem.problem.examples)" :key="index" class="problem-example">
              <div class="example-item">
                <strong v-html="renderMarkdown('样例 ' + (index + 1))"></strong>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown('**输入：**')"></div>
                  <div class="example-content" v-html="renderMarkdown(example.input)"></div>
                </div>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown('**输出：**')"></div>
                  <div class="example-content" v-html="renderMarkdown(example.output)"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="problem-empty">
        <i class="el-icon-info"></i>
        <span>暂无题目信息</span>
      </div>

      <span slot="footer">
        <el-button @click="showProblemDetailDialog = false">关闭</el-button>
      </span>
    </el-dialog>

    <!-- 标签题目详情对话框 -->
    <el-dialog
      title="题目详情"
      :visible.sync="showTagProblemDetailDialog"
      width="900px"
      :close-on-click-modal="false"
      append-to-body
    >
      <div v-if="loadingTagProblemDetail" v-loading="true" style="min-height: 200px;"></div>
      <div v-else-if="tagProblemDetail">
        <div class="problem-detail-view">
          <h3 style="margin-top: 0;" v-html="renderMarkdown(tagProblemDetail.problem.title)"></h3>

          <div class="problem-meta-info" style="margin: 15px 0;">
            <el-tag size="small" type="info">
              <span v-html="renderMarkdown('**时间限制:** ' + tagProblemDetail.problem.timeLimit + 'ms')"></span>
            </el-tag>
            <el-tag size="small" type="warning">
              <span v-html="renderMarkdown('**内存限制:** ' + tagProblemDetail.problem.memoryLimit + 'MB')"></span>
            </el-tag>
            <el-tag size="small" type="primary">
              <span v-html="renderMarkdown('**判题模式:** ' + getJudgeModeText(tagProblemDetail.problem.judgeMode))"></span>
            </el-tag>
            <el-tag size="small" type="success">
              <span v-html="renderMarkdown('**难度:** ' + getDifficultyName(tagProblemDetail.problem.difficulty))"></span>
            </el-tag>
          </div>

          <el-divider></el-divider>

          <div class="problem-section">
            <h4>题目描述</h4>
            <div class="problem-description" v-html="renderMarkdown(tagProblemDetail.problem.description)"></div>
          </div>

          <div class="problem-section" v-if="tagProblemDetail.problem.input">
            <h4>输入格式</h4>
            <div class="problem-io" v-html="renderMarkdown(tagProblemDetail.problem.input)"></div>
          </div>

          <div class="problem-section" v-if="tagProblemDetail.problem.output">
            <h4>输出格式</h4>
            <div class="problem-io" v-html="renderMarkdown(tagProblemDetail.problem.output)"></div>
          </div>

          <div class="problem-section" v-if="tagProblemDetail.problem.hint">
            <h4>提示</h4>
            <div class="problem-hint" v-html="renderMarkdown(tagProblemDetail.problem.hint)"></div>
          </div>

          <div class="problem-section" v-if="parseExamples(tagProblemDetail.problem.examples).length > 0">
            <h4>样例</h4>
            <div v-for="(example, index) in parseExamples(tagProblemDetail.problem.examples)" :key="index" class="problem-example">
              <div class="example-item">
                <strong v-html="renderMarkdown('样例 ' + (index + 1))"></strong>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown('**输入：**')"></div>
                  <div class="example-content" v-html="renderMarkdown(example.input)"></div>
                </div>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown('**输出：**')"></div>
                  <div class="example-content" v-html="renderMarkdown(example.output)"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <span slot="footer">
        <el-button type="primary" @click="showTagProblemDetailDialog = false">关闭</el-button>
      </span>
    </el-dialog>

    <div class="exam-paper-container" v-if="!showCreateDialog">
    <el-card class="exam-paper-card">
      <div slot="header" class="card-header">
        <div class="header-left">
          <i class="el-icon-document-copy"></i>
          <span>试卷库管理</span>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="el-icon-plus" @click="openCreateDialog" size="small">
            创建试卷
          </el-button>
        </div>
      </div>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-input
              v-model="filters.keyword"
              placeholder="搜索试卷标题"
              clearable
              @clear="loadPapers"
              @keyup.enter.native="loadPapers"
            >
              <el-button slot="append" icon="el-icon-search" @click="loadPapers"></el-button>
            </el-input>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.isShared" placeholder="共享状态" clearable @change="loadPapers">
              <el-option label="全部" value=""></el-option>
              <el-option label="私有试卷" value="0"></el-option>
              <el-option label="共享试卷" value="1"></el-option>
            </el-select>
          </el-col>
        </el-row>
      </div>

      <!-- 试卷列表 -->
      <el-table
        :data="papers"
        v-loading="loading"
        stripe
        border
        style="width: 100%; margin-top: 20px"
      >
        <el-table-column prop="title" label="试卷标题" min-width="200">
          <template slot-scope="{ row }">
            <div>
              <div class="paper-title">
                {{ row.title }}
                <el-tag v-if="row.isShared === 1" size="mini" type="success" style="margin-left: 8px;">共享</el-tag>
                <el-tag v-else size="mini" type="info" style="margin-left: 8px;">私有</el-tag>
              </div>
              <div class="paper-desc" v-if="row.description">{{ row.description }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="questionCount" label="题目数量" width="100" align="center"></el-table-column>
        <el-table-column prop="totalScore" label="总分" width="80" align="center"></el-table-column>
        <el-table-column prop="creator.username" label="创建者" width="150">
          <template slot-scope="{ row }">
            <div v-if="row.creator">
              {{ row.creator.username }}
              <el-tag v-if="canEditPaper(row)" size="mini" type="info" style="margin-left: 8px;">我</el-tag>
            </div>
            <div v-else>-</div>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template slot-scope="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template slot-scope="{ row }">
            <el-button-group>
              <el-button size="mini" icon="el-icon-view" @click.stop="viewPaper(row)">查看</el-button>
              <el-button
                v-if="canEditPaper(row)"
                size="mini"
                icon="el-icon-edit"
                type="primary"
                @click.stop="editPaper(row)"
              >编辑</el-button>
              <el-button
                v-if="canEditPaper(row)"
                size="mini"
                icon="el-icon-delete"
                type="danger"
                @click.stop="deletePaper(row)"
              >删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="pagination.currentPage"
          :page-sizes="[10, 20, 50, 100]"
          :page-size="pagination.pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
        >
        </el-pagination>
      </div>
    </el-card>
  </div>

  <!-- 创建/编辑试卷页面 -->
  <div v-if="showCreateDialog" class="exam-paper-editor-page">
    <div class="editor-toolbar">
      <div class="editor-title-wrap">
        <el-button icon="el-icon-arrow-left" @click="showCreateDialog = false">返回试卷列表</el-button>
        <h3 class="editor-page-title">{{ isEditMode ? '编辑试卷' : '创建试卷' }}</h3>
      </div>
      <div class="editor-toolbar-actions">
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="savePaper" :loading="saving">保存</el-button>
      </div>
    </div>

    <el-card class="exam-paper-editor-card" shadow="never">
      <el-form :model="paperForm" :rules="paperRules" ref="paperForm" label-width="100px">
      <el-form-item label="试卷标题" prop="title">
        <el-input v-model="paperForm.title" placeholder="请输入试卷标题"></el-input>
      </el-form-item>
      <el-form-item label="试卷描述">
        <el-input
          v-model="paperForm.description"
          type="textarea"
          :rows="2"
          placeholder="请输入试卷描述（可选）"
        ></el-input>
      </el-form-item>
      <el-form-item label="共享设置">
        <el-switch v-model="paperForm.isShared" active-text="共享" inactive-text="私有"></el-switch>
        <div style="color: #909399; font-size: 12px; margin-top: 8px;">
          共享试卷可以被其他教师查看和使用
        </div>
      </el-form-item>

      <el-divider>题目选择</el-divider>

      <!-- 题目选择区域 -->
      <div class="questions-selector">
        <!-- 左侧：题库面板 -->
        <div class="question-bank-panel">
          <div class="panel-header">
            <span class="panel-title">题库 ({{ questionBank.length }}/{{ questionBankTotal }})</span>
            <el-button
              type="text"
              icon="el-icon-refresh"
              @click="loadQuestionBank"
              :loading="questionsLoading"
              size="small"
            >
              刷新
            </el-button>
          </div>

          <div class="filter-section">
            <el-input
              v-model="questionFilters.keyword"
              placeholder="搜索题目"
              prefix-icon="el-icon-search"
              clearable
              @clear="loadQuestionBank"
              @keyup.enter.native="loadQuestionBank"
              size="small"
              class="search-input"
            >
              <el-button
                slot="append"
                icon="el-icon-search"
                @click="loadQuestionBank"
              >
                搜索
              </el-button>
            </el-input>

            <el-select
              v-model="questionFilters.type"
              placeholder="题型"
              clearable
              @change="loadQuestionBank"
              size="small"
              class="filter-select"
            >
              <el-option label="全部" value=""></el-option>
              <el-option label="单选题" value="single_choice"></el-option>
              <el-option label="多选题" value="multiple_choice"></el-option>
              <el-option label="判断题" value="judge"></el-option>
              <el-option label="主观题" value="subjective"></el-option>
            </el-select>

            <el-select
              v-model="questionFilters.course"
              placeholder="课程"
              clearable
              @change="loadQuestionBank"
              size="small"
              class="filter-select"
            >
              <el-option label="全部" value=""></el-option>
              <el-option
                v-for="course in commonCourses"
                :key="course"
                :label="course"
                :value="course"
              />
            </el-select>

            <el-input
              v-model="questionFilters.tag"
              placeholder="标签筛选"
              prefix-icon="el-icon-price-tag"
              clearable
              @clear="loadQuestionBank"
              @keyup.enter.native="loadQuestionBank"
              size="small"
              class="filter-select"
            ></el-input>
          </div>

          <div class="quick-add-section">
            <div class="quick-add-title">按 ID 快速添加客观题</div>
            <div class="quick-add-row">
              <el-input
                v-model.trim="quickAddQuestionId"
                size="small"
                clearable
                placeholder="输入客观题 ID，例如 1024"
                @keyup.enter.native="quickAddObjectiveQuestion"
              ></el-input>
              <el-button
                class="quick-add-btn"
                type="primary"
                size="small"
                :loading="quickAddQuestionLoading"
                @click="quickAddObjectiveQuestion"
              >
                添加
              </el-button>
            </div>
          </div>

          <div class="question-list" v-loading="questionsLoading">
            <el-alert
              v-if="questionBank.length === 0 && !questionsLoading"
              title="题库为空"
              type="info"
              :closable="false"
              style="margin-bottom: 10px;"
            >
              <template slot="default">
                <div>当前题库没有可用的题目。</div>
                <div style="font-size: 12px; margin-top: 5px; color: #909399;">
                  提示：教师端只能看到自己创建的题目和共享的题目。如需查看所有题目，请使用管理员账户。
                </div>
              </template>
            </el-alert>
            <el-table
              :data="questionBank"
              size="small"
              :show-header="false"
              :empty-text="questionBank.length === 0 ? '暂无题目' : '搜索题目'"
              style="font-size: 12px;"
            >
              <el-table-column>
                <template slot-scope="{ row }">
                  <div class="question-item" v-if="row && row.type">
                    <div class="question-header" @click="toggleQuestionDetail(row)">
                      <el-tag size="mini" :type="getQuestionTypeTag(row.type)">
                        {{ getQuestionTypeLabel(row.type) }}
                      </el-tag>
                      <span style="margin-left: 10px;">{{ row.title }}</span>
                      <el-tag v-if="row.isShared === 1" size="mini" type="success" style="margin-left: 5px;">共享</el-tag>
                      <el-tag v-else-if="row.creatorId === currentUserId" size="mini" type="info" style="margin-left: 5px;">私有</el-tag>
                      <i :class="row.showDetail ? 'el-icon-arrow-up' : 'el-icon-arrow-down'" style="margin-left: auto; color: #909399;"></i>
                    </div>
                    <el-collapse-transition>
                      <div v-show="row.showDetail" class="question-detail-content">
                        <div class="question-description markdown-body" v-html="renderMarkdown(row.content || row.title)" v-highlight></div>
                        <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="question-options">
                          <div v-for="(option, index) in parseOptions(row.options)" :key="index" class="option-item">
                            <span class="option-label">{{ option.label }}.</span>
                            <span class="option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                          </div>
                        </div>
                        <div class="question-answer-meta" v-if="row && row.type">
                          <span class="meta-label">正确答案：</span>
                          <span class="meta-value">{{ formatAnswer(row) }}</span>
                        </div>
                      </div>
                    </el-collapse-transition>
                  </div>
                </template>
              </el-table-column>
              <el-table-column width="60" align="center">
                <template slot-scope="{ row }">
                  <el-button
                    v-if="row && row.id"
                    size="mini"
                    :type="isQuestionSelected(row) ? 'info' : 'primary'"
                    icon="el-icon-plus"
                    :disabled="isQuestionSelected(row)"
                    :title="isQuestionSelected(row) ? '已添加' : '添加题目'"
                    style="padding: 5px 8px;"
                    @click.stop="addQuestion(row)"
                  >
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <div v-if="questionBankTotal > 0" style="padding: 10px; text-align: center;">
              <el-pagination
                @current-change="handleQuestionPageChange"
                :current-page="questionCurrentPage"
                :page-size="questionPageSize"
                small
                layout="prev, pager, next"
                :total="questionBankTotal"
              >
              </el-pagination>
            </div>
          </div>

          <div class="add-programming-section">
            <el-button
              type="success"
              size="small"
              icon="el-icon-plus"
              @click="showAddProgrammingDialog = true"
              style="width: 100%;"
            >
              添加编程题
            </el-button>
          </div>
        </div>

        <!-- 右侧：已选题目 -->
        <div class="selected-questions-panel">
          <div class="panel-header">
            <span class="panel-title">已选题目</span>
            <el-tag size="small" type="info">{{ paperForm.questions.length }} 题</el-tag>
            <el-tag size="small" type="success" style="margin-left: 8px;">
              总分: {{ getSelectedScore() }} 分
            </el-tag>
          </div>

          <div class="selected-list" v-if="paperForm.questions.length > 0">
            <transition-group name="list">
              <template v-for="(q, index) in paperForm.questions">
                <div
                  v-if="q && (q.questionId || q.problemId)"
                  :key="q.questionId || q.problemId || index"
                  class="selected-item"
                >
                  <div class="item-header">
                    <span class="item-order">{{ index + 1 }}.</span>
                    <el-tag size="mini" :type="getQuestionTypeTag(q.questionType)">
                      {{ getQuestionTypeLabel(q.questionType) }}
                    </el-tag>
                    <el-input-number
                      v-model="q.score"
                      :min="1"
                      :max="100"
                      size="mini"
                      style="margin-left: 10px;"
                    ></el-input-number>
                    <span style="margin-left: 5px;">分</span>
                    <el-button-group style="margin-left: auto;">
                      <el-button
                        size="mini"
                        type="primary"
                        icon="el-icon-top"
                        :disabled="index === 0"
                        @click="moveQuestionUp(index)"
                      ></el-button>
                      <el-button
                        size="mini"
                        type="primary"
                        icon="el-icon-bottom"
                        :disabled="index === paperForm.questions.length - 1"
                        @click="moveQuestionDown(index)"
                      ></el-button>
                      <el-button
                        size="mini"
                        type="danger"
                        icon="el-icon-delete"
                        @click="removeQuestion(index)"
                      ></el-button>
                    </el-button-group>
                  </div>
                </div>
                <div v-if="q && (q.questionId || q.problemId)" class="item-title" :key="'title-' + (q.questionId || q.problemId || index)">{{ getQuestionTitle(q) }}</div>
                <div v-if="q && (q.questionId || q.problemId)" class="item-content" :key="'content-' + (q.questionId || q.problemId || index)">
                  <!-- 客观题 -->
                  <div v-if="q.question && (q.question.title || q.question.content)">
                    <div class="item-description markdown-body" v-html="renderMarkdown(q.question.content || q.question.title)" v-highlight></div>
                    <div v-if="q.question.type === 'single_choice' || q.question.type === 'multiple_choice'" class="item-options">
                      <div v-for="(option, index) in parseOptions(q.question.options)" :key="index" class="option-item">
                        <span class="option-label">{{ option.label }}.</span>
                        <span class="option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                      </div>
                    </div>
                    <!-- 显示课程和标签 -->
                    <div class="item-meta-tags" style="margin-top: 8px;">
                      <el-tag v-if="q.question.course" type="warning" size="mini" style="margin-right: 5px;">
                        <i class="el-icon-collection"></i> {{ q.question.course }}
                      </el-tag>
                      <el-tag
                        v-for="(tag, idx) in parseQuestionTags(q.question.tags)"
                        :key="idx"
                        size="mini"
                        type="info"
                        style="margin-right: 3px;"
                      >
                        {{ tag }}
                      </el-tag>
                    </div>
                    <div class="item-answer">
                      <span class="meta-label">答案：</span>
                      <span class="meta-value">{{ formatAnswer(q.question) }}</span>
                    </div>
                  </div>
                  <!-- 编程题 -->
                  <div v-else-if="q.problemId">
                    <div v-if="q.problem && q.problem.title" class="programming-item">
                      <div class="item-title">
                        <strong>{{ q.problem.title }}</strong>
                        <el-button
                          size="mini"
                          type="text"
                          icon="el-icon-view"
                          @click.stop="viewProblemDetail(q.problem)"
                          style="margin-left: 10px;"
                        >
                          查看详情
                        </el-button>
                      </div>
                      <div class="item-description" v-html="renderMarkdown(q.problem.description)"></div>
                      <div class="problem-meta">
                        <el-tag size="small">时间: {{ q.problem.timeLimit }}ms</el-tag>
                        <el-tag size="small" type="warning">内存: {{ q.problem.memoryLimit }}MB</el-tag>
                        <el-tag size="small" type="primary">判题模式: {{ getJudgeModeText(q.problem.judgeMode) }}</el-tag>
                      </div>
                    </div>
                    <div v-else class="item-description">
                      BingOJ 编程题 - {{ q.problemId }}
                      <el-button
                        size="mini"
                        type="text"
                        icon="el-icon-view"
                        @click.stop="fetchAndviewProblemDetail(q.problemId)"
                        style="margin-left: 10px;"
                      >
                        查看详情
                      </el-button>
                    </div>
                  </div>
                </div>
              </template>
            </transition-group>
          </div>

          <div v-else class="empty-selected">
            <i class="el-icon-document"></i>
            <p>暂未选择题目</p>
            <p class="hint">从左侧题库中添加题目</p>
          </div>
        </div>
      </div>
      </el-form>

      <div class="editor-footer-actions">
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="savePaper" :loading="saving">保存</el-button>
      </div>
    </el-card>
  </div>

  <!-- 添加编程题对话框 -->
  <el-dialog title="添加编程题" :visible.sync="showAddProgrammingDialog" width="1100px">
      <el-form :model="programmingForm" label-width="120px">
        <el-form-item label="方式选择">
          <el-radio-group v-model="programmingInputMode" @change="handleProgrammingInputModeChange">
            <el-radio label="manual">手动输入题目ID</el-radio>
            <el-radio label="tag">按标签选择题目</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 手动输入模式 -->
        <template v-if="programmingInputMode === 'manual'">
          <el-form-item label="题目ID" required>
            <el-input v-model="programmingForm.problemId" placeholder="请输入 BingOJ 题目 ID（如 0001）" style="width: 300px;" />
            <el-button
              type="primary"
              icon="el-icon-search"
              style="margin-left: 10px;"
              @click="fetchProgrammingProblemInfo"
              :loading="fetchingProblem"
            >
              获取题目信息
            </el-button>
          </el-form-item>
        </template>

        <!-- 标签选择模式 -->
        <template v-if="programmingInputMode === 'tag'">
          <el-form-item label="选择标签">
            <div v-if="problemTagsLoading" v-loading="true" style="min-height: 100px;"></div>
            <div v-else>
              <div v-for="(tagsAndClassification, index) in problemTagsAndClassificationList" :key="index" style="margin-bottom: 15px;">
                <div style="margin-bottom: 8px; font-weight: bold; color: #606266;">
                  {{ tagsAndClassification.classification ? tagsAndClassification.classification.name : '未分类' }}
                </div>
                <el-tag
                  v-for="tag in tagsAndClassification.tagList"
                  :key="tag.id"
                  :type="isTagSelected(tag.id) ? 'primary' : 'info'"
                  :color="isTagSelected(tag.id) ? (tag.color || '#409eff') : ''"
                  effect="dark"
                  @click="toggleProblemTag(tag)"
                  style="margin-right: 10px; margin-bottom: 10px; cursor: pointer;"
                  size="medium"
                >
                  {{ tag.name }}
                </el-tag>
              </div>
            </div>
          </el-form-item>

          <!-- 已选标签显示 -->
          <el-form-item v-if="selectedProblemTagIds.length > 0" label="已选标签">
            <el-tag
              v-for="tagId in selectedProblemTagIds"
              :key="tagId"
              closable
              @close="removeSelectedTag(tagId)"
              style="margin-right: 10px;"
              type="primary"
            >
              {{ getTagName(tagId) }}
            </el-tag>
            <el-button type="text" size="small" @click="clearAllTags" style="margin-left: 10px;">清空</el-button>
          </el-form-item>

          <!-- 标签筛选结果 -->
          <el-form-item v-if="filteredProblemsByTag.length > 0" label="题目列表">
            <el-alert
              type="info"
              :closable="false"
              style="margin-bottom: 10px;"
            >
              <span slot="title">
                已选标签下共有 {{ filteredProblemsTotal }} 道题目（当前显示前 {{ filteredProblemsByTag.length }} 道），点击题号查看详情
              </span>
            </el-alert>
            <el-table
              :data="filteredProblemsByTag"
              stripe
              border
              max-height="300"
              style="width: 100%"
            >
              <el-table-column prop="problemId" label="题号" width="120">
                <template slot-scope="{ row }">
                  <el-link type="primary" @click="viewTagProblemDetail(row)">{{ row.problemId }}</el-link>
                </template>
              </el-table-column>
              <el-table-column prop="title" label="题名" min-width="200" show-overflow-tooltip></el-table-column>
              <el-table-column prop="difficulty" label="难度" width="80" align="center">
                <template slot-scope="{ row }">
                  <el-tag :type="getDifficultyTagType(row.difficulty)" size="mini">
                    {{ getDifficultyName(row.difficulty) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="120" align="center">
                <template slot-scope="{ row }">
                  <el-button type="primary" size="mini" @click="selectProblemByTag(row)">
                    添加此题
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <!-- 分页 -->
            <div style="margin-top: 15px; text-align: center;">
              <el-pagination
                v-if="filteredProblemsTotal > 0"
                @current-change="handleFilteredProblemsPageChange"
                :current-page="filteredProblemsCurrentPage"
                :page-size="filteredProblemsPageSize"
                :total="filteredProblemsTotal"
                layout="prev, pager, next, total"
                small
              >
              </el-pagination>
            </div>
          </el-form-item>
        </template>

        <!-- 题目预览区域 -->
        <div v-if="programmingProblemPreview" class="problem-preview">
          <el-divider content-position="left">题目预览</el-divider>
          <el-card>
            <h3>{{ programmingProblemPreview.problem.title }}</h3>
            <div class="problem-description" v-html="renderMarkdown(programmingProblemPreview.problem.description)"></div>
            <div class="problem-meta">
              <el-tag size="small">题目ID: {{ programmingProblemPreview.problem.problemId }}</el-tag>
              <el-tag size="small" type="info">时间限制: {{ programmingProblemPreview.problem.timeLimit }}ms</el-tag>
              <el-tag size="small" type="warning">内存限制: {{ programmingProblemPreview.problem.memoryLimit }}MB</el-tag>
              <el-tag size="small" type="primary">判题模式: {{ getJudgeModeText(programmingProblemPreview.problem.judgeMode) }}</el-tag>
            </div>
            <div v-if="programmingProblemPreview.problem.input" class="problem-section">
              <h4>输入格式</h4>
              <div v-html="renderMarkdown(programmingProblemPreview.problem.input)"></div>
            </div>
            <div v-if="programmingProblemPreview.problem.output" class="problem-section">
              <h4>输出格式</h4>
              <div v-html="renderMarkdown(programmingProblemPreview.problem.output)"></div>
            </div>
            <div v-if="programmingProblemPreview.problem.hint" class="problem-section">
              <h4>提示</h4>
              <div v-html="renderMarkdown(programmingProblemPreview.problem.hint)"></div>
            </div>
          </el-card>
        </div>

        <el-form-item label="分值" required>
          <el-input-number v-model="programmingForm.score" :min="1" :max="100" />
        </el-form-item>
      </el-form>

      <span slot="footer">
        <el-button @click="showAddProgrammingDialog = false">取消</el-button>
        <el-button
          v-if="programmingInputMode === 'manual'"
          type="primary"
          @click="confirmAddProgrammingQuestion"
          :disabled="!programmingProblemPreview"
        >
          确定添加
        </el-button>
      </span>
    </el-dialog>

    <!-- 查看试卷对话框 -->
    <el-dialog title="试卷详情" :visible.sync="showViewDialog" width="80%">
      <div v-if="currentPaper && validQuestions.length > 0" class="paper-detail">
        <div class="paper-info">
          <h2>{{ currentPaper.title }}</h2>
          <div class="meta-info">
            <el-tag v-if="currentPaper.isShared === 1" type="success">共享试卷</el-tag>
            <el-tag v-else type="info">私有试卷</el-tag>
            <span style="margin-left: 20px;">题目数量：{{ currentPaper.questionCount }}</span>
            <span style="margin-left: 20px;">总分：{{ currentPaper.totalScore }} 分</span>
          </div>
          <p v-if="currentPaper.description" class="description">{{ currentPaper.description }}</p>
        </div>

        <el-divider></el-divider>

        <div class="questions-list">
          <div v-for="(q, index) in validQuestions" :key="index" class="question-item">
            <div class="question-header">
              <span class="question-order">{{ index + 1 }}.</span>
              <el-tag size="small" :type="getQuestionTypeTag(q?.questionType)">
                {{ getQuestionTypeLabel(q?.questionType) }}
              </el-tag>
              <el-tag size="small" type="warning" style="margin-left: 8px;">{{ q?.score || 0 }} 分</el-tag>
            </div>
            <div class="question-content">
              <!-- 客观题 -->
              <div v-if="q?.question && (q.question.title || q.question.content)">
                <div class="question-title markdown-body" v-html="renderMarkdown(q.question.content || q.question.title)" v-highlight></div>
                <div v-if="q.question.type === 'single_choice' || q.question.type === 'multiple_choice'" class="question-options">
                  <div v-for="(option, index) in parseOptions(q.question.options)" :key="index" class="option-item">
                    <span class="option-label">{{ option.label }}.</span>
                    <span class="option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                  </div>
                </div>
                <!-- 显示课程和标签 -->
                <div v-if="q.question.course || q.question.tags" class="question-tags-course" style="margin-top: 8px;">
                  <el-tag v-if="q.question.course" type="warning" size="mini" style="margin-right: 5px;">
                    <i class="el-icon-collection"></i> {{ q.question.course }}
                  </el-tag>
                  <el-tag
                    v-for="(tag, idx) in parseQuestionTags(q.question.tags)"
                    :key="idx"
                    size="mini"
                    type="info"
                    style="margin-right: 3px;"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
                <div v-if="q.question.type !== 'subjective'" class="question-meta">
                  <span class="meta-label">正确答案：</span>
                  <span class="meta-value">{{ formatAnswer(q.question) }}</span>
                </div>
              </div>
              <!-- 编程题 -->
              <div v-else-if="q?.problemId">
                <div v-if="q.problem && q.problem.title">
                  <div class="question-title">
                    <strong>{{ q.problem.title }}</strong>
                    <el-button
                      size="mini"
                      type="text"
                      icon="el-icon-view"
                      @click.stop="viewProblemDetail(q.problem)"
                      style="margin-left: 10px;"
                    >
                      查看详情
                    </el-button>
                  </div>
                  <div class="problem-description" v-html="renderMarkdown(q.problem.description)"></div>
                  <div class="problem-meta">
                    <el-tag size="small" type="info">时间: {{ q.problem.timeLimit }}ms</el-tag>
                    <el-tag size="small" type="warning">内存: {{ q.problem.memoryLimit }}MB</el-tag>
                    <el-tag size="small" type="primary">判题模式: {{ getJudgeModeText(q.problem.judgeMode) }}</el-tag>
                  </div>
                </div>
                <div v-else>
                  <div class="question-title">
                    <strong>BingOJ 编程题 - {{ q.problemId }}</strong>
                    <el-button
                      size="mini"
                      type="text"
                      icon="el-icon-view"
                      @click.stop="fetchAndviewProblemDetail(q.problemId)"
                      style="margin-left: 10px;"
                    >
                      查看详情
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else-if="currentPaper" class="empty-questions">
        <i class="el-icon-info"></i>
        <p>该试卷暂无题目</p>
      </div>
      <div v-else class="paper-loading">
        <i class="el-icon-loading"></i>
        <span>加载中...</span>
      </div>
    </el-dialog>
  </div>
</div>
</template>

<script>
import { mapGetters } from 'vuex'
import classroomApi from '@/api/classroom'
import api from '@/common/api'
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})
md.use(katex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'ExamPaper',
  data() {
    return {
      loading: false,
      papers: [],
      filters: {
        keyword: '',
        isShared: ''
      },
      pagination: {
        currentPage: 1,
        pageSize: 20,
        total: 0
      },
      showCreateDialog: false,
      showViewDialog: false,
      showAddProgrammingDialog: false,
      isEditMode: false,
      saving: false,
      currentPaper: null,
      isEditing: false, // 标记是否正在编辑，防止竞态条件
      paperForm: {
        title: '',
        description: '',
        isShared: false,
        questions: []
      },
      paperRules: {
        title: [{ required: true, message: '请输入试卷标题', trigger: 'blur' }]
      },
      // 题库相关
      questionBank: [],
      questionsLoading: false,
      questionFilters: {
        keyword: '',
        type: '',
        course: '',
        tag: ''
      },
      questionCurrentPage: 1,
      questionPageSize: 10,
      questionBankTotal: 0,
      quickAddQuestionId: '',
      quickAddQuestionLoading: false,
      // 常用课程列表
      commonCourses: [
        '数据结构',
        '算法设计与分析',
        '计算机网络',
        '操作系统',
        '计算机组成原理',
        '高等数学',
        '线性代数',
        '政治',
        '英语'
      ],
      // 编程题相关
      programmingInputMode: 'manual', // 'manual' 或 'tag'
      programmingForm: {
        problemId: '',
        score: 10
      },
      programmingProblemPreview: null,
      fetchingProblem: false,
      // 标签筛选相关
      problemTagsAndClassificationList: [],
      selectedProblemTagIds: [], // 改为数组，支持多选
      filteredProblemsByTag: [],
      filteredProblemsTotal: 0, // 添加总数字段
      problemTagsLoading: false,
      // 标签题目分页
      filteredProblemsCurrentPage: 1,
      filteredProblemsPageSize: 30,
      // 标签题目详情弹窗
      showTagProblemDetailDialog: false,
      tagProblemDetail: null,
      loadingTagProblemDetail: false,
      // 查看编程题详情
      showProblemDetailDialog: false,
      currentViewProblem: null,
      fetchingViewProblem: false
    }
  },
  computed: {
    ...mapGetters(['userInfo']),
    currentUserId() {
      return this.userInfo?.uid || ''
    },
    // 安全地获取当前试卷的题目列表，过滤掉无效的题目
    validQuestions() {
      if (!this.currentPaper || !this.currentPaper.questions) {
        return []
      }
      return this.currentPaper.questions.filter(q => {
        // 确保题目对象存在且包含必要的字段
        return q &&
               typeof q === 'object' &&
               (q.questionId || q.problemId) &&
               q.questionType &&
               typeof q.score === 'number'
      })
    }
  },
  mounted() {
    this.loadPapers()
    this.loadProblemTagsAndClassification()
  },
  watch: {
    showCreateDialog(newVal, oldVal) {
      // 当对话框从打开变为关闭时，总是重置表单，避免残留数据
      if (oldVal === true && newVal === false) {
        // 取消任何正在进行的编辑操作
        this.isEditing = false
        this.isEditMode = false
        this.currentPaper = null
        this.paperForm = {
          title: '',
          description: '',
          isShared: false,
          questions: []
        }
        this.quickAddQuestionId = ''
        this.quickAddQuestionLoading = false
        if (this.$refs.paperForm) {
          this.$refs.paperForm.clearValidate()
        }
      }
    },
    showProblemDetailDialog(newVal, oldVal) {
      if (newVal === true) {
        // 延迟检查对话框是否真的被创建了
        this.$nextTick(() => {
          setTimeout(() => {
            const dialogWrappers = document.querySelectorAll('.el-dialog__wrapper')
            // 如果编程题详情对话框没有出现，尝试手动创建
            const problemDetailDialogExists = Array.from(dialogWrappers).some(wrapper => {
              const title = wrapper.querySelector('.el-dialog__title')
              return title && title.textContent.trim() === '编程题详情'
            })

            if (!problemDetailDialogExists) {
              // 尝试使用 Element UI 的 API
              this.$message.warning('编程题详情对话框无法打开，请刷新页面后重试')
            }
          }, 100)
        })
      }
    }
  },
  methods: {
    openCreateDialog() {
      // 如果对话框已经打开，先关闭它，触发 watch 的重置逻辑
      if (this.showCreateDialog) {
        this.showCreateDialog = false
        // 等待对话框关闭动画完成后再重新打开
        this.$nextTick(() => {
          setTimeout(() => {
            this.doOpenCreateDialog()
          }, 100)
        })
      } else {
        this.doOpenCreateDialog()
      }
    },
    doOpenCreateDialog() {
      // 取消任何正在进行的编辑操作
      this.isEditing = false

      // 强制重置所有状态
      this.isEditMode = false
      this.currentPaper = null
      this.paperForm = {
        title: '',
        description: '',
        isShared: false,
        questions: []
      }

      // 清空题库，避免被之前的题目影响
      this.questionBank = []
      this.questionBankTotal = 0
      this.quickAddQuestionId = ''
      this.quickAddQuestionLoading = false

      // 在下一个tick清除表单验证并打开对话框
      this.$nextTick(() => {
        if (this.$refs.paperForm) {
          this.$refs.paperForm.clearValidate()
        }
        // 加载题库
        this.loadQuestionBank()
        // 打开对话框
        this.showCreateDialog = true
      })
    },
    canEditPaper(paper) {
      // 检查是否可以编辑试卷（只有创建者可以编辑）
      return paper && paper.creatorId && this.currentUserId && paper.creatorId === this.currentUserId
    },
    async loadPapers() {
      this.loading = true
      try {
        const res = await classroomApi.getExamPaperList({
          page: this.pagination.currentPage,
          limit: this.pagination.pageSize,
          keyword: this.filters.keyword || undefined,
          isShared: this.filters.isShared
        })
        if (res.data.code === 200) {
          this.papers = res.data.data.papers || []
          this.pagination.total = res.data.data.total || 0
        }
      } catch (error) {
        this.$message.error('加载失败')
      } finally {
        this.loading = false
      }
    },
    handleSizeChange(val) {
      this.pagination.pageSize = val
      this.loadPapers()
    },
    handleCurrentChange(val) {
      this.pagination.currentPage = val
      this.loadPapers()
    },
    // 加载题库
    async loadQuestionBank() {
      this.questionsLoading = true
      try {
        const res = await this.$store.dispatch('classroom/getQuestionBank', {
          classroomId: '0', // 不限制班级，获取所有题目
          page: this.questionCurrentPage,
          limit: this.questionPageSize,
          keyword: this.questionFilters.keyword || undefined,
          type: this.questionFilters.type || undefined,
          course: this.questionFilters.course || undefined,
          tag: this.questionFilters.tag || undefined
        })
        // Vuex action 返回的是 res.data，结构为 { code, message, data }
        if (res && res.code === 200) {
          // 过滤掉 undefined 或 null 的题目
          this.questionBank = (res.data.questions || []).filter(q => {
            const isValid = q && q.id && typeof q === 'object'
            return isValid
          })
          this.questionBankTotal = res.data.total || 0
        }
      } catch (error) {
        this.$message.error('加载题库失败')
      } finally {
        this.questionsLoading = false
      }
    },
    handleQuestionPageChange(page) {
      this.questionCurrentPage = page
      this.loadQuestionBank()
    },
    isObjectiveQuestionType(type) {
      return ['single_choice', 'multiple_choice', 'judge'].includes(type)
    },
    async quickAddObjectiveQuestion() {
      const questionId = String(this.quickAddQuestionId || '').trim()
      if (!questionId) {
        this.$message.warning('请输入题目ID')
        return
      }
      if (!/^\d+$/.test(questionId)) {
        this.$message.warning('题目ID必须是数字')
        return
      }
      if (this.paperForm.questions.some(q => q && String(q.questionId) === questionId)) {
        this.$message.warning('该题目已添加')
        return
      }

      this.quickAddQuestionLoading = true
      try {
        const res = await classroomApi.getQuestionDetail(questionId)
        if (!res || !res.data || res.data.code !== 200 || !res.data.data) {
          this.$message.error('未找到该题目')
          return
        }
        const question = res.data.data
        if (!this.isObjectiveQuestionType(question.type)) {
          this.$message.warning('该题不是客观题，仅支持单选/多选/判断题')
          return
        }
        this.addQuestion(question)
        this.quickAddQuestionId = ''
      } catch (error) {
        this.$message.error('根据ID获取题目失败')
      } finally {
        this.quickAddQuestionLoading = false
      }
    },
    // 添加客观题
    addQuestion(question) {
      // 检查参数是否有效
      if (!question || !question.id) {
        this.$message.warning('题目数据无效')
        return
      }

      // 检查是否已添加
      if (this.isQuestionSelected(question)) {
        this.$message.warning('该题目已添加')
        return
      }

      // 添加题目
      this.paperForm.questions.push({
        questionId: question.id,
        problemId: null,
        questionType: question.type,
        score: question.score || this.getDefaultScore(question.type),
        title: question.title,
        question: question
      })

      this.$message.success('添加成功')
    },
    isQuestionSelected(question) {
      if (!question || !question.id) return false
      return this.paperForm.questions.some(q => q && q.questionId === question.id)
    },
    getDefaultScore(type) {
      const scores = {
        single_choice: 2,
        multiple_choice: 5,
        judge: 1,
        subjective: 5
      }
      return scores[type] || 10
    },
    removeQuestion(index) {
      this.paperForm.questions.splice(index, 1)
      this.$forceUpdate()
    },
    moveQuestionUp(index) {
      if (index > 0) {
        const temp = this.paperForm.questions[index]
        this.$set(this.paperForm.questions, index, this.paperForm.questions[index - 1])
        this.$set(this.paperForm.questions, index - 1, temp)
      }
    },
    moveQuestionDown(index) {
      if (index < this.paperForm.questions.length - 1) {
        const temp = this.paperForm.questions[index]
        this.$set(this.paperForm.questions, index, this.paperForm.questions[index + 1])
        this.$set(this.paperForm.questions, index + 1, temp)
      }
    },
    getQuestionTitle(q) {
      // 检查 q 是否为 null 或 undefined
      if (!q) {
        return '未知题目'
      }

      try {
        // 优先使用直接存储的标题（已保存的试卷）
        if (q.title) {
          return q.title
        }
        // 检查是否有完整的题目对象（从题库添加的新题目）
        if (q.question && q.question.title) {
          return q.question.title
        }
        // 检查是否有编程题对象
        if (q.problem && q.problem.title) {
          return q.problem.title
        }
        // 回退到显示ID
        if (q.questionId) {
          return `题目ID: ${q.questionId}`
        } else if (q.problemId) {
          return `BingOJ 编程题 - ${q.problemId}`
        }
        return '未知题目'
      } catch (error) {
        return '未知题目'
      }
    },
    // 编程题相关
    async fetchProgrammingProblemInfo() {
      if (!this.programmingForm.problemId) {
        this.$message.warning('请输入题目ID')
        return
      }

      this.fetchingProblem = true
      try {
        const res = await api.getProblem(this.programmingForm.problemId, '0', undefined)

        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          this.programmingProblemPreview = res.data.data
        } else {
          this.$message.error('获取题目信息失败')
        }
      } catch (error) {
        this.$message.error('获取题目信息失败')
      } finally {
        this.fetchingProblem = false
      }
    },
    confirmAddProgrammingQuestion() {
      if (!this.programmingForm.problemId) {
        this.$message.warning('请输入题目ID')
        return
      }

      if (!this.programmingProblemPreview) {
        this.$message.warning('请先点击"获取题目信息"按钮')
        return
      }

      // 检查是否已添加
      if (this.paperForm.questions.some(q => q.problemId === this.programmingForm.problemId)) {
        this.$message.warning('该编程题已添加')
        return
      }

      // 添加编程题
      this.paperForm.questions.push({
        questionId: null,
        problemId: this.programmingForm.problemId,
        questionType: 'programming',
        score: this.programmingForm.score,
        title: this.programmingProblemPreview.problem.title,
        problem: this.programmingProblemPreview.problem
      })

      this.$message.success('添加成功')
      this.showAddProgrammingDialog = false
      this.resetProgrammingForm()
    },
    resetProgrammingForm() {
      this.programmingForm = {
        problemId: '',
        score: 10
      }
      this.programmingProblemPreview = null
    },
    // ==================== 标签筛选相关方法 ====================
    // 加载标签和分类
    async loadProblemTagsAndClassification() {
      this.problemTagsLoading = true
      try {
        const res = await api.getProblemTagsAndClassification('ME')
        if (res && res.data && res.data.data) {
          this.problemTagsAndClassificationList = res.data.data
        }
      } catch (error) {
        console.error('加载标签失败:', error)
      } finally {
        this.problemTagsLoading = false
      }
    },
    // 切换输入模式
    handleProgrammingInputModeChange(mode) {
      if (mode === 'tag') {
        // 切换到标签模式时，清空手动输入的内容
        this.programmingForm.problemId = ''
        this.programmingProblemPreview = null
      } else {
        // 切换到手动模式时，清空标签选择的内容
        this.selectedProblemTagIds = []
        this.filteredProblemsByTag = []
        this.filteredProblemsTotal = 0
      }
    },
    // 判断标签是否已选中
    isTagSelected(tagId) {
      return this.selectedProblemTagIds.includes(tagId)
    },
    // 切换标签选中状态（点击选择/取消）
    async toggleProblemTag(tag) {
      const index = this.selectedProblemTagIds.indexOf(tag.id)
      if (index > -1) {
        // 已选中，取消选择
        this.selectedProblemTagIds.splice(index, 1)
      } else {
        // 未选中，添加到已选列表
        this.selectedProblemTagIds.push(tag.id)
      }
      // 重置到第一页并重新加载题目列表
      this.filteredProblemsCurrentPage = 1
      await this.loadProblemsByTags()
    },
    // 移除已选标签
    async removeSelectedTag(tagId) {
      const index = this.selectedProblemTagIds.indexOf(tagId)
      if (index > -1) {
        this.selectedProblemTagIds.splice(index, 1)
        this.filteredProblemsCurrentPage = 1
        await this.loadProblemsByTags()
      }
    },
    // 清空所有已选标签
    clearAllTags() {
      this.selectedProblemTagIds = []
      this.filteredProblemsByTag = []
      this.filteredProblemsTotal = 0
      this.filteredProblemsCurrentPage = 1
    },
    // 标签题目分页改变
    async handleFilteredProblemsPageChange(page) {
      this.filteredProblemsCurrentPage = page
      await this.loadProblemsByTags()
    },
    // 根据已选标签加载题目列表
    async loadProblemsByTags() {
      if (this.selectedProblemTagIds.length === 0) {
        this.filteredProblemsByTag = []
        this.filteredProblemsTotal = 0
        return
      }

      try {
        const res = await api.getProblemList({
          oj: 'ME',
          tagId: this.selectedProblemTagIds.join(','), // 传递逗号分隔的标签ID字符串
          limit: this.filteredProblemsPageSize,
          currentPage: this.filteredProblemsCurrentPage
        })

        if (res && res.data && res.data.data) {
          this.filteredProblemsByTag = res.data.data.records || []
          this.filteredProblemsTotal = res.data.data.total || 0
        } else {
          this.filteredProblemsByTag = []
          this.filteredProblemsTotal = 0
        }
      } catch (error) {
        console.error('获取题目列表失败:', error)
        this.$message.error('获取题目列表失败')
        this.filteredProblemsByTag = []
        this.filteredProblemsTotal = 0
      }
    },
    // 获取标签名称
    getTagName(tagId) {
      const tag = this.getTagById(tagId)
      return tag ? tag.name : ''
    },
    // 根据ID获取标签对象
    getTagById(tagId) {
      for (const group of this.problemTagsAndClassificationList) {
        const tag = group.tagList.find(t => t.id === tagId)
        if (tag) {
          return tag
        }
      }
      return null
    },
    // 查看标签题目的详情
    async viewTagProblemDetail(problem) {
      this.loadingTagProblemDetail = true
      this.showTagProblemDetailDialog = true

      try {
        const res = await api.getProblem(problem.problemId, '0', undefined)
        if (res && res.status === 200 && res.data && res.data.data) {
          this.tagProblemDetail = res.data.data
        }
      } catch (error) {
        console.error('获取题目详情失败:', error)
        this.$message.error('获取题目详情失败')
      } finally {
        this.loadingTagProblemDetail = false
      }
    },
    // 使用标签查看的题目
    useTagProblem() {
      if (this.tagProblemDetail) {
        this.programmingForm.problemId = this.tagProblemDetail.problem.problemId
        this.programmingInputMode = 'manual'
        this.fetchProgrammingProblemInfo()
        this.showTagProblemDetailDialog = false
      }
    },
    // 通过标签选择题目（直接添加）
    async selectProblemByTag(problem) {
      this.programmingForm.problemId = problem.problemId

      // 获取完整题目信息
      this.fetchingProblem = true
      try {
        const res = await api.getProblem(this.programmingForm.problemId, '0', undefined)

        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          const problemData = res.data.data

          // 检查是否已添加
          if (this.paperForm.questions.some(q => q.problemId === this.programmingForm.problemId)) {
            this.$message.warning('该编程题已添加')
            return
          }

          // 直接添加编程题
          this.paperForm.questions.push({
            questionId: null,
            problemId: this.programmingForm.problemId,
            questionType: 'programming',
            score: this.programmingForm.score,
            title: problemData.problem.title,
            problem: problemData.problem
          })

          this.$message.success(`已添加题目：${problem.problemId} - ${problem.title}`)
          // 清空表单以便继续添加，但保持对话框打开
          this.programmingForm.problemId = ''
          this.programmingProblemPreview = null
        } else {
          this.$message.error('获取题目信息失败')
        }
      } catch (error) {
        console.error('获取题目信息失败:', error)
        this.$message.error('获取题目信息失败')
      } finally {
        this.fetchingProblem = false
      }
    },
    // 获取难度标签类型
    getDifficultyTagType(difficulty) {
      const typeMap = {
        0: 'info',     // 入门
        1: 'success',  // 简单
        2: '',         // 中等（默认灰色）
        3: 'warning',  // 困难
        4: 'danger',   // 大师
        5: 'danger'    // 专家
      }
      return typeMap[difficulty] || ''
    },
    // 获取难度名称（6个梯度）
    getDifficultyName(difficulty) {
      const nameMap = {
        0: '入门',
        1: '简单',
        2: '中等',
        3: '困难',
        4: '大师',
        5: '专家'
      }
      return nameMap[difficulty] || '未知'
    },
    getSelectedScore() {
      return this.paperForm.questions.reduce((sum, q) => sum + (q.score || 0), 0)
    },
    async viewPaper(row) {
      try {
        const res = await classroomApi.getExamPaperDetail(row.id)
        if (res.data.code === 200) {
          const paper = res.data.data
          // 安全地处理题目数据，过滤掉无效的题目
          this.currentPaper = {
            ...paper,
            questions: (paper.questions || []).filter(q => {
              // 确保题目对象存在且包含必要的字段
              return q &&
                     typeof q === 'object' &&
                     (q.questionId || q.problemId) &&
                     q.questionType &&
                     typeof q.score === 'number'
            })
          }
          this.showViewDialog = true
        }
      } catch (error) {
        this.$message.error('加载试卷详情失败')
      }
    },
    async editPaper(row) {
      // 权限检查：只能编辑自己创建的试卷
      if (!this.canEditPaper(row)) {
        this.$message.warning('只能编辑自己创建的试卷')
        return
      }

      // 如果对话框已经打开，先关闭它，触发 watch 的重置逻辑
      if (this.showCreateDialog) {
        this.showCreateDialog = false
        // 等待对话框关闭后再打开编辑对话框
        await new Promise(resolve => {
          this.$nextTick(() => {
            setTimeout(resolve, 100)
          })
        })
      }

      // 设置编辑标志
      this.isEditing = true

      try {
        const res = await classroomApi.getExamPaperDetail(row.id)

        // 再次检查是否已取消编辑（双重检查）
        if (!this.isEditing) {
          return
        }

        if (res.data.code === 200) {
          const paper = res.data.data

          // 安全地映射题目数据，确保所有必需字段都有值
          const validQuestions = (paper.questions || []).filter(q => {
            const isValid = q && typeof q === 'object' && (q.questionId || q.problemId) && q.questionType && typeof q.score === 'number'
            return isValid
          })

          this.paperForm = {
            title: paper.title || '',
            description: paper.description || '',
            isShared: paper.isShared === 1,
            questions: validQuestions.map(q => {
              // 安全地生成标题
              let title = '未知题目'
              if (q.title) {
                title = q.title
              } else if (q.question && q.question.title) {
                title = q.question.title
              } else if (q.problem && q.problem.title) {
                title = q.problem.title
              } else if (q.questionId) {
                title = `题目ID: ${q.questionId}`
              } else if (q.problemId) {
                title = `BingOJ 编程题 - ${q.problemId}`
              }

              return {
                questionId: q.questionId || null,
                problemId: q.problemId || null,
                questionType: q.questionType || 'unknown',
                score: q.score || 0,
                title: title,
                question: q.question || null,
                problem: q.problem || null
              }
            })
          }

          this.isEditMode = true
          this.currentPaper = {
            ...paper,
            questions: (paper.questions || []).filter(q => {
              return q &&
                     typeof q === 'object' &&
                     (q.questionId || q.problemId) &&
                     q.questionType &&
                     typeof q.score === 'number'
            })
          }

          this.showCreateDialog = true
          this.loadQuestionBank()
        }
      } catch (error) {
        this.$message.error('加载试卷详情失败')
      } finally {
        // 清除编辑标志
        this.isEditing = false
      }
    },
    async savePaper() {
      if (this.paperForm.questions.length === 0) {
        this.$message.warning('请至少添加一道题目')
        return
      }

      this.saving = true
      try {
        const data = {
          title: this.paperForm.title,
          description: this.paperForm.description,
          isShared: this.paperForm.isShared ? 1 : 0,
          questions: this.paperForm.questions.map(q => ({
            questionId: q.questionId,
            problemId: q.problemId,
            questionType: q.questionType,
            score: q.score
          }))
        }

        let res
        if (this.isEditMode) {
          res = await classroomApi.updateExamPaper(this.currentPaper.id, data)
        } else {
          res = await classroomApi.createExamPaper(data)
        }

        if (res.data.code === 200) {
          this.$message.success(this.isEditMode ? '更新成功' : '创建成功')
          this.showCreateDialog = false
          this.loadPapers()
          // 重置表单
          this.paperForm = {
            title: '',
            description: '',
            isShared: false,
            questions: []
          }
        } else {
          this.$message.error(res.data.message || '操作失败')
        }
      } catch (error) {
        this.$message.error('操作失败')
      } finally {
        this.saving = false
      }
    },
    deletePaper(row) {
      this.$confirm('确定要删除这份试卷吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await classroomApi.deleteExamPaper(row.id)
          if (res.data.code === 200) {
            this.$message.success('删除成功')
            this.loadPapers()
          } else {
            this.$message.error(res.data.message || '删除失败')
          }
        } catch (error) {
          this.$message.error('删除失败')
        }
      })
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN')
    },
    getQuestionTypeTag(type) {
      if (!type) return 'info'
      const tagMap = {
        single_choice: 'success',
        multiple_choice: 'warning',
        judge: 'info',
        subjective: 'primary',
        programming: 'danger'
      }
      return tagMap[type] || 'info'
    },
    getQuestionTypeLabel(type) {
      if (!type) return '未知类型'
      const labelMap = {
        single_choice: '单选题',
        multiple_choice: '多选题',
        judge: '判断题',
        subjective: '主观题',
        programming: '编程题'
      }
      return labelMap[type] || type
    },
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },
    formatAnswer(question) {
      if (!question) return '-'

      try {
        const type = question?.type
        if (!type) {
          return '-'
        }

        switch (type) {
          case 'single_choice':
            // 单选题：答案直接是字母，如 "A"
            return question.answer || '-'

          case 'multiple_choice':
            // 多选题：答案是JSON数组字符串，如 '["A","B","C"]'
            if (!question.answer) return '-'
            try {
              const parsed = JSON.parse(question.answer)
              if (Array.isArray(parsed)) {
                return parsed.join('、')
              }
              return question.answer
            } catch (e) {
              return question.answer
            }

          case 'judge':
            // 判断题：统一仅按 true/false 显示
            const answer = question.answer
            if (!answer) return '-'

            const normalizedAnswer = String(answer).toLowerCase().trim()
            if (normalizedAnswer === 'true') {
              return '正确'
            } else if (normalizedAnswer === 'false') {
              return '错误'
            }
            return '-'

          case 'subjective':
            // 主观题：显示参考答案或提示
            return question.answer && question.answer !== '需人工评分' ? '有参考答案' : '需人工评分'

          default:
            return '-'
        }
      } catch (error) {
        return '-'
      }
    },
    toggleQuestionDetail(row) {
      if (!row) return
      this.$set(row, 'showDetail', !row.showDetail)
    },
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        const options = JSON.parse(optionsStr)
        if (Array.isArray(options)) {
          const labels = ['A', 'B', 'C', 'D', 'E', 'F']
          return options.map((opt, index) => ({
            label: labels[index] || String.fromCharCode(65 + index),
            text: opt
          }))
        }
      } catch (e) {
        // 解析失败
      }
      return []
    },
    // 解析题目标签
    parseQuestionTags(tags) {
      if (!tags) return []
      try {
        return JSON.parse(tags)
      } catch (e) {
        return []
      }
    },
    // 查看编程题详情
    viewProblemDetail(problem) {
      if (!problem) {
        this.$message.error('题目数据不存在')
        return
      }
      this.currentViewProblem = {
        problem: problem
      }
      this.showProblemDetailDialog = true
    },
    // 获取并查看编程题详情
    async fetchAndviewProblemDetail(problemId) {
      if (!problemId) {
        this.$message.error('题目ID不存在')
        return
      }

      this.fetchingViewProblem = true
      this.showProblemDetailDialog = true
      this.currentViewProblem = null

      // 强制 Vue 更新
      this.$forceUpdate()

      // 延迟检查 DOM
      this.$nextTick(() => {
        // 强制设置编程题详情对话框的 z-index
        setTimeout(() => {
          const dialogWrappers = document.querySelectorAll('.el-dialog__wrapper')

          // 编程题详情对话框应该是对话框列表中的最后一个
          if (dialogWrappers.length >= 2) {
            // 找到真正的编程题详情对话框（通过标题识别）
            let problemDetailWrapper = null
            let problemDetailDialog = null

            dialogWrappers.forEach((wrapper, i) => {
              const dialog = wrapper.querySelector('.el-dialog')
              const header = dialog ? dialog.querySelector('.el-dialog__header') : null
              const title = header ? header.querySelector('.el-dialog__title') : null
              const titleText = title ? title.textContent.trim() : ''

              if (titleText === '编程题详情') {
                problemDetailWrapper = wrapper
                problemDetailDialog = dialog
              }
            })

            // 如果没找到，使用最后一个作为后备
            if (!problemDetailWrapper) {
              problemDetailWrapper = dialogWrappers[dialogWrappers.length - 1]
              problemDetailDialog = problemDetailWrapper.querySelector('.el-dialog')
            }

            if (problemDetailWrapper && problemDetailDialog) {
              // 使用 CSS.setProperty 和 !important
              problemDetailWrapper.style.setProperty('z-index', '10000', 'important')
              if (problemDetailDialog) {
                problemDetailDialog.style.setProperty('z-index', '10000', 'important')
              }

              // 查找并设置遮罩层的 z-index
              const modals = document.querySelectorAll('.v-modal')
              modals.forEach(m => {
                m.style.setProperty('z-index', '9999', 'important')
              })
            }
          }
        }, 100)
      })

      try {
        const res = await api.getProblem(problemId, '0', undefined)

        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          this.currentViewProblem = res.data.data
        } else {
          this.$message.error('获取题目信息失败')
          this.showProblemDetailDialog = false
        }
      } catch (error) {
        console.error('Error fetching problem:', error)
        this.$message.error('获取题目信息失败')
        this.showProblemDetailDialog = false
      } finally {
        this.fetchingViewProblem = false
      }
    },
    // 解析编程题样例（从XML格式字符串中提取）
    parseExamples(examplesStr) {
      if (!examplesStr || typeof examplesStr !== 'string') {
        return []
      }

      // 匹配所有的 <input>...</input> 和 <output>...</output> 标签
      const inputRegex = /<input>([\s\S]*?)<\/input>/g
      const outputRegex = /<output>([\s\S]*?)<\/output>/g

      const inputs = []
      const outputs = []

      let match
      while ((match = inputRegex.exec(examplesStr)) !== null) {
        inputs.push(match[1].trim())
      }

      // 重置正则表达式的lastIndex
      inputRegex.lastIndex = 0

      while ((match = outputRegex.exec(examplesStr)) !== null) {
        outputs.push(match[1].trim())
      }

      // 组合成样例数组，将输入输出包裹在代码块中
      const examples = []
      const maxCount = Math.max(inputs.length, outputs.length)

      for (let i = 0; i < maxCount; i++) {
        // 使用代码块格式，保留换行
        const inputText = inputs[i] || ''
        const outputText = outputs[i] || ''

        examples.push({
          input: '```\n' + inputText + '\n```',
          output: '```\n' + outputText + '\n```'
        })
      }

      return examples
    },
    // 获取判题模式文本
    getJudgeModeText(mode) {
      const modeMap = {
        'default': '默认模式',
        'spj': '特殊判题 (SPJ)',
        'interactive': '交互式',
        'subtask': '子任务'
      }
      return modeMap[mode] || mode || '默认模式'
    }
  }
}
</script>

<style scoped>
.exam-paper-container {
  padding: 16px;
  max-width: 1320px;
  margin: 0 auto;
}

.exam-paper-editor-page {
  padding: 16px;
  max-width: 1380px;
  margin: 0 auto;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.editor-title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}

.editor-page-title {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.editor-toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.exam-paper-editor-card {
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.exam-paper-card {
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #ebeef5;
}

.editor-toolbar-actions .el-button,
.editor-footer-actions .el-button,
.header-right .el-button,
.quick-add-btn,
.add-programming-section .el-button {
  border-radius: 6px;
}

.editor-toolbar-actions .el-button--primary,
.editor-footer-actions .el-button--primary,
.header-right .el-button--primary,
.quick-add-btn.el-button--primary,
.add-programming-section .el-button--primary {
  background: #2f6ff6;
  border-color: #2f6ff6;
}

.editor-toolbar-actions .el-button--primary:hover,
.editor-footer-actions .el-button--primary:hover,
.header-right .el-button--primary:hover,
.quick-add-btn.el-button--primary:hover,
.add-programming-section .el-button--primary:hover {
  background: #285fdb;
  border-color: #285fdb;
}

.filter-bar {
  margin-bottom: 14px;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #fff;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.paper-title {
  font-weight: 600;
  color: #303133;
}

.paper-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

/* 题目选择区域 */
.questions-selector {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  padding: 14px 0;
  min-height: 460px;
}

/* 左侧题库面板 */
.question-bank-panel {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #ebeef5;
  overflow: hidden;
}

.panel-header {
  padding: 12px;
  background: #fff;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-weight: 600;
  color: #303133;
}

.filter-section {
  padding: 12px;
  border-bottom: 1px solid #f0f2f5;
}

.quick-add-section {
  padding: 10px 12px 12px;
  border-bottom: 1px solid #f0f2f5;
  background: #fcfcfd;
}

.quick-add-title {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.quick-add-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.quick-add-row .el-input {
  flex: 1;
}

.quick-add-btn {
  min-width: 86px;
}

.search-input {
  width: 100%;
  margin-bottom: 10px;
}

.filter-select {
  width: 100%;
}

.question-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  min-height: 300px;
}

.question-item {
  padding: 10px;
  margin-bottom: 8px;
  background: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 4px;
  transition: all 0.2s;
}

.question-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.question-header {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 5px;
  user-select: none;
}

.question-header:hover {
  background-color: #f5f7fa;
  border-radius: 4px;
}

.question-detail-content {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #e0e6ed;
}

.question-description {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 10px;
  font-size: 13px;
}

.question-options {
  margin: 10px 0;
}

.option-item {
  padding: 5px 0;
  color: #606266;
  font-size: 13px;
}

.option-label {
  font-weight: 600;
  color: #409eff;
  margin-right: 8px;
}

.option-text {
  color: #606266;
}

.question-answer-meta {
  padding: 8px;
  background-color: #f0f9ff;
  border-radius: 4px;
  font-size: 13px;
  display: inline-block;
}

.add-programming-section {
  padding: 15px;
  border-top: 1px solid #e0e6ed;
}

/* 右侧已选题目面板 */
.selected-questions-panel {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #ebeef5;
  overflow: hidden;
}

.selected-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  max-height: 500px;
}

.selected-item {
  background: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 10px;
  transition: all 0.3s;
}

.selected-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.item-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.item-order {
  font-weight: 600;
  margin-right: 8px;
  color: #409eff;
}

.item-title {
  color: #606266;
  line-height: 1.5;
}

.empty-selected {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.empty-selected i {
  font-size: 48px;
  margin-bottom: 10px;
}

.empty-selected p {
  margin: 5px 0;
}

.empty-selected .hint {
  font-size: 12px;
  color: #c0c4cc;
}

.item-content {
  padding-left: 28px;
  color: #606266;
  font-size: 13px;
}

.item-description {
  line-height: 1.6;
  margin-bottom: 8px;
}

.item-options {
  margin: 8px 0;
  padding-left: 10px;
}

.item-answer {
  padding: 6px 10px;
  background-color: #f0f9ff;
  border-radius: 4px;
  font-size: 12px;
  display: inline-block;
}

.item-answer .meta-label {
  font-weight: 600;
  color: #303133;
}

.item-answer .meta-value {
  color: #67C23A;
  font-weight: 500;
}

.programming-item {
  width: 100%;
}

.programming-item .item-title {
  color: #303133;
  font-weight: 600;
  margin-bottom: 8px;
  font-size: 14px;
}

.programming-item .problem-meta {
  margin-top: 8px;
  display: flex;
  gap: 8px;
}

/* 列表过渡动画 */
.list-enter-active,
.list-leave-active {
  transition: all 0.3s;
}

.list-enter,
.list-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

.list-move {
  transition: transform 0.3s;
}

/* 试卷详情 */
.paper-detail {
  padding: 20px;
}

.paper-detail h2 {
  margin-top: 0;
  color: #2c3e50;
}

.paper-detail .meta-info {
  margin: 10px 0;
  color: #606266;
}

.paper-detail .description {
  color: #909399;
  margin-top: 10px;
}

.paper-detail .questions-list {
  margin-top: 20px;
}

.paper-detail .question-item {
  padding: 15px;
  border-bottom: 1px solid #EBEEF5;
}

.paper-detail .question-item:last-child {
  border-bottom: none;
}

.paper-detail .question-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.paper-detail .question-order {
  font-weight: 600;
  margin-right: 8px;
}

.paper-detail .question-content {
  color: #606266;
  line-height: 1.6;
}

.markdown-body {
  line-height: 1.6;
}

/* 查看试卷对话框样式 */
.paper-detail {
  padding: 20px;
}

.paper-detail .paper-info {
  margin-bottom: 20px;
}

.paper-detail .paper-info h2 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #303133;
}

.paper-detail .meta-info {
  margin: 10px 0;
  color: #606266;
}

.paper-detail .description {
  color: #909399;
  margin-top: 10px;
}

.paper-detail .questions-list {
  max-height: 60vh;
  overflow-y: auto;
}

.paper-detail .question-item {
  padding: 15px;
  border-bottom: 1px solid #EBEEF5;
}

.paper-detail .question-item:last-child {
  border-bottom: none;
}

.paper-detail .question-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.paper-detail .question-order {
  font-weight: 600;
  margin-right: 8px;
}

.paper-detail .question-content {
  color: #606266;
  line-height: 1.6;
}

.paper-detail .question-title {
  margin-bottom: 10px;
  font-weight: 500;
}

.paper-detail .question-meta {
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 14px;
  display: inline-block;
  margin-top: 10px;
}

.paper-detail .meta-label {
  font-weight: 600;
  color: #303133;
}

.paper-detail .meta-value {
  color: #67C23A;
  font-weight: 500;
}

.paper-detail .programming-info {
  margin-top: 10px;
}

.paper-detail .problem-description {
  color: #606266;
  line-height: 1.6;
  margin: 10px 0;
}

/* 编程题预览 */
.problem-preview {
  margin: 20px 0;
}

.problem-preview h3 {
  margin-top: 0;
  color: #2c3e50;
}

.problem-preview .problem-description {
  color: #606266;
  line-height: 1.6;
  margin: 15px 0;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.problem-preview .problem-meta {
  margin: 10px 0;
}

.problem-preview .problem-section {
  margin-top: 15px;
  padding: 15px;
  background-color: #fafbfc;
  border-left: 3px solid #409eff;
  border-radius: 4px;
}

.problem-preview .problem-section h4 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #409eff;
  font-size: 14px;
}

@media screen and (max-width: 768px) {
  .questions-selector {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .exam-paper-container {
    padding: 10px;
  }

  .exam-paper-editor-page {
    padding: 10px;
  }

  .editor-toolbar {
    flex-direction: column;
    align-items: flex-start;
  }

  .editor-toolbar-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .quick-add-row {
    flex-direction: column;
    align-items: stretch;
  }

  .quick-add-btn {
    width: 100%;
  }
}

/* 编程题详情对话框样式 */
.problem-detail-view {
  overflow-y: auto;
  max-height: 60vh;
}

/* 重置所有 markdown 渲染元素的默认样式 */
.problem-detail-view >>> * {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.problem-detail-view h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #2c3e50;
  font-size: 20px;
}

.problem-detail-view .problem-meta-info {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

.problem-detail-view .problem-section {
  margin: 15px 0;
}

.problem-detail-view .problem-section h4 {
  margin: 0 0 10px 0;
  color: #409eff;
  font-size: 16px;
}

.problem-detail-view .problem-description,
.problem-detail-view .problem-io,
.problem-detail-view .problem-hint {
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
  line-height: 1.6;
  color: #606266;
  margin: 0;
}

/* 重置描述内的段落样式 */
.problem-detail-view .problem-description p,
.problem-detail-view .problem-io p,
.problem-detail-view .problem-hint p {
  margin: 0;
}

.problem-detail-view .problem-section h4 {
  margin: 0 0 10px 0;
  color: #409eff;
  font-size: 16px;
}

.problem-detail-view .problem-example {
  margin: 10px 0;
}

.problem-detail-view .example-item {
  padding: 15px;
  background-color: #f9fafc;
  border-radius: 4px;
}

.problem-detail-view .example-block {
  margin: 10px 0;
}

.problem-detail-view .example-label {
  font-weight: bold;
  color: #606266;
  margin-bottom: 5px;
}

.problem-detail-view .example-content {
  margin: 0;
  padding: 10px;
  background-color: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.6;
}

/* Markdown 代码块样式 */
.problem-detail-view .example-content pre {
  margin: 0;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
  overflow-x: auto;
}

.problem-detail-view .example-content code {
  font-family: 'Courier New', Courier, monospace;
  font-size: 13px;
  background-color: #f5f7fa;
  padding: 2px 4px;
  border-radius: 3px;
}

.problem-detail-view .problem-meta-info {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.problem-empty {
  text-align: center;
  padding: 40px 0;
  color: #909399;
}

.problem-empty i {
  font-size: 48px;
  margin-bottom: 10px;
  display: block;
}

.problem-empty span {
  font-size: 14px;
}

.empty-questions {
  text-align: center;
  padding: 40px 0;
  color: #909399;
}

.empty-questions i {
  font-size: 48px;
  margin-bottom: 10px;
  display: block;
}

.empty-questions p {
  font-size: 14px;
  margin: 0;
}

.paper-loading {
  text-align: center;
  padding: 40px 0;
  color: #909399;
}

.paper-loading i {
  font-size: 32px;
  margin-right: 10px;
}

/* 确保编程题详情对话框在最上层 */
.problem-detail-dialog-main {
  z-index: 10000 !important;
}

.problem-detail-dialog-main .v-modal {
  z-index: 9999 !important;
}

/* 确保编程题详情对话框在最上层 - 全局样式 */
.problem-detail-dialog-wrapper {
  z-index: 10000 !important;
}

.problem-detail-dialog-wrapper.el-dialog {
  z-index: 10000 !important;
}
</style>

<style>
/* 全局样式：确保编程题详情对话框在最上层 */
.problem-detail-dialog-wrapper {
  z-index: 10000 !important;
}

.problem-detail-dialog-wrapper.el-dialog__wrapper {
  z-index: 10000 !important;
}

.problem-detail-dialog-wrapper .el-dialog {
  z-index: 10000 !important;
}

/* 试卷对话框样式 */
.exam-paper-dialog .el-dialog__footer {
  text-align: right;
  padding: 15px 20px;
  border-top: 1px solid #e0e6ed;
}

.exam-paper-dialog .el-dialog__footer .el-button {
  margin-left: 10px;
}
</style>
