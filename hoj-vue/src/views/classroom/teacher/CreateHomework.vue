<template>
  <div class="create-homework-wrapper">
    <!-- 编辑模式数据加载遮罩 -->
    <div v-if="loadingData" class="loading-overlay">
      <el-icon class="is-loading"><i class="el-icon-loading"></i></el-icon>
      <p>正在加载作业数据...</p>
    </div>

    <!-- 顶部固定导航栏 -->
    <div class="create-homework-header">
      <div class="header-content">
        <div class="header-left">
          <i class="el-icon-edit-outline"></i>
          <h2>{{ isEditMode ? '编辑作业' : '创建作业' }}</h2>
        </div>
        <div class="header-actions">
          <el-button @click="goBack" size="medium">
            <i class="el-icon-back"></i>
            返回
          </el-button>
          <el-button type="primary" @click="saveHomework" :loading="submitting" size="medium">
            <i class="el-icon-check"></i>
            {{ isEditMode ? '保存' : '创建' }}
          </el-button>
        </div>
      </div>
    </div>

    <div class="create-homework-content">
      <el-form :model="form" :rules="rules" ref="homeworkForm" label-width="100px" class="homework-form">
        <!-- 第一板块：基本信息 -->
        <el-card class="form-section basic-info-card" shadow="hover">
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
              <div class="mode-label">作业模式</div>
              <el-radio-group v-model="form.isExamMode">
                <el-radio :label="0">普通作业模式</el-radio>
                <el-radio :label="1">考试模式</el-radio>
              </el-radio-group>
            </div>

            <div class="setting-item">
              <el-checkbox v-model="form.showHomework">
                <span class="setting-label">{{ $t('m.Show_Homework_After_Complete') }}</span>
              </el-checkbox>
              <div class="setting-tip">
                {{ form.isExamMode === 1 ? '考试模式下，需等到考试结束时间后才能查看' : $t('m.Show_Homework_Tip') }}
              </div>
            </div>
            <div class="setting-item">
              <el-checkbox v-model="form.showScore">
                <span class="setting-label">{{ $t('m.Show_Score_After_Complete') }}</span>
              </el-checkbox>
              <div class="setting-tip">
                {{ form.isExamMode === 1 ? '考试模式下，需等到考试结束时间后才能查看' : $t('m.Show_Score_Tip') }}
              </div>
            </div>
            <div class="setting-item">
              <el-checkbox v-model="form.showAnswer">
                <span class="setting-label">允许学生提交后查看答案</span>
              </el-checkbox>
              <div class="setting-tip">
                {{ form.isExamMode === 1 ? '考试模式下，需等到考试结束时间后才能查看' : '学生提交作业后，可以查看每道题的正确答案（主观题显示参考答案，编程题显示题库链接）' }}
              </div>
            </div>
          </div>

          <!-- 考试模式配置 -->
          <div v-if="form.isExamMode === 1" class="exam-config-section">
            <el-divider content-position="left">时间设置</el-divider>

            <el-form-item label="考试时长" prop="examDuration" required>
              <el-input-number
                v-model="form.examDuration"
                :min="1"
                :max="600"
                :step="5"
                controls-position="right"
              />
              <span class="unit-label">分钟</span>
              <span class="setting-tip">学生从开始答题算起的时长限制</span>
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
                  <strong>⚠️ 时间区间不匹配：</strong>
                  作业开放时间段为 {{ timeRangeMinutes }} 分钟，但考试时长为 {{ form.examDuration }} 分钟。
                  <br>
                  <span style="color: #E6A23C;">学生可以在作业开放时间段内的任意时刻开始考试，</span>
                  从开始答题起计时 {{ form.examDuration }} 分钟后或者考试截止时间自动交卷。
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
                  <strong>✓ 时间配置一致：</strong>
                  作业开放时间段和考试时长一致（{{ form.examDuration }} 分钟）。
                  <br>
                  学生必须在作业开始时间后立即开始考试，考试时长固定为 {{ form.examDuration }} 分钟。
                </div>
              </template>
            </el-alert>

            <el-form-item label="允许交卷时间" prop="allowSubmitAfterMinutes">
              <el-input-number
                v-model="form.allowSubmitAfterMinutes"
                :min="0"
                :max="600"
                :step="5"
                controls-position="right"
              />
              <span class="unit-label">分钟</span>
              <span class="setting-tip">开考后至少需要答题多少分钟才能交卷（0表示可以立即交卷）</span>
            </el-form-item>

            <el-divider content-position="left">防作弊设置</el-divider>

            <div class="exam-settings-grid">
              <div class="exam-setting-item">
                <el-checkbox v-model="form.disableCopyPaste">
                  <span class="setting-label">禁止复制粘贴</span>
                </el-checkbox>
                <div class="setting-tip">禁止学生在考试期间复制或粘贴内容</div>
              </div>

              <div class="exam-setting-item">
                <el-checkbox v-model="form.requireFullscreen">
                  <span class="setting-label">要求全屏模式</span>
                </el-checkbox>
                <div class="setting-tip">学生必须在全屏模式下答题，退出全屏会记录违规</div>
              </div>

              <div class="exam-setting-item">
                <el-checkbox v-model="form.disallowTabSwitch">
                  <span class="setting-label">禁止切换标签页</span>
                </el-checkbox>
                <div class="setting-tip">检测学生切换浏览器标签页的行为并记录</div>
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
                  <strong>考试模式说明：</strong>
                  <ul style="margin: 10px 0 0 20px; padding: 0;">
                    <li>每位学生的题目顺序会随机打乱，防止作弊</li>
                    <li>超时未交卷将自动强制收卷</li>
                    <li>教师可以实时监控学生答题状态</li>
                    <li>所有违规行为（退出全屏、切换标签页等）都会被记录</li>
                  </ul>
                </div>
              </template>
            </el-alert>
          </div>
        </el-card>

      <!-- 第二板块：题目管理（题库 + 已选题目） -->
      <el-card class="form-section questions-management-card" shadow="hover">
        <div slot="header" class="card-header">
          <span class="header-icon">
            <i class="el-icon-collection"></i>
          </span>
          <span class="header-title">题目管理</span>
          <div class="header-stats">
            <el-tag size="small" type="info">已选 {{ selectedQuestions.length }} 题</el-tag>
            <el-tag size="small" type="success" style="margin-left: 8px;">总分: {{ getTotalScore() }} 分</el-tag>
          </div>
          <div class="header-actions">
            <el-button
              type="success"
              size="small"
              icon="el-icon-plus"
              @click="showAddProgrammingDialog = true"
            >
              添加编程题
            </el-button>
            <el-button
              type="primary"
              size="small"
              icon="el-icon-view"
              @click="showFullPreview"
              :disabled="selectedQuestions.length === 0"
              plain
            >
              全卷预览
            </el-button>
          </div>
        </div>

        <div class="questions-container-layout">
          <!-- 左侧：题库 -->
          <div class="question-bank-panel">
            <div class="panel-header">
              <span class="panel-title">题库</span>
              <el-tag size="mini" type="info">共 {{ total }} 题</el-tag>
            </div>

            <div class="filter-section">
              <el-input
                v-model="searchKeyword"
                :placeholder="$t('m.Search_Questions')"
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
                v-model="filterType"
                :placeholder="$t('m.Question_Type')"
                clearable
                @change="loadQuestionBank"
                size="small"
                class="filter-select"
              >
                <el-option :label="$t('m.All')" value="" />
                <el-option :label="$t('m.Single_Choice')" value="single_choice" />
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice" />
                <el-option :label="$t('m.Judge')" value="judge" />
                <el-option :label="$t('m.Subjective')" value="subjective" />
                <el-option :label="$t('m.Programming')" value="programming" />
              </el-select>
            </div>

            <div class="question-list" v-loading="questionsLoading">
              <el-table
                :data="questionBank"
                stripe
                class="question-table"
                height="500"
              >
                <el-table-column prop="title" :label="$t('m.Question_Title')" min-width="180" show-overflow-tooltip />
                <el-table-column prop="type" :label="$t('m.Question_Type')" width="100" align="center">
                  <template slot-scope="{ row }">
                    <el-tag :type="getQuestionTypeColor(row.type)" size="mini">
                      {{ getQuestionTypeText(row.type) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="difficulty" :label="$t('m.Difficulty')" width="100" align="center">
                  <template slot-scope="{ row }">
                    <el-rate :value="getDifficultyStars(row.difficulty)" disabled show-score text-color="#ff9900" />
                  </template>
                </el-table-column>
                <el-table-column :label="$t('m.Operation')" width="140" align="center" fixed="right">
                  <template slot-scope="{ row }">
                    <el-button
                      :type="isQuestionSelected(row) ? 'info' : 'primary'"
                      size="mini"
                      icon="el-icon-plus"
                      @click="addQuestion(row)"
                      :disabled="isQuestionSelected(row)"
                      plain
                    >
                      {{ isQuestionSelected(row) ? '已添加' : '添加' }}
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>

              <div class="pagination-wrapper">
                <el-pagination
                  v-if="total > 0"
                  :current-page="currentPage"
                  :page-size="pageSize"
                  :total="total"
                  layout="prev, pager, next"
                  @current-change="handlePageChange"
                  small
                />
              </div>
            </div>
          </div>

          <!-- 右侧：已选题目 -->
          <div class="selected-questions-panel">
            <div class="panel-header">
              <span class="panel-title">已选题目</span>
              <el-tag size="mini" type="success">{{ selectedQuestions.length }} 题</el-tag>
            </div>

            <div class="selected-questions-list" v-if="selectedQuestions.length > 0">
              <transition-group name="list" tag="div" class="questions-container">
                <div
                  v-for="(question, index) in selectedQuestions"
                  :key="question.id || question.problemId"
                  class="question-item-card"
                >
                  <div class="question-index">
                    <span class="index-number">{{ index + 1 }}</span>
                  </div>

                  <div class="question-content">
                    <div class="question-title-row">
                      <el-tag :type="getQuestionTypeColor(question.type)" size="mini">
                        {{ getQuestionTypeText(question.type) }}
                      </el-tag>
                      <span class="question-title-text" :title="question.title">{{ question.title }}</span>
                    </div>

                    <div class="question-meta">
                      <div class="score-editor">
                        <span class="score-label">分值:</span>
                        <el-input-number
                          v-model="question.score"
                          :min="1"
                          :max="100"
                          size="mini"
                          controls-position="right"
                        />
                        <span class="score-unit">分</span>
                      </div>
                    </div>
                  </div>

                  <div class="question-actions">
                    <el-button-group>
                      <el-tooltip content="上移" placement="top">
                        <el-button
                          size="mini"
                          icon="el-icon-top"
                          :disabled="index === 0"
                          @click="moveUp(index)"
                        />
                      </el-tooltip>
                      <el-tooltip content="下移" placement="top">
                        <el-button
                          size="mini"
                          icon="el-icon-bottom"
                          :disabled="index === selectedQuestions.length - 1"
                          @click="moveDown(index)"
                        />
                      </el-tooltip>
                      <el-tooltip content="移除" placement="top">
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
              </transition-group>
            </div>

            <div v-else class="empty-selected">
              <i class="el-icon-document"></i>
              <p>暂未选择题目</p>
              <p class="hint">从左侧题库中添加题目</p>
            </div>
          </div>
        </div>
      </el-card>

      </el-form>

    </div>

    <!-- 添加 BingOJ 编程题对话框 -->
    <el-dialog title="添加 BingOJ 编程题" :visible.sync="showAddProgrammingDialog" width="900px">
      <el-form :model="programmingForm" label-width="120px">
        <el-form-item label="BingOJ 题目 ID" required>
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
          <div style="margin-top: 8px; color: #909399; font-size: 12px;">
            <i class="el-icon-info"></i>
            输入 BingOJ 题库中的题目 ID，点击"获取题目信息"预览题目内容
          </div>
        </el-form-item>

        <!-- 题目预览区域 -->
        <div v-if="programmingProblemPreview" class="problem-preview">
          <el-divider content-position="left">题目预览</el-divider>
          <el-card>
            <h3>{{ programmingProblemPreview.problem.title }}</h3>
            <div class="problem-meta">
              <el-tag size="small">题目ID: {{ programmingProblemPreview.problem.problemId }}</el-tag>
              <el-tag size="small" type="info">时间限制: {{ programmingProblemPreview.problem.timeLimit }}ms</el-tag>
              <el-tag size="small" type="warning">内存限制: {{ programmingProblemPreview.problem.memoryLimit }}MB</el-tag>
              <el-tag size="small" type="success">判题模式: {{ getJudgeModeText(programmingProblemPreview.problem.judgeMode) }}</el-tag>
            </div>
            <div class="problem-content">
              <div class="content-section">
                <h4>题目描述</h4>
                <div v-html="renderMarkdown(programmingProblemPreview.problem.description)"></div>
              </div>
              <div class="content-section" v-if="programmingProblemPreview.problem.input">
                <h4>输入格式</h4>
                <div v-html="renderMarkdown(programmingProblemPreview.problem.input)"></div>
              </div>
              <div class="content-section" v-if="programmingProblemPreview.problem.output">
                <h4>输出格式</h4>
                <div v-html="renderMarkdown(programmingProblemPreview.problem.output)"></div>
              </div>
              <div class="content-section" v-if="programmingExamples.length > 0">
                <h4>样例</h4>
                <div v-for="(example, index) in programmingExamples" :key="index" class="example-item">
                  <el-alert :title="`样例 ${index + 1}`" type="info" :closable="false">
                    <div slot="default">
                      <p><strong>输入：</strong></p>
                      <pre>{{ example.input }}</pre>
                      <p><strong>输出：</strong></p>
                      <pre>{{ example.output }}</pre>
                    </div>
                  </el-alert>
                </div>
              </div>
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

    <!-- 全卷预览对话框 -->
    <el-dialog
      title="全卷预览（学生视角）"
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
          <h3>{{ form.title || '作业标题' }}</h3>
          <p><strong>描述：</strong>{{ form.description || '无' }}</p>
          <p><strong>开始时间：</strong>{{ formatTime(form.startTime) }}</p>
          <p><strong>结束时间：</strong>{{ formatTime(form.endTime) }}</p>
          <p><strong>总分：</strong>{{ getTotalScore() }} 分</p>
          <p><strong>题目数量：</strong>{{ selectedQuestions.length }} 题</p>
        </el-card>

        <!-- 题目列表（和学生看到的一样） -->
        <div class="questions-preview">
          <el-divider content-position="left">题目内容</el-divider>
          <div v-for="(item, index) in selectedQuestions" :key="item.id || item.problemId" class="question-item">
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <span class="question-type">({{ getQuestionTypeText(item.type) }})</span>
              <span class="question-score">{{ item.score }}分</span>
            </div>
            <div class="question-title markdown-body" v-html="renderMarkdown(item.title)"></div>
            <div class="question-content markdown-body" v-html="renderMarkdown(item.content)"></div>

            <!-- 单选题选项（预览模式，不可作答） -->
            <div v-if="item.type === 'single_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
                <div class="option-preview">
                  <span class="option-letter">{{ option.letter }}.</span>
                  <span class="option-text markdown-body" v-html="renderMarkdown(option.text)"></span>
                </div>
              </div>
            </div>

            <!-- 多选题选项（预览模式，不可作答） -->
            <div v-if="item.type === 'multiple_choice'" class="question-options">
              <div v-for="(option, idx) in parseOptions(item.options)" :key="idx" class="option-item">
                <div class="option-preview">
                  <span class="option-letter">{{ option.letter }}.</span>
                  <span class="option-text markdown-body" v-html="renderMarkdown(option.text)"></span>
                </div>
              </div>
            </div>

            <!-- 判断题（预览模式，不可作答） -->
            <div v-if="item.type === 'judge'" class="question-options">
              <div class="option-preview">
                <span class="option-letter">✓</span>
                <span class="option-text">正确</span>
              </div>
              <div class="option-preview">
                <span class="option-letter">✗</span>
                <span class="option-text">错误</span>
              </div>
            </div>

            <!-- 主观题 -->
            <div v-if="item.type === 'subjective'" class="subjective-preview">
              <el-alert type="info" :closable="false">
                <i class="el-icon-edit"></i> 主观题，学生需要在此处输入文字答案
              </el-alert>
            </div>

            <!-- 编程题 -->
            <div v-if="item.type === 'programming'" class="programming-preview">
              <template v-if="getProgrammingProblemInfo(item.problemId)">
                <div class="problem-preview">
                  <el-divider content-position="left">编程题详情</el-divider>
                  <el-card>
                    <h3>{{ getProgrammingProblemInfo(item.problemId).problem.title }}</h3>
                    <div class="problem-meta">
                      <el-tag size="small">题目ID: {{ getProgrammingProblemInfo(item.problemId).problem.problemId }}</el-tag>
                      <el-tag size="small" type="info">时间限制: {{ getProgrammingProblemInfo(item.problemId).problem.timeLimit }}ms</el-tag>
                      <el-tag size="small" type="warning">内存限制: {{ getProgrammingProblemInfo(item.problemId).problem.memoryLimit }}MB</el-tag>
                      <el-tag size="small" type="success">判题模式: {{ getJudgeModeText(getProgrammingProblemInfo(item.problemId).problem.judgeMode) }}</el-tag>
                    </div>
                    <div class="problem-content">
                      <div class="content-section">
                        <h4>题目描述</h4>
                        <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.description)"></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.input">
                        <h4>输入格式</h4>
                        <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.input)"></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingProblemInfo(item.problemId).problem.output">
                        <h4>输出格式</h4>
                        <div v-html="renderMarkdown(getProgrammingProblemInfo(item.problemId).problem.output)"></div>
                      </div>
                      <div class="content-section" v-if="getProgrammingExamplesForProblem(item.problemId).length > 0">
                        <h4>样例</h4>
                        <div v-for="(example, idx) in getProgrammingExamplesForProblem(item.problemId)" :key="idx" class="example-item">
                          <el-alert :title="`样例 ${idx + 1}`" type="info" :closable="false">
                            <div slot="default">
                              <p><strong>输入：</strong></p>
                              <pre>{{ example.input }}</pre>
                              <p><strong>输出：</strong></p>
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
                <p>学生将看到完整的编程题目界面，包括题目描述、样例、代码编辑器和提交按钮</p>
              </el-alert>

              <el-alert v-if="!item.problemId" type="warning" :closable="false">
                此编程题未关联 BingOJ 题目 ID
              </el-alert>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 题目详情对话框 -->
    <el-dialog :title="$t('m.Question_Detail')" :visible.sync="showDetailDialog" width="800px">
      <div v-if="currentQuestion" class="question-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item :label="$t('m.Question_Type')">
            {{ getQuestionTypeText(currentQuestion.type) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Question_Title')">
            <div v-html="renderMarkdown(currentQuestion.title)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Content')">
            <div v-html="renderMarkdown(currentQuestion.content)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Options')" v-if="currentQuestion.options">
            <div v-html="renderOptions(currentQuestion.options)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Answer')">
            {{ currentQuestion.answer }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Difficulty')">
            <el-rate :value="currentQuestion.difficulty" :max="3" disabled />
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Score')">
            {{ currentQuestion.score }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import teacherAuth from '@/mixins/teacherAuth'
import moment from 'moment'
import { getJudgeInfo } from '@/common/judgeTerminal'
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
      searchKeyword: '',
      filterType: '',
      currentPage: 1,
      pageSize: 10,
      total: 0,
      questionBank: [],
      selectedQuestions: [],
      showDetailDialog: false,
      showFullPreviewDialog: false, // 全卷预览对话框
      currentQuestion: null,
      // BingOJ 编程题相关
      showAddProgrammingDialog: false,
      programmingForm: {
        problemId: '',
        score: 10
      },
      programmingProblemPreview: null,
      programmingExamples: [],
      fetchingProblem: false,
      programmingProblemsCache: {}, // 缓存编程题信息，用于预览
      form: {
        title: '',
        description: '',
        startTime: null,
        endTime: null,
        showHomework: false,
        showScore: false,
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
    isEditMode() {
      return !!this.$route.query.editId
    },
    editId() {
      return this.$route.query.editId
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
    }
  },
  mounted() {
    this.loadQuestionBank()
    // 如果是编辑模式，加载作业数据
    if (this.isEditMode) {
      this.loadHomeworkData()
    }
  },
  methods: {
    async loadQuestionBank() {
      this.questionsLoading = true
      try {
        const params = {
          classroomId: this.classroomId,
          page: this.currentPage,
          limit: this.pageSize
        }
        if (this.searchKeyword) {
          params.keyword = this.searchKeyword
        }
        if (this.filterType) {
          params.type = this.filterType
        }

        const res = await this.$store.dispatch('classroom/getQuestionBank', params)
        if (res.code === 200) {
          const questions = res.data.questions || res.data || []
          this.questionBank = questions.map(q => ({
            ...q,
            difficulty: parseInt(q.difficulty) || 2, // 确保是数字类型
            score: q.score || 10 // 默认分值
          }))
          this.total = res.data.total || questions.length
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.questionsLoading = false
      }
    },
    handleSelectionChange(selection) {
      // 不再使用勾选框，此方法保留但不做任何操作
    },
    isQuestionSelected(question) {
      // 检查题目是否已添加（通过id或problemId）
      return this.selectedQuestions.some(sq =>
        (sq.id === question.id) ||
        (sq.problemId && sq.problemId === question.problemId)
      )
    },
    addQuestion(question) {
      // 检查是否已添加
      if (this.isQuestionSelected(question)) {
        this.$message.warning('该题目已添加')
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
          subjective: 5,         // 主观题默认5分
          programming: 20        // 编程题默认20分
        }
        newQuestion.score = defaultScores[newQuestion.type] || 10
      }

      this.selectedQuestions.push(newQuestion)

      this.$message.success('添加成功')
    },
    removeQuestion(question) {
      const index = this.selectedQuestions.findIndex(q => q.id === question.id)
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
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
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
        subjective: 'info',
        programming: 'danger'
      }
      return colorMap[type] || 'info'
    },
    // 将数据库中的难度值（1-3）转换为 el-rate 的星星数（1-5）
    getDifficultyStars(difficulty) {
      // 数据库: 1=简单, 2=中等, 3=困难
      // 显示: 映射为 1星, 3星, 5星
      const difficultyMap = {
        1: 1, // 简单 -> 1星
        2: 3, // 中等 -> 3星
        3: 5  // 困难 -> 5星
      }
      return difficultyMap[difficulty] || 3 // 默认3星
    },
    renderOptions(options) {
      if (!options) return '-'
      try {
        const opts = JSON.parse(options)
        return opts.map(opt => {
          // Render each option with markdown
          return this.renderMarkdown(opt)
        }).join('<br>')
      } catch (e) {
        return options.replace(/\n/g, '<br>')
      }
    },
    handlePageChange(page) {
      this.currentPage = page
      this.loadQuestionBank()
    },
    async loadHomeworkData() {
      this.loadingData = true
      try {
        console.log('开始加载作业数据, editId:', this.editId)
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', this.editId)
        console.log('getHomeworkDetail 响应:', res)
        if (res.code === 200 && res.data) {
          const homework = res.data
          console.log('作业数据:', homework)
          console.log('题目列表:', homework.questions)

          // 先填充表单数据（包括 isExamMode），立即更新界面
          this.form = {
            title: homework.title,
            description: homework.description,
            startTime: new Date(homework.startTime),
            endTime: new Date(homework.endTime),
            showHomework: homework.showHomework === 1,
            showScore: homework.showScore === 1,
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
              console.log(`处理题目 ${index + 1}:`, item)
              // 编程题：使用 problemId
              if (item.problemId) {
                console.log('  -> 这是编程题')
                // 从缓存获取编程题详情
                const cached = this.programmingProblemsCache[item.problemId]
                console.log('  -> 缓存数据:', cached)
                const score = item.score || item.question?.score || 20
                if (cached && cached.problem) {
                  console.log('  -> 使用缓存数据')
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
                console.log('  -> 使用基本信息（无缓存）')
                return {
                  id: item.id,
                  problemId: item.problemId,
                  title: `BingOJ 编程题 - ${item.problemId}`,
                  type: 'programming',
                  difficulty: 5,
                  score: score
                }
              }
              // 普通题目：使用 question
              if (item.question && item.question.id) {
                console.log('  -> 这是普通题目')
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
            console.log('最终 selectedQuestions:', this.selectedQuestions)
          }

          // 关键数据填充完成，隐藏 loading
          this.loadingData = false

          // 异步加载编程题详情（不阻塞界面）
          const programmingProblemIds = homework.questions
            .filter(q => q.problemId)
            .map(q => q.problemId)

          if (programmingProblemIds.length > 0) {
            console.log('预加载编程题数据:', programmingProblemIds)
            // 不使用 await，让加载在后台进行，完成后会自动更新界面
            this.loadProgrammingProblemByIds(programmingProblemIds)
          }
        }
      } catch (error) {
        console.error('加载作业数据失败:', error)
        this.$message.error('加载作业数据失败: ' + (error.message || error))
      } finally {
        this.loadingData = false
      }
    },
    async saveHomework() {
      if (!this.$refs.homeworkForm) {
        this.$message.error('表单未初始化')
        return
      }

      this.$refs.homeworkForm.validate(async (valid) => {
        if (!valid) {
          this.$message.warning('请填写所有必填项')
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
              result.questionId = q.id
              result.problemId = null
            }
            return result
          })
        }

        // 调试：打印提交的数据
        console.log('=== 提交作业数据 ===')
        console.log('完整数据:', JSON.stringify(data, null, 2))
        console.log('questions 数量:', data.questions.length)
        data.questions.forEach((q, idx) => {
          console.log(`题目${idx + 1}:`, JSON.stringify(q, null, 2))
        })

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

          console.log('=== 后端响应 ===')
          console.log('完整响应:', res)

          if (res.code === 200) {
            this.$message.success(this.isEditMode ? this.$t('m.Update_Success') : this.$t('m.Create_Success'))
            this.goBack()
          } else {
            this.$message.error(res.message || (this.isEditMode ? '保存失败' : '创建失败'))
          }
        } catch (error) {
          console.error('=== 请求异常 ===')
          console.error('错误详情:', error)
          this.$message.error(this.isEditMode ? '保存失败' : '创建失败')
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
        this.$message.warning('请输入 BingOJ 题目 ID')
        return
      }

      // 获取当前登录用户信息
      const userInfo = this.$store.getters.userInfo
      console.log('当前用户信息:', userInfo)

      if (!userInfo || !userInfo.username) {
        this.$message.warning('请先登录')
        return
      }

      // 从 localStorage 获取 token
      const token = localStorage.getItem('token')
      console.log('当前 Token:', token ? token.substring(0, 20) + '...' : '无')

      if (!token) {
        this.$message.warning('未找到登录凭证，请重新登录')
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

        const res = await getJudgeInfo(requestData)
        console.log('HOJ API 响应:', res)

        if (res.code === 200 && res.data) {
          this.programmingProblemPreview = res.data
          this.extractProgrammingExamples()
          this.$message.success('题目获取成功')
        } else {
          this.$message.error(res.message || '获取题目失败')
        }
      } catch (error) {
        console.error('获取题目异常:', error)
        this.$message.error('网络错误，获取题目失败')
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
        this.$message.warning('请输入 BingOJ 题目 ID')
        return
      }

      // 必须先预览题目
      if (!this.programmingProblemPreview) {
        this.$message.warning('请先点击"获取题目信息"按钮预览题目')
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
        content: `BingOJ 题目 ID: ${this.programmingForm.problemId}`
      }

      // 检查是否已经添加过该题目
      const exists = this.selectedQuestions.some(q => q.problemId === this.programmingForm.problemId)
      if (exists) {
        this.$message.warning('该题目已添加')
        return
      }

      // 添加到已选题目
      this.selectedQuestions.push(tempQuestion)
      this.$message.success('添加成功')

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
    // 渲染 Markdown
    renderMarkdown(text) {
      if (!text) return ''
      try {
        const rendered = md.render(text)
        console.log('渲染 Markdown，输入长度:', text.length, '输出长度:', rendered.length)
        return rendered
      } catch (e) {
        console.error('Markdown 渲染失败:', e)
        return text
      }
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
    },
    // 全卷预览相关方法
    async showFullPreview() {
      if (this.selectedQuestions.length === 0) {
        this.$message.warning('请先添加题目')
        return
      }

      // 检查是否有编程题需要加载
      const hasProgrammingQuestions = this.selectedQuestions.some(q => q.type === 'programming' && q.problemId)
      if (hasProgrammingQuestions) {
        const loading = this.$loading({
          lock: true,
          text: '正在加载编程题数据...',
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
      console.log('预览对话框已打开，编程题缓存:', this.programmingProblemsCache)
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
        console.warn('未登录，无法加载编程题数据')
        return
      }

      console.log('开始批量加载编程题:', problemIds)

      // 过滤掉已缓存的
      const uncachedIds = problemIds.filter(id => !this.programmingProblemsCache[id])
      console.log('需要加载的编程题:', uncachedIds)

      for (const problemId of uncachedIds) {
        try {
          console.log('加载编程题:', problemId)
          const res = await getJudgeInfo({
            pid: problemId,
            cid: '0',
            mode: 'normal',
            username: userInfo.username,
            token: token,
            password: ''
          })

          if (res.code === 200 && res.data) {
            console.log('编程题加载成功:', problemId)
            console.log('题目描述包含数学公式:', res.data.problem.description.includes('$'))
            this.$set(this.programmingProblemsCache, problemId, res.data)
          }
        } catch (error) {
          console.error('加载编程题信息失败:', problemId, error)
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
    parseOptions(optionsStr) {
      if (!optionsStr) return []
      try {
        const options = JSON.parse(optionsStr)
        return options.map((opt, idx) => ({
          letter: String.fromCharCode(65 + idx), // A, B, C, D...
          text: opt
        }))
      } catch (e) {
        return []
      }
    },
    formatTime(time) {
      if (!time) return '-'
      return moment(time).format('YYYY-MM-DD HH:mm')
    },

    goBack() {
      this.$router.go(-1)
    }
  }
}
</script>

<style scoped>
/* 整体布局 - 清爽浅色主题 */
.create-homework-wrapper {
  min-height: 100vh;
  background: #f5f7fa;
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

/* 题目管理卡片 - 左右布局 */
.questions-management-card {
  background: #fff;
}

.questions-container-layout {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 20px;
  padding: 0 20px 20px;
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

/* 右侧已选题目面板 */
.selected-questions-panel {
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
  align-items: center;
  gap: 12px;
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

.question-content {
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

.question-meta {
  display: flex;
  align-items: center;
  gap: 12px;
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
</style>

<!-- 引入 KaTeX 样式 -->
<style>
@import '~katex/dist/katex.min.css';
</style>
