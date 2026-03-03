<template>
  <div class="exam-paper-admin-container">
    <el-card class="exam-paper-card">
      <div slot="header" class="card-header">
        <div class="header-left">
          <i class="el-icon-document-copy"></i>
          <span>试卷库管理</span>
        </div>
        <div class="header-right">
          <el-button icon="el-icon-refresh" @click="loadPapers" size="small">刷新</el-button>
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
            <div class="paper-title">
              {{ row.title }}
              <el-tag v-if="row.isShared === 1" size="mini" type="success" style="margin-left: 8px;">共享</el-tag>
              <el-tag v-else size="mini" type="info" style="margin-left: 8px;">私有</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="questionCount" label="题目数量" width="100" align="center"></el-table-column>
        <el-table-column prop="totalScore" label="总分" width="80" align="center"></el-table-column>

        <el-table-column prop="creator.username" label="创建者" width="150">
          <template slot-scope="{ row }">
            <div v-if="row.creator">
              {{ row.creator.username }}
            </div>
            <div v-else>-</div>
          </template>
        </el-table-column>

        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template slot-scope="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template slot-scope="{ row }">
            <el-button-group>
              <el-button size="mini" icon="el-icon-edit" @click.stop="editPaper(row)">编辑</el-button>
              <el-button
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

    <!-- 编辑试卷对话框（合并查看和编辑功能） -->
    <el-dialog :title="isEditMode ? '编辑试卷' : '查看试卷'" :visible.sync="showEditDialog" width="80%" top="5vh" @close="resetEditForm">
      <el-form v-if="editForm" :model="editForm" :rules="editRules" ref="editForm" label-width="100px">
        <el-form-item label="试卷标题" prop="title">
          <el-input v-model="editForm.title" placeholder="请输入试卷标题" :disabled="!isEditMode"></el-input>
        </el-form-item>
        <el-form-item label="试卷描述">
          <el-input type="textarea" v-model="editForm.description" :rows="3" placeholder="请输入试卷描述" :disabled="!isEditMode"></el-input>
        </el-form-item>
        <el-form-item label="共享状态">
          <el-switch
            v-model="editForm.isShared"
            active-text="共享"
            inactive-text="私有"
            :disabled="!isEditMode"
            @change="handleSharedChange"
          ></el-switch>
        </el-form-item>

        <el-divider>题目列表</el-divider>

        <!-- 添加题目按钮 -->
        <div v-if="isEditMode" style="margin-bottom: 15px;">
          <el-button icon="el-icon-plus" type="primary" @click="showAddQuestionPanel = !showAddQuestionPanel">
            {{ showAddQuestionPanel ? '收起题库' : '添加题目' }}
          </el-button>
          <el-tag style="margin-left: 10px;">已选 {{ editForm.questions.length }} 题，总分 {{ getTotalScore() }} 分</el-tag>
        </div>

        <!-- 题目选择面板（折叠） -->
        <div v-if="isEditMode && showAddQuestionPanel" class="questions-selector">
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
            </div>

            <div class="question-list" v-loading="questionsLoading">
              <el-table
                :data="questionBank"
                size="small"
                :show-header="false"
                :empty-text="questionBank.length === 0 ? '暂无题目' : '搜索题目'"
                style="font-size: 12px;"
              >
                <el-table-column>
                  <template slot-scope="{ row }">
                    <div class="question-item" v-if="row">
                      <div class="question-header" @click="toggleQuestionDetail(row)">
                        <el-tag size="mini" :type="getQuestionTypeTag(row.type)">
                          {{ getQuestionTypeLabel(row.type) }}
                        </el-tag>
                        <span style="margin-left: 10px;">{{ row.title }}</span>
                        <i :class="row.showDetail ? 'el-icon-arrow-up' : 'el-icon-arrow-down'" style="margin-left: auto; color: #909399;"></i>
                      </div>
                      <el-collapse-transition>
                        <div v-show="row.showDetail" class="question-detail-content">
                          <div class="question-description" v-html="renderMarkdown(row.content || row.title)"></div>
                          <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="question-options">
                            <div v-for="(option, index) in parseOptions(row.options)" :key="index" class="option-item">
                              <span class="option-label">{{ option.label }}.</span>
                              <span class="option-text" v-html="renderMarkdown(option.text)"></span>
                            </div>
                          </div>
                          <div class="question-answer-meta">
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

          <!-- 右侧：已选题目（可编辑模式下显示） -->
          <div class="selected-questions-panel">
            <div class="panel-header">
              <span class="panel-title">已选题目</span>
            </div>

            <div class="selected-list">
              <transition-group name="list">
                <template v-for="(q, index) in editForm.questions">
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
                          icon="el-icon-top"
                          size="mini"
                          type="primary"
                          :disabled="index === 0"
                          @click="moveQuestion(index, -1)"
                        ></el-button>
                        <el-button
                          icon="el-icon-bottom"
                          size="mini"
                          type="primary"
                          :disabled="index === editForm.questions.length - 1"
                          @click="moveQuestion(index, 1)"
                        ></el-button>
                        <el-button
                          type="danger"
                          icon="el-icon-delete"
                          size="mini"
                          @click="removeQuestion(index)"
                        ></el-button>
                      </el-button-group>
                    </div>
                  </div>
                  <div v-if="q && (q.questionId || q.problemId)" class="item-content" :key="'content-' + (q.questionId || q.problemId || index)">
                    <!-- 客观题 -->
                    <div v-if="q.question && (q.question.title || q.question.content)">
                      <div class="item-description" v-html="renderMarkdown(q.question.content || q.question.title)"></div>
                      <div v-if="q.question.type === 'single_choice' || q.question.type === 'multiple_choice'" class="item-options">
                        <div v-for="(option, index) in parseOptions(q.question.options)" :key="index" class="option-item">
                          <span class="option-label">{{ option.label }}.</span>
                          <span class="option-text" v-html="renderMarkdown(option.text)"></span>
                        </div>
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
          </div>
        </div>

        <!-- 题目列表（查看模式或折叠编辑模式） -->
        <div v-else class="questions-edit-list">
          <template v-for="(q, index) in editForm.questions">
            <div v-if="q && (q.questionId || q.problemId)" :key="index" class="question-edit-item">
              <div class="question-edit-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <el-tag size="small" :type="getQuestionTypeTag(q.questionType)">
                  {{ getQuestionTypeLabel(q.questionType) }}
                </el-tag>
                <el-input-number
                  v-model="q.score"
                  :min="1"
                  :max="100"
                  size="mini"
                  style="width: 120px; margin-left: 10px;"
                  :disabled="!isEditMode"
                ></el-input-number>
                <span>分</span>
                <el-tooltip v-if="isEditMode" content="修改分数只针对此题在试卷中的分值，不会修改题库中的分数" placement="top">
                  <i class="el-icon-question" style="margin-left: 5px; color: #909399; cursor: help;"></i>
                </el-tooltip>
                <el-button-group style="margin-left: auto;" v-if="isEditMode">
                  <el-button
                    icon="el-icon-top"
                    size="mini"
                    type="primary"
                    :disabled="index === 0"
                    @click="moveQuestion(index, -1)"
                  ></el-button>
                  <el-button
                    icon="el-icon-bottom"
                    size="mini"
                    type="primary"
                    :disabled="index === editForm.questions.length - 1"
                    @click="moveQuestion(index, 1)"
                  ></el-button>
                  <el-button
                    type="danger"
                    icon="el-icon-delete"
                    size="mini"
                    @click="removeQuestion(index)"
                  ></el-button>
                </el-button-group>
              </div>
              <div class="question-edit-content" v-if="q && (q.questionId || q.problemId)">
                <div v-if="q.question && (q.question.title || q.question.content)">
                  <div class="question-title" v-html="renderMarkdown(q.question.content || q.question.title)"></div>
                  <div v-if="q.question.type !== 'subjective'" class="question-meta">
                    <span class="meta-label">正确答案：</span>
                    <span class="meta-value">{{ formatAnswer(q.question) }}</span>
                  </div>
                </div>
                <div v-else-if="q.problemId">
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
                    <div class="question-description" v-html="renderMarkdown(q.problem.description)"></div>
                    <div class="question-meta">
                      <el-tag size="small" type="info">时间: {{ q.problem.timeLimit }}ms</el-tag>
                      <el-tag size="small" type="warning">内存: {{ q.problem.memoryLimit }}MB</el-tag>
                      <el-tag size="small" type="primary">判题模式: {{ getJudgeModeText(q.problem.judgeMode) }}</el-tag>
                    </div>
                  </div>
                  <div v-else>
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
          </template>
        </div>

        <el-form-item style="margin-top: 20px;">
          <el-alert
            v-if="isEditMode"
            title="提示"
            type="info"
            :closable="false"
            description="管理员可以添加题库中的题目或编程题，也可以调整题目顺序、修改分值、或移除题目。修改分值只针对此题在试卷中的分数，不会修改题库中的原始分数。如需修改题目内容，请到题库管理中操作。"
          >
          </el-alert>
          <el-alert
            v-else
            title="提示"
            type="warning"
            :closable="false"
            description="当前为查看模式，如需编辑试卷请点击右下角编辑按钮。"
          >
          </el-alert>
        </el-form-item>
      </el-form>

      <span slot="footer">
        <el-button @click="showEditDialog = false">关闭</el-button>
        <el-button v-if="!isEditMode" type="primary" @click="isEditMode = true">编辑</el-button>
        <template v-else>
          <el-button @click="isEditMode = false">取消编辑</el-button>
          <el-button type="primary" @click="confirmEdit" :loading="saving">保存</el-button>
        </template>
      </span>
    </el-dialog>

    <!-- 添加编程题对话框 -->
    <el-dialog title="添加编程题" :visible.sync="showAddProgrammingDialog" width="900px">
      <el-form :model="programmingForm" label-width="120px">
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
              <el-tag size="small" type="success">判题模式: {{ programmingProblemPreview.problem.judgeMode }}</el-tag>
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
          type="primary"
          @click="confirmAddProgrammingQuestion"
          :disabled="!programmingProblemPreview"
        >
          确定添加
        </el-button>
      </span>
    </el-dialog>

    <!-- 查看编程题详情对话框 -->
    <el-dialog
      title="编程题详情"
      :visible.sync="showProblemDetailDialog"
      width="900px"
      :close-on-click-modal="false"
      :modal="true"
      :modal-append-to-body="true"
      append-to-body
      destroy-on-close
      center
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
  </div>
</template>

<script>
import api from '@/api/classroom'
import problemApi from '@/common/api'
import { mapGetters } from 'vuex'
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'

const md = new MarkdownIt()
md.use(katex)

export default {
  name: 'ExamPaperAdmin',
  data() {
    return {
      loading: false,
      saving: false,
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
      showEditDialog: false,
      isEditMode: false,
      currentPaper: null,
      editForm: null,
      editRules: {
        title: [{ required: true, message: '请输入试卷标题', trigger: 'blur' }]
      },
      // 题库相关
      showAddQuestionPanel: false,
      questionBank: [],
      questionsLoading: false,
      questionFilters: {
        keyword: '',
        type: ''
      },
      questionCurrentPage: 1,
      questionPageSize: 10,
      questionBankTotal: 0,
      // 编程题相关
      showAddProgrammingDialog: false,
      programmingForm: {
        problemId: '',
        score: 10
      },
      programmingProblemPreview: null,
      fetchingProblem: false,
      // 查看编程题详情
      showProblemDetailDialog: false,
      currentViewProblem: null,
      fetchingViewProblem: false
    }
  },
  computed: {
    ...mapGetters(['userInfo'])
  },
  mounted() {
    this.loadPapers()
  },
  methods: {
    async loadPapers() {
      this.loading = true
      try {
        const res = await api.adminGetExamPaperList({
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
    async editPaper(row) {
      try {
        const res = await api.adminGetExamPaperDetail(row.id)
        if (res.data.code === 200) {
          const paper = res.data.data
          this.currentPaper = paper

          // 先过滤掉无效的题目
          const validQuestions = (paper.questions || []).filter(q => {
            const isValid = q && typeof q === 'object' && (q.questionId || q.problemId) && q.questionType && typeof q.score === 'number'
            return isValid
          })

          this.editForm = {
            id: paper.id,
            title: paper.title,
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
          this.showEditDialog = true
          // 加载题库数据
          this.loadQuestionBank()
        }
      } catch (error) {
        this.$message.error('加载试卷详情失败')
      }
    },
    resetEditForm() {
      this.editForm = null
      this.currentPaper = null
      this.isEditMode = false
      if (this.$refs.editForm) {
        this.$refs.editForm.clearValidate()
      }
    },
    removeQuestion(index) {
      this.$confirm('确定要移除这道题目吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.editForm.questions.splice(index, 1)
        // 确保题库列表保持完整，不重新加载题库
        // 如果题目不在当前页的题库列表中，可能需要搜索或翻页找到
      })
    },
    moveQuestion(index, direction) {
      const newIndex = index + direction
      if (newIndex < 0 || newIndex >= this.editForm.questions.length) return

      const temp = this.editForm.questions[index]
      this.editForm.questions.splice(index, 1)
      this.editForm.questions.splice(newIndex, 0, temp)
    },
    async confirmEdit() {
      this.$refs.editForm.validate(async (valid) => {
        if (!valid) return

        if (this.editForm.questions.length === 0) {
          this.$message.warning('试卷至少需要一道题目')
          return
        }

        this.saving = true
        try {
          const data = {
            title: this.editForm.title,
            description: this.editForm.description,
            isShared: this.editForm.isShared ? 1 : 0,
            questions: this.editForm.questions.map((q, index) => ({
              questionId: q.questionType === 'programming' ? null : q.questionId,
              problemId: q.questionType === 'programming' ? q.problemId : null,
              questionType: q.questionType,
              score: q.score
            }))
          }

          const res = await api.adminUpdateExamPaper(this.editForm.id, data)

          if (res.data.code === 200) {
            this.$message.success('更新成功')
            this.showEditDialog = false
            this.loadPapers()
          } else {
            this.$message.error(res.data.message || '更新失败')
          }
        } catch (error) {
          this.$message.error('更新失败')
        } finally {
          this.saving = false
        }
      })
    },
    handleSharedChange(newValue) {
      // El switch 的 v-model 会自动更新
    },
    deletePaper(row) {
      this.$confirm('确定要删除这份试卷吗？删除后创建者也无法使用。', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await api.adminDeleteExamPaper(row.id)
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
      const labelMap = {
        single_choice: '单选题',
        multiple_choice: '多选题',
        judge: '判断题',
        subjective: '主观题',
        programming: '编程题'
      }
      return labelMap[type] || type
    },
    formatAnswer(question) {
      if (!question) return '-'

      try {
        switch (question.type) {
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
            // 判断题：统一显示为 "正确" 或 "错误"
            // 数据库中可能有多种格式："正确"/"错误"、"对"/"错"、true/false
            const answer = question.answer
            if (!answer) return '-'

            // 标准化答案
            const normalizedAnswer = String(answer).toLowerCase().trim()
            if (['正确', '对', 'true', '1', '√', '✓'].some(v => normalizedAnswer === v.toLowerCase())) {
              return '正确'
            } else if (['错误', '错', 'false', '0', '×', '✗'].some(v => normalizedAnswer === v.toLowerCase())) {
              return '错误'
            }
            // 如果无法识别，返回原始值
            return answer

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
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },
    // 加载题库（管理员可以查看所有题目，包括私有的）
    async loadQuestionBank() {
      this.questionsLoading = true
      try {
        const res = await api.adminGetQuestionBank({
          page: this.questionCurrentPage,
          limit: this.questionPageSize,
          keyword: this.questionFilters.keyword || undefined,
          type: this.questionFilters.type || undefined
        })
        if (res && res.data && res.data.code === 200) {
          // 过滤掉 undefined 或 null 的题目
          this.questionBank = (res.data.data.questions || []).filter(q => {
            const isValid = q && q.id
            return isValid
          })
          this.questionBankTotal = res.data.data.total || 0
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
      const newQuestion = {
        questionId: question.id,
        problemId: null,
        questionType: question.type,
        score: question.score || this.getDefaultScore(question.type),
        title: question.title,
        question: question
      }
      this.editForm.questions.push(newQuestion)

      this.$message.success('添加成功')
    },
    isQuestionSelected(question) {
      if (!question || !question.id) return false
      return this.editForm.questions.some(q => q && q.questionId === question.id)
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
    getQuestionTitle(q) {
      // 检查 q 是否为 null 或 undefined
      if (!q) {
        return '未知题目'
      }

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
    },
    getTotalScore() {
      return this.editForm.questions.reduce((sum, q) => sum + (q.score || 0), 0)
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
    // 编程题相关
    async fetchProgrammingProblemInfo() {
      if (!this.programmingForm.problemId) {
        this.$message.warning('请输入题目ID')
        return
      }

      this.fetchingProblem = true
      try {
        const res = await problemApi.getProblem(this.programmingForm.problemId, '0', undefined)

        // 适配后端返回的数据结构
        // axios响应: res.data = {status: 200, data: {problem: {...}}, msg: "success"}
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
      if (this.editForm.questions.some(q => q.problemId === this.programmingForm.problemId)) {
        this.$message.warning('该编程题已添加')
        return
      }

      // 添加编程题
      this.editForm.questions.push({
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
    // 查看编程题详情
    viewProblemDetail(problem) {
      this.currentViewProblem = {
        problem: problem
      }
      this.showProblemDetailDialog = true
    },
    // 获取并查看编程题详情
    async fetchAndviewProblemDetail(problemId) {
      this.fetchingViewProblem = true
      this.showProblemDetailDialog = true
      this.currentViewProblem = null

      try {
        const res = await problemApi.getProblem(problemId, '0', undefined)

        // 适配后端返回的数据结构
        // axios响应: res.data = {status: 200, data: {problem: {...}}, msg: "success"}
        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          this.currentViewProblem = res.data.data
        } else {
          this.$message.error('获取题目信息失败')
          this.showProblemDetailDialog = false
        }
      } catch (error) {
        console.error('获取题目信息失败:', error)
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
.exam-paper-admin-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.filter-bar {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.paper-title {
  font-weight: 600;
  color: #303133;
}

.questions-edit-list {
  max-height: 60vh;
  overflow-y: auto;
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  padding: 10px;
}

.question-edit-item {
  background-color: #fff;
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  margin-bottom: 10px;
  padding: 15px;
  transition: all 0.3s;
}

.question-edit-item:hover {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.question-edit-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.question-number {
  font-weight: 600;
  margin-right: 8px;
}

.question-edit-content {
  padding-left: 30px;
  color: #606266;
}

.question-title {
  line-height: 1.6;
  margin-bottom: 10px;
}

.question-meta {
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 14px;
  display: inline-block;
}

.meta-label {
  font-weight: 600;
  color: #303133;
}

.meta-value {
  color: #67C23A;
  font-weight: 500;
}

.markdown-body {
  line-height: 1.6;
}

/* 题目选择区域 */
.questions-selector {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  padding: 20px 0;
  min-height: 500px;
}

/* 左侧题库面板 */
.question-bank-panel {
  display: flex;
  flex-direction: column;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
  overflow: hidden;
}

.panel-header {
  padding: 15px;
  background: #fff;
  border-bottom: 1px solid #e0e6ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-weight: 600;
  color: #303133;
}

.filter-section {
  padding: 15px;
  border-bottom: 1px solid #e0e6ed;
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
  padding: 15px;
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
  background: #f0f9ff;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
  overflow: hidden;
}

.selected-list {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
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
  font-weight: 500;
  margin-bottom: 8px;
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

  .exam-paper-admin-container {
    padding: 10px;
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

.problem-detail-view .problem-example {
  margin: 10px 0;
}

.problem-detail-view .example-item {
  padding: 15px;
  background-color: #f9fafc;
  border-radius: 4px;
  margin: 0;
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
</style>
