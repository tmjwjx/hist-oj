<template>
  <div class="create-homework-wrapper">
    <!-- 编辑模式数据加载遮罩 -->
    <div v-if="loadingData" class="loading-overlay">
      <el-icon class="is-loading"><i class="el-icon-loading"></i></el-icon>
      <p>{{ $t('m.CH_Loading_Homework_Data') }}</p>
    </div>

    <!-- 顶部固定导航栏 -->
    <div class="create-homework-header">
      <div class="header-content">
        <div class="header-left">
          <i class="el-icon-edit-outline"></i>
          <h2>{{ isEditMode ? $t('m.CH_Edit_Homework') : $t('m.CH_Create_Homework') }}</h2>
        </div>
        <div class="header-actions">
          <el-button @click="goBack" size="medium">
            <i class="el-icon-back"></i>
            {{ $t('m.Back') }}
          </el-button>
          <el-button v-if="isConfigPage" type="primary" @click="switchEditorPage('questions')" size="medium">
            <i class="el-icon-right"></i>
            {{ $t('m.CH_Next_Question_Management') }}
          </el-button>
          <el-button v-else type="primary" @click="saveHomework" :loading="submitting" size="medium">
            <i class="el-icon-check"></i>
            {{ isEditMode ? $t('m.Save') : $t('m.Create') }}
          </el-button>
        </div>
      </div>
    </div>

    <div class="create-homework-content">
      <div class="editor-page-switch">
        <el-button :type="isConfigPage ? 'primary' : 'default'" size="medium" @click="switchEditorPage('config')">
          {{ $t('m.CH_Step_Homework_Config') }}
        </el-button>
        <el-button :type="isQuestionPage ? 'primary' : 'default'" size="medium" @click="switchEditorPage('questions')">
          {{ $t('m.CH_Step_Question_Management') }}
        </el-button>
      </div>
      <el-form :model="form" :rules="rules" ref="homeworkForm" label-width="100px" class="homework-form">
        <!-- 第一板块：基本信息 -->
        <el-card v-if="isConfigPage" class="form-section basic-info-card" shadow="hover">
          <div slot="header" class="card-header">
            <span class="header-icon">
              <i class="el-icon-document"></i>
            </span>
            <span class="header-title">{{ $t('m.Basic_Info') }}</span>
          </div>

          <div class="basic-info-grid">
            <el-form-item :label="$t('m.Homework_Title')" prop="title" class="full-width">
              <el-input
                v-model="form.title"
                :placeholder="$t('m.Please_Enter_Homework_Title')"
                prefix-icon="el-icon-edit"
                size="medium"
              />
            </el-form-item>

            <el-form-item :label="$t('m.Description')" prop="description" class="full-width">
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="3"
                :placeholder="$t('m.Homework_Description_Tip')"
                size="medium"
              />
            </el-form-item>

            <el-form-item :label="$t('m.Start_Time')" prop="startTime">
              <el-date-picker
                v-model="form.startTime"
                type="datetime"
                :placeholder="$t('m.Please_Select_Start_Time')"
                style="width: 100%"
                size="medium"
                prefix-icon="el-icon-time"
              />
            </el-form-item>

            <el-form-item :label="$t('m.End_Time')" prop="endTime">
              <el-date-picker
                v-model="form.endTime"
                type="datetime"
                :placeholder="$t('m.Please_Select_End_Time')"
                style="width: 100%"
                size="medium"
                prefix-icon="el-icon-time"
              />
            </el-form-item>
          </div>

          <div class="display-settings">
            <!-- 作业模式选择 -->
            <div class="mode-selection">
              <div class="mode-label">{{ $t('m.CH_Homework_Mode') }}</div>
              <el-radio-group v-model="form.isExamMode">
                <el-radio :label="0">{{ $t('m.CH_Normal_Homework_Mode') }}</el-radio>
                <el-radio :label="1">{{ $t('m.CH_Exam_Mode') }}</el-radio>
              </el-radio-group>
            </div>

            <div class="setting-item">
              <el-checkbox v-model="form.showHomework">
                <span class="setting-label">{{ $t('m.Show_Homework_After_Complete') }}</span>
              </el-checkbox>
              <div class="setting-tip">
                {{ form.isExamMode === 1 ? $t('m.CH_Exam_View_After_End') : $t('m.Show_Homework_Tip') }}
              </div>
            </div>
            <div class="setting-item">
              <el-checkbox v-model="form.showScore">
                <span class="setting-label">{{ $t('m.Show_Score_After_Complete') }}</span>
              </el-checkbox>
              <div class="setting-tip">
                {{ form.isExamMode === 1 ? $t('m.CH_Exam_View_After_End') : $t('m.Show_Score_Tip') }}
              </div>
            </div>
            <div class="setting-item">
              <el-checkbox v-model="form.showRank">
                  <span class="setting-label">{{ $t('m.CH_Show_Rank_After_End') }}</span>
                </el-checkbox>
                <div class="setting-tip">
                  {{ form.isExamMode === 1 ? $t('m.CH_Exam_Rank_After_End') : $t('m.CH_Rank_Tip') }}
              </div>
            </div>
            <div class="setting-item">
              <el-checkbox v-model="form.showAnswer">
                  <span class="setting-label">{{ $t('m.CH_Allow_View_Answer') }}</span>
                </el-checkbox>
                <div class="setting-tip">
                  {{ form.isExamMode === 1 ? $t('m.CH_Exam_View_After_End') : $t('m.CH_Answer_Tip') }}
              </div>
            </div>
          </div>

          <!-- 考试模式配置 -->
          <div v-if="form.isExamMode === 1" class="exam-config-section">
            <el-divider content-position="left">{{ $t('m.CH_Time_Settings') }}</el-divider>

            <el-form-item :label="$t('m.CH_Exam_Duration')" prop="examDuration" required>
              <el-input-number
                v-model="form.examDuration"
                :min="1"
                :max="600"
                :step="5"
                controls-position="right"
              />
              <span class="unit-label">{{ $t('m.CH_Minutes') }}</span>
              <span class="setting-tip">{{ $t('m.CH_Exam_Duration_Tip') }}</span>
            </el-form-item>

            <!-- 时间差异提醒 -->
            <el-alert
              v-if="timeRangeMismatch"
              type="warning"
              :closable="false"
              show-icon
              style="margin-bottom: 20px;"
            >
              <template slot="title">
                <div style="line-height: 1.6;">
                  <strong>{{ $t('m.CH_Time_Mismatch_Title') }}</strong>
                  {{ $t('m.CH_Time_Mismatch_Desc', { minutes: timeRangeMinutes, duration: form.examDuration }) }}
                  <br>
                  <span style="color: #E6A23C;">{{ $t('m.CH_Time_Mismatch_Hint') }}</span>
                  {{ $t('m.CH_Time_Mismatch_Auto_Submit', { duration: form.examDuration }) }}
                </div>
              </template>
            </el-alert>

            <el-alert
              v-else
              type="success"
              :closable="false"
              show-icon
              style="margin-bottom: 20px;"
            >
              <template slot="title">
                <div style="line-height: 1.6;">
                  <strong>{{ $t('m.CH_Time_Match_Title') }}</strong>
                  {{ $t('m.CH_Time_Match_Desc', { duration: form.examDuration }) }}
                  <br>
                  {{ $t('m.CH_Time_Match_Note', { duration: form.examDuration }) }}
                </div>
              </template>
            </el-alert>

            <el-form-item :label="$t('m.CH_Allow_Submit_Time')" prop="allowSubmitAfterMinutes">
              <el-input-number
                v-model="form.allowSubmitAfterMinutes"
                :min="0"
                :max="600"
                :step="5"
                controls-position="right"
              />
              <span class="unit-label">{{ $t('m.CH_Minutes') }}</span>
              <span class="setting-tip">{{ $t('m.CH_Allow_Submit_Tip') }}</span>
            </el-form-item>

            <el-divider content-position="left">{{ $t('m.CH_Anti_Cheat_Settings') }}</el-divider>

            <div class="exam-settings-grid">
              <div class="exam-setting-item">
                <el-checkbox v-model="form.disableCopyPaste">
                  <span class="setting-label">{{ $t('m.CH_Disable_Copy_Paste') }}</span>
                </el-checkbox>
                <div class="setting-tip">{{ $t('m.CH_Disable_Copy_Paste_Tip') }}</div>
              </div>

              <div class="exam-setting-item">
                <el-checkbox v-model="form.requireFullscreen">
                  <span class="setting-label">{{ $t('m.CH_Require_Fullscreen') }}</span>
                </el-checkbox>
                <div class="setting-tip">{{ $t('m.CH_Fullscreen_Tip') }}</div>
              </div>

              <div class="exam-setting-item">
                <el-checkbox v-model="form.disallowTabSwitch">
                  <span class="setting-label">{{ $t('m.CH_Disallow_Tab_Switch') }}</span>
                </el-checkbox>
                <div class="setting-tip">{{ $t('m.CH_Tab_Switch_Tip') }}</div>
              </div>
            </div>

            <el-alert
              type="info"
              :closable="false"
              show-icon
              style="margin-top: 15px;"
            >
              <template slot="title">
                <div style="line-height: 1.8;">
                  <strong>{{ $t('m.CH_Exam_Mode_Notes') }}</strong>
                  <ul style="margin: 10px 0 0 20px; padding: 0;">
                    <li>{{ $t('m.CH_Exam_Note_1') }}</li>
                    <li>{{ $t('m.CH_Exam_Note_2') }}</li>
                    <li>{{ $t('m.CH_Exam_Note_3') }}</li>
                  </ul>
                </div>
              </template>
            </el-alert>
          </div>
        </el-card>

      <!-- 第二板块：题目管理（题库 + 已选题目） -->
      <el-card v-if="isQuestionPage" class="form-section questions-management-card" shadow="hover">
        <div slot="header" class="card-header">
          <span class="header-icon">
            <i class="el-icon-collection"></i>
          </span>
          <span class="header-title">{{ $t('m.CH_Question_Management') }}</span>
          <div class="header-stats">
            <el-tag size="small" type="info">{{ $t('m.CH_Selected_Count', { count: selectedQuestions.length }) }}</el-tag>
            <el-tag size="small" type="success" style="margin-left: 8px;">{{ $t('m.CH_Total_Score_Point', { score: getTotalScore() }) }}</el-tag>
          </div>
          <div class="header-actions">
            <el-button
              size="small"
              icon="el-icon-back"
              @click="switchEditorPage('config')"
              plain
              style="margin-right: 8px;"
            >
              {{ $t('m.CH_Back_To_Config') }}
            </el-button>
            <el-button
              type="primary"
              size="small"
              icon="el-icon-view"
              @click="showFullPreview"
              :disabled="selectedQuestions.length === 0"
              plain
            >
              {{ $t('m.CH_Full_Preview') }}
            </el-button>
          </div>
        </div>

        <div class="questions-container-layout">
          <!-- 题目区域 -->
          <div class="selected-questions-panel-full">
            <!-- 顶部工具栏：添加题目按钮 -->
            <div class="questions-toolbar">
              <div class="toolbar-left">
                <span class="panel-title">{{ $t('m.CH_Selected_Questions') }}</span>
                <el-tag size="mini" type="success">{{ $t('m.CH_Count_Questions', { count: selectedQuestions.length }) }}</el-tag>
              </div>
              <div class="toolbar-right">
                <el-button-group>
                  <el-button type="primary" icon="el-icon-collection" @click="openQuestionSelectorDialog">
                    {{ $t('m.CH_Objective_Bank') }}
                  </el-button>
                  <el-button type="success" icon="el-icon-document" @click="showAddProgrammingDialog = true">
                    {{ $t('m.CH_Programming_Bank') }}
                  </el-button>
                  <el-button type="warning" icon="el-icon-download" @click="openImportPaperDialog">
                    {{ $t('m.CH_Import_Paper') }}
                  </el-button>
                </el-button-group>
              </div>
            </div>

            <!-- 题目列表 -->
            <div class="selected-questions-list" v-if="selectedQuestions.length > 0">
              <transition-group name="list" tag="div" class="questions-container">
                <div
                  v-for="(question, index) in selectedQuestions"
                  :key="question.id || question.problemId"
                  class="question-item-card"
                >
                  <div class="question-item-main">
                    <div class="question-index">
                      <span class="index-number">{{ index + 1 }}</span>
                    </div>

                    <div class="selected-question-content">
                      <div class="question-title-row">
                        <el-tag :type="getQuestionTypeColor(question.type)" size="mini">
                          {{ getQuestionTypeText(question.type) }}
                        </el-tag>
                        <span class="question-title-text" :title="question.title">{{ question.title }}</span>
                      </div>

                      <div class="selected-question-meta">
                        <div class="score-editor">
                          <span class="score-label">{{ $t('m.CH_Score_Label') }}</span>
                          <el-input-number
                            v-model="question.score"
                            :min="1"
                            :max="100"
                            size="mini"
                            controls-position="right"
                          />
                          <span class="score-unit">{{ $t('m.CH_Score_Unit') }}</span>
                        </div>

                        <div class="question-tags-course" style="margin-top: 5px;">
                          <el-tag v-if="question.course" type="warning" size="mini" style="margin-right: 5px;">
                            <i class="el-icon-collection"></i> {{ question.course }}
                          </el-tag>
                          <el-tag
                            v-for="(tag, idx) in parseQuestionTags(question.tags)"
                            :key="idx"
                            size="mini"
                            type="info"
                            style="margin-right: 3px;"
                          >
                            {{ tag }}
                          </el-tag>
                        </div>
                      </div>
                    </div>

                    <div class="question-actions">
                      <el-button-group>
                        <el-tooltip :content="$t('m.View_Detail')" placement="top">
                          <el-button
                            size="mini"
                            icon="el-icon-view"
                            @click="viewQuestionDetail(question)"
                          />
                        </el-tooltip>
                        <el-tooltip :content="$t('m.Move_Up')" placement="top">
                          <el-button
                            size="mini"
                            icon="el-icon-top"
                            :disabled="index === 0"
                            @click="moveUp(index)"
                          />
                        </el-tooltip>
                        <el-tooltip :content="$t('m.Move_Down')" placement="top">
                          <el-button
                            size="mini"
                            icon="el-icon-bottom"
                            :disabled="index === selectedQuestions.length - 1"
                            @click="moveDown(index)"
                          />
                        </el-tooltip>
                        <el-tooltip :content="$t('m.Remove')" placement="top">
                          <el-button
                            size="mini"
                            icon="el-icon-delete"
                            type="danger"
                            @click="removeQuestion(question)"
                          />
                        </el-tooltip>
                      </el-button-group>
                    </div>
                  </div>

                  <div class="selected-question-expanded">
                    <div class="selected-question-block">
                      <div class="markdown-body preview-markdown" v-html="renderMarkdown(question.title)" v-highlight></div>
                      <div v-if="question.content" class="markdown-body preview-markdown" v-html="renderMarkdown(question.content)" v-highlight></div>
                    </div>

                    <div
                      v-if="question.type === 'single_choice' || question.type === 'multiple_choice'"
                      class="selected-question-options"
                    >
                      <div v-for="(option, idx) in parseOptions(question.options)" :key="idx" class="selected-question-option">
                        <span class="option-letter">{{ option.letter }}.</span>
                        <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                      </div>
                    </div>

                    <div v-if="question.type === 'judge'" class="selected-question-options">
                      <div class="selected-question-option">
                        <span class="option-letter">✓</span>
                        <span class="option-text">{{ $t('m.True') }}</span>
                      </div>
                      <div class="selected-question-option">
                        <span class="option-letter">✗</span>
                        <span class="option-text">{{ $t('m.False') }}</span>
                      </div>
                    </div>

                    <div v-if="question.type === 'composite'" class="selected-question-options composite-preview-list">
                      <div
                        v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(question.options)"
                        :key="subQuestion.id || subIndex"
                        class="composite-preview-item"
                      >
                        <div class="composite-preview-header">
                          <span>{{ $t('m.CH_Sub_Question', { index: subIndex + 1 }) }}</span>
                          <span>{{ Number(subQuestion.score || 0) }}{{ $t('m.CH_Score_Unit') }}</span>
                        </div>
                        <div class="markdown-body preview-markdown" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
                        <div class="selected-question-options" style="padding-left: 0;">
                          <div v-for="(option, idx) in parseOptions(subQuestion.options)" :key="idx" class="selected-question-option">
                            <span class="option-letter">{{ option.letter }}.</span>
                            <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                          </div>
                        </div>
                        <div class="selected-question-answer answer-info compact-answer-info">
                          <el-tag size="mini" type="success">
                            {{ $t('m.CH_Correct_Answer') }}{{ getCompositeCorrectAnswer(question.answer, subQuestion.id, subIndex) }}
                          </el-tag>
                        </div>
                      </div>
                    </div>

                    <div v-if="question.type === 'fill_blank'" class="selected-question-answer answer-info compact-answer-info">
                      <el-tag type="success" size="mini">{{ $t('m.CH_Reference_Answer') }}{{ formatAnswerForDisplay(question) }}</el-tag>
                    </div>
                    <div v-if="question.type === 'subjective'" class="selected-question-answer answer-info compact-answer-info">
                      <div
                        class="markdown-body preview-markdown"
                        v-html="renderMarkdown(formatAnswerForDisplay(question))"
                        v-highlight
                      ></div>
                    </div>
                    <div v-if="question.type === 'judge'" class="selected-question-answer answer-info compact-answer-info">
                      <el-tag type="success" size="mini">{{ $t('m.CH_Reference_Answer') }}{{ formatAnswerForDisplay(question) }}</el-tag>
                    </div>
                    <div v-if="question.type === 'single_choice' || question.type === 'multiple_choice'" class="selected-question-answer answer-info compact-answer-info">
                      <el-tag type="success" size="mini">{{ $t('m.CH_Reference_Answer') }}{{ formatAnswerForDisplay(question) }}</el-tag>
                    </div>

                    <div v-if="question.type === 'programming'" class="selected-programming-detail">
                      <template v-if="getProgrammingProblemInfo(question.problemId)">
                        <div class="content-section">
                          <h4>{{ $t('m.Problem_Description') }}</h4>
                          <div
                            class="markdown-body preview-markdown"
                            v-html="renderMarkdown(getProgrammingProblemInfo(question.problemId).problem.description)"
                            v-highlight
                          ></div>
                        </div>
                        <div class="content-section" v-if="getProgrammingProblemInfo(question.problemId).problem.input">
                          <h4>{{ $t('m.CH_Input_Format') }}</h4>
                          <div
                            class="markdown-body preview-markdown"
                            v-html="renderMarkdown(getProgrammingProblemInfo(question.problemId).problem.input)"
                            v-highlight
                          ></div>
                        </div>
                        <div class="content-section" v-if="getProgrammingProblemInfo(question.problemId).problem.output">
                          <h4>{{ $t('m.CH_Output_Format') }}</h4>
                          <div
                            class="markdown-body preview-markdown"
                            v-html="renderMarkdown(getProgrammingProblemInfo(question.problemId).problem.output)"
                            v-highlight
                          ></div>
                        </div>
                      </template>
                      <el-alert v-else type="info" :closable="false">
                        {{ $t('m.CH_Added_Programming', { name: question.problemId || question.title }) }}
                      </el-alert>
                    </div>
                  </div>
                </div>
              </transition-group>
            </div>

            <div v-else class="empty-selected">
              <i class="el-icon-document"></i>
              <p>{{ $t('m.CH_No_Questions_Selected') }}</p>
              <p class="hint">{{ $t('m.CH_Add_From_Toolbar') }}</p>
            </div>
          </div>
        </div>
      </el-card>

      </el-form>

    </div>

    <!-- 添加客观题对话框 -->
    <el-dialog
      :title="$t('m.CH_Add_Question')"
      :visible.sync="showQuestionSelectorDialog"
      width="1200px"
      append-to-body
      class="question-selector-dialog-wrapper"
    >
      <div class="question-selector-dialog">
        <div class="selector-filter-section">
          <el-row :gutter="10">
            <el-col :xs="24" :sm="12" :md="8">
              <el-input
                v-model="questionFilters.keyword"
                :placeholder="$t('m.CH_Search_By_Title')"
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
                :placeholder="$t('m.CH_Search_By_Id')"
                prefix-icon="el-icon-ticket"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              ></el-input>
            </el-col>
            <el-col :xs="24" :sm="12" :md="5">
              <el-select
                v-model="questionFilters.type"
                :placeholder="$t('m.CH_Filter_By_Type')"
                clearable
                @change="loadQuestionBank"
                style="width: 100%;"
              >
                <el-option :label="$t('m.CH_All_Types')" value=""></el-option>
                <el-option :label="$t('m.Single_Choice')" value="single_choice"></el-option>
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice"></el-option>
                <el-option :label="$t('m.Judge')" value="judge"></el-option>
                <el-option :label="$t('m.CH_Fill_Blank')" value="fill_blank"></el-option>
                <el-option :label="$t('m.Subjective')" value="subjective"></el-option>
                <el-option :label="$t('m.CH_Composite')" value="composite"></el-option>
              </el-select>
            </el-col>
            <el-col :xs="24" :sm="12" :md="5">
              <el-select
                v-model="questionFilters.course"
                :placeholder="$t('m.CH_Filter_By_Course')"
                clearable
                @change="loadQuestionBank"
                style="width: 100%;"
              >
                <el-option :label="$t('m.CH_All_Courses')" value=""></el-option>
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
                :placeholder="$t('m.CH_Filter_By_Tag')"
                prefix-icon="el-icon-price-tag"
                clearable
                @clear="loadQuestionBank"
                @keyup.enter.native="loadQuestionBank"
              ></el-input>
            </el-col>
            <el-col :xs="24" :sm="8">
              <el-select
                v-model="questionFilters.sortKey"
                :placeholder="$t('m.CH_Sort_By')"
                @change="handleQuestionSortChange"
                style="width: 100%;"
              >
                <el-option :label="$t('m.CH_Sort_Newest')" value="create_desc"></el-option>
                <el-option :label="$t('m.CH_Sort_Oldest')" value="create_asc"></el-option>
                <el-option :label="$t('m.CH_Sort_Id_Asc')" value="id_asc"></el-option>
                <el-option :label="$t('m.CH_Sort_Id_Desc')" value="id_desc"></el-option>
              </el-select>
            </el-col>
            <el-col :xs="24" :sm="6" class="selector-tools">
              <el-checkbox v-model="questionFilters.onlyUnselected">{{ $t('m.CH_Only_Unselected') }}</el-checkbox>
              <el-button
                type="text"
                icon="el-icon-refresh"
                @click="loadQuestionBank"
                :loading="questionsLoading"
              >
                {{ $t('m.CH_Refresh_Bank') }}
              </el-button>
              <el-button type="text" @click="resetQuestionSelectorFilters(); loadQuestionBank()">{{ $t('m.CH_Reset_Filters') }}</el-button>
            </el-col>
          </el-row>
        </div>

        <div class="selector-summary">
          <el-tag size="small" type="info">{{ $t('m.CH_Selector_Summary', { showing: filteredQuestionBank.length, page: questionBank.length, total: questionBankTotal }) }}</el-tag>
          <el-tag size="small" type="success">{{ $t('m.CH_Selected_Count', { count: selectedQuestions.length }) }}</el-tag>
        </div>

        <div class="question-selector-list" v-loading="questionsLoading">
          <el-alert
            v-if="questionBank.length === 0 && !questionsLoading"
            :title="$t('m.CH_Bank_Empty')"
            type="info"
            :closable="false"
            style="margin-bottom: 10px;"
          >
            <template slot="default">
              <div>{{ $t('m.CH_Bank_Empty_Desc') }}</div>
            </template>
          </el-alert>
          <el-empty
            v-else-if="filteredQuestionBank.length === 0 && !questionsLoading"
            :description="$t('m.CH_All_Filtered_Desc')"
          ></el-empty>

          <div
            v-for="row in filteredQuestionBank"
            :key="'selector-question-' + row.id"
            class="selector-question-item"
          >
            <div class="selector-question-header">
              <div class="selector-header-main">
                <el-tag size="mini" :type="getQuestionTypeColor(row.type)">
                  {{ getQuestionTypeText(row.type) }}
                </el-tag>
                <el-tag size="mini" type="info">ID: {{ row.id }}</el-tag>
                <el-tag size="mini" :type="getDifficultyTagType(row.difficulty)">
                  {{ $t('m.CH_Difficulty_Label', { name: getDifficultyName(row.difficulty) }) }}
                </el-tag>
                <span class="selector-question-title">{{ row.title }}</span>
              </div>
              <el-button
                size="mini"
                :type="isQuestionSelected(row) ? 'danger' : 'primary'"
                plain
                @click="toggleObjectiveQuestion(row)"
              >
                {{ isQuestionSelected(row) ? $t('m.Remove') : $t('m.CH_Add') }}
              </el-button>
            </div>

            <div class="selector-question-desc markdown-body preview-markdown" v-html="renderMarkdown(row.content || row.title)" v-highlight></div>

            <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="selector-question-options">
              <div v-for="(option, idx) in parseOptions(row.options)" :key="idx" class="selector-option-item">
                <span class="option-letter">{{ option.letter }}.</span>
                <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
              </div>
            </div>

            <div v-if="row.type === 'composite'" class="selector-question-options">
              <div
                v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(row.options)"
                :key="subQuestion.id || subIndex"
                class="composite-preview-item"
              >
                <div class="composite-preview-header">
                  <span>{{ $t('m.CH_Sub_Question', { index: subIndex + 1 }) }}</span>
                  <el-tag size="mini" type="warning">{{ Number(subQuestion.score || 0) }} {{ $t('m.CH_Score_Unit') }}</el-tag>
                </div>
                <div class="markdown-body preview-markdown" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
                <div class="selected-question-options">
                  <div v-for="(option, idx) in parseOptions(subQuestion.options)" :key="idx" class="selected-question-option">
                    <span class="option-letter">{{ option.letter }}.</span>
                    <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                  </div>
                </div>
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
            <div class="selected-question-answer answer-info compact-answer-info" v-if="row && row.type">
              <span class="meta-label">{{ $t('m.CH_Correct_Answer') }}</span>
              <span class="meta-value">{{ formatAnswerForDisplay(row) }}</span>
            </div>
            <div v-if="row.analysis" class="selector-analysis markdown-body preview-markdown" v-html="renderMarkdown(row.analysis)" v-highlight></div>
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
        <el-button @click="showQuestionSelectorDialog = false">{{ $t('m.Close') }}</el-button>
      </span>
    </el-dialog>

    <!-- 添加 BingOJ 编程题对话框 -->
    <el-dialog :title="$t('m.CH_Add_Programming_Question')" :visible.sync="showAddProgrammingDialog" width="1100px">
      <el-form :model="programmingForm" label-width="120px">
        <el-form-item :label="$t('m.CH_Input_Mode')">
          <el-radio-group v-model="programmingInputMode" @change="handleProgrammingInputModeChange">
            <el-radio label="manual">{{ $t('m.CH_Manual_Input_Id') }}</el-radio>
            <el-radio label="tag">{{ $t('m.CH_Select_By_Tag') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 手动输入模式 -->
        <template v-if="programmingInputMode === 'manual'">
          <el-form-item :label="$t('m.CH_BingOJ_Problem_Id')" required>
            <el-input v-model="programmingForm.problemId" :placeholder="$t('m.CH_Enter_BingOJ_Id')" style="width: 300px;" />
            <el-button
              type="primary"
              icon="el-icon-search"
              style="margin-left: 10px;"
              @click="fetchProgrammingProblemInfo"
              :loading="fetchingProblem"
            >
              {{ $t('m.CH_Fetch_Problem_Info') }}
            </el-button>
            <div style="margin-top: 8px; color: #909399; font-size: 12px;">
              <i class="el-icon-info"></i>
              {{ $t('m.CH_BingOJ_Id_Hint') }}
            </div>
          </el-form-item>
        </template>

        <!-- 标签选择模式 -->
        <template v-if="programmingInputMode === 'tag'">
          <el-form-item :label="$t('m.CH_Select_Tags')">
            <div v-if="problemTagsLoading" v-loading="true" style="min-height: 100px;"></div>
            <div v-else>
              <div v-for="(tagsAndClassification, index) in problemTagsAndClassificationList" :key="index" style="margin-bottom: 15px;">
                <div style="margin-bottom: 8px; font-weight: bold; color: #606266;">
                  {{ tagsAndClassification.classification ? tagsAndClassification.classification.name : $t('m.CH_Uncategorized') }}
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
          <el-form-item v-if="selectedProblemTagIds.length > 0" :label="$t('m.CH_Selected_Tags')">
            <el-tag
              v-for="tagId in selectedProblemTagIds"
              :key="tagId"
              closable
              @close="removeSelectedTag(tagId)"
              :type="getTagById(tagId)?.type || 'primary'"
              :color="getTagById(tagId)?.color || '#409eff'"
              effect="dark"
              style="margin-right: 10px; margin-bottom: 10px;"
              size="medium"
            >
              {{ getTagName(tagId) }}
            </el-tag>
            <el-button type="text" size="small" @click="clearAllTags" style="margin-left: 10px;">{{ $t('m.CH_Clear') }}</el-button>
          </el-form-item>

          <!-- 标签筛选结果 -->
          <el-form-item v-if="filteredProblemsByTag.length > 0" :label="$t('m.CH_Problem_List')">
            <el-alert
              type="info"
              :closable="false"
              style="margin-bottom: 10px;"
            >
              <span slot="title">
                {{ $t('m.CH_Tag_Problems_Before') }}<strong>{{ filteredProblemsTotal }}</strong>{{ $t('m.CH_Tag_Problems_Middle') }}{{ filteredProblemsByTag.length }}{{ $t('m.CH_Tag_Problems_After') }}
              </span>
            </el-alert>
            <el-table
              :data="filteredProblemsByTag"
              stripe
              border
              max-height="300"
              style="width: 100%"
            >
              <el-table-column prop="problemId" :label="$t('m.CH_Problem_Number')" width="120">
                <template slot-scope="{ row }">
                  <el-link type="primary" @click="viewProblemDetail(row)">{{ row.problemId }}</el-link>
                </template>
              </el-table-column>
              <el-table-column prop="title" :label="$t('m.CH_Problem_Title')" min-width="200" show-overflow-tooltip></el-table-column>
              <el-table-column prop="difficulty" :label="$t('m.Difficulty')" width="80" align="center">
                <template slot-scope="{ row }">
                  <el-tag :type="getDifficultyTagType(row.difficulty)" size="mini">
                    {{ getDifficultyName(row.difficulty) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="$t('m.Operation')" width="120" align="center">
                <template slot-scope="{ row }">
                  <el-button type="primary" size="mini" @click="selectProblemByTag(row)">
                    {{ $t('m.CH_Add_This_Problem') }}
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

        <!-- 题目详情查看弹窗 -->
        <el-dialog
          :title="$t('m.Question_Detail')"
          :visible.sync="showTagProblemDetailDialog"
          width="900px"
          append-to-body
        >
          <div v-if="tagProblemDetail" v-loading="loadingTagProblemDetail">
            <div class="problem-detail-content">
              <h3>{{ tagProblemDetail.title }}</h3>
              <div class="problem-meta">
                <el-tag size="small">{{ $t('m.CH_Problem_Id_Label', { id: tagProblemDetail.problemId }) }}</el-tag>
                <el-tag size="small" type="info">{{ $t('m.CH_Time_Limit_Label', { limit: tagProblemDetail.timeLimit }) }}</el-tag>
                <el-tag size="small" type="warning">{{ $t('m.CH_Memory_Limit_Label', { limit: tagProblemDetail.memoryLimit }) }}</el-tag>
                <el-tag size="small" type="success">{{ $t('m.CH_Difficulty_Label', { name: getDifficultyName(tagProblemDetail.difficulty) }) }}</el-tag>
              </div>
              <div class="problem-body">
                <div class="content-section">
                  <h4>{{ $t('m.Problem_Description') }}</h4>
                  <div class="markdown-body preview-markdown" v-html="renderMarkdown(tagProblemDetail.description)" v-highlight></div>
                </div>
                <div v-if="tagProblemDetail.input" class="content-section">
                  <h4>{{ $t('m.CH_Input_Format') }}</h4>
                  <div class="markdown-body preview-markdown" v-html="renderMarkdown(tagProblemDetail.input)" v-highlight></div>
                </div>
                <div v-if="tagProblemDetail.output" class="content-section">
                  <h4>{{ $t('m.CH_Output_Format') }}</h4>
                  <div class="markdown-body preview-markdown" v-html="renderMarkdown(tagProblemDetail.output)" v-highlight></div>
                </div>
              </div>
              <div slot="footer" style="text-align: right;">
                <el-button type="primary" @click="showTagProblemDetailDialog = false">{{ $t('m.Close') }}</el-button>
              </div>
            </div>
          </div>
        </el-dialog>

        <!-- 题目预览区域 -->
        <div v-if="programmingProblemPreview" class="problem-preview">
          <el-divider content-position="left">{{ $t('m.CH_Problem_Preview') }}</el-divider>
          <el-card>
            <h3>{{ programmingProblemPreview.problem.title }}</h3>
            <div class="problem-meta">
              <el-tag size="small">{{ $t('m.CH_Problem_Id_Label', { id: programmingProblemPreview.problem.problemId }) }}</el-tag>
              <el-tag size="small" type="info">{{ $t('m.CH_Time_Limit_Label', { limit: programmingProblemPreview.problem.timeLimit }) }}</el-tag>
              <el-tag size="small" type="warning">{{ $t('m.CH_Memory_Limit_Label', { limit: programmingProblemPreview.problem.memoryLimit }) }}</el-tag>
              <el-tag size="small" type="success">{{ $t('m.CH_Judge_Mode_Label', { mode: getJudgeModeText(programmingProblemPreview.problem.judgeMode) }) }}</el-tag>
            </div>
            <div class="problem-content">
              <div class="content-section">
                <h4>{{ $t('m.Problem_Description') }}</h4>
                <div class="markdown-body preview-markdown" v-html="renderMarkdown(programmingProblemPreview.problem.description)" v-highlight></div>
              </div>
              <div class="content-section" v-if="programmingProblemPreview.problem.input">
                <h4>{{ $t('m.CH_Input_Format') }}</h4>
                <div class="markdown-body preview-markdown" v-html="renderMarkdown(programmingProblemPreview.problem.input)" v-highlight></div>
              </div>
              <div class="content-section" v-if="programmingProblemPreview.problem.output">
                <h4>{{ $t('m.CH_Output_Format') }}</h4>
                <div class="markdown-body preview-markdown" v-html="renderMarkdown(programmingProblemPreview.problem.output)" v-highlight></div>
              </div>
              <div class="content-section" v-if="programmingExamples.length > 0">
                <h4>{{ $t('m.CH_Examples') }}</h4>
                <div v-for="(example, index) in programmingExamples" :key="index" class="example-item">
                  <el-alert :title="$t('m.CH_Example_N', { index: index + 1 })" type="info" :closable="false">
                    <div slot="default">
                      <p><strong>{{ $t('m.CH_Input_Colon') }}</strong></p>
                      <pre>{{ example.input }}</pre>
                      <p><strong>{{ $t('m.CH_Output_Colon') }}</strong></p>
                      <pre>{{ example.output }}</pre>
                    </div>
                  </el-alert>
                </div>
              </div>
            </div>
          </el-card>
        </div>

        <el-form-item :label="$t('m.Score')" required>
          <el-input-number v-model="programmingForm.score" :min="1" :max="100" />
        </el-form-item>
      </el-form>

      <span slot="footer">
        <el-button @click="showAddProgrammingDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button
          v-if="programmingInputMode === 'manual'"
          type="primary"
          @click="confirmAddProgrammingQuestion"
          :disabled="!programmingProblemPreview"
        >
          {{ $t('m.CH_Confirm_Add') }}
        </el-button>
      </span>
    </el-dialog>

    <!-- 导入试卷对话框 -->
    <el-dialog :title="$t('m.CH_Import_Paper')" :visible.sync="showImportPaperDialog" width="800px">
      <div v-loading="loadingPapers" :element-loading-text="$t('m.CH_Loading_Papers')">
        <el-form label-width="80px">
          <el-form-item :label="$t('m.CH_Paper_Filter')">
            <el-input
              v-model="paperSearchKeyword"
              :placeholder="$t('m.CH_Search_Paper')"
              prefix-icon="el-icon-search"
              clearable
              @clear="loadExamPaperList"
              @keyup.enter.native="loadExamPaperList"
              style="width: 300px;"
            >
              <el-button
                slot="append"
                icon="el-icon-search"
                @click="loadExamPaperList"
              >
                {{ $t('m.Search') }}
              </el-button>
            </el-input>
          </el-form-item>
        </el-form>

        <el-divider content-position="left">{{ $t('m.CH_Paper_List') }}</el-divider>

        <div v-if="examPapers.length === 0 && !loadingPapers" class="empty-papers">
          <i class="el-icon-info"></i>
          <p>{{ $t('m.CH_No_Papers') }}</p>
          <p class="hint">{{ $t('m.CH_No_Papers_Hint') }}</p>
        </div>

        <div v-else class="paper-list">
          <el-radio-group v-model="selectedPaperId" class="paper-radio-group">
            <div
              v-for="paper in examPapers"
              :key="paper.id"
              class="paper-item"
              :class="{ 'is-selected': selectedPaperId === paper.id }"
            >
              <el-radio :label="paper.id" class="paper-radio">
                <div class="paper-content">
                  <div class="paper-header">
                    <span class="paper-title">{{ paper.title }}</span>
                    <div class="paper-tags">
                      <el-tag v-if="paper.isShared === 1" size="mini" type="success">{{ $t('m.CH_Shared') }}</el-tag>
                      <el-tag v-else size="mini" type="info">{{ $t('m.CH_Private') }}</el-tag>
                      <el-tag size="mini" type="primary">{{ paper.questionCount }}{{ $t('m.CH_Question_Unit') }}</el-tag>
                      <el-tag size="mini" type="warning">{{ paper.totalScore }}{{ $t('m.CH_Score_Unit') }}</el-tag>
                    </div>
                  </div>
                  <div v-if="paper.description" class="paper-description">{{ paper.description }}</div>
                  <div class="paper-meta">
                    <span class="creator">
                      <i class="el-icon-user"></i>
                      {{ paper.creator && paper.creator.username ? paper.creator.username : (paper.creatorId || $t('m.Unknown')) }}
                    </span>
                    <span class="create-time">
                      <i class="el-icon-time"></i>
                      {{ formatPaperTime(paper.createdAt) }}
                    </span>
                  </div>
                </div>
              </el-radio>
            </div>
          </el-radio-group>
        </div>

        <el-pagination
          v-if="paperTotal > 0"
          @current-change="handlePaperPageChange"
          :current-page="paperCurrentPage"
          :page-size="paperPageSize"
          :total="paperTotal"
          layout="prev, pager, next, total"
          style="margin-top: 20px; text-align: center;"
        />
      </div>

      <span slot="footer">
        <el-button @click="showImportPaperDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button
          type="primary"
          @click="confirmImportPaper"
          :disabled="!selectedPaperId"
          :loading="importingPaper"
        >
          {{ $t('m.CH_Confirm_Import') }}
        </el-button>
      </span>
    </el-dialog>

    <!-- 全卷预览对话框 -->
    <el-dialog
      :title="$t('m.CH_Full_Preview_Title')"
      :visible.sync="showFullPreviewDialog"
      width="900px"
      top="5vh"
      :append-to-body="false"
      :close-on-click-modal="false"
      @opened="handlePreviewOpened"
    >
      <div class="full-preview-container" v-if="selectedQuestions.length > 0">
        <!-- 作业基本信息 -->
        <el-card class="preview-info" shadow="never">
          <h3>{{ form.title || $t('m.CH_Homework_Title_Placeholder') }}</h3>
          <p><strong>{{ $t('m.CH_Description_Colon') }}</strong>{{ form.description || $t('m.None') }}</p>
          <p><strong>{{ $t('m.CH_Start_Time_Colon') }}</strong>{{ formatTime(form.startTime) }}</p>
          <p><strong>{{ $t('m.CH_End_Time_Colon') }}</strong>{{ formatTime(form.endTime) }}</p>
          <p><strong>{{ $t('m.CH_Total_Score_Colon') }}</strong>{{ getTotalScore() }} {{ $t('m.CH_Score_Unit') }}</p>
          <p><strong>{{ $t('m.CH_Question_Count_Colon') }}</strong>{{ selectedQuestions.length }} {{ $t('m.CH_Question_Unit') }}</p>
        </el-card>

        <!-- 题目列表（和学生看到的一样） -->
        <div class="questions-preview">
          <el-divider content-position="left">{{ $t('m.CH_Question_Content') }}</el-divider>
          <div v-for="(item, index) in selectedQuestions" :key="item.id || item.problemId" class="question-item">
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <span class="question-type">({{ getQuestionTypeText(item.type) }})</span>
              <span class="question-score">{{ item.score }}{{ $t('m.CH_Score_Unit') }}</span>
            </div>
            <div class="question-title markdown-body preview-markdown" v-html="renderMarkdown(item.title)" v-highlight></div>
            <div
              class="question-content markdown-body preview-markdown"
              v-if="item.content"
              v-html="renderMarkdown(item.content)"
              v-highlight
            ></div>

            <!-- 单选题选项（预览模式，不可作答） -->
            <div v-if="item.type === 'single_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
                <div class="option-preview">
                  <span class="option-letter">{{ option.letter }}.</span>
                  <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                </div>
              </div>
            </div>

            <!-- 多选题选项（预览模式，不可作答） -->
            <div v-if="item.type === 'multiple_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
                <div class="option-preview">
                  <span class="option-letter">{{ option.letter }}.</span>
                  <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                </div>
              </div>
            </div>

            <!-- 判断题（预览模式，不可作答） -->
            <div v-if="item.type === 'judge'" class="question-options">
              <div class="option-preview">
                <span class="option-letter">✓</span>
                <span class="option-text">{{ $t('m.True') }}</span>
              </div>
              <div class="option-preview">
                <span class="option-letter">✗</span>
                <span class="option-text">{{ $t('m.False') }}</span>
              </div>
            </div>

            <!-- 组合题 -->
            <div v-if="item.type === 'composite'" class="question-options composite-preview-list">
              <div
                v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(item.options)"
                :key="subQuestion.id || subIndex"
                class="composite-preview-item"
              >
                <div class="composite-preview-header">
                  <span>{{ $t('m.CH_Sub_Question', { index: subIndex + 1 }) }}</span>
                  <span>{{ Number(subQuestion.score || 0) }}{{ $t('m.CH_Score_Unit') }}</span>
                </div>
                <div class="markdown-body preview-markdown" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
                <div class="question-options" style="padding-left: 0;">
                  <div v-for="(option, idx) in parseOptions(subQuestion.options)" :key="idx" class="option-item">
                    <div class="option-preview">
                      <span class="option-letter">{{ option.letter }}.</span>
                      <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 填空题 -->
            <div v-if="item.type === 'fill_blank'" class="subjective-preview">
              <el-alert type="success" :closable="false">
                <i class="el-icon-edit"></i> {{ $t('m.CH_Fill_Blank_Preview_Tip') }}
              </el-alert>
            </div>

            <!-- 主观题 -->
            <div v-if="item.type === 'subjective'" class="subjective-preview">
              <el-alert type="info" :closable="false">
                <i class="el-icon-edit"></i> {{ $t('m.CH_Subjective_Preview_Tip') }}
              </el-alert>
            </div>

            <!-- 编程题 -->
            <div v-if="item.type === 'programming'" class="programming-preview">
              <template v-if="getProgrammingProblemInfo(item.problemId)">
                <div class="problem-preview">
                  <el-divider content-position="left">{{ $t('m.CH_Programming_Detail') }}</el-divider>
                  <el-card>
                    <h3>{{ getProgrammingProblemInfo(item.problemId).problem.title }}</h3>
                    <div class="problem-meta">
                      <el-tag size="small">{{ $t('m.CH_Problem_Id_Label', { id: getProgrammingProblemInfo(item.problemId).problem.problemId }) }}</el-tag>
                      <el-tag size="small" type="info">{{ $t('m.CH_Time_Limit_Label', { limit: getProgrammingProblemInfo(item.problemId).problem.timeLimit }) }}</el-tag>
                      <el-tag size="small" type="warning">{{ $t('m.CH_Memory_Limit_Label', { limit: getProgrammingProblemInfo(item.problemId).problem.memoryLimit }) }}</el-tag>
                      <el-tag size="small" type="success">{{ $t('m.CH_Judge_Mode_Label', { mode: getJudgeModeText(getProgrammingProblemInfo(item.problemId).problem.judgeMode) }) }}</el-tag>
                    </div>
                    <div class="problem-content">
                      <div class="content-section">
                        <h4>{{ $t('m.Problem_Description') }}</h4>
                        <div class="markdown-body preview-markdown" v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.description)" v-highlight></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.input">
                        <h4>{{ $t('m.CH_Input_Format') }}</h4>
                        <div class="markdown-body preview-markdown" v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.input)" v-highlight></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.output">
                        <h4>{{ $t('m.CH_Output_Format') }}</h4>
                        <div class="markdown-body preview-markdown" v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.output)" v-highlight></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingExamplesForProblem(item.problemId).length > 0">
                        <h4>{{ $t('m.CH_Examples') }}</h4>
                        <div v-for="(example, idx) in getProgrammingExamplesForProblem(item.problemId)" :key="idx" class="example-item">
                          <el-alert :title="$t('m.CH_Example_N', { index: idx + 1 })" type="info" :closable="false">
                            <div slot="default">
                              <p><strong>{{ $t('m.CH_Input_Colon') }}</strong></p>
                              <pre>{{ example.input }}</pre>
                              <p><strong>{{ $t('m.CH_Output_Colon') }}</strong></p>
                              <pre>{{ example.output }}</pre>
                            </div>
                          </el-alert>
                        </div>
                      </div>
                    </div>
                  </el-card>
                </div>
              </template>

              <el-alert v-else type="warning" :closable="false">
                <p>{{ $t('m.CH_Programming_Preview_Tip') }}</p>
              </el-alert>

              <el-alert v-if="!item.problemId" type="warning" :closable="false">
                {{ $t('m.CH_No_BingOJ_Id') }}
              </el-alert>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 题目详情对话框 -->
    <el-dialog :title="currentQuestion?.type === 'programming' ? $t('m.CH_Programming_Detail') : $t('m.Question_Detail')" :visible.sync="showDetailDialog" width="900px">
      <div v-if="currentQuestion" class="question-detail">
        <!-- 编程题详情 -->
        <div v-if="currentQuestion.type === 'programming' && currentQuestion.problem">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="$t('m.CH_Problem_Id_Item')">
              {{ currentQuestion.problem.problemId }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.Question_Type')">
              <el-tag type="danger" size="small">{{ $t('m.Programming') }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.Question_Title')" :span="2">
              <strong>{{ currentQuestion.problem.title }}</strong>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.Time_Limit')">
              {{ currentQuestion.problem.timeLimit }} ms
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.Memory_Limit')">
              {{ currentQuestion.problem.memoryLimit }} MB
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.Problem_Description')" :span="2">
              <div v-html="renderMarkdown(currentQuestion.problem.description)" class="markdown-body detail-content preview-markdown" v-highlight></div>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.CH_Input_Format')" v-if="currentQuestion.problem.input">
              <div v-html="renderMarkdown(currentQuestion.problem.input)" class="markdown-body detail-content preview-markdown" v-highlight></div>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.CH_Output_Format')" v-if="currentQuestion.problem.output">
              <div v-html="renderMarkdown(currentQuestion.problem.output)" class="markdown-body detail-content preview-markdown" v-highlight></div>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.CH_Examples')" :span="2" v-if="currentQuestion.problem.examples">
              <div v-for="(example, idx) in parseExamples(currentQuestion.problem.examples)" :key="idx" class="example-box">
                <div class="example-title">{{ $t('m.CH_Example_N', { index: idx + 1 }) }}</div>
                <div><strong>{{ $t('m.CH_Input_Colon') }}</strong><pre>{{ example.input }}</pre></div>
                <div><strong>{{ $t('m.CH_Output_Colon') }}</strong><pre>{{ example.output }}</pre></div>
              </div>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('m.CH_Hint')" v-if="currentQuestion.problem.hint">
              <div v-html="renderMarkdown(currentQuestion.problem.hint)" class="markdown-body detail-content preview-markdown" v-highlight></div>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 客观题详情 -->
        <el-descriptions v-else :column="1" border>
          <el-descriptions-item :label="$t('m.Question_Type')">
            {{ getQuestionTypeText(currentQuestion.type) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Question_Title')">
            <div v-html="renderMarkdown(currentQuestion.title)" class="markdown-body detail-content preview-markdown" v-highlight></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Content')" v-if="currentQuestion.content">
            <div v-html="renderMarkdown(currentQuestion.content)" class="markdown-body detail-content preview-markdown" v-highlight></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Options')" v-if="currentQuestion.options">
            <div v-if="currentQuestion.type !== 'composite'" v-html="renderOptions(currentQuestion.options)" class="markdown-body detail-content preview-markdown" v-highlight></div>
            <div v-else class="composite-detail-list">
              <div
                v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(currentQuestion.options)"
                :key="subQuestion.id || subIndex"
                class="composite-detail-item"
              >
                <div class="composite-detail-header">
                  <span>{{ $t('m.CH_Sub_Question', { index: subIndex + 1 }) }}</span>
                  <span>{{ Number(subQuestion.score || 0) }}{{ $t('m.CH_Score_Unit') }}</span>
                </div>
                <div class="markdown-body detail-content preview-markdown" v-html="renderMarkdown(subQuestion.content || '')" v-highlight></div>
                <div class="composite-detail-options">
                  <div v-for="(option, optionIndex) in parseOptions(subQuestion.options)" :key="optionIndex" class="composite-detail-option">
                    <span class="option-letter">{{ option.letter }}.</span>
                    <span class="option-text markdown-body preview-markdown" v-html="renderMarkdown(option.text)" v-highlight></span>
                  </div>
                </div>
                <div class="composite-detail-answer answer-info compact-answer-info">
                  <el-tag size="mini" type="success">
                    {{ $t('m.CH_Correct_Answer') }}{{ getCompositeCorrectAnswer(currentQuestion.answer, subQuestion.id, subIndex) }}
                  </el-tag>
                </div>
              </div>
              <el-empty
                v-if="parseCompositeSubQuestions(currentQuestion.options).length === 0"
                :description="$t('m.CH_No_Composite_Data')"
                :image-size="100"
              />
            </div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Answer')">
            <div class="answer-info compact-answer-info">
              <div
                v-if="currentQuestion.type === 'subjective'"
                class="markdown-body detail-content preview-markdown"
                v-html="renderMarkdown(formatAnswerForDisplay(currentQuestion))"
                v-highlight
              ></div>
              <span v-else>{{ formatAnswerForDisplay(currentQuestion) }}</span>
            </div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Difficulty')">
            <el-rate :value="getDifficultyStars(currentQuestion.difficulty)" :max="3" disabled />
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Score')">
            {{ currentQuestion.score }} {{ $t('m.CH_Score_Unit') }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.CH_Course')" v-if="currentQuestion.course">
            {{ currentQuestion.course }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Tags')" v-if="currentQuestion.tags">
            <el-tag v-for="(tag, idx) in parseQuestionTags(currentQuestion.tags)" :key="idx" size="mini" type="info" style="margin-right: 5px;">
              {{ tag }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import teacherAuth from '@/mixins/teacherAuth'
import moment from 'moment'
import { getClassroomProblem } from '@/common/classroomProblem'
import api from '@/common/api'
import MarkdownIt from 'markdown-it'
import MarkdownItKatex from '@iktakahiro/markdown-it-katex'

// 配置 markdown-it 和 KaTeX
const md = new MarkdownIt({
  html: true,
  linkify: true,
  // 配置不要把单独的 ^ 当作上标
  typographer: false
})
// 配置 KaTeX，只在 $...$ 或 $$...$$ 中渲染数学公式
md.use(MarkdownItKatex, {
  throwOnError: false,
  errorColor: '#cc0000',
  strict: false,
  // 禁用自动识别上标/下标，避免 ^9 被当作上标渲染
  enableSuperscript: false,
  enableSubscript: false
})

export default {
  name: 'CreateHomework',
  mixins: [teacherAuth],
  data() {
    // 从路由参数读取初始 isExamMode，避免页面闪烁
    const initialIsExamMode = this.$route.query.isExamMode ? parseInt(this.$route.query.isExamMode) : 0

    return {
      submitting: false,
      loadingData: false, // 编辑模式下加载数据的loading状态
      questionsLoading: false,
      showQuestionSelectorDialog: false,
      questionBank: [],
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
      selectedQuestions: [],
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
      showDetailDialog: false,
      showFullPreviewDialog: false, // 全卷预览对话框
      currentQuestion: null,
      // BingOJ 编程题相关
      showAddProgrammingDialog: false,
      programmingInputMode: 'manual', // 'manual' 或 'tag'
      programmingForm: {
        problemId: '',
        score: 10
      },
      programmingProblemPreview: null,
      programmingExamples: [],
      fetchingProblem: false,
      programmingProblemsCache: {}, // 缓存编程题信息，用于预览
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
      // 导入试卷相关
      showImportPaperDialog: false,
      examPapers: [],
      loadingPapers: false,
      paperSearchKeyword: '',
      selectedPaperId: null,
      paperCurrentPage: 1,
      paperPageSize: 10,
      paperTotal: 0,
      importingPaper: false,
      form: {
        title: '',
        description: '',
        startTime: null,
        endTime: null,
        showHomework: false,
        showScore: false,
        showRank: false,
        showAnswer: false,
        // 考试模式字段 - 使用路由参数作为初始值
        isExamMode: initialIsExamMode,
        examDuration: 60,
        allowSubmitAfterMinutes: 0,
        disableCopyPaste: true,
        requireFullscreen: true,
        disallowTabSwitch: true
      },
      rules: {
        title: [{ required: true, message: this.$t('m.Please_Enter_Homework_Title'), trigger: 'blur' }],
        startTime: [{ required: true, message: this.$t('m.Please_Select_Start_Time'), trigger: 'change' }],
        endTime: [{ required: true, message: this.$t('m.Please_Select_End_Time'), trigger: 'change' }]
      }
    }
  },
  computed: {
    classroomId() {
      return this.$route.params.classroomId
    },
    // 判断是否是编辑模式（支持教师端和管理员端两种路由）
    isEditMode() {
      return !!(this.$route.query.editId || this.$route.params.homeworkId)
    },
    // 获取作业ID（支持教师端 query 参数和管理员端 params 参数）
    editId() {
      return this.$route.query.editId || this.$route.params.homeworkId
    },
    // 判断是否在管理员路由下
    isAdminRoute() {
      return this.$route.path.startsWith('/admin/')
    },
    currentEditorPage() {
      return this.$route.query.page === 'questions' ? 'questions' : 'config'
    },
    isConfigPage() {
      return this.currentEditorPage === 'config'
    },
    isQuestionPage() {
      return this.currentEditorPage === 'questions'
    },
    // 计算作业时间区间（分钟）
    timeRangeMinutes() {
      if (!this.form.startTime || !this.form.endTime) return 0
      const start = new Date(this.form.startTime)
      const end = new Date(this.form.endTime)
      const diffMs = end - start
      const diffMinutes = Math.floor(diffMs / (1000 * 60))
      return diffMinutes > 0 ? diffMinutes : 0
    },
    // 时间区间和考试时长是否不匹配
    timeRangeMismatch() {
      return this.timeRangeMinutes > 0 && this.timeRangeMinutes !== this.form.examDuration
    },
    filteredQuestionBank() {
      if (!Array.isArray(this.questionBank)) {
        return []
      }
      if (!this.questionFilters.onlyUnselected) {
        return this.questionBank
      }
      return this.questionBank.filter(question => !this.isQuestionSelected(question))
    }
  },
  mounted() {
    this.loadProblemTagsAndClassification()
    // 如果是编辑模式，加载作业数据
    if (this.isEditMode) {
      this.loadHomeworkData()
    }
  },
  methods: {
    switchEditorPage(page) {
      const targetPage = page === 'questions' ? 'questions' : 'config'
      this.$router.replace({
        path: this.$route.path,
        query: {
          ...this.$route.query,
          page: targetPage
        }
      })
    },
    openQuestionSelectorDialog() {
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
    async loadQuestionBank() {
      this.questionsLoading = true
      try {
        const questionId = String(this.questionFilters.questionId || '').trim()
        const keyword = String(this.questionFilters.keyword || '').trim()
        const sortParams = this.getQuestionSortParams()
        const params = {
          classroomId: this.classroomId,
          page: this.questionCurrentPage,
          limit: this.questionPageSize,
          type: this.questionFilters.type || undefined,
          course: this.questionFilters.course || undefined,
          tag: this.questionFilters.tag || undefined,
          sortBy: sortParams.sortBy,
          sortOrder: sortParams.sortOrder
        }

        if (questionId) {
          params.questionId = questionId
        } else if (keyword) {
          params.keyword = keyword
        }

        const res = await this.$store.dispatch('classroom/getQuestionBank', params)
        if (res.code === 200) {
          const questions = res.data.questions || res.data || []
          this.questionBank = questions.map(q => ({
            ...q,
            difficulty: parseInt(q.difficulty) || 2, // 确保是数字类型
            score: q.score || 10 // 默认分值
          }))
          this.questionBankTotal = res.data.total || questions.length
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.questionsLoading = false
      }
    },
    handleQuestionPageChange(page) {
      this.questionCurrentPage = page
      this.loadQuestionBank()
    },
    toggleObjectiveQuestion(question) {
      if (!question || !question.id) {
        return
      }
      if (this.isQuestionSelected(question)) {
        this.removeQuestion(question)
        this.$message.success(this.$t('m.CH_Removed_Question'))
      } else {
        this.addQuestion(question)
      }
    },
    isQuestionSelected(question) {
      // 检查题目是否已添加（通过id或problemId）
      return this.selectedQuestions.some(sq => {
        // 客观题：使用 id 判断
        if (question.id && sq.id === question.id) {
          return true
        }
        // 编程题：使用 problemId 判断
        if (question.problemId && sq.problemId === question.problemId) {
          return true
        }
        // 兼容：question 有 questionId 字段的情况
        if (question.questionId && sq.id === question.questionId) {
          return true
        }
        return false
      })
    },
    addQuestion(question) {
      // 检查是否已添加
      if (this.isQuestionSelected(question)) {
        this.$message.warning(this.$t('m.CH_Question_Already_Added'))
        return
      }

      // 添加题目（深拷贝，避免修改原数据）
      const newQuestion = JSON.parse(JSON.stringify(question))

      // 设置默认分数（如果没有分数或分数为0）
      if (!newQuestion.score || newQuestion.score === 0) {
        // 根据题型设置默认分数
        const defaultScores = {
          single_choice: 2,      // 单选题默认2分
          multiple_choice: 5,    // 多选题默认5分
          judge: 1,              // 判断题默认1分
          fill_blank: 2,         // 填空题默认2分
          composite: 10,         // 组合题默认10分
          subjective: 5,         // 主观题默认5分
          programming: 20        // 编程题默认20分
        }
        newQuestion.score = defaultScores[newQuestion.type] || 10
      }

      this.selectedQuestions.push(newQuestion)

      this.$message.success(this.$t('m.CH_Add_Success'))
    },
    removeQuestion(question) {
      // 根据题目类型查找索引
      const index = this.selectedQuestions.findIndex(q => {
        // 客观题：通过 id 或 questionId 判断
        if (question.id) {
          return q.id === question.id || q.questionId === question.id
        }
        // 编程题：通过 problemId 判断
        if (question.problemId) {
          return q.problemId === question.problemId
        }
        return false
      })
      if (index > -1) {
        this.selectedQuestions.splice(index, 1)
      }
    },
    moveUp(index) {
      if (index > 0) {
        const temp = this.selectedQuestions[index - 1]
        this.$set(this.selectedQuestions, index - 1, this.selectedQuestions[index])
        this.$set(this.selectedQuestions, index, temp)
      }
    },
    moveDown(index) {
      if (index < this.selectedQuestions.length - 1) {
        const temp = this.selectedQuestions[index + 1]
        this.$set(this.selectedQuestions, index + 1, this.selectedQuestions[index])
        this.$set(this.selectedQuestions, index, temp)
      }
    },
    viewQuestion(question) {
      // 深拷贝题目对象，确保不影响原数据
      this.currentQuestion = {
        ...question,
        difficulty: parseInt(question.difficulty) || 2 // 确保是数字类型
      }
      this.showDetailDialog = true
    },
    // 查看题目详情（支持客观题和编程题）
    async viewQuestionDetail(question) {
      if (question.type === 'programming') {
        // 编程题：需要加载完整信息
        if (!this.programmingProblemsCache[question.problemId]) {
          // 缓存中没有，需要先加载
          try {
            const loading = this.$loading({
              lock: true,
              text: this.$t('m.CH_Loading_Question'),
              spinner: 'el-icon-loading',
              background: 'rgba(0, 0, 0, 0.7)'
            })
            await this.loadProgrammingProblemByIds([question.problemId])
            loading.close()
          } catch (error) {
            this.$message.error(this.$t('m.CH_Load_Question_Failed'))
            return
          }
        }
        // 从缓存获取完整题目信息
        const fullProblem = this.programmingProblemsCache[question.problemId]
        this.currentQuestion = {
          ...question,
          problem: fullProblem
        }
      } else {
        // 客观题：直接显示
        this.currentQuestion = {
          ...question,
          difficulty: parseInt(question.difficulty) || 2
        }
      }
      this.showDetailDialog = true
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        fill_blank: this.$t('m.CH_Fill_Blank'),
        composite: this.$t('m.CH_Composite'),
        subjective: this.$t('m.Subjective'),
        programming: this.$t('m.Programming')
      }
      return map[type] || type
    },
    getQuestionTypeColor(type) {
      const colorMap = {
        single_choice: 'primary',
        multiple_choice: 'success',
        judge: 'warning',
        fill_blank: 'success',
        composite: 'danger',
        subjective: 'info',
        programming: 'danger'
      }
      return colorMap[type] || 'info'
    },
    getDifficultyStars(difficulty) {
      return parseInt(difficulty) || 1
    },
    parseMaybeSerializedJson(rawValue, maxDepth = 2) {
      if (rawValue === null || rawValue === undefined) return rawValue
      let current = rawValue
      for (let i = 0; i < maxDepth; i++) {
        if (typeof current !== 'string') break
        const trimmed = current.trim()
        if (!trimmed) return ''
        const looksLikeJson = (
          (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
          (trimmed.startsWith('[') && trimmed.endsWith(']')) ||
          (trimmed.startsWith('"') && trimmed.endsWith('"'))
        )
        if (!looksLikeJson) return trimmed
        try {
          current = JSON.parse(trimmed)
        } catch (e) {
          return trimmed
        }
      }
      return current
    },
    normalizeTextValue(rawValue) {
      const parsed = this.parseMaybeSerializedJson(rawValue)
      if (parsed === null || parsed === undefined) return ''
      if (typeof parsed === 'string') return parsed
      if (typeof parsed === 'number' || typeof parsed === 'boolean') return String(parsed)
      return ''
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
    // 解析题目标签
    parseQuestionTags(tags) {
      const parsed = this.parseMaybeSerializedJson(tags)
      if (!Array.isArray(parsed)) return []
      return parsed.map(tag => String(tag || '').trim()).filter(Boolean)
    },
    renderOptions(options) {
      if (!options) return '-'
      const parsedOptions = this.parseOptions(options)
      if (parsedOptions.length > 0) {
        return parsedOptions.map(opt => (
          `<div class="option-line"><strong>${opt.letter}.</strong> ${this.renderMarkdown(opt.text)}</div>`
        )).join('')
      }
      return this.renderMarkdown(this.normalizeTextValue(options))
    },
    parseAnswerArray(answerInput, { allowCommaSplit = true } = {}) {
      const parsed = this.parseMaybeSerializedJson(answerInput)
      if (Array.isArray(parsed)) {
        return parsed.map(item => String(item || '').trim()).filter(Boolean)
      }
      const raw = String(parsed || '').trim()
      if (!raw) return []
      if (allowCommaSplit && (raw.includes(',') || raw.includes('，'))) {
        return raw.split(/[，,]/).map(item => item.trim()).filter(Boolean)
      }
      return [raw]
    },
    parseCompositeAnswerMap(answerInput) {
      const parsed = this.parseMaybeSerializedJson(answerInput)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed
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
    formatAnswerForDisplay(question) {
      if (!question) return '-'
      switch (question.type) {
        case 'single_choice': {
          const value = String(this.parseMaybeSerializedJson(question.answer) || '').trim()
          return value || '-'
        }
        case 'multiple_choice': {
          const answers = this.parseAnswerArray(question.answer)
          return answers.length > 0 ? answers.join('、') : '-'
        }
        case 'judge': {
          const normalized = String(this.parseMaybeSerializedJson(question.answer) || '').trim().toLowerCase()
          if (['true', '1', 'yes', 'y', '正确'].includes(normalized)) return this.$t('m.True')
          if (['false', '0', 'no', 'n', '错误'].includes(normalized)) return this.$t('m.False')
          return '-'
        }
        case 'fill_blank': {
          const answers = this.parseAnswerArray(question.answer, { allowCommaSplit: false })
          return answers.length > 0 ? answers.join(' / ') : this.$t('m.CH_No_Answer_Set')
        }
        case 'subjective': {
          const value = this.normalizeTextValue(question.answer).trim()
          return value || this.$t('m.CH_No_Reference_Answer')
        }
        case 'composite':
          return this.$t('m.CH_See_Sub_Answers')
        default: {
          const value = this.normalizeTextValue(question.answer).trim()
          return value || '-'
        }
      }
    },
    // 解析编程题样例
    parseExamples(examples) {
      if (!examples) return []
      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      const result = []
      let match
      while ((match = regex.exec(examples)) !== null) {
        result.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }
      return result
    },
    async loadHomeworkData() {
      this.loadingData = true
      try {
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', this.editId)
        if (res.code === 200 && res.data) {
          const homework = res.data

          // 先填充表单数据（包括 isExamMode），立即更新界面
          this.form = {
            title: homework.title,
            description: homework.description,
            startTime: new Date(homework.startTime),
            endTime: new Date(homework.endTime),
            showHomework: homework.showHomework === 1,
            showScore: homework.showScore === 1,
            showRank: homework.showRank === 1,
            showAnswer: homework.showAnswer === 1,
            // 考试模式字段
            isExamMode: homework.isExamMode || 0,
            examDuration: homework.examDuration || 60,
            allowSubmitAfterMinutes: homework.allowSubmitAfterMinutes || 0,
            disableCopyPaste: homework.disableCopyPaste !== 0,
            requireFullscreen: homework.requireFullscreen !== 0,
            disallowTabSwitch: homework.disallowTabSwitch !== 0
          }

          // 填充已选题目
          if (homework.questions && homework.questions.length > 0) {
            this.selectedQuestions = homework.questions.map((item, index) => {
              // 编程题：使用 problemId
              if (item.problemId) {
                // 从缓存获取编程题详情
                const cached = this.programmingProblemsCache[item.problemId]
                const score = item.score || item.question?.score || 20
                if (cached && cached.problem) {
                  return {
                    id: item.id,
                    problemId: item.problemId,
                    title: cached.problem.title,
                    type: 'programming',
                    difficulty: 5,
                    score: score,
                    content: cached.problem.description || '',
                    input: cached.problem.input || '',
                    output: cached.problem.output || '',
                    examples: cached.problem.examples || '',
                    hint: cached.problem.hint || '',
                    judgeMode: cached.problem.judgeMode || '',
                    timeLimit: cached.problem.timeLimit || 0,
                    memoryLimit: cached.problem.memoryLimit || 0
                  }
                }
                // 如果缓存中没有，返回基本信息（不影响编辑）
                return {
                  id: item.id,
                  problemId: item.problemId,
                  title: this.$t('m.CH_BingOJ_Problem_Title', { id: item.problemId }),
                  type: 'programming',
                  difficulty: 5,
                  score: score
                }
              }
              // 普通题目：使用 question
              if (item.question && item.question.id) {
                return {
                  id: item.question.id,
                  title: item.question.title,
                  type: item.question.type,
                  difficulty: parseInt(item.question.difficulty) || 2,
                  score: item.score || item.question.score || 10,
                  content: item.question.content,
                  options: item.question.options,
                  answer: item.question.answer
                }
              }
              // 兜底：如果既没有 problemId 也没有 question，跳过
              console.warn('  -> 跳过无效题目:', item)
              return null
            }).filter(q => q !== null)
          }

          // 关键数据填充完成，隐藏 loading
          this.loadingData = false

          // 异步加载编程题详情（不阻塞界面）
          const programmingProblemIds = homework.questions
            .filter(q => q.problemId)
            .map(q => q.problemId)

          if (programmingProblemIds.length > 0) {
            // 不使用 await，让加载在后台进行，完成后会自动更新界面
            this.loadProgrammingProblemByIds(programmingProblemIds)
          }
        }
      } catch (error) {
        console.error('加载作业数据失败:', error)
        this.$message.error(this.$t('m.CH_Load_Homework_Failed') + ': ' + (error.message || error))
      } finally {
        this.loadingData = false
      }
    },
    async saveHomework() {
      if (!this.$refs.homeworkForm) {
        this.$message.error(this.$t('m.CH_Form_Not_Ready'))
        return
      }

      this.$refs.homeworkForm.validate(async (valid) => {
        if (!valid) {
          this.$message.warning(this.$t('m.CH_Fill_All_Required'))
          return
        }
        if (this.selectedQuestions.length === 0) {
          this.$message.warning(this.$t('m.Please_Select_At_Least_One_Question'))
          return
        }

        this.submitting = true
        const data = {
          classroomId: Number(this.classroomId),
          title: this.form.title,
          description: this.form.description || '',
          startTime: moment(this.form.startTime).format(),
          endTime: moment(this.form.endTime).format(),
          showHomework: this.form.showHomework ? 1 : 0,
          showScore: this.form.showScore ? 1 : 0,
          showRank: this.form.showRank ? 1 : 0,
          showAnswer: this.form.showAnswer ? 1 : 0,
          // 考试模式字段
          isExamMode: this.form.isExamMode,
          examDuration: this.form.examDuration,
          allowSubmitAfterMinutes: this.form.allowSubmitAfterMinutes,
          disableCopyPaste: this.form.disableCopyPaste ? 1 : 0,
          requireFullscreen: this.form.requireFullscreen ? 1 : 0,
          disallowTabSwitch: this.form.disallowTabSwitch ? 1 : 0,
          questions: this.selectedQuestions.map((q, index) => {
            const result = {
              score: Number(q.score) || 10,
              questionOrder: index + 1,
              questionType: q.type // 添加题目类型，用于后端设置默认分数
            }
            // 编程题使用 problemId,其他题型使用 questionId
            if (q.type === 'programming' && q.problemId) {
              result.problemId = q.problemId
              result.questionId = null
            } else {
              const rawQuestionId = q.questionId !== undefined && q.questionId !== null ? q.questionId : q.id
              const numericQuestionId = Number(rawQuestionId)
              result.questionId = (Number.isInteger(numericQuestionId) && numericQuestionId > 0) ? numericQuestionId : null
              result.problemId = null
            }
            return result
          })
        }

        const invalidQuestion = data.questions.find(item => !item.problemId && !item.questionId)
        if (invalidQuestion) {
          this.$message.error(this.$t('m.CH_Invalid_Question_Id'))
          this.submitting = false
          return
        }

        try {
          let res
          if (this.isEditMode) {
            // 编辑模式：调用更新接口
            data.id = Number(this.editId)
            res = await this.$store.dispatch('classroom/updateHomework', data)
          } else {
            // 创建模式：调用创建接口
            res = await this.$store.dispatch('classroom/createHomework', data)
          }

          if (res.code === 200) {
            this.$message.success(this.isEditMode ? this.$t('m.Update_Success') : this.$t('m.Create_Success'))
            this.goBack()
          } else {
            this.$message.error(res.message || (this.isEditMode ? this.$t('m.CH_Save_Failed') : this.$t('m.CH_Create_Failed')))
          }
        } catch (error) {
          console.error('保存作业失败:', error)
          this.$message.error(this.isEditMode ? this.$t('m.CH_Save_Failed') : this.$t('m.CH_Create_Failed'))
        } finally {
          this.submitting = false
        }
      })
    },
    // 获取 HOJ 题目信息
    async fetchProgrammingProblemInfo() {
      console.log('=== 开始获取 HOJ 题目信息 ===')
      console.log('题目ID:', this.programmingForm.problemId)

      if (!this.programmingForm.problemId) {
        this.$message.warning(this.$t('m.CH_Enter_BingOJ_Id'))
        return
      }

      // 获取当前登录用户信息
      const userInfo = this.$store.getters.userInfo
      console.log('当前用户信息:', userInfo)

      if (!userInfo || !userInfo.username) {
        this.$message.warning(this.$t('m.CH_Please_Login'))
        return
      }

      // 从 localStorage 获取 token
      const token = localStorage.getItem('token')
      console.log('当前 Token:', token ? token.substring(0, 20) + '...' : '无')

      if (!token) {
        this.$message.warning(this.$t('m.CH_No_Token'))
        return
      }

      this.fetchingProblem = true
      try {
        const requestData = {
          pid: this.programmingForm.problemId,
          cid: '0',
          mode: 'normal',
          username: userInfo.username,
          token: token, // 使用 token 而不是密码
          password: ''
        }
        console.log('请求HOJ API, 参数:', { ...requestData, token: token.substring(0, 20) + '...' })

        const res = await getClassroomProblem(requestData)
        console.log('HOJ API 响应:', res)

        if (res.code === 200 && res.data) {
          this.programmingProblemPreview = res.data
          this.extractProgrammingExamples()
          this.$message.success(this.$t('m.CH_Fetch_Success'))
        } else {
          this.$message.error(res.message || this.$t('m.CH_Fetch_Failed'))
        }
      } catch (error) {
        console.error('获取题目异常:', error)
        this.$message.error(this.$t('m.CH_Fetch_Network_Error'))
      } finally {
        this.fetchingProblem = false
      }
    },
    // 提取样例
    extractProgrammingExamples() {
      if (!this.programmingProblemPreview || !this.programmingProblemPreview.problem.examples) {
        this.programmingExamples = []
        return
      }

      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      const examples = []
      let match

      while ((match = regex.exec(this.programmingProblemPreview.problem.examples)) !== null) {
        examples.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }

      this.programmingExamples = examples
    },
    // 确认添加编程题
    confirmAddProgrammingQuestion() {
      if (!this.programmingForm.problemId) {
        this.$message.warning(this.$t('m.CH_Enter_BingOJ_Id'))
        return
      }

      // 必须先预览题目
      if (!this.programmingProblemPreview) {
        this.$message.warning(this.$t('m.CH_Preview_First'))
        return
      }

      // 使用预览获取的真实题目信息
      const tempQuestion = {
        id: `hoj_${this.programmingForm.problemId}`, // 临时ID,使用 hoj_ 前缀
        problemId: this.programmingForm.problemId, // HOJ 题目ID,保持字符串类型
        title: this.programmingProblemPreview.problem.title, // 使用真实标题
        type: 'programming',
        difficulty: 5, // 默认难度
        score: this.programmingForm.score,
        content: this.$t('m.CH_BingOJ_Id_Content', { id: this.programmingForm.problemId })
      }

      // 检查是否已经添加过该题目
      const exists = this.selectedQuestions.some(q => q.problemId === this.programmingForm.problemId)
      if (exists) {
        this.$message.warning(this.$t('m.CH_Question_Already_Added'))
        return
      }

      // 添加到已选题目
      this.selectedQuestions.push(tempQuestion)
      this.$message.success(this.$t('m.CH_Add_Success'))

      // 关闭对话框并重置
      this.showAddProgrammingDialog = false
      this.resetProgrammingForm()
    },
    // 重置编程题表单
    resetProgrammingForm() {
      this.programmingForm = {
        problemId: '',
        score: 10
      }
      this.programmingProblemPreview = null
      this.programmingExamples = []
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
        this.programmingExamples = []
      } else {
        // 切换到手动模式时，清空标签选择的内容
        this.selectedProblemTagIds = []
        this.filteredProblemsByTag = []
        this.filteredProblemsTotal = 0
      }
    },
    // 检查标签是否已选中
    isTagSelected(tagId) {
      return this.selectedProblemTagIds.includes(tagId)
    },
    // 切换标签选中状态
    async toggleProblemTag(tag) {
      const index = this.selectedProblemTagIds.indexOf(tag.id)
      if (index > -1) {
        // 已选中，取消选中
        this.selectedProblemTagIds.splice(index, 1)
      } else {
        // 未选中，添加到已选列表
        this.selectedProblemTagIds.push(tag.id)
      }
      // 重置到第一页并重新加载题目列表
      this.filteredProblemsCurrentPage = 1
      await this.loadProblemsByTags()
    },
    // 移除选中的标签
    removeSelectedTag(tagId) {
      const index = this.selectedProblemTagIds.indexOf(tagId)
      if (index > -1) {
        this.selectedProblemTagIds.splice(index, 1)
        this.filteredProblemsCurrentPage = 1
        this.loadProblemsByTags()
      }
    },
    // 清空所有标签
    clearAllTags() {
      this.selectedProblemTagIds = []
      this.filteredProblemsByTag = []
      this.filteredProblemsTotal = 0
      this.filteredProblemsCurrentPage = 1
    },
    // 根据已选标签加载题目
    async loadProblemsByTags() {
      if (this.selectedProblemTagIds.length === 0) {
        this.filteredProblemsByTag = []
        this.filteredProblemsTotal = 0
        return
      }

      try {
        const res = await api.getProblemList({
          oj: 'ME',
          tagId: this.selectedProblemTagIds.join(','),
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
        this.$message.error(this.$t('m.CH_Load_Problem_List_Failed'))
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
      for (const group of this.problemTagsAndClassificationList) {
        const tag = group.tagList.find(t => t.id === tagId)
        if (tag) return tag.name
      }
      return ''
    },
    // 获取标签对象
    getTagById(tagId) {
      for (const group of this.problemTagsAndClassificationList) {
        const tag = group.tagList.find(t => t.id === tagId)
        if (tag) return tag
      }
      return null
    },
    // 查看标签下的题目详情
    async viewProblemDetail(problem) {
      this.loadingTagProblemDetail = true
      this.showTagProblemDetailDialog = true
      this.tagProblemDetail = null

      try {
        const userInfo = this.$store.getters.userInfo
        const token = localStorage.getItem('token')

        if (!userInfo || !token) {
          this.$message.warning(this.$t('m.CH_Please_Login'))
          this.showTagProblemDetailDialog = false
          return
        }

        const res = await getClassroomProblem({
          pid: problem.problemId,
          cid: '0',
          mode: 'normal',
          username: userInfo.username,
          token: token,
          password: ''
        })

        if (res.code === 200 && res.data) {
          this.tagProblemDetail = res.data.problem
        } else {
          this.$message.error(this.$t('m.CH_Load_Problem_Detail_Failed'))
        }
      } catch (error) {
        console.error('获取题目详情失败:', error)
        this.$message.error(this.$t('m.CH_Load_Problem_Detail_Failed'))
      } finally {
        this.loadingTagProblemDetail = false
      }
    },
    // 使用查看的题目
    useThisProblem() {
      if (!this.tagProblemDetail) return

      this.programmingForm.problemId = this.tagProblemDetail.problemId
      this.programmingInputMode = 'manual' // 切换回手动模式

      // 获取题目信息
      this.fetchProgrammingProblemInfo()

      // 关闭详情弹窗
      this.showTagProblemDetailDialog = false
      this.$message.success(this.$t('m.CH_Problem_Selected'))
    },
    // 通过标签选择题目（直接添加）
    async selectProblemByTag(problem) {
      this.programmingForm.problemId = problem.problemId

      // 获取完整题目信息
      const userInfo = this.$store.getters.userInfo
      const token = localStorage.getItem('token')

      if (!userInfo || !userInfo.username || !token) {
        this.$message.warning(this.$t('m.CH_Please_Login'))
        return
      }

      this.fetchingProblem = true
      try {
        const requestData = {
          pid: this.programmingForm.problemId,
          cid: '0',
          mode: 'normal',
          username: userInfo.username,
          token: token,
          password: ''
        }

        const res = await getClassroomProblem(requestData)

        if (res.code === 200 && res.data) {
          const problemData = res.data

          // 检查是否已经添加过该题目
          const exists = this.selectedQuestions.some(q => q.problemId === this.programmingForm.problemId)
          if (exists) {
            this.$message.warning(this.$t('m.CH_Programming_Already_Added'))
            return
          }

          // 直接添加编程题
          const tempQuestion = {
            id: `hoj_${this.programmingForm.problemId}`,
            problemId: this.programmingForm.problemId,
            title: problemData.problem.title,
            type: 'programming',
            difficulty: this.parseDifficultyLevel(problemData.problem.difficulty ?? problem.difficulty, 5),
            score: this.programmingForm.score,
            content: this.$t('m.CH_BingOJ_Id_Content', { id: this.programmingForm.problemId })
          }

          this.selectedQuestions.push(tempQuestion)
          this.$message.success(this.$t('m.CH_Added_Problem', { id: problem.problemId, title: problem.title }))
          // 清空表单以便继续添加，但保持对话框打开
          this.programmingForm.problemId = ''
          this.programmingProblemPreview = null
          this.programmingExamples = []
        } else {
          this.$message.error(res.message || this.$t('m.CH_Get_Problem_Info_Failed'))
        }
      } catch (error) {
        console.error('获取题目信息失败:', error)
        this.$message.error(this.$t('m.CH_Get_Problem_Info_Failed'))
      } finally {
        this.fetchingProblem = false
      }
    },
    // 获取难度标签类型
    getDifficultyTagType(difficulty) {
      const normalizedDifficulty = this.parseDifficultyLevel(difficulty, -1)
      const typeMap = {
        0: 'info',     // 入门
        1: 'success',  // 简单
        2: '',         // 中等（默认灰色）
        3: 'warning',  // 困难
        4: 'danger',   // 大师
        5: 'danger'    // 专家
      }
      return typeMap[normalizedDifficulty] || ''
    },
    // 获取难度名称（6个梯度）
    getDifficultyName(difficulty) {
      const normalizedDifficulty = this.parseDifficultyLevel(difficulty, -1)
      const nameMap = {
        0: this.$t('m.CH_Diff_Beginner'),
        1: this.$t('m.CH_Diff_Easy'),
        2: this.$t('m.CH_Diff_Medium'),
        3: this.$t('m.CH_Diff_Hard'),
        4: this.$t('m.CH_Diff_Master'),
        5: this.$t('m.CH_Diff_Expert')
      }
      if (nameMap[normalizedDifficulty]) {
        return nameMap[normalizedDifficulty]
      }

      const raw = (difficulty === undefined || difficulty === null) ? '' : String(difficulty).trim()
      const textMap = {
        beginner: this.$t('m.CH_Diff_Beginner'),
        easy: this.$t('m.CH_Diff_Easy'),
        medium: this.$t('m.CH_Diff_Medium'),
        hard: this.$t('m.CH_Diff_Hard'),
        master: this.$t('m.CH_Diff_Master'),
        expert: this.$t('m.CH_Diff_Expert'),
        入门: this.$t('m.CH_Diff_Beginner'),
        简单: this.$t('m.CH_Diff_Easy'),
        中等: this.$t('m.CH_Diff_Medium'),
        困难: this.$t('m.CH_Diff_Hard'),
        大师: this.$t('m.CH_Diff_Master'),
        专家: this.$t('m.CH_Diff_Expert')
      }
      return textMap[raw.toLowerCase()] || textMap[raw] || this.$t('m.Unknown')
    },
    parseDifficultyLevel(difficulty, fallback = 2) {
      if (difficulty === null || difficulty === undefined || difficulty === '') {
        return fallback
      }

      const parsed = Number.parseInt(String(difficulty).trim(), 10)
      if (Number.isInteger(parsed) && parsed >= 0 && parsed <= 5) {
        return parsed
      }

      const text = String(difficulty).trim().toLowerCase()
      const textMap = {
        beginner: 0,
        easy: 1,
        medium: 2,
        hard: 3,
        master: 4,
        expert: 5,
        入门: 0,
        简单: 1,
        中等: 2,
        困难: 3,
        大师: 4,
        专家: 5
      }

      if (Object.prototype.hasOwnProperty.call(textMap, text)) {
        return textMap[text]
      }

      return fallback
    },
    // 渲染 Markdown
    renderMarkdown(text) {
      const normalizedText = this.normalizeTextValue(text)
      if (!normalizedText) return ''
      try {
        return md.render(this.normalizeQuestionImagePath(normalizedText))
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return normalizedText
      }
    },
    // 获取判题模式文本
    getJudgeModeText(mode) {
      const modeMap = {
        'default': this.$t('m.CH_Judge_Mode_Default'),
        'spj': this.$t('m.CH_Judge_Mode_SPJ'),
        'interactive': this.$t('m.CH_Judge_Mode_Interactive'),
        'subtask': this.$t('m.CH_Judge_Mode_Subtask')
      }
      return modeMap[mode] || mode || this.$t('m.CH_Judge_Mode_Default')
    },
    // 全卷预览相关方法
    async showFullPreview() {
      if (this.selectedQuestions.length === 0) {
        this.$message.warning(this.$t('m.CH_Add_Question_First'))
        return
      }

      // 检查是否有编程题需要加载
      const hasProgrammingQuestions = this.selectedQuestions.some(q => q.type === 'programming' && q.problemId)
      if (hasProgrammingQuestions) {
        const loading = this.$loading({
          lock: true,
          text: this.$t('m.CH_Loading_Programming_Data'),
          spinner: 'el-icon-loading',
          background: 'rgba(0, 0, 0, 0.7)'
        })

        try {
          // 先加载所有编程题的信息，等待加载完成后再显示预览
          await this.loadProgrammingProblemsForPreview()
        } finally {
          loading.close()
        }
      }

      this.showFullPreviewDialog = true
    },
    handlePreviewOpened() {
      // 对话框打开后，确保 KaTeX 样式已加载
      // 不需要强制刷新，Vue 的响应式系统会自动处理
    },
    async loadProgrammingProblemsForPreview() {
      // 获取所有编程题的 problemId
      const programmingQuestions = this.selectedQuestions.filter(q => q.type === 'programming' && q.problemId)
      const problemIds = programmingQuestions.map(q => q.problemId)

      await this.loadProgrammingProblemByIds(problemIds)
    },
    // 批量加载编程题数据（用于编辑和预览）
    async loadProgrammingProblemByIds(problemIds) {
      if (!problemIds || problemIds.length === 0) return

      const userInfo = this.$store.getters.userInfo
      const token = localStorage.getItem('token')

      if (!userInfo || !token) {
        return
      }

      // 过滤掉已缓存的
      const uncachedIds = problemIds.filter(id => !this.programmingProblemsCache[id])

      for (const problemId of uncachedIds) {
        try {
          const res = await getClassroomProblem({
            pid: problemId,
            cid: '0',
            mode: 'normal',
            username: userInfo.username,
            token: token,
            password: ''
          })

          if (res.code === 200 && res.data) {
            this.$set(this.programmingProblemsCache, problemId, res.data)
          }
        } catch (error) {
          console.error('加载编程题信息失败:', problemId, error.message)
        }
      }
    },
    getProgrammingProblemInfo(problemId) {
      return this.programmingProblemsCache[problemId] || null
    },
    getProgrammingExamplesForProblem(problemId) {
      const problemInfo = this.programmingProblemsCache[problemId]
      if (!problemInfo || !problemInfo.problem || !problemInfo.problem.examples) {
        return []
      }

      // 解析样例
      const examples = []
      const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g
      let match

      while ((match = regex.exec(problemInfo.problem.examples)) !== null) {
        examples.push({
          input: match[1].trim(),
          output: match[2].trim()
        })
      }

      return examples
    },
    getTotalScore() {
      return this.selectedQuestions.reduce((sum, q) => sum + (q.score || 0), 0)
    },
    stripOptionPrefix(optionText, letterHint = '') {
      const normalizedText = String(optionText || '').trim()
      if (!normalizedText) return ''
      const escapedHint = letterHint ? letterHint.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') : '[A-Za-z]'
      const prefixRegex = new RegExp(`^\\s*(?:${escapedHint}|[A-Za-z])\\s*[\\.\\)、:：]\\s*`)
      return normalizedText.replace(prefixRegex, '').trim()
    },
    parseOptions(optionsInput) {
      if (!optionsInput) return []
      const options = this.parseMaybeSerializedJson(optionsInput)
      if (!Array.isArray(options)) return []
      return options.map((opt, idx) => {
        if (opt && typeof opt === 'object' && !Array.isArray(opt)) {
          const normalizedLetter = String(opt.letter || opt.label || String.fromCharCode(65 + idx))
            .trim()
            .replace(/[^A-Za-z0-9]/g, '')
            .toUpperCase()
          const letter = normalizedLetter || String.fromCharCode(65 + idx)
          const rawText = opt.text !== undefined
            ? opt.text
            : (opt.content !== undefined ? opt.content : '')
          return {
            letter,
            text: this.stripOptionPrefix(rawText, letter)
          }
        }
        const letter = String.fromCharCode(65 + idx)
        return {
          letter,
          text: this.stripOptionPrefix(opt, letter)
        }
      })
    },
    parseCompositeSubQuestions(optionsInput) {
      if (!optionsInput) return []
      const parsed = this.parseMaybeSerializedJson(optionsInput)
      if (!Array.isArray(parsed)) return []
      return parsed.map((subQuestion, index) => {
        const optionSource = subQuestion && subQuestion.options !== undefined
          ? subQuestion.options
          : (subQuestion && subQuestion.choiceOptions !== undefined ? subQuestion.choiceOptions : [])
        const parsedOptionSource = this.parseMaybeSerializedJson(optionSource)
        return {
          id: String((subQuestion && subQuestion.id) || `sq_${index + 1}`),
          content: (subQuestion && subQuestion.content) || '',
          options: Array.isArray(parsedOptionSource) ? parsedOptionSource : [],
          score: Number((subQuestion && (subQuestion.score || subQuestion.subScore)) || 0)
        }
      })
    },
    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },

    goBack() {
      // 根据当前路由判断返回路径
      if (this.isAdminRoute) {
        // 管理员路由：返回管理员班级管理页面
        if (this.isEditMode && this.editId) {
          // 编辑模式，返回作业详情（管理员路由）
          this.$router.push({
            path: '/admin/classroom',
            query: {
              classroomId: this.classroomId,
              homeworkId: this.editId,
              activeTab: 'homework'
            }
          })
        } else {
          // 创建模式，返回班级管理页面
          this.$router.push({
            path: '/admin/classroom',
            query: {
              classroomId: this.classroomId,
              activeTab: 'homework'
            }
          })
        }
      } else {
        // 教师路由：返回教师端页面
        if (this.isEditMode && this.editId) {
          // 编辑模式，返回到作业详情页
          this.$router.push({
            name: 'TeacherHomeworkDetail',
            params: {
              classroomId: this.classroomId,
              homeworkId: this.editId
            }
          })
        } else {
          // 创建模式，返回到作业列表页
          this.$router.push({
            name: 'TeacherHomework',
            params: {
              classroomId: this.classroomId
            },
            query: {
              tab: 'homework'
            }
          })
        }
      }
    },

    // ==================== 导入试卷相关方法 ====================
    // 打开导入试卷对话框
    openImportPaperDialog() {
      this.showImportPaperDialog = true
      this.paperSearchKeyword = ''
      this.selectedPaperId = null
      this.paperCurrentPage = 1
      this.loadExamPaperList()
    },

    // 加载试卷列表
    async loadExamPaperList() {
      this.loadingPapers = true
      try {
        const searchKeyword = (this.paperSearchKeyword || '').trim()
        const paperIdCandidate = Number.parseInt(searchKeyword, 10)
        const params = {
          page: this.paperCurrentPage,
          limit: this.paperPageSize
        }
        if (searchKeyword) {
          params.keyword = searchKeyword
          if (/^\d+$/.test(searchKeyword) && Number.isInteger(paperIdCandidate) && paperIdCandidate > 0) {
            params.paperId = paperIdCandidate
          }
        }

        const res = await this.$store.dispatch('classroom/getExamPaperList', params)
        if (res.code === 200) {
          this.examPapers = res.data.papers || res.data || []
          this.paperTotal = res.data.total || 0
        }
      } catch (error) {
        console.error('加载试卷列表失败:', error)
        this.$message.error(this.$t('m.CH_Load_Papers_Failed'))
      } finally {
        this.loadingPapers = false
      }
    },

    // 试卷列表分页
    handlePaperPageChange(page) {
      this.paperCurrentPage = page
      this.loadExamPaperList()
    },

    // 确认导入试卷
    async confirmImportPaper() {
      if (!this.selectedPaperId) {
        this.$message.warning(this.$t('m.CH_Select_Paper_First'))
        return
      }

      this.importingPaper = true
      try {
        // 确保参数是数字类型，并验证有效性
        const paperId = parseInt(this.selectedPaperId)
        const classroomId = parseInt(this.classroomId)

        // 验证转换结果
        if (isNaN(paperId) || paperId <= 0) {
          this.$message.error(this.$t('m.CH_Select_Valid_Paper'))
          return
        }
        if (isNaN(classroomId) || classroomId <= 0) {
          this.$message.error(this.$t('m.CH_Invalid_Classroom_Id'))
          return
        }

        const requestData = {
          paperId: paperId,
          classroomId: classroomId
        }

        const res = await this.$store.dispatch('classroom/importExamPaper', requestData)

        // 检查响应数据结构
        if (!res) {
          this.$message.error(this.$t('m.CH_Import_No_Data'))
          return
        }

        if (res.code === 200) {
          // 导入成功，将题目添加到已选题目列表
          // 注意：res.data 可能是题目数组，也可能包含 {questions: [...]}
          let questions = []
          if (Array.isArray(res.data)) {
            questions = res.data
          } else if (res.data && Array.isArray(res.data.questions)) {
            questions = res.data.questions
          }

          if (questions.length > 0) {
            questions.forEach(q => {
              // 检查是否已存在
              // 客观题：使用 id 或 questionId 判断
              // 编程题：使用 problemId 判断
              const exists = this.selectedQuestions.some(sq => {
                // 客观题：比较 id 或 questionId
                const incomingQuestionId = q.questionId !== undefined && q.questionId !== null ? q.questionId : q.id
                if (incomingQuestionId !== undefined && incomingQuestionId !== null && incomingQuestionId !== '') {
                  const incomingKey = String(incomingQuestionId)
                  return String(sq.id) === incomingKey || String(sq.questionId) === incomingKey
                }
                // 编程题：比较 problemId
                if (q.problemId) {
                  return sq.problemId === q.problemId
                }
                return false
              })

              if (!exists) {
                // 为每个题目设置唯一标识
                let questionId
                if (q.questionId || q.id) {
                  // 客观题：统一转为数字，避免后端 uint64 解析失败
                  const numericQuestionId = Number(q.questionId || q.id)
                  if (!Number.isInteger(numericQuestionId) || numericQuestionId <= 0) {
                    console.warn('跳过无效题目ID:', q)
                    return
                  }
                  questionId = numericQuestionId
                } else {
                  // 编程题：使用 problemId（已经是字符串）
                  questionId = q.problemId
                }

                this.selectedQuestions.push({
                  id: questionId,
                  questionId: (q.questionId || q.id) ? questionId : undefined,
                  problemId: q.problemId,
                  type: q.questionType || q.type,
                  title: q.title,
                  difficulty: q.difficulty || 2,
                  score: q.score || 10,
                  questionOrder: q.questionOrder
                })
              }
            })

            this.$message.success(this.$t('m.CH_Import_Success', { count: questions.length }))
            this.showImportPaperDialog = false
          } else {
            this.$message.warning(this.$t('m.CH_Paper_No_Questions'))
          }
        } else {
          this.$message.error(res.message || this.$t('m.CH_Import_Failed'))
        }
      } catch (error) {
        console.error('导入试卷失败:', error)
        // 检查是否是502错误（网关错误）
        if (error.response && error.response.status === 502) {
          this.$message.error(this.$t('m.CH_Import_Busy'))
        } else if (error.response && error.response.status === 404) {
          this.$message.error(this.$t('m.CH_Import_Not_Found'))
        } else if (error.response && error.response.status === 403) {
          this.$message.error(this.$t('m.CH_Import_No_Permission'))
        } else {
          this.$message.error(this.$t('m.CH_Import_Failed_Msg', { msg: error.message || this.$t('m.CH_Import_Bad_Params') }))
        }
      } finally {
        this.importingPaper = false
      }
    },

    // 格式化试卷时间
    formatPaperTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    }
  }
}
</script>

<style scoped>
/* 整体布局 - 清爽浅色主题 */
.create-homework-wrapper {
  min-height: 100vh;
}

/* 顶部固定导航栏 - 柔和蓝色 */
.create-homework-header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #5b9bd5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.create-homework-header .header-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 14px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left i {
  font-size: 24px;
  color: #fff;
}

.header-left h2 {
  margin: 0;
  color: #fff;
  font-size: 20px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* 内容区域 */
.create-homework-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px 24px 24px;
}

.editor-page-switch {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.homework-form {
  max-width: 100%;
}

/* 卡片通用样式 */
.form-section {
  margin-bottom: 20px;
  border-radius: 12px;
  background: #fff;
  border: 1px solid #e0e6ed;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.form-section:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0;
  background: #f8f9fa;
  margin: -20px -20px 20px -20px;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e6ed;
  border-radius: 12px 12px 0 0;
}

.header-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: #5b9bd5;
  border-radius: 8px;
  color: #fff;
  font-size: 18px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.header-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: 12px;
}

.header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 10px;
}

.add-programming-btn {
  margin-left: auto;
}

/* 基本信息卡片 */
.basic-info-card {
  background: #fff;
}

.basic-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
  padding: 0 20px;
}

.basic-info-grid .full-width {
  grid-column: 1 / -1;
}

/* 显示设置 */
/* 模式选择区域 - 方形样式 */
.mode-selection {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 14px;
  background-color: #fff;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
}

.mode-label {
  font-weight: 600;
  color: #2c3e50;
  font-size: 14px;
  min-width: 70px;
}

/* 将 radio 改为方形按钮样式 */
.mode-selection ::v-deep .el-radio-group {
  display: flex;
  gap: 10px;
}

.mode-selection ::v-deep .el-radio {
  margin-right: 0;
}

.mode-selection ::v-deep .el-radio__input {
  display: none; /* 隐藏原圆形 radio */
}

.mode-selection ::v-deep .el-radio__label {
  padding: 8px 16px;
  border: 2px solid #dcdfe6;
  border-radius: 4px; /* 方形圆角 */
  background-color: #fff;
  color: #606266;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
  user-select: none;
}

.mode-selection ::v-deep .el-radio__label:hover {
  border-color: #5b9bd5;
  color: #5b9bd5;
}

/* 选中状态 */
.mode-selection ::v-deep .el-radio.is-checked .el-radio__label {
  background-color: #5b9bd5;
  border-color: #5b9bd5;
  color: #fff;
  font-weight: 500;
}

.display-settings {
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  margin: 0 20px;
}

.setting-item {
  margin-bottom: 12px;
  padding: 14px;
  background-color: #fff;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
  transition: all 0.2s ease;
}

.setting-item:hover {
  border-color: #5b9bd5;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  transform: translateY(-1px);
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-label {
  font-weight: 500;
  color: #2c3e50;
}

.setting-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #7f8c8d;
  line-height: 1.6;
  padding-left: 4px;
}

/* 题目管理卡片 - 单列布局 */
.questions-management-card {
  background: #fff;
}

.questions-container-layout {
  display: block;
  padding: 0 20px 20px;
}

/* 顶部工具栏 */
.questions-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 15px 20px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toolbar-right .el-button-group {
  display: flex;
  gap: 8px;
}

.toolbar-right .el-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* 左侧添加题目面板 */
.add-questions-panel {
  display: flex;
  flex-direction: column;
}

.action-card {
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 20px;
}

.action-item h4 {
  margin: 15px 0 10px;
  color: #303133;
  font-size: 16px;
}

.action-item p {
  margin: 0 0 15px;
  color: #909399;
  font-size: 13px;
}

.action-item .el-button {
  width: 100%;
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #fff;
  border-bottom: 1px solid #e0e6ed;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: #2c3e50;
}

.filter-section {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  background: #fff;
  border-bottom: 1px solid #e0e6ed;
}

.search-input {
  flex: 1;
  max-width: none;
}

.filter-select {
  width: 140px;
}

.filter-stats {
  display: flex;
  gap: 8px;
}

.question-list {
  position: relative;
  padding: 16px;
  flex: 1;
  overflow: hidden;
}

.question-table {
  border-radius: 6px;
  overflow: hidden;
}

.question-table >>> th {
  background: #f8f9fa;
  font-weight: 600;
  color: #2c3e50;
  border-bottom: 1px solid #e0e6ed;
}

.question-table >>> tr:hover {
  background-color: #f8f9fa !important;
}

.score-text {
  font-weight: 600;
  color: #5b9bd5;
}

.pagination-wrapper {
  margin-top: 12px;
  display: flex;
  justify-content: center;
  padding: 12px 0 0;
}

.question-selector-dialog {
  display: flex;
  flex-direction: column;
}

.selector-filter-section {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 12px;
  padding: 12px;
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

.selector-question-desc {
  margin: 8px 0;
}

.selector-question-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.selector-option-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #f8f9fa;
}

.selector-question-meta {
  margin: 8px 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.selector-analysis {
  margin-top: 8px;
}

/* 右侧已选题目面板 */
.selected-questions-panel,
.selected-questions-panel-full {
  display: flex;
  flex-direction: column;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
  overflow: hidden;
}

.selected-questions-list {
  flex: 1;
  max-height: 520px;
  overflow-y: auto;
  padding: 12px;
}

.empty-selected {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #95a5a6;
}

.empty-selected i {
  font-size: 48px;
  margin-bottom: 12px;
  color: #bdc3c7;
}

.empty-selected p {
  margin: 4px 0;
  font-size: 14px;
}

.empty-selected .hint {
  font-size: 12px;
  color: #bdc3c7;
}

.questions-container {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.question-item-card {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  padding: 12px;
  background: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.question-item-card:hover {
  border-color: #5b9bd5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.question-item-main {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.question-index {
  flex-shrink: 0;
}

.index-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #5b9bd5;
  color: #fff;
  border-radius: 6px;
  font-weight: 700;
  font-size: 14px;
}

.selected-question-content {
  flex: 1;
  min-width: 0;
}

.question-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.question-title-text {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: #2c3e50;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selected-question-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.score-editor {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #e0e6ed;
}

.score-label {
  font-size: 12px;
  color: #2c3e50;
  font-weight: 500;
}

.score-unit {
  font-size: 12px;
  color: #7f8c8d;
}

.question-actions {
  flex-shrink: 0;
}

.selected-question-expanded {
  border-top: 1px solid #eef2f6;
  padding-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.selected-question-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.selected-question-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-left: 0;
}

.selected-question-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #f8f9fa;
}

.selected-question-answer {
  display: flex;
  align-items: center;
}

.selected-programming-detail {
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 10px 12px;
  background: #fcfcfd;
}

/* 列表过渡动画 */
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter,
.list-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

/* 对话框样式优化 */
.question-detail {
  padding: 10px;
}

.detail-content {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  line-height: 1.8;
  max-height: 400px;
  overflow-y: auto;
}

.composite-detail-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.composite-detail-item {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  padding: 12px;
  background: #fff;
}

.composite-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
  color: #303133;
}

.composite-detail-options {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.composite-detail-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #fafafa;
}

.composite-detail-answer {
  margin-top: 10px;
}

.example-box {
  margin-bottom: 15px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  border: 1px solid #e0e6ed;
}

.example-title {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 8px;
}

.example-box pre {
  margin: 5px 0;
  padding: 8px;
  background: white;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 编程题预览样式 */
.problem-preview {
  margin: 20px 0;
}

.problem-preview h3 {
  margin-top: 0;
  color: #2c3e50;
  font-size: 18px;
}

.problem-meta {
  margin: 15px 0;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.problem-content {
  margin-top: 20px;
}

.content-section {
  margin-bottom: 20px;
}

.content-section h4 {
  color: #5b9bd5;
  font-size: 16px;
  margin-bottom: 10px;
  border-left: 3px solid #5b9bd5;
  padding-left: 10px;
}

.content-section pre {
  background-color: #f8f9fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 10px 0;
  border: 1px solid #e0e6ed;
}

.example-item {
  margin-bottom: 15px;
}

.example-item pre {
  margin: 8px 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 全卷预览样式 */
.full-preview-container {
  max-height: 70vh;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 10px;
}

/* 确保对话框内容可以正常滚动 */
.full-preview-container >>> .el-card {
  margin-bottom: 20px;
  overflow: visible;
}

.full-preview-container >>> .preview-markdown pre,
.question-detail >>> .preview-markdown pre,
.problem-detail-content >>> .preview-markdown pre,
.problem-preview >>> .preview-markdown pre,
.selected-question-expanded >>> .preview-markdown pre {
  margin-left: 0 !important;
  text-indent: 0 !important;
  padding: 10px 12px !important;
  overflow-x: auto !important;
}

.full-preview-container >>> .preview-markdown pre code,
.full-preview-container >>> .preview-markdown code.hljs,
.question-detail >>> .preview-markdown pre code,
.question-detail >>> .preview-markdown code.hljs,
.problem-detail-content >>> .preview-markdown pre code,
.problem-detail-content >>> .preview-markdown code.hljs,
.problem-preview >>> .preview-markdown pre code,
.problem-preview >>> .preview-markdown code.hljs,
.selected-question-expanded >>> .preview-markdown pre code,
.selected-question-expanded >>> .preview-markdown code.hljs {
  margin-left: 0 !important;
  text-indent: 0 !important;
  display: block;
  white-space: pre !important;
}

.full-preview-container >>> .preview-markdown p,
.question-detail >>> .preview-markdown p,
.problem-detail-content >>> .preview-markdown p,
.problem-preview >>> .preview-markdown p,
.selected-question-expanded >>> .preview-markdown p {
  text-indent: 0 !important;
}

.selected-question-expanded >>> .preview-markdown pre {
  padding-left: 0 !important;
}

.selected-question-expanded >>> .preview-markdown pre code,
.selected-question-expanded >>> .preview-markdown code.hljs {
  padding-left: 0 !important;
}

.full-preview-container >>> .question-item {
  overflow: visible;
}

.preview-info {
  margin-bottom: 20px;
  background-color: #f8f9fa;
  border: 1px solid #e0e6ed;
  border-radius: 8px;
  padding: 16px;
}

.preview-info h3 {
  margin: 0 0 15px 0;
  color: #5b9bd5;
  font-size: 20px;
}

.preview-info p {
  margin: 8px 0;
  color: #2c3e50;
}

.questions-preview {
  margin-top: 20px;
}

.question-item {
  padding: 20px;
  margin-bottom: 20px;
  background-color: #fff;
  border: 1px solid #e0e6ed;
  border-radius: 8px;
}

.question-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e0e6ed;
}

.question-number {
  font-size: 18px;
  font-weight: bold;
  color: #5b9bd5;
}

.question-type {
  color: #7f8c8d;
  font-size: 14px;
}

.question-score {
  margin-left: auto;
  color: #ff9800;
  font-weight: bold;
}

.question-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 15px;
  color: #2c3e50;
}

.question-content {
  margin: 15px 0;
  color: #606266;
  line-height: 1.8;
  font-size: 14px;
}

.question-options {
  margin-top: 15px;
  padding-left: 20px;
}

.option-item {
  margin-bottom: 10px;
}

.composite-preview-list {
  margin-top: 12px;
  padding-left: 0;
}

.composite-preview-item {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 12px;
  background: #fff;
}

.composite-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
  color: #303133;
}

.option-preview {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 4px;
  cursor: default;
  border: 1px solid #e0e6ed;
}

.option-preview:hover {
  background-color: #e9ecef;
}

.option-letter {
  display: inline-block;
  min-width: 30px;
  font-weight: bold;
  color: #5b9bd5;
  font-size: 15px;
}

.option-text {
  flex: 1;
  color: #606266;
  line-height: 1.6;
}

.subjective-preview,
.programming-preview {
  margin-top: 15px;
  padding: 15px;
  background-color: #fff9e6;
  border-left: 3px solid #ffa726;
  border-radius: 4px;
  border: 1px solid #ffecb3;
}

.programming-detail {
  margin-top: 15px;
  padding: 15px;
  background-color: #fff;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
}

.programming-meta {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.programming-content h4 {
  color: #5b9bd5;
  font-size: 16px;
  margin-bottom: 10px;
  border-left: 3px solid #5b9bd5;
  padding-left: 10px;
}

.programming-content {
  /* 确保长内容不会破坏布局 */
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.programming-content >>> * {
  max-width: 100%;
  overflow-x: auto;
}

/* 特别处理 pre/code 块，防止内容溢出 */
.programming-content >>> pre,
.programming-content >>> code {
  white-space: pre-wrap !important;
  word-break: break-word !important;
  max-width: 100% !important;
  overflow-x: auto !important;
}

/* 确保 Markdown 渲染的数学公式等不会撑破容器 */
.programming-content >>> .katex,
.programming-content >>> .katex-display {
  max-width: 100% !important;
  overflow-x: auto !important;
}

/* 确保 table 不会撑破容器 */
.programming-content >>> table {
  max-width: 100% !important;
  overflow-x: auto !important;
  display: block !important;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .create-homework-header .header-content {
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px;
  }

  .create-homework-content {
    padding: 16px;
  }

  .basic-info-grid {
    grid-template-columns: 1fr;
    padding: 0 12px;
  }

  .questions-container-layout {
    grid-template-columns: 1fr;
  }

  .questions-toolbar {
    flex-direction: column;
    gap: 15px;
    padding: 15px;
  }

  .toolbar-left {
    width: 100%;
  }

  .toolbar-right .el-button-group {
    display: flex;
    flex-wrap: wrap;
    width: 100%;
  }

  .toolbar-right .el-button {
    flex: 1;
    min-width: 120px;
  }

  .selector-summary {
    flex-wrap: wrap;
  }

  .selector-question-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .selector-tools {
    justify-content: flex-start;
  }

  .filter-section {
    flex-direction: column;
    gap: 12px;
  }

  .search-input {
    max-width: 100%;
  }

  .question-item-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .question-actions {
    width: 100%;
    display: flex;
    justify-content: flex-end;
  }

  .question-list {
    padding: 0 12px 16px;
  }

  .exam-config-section {
    padding: 0 12px;
  }
}

/* 考试模式配置区域 - 确保与基本信息对齐 */
.exam-config-section {
  padding: 0 20px;
}

/* 调整考试模式下 el-divider 的左对齐，与表单标签对齐 */
.exam-config-section ::v-deep .el-divider {
  margin: 20px 0 24px 0;
}

.exam-config-section ::v-deep .el-divider__text {
  padding-left: 0;
}

.exam-config-section ::v-deep .el-divider__text.is-left {
  left: 0;
}

/* 考试模式设置网格布局 */
.exam-settings-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  padding: 0;
}

.exam-setting-item {
  padding: 14px;
  background-color: #fff;
  border-radius: 8px;
  border: 1px solid #e0e6ed;
  transition: all 0.2s ease;
}

.exam-setting-item:hover {
  border-color: #5b9bd5;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

/* 单位标签样式 */
.unit-label {
  margin: 0 8px;
  color: #606266;
  font-size: 14px;
}

/* 编辑模式数据加载遮罩 */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.95);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 9999;
}

.loading-overlay i {
  font-size: 48px;
  color: #5b9bd5;
  margin-bottom: 20px;
}

.loading-overlay p {
  font-size: 16px;
  color: #606266;
  margin: 0;
}

/* ==================== 导入试卷对话框样式 ==================== */
.empty-papers {
  text-align: center;
  padding: 60px 20px;
  color: #909399;
}

.empty-papers i {
  font-size: 48px;
  margin-bottom: 16px;
  display: block;
}

.empty-papers p {
  margin: 8px 0;
  font-size: 14px;
}

.empty-papers .hint {
  font-size: 12px;
  color: #c0c4cc;
}

.paper-list {
  max-height: 500px;
  overflow-y: auto;
}

.paper-radio-group {
  display: block;
  width: 100%;
}

.paper-item {
  margin-bottom: 16px;
  border: 2px solid #e0e6ed;
  border-radius: 8px;
  padding: 16px;
  background: #fafbfc;
  transition: all 0.3s ease;
  cursor: pointer;
}

.paper-item:hover {
  border-color: #5b9bd5;
  background: #f0f7ff;
  box-shadow: 0 2px 8px rgba(91, 155, 213, 0.2);
}

.paper-item.is-selected {
  border-color: #5b9bd5;
  background: #e6f3ff;
  box-shadow: 0 2px 12px rgba(91, 155, 213, 0.3);
}

.paper-radio {
  display: block;
  margin: 0;
  width: 100%;
}

.paper-radio >>> .el-radio__label {
  padding-left: 8px;
  width: 100%;
}

.paper-radio >>> .el-radio__input {
  vertical-align: top;
  margin-top: 4px;
}

.paper-content {
  display: inline-block;
  width: calc(100% - 28px);
  vertical-align: top;
}

.paper-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.paper-title {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  margin-right: 12px;
}

.paper-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.paper-description {
  font-size: 13px;
  color: #606266;
  margin: 8px 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.paper-meta {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  font-size: 12px;
  color: #909399;
}

.paper-meta span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.paper-meta i {
  font-size: 14px;
}
</style>

<!-- 引入 KaTeX 样式 -->
<style>
@import '~katex/dist/katex.min.css';
</style>
