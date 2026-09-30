<template>
  <div class="exam-paper-admin-container">
    <el-card class="exam-paper-card" v-if="!showEditDialog">
      <div slot="header" class="card-header">
        <div class="header-left">
          <i class="el-icon-document-copy"></i>
          <span>{{ $t('m.Exam_Paper_List') }}</span>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="el-icon-plus" @click="openCreateDialog" size="small">{{ $t('m.Exam_Create_Paper') }}</el-button>
          <el-button icon="el-icon-refresh" @click="loadPapers" size="small">{{ $t('m.Refresh') }}</el-button>
        </div>
      </div>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-input
              v-model="filters.keyword"
              :placeholder="$t('m.Exam_Search_Paper_Placeholder')"
              clearable
              @clear="loadPapers"
              @keyup.enter.native="loadPapers"
            >
              <el-button slot="append" icon="el-icon-search" @click="loadPapers"></el-button>
            </el-input>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.isShared" :placeholder="$t('m.Exam_Shared_Status')" clearable @change="loadPapers">
              <el-option :label="$t('m.All')" value=""></el-option>
              <el-option :label="$t('m.Exam_Private_Paper')" value="0"></el-option>
              <el-option :label="$t('m.Exam_Shared_Paper')" value="1"></el-option>
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.isPublic" :placeholder="$t('m.Exam_Public_Status')" clearable @change="loadPapers">
              <el-option :label="$t('m.All')" value=""></el-option>
              <el-option :label="$t('m.Exam_Not_Public')" value="0"></el-option>
              <el-option :label="$t('m.Exam_Public')" value="1"></el-option>
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
        <el-table-column prop="id" :label="$t('m.Exam_Paper_ID')" width="90" align="center"></el-table-column>

        <el-table-column prop="title" :label="$t('m.Exam_Paper_Title')" min-width="200">
          <template slot-scope="{ row }">
            <div class="paper-title">
              {{ row.title }}
              <el-tag v-if="row.isShared === 1" size="mini" type="success" style="margin-left: 8px;">{{ $t('m.Exam_Shared') }}</el-tag>
              <el-tag v-else size="mini" type="info" style="margin-left: 8px;">{{ $t('m.Exam_Private') }}</el-tag>
              <el-tag v-if="row.isPublic === 1" size="mini" type="danger" style="margin-left: 8px;">{{ $t('m.Exam_Public_On_Main') }}</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="questionCount" :label="$t('m.Exam_Question_Count')" width="100" align="center"></el-table-column>
        <el-table-column prop="totalScore" :label="$t('m.Exam_Total_Score')" width="80" align="center"></el-table-column>

        <el-table-column prop="creator.username" :label="$t('m.Contest_Creator')" width="150">
          <template slot-scope="{ row }">
            <div v-if="row.creator && row.creator.username">
              {{ row.creator.username }}
            </div>
            <div v-else>{{ row.creatorId || '-' }}</div>
          </template>
        </el-table-column>

        <el-table-column prop="createdAt" :label="$t('m.Exam_Create_Time')" width="180">
          <template slot-scope="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column :label="$t('m.Operation')" width="200" fixed="right">
          <template slot-scope="{ row }">
            <el-button-group>
              <el-button size="mini" icon="el-icon-edit" @click.stop="editPaper(row)">{{ $t('m.Edit') }}</el-button>
              <el-button
                size="mini"
                icon="el-icon-delete"
                type="danger"
                @click.stop="deletePaper(row)"
              >{{ $t('m.Delete') }}</el-button>
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
          <el-button icon="el-icon-arrow-left" @click="closeEditPage">{{ $t('m.Exam_Back_To_Paper_List') }}</el-button>
          <h3 class="editor-page-title">{{ dialogTitle }}</h3>
        </div>
        <div class="editor-toolbar-actions">
          <el-button v-if="isCreateMode" @click="closeEditPage">{{ $t('m.Cancel') }}</el-button>
          <el-button v-if="!isEditMode" type="primary" @click="isEditMode = true">{{ $t('m.Edit') }}</el-button>
          <template v-else>
            <el-button v-if="!isCreateMode" @click="isEditMode = false">{{ $t('m.Exam_Cancel_Edit') }}</el-button>
            <el-button type="primary" @click="confirmEdit" :loading="saving">{{ isCreateMode ? $t('m.Create') : $t('m.Save') }}</el-button>
          </template>
        </div>
      </div>

      <el-card class="exam-paper-editor-card" shadow="never">
        <el-form v-if="editForm" :model="editForm" :rules="editRules" ref="editForm" label-width="100px">
        <el-form-item :label="$t('m.Exam_Paper_Title')" prop="title">
          <el-input v-model="editForm.title" :placeholder="$t('m.Exam_Paper_Title_Placeholder')" :disabled="!isEditMode"></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.Exam_Paper_Description')">
          <el-input type="textarea" v-model="editForm.description" :rows="3" :placeholder="$t('m.Exam_Paper_Description_Placeholder')" :disabled="!isEditMode"></el-input>
        </el-form-item>
        <el-form-item :label="$t('m.Exam_Shared_Status')">
          <el-switch
            v-model="editForm.isShared"
            :active-text="$t('m.Exam_Shared')"
            :inactive-text="$t('m.Exam_Private')"
            :disabled="!isEditMode"
            @change="handleSharedChange"
          ></el-switch>
        </el-form-item>
        <el-form-item :label="$t('m.Exam_Public_On_Main')">
          <el-switch
            v-model="editForm.isPublic"
            :active-text="$t('m.Exam_Public')"
            :inactive-text="$t('m.Exam_Not_Public')"
            :disabled="!isEditMode"
          ></el-switch>
        </el-form-item>

        <el-divider>{{ $t('m.Problem_List') }}</el-divider>

        <div v-if="isEditMode" class="selected-questions-toolbar">
          <el-button icon="el-icon-plus" type="primary" @click="openQuestionSelectorDialog('objective')">
            {{ $t('m.Exam_Add_Question') }}
          </el-button>
          <el-tag size="small" type="info">{{ $t('m.Exam_Selected_Count', { count: selectedEditQuestions.length }) }}</el-tag>
          <el-tag size="small" type="success">{{ $t('m.Exam_Total_Score_Short', { score: getTotalScore() }) }}</el-tag>
        </div>

        <div v-if="selectedEditQuestions.length > 0" class="questions-edit-layout">
          <div class="question-index-panel" v-if="selectedEditQuestions.length > 1">
            <div class="question-index-title">{{ $t('m.Exam_Index_Nav') }}</div>
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
                  <span>{{ $t('m.Exam_Score_Unit') }}</span>
                  <el-tooltip v-if="isEditMode" :content="$t('m.Exam_Score_Change_Tip')" placement="top">
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
                          <span class="composite-sub-title">{{ $t('m.Exam_Sub_Question', { index: subIndex + 1 }) }}</span>
                          <el-tag size="mini" type="warning">{{ $t('m.Exam_Score_Short', { score: Number(subQuestion.score || 0) }) }}</el-tag>
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
                          <span class="meta-label">{{ $t('m.Exam_Correct_Answer_Label') }}</span>
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
                      <span class="meta-label">{{ $t('m.Exam_Correct_Answer_Label') }}</span>
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
                          {{ $t('m.View_Detail') }}
                        </el-button>
                      </div>
                      <div class="question-description" v-html="renderMarkdown(q.problem.description)"></div>
                      <div class="question-meta">
                        <el-tag size="small" type="info">{{ $t('m.Exam_Time_Label', { time: q.problem.timeLimit }) }}</el-tag>
                        <el-tag size="small" type="warning">{{ $t('m.Exam_Memory_Label', { mem: q.problem.memoryLimit }) }}</el-tag>
                        <el-tag size="small" type="primary">{{ $t('m.Exam_Judge_Mode_Label', { mode: getJudgeModeText(q.problem.judgeMode) }) }}</el-tag>
                        <el-tag size="small" type="success">{{ $t('m.Exam_Difficulty_Label', { diff: getDifficultyName(q.problem.difficulty) }) }}</el-tag>
                      </div>
                    </div>
                    <div v-else>
                      <strong>{{ $t('m.Exam_BingOJ_Problem') }} - {{ q.problemId }}</strong>
                      <el-button
                        size="mini"
                        type="text"
                        icon="el-icon-view"
                        @click.stop="fetchAndviewProblemDetail(q.problemId)"
                        style="margin-left: 10px;"
                      >
                        {{ $t('m.View_Detail') }}
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
          <p>{{ $t('m.Exam_No_Question_Selected') }}</p>
          <p class="hint">{{ $t('m.Exam_Click_Add_Question') }}</p>
        </div>

        <el-form-item style="margin-top: 20px;">
          <el-alert
            v-if="isEditMode"
            :title="$t('m.Prompt')"
            type="info"
            :closable="false"
            :description="$t('m.Exam_Edit_Mode_Tip')"
          >
          </el-alert>
          <el-alert
            v-else
            :title="$t('m.Prompt')"
            type="warning"
            :closable="false"
            :description="$t('m.Exam_View_Mode_Tip')"
          >
          </el-alert>
        </el-form-item>
        </el-form>

        <div class="editor-footer-actions">
          <el-button @click="closeEditPage">{{ isCreateMode ? $t('m.Cancel') : $t('m.Close') }}</el-button>
          <el-button v-if="!isEditMode" type="primary" @click="isEditMode = true">{{ $t('m.Edit') }}</el-button>
          <template v-else>
            <el-button v-if="!isCreateMode" @click="isEditMode = false">{{ $t('m.Exam_Cancel_Edit') }}</el-button>
            <el-button type="primary" @click="confirmEdit" :loading="saving">{{ isCreateMode ? $t('m.Create') : $t('m.Save') }}</el-button>
          </template>
        </div>
      </el-card>
    </div>

    <el-dialog
      :title="$t('m.Exam_Add_Question')"
      :visible.sync="showQuestionSelectorDialog"
      width="1200px"
      append-to-body
      class="question-selector-dialog-wrapper"
      @opened="handleQuestionSelectorOpened"
      @close="handleQuestionSelectorClose"
    >
      <div class="question-selector-dialog">
        <el-tabs
          v-model="questionSelectorTab"
          class="question-selector-tabs"
          @tab-click="handleQuestionSelectorTabChange"
        >
          <el-tab-pane :label="$t('m.Exam_Objective')" name="objective">
        <div class="filter-section selector-filter-section">
          <el-row :gutter="10">
            <el-col :xs="24" :sm="12" :md="8">
              <el-input
                v-model="questionFilters.keyword"
                :placeholder="$t('m.Exam_Search_By_Title')"
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
                :placeholder="$t('m.Exam_Search_By_ID')"
                prefix-icon="el-icon-ticket"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              ></el-input>
            </el-col>
            <el-col :xs="24" :sm="12" :md="5">
              <el-select
                v-model="questionFilters.type"
                :placeholder="$t('m.Exam_Type_Filter')"
                clearable
                @change="loadQuestionBank"
                style="width: 100%;"
              >
                <el-option :label="$t('m.Exam_All_Types')" value=""></el-option>
                <el-option :label="$t('m.Single_Choice')" value="single_choice"></el-option>
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice"></el-option>
                <el-option :label="$t('m.Judge')" value="judge"></el-option>
                <el-option :label="$t('m.Practice_Fill_Blank')" value="fill_blank"></el-option>
                <el-option :label="$t('m.Subjective')" value="subjective"></el-option>
                <el-option :label="$t('m.Practice_Composite')" value="composite"></el-option>
              </el-select>
            </el-col>
            <el-col :xs="24" :sm="12" :md="5">
              <el-select
                v-model="questionFilters.course"
                :placeholder="$t('m.Exam_Course_Filter')"
                clearable
                @change="loadQuestionBank"
                style="width: 100%;"
              >
                <el-option :label="$t('m.Exam_All_Courses')" value=""></el-option>
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
                :placeholder="$t('m.Exam_Tag_Filter')"
                prefix-icon="el-icon-price-tag"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              ></el-input>
            </el-col>
            <el-col :xs="24" :sm="8">
              <el-select
                v-model="questionFilters.sortKey"
                :placeholder="$t('m.Exam_Sort_By')"
                @change="handleQuestionSortChange"
                style="width: 100%;"
              >
                <el-option :label="$t('m.Exam_Sort_Newest')" value="create_desc"></el-option>
                <el-option :label="$t('m.Exam_Sort_Oldest')" value="create_asc"></el-option>
                <el-option :label="$t('m.Exam_Sort_ID_Asc')" value="id_asc"></el-option>
                <el-option :label="$t('m.Exam_Sort_ID_Desc')" value="id_desc"></el-option>
              </el-select>
            </el-col>
            <el-col :xs="24" :sm="6" class="selector-tools">
              <el-checkbox v-model="questionFilters.onlyUnselected">{{ $t('m.Exam_Only_Unselected') }}</el-checkbox>
              <el-button
                type="text"
                icon="el-icon-refresh"
                @click="loadQuestionBank"
                :loading="questionsLoading"
              >
                {{ $t('m.Exam_Refresh_Bank') }}
              </el-button>
              <el-button type="text" @click="resetQuestionSelectorFilters(); loadQuestionBank()">{{ $t('m.Exam_Reset_Filters') }}</el-button>
            </el-col>
          </el-row>
        </div>

        <div class="selector-summary">
          <el-tag size="small" type="info">{{ $t('m.Exam_Bank_Summary', { shown: filteredQuestionBank.length, page: questionBank.length, total: questionBankTotal }) }}</el-tag>
          <el-tag size="small" type="success">{{ $t('m.Exam_Selected_Count', { count: selectedEditQuestions.length }) }}</el-tag>
        </div>

        <div class="question-selector-list" v-loading="questionsLoading">
          <el-empty
            v-if="!questionsLoading && questionBank.length > 0 && filteredQuestionBank.length === 0"
            :description="$t('m.Exam_All_Filtered')"
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
                <el-tag v-if="row.isShared === 1" size="mini" type="success">{{ $t('m.Exam_Shared') }}</el-tag>
                <el-tag v-else size="mini" type="info">{{ $t('m.Exam_Private') }}</el-tag>
                <el-tag v-if="row.creator && row.creator.username" size="mini" type="warning">
                  {{ $t('m.Contest_Creator') }}: {{ row.creator.username }}
                </el-tag>
              </div>
              <el-button
                size="mini"
                :type="isQuestionSelected(row) ? 'danger' : 'primary'"
                plain
                @click="toggleObjectiveQuestion(row)"
              >
                {{ isQuestionSelected(row) ? $t('m.Remove') : $t('m.Exam_Add') }}
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
              <span class="meta-label">{{ $t('m.Exam_Correct_Answer_Label') }}</span>
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
          </el-tab-pane>

          <el-tab-pane :label="$t('m.Programming')" name="programming">
            <el-form :model="programmingForm" label-width="100px" class="programming-selector-form">
              <el-form-item :label="$t('m.Exam_Add_Method')">
                <el-radio-group v-model="programmingInputMode" @change="handleProgrammingInputModeChange">
                  <el-radio label="manual">{{ $t('m.Exam_Add_By_ID') }}</el-radio>
                  <el-radio label="tag">{{ $t('m.Exam_Add_By_Tag') }}</el-radio>
                </el-radio-group>
              </el-form-item>

              <template v-if="programmingInputMode === 'manual'">
                <el-form-item :label="$t('m.Problem_ID')" required>
                  <el-input
                    v-model.trim="programmingForm.problemId"
                    :placeholder="$t('m.Exam_Problem_ID_Placeholder')"
                    style="width: 320px;"
                    @keyup.enter.native="fetchProgrammingProblemInfo"
                  />
                  <el-button
                    type="primary"
                    icon="el-icon-search"
                    :loading="fetchingProblem"
                    style="margin-left: 10px;"
                    @click="fetchProgrammingProblemInfo"
                  >{{ $t('m.Exam_Fetch_Problem') }}</el-button>
                </el-form-item>
              </template>

              <template v-else>
                <el-form-item :label="$t('m.Exam_Problem_Tags')">
                  <div v-loading="problemTagsLoading" class="programming-tag-list">
                    <div
                      v-for="group in problemTagsAndClassificationList"
                      :key="group.classification ? group.classification.id : 'uncategorized'"
                      class="programming-tag-group"
                    >
                      <div class="programming-tag-group-title">
                        {{ group.classification ? group.classification.name : $t('m.Exam_Uncategorized') }}
                      </div>
                      <el-tag
                        v-for="tag in group.tagList"
                        :key="tag.id"
                        :type="isTagSelected(tag.id) ? 'primary' : 'info'"
                        :color="isTagSelected(tag.id) ? (tag.color || '#409eff') : ''"
                        effect="dark"
                        size="small"
                        class="programming-tag"
                        @click="toggleProblemTag(tag)"
                      >{{ tag.name }}</el-tag>
                    </div>
                  </div>
                </el-form-item>
                <el-form-item v-if="selectedProblemTagIds.length" :label="$t('m.Exam_Selected_Tags')">
                  <el-tag
                    v-for="tagId in selectedProblemTagIds"
                    :key="tagId"
                    closable
                    type="primary"
                    class="selected-programming-tag"
                    @close="removeSelectedTag(tagId)"
                  >{{ getTagName(tagId) }}</el-tag>
                  <el-button type="text" size="small" @click="clearAllTags">{{ $t('m.Exam_Clear') }}</el-button>
                </el-form-item>
                <el-form-item v-if="filteredProblemsTotal" :label="$t('m.Problem_List')">
                  <el-table :data="filteredProblemsByTag" stripe border max-height="280">
                    <el-table-column prop="problemId" :label="$t('m.Exam_Problem_Number')" width="120" />
                    <el-table-column prop="title" :label="$t('m.Problem')" min-width="220" show-overflow-tooltip />
                    <el-table-column :label="$t('m.Operation')" width="100" align="center">
                      <template slot-scope="{ row }">
                        <el-button type="primary" size="mini" @click="selectProblemByTag(row)">{{ $t('m.Exam_Add') }}</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                  <el-pagination
                    small
                    layout="prev, pager, next"
                    :current-page="filteredProblemsCurrentPage"
                    :page-size="filteredProblemsPageSize"
                    :total="filteredProblemsTotal"
                    class="programming-pagination"
                    @current-change="handleFilteredProblemsPageChange"
                  />
                </el-form-item>
              </template>

              <div v-if="programmingProblemPreview" class="problem-preview compact-problem-preview">
                <el-divider content-position="left">{{ $t('m.Exam_Problem_Preview') }}</el-divider>
                <el-card shadow="never">
                  <h3>{{ programmingProblemPreview.problem.title }}</h3>
                  <div class="problem-description" v-html="renderMarkdown(programmingProblemPreview.problem.description)"></div>
                  <div class="problem-meta">
                    <el-tag size="small">{{ $t('m.Exam_Problem_ID_Label') }}{{ programmingProblemPreview.problem.problemId }}</el-tag>
                    <el-tag size="small" type="info">{{ $t('m.Exam_Time_Limit_Label') }}{{ programmingProblemPreview.problem.timeLimit }}ms</el-tag>
                    <el-tag size="small" type="warning">{{ $t('m.Exam_Memory_Limit_Label') }}{{ programmingProblemPreview.problem.memoryLimit }}MB</el-tag>
                  </div>
                </el-card>
              </div>

              <el-form-item :label="$t('m.Score')" required>
                <el-input-number v-model="programmingForm.score" :min="1" :max="100" />
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </div>
      <span
        v-if="questionSelectorTab === 'programming' && programmingInputMode === 'manual'"
        slot="footer"
      >
        <el-button
          type="primary"
          :disabled="!programmingProblemPreview"
          @click="confirmAddProgrammingQuestion"
        >{{ $t('m.Exam_Confirm_Add') }}</el-button>
      </span>
    </el-dialog>

    <!-- 查看编程题详情对话框 -->
    <el-dialog
      :title="$t('m.Exam_Programming_Detail')"
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
              <span v-html="renderMarkdown('**' + $t('m.Time_Limit') + ':** ' + currentViewProblem.problem.timeLimit + 'ms')"></span>
            </el-tag>
            <el-tag size="small" type="warning">
              <span v-html="renderMarkdown('**' + $t('m.Memory_Limit') + ':** ' + currentViewProblem.problem.memoryLimit + 'MB')"></span>
            </el-tag>
            <el-tag size="small" type="primary">
              <span v-html="renderMarkdown('**' + $t('m.Exam_Judge_Mode') + ':** ' + getJudgeModeText(currentViewProblem.problem.judgeMode))"></span>
            </el-tag>
            <el-tag size="small" type="success">
              <span v-html="renderMarkdown('**' + $t('m.Difficulty') + ':** ' + getDifficultyName(currentViewProblem.problem.difficulty))"></span>
            </el-tag>
          </div>

          <el-divider></el-divider>

          <div class="problem-section">
            <h4>{{ $t('m.Problem_Description') }}</h4>
            <div class="problem-description" v-html="renderMarkdown(currentViewProblem.problem.description)"></div>
          </div>

          <div class="problem-section" v-if="currentViewProblem.problem.input">
            <h4>{{ $t('m.Exam_Input_Format') }}</h4>
            <div class="problem-io" v-html="renderMarkdown(currentViewProblem.problem.input)"></div>
          </div>

          <div class="problem-section" v-if="currentViewProblem.problem.output">
            <h4>{{ $t('m.Exam_Output_Format') }}</h4>
            <div class="problem-io" v-html="renderMarkdown(currentViewProblem.problem.output)"></div>
          </div>

          <div class="problem-section" v-if="currentViewProblem.problem.hint">
            <h4>{{ $t('m.Prompt') }}</h4>
            <div class="problem-hint" v-html="renderMarkdown(currentViewProblem.problem.hint)"></div>
          </div>

          <div class="problem-section" v-if="parseExamples(currentViewProblem.problem.examples).length > 0">
            <h4>{{ $t('m.Exam_Samples') }}</h4>
            <div v-for="(example, index) in parseExamples(currentViewProblem.problem.examples)" :key="index" class="problem-example">
              <div class="example-item">
                <strong v-html="renderMarkdown($t('m.Exam_Samples') + ' ' + (index + 1))"></strong>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown($t('m.Exam_Input_Label'))"></div>
                  <div class="example-content" v-html="renderMarkdown(example.input)"></div>
                </div>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown($t('m.Exam_Output_Label'))"></div>
                  <div class="example-content" v-html="renderMarkdown(example.output)"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="problem-empty">
        <i class="el-icon-info"></i>
        <span>{{ $t('m.Exam_No_Problem_Info') }}</span>
      </div>

      <span slot="footer">
        <el-button @click="showProblemDetailDialog = false">{{ $t('m.Close') }}</el-button>
      </span>
    </el-dialog>

    <!-- 标签题目详情对话框 -->
    <el-dialog
      :title="$t('m.Exam_Problem_Detail')"
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
              <span v-html="renderMarkdown('**' + $t('m.Time_Limit') + ':** ' + tagProblemDetail.problem.timeLimit + 'ms')"></span>
            </el-tag>
            <el-tag size="small" type="warning">
              <span v-html="renderMarkdown('**' + $t('m.Memory_Limit') + ':** ' + tagProblemDetail.problem.memoryLimit + 'MB')"></span>
            </el-tag>
            <el-tag size="small" type="primary">
              <span v-html="renderMarkdown('**' + $t('m.Exam_Judge_Mode') + ':** ' + getJudgeModeText(tagProblemDetail.problem.judgeMode))"></span>
            </el-tag>
            <el-tag size="small" type="success">
              <span v-html="renderMarkdown('**' + $t('m.Difficulty') + ':** ' + getDifficultyName(tagProblemDetail.problem.difficulty))"></span>
            </el-tag>
          </div>

          <el-divider></el-divider>

          <div class="problem-section">
            <h4>{{ $t('m.Problem_Description') }}</h4>
            <div class="problem-description" v-html="renderMarkdown(tagProblemDetail.problem.description)"></div>
          </div>

          <div class="problem-section" v-if="tagProblemDetail.problem.input">
            <h4>{{ $t('m.Exam_Input_Format') }}</h4>
            <div class="problem-io" v-html="renderMarkdown(tagProblemDetail.problem.input)"></div>
          </div>

          <div class="problem-section" v-if="tagProblemDetail.problem.output">
            <h4>{{ $t('m.Exam_Output_Format') }}</h4>
            <div class="problem-io" v-html="renderMarkdown(tagProblemDetail.problem.output)"></div>
          </div>

          <div class="problem-section" v-if="tagProblemDetail.problem.hint">
            <h4>{{ $t('m.Prompt') }}</h4>
            <div class="problem-hint" v-html="renderMarkdown(tagProblemDetail.problem.hint)"></div>
          </div>

          <div class="problem-section" v-if="parseExamples(tagProblemDetail.problem.examples).length > 0">
            <h4>{{ $t('m.Exam_Samples') }}</h4>
            <div v-for="(example, index) in parseExamples(tagProblemDetail.problem.examples)" :key="index" class="problem-example">
              <div class="example-item">
                <strong v-html="renderMarkdown($t('m.Exam_Samples') + ' ' + (index + 1))"></strong>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown($t('m.Exam_Input_Label'))"></div>
                  <div class="example-content" v-html="renderMarkdown(example.input)"></div>
                </div>
                <div class="example-block">
                  <div class="example-label" v-html="renderMarkdown($t('m.Exam_Output_Label'))"></div>
                  <div class="example-content" v-html="renderMarkdown(example.output)"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <span slot="footer">
        <el-button type="primary" @click="showTagProblemDetailDialog = false">{{ $t('m.Close') }}</el-button>
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
  props: {
    startInCreateMode: {
      type: Boolean,
      default: false
    }
  },
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
        title: [{ required: true, message: this.$t('m.Exam_Paper_Title_Placeholder'), trigger: 'blur' }]
      },
      // 题库相关
      showAddQuestionPanel: false,
      showQuestionSelectorDialog: false,
      questionSelectorTab: 'objective',
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
      questionLoadRetryTimer: null,
      questionRequestId: 0,
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
        return this.$t('m.Exam_Create_Paper')
      }
      return this.isEditMode ? this.$t('m.Exam_Edit_Paper') : this.$t('m.Exam_View_Paper')
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
    if (this.startInCreateMode) {
      this.$nextTick(() => this.openCreateDialog())
    }
  },
  beforeDestroy() {
    if (this.questionLoadRetryTimer) {
      clearTimeout(this.questionLoadRetryTimer)
      this.questionLoadRetryTimer = null
    }
  },
  watch: {
    startInCreateMode(enabled) {
      if (enabled && !this.showEditDialog) {
        this.openCreateDialog()
      } else if (!enabled && this.isCreateMode) {
        this.showEditDialog = false
      }
    },
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
        this.$message.error(this.$t('m.Load_Failed'))
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
    openQuestionSelectorDialog(tab = 'objective') {
      if (!this.isEditMode) {
        return
      }
      this.questionSelectorTab = tab === 'programming' ? 'programming' : 'objective'
      this.questionCurrentPage = 1
      this.resetProgrammingForm()
      this.programmingInputMode = 'manual'
      this.clearAllTags()
      this.showQuestionSelectorDialog = true
    },
    handleQuestionSelectorOpened() {
      const activeTab = this.questionSelectorTab === 'programming' ? 'programming' : 'objective'
      this.questionSelectorTab = activeTab
      this.$nextTick(() => {
        if (activeTab === 'objective') this.loadQuestionBank()
        else this.loadProblemTagsAndClassification()
      })
    },
    handleQuestionSelectorClose() {
      this.questionRequestId += 1
      this.questionsLoading = false
      if (this.questionLoadRetryTimer) {
        clearTimeout(this.questionLoadRetryTimer)
        this.questionLoadRetryTimer = null
      }
    },
    handleQuestionSelectorTabChange(tabPane) {
      const activeTab = tabPane && tabPane.name === 'programming' ? 'programming' : 'objective'
      if (activeTab === 'programming') {
        this.loadProblemTagsAndClassification()
      } else {
        this.loadQuestionBank()
      }
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
      if (!this.startInCreateMode && this.$route.name !== 'admin-create-exam-paper') {
        this.$router.push({ name: 'admin-create-exam-paper' })
        return
      }
      this.currentPaper = null
      this.isEditMode = true
      this.showAddQuestionPanel = false
      this.showQuestionSelectorDialog = false
      this.questionSelectorTab = 'objective'
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
            let title = this.$t('m.Exam_Unknown_Question')
            if (q.title) {
              title = q.title
            } else if (q.question && q.question.title) {
              title = q.question.title
            } else if (q.problem && q.problem.title) {
              title = q.problem.title
            } else if (q.questionId) {
              title = this.$t('m.Problem_ID') + ': ' + q.questionId
            } else if (q.problemId) {
              title = this.$t('m.Exam_BingOJ_Problem') + ' - ' + q.problemId
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
        this.$message.error(this.$t('m.Exam_Load_Detail_Failed'))
      }
    },
    resetEditForm() {
      this.editForm = null
      this.currentPaper = null
      this.isEditMode = false
      this.showAddQuestionPanel = false
      this.showQuestionSelectorDialog = false
      this.questionSelectorTab = 'objective'
      this.resetQuestionSelectorFilters()
      this.quickAddQuestionId = ''
      this.quickAddQuestionLoading = false
      if (this.$refs.editForm) {
        this.$refs.editForm.clearValidate()
      }
    },
    removeQuestion(index) {
      this.$confirm(this.$t('m.Exam_Remove_Question_Confirm'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.OK'),
        cancelButtonText: this.$t('m.Cancel'),
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
          this.$message.warning(this.$t('m.Exam_Need_One_Question'))
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
            this.$message.error(res.data.message || (isCreateOperation ? this.$t('m.Create_Failed') : this.$t('m.Update_Failed')))
            return
          }

          // 创建时如果勾选了主界面公开，补一次管理员更新
          if (isCreateOperation && this.editForm.isPublic) {
            const createdPaperId = res && res.data && res.data.data ? res.data.data.id : null
            if (createdPaperId) {
              const publishRes = await api.adminUpdateExamPaper(createdPaperId, data)
              if (publishRes.data.code !== 200) {
                this.$message.warning(this.$t('m.Exam_Publish_Fail_After_Create'))
              }
            } else {
              this.$message.warning(this.$t('m.Exam_No_ID_After_Create'))
            }
          }

          this.$message.success(isCreateOperation ? this.$t('m.Create_Success') : this.$t('m.Update_Successfully'))
          this.showEditDialog = false
          this.loadPapers()
          if (isCreateOperation) {
            this.questionCurrentPage = 1
            if (this.$route.name === 'admin-create-exam-paper') {
              this.$router.push({ name: 'admin-exam-paper' })
            }
          }
        } catch (error) {
          this.$message.error(isCreateOperation ? this.$t('m.Create_Failed') : this.$t('m.Update_Failed'))
        } finally {
          this.saving = false
        }
      })
    },
    closeEditPage() {
      if (this.$route.name === 'admin-create-exam-paper') {
        this.$router.push({ name: 'admin-exam-paper' })
        return
      }
      this.showEditDialog = false
    },
    handleSharedChange(newValue) {
      // El switch 的 v-model 会自动更新
    },
    deletePaper(row) {
      this.$confirm(this.$t('m.Exam_Delete_Paper_Confirm'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.OK'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await api.adminDeleteExamPaper(row.id)
          if (res.data.code === 200) {
            this.$message.success(this.$t('m.Delete_successfully'))
            this.loadPapers()
          } else {
            this.$message.error(res.data.message || this.$t('m.Delete_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Delete_Failed'))
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
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        fill_blank: this.$t('m.Practice_Fill_Blank'),
        subjective: this.$t('m.Subjective'),
        composite: this.$t('m.Practice_Composite'),
        programming: this.$t('m.Programming')
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
                return parsed.join(this.$t('m.Exam_Answer_Separator'))
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
              return this.$t('m.Exam_Judge_Correct')
            } else if (normalizedAnswer === 'false') {
              return this.$t('m.Exam_Judge_Wrong')
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
            return question.answer || this.$t('m.Exam_Need_Manual_Grading')

          case 'composite':
            return this.$t('m.Exam_Composite_Scoring')

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
    async loadQuestionBank({ retry = true } = {}) {
      const requestId = ++this.questionRequestId
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
        const responseData = res && res.data ? res.data : null
        const responseCode = responseData && (responseData.code || responseData.status)
        if (responseCode === 200 && responseData.data) {
          // 过滤掉 undefined 或 null 的题目
          this.questionBank = (responseData.data.questions || []).filter(q => {
            const isValid = q && q.id
            return isValid
          })
          this.questionBankTotal = responseData.data.total || 0
        } else {
          throw new Error((responseData && (responseData.message || responseData.msg)) || this.$t('m.Exam_Bank_Response_Error'))
        }
      } catch (error) {
        if (requestId !== this.questionRequestId) return
        if (retry && this.showQuestionSelectorDialog && this.questionSelectorTab === 'objective') {
          if (this.questionLoadRetryTimer) clearTimeout(this.questionLoadRetryTimer)
          this.questionLoadRetryTimer = setTimeout(() => {
            this.questionLoadRetryTimer = null
            this.loadQuestionBank({ retry: false })
          }, 500)
        } else {
          this.$message.error(this.$t('m.Exam_Load_Bank_Failed'))
        }
      } finally {
        if (requestId === this.questionRequestId) this.questionsLoading = false
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
        this.$message.success(this.$t('m.Exam_Question_Removed'))
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
        this.$message.warning(this.$t('m.Enter_Problem_ID'))
        return
      }
      if (!/^\d+$/.test(questionId)) {
        this.$message.warning(this.$t('m.Exam_Problem_ID_Numeric'))
        return
      }
      if (this.selectedEditQuestions.some(q => q && String(q.questionId) === questionId)) {
        this.$message.warning(this.$t('m.Exam_Already_Added'))
        return
      }

      this.quickAddQuestionLoading = true
      try {
        const res = await api.getQuestionDetail(questionId)
        if (!res || !res.data || res.data.code !== 200 || !res.data.data) {
          this.$message.error(this.$t('m.Exam_Question_Not_Found'))
          return
        }
        const question = res.data.data
        if (!this.isObjectiveQuestionType(question.type)) {
          this.$message.warning(this.$t('m.Exam_Not_Objective'))
          return
        }
        this.addQuestion(question)
        this.quickAddQuestionId = ''
      } catch (error) {
        this.$message.error(this.$t('m.Exam_Fetch_By_ID_Failed'))
      } finally {
        this.quickAddQuestionLoading = false
      }
    },
    // 添加客观题
    addQuestion(question) {
      // 检查参数是否有效
      if (!question || !question.id) {
        this.$message.warning(this.$t('m.Exam_Invalid_Question_Data'))
        return
      }

      // 检查是否已添加
      if (this.isQuestionSelected(question)) {
        this.$message.warning(this.$t('m.Exam_Already_Added'))
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

      this.$message.success(this.$t('m.Exam_Add_Success'))
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
        return this.$t('m.Exam_Unknown_Question')
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
        return this.$t('m.Problem_ID') + ': ' + q.questionId
      } else if (q.problemId) {
        return this.$t('m.Exam_BingOJ_Problem') + ' - ' + q.problemId
      }
      return this.$t('m.Exam_Unknown_Question')
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
        this.$message.warning(this.$t('m.Enter_Problem_ID'))
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
          this.$message.error(this.$t('m.Exam_Fetch_Problem_Failed'))
        }
      } catch (error) {
        this.programmingProblemPreview = null
        this.$message.error(this.getRequestErrorMessage(error, this.$t('m.Exam_Fetch_Problem_Failed')))
      } finally {
        this.fetchingProblem = false
      }
    },
    confirmAddProgrammingQuestion() {
      if (!this.programmingForm.problemId) {
        this.$message.warning(this.$t('m.Enter_Problem_ID'))
        return
      }

      if (!this.programmingProblemPreview) {
        this.$message.warning(this.$t('m.Exam_Click_Fetch_First'))
        return
      }

      // 检查是否已添加
      if (this.selectedEditQuestions.some(q => q.problemId === this.programmingForm.problemId)) {
        this.$message.warning(this.$t('m.Exam_Problem_Already_Added'))
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

      this.$message.success(this.$t('m.Exam_Add_Success'))
      this.showQuestionSelectorDialog = false
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
        this.$message.error(this.$t('m.Exam_Get_Problem_List_Failed'))
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
        this.$message.error(this.getRequestErrorMessage(error, this.$t('m.Exam_Get_Problem_Detail_Failed')))
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
            this.$message.warning(this.$t('m.Exam_Problem_Already_Added'))
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

          this.$message.success(this.$t('m.Exam_Problem_Added', { id: problem.problemId, title: problem.title }))
          // 清空表单以便继续添加，但保持对话框打开
          this.programmingForm.problemId = ''
          this.programmingProblemPreview = null
        } else {
          this.$message.error(this.$t('m.Exam_Fetch_Problem_Failed'))
        }
      } catch (error) {
        console.error('获取题目信息失败:', error)
        this.programmingProblemPreview = null
        this.$message.error(this.getRequestErrorMessage(error, this.$t('m.Exam_Fetch_Problem_Failed')))
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
        0: this.$t('m.Difficulty_Entry'),
        1: this.$t('m.Exam_Diff_Easy'),
        2: this.$t('m.Difficulty_Medium'),
        3: this.$t('m.Difficulty_Hard'),
        4: this.$t('m.Exam_Diff_Master'),
        5: this.$t('m.Exam_Diff_Expert')
      }
      return nameMap[difficulty] || this.$t('m.Unknown')
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
          this.$message.error(this.$t('m.Exam_Fetch_Problem_Failed'))
          this.showProblemDetailDialog = false
        }
      } catch (error) {
        console.error('获取题目信息失败:', error)
        this.$message.error(this.getRequestErrorMessage(error, this.$t('m.Exam_Fetch_Problem_Failed')))
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
        'default': this.$t('m.Exam_Judge_Mode_Default'),
        'spj': this.$t('m.Exam_Judge_Mode_SPJ'),
        'interactive': this.$t('m.Exam_Judge_Mode_Interactive'),
        'subtask': this.$t('m.Subtask')
      }
      return modeMap[mode] || mode || this.$t('m.Exam_Judge_Mode_Default')
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

.programming-selector-form {
  padding: 8px 4px 0;
}

.programming-tag-list {
  min-height: 90px;
}

.programming-tag-group {
  margin-bottom: 14px;
}

.programming-tag-group-title {
  margin-bottom: 8px;
  color: #606266;
  font-weight: 600;
}

.programming-tag {
  margin: 0 8px 8px 0;
  cursor: pointer;
}

.selected-programming-tag {
  margin: 0 8px 8px 0;
}

.programming-pagination {
  margin-top: 10px;
  text-align: center;
}

.compact-problem-preview {
  margin: 10px 0 18px;
}
</style>
