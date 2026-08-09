<template>
  <div class="exam-paper-admin-container">
    <el-card class="exam-paper-card" v-if="!showEditDialog">
      <div slot="header" class="card-header">
        <div class="header-left">
          <i class="el-icon-document-copy"></i>
          <span>试卷库管理</span>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="el-icon-plus" @click="openCreateDialog" size="small">创建试卷</el-button>
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
          <el-col :span="6">
            <el-select v-model="filters.isPublic" placeholder="主界面公开状态" clearable @change="loadPapers">
              <el-option label="全部" value=""></el-option>
              <el-option label="未公开" value="0"></el-option>
              <el-option label="已公开" value="1"></el-option>
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
        <el-table-column prop="id" label="试卷ID" width="90" align="center"></el-table-column>

        <el-table-column prop="title" label="试卷标题" min-width="200">
          <template slot-scope="{ row }">
            <div class="paper-title">
              {{ row.title }}
              <el-tag v-if="row.isShared === 1" size="mini" type="success" style="margin-left: 8px;">共享</el-tag>
              <el-tag v-else size="mini" type="info" style="margin-left: 8px;">私有</el-tag>
              <el-tag v-if="row.isPublic === 1" size="mini" type="danger" style="margin-left: 8px;">主界面公开</el-tag>
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

    <!-- 编辑试卷页面（合并查看和编辑功能） -->
    <div v-if="showEditDialog" class="exam-paper-editor-page">
      <div class="editor-toolbar">
        <div class="editor-title-wrap">
          <el-button icon="el-icon-arrow-left" @click="closeEditPage">返回试卷列表</el-button>
          <h3 class="editor-page-title">{{ dialogTitle }}</h3>
        </div>
        <div class="editor-toolbar-actions">
          <el-button v-if="isCreateMode" @click="closeEditPage">取消</el-button>
          <el-button v-if="!isEditMode" type="primary" @click="isEditMode = true">编辑</el-button>
          <template v-else>
            <el-button v-if="!isCreateMode" @click="isEditMode = false">取消编辑</el-button>
            <el-button type="primary" @click="confirmEdit" :loading="saving">{{ isCreateMode ? '创建' : '保存' }}</el-button>
          </template>
        </div>
      </div>

      <el-card class="exam-paper-editor-card" shadow="never">
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
        <el-form-item label="主界面公开">
          <el-switch
            v-model="editForm.isPublic"
            active-text="公开"
            inactive-text="不公开"
            :disabled="!isEditMode"
          ></el-switch>
        </el-form-item>

        <el-divider>题目列表</el-divider>

        <div v-if="isEditMode" class="selected-questions-toolbar">
          <el-button icon="el-icon-plus" type="primary" @click="openQuestionSelectorDialog">添加题目</el-button>
          <el-button
            type="success"
            size="small"
            icon="el-icon-plus"
            @click="showAddProgrammingDialog = true"
          >
            添加编程题
          </el-button>
          <el-tag size="small" type="info">已选 {{ selectedEditQuestions.length }} 题</el-tag>
          <el-tag size="small" type="success">总分 {{ getTotalScore() }} 分</el-tag>
        </div>

        <div v-if="selectedEditQuestions.length > 0" class="questions-edit-layout">
          <div class="question-index-panel" v-if="selectedEditQuestions.length > 1">
            <div class="question-index-title">题号导航</div>
            <el-button
              v-for="(q, index) in selectedEditQuestions"
              v-if="q && (q.questionId || q.problemId)"
              :key="'index-nav-' + index"
              size="mini"
              plain
              @click="scrollToEditQuestion(index)"
            >
              {{ index + 1 }}
            </el-button>
          </div>

          <div class="questions-edit-list" ref="editQuestionScroll">
            <template v-for="(q, index) in selectedEditQuestions">
              <div v-if="q && (q.questionId || q.problemId)" :key="index" class="question-edit-item" ref="editQuestionItem">
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
                      :disabled="index === selectedEditQuestions.length - 1"
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
                    <div class="question-title markdown-body" v-html="renderMarkdown(q.question.content || q.question.title)" v-highlight></div>
                    <div v-if="q.question.type === 'single_choice' || q.question.type === 'multiple_choice'" class="question-options">
                      <div v-for="(option, optionIndex) in parseOptions(q.question.options)" :key="`question-option-${index}-${optionIndex}`" class="option-item">
                        <span class="option-label">{{ option.label }}.</span>
                        <span class="option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                      </div>
                    </div>
                    <div v-if="q.question.type === 'composite'" class="composite-question-block">
                      <div
                        v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(q.question.options)"
                        :key="subQuestion.id || subIndex"
                        class="composite-sub-question"
                      >
                        <div class="composite-sub-header">
                          <span class="composite-sub-title">子题 {{ subIndex + 1 }}</span>
                          <el-tag size="mini" type="warning">{{ Number(subQuestion.score || 0) }} 分</el-tag>
                        </div>
                        <div class="markdown-body composite-sub-content" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
                        <div class="question-options" v-if="subQuestion.options && subQuestion.options.length">
                          <div
                            v-for="(option, optionIndex) in parseOptions(subQuestion.options)"
                            :key="`${subQuestion.id || subIndex}_${optionIndex}`"
                            class="option-item"
                          >
                            <span class="option-label">{{ option.label }}.</span>
                            <span class="option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
                          </div>
                        </div>
                        <div class="question-meta answer-info compact-answer-info">
                          <span class="meta-label">正确答案：</span>
                          <span class="meta-value">{{ getCompositeCorrectAnswer(q.question.answer, subQuestion.id, subIndex) }}</span>
                        </div>
                      </div>
                    </div>
                    <div class="selector-question-meta">
                      <el-tag v-if="q.question.course" size="mini" type="warning">
                        <i class="el-icon-collection"></i> {{ q.question.course }}
                      </el-tag>
                      <el-tag
                        v-for="(tag, idx) in parseQuestionTags(q.question.tags)"
                        :key="`question-tag-${index}-${idx}`"
                        size="mini"
                        type="info"
                      >
                        {{ tag }}
                      </el-tag>
                    </div>
                    <div v-if="q.question.type !== 'composite'" class="question-meta answer-info compact-answer-info">
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
                        <el-tag size="small" type="success">难度: {{ getDifficultyName(q.problem.difficulty) }}</el-tag>
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
        </div>

        <div v-else class="empty-selected">
          <i class="el-icon-document"></i>
          <p>暂未选择题目</p>
          <p class="hint">点击上方“添加题目”按钮</p>
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

        <div class="editor-footer-actions">
          <el-button @click="closeEditPage">{{ isCreateMode ? '取消' : '关闭' }}</el-button>
          <el-button v-if="!isEditMode" type="primary" @click="isEditMode = true">编辑</el-button>
          <template v-else>
            <el-button v-if="!isCreateMode" @click="isEditMode = false">取消编辑</el-button>
            <el-button type="primary" @click="confirmEdit" :loading="saving">{{ isCreateMode ? '创建' : '保存' }}</el-button>
          </template>
        </div>
      </el-card>
    </div>

    <el-dialog
      title="添加题目"
      :visible.sync="showQuestionSelectorDialog"
      width="1200px"
      append-to-body
      class="question-selector-dialog-wrapper"
    >
      <div class="question-selector-dialog">
        <div class="filter-section selector-filter-section">
          <el-row :gutter="10">
            <el-col :xs="24" :sm="12" :md="8">
              <el-input
                v-model="questionFilters.keyword"
                placeholder="按题目标题搜索"
                prefix-icon="el-icon-search"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              >
                <el-button slot="append" icon="el-icon-search" @click="loadQuestionBank"></el-button>
              </el-input>
            </el-col>
            <el-col :xs="24" :sm="12" :md="6">
              <el-input
                v-model="questionFilters.questionId"
                placeholder="按题目ID搜索"
                prefix-icon="el-icon-ticket"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              ></el-input>
            </el-col>
            <el-col :xs="24" :sm="12" :md="5">
              <el-select
                v-model="questionFilters.type"
                placeholder="题型筛选"
                clearable
                @change="loadQuestionBank"
                style="width: 100%;"
              >
                <el-option label="全部题型" value=""></el-option>
                <el-option label="单选题" value="single_choice"></el-option>
                <el-option label="多选题" value="multiple_choice"></el-option>
                <el-option label="判断题" value="judge"></el-option>
                <el-option label="填空题" value="fill_blank"></el-option>
                <el-option label="主观题" value="subjective"></el-option>
                <el-option label="组合题" value="composite"></el-option>
              </el-select>
            </el-col>
            <el-col :xs="24" :sm="12" :md="5">
              <el-select
                v-model="questionFilters.course"
                placeholder="课程筛选"
                clearable
                @change="loadQuestionBank"
                style="width: 100%;"
              >
                <el-option label="全部课程" value=""></el-option>
                <el-option
                  v-for="course in commonCourses"
                  :key="course"
                  :label="course"
                  :value="course"
                />
              </el-select>
            </el-col>
          </el-row>
          <el-row :gutter="10" style="margin-top: 10px;">
            <el-col :xs="24" :sm="10">
              <el-input
                v-model="questionFilters.tag"
                placeholder="按标签筛选"
                prefix-icon="el-icon-price-tag"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              ></el-input>
            </el-col>
            <el-col :xs="24" :sm="8">
              <el-select
                v-model="questionFilters.sortKey"
                placeholder="排序方式"
                @change="handleQuestionSortChange"
                style="width: 100%;"
              >
                <el-option label="最新创建" value="create_desc"></el-option>
                <el-option label="最早创建" value="create_asc"></el-option>
                <el-option label="ID升序" value="id_asc"></el-option>
                <el-option label="ID降序" value="id_desc"></el-option>
              </el-select>
            </el-col>
            <el-col :xs="24" :sm="6" class="selector-tools">
              <el-checkbox v-model="questionFilters.onlyUnselected">仅看未添加</el-checkbox>
              <el-button
                type="text"
                icon="el-icon-refresh"
                @click="loadQuestionBank"
                :loading="questionsLoading"
              >
                刷新题库
              </el-button>
              <el-button type="text" @click="resetQuestionSelectorFilters(); loadQuestionBank()">重置筛选</el-button>
            </el-col>
          </el-row>
        </div>

        <div class="selector-summary">
          <el-tag size="small" type="info">当前显示 {{ filteredQuestionBank.length }} / 本页 {{ questionBank.length }} / 总计 {{ questionBankTotal }} 题</el-tag>
          <el-tag size="small" type="success">已选 {{ selectedEditQuestions.length }} 题</el-tag>
        </div>

        <div class="question-selector-list" v-loading="questionsLoading">
          <el-empty
            v-if="!questionsLoading && questionBank.length > 0 && filteredQuestionBank.length === 0"
            description="当前页题目均已添加或被筛选条件过滤"
          ></el-empty>
          <div
            v-for="row in filteredQuestionBank"
            :key="'selector-question-' + row.id"
            class="selector-question-item"
          >
            <div class="selector-question-header">
              <div class="selector-header-main">
                <el-tag size="mini" :type="getQuestionTypeTag(row.type)">
                  {{ getQuestionTypeLabel(row.type) }}
                </el-tag>
                <el-tag size="mini" type="info" style="margin-left: 6px;">ID: {{ row.id }}</el-tag>
                <span class="selector-question-title">{{ row.title }}</span>
                <el-tag v-if="row.isShared === 1" size="mini" type="success">共享</el-tag>
                <el-tag v-else size="mini" type="info">私有</el-tag>
                <el-tag v-if="row.creator && row.creator.username" size="mini" type="warning">
                  创建者: {{ row.creator.username }}
                </el-tag>
              </div>
              <el-button
                size="mini"
                :type="isQuestionSelected(row) ? 'danger' : 'primary'"
                plain
                @click="toggleObjectiveQuestion(row)"
              >
                {{ isQuestionSelected(row) ? '移除' : '添加' }}
              </el-button>
            </div>

            <div class="question-description markdown-body" v-html="renderMarkdown(row.content || row.title)" v-highlight></div>
            <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="question-options">
              <div v-for="(option, index) in parseOptions(row.options)" :key="index" class="option-item">
                <span class="option-label">{{ option.label }}.</span>
                <span class="option-text markdown-body" v-html="renderMarkdown(option.text)" v-highlight></span>
              </div>
            </div>
            <div class="selector-question-meta">
              <el-tag v-if="row.course" size="mini" type="warning">
                <i class="el-icon-collection"></i> {{ row.course }}
              </el-tag>
              <el-tag
                v-for="(tag, idx) in parseQuestionTags(row.tags)"
                :key="'selector-tag-' + row.id + '-' + idx"
                size="mini"
                type="info"
              >
                {{ tag }}
              </el-tag>
            </div>
            <div class="question-answer-meta answer-info compact-answer-info" v-if="row && row.type">
              <span class="meta-label">正确答案：</span>
              <span class="meta-value">{{ formatAnswer(row) }}</span>
            </div>
          </div>

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
      </div>
      <span slot="footer">
        <el-button @click="showQuestionSelectorDialog = false">关闭</el-button>
      </span>
    </el-dialog>

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
          v-if="programmingInputMode === 'manual'"
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
              <span v-html="renderMarkdown('**难度:** ' + getDifficultyName(currentViewProblem.problem.difficulty))"></span>
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
  </div>
</template>

<script>
import api from '@/api/classroom'
import problemApi from '@/common/api'
import { mapGetters } from 'vuex'
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
  name: 'ExamPaperAdmin',
  data() {
    return {
      loading: false,
      saving: false,
      papers: [],
      filters: {
        keyword: '',
        isShared: '',
        isPublic: ''
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
      showQuestionSelectorDialog: false,
      questionBank: [],
      questionsLoading: false,
      questionFilters: {
        keyword: '',
        questionId: '',
        type: '',
        course: '',
        tag: '',
        sortKey: 'create_desc',
        onlyUnselected: false
      },
      questionCurrentPage: 1,
      questionPageSize: 10,
      questionBankTotal: 0,
      quickAddQuestionId: '',
      quickAddQuestionLoading: false,
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
      showAddProgrammingDialog: false,
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
    isCreateMode() {
      return this.isEditMode && this.editForm && !this.editForm.id
    },
    dialogTitle() {
      if (this.isCreateMode) {
        return '创建试卷'
      }
      return this.isEditMode ? '编辑试卷' : '查看试卷'
    },
    filteredQuestionBank() {
      if (!Array.isArray(this.questionBank)) {
        return []
      }
      if (!this.questionFilters.onlyUnselected || !this.editForm) {
        return this.questionBank
      }
      return this.questionBank.filter(q => !this.isQuestionSelected(q))
    },
    selectedEditQuestions() {
      return this.editForm && Array.isArray(this.editForm.questions)
        ? this.editForm.questions
        : []
    }
  },
  mounted() {
    this.loadPapers()
    this.loadProblemTagsAndClassification()
  },
  watch: {
    showEditDialog(newVal, oldVal) {
      if (oldVal === true && newVal === false) {
        this.resetEditForm()
      }
    }
  },
  methods: {
    async loadPapers() {
      this.loading = true
      try {
        const res = await api.adminGetExamPaperList({
          page: this.pagination.currentPage,
          limit: this.pagination.pageSize,
          keyword: this.filters.keyword || undefined,
          isShared: this.filters.isShared,
          isPublic: this.filters.isPublic
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
    openQuestionSelectorDialog() {
      if (!this.isEditMode) {
        return
      }
      this.showQuestionSelectorDialog = true
      this.questionCurrentPage = 1
      this.loadQuestionBank()
    },
    resetQuestionSelectorFilters() {
      this.questionFilters = {
        keyword: '',
        questionId: '',
        type: '',
        course: '',
        tag: '',
        sortKey: 'create_desc',
        onlyUnselected: false
      }
      this.questionCurrentPage = 1
    },
    getQuestionSortParams() {
      const sortKey = this.questionFilters.sortKey || 'create_desc'
      switch (sortKey) {
        case 'id_asc':
          return { sortBy: 'id', sortOrder: 'asc' }
        case 'id_desc':
          return { sortBy: 'id', sortOrder: 'desc' }
        case 'create_asc':
          return { sortBy: 'createTime', sortOrder: 'asc' }
        case 'create_desc':
        default:
          return { sortBy: 'createTime', sortOrder: 'desc' }
      }
    },
    handleQuestionSortChange() {
      this.questionCurrentPage = 1
      this.loadQuestionBank()
    },
    openCreateDialog() {
      this.currentPaper = null
      this.isEditMode = true
      this.showAddQuestionPanel = false
      this.showQuestionSelectorDialog = false
      this.resetQuestionSelectorFilters()
      this.questionCurrentPage = 1
      this.quickAddQuestionId = ''
      this.quickAddQuestionLoading = false
      this.editForm = {
        id: null,
        title: '',
        description: '',
        isShared: false,
        isPublic: false,
        questions: []
      }
      this.showEditDialog = true
      this.loadQuestionBank()
      this.$nextTick(() => {
        if (this.$refs.editForm) {
          this.$refs.editForm.clearValidate()
        }
      })
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
            isPublic: paper.isPublic === 1,
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
          this.showQuestionSelectorDialog = false
          // 加载题库数据
          this.loadQuestionBank()
          this.hydrateProgrammingProblemDetails(this.selectedEditQuestions)
        }
      } catch (error) {
        this.$message.error('加载试卷详情失败')
      }
    },
    resetEditForm() {
      this.editForm = null
      this.currentPaper = null
      this.isEditMode = false
      this.showAddQuestionPanel = false
      this.showQuestionSelectorDialog = false
      this.resetQuestionSelectorFilters()
      this.quickAddQuestionId = ''
      this.quickAddQuestionLoading = false
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
        const questions = this.ensureEditFormQuestions()
        if (!questions) return
        questions.splice(index, 1)
        // 确保题库列表保持完整，不重新加载题库
        // 如果题目不在当前页的题库列表中，可能需要搜索或翻页找到
      })
    },
    moveQuestion(index, direction) {
      const questions = this.ensureEditFormQuestions()
      if (!questions) return
      const newIndex = index + direction
      if (newIndex < 0 || newIndex >= questions.length) return

      const temp = questions[index]
      questions.splice(index, 1)
      questions.splice(newIndex, 0, temp)
    },
    async confirmEdit() {
      this.$refs.editForm.validate(async (valid) => {
        if (!valid) return

        const questions = this.selectedEditQuestions
        if (questions.length === 0) {
          this.$message.warning('试卷至少需要一道题目')
          return
        }

        this.saving = true
        const isCreateOperation = !this.editForm.id
        try {
          const data = {
            title: this.editForm.title,
            description: this.editForm.description,
            isShared: this.editForm.isShared ? 1 : 0,
            isPublic: this.editForm.isPublic ? 1 : 0,
            questions: questions.map(q => ({
              questionId: q.questionType === 'programming' ? null : q.questionId,
              problemId: q.questionType === 'programming' ? q.problemId : null,
              questionType: q.questionType,
              score: q.score
            }))
          }

          let res
          if (!isCreateOperation) {
            res = await api.adminUpdateExamPaper(this.editForm.id, data)
          } else {
            // 管理员创建试卷时复用教师端创建接口
            res = await api.createExamPaper({
              title: data.title,
              description: data.description,
              isShared: data.isShared,
              questions: data.questions
            })
          }

          if (res.data.code !== 200) {
            this.$message.error(res.data.message || (isCreateOperation ? '创建失败' : '更新失败'))
            return
          }

          // 创建时如果勾选了主界面公开，补一次管理员更新
          if (isCreateOperation && this.editForm.isPublic) {
            const createdPaperId = res && res.data && res.data.data ? res.data.data.id : null
            if (createdPaperId) {
              const publishRes = await api.adminUpdateExamPaper(createdPaperId, data)
              if (publishRes.data.code !== 200) {
                this.$message.warning('试卷已创建，但设置主界面公开失败，请在列表中编辑试卷后重试')
              }
            } else {
              this.$message.warning('试卷已创建，但未拿到试卷ID，无法自动设置主界面公开')
            }
          }

          this.$message.success(isCreateOperation ? '创建成功' : '更新成功')
          this.showEditDialog = false
          this.loadPapers()
          if (isCreateOperation) {
            this.questionCurrentPage = 1
          }
        } catch (error) {
          this.$message.error(isCreateOperation ? '创建失败' : '更新失败')
        } finally {
          this.saving = false
        }
      })
    },
    closeEditPage() {
      this.showEditDialog = false
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
        fill_blank: 'success',
        subjective: 'primary',
        composite: 'danger',
        programming: 'danger'
      }
      return tagMap[type] || 'info'
    },
    getQuestionTypeLabel(type) {
      const labelMap = {
        single_choice: '单选题',
        multiple_choice: '多选题',
        judge: '判断题',
        fill_blank: '填空题',
        subjective: '主观题',
        composite: '组合题',
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

          case 'fill_blank':
            if (!question.answer) return '-'
            try {
              const parsed = typeof question.answer === 'string' ? JSON.parse(question.answer) : question.answer
              if (Array.isArray(parsed)) {
                const answers = parsed.map(item => String(item || '').trim()).filter(Boolean)
                return answers.length > 0 ? answers.join(' / ') : '-'
              }
              return String(question.answer || '').trim() || '-'
            } catch (e) {
              return String(question.answer || '').trim() || '-'
            }

          case 'subjective':
            return question.answer || '需人工评分'

          case 'composite':
            return '组合题（按子题判分）'

          default:
            return '-'
        }
      } catch (error) {
        return '-'
      }
    },
    normalizeQuestionImagePath(rawText) {
      let content = String(rawText || '')
      if (!content) return content
      content = content.replace(
        /(<img\b[^>]*\bsrc=["'])(uploads\/classroom\/[^"']+)(["'][^>]*>)/gi,
        '$1/$2$3'
      )
      content = content.replace(
        /(!\[[^\]]*]\()(uploads\/classroom\/[^)\s]+)(\))/gi,
        '$1/$2$3'
      )
      return content
    },
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(this.normalizeQuestionImagePath(text))
    },
    // 加载题库（管理员可以查看所有题目，包括私有的）
    async loadQuestionBank() {
      this.questionsLoading = true
      try {
        const questionId = String(this.questionFilters.questionId || '').trim()
        const keyword = String(this.questionFilters.keyword || '').trim()
        const searchKeyword = questionId || keyword
        const searchField = questionId ? 'id' : 'title'
        const sortParams = this.getQuestionSortParams()
        const res = await api.adminGetQuestionBank({
          page: this.questionCurrentPage,
          limit: this.questionPageSize,
          keyword: searchKeyword || undefined,
          searchField: searchKeyword ? searchField : undefined,
          type: this.questionFilters.type || undefined,
          course: this.questionFilters.course || undefined,
          tag: this.questionFilters.tag || undefined,
          sortBy: sortParams.sortBy,
          sortOrder: sortParams.sortOrder
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
    toggleObjectiveQuestion(question) {
      if (!question || !question.id || !this.editForm) {
        return
      }
      if (this.isQuestionSelected(question)) {
        this.removeObjectiveQuestion(question.id)
        this.$message.success('已移除题目')
      } else {
        this.addQuestion(question)
      }
    },
    ensureEditFormQuestions() {
      if (!this.editForm) {
        return null
      }
      if (!Array.isArray(this.editForm.questions)) {
        this.$set(this.editForm, 'questions', [])
      }
      return this.editForm.questions
    },
    removeObjectiveQuestion(questionId) {
      const questions = this.ensureEditFormQuestions()
      if (!questions) {
        return
      }
      const index = questions.findIndex(
        q => q && q.questionId === questionId
      )
      if (index !== -1) {
        questions.splice(index, 1)
      }
    },
    isObjectiveQuestionType(type) {
      return ['single_choice', 'multiple_choice', 'judge', 'fill_blank', 'composite'].includes(type)
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
      if (this.selectedEditQuestions.some(q => q && String(q.questionId) === questionId)) {
        this.$message.warning('该题目已添加')
        return
      }

      this.quickAddQuestionLoading = true
      try {
        const res = await api.getQuestionDetail(questionId)
        if (!res || !res.data || res.data.code !== 200 || !res.data.data) {
          this.$message.error('未找到该题目')
          return
        }
        const question = res.data.data
        if (!this.isObjectiveQuestionType(question.type)) {
          this.$message.warning('该题不是客观题，仅支持单选/多选/判断/填空/组合题')
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
      const questions = this.ensureEditFormQuestions()
      if (!questions) {
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
      questions.push(newQuestion)

      this.$message.success('添加成功')
    },
    isQuestionSelected(question) {
      if (!question || !question.id) return false
      return this.selectedEditQuestions.some(q => q && q.questionId === question.id)
    },
    getDefaultScore(type) {
      const scores = {
        single_choice: 2,
        multiple_choice: 5,
        judge: 1,
        fill_blank: 2,
        subjective: 5,
        composite: 10
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
    getRequestErrorMessage(error, fallback) {
      return error?.data?.msg ||
        error?.data?.message ||
        error?.response?.data?.msg ||
        error?.response?.data?.message ||
        fallback
    },
    getTotalScore() {
      return this.selectedEditQuestions.reduce((sum, q) => sum + (q.score || 0), 0)
    },
    toggleQuestionDetail(row) {
      if (!row) return
      this.$set(row, 'showDetail', !row.showDetail)
    },
    parseOptions(optionsInput) {
      if (!optionsInput) return []
      try {
        const options = Array.isArray(optionsInput)
          ? optionsInput
          : JSON.parse(optionsInput)
        if (Array.isArray(options)) {
          const labels = ['A', 'B', 'C', 'D', 'E', 'F']
          return options.map((opt, index) => ({
            label: labels[index] || String.fromCharCode(65 + index),
            text: this.stripOptionPrefix(opt)
          }))
        }
      } catch (e) {
        // 解析失败
      }
      return []
    },
    stripOptionPrefix(optionText) {
      return String(optionText || '').replace(/^[A-Z]\.\s*/, '')
    },
    parseCompositeSubQuestions(optionsInput) {
      if (!optionsInput) return []
      try {
        const parsed = Array.isArray(optionsInput)
          ? optionsInput
          : JSON.parse(optionsInput)
        if (!Array.isArray(parsed)) return []
        return parsed.map((subQuestion, index) => ({
          id: String(subQuestion.id || `sq_${index + 1}`),
          content: subQuestion.content || '',
          options: Array.isArray(subQuestion.options)
            ? subQuestion.options
            : (Array.isArray(subQuestion.choiceOptions) ? subQuestion.choiceOptions : []),
          score: Number(subQuestion.score || subQuestion.subScore || 0)
        }))
      } catch (e) {
        return []
      }
    },
    parseCompositeAnswerMap(answerInput) {
      if (!answerInput) return {}
      try {
        const parsed = typeof answerInput === 'string' ? JSON.parse(answerInput) : answerInput
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
          return parsed
        }
      } catch (e) {
        // ignore parse failure
      }
      return {}
    },
    getCompositeCorrectAnswer(answerInput, subQuestionId, subIndex) {
      const answerMap = this.parseCompositeAnswerMap(answerInput)
      const candidates = [
        String(subQuestionId || ''),
        String(subIndex + 1),
        String(subIndex),
        `sub_${subIndex + 1}`
      ]
      for (const key of candidates) {
        if (key && Object.prototype.hasOwnProperty.call(answerMap, key)) {
          const value = String(answerMap[key] || '').trim()
          if (value) return value
        }
      }
      return '-'
    },
    scrollToEditQuestion(index) {
      this.$nextTick(() => {
        const container = this.$refs.editQuestionScroll
        const itemRefs = this.$refs.editQuestionItem
        const target = Array.isArray(itemRefs) ? itemRefs[index] : itemRefs
        if (!container || !target) return
        const targetTop = target.offsetTop - container.offsetTop
        container.scrollTo({
          top: Math.max(targetTop - 8, 0),
          behavior: 'smooth'
        })
      })
    },
    async hydrateProgrammingProblemDetails(questions) {
      if (!Array.isArray(questions) || questions.length === 0) return
      const tasks = questions
        .filter(q => q && q.problemId && (!q.problem || !q.problem.title))
        .map(async q => {
          try {
            const res = await problemApi.getProblem(q.problemId, '0', undefined, true)
            if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
              this.$set(q, 'problem', res.data.data.problem)
            }
          } catch (e) {
            // ignore single problem fetch errors
          }
        })
      if (tasks.length > 0) {
        await Promise.all(tasks)
      }
    },
    parseQuestionTags(tags) {
      if (!tags) return []
      try {
        return JSON.parse(tags)
      } catch (e) {
        return []
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
        const res = await problemApi.getProblem(this.programmingForm.problemId, '0', undefined, true)

        // 适配后端返回的数据结构
        // axios响应: res.data = {status: 200, data: {problem: {...}}, msg: "success"}
        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          this.programmingProblemPreview = res.data.data
        } else {
          this.$message.error('获取题目信息失败')
        }
      } catch (error) {
        this.programmingProblemPreview = null
        this.$message.error(this.getRequestErrorMessage(error, '获取题目信息失败'))
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
      if (this.selectedEditQuestions.some(q => q.problemId === this.programmingForm.problemId)) {
        this.$message.warning('该编程题已添加')
        return
      }
      const questions = this.ensureEditFormQuestions()
      if (!questions) return

      // 添加编程题
      questions.push({
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
        const res = await problemApi.getProblemTagsAndClassification('ME')
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
        this.filteredProblemsCurrentPage = 1
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
    // 根据已选标签加载题目列表
    async loadProblemsByTags() {
      if (this.selectedProblemTagIds.length === 0) {
        this.filteredProblemsByTag = []
        this.filteredProblemsTotal = 0
        return
      }

      try {
        const res = await problemApi.getProblemList({
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
    // 标签题目分页改变
    async handleFilteredProblemsPageChange(page) {
      this.filteredProblemsCurrentPage = page
      await this.loadProblemsByTags()
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
        const res = await problemApi.getProblem(problem.problemId, '0', undefined, true)
        if (res && res.status === 200 && res.data && res.data.data) {
          this.tagProblemDetail = res.data.data
        }
      } catch (error) {
        console.error('获取题目详情失败:', error)
        this.tagProblemDetail = null
        this.showTagProblemDetailDialog = false
        this.$message.error(this.getRequestErrorMessage(error, '获取题目详情失败'))
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
        const res = await problemApi.getProblem(this.programmingForm.problemId, '0', undefined, true)

        if (res && res.status === 200 && res.data && res.data.data && res.data.data.problem) {
          const problemData = res.data.data

          // 检查是否已添加
          if (this.selectedEditQuestions.some(q => q.problemId === this.programmingForm.problemId)) {
            this.$message.warning('该编程题已添加')
            return
          }
          const questions = this.ensureEditFormQuestions()
          if (!questions) return

          // 直接添加编程题
          questions.push({
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
        this.programmingProblemPreview = null
        this.$message.error(this.getRequestErrorMessage(error, '获取题目信息失败'))
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
        const res = await problemApi.getProblem(problemId, '0', undefined, true)

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
        this.$message.error(this.getRequestErrorMessage(error, '获取题目信息失败'))
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
  padding: 16px;
  max-width: 1320px;
  margin: 0 auto;
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

.editor-footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #ebeef5;
}

.exam-paper-card,
.exam-paper-editor-card {
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

.selected-questions-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.empty-selected {
  min-height: 220px;
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.empty-selected i {
  font-size: 44px;
  margin-bottom: 10px;
}

.empty-selected p {
  margin: 4px 0;
}

.empty-selected .hint {
  font-size: 12px;
  color: #c0c4cc;
}

.question-selector-dialog {
  display: flex;
  flex-direction: column;
}

.selector-filter-section {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 12px;
}

.selector-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.selector-tools {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
}

.question-selector-list {
  max-height: 56vh;
  overflow-y: auto;
  padding-right: 4px;
}

.selector-question-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 10px;
  background: #fff;
}

.selector-question-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
}

.selector-header-main {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.selector-question-title {
  font-weight: 600;
  color: #303133;
}

.selector-question-meta {
  margin: 8px 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.questions-edit-layout {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.questions-edit-list {
  flex: 1;
  max-height: 60vh;
  overflow-y: auto;
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  padding: 10px;
}

.question-index-panel {
  width: 90px;
  max-height: 60vh;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 8px;
  background: #fafafa;
  position: sticky;
  top: 0;
}

.question-index-title {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.question-index-panel .el-button {
  width: 100%;
  margin: 0 0 6px 0;
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

.composite-question-block {
  margin-top: 10px;
}

.composite-sub-question {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 10px;
  margin-bottom: 10px;
  background: #fcfcfd;
}

.composite-sub-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.composite-sub-title {
  font-weight: 600;
  color: #303133;
}

.composite-sub-content {
  margin-bottom: 8px;
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

.questions-edit-list >>> .markdown-body pre {
  padding: 0 !important;
}

.questions-edit-list >>> .markdown-body pre ol.pre-numbering {
  display: none !important;
}

.questions-edit-list >>> .markdown-body img {
  max-width: 100%;
  height: auto;
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

  .selector-question-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .selector-tools {
    justify-content: flex-start;
  }

  .questions-edit-layout {
    flex-direction: column;
  }

  .question-index-panel {
    width: 100%;
    max-height: none;
    position: static;
    display: grid;
    grid-template-columns: repeat(8, minmax(0, 1fr));
    gap: 6px;
  }

  .question-index-title {
    grid-column: 1 / -1;
    margin-bottom: 0;
  }

  .question-index-panel .el-button {
    margin: 0;
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
  padding: 0;
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
