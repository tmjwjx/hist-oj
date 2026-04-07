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
          <!-- 题目区域 -->
          <div class="selected-questions-panel-full">
            <!-- 顶部工具栏：添加题目按钮 -->
            <div class="questions-toolbar">
              <div class="toolbar-left">
                <span class="panel-title">已选题目</span>
                <el-tag size="mini" type="success">{{ selectedQuestions.length }} 题</el-tag>
              </div>
              <div class="toolbar-right">
                <el-button-group>
                  <el-button type="primary" icon="el-icon-collection" @click="goToQuestionBank">
                    客观题题库
                  </el-button>
                  <el-button type="success" icon="el-icon-document" @click="showAddProgrammingDialog = true">
                    编程题题库
                  </el-button>
                  <el-button type="warning" icon="el-icon-download" @click="openImportPaperDialog">
                    导入试卷
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

                      <!-- 显示课程和标签 -->
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
                      <el-tooltip content="查看详情" placement="top">
                        <el-button
                          size="mini"
                          icon="el-icon-view"
                          @click="viewQuestionDetail(question)"
                        />
                      </el-tooltip>
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
              <p class="hint">从上方工具栏中添加题目</p>
            </div>
          </div>
        </div>
      </el-card>

      </el-form>

    </div>

    <!-- 添加 BingOJ 编程题对话框 -->
    <el-dialog title="添加 BingOJ 编程题" :visible.sync="showAddProgrammingDialog" width="1100px">
      <el-form :model="programmingForm" label-width="120px">
        <el-form-item label="方式选择">
          <el-radio-group v-model="programmingInputMode" @change="handleProgrammingInputModeChange">
            <el-radio label="manual">手动输入题目ID</el-radio>
            <el-radio label="tag">按标签选择题目</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 手动输入模式 -->
        <template v-if="programmingInputMode === 'manual'">
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
              :type="getTagById(tagId)?.type || 'primary'"
              :color="getTagById(tagId)?.color || '#409eff'"
              effect="dark"
              style="margin-right: 10px; margin-bottom: 10px;"
              size="medium"
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
                已选标签下共有 <strong>{{ filteredProblemsTotal }}</strong> 道题目（当前显示前 {{ filteredProblemsByTag.length }} 道），点击题号可查看详情
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
                  <el-link type="primary" @click="viewProblemDetail(row)">{{ row.problemId }}</el-link>
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

        <!-- 题目详情查看弹窗 -->
        <el-dialog
          title="题目详情"
          :visible.sync="showTagProblemDetailDialog"
          width="900px"
          append-to-body
        >
          <div v-if="tagProblemDetail" v-loading="loadingTagProblemDetail">
            <div class="problem-detail-content">
              <h3>{{ tagProblemDetail.title }}</h3>
              <div class="problem-meta">
                <el-tag size="small">题目ID: {{ tagProblemDetail.problemId }}</el-tag>
                <el-tag size="small" type="info">时间限制: {{ tagProblemDetail.timeLimit }}ms</el-tag>
                <el-tag size="small" type="warning">内存限制: {{ tagProblemDetail.memoryLimit }}MB</el-tag>
                <el-tag size="small" type="success">难度: {{ getDifficultyName(tagProblemDetail.difficulty) }}</el-tag>
              </div>
              <div class="problem-body">
                <div class="content-section">
                  <h4>题目描述</h4>
                  <div v-html="renderMarkdown(tagProblemDetail.description)"></div>
                </div>
                <div v-if="tagProblemDetail.input" class="content-section">
                  <h4>输入格式</h4>
                  <div v-html="renderMarkdown(tagProblemDetail.input)"></div>
                </div>
                <div v-if="tagProblemDetail.output" class="content-section">
                  <h4>输出格式</h4>
                  <div v-html="renderMarkdown(tagProblemDetail.output)"></div>
                </div>
              </div>
              <div slot="footer" style="text-align: right;">
                <el-button type="primary" @click="showTagProblemDetailDialog = false">关闭</el-button>
              </div>
            </div>
          </div>
        </el-dialog>

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
          v-if="programmingInputMode === 'manual'"
          type="primary"
          @click="confirmAddProgrammingQuestion"
          :disabled="!programmingProblemPreview"
        >
          确定添加
        </el-button>
      </span>
    </el-dialog>

    <!-- 导入试卷对话框 -->
    <el-dialog title="导入试卷" :visible.sync="showImportPaperDialog" width="800px">
      <div v-loading="loadingPapers" element-loading-text="加载试卷列表中...">
        <el-form label-width="80px">
          <el-form-item label="试卷筛选">
            <el-input
              v-model="paperSearchKeyword"
              placeholder="搜索试卷名称"
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
                搜索
              </el-button>
            </el-input>
          </el-form-item>
        </el-form>

        <el-divider content-position="left">试卷列表</el-divider>

        <div v-if="examPapers.length === 0 && !loadingPapers" class="empty-papers">
          <i class="el-icon-info"></i>
          <p>暂无可用试卷</p>
          <p class="hint">您只能导入自己创建的试卷或共享试卷</p>
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
                      <el-tag v-if="paper.isShared === 1" size="mini" type="success">共享</el-tag>
                      <el-tag v-else size="mini" type="info">私有</el-tag>
                      <el-tag size="mini" type="primary">{{ paper.questionCount }}题</el-tag>
                      <el-tag size="mini" type="warning">{{ paper.totalScore }}分</el-tag>
                    </div>
                  </div>
                  <div v-if="paper.description" class="paper-description">{{ paper.description }}</div>
                  <div class="paper-meta">
                    <span class="creator">
                      <i class="el-icon-user"></i>
                      {{ paper.creator ? paper.creator.username : '未知' }}
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
        <el-button @click="showImportPaperDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="confirmImportPaper"
          :disabled="!selectedPaperId"
          :loading="importingPaper"
        >
          确定导入
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
    <el-dialog :title="currentQuestion?.type === 'programming' ? '编程题详情' : $t('m.Question_Detail')" :visible.sync="showDetailDialog" width="900px">
      <div v-if="currentQuestion" class="question-detail">
        <!-- 编程题详情 -->
        <div v-if="currentQuestion.type === 'programming' && currentQuestion.problem">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="题目 ID">
              {{ currentQuestion.problem.problemId }}
            </el-descriptions-item>
            <el-descriptions-item label="题目类型">
              <el-tag type="danger" size="small">编程题</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="题目标题" :span="2">
              <strong>{{ currentQuestion.problem.title }}</strong>
            </el-descriptions-item>
            <el-descriptions-item label="时间限制">
              {{ currentQuestion.problem.timeLimit }} ms
            </el-descriptions-item>
            <el-descriptions-item label="内存限制">
              {{ currentQuestion.problem.memoryLimit }} MB
            </el-descriptions-item>
            <el-descriptions-item label="题目描述" :span="2">
              <div v-html="renderMarkdown(currentQuestion.problem.description)" class="markdown-body detail-content"></div>
            </el-descriptions-item>
            <el-descriptions-item label="输入格式" v-if="currentQuestion.problem.input">
              <div v-html="renderMarkdown(currentQuestion.problem.input)" class="markdown-body detail-content"></div>
            </el-descriptions-item>
            <el-descriptions-item label="输出格式" v-if="currentQuestion.problem.output">
              <div v-html="renderMarkdown(currentQuestion.problem.output)" class="markdown-body detail-content"></div>
            </el-descriptions-item>
            <el-descriptions-item label="样例" :span="2" v-if="currentQuestion.problem.examples">
              <div v-for="(example, idx) in parseExamples(currentQuestion.problem.examples)" :key="idx" class="example-box">
                <div class="example-title">样例 {{ idx + 1 }}</div>
                <div><strong>输入：</strong><pre>{{ example.input }}</pre></div>
                <div><strong>输出：</strong><pre>{{ example.output }}</pre></div>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="提示" v-if="currentQuestion.problem.hint">
              <div v-html="renderMarkdown(currentQuestion.problem.hint)" class="markdown-body detail-content"></div>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 客观题详情 -->
        <el-descriptions v-else :column="1" border>
          <el-descriptions-item :label="$t('m.Question_Type')">
            {{ getQuestionTypeText(currentQuestion.type) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Question_Title')">
            <div v-html="renderMarkdown(currentQuestion.title)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Content')" v-if="currentQuestion.content">
            <div v-html="renderMarkdown(currentQuestion.content)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Options')" v-if="currentQuestion.options">
            <div v-html="renderOptions(currentQuestion.options)" class="markdown-body"></div>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Answer')">
            {{ currentQuestion.answer }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Difficulty')">
            <el-rate :value="getDifficultyStars(currentQuestion.difficulty)" :max="3" disabled />
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Score')">
            {{ currentQuestion.score }} 分
          </el-descriptions-item>
          <el-descriptions-item label="课程" v-if="currentQuestion.course">
            {{ currentQuestion.course }}
          </el-descriptions-item>
          <el-descriptions-item label="标签" v-if="currentQuestion.tags">
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
import { getJudgeInfo } from '@/common/judgeTerminal'
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
      searchKeyword: '',
      filterType: '',
      filterCourse: '', // 课程筛选
      filterTag: '', // 标签筛选
      currentPage: 1,
      pageSize: 10,
      total: 0,
      questionBank: [],
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
    this.loadProblemTagsAndClassification()
    // 如果是编辑模式，加载作业数据
    if (this.isEditMode) {
      this.loadHomeworkData()
    }
    // 检查是否有从题库浏览器返回的题目
    this.loadSelectedQuestionsFromStore()
  },
  methods: {
    goToQuestionBank() {
      this.$router.push({
        name: 'QuestionBankBrowser',
        params: { classroomId: this.classroomId }
      })
    },
    loadSelectedQuestionsFromStore() {
      const storedQuestions = this.$store.state.classroom.selectedQuestions || []
      if (storedQuestions.length > 0) {
        // 将题库浏览器选中的题目添加到已选题目列表
        storedQuestions.forEach(question => {
          if (!this.selectedQuestions.find(q => q.id === question.id)) {
            this.selectedQuestions.push({
              ...question,
              questionOrder: this.selectedQuestions.length + 1,
              question: question // 保存完整的题目信息
            })
          }
        })
        // 清空 store 中的临时题目
        this.$store.commit('classroom/SET_SELECTED_QUESTIONS', [])
        this.$message.success(`已从题库添加 ${storedQuestions.length} 道题目`)
      }
    },
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
        if (this.filterCourse) {
          params.course = this.filterCourse
        }
        if (this.filterTag) {
          params.tag = this.filterTag
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
          composite: 10,         // 组合题默认10分
          subjective: 5,         // 主观题默认5分
          programming: 20        // 编程题默认20分
        }
        newQuestion.score = defaultScores[newQuestion.type] || 10
      }

      this.selectedQuestions.push(newQuestion)

      this.$message.success('添加成功')
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
              text: '加载题目中...',
              spinner: 'el-icon-loading',
              background: 'rgba(0, 0, 0, 0.7)'
            })
            await this.loadProgrammingProblemByIds([question.problemId])
            loading.close()
          } catch (error) {
            this.$message.error('加载题目失败')
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
        composite: '组合题',
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
        composite: 'danger',
        subjective: 'info',
        programming: 'danger'
      }
      return colorMap[type] || 'info'
    },
    getDifficultyStars(difficulty) {
      return parseInt(difficulty) || 1
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
    handlePageChange(page) {
      this.currentPage = page
      this.loadQuestionBank()
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
                  title: `BingOJ 编程题 - ${item.problemId}`,
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
          this.$message.error('存在无效题目ID，请删除后重新添加')
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
            this.$message.error(res.message || (this.isEditMode ? '保存失败' : '创建失败'))
          }
        } catch (error) {
          console.error('保存作业失败:', error)
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
          this.$message.warning('请先登录')
          this.showTagProblemDetailDialog = false
          return
        }

        const res = await getJudgeInfo({
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
          this.$message.error('获取题目详情失败')
        }
      } catch (error) {
        console.error('获取题目详情失败:', error)
        this.$message.error('获取题目详情失败')
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
      this.$message.success('已选择题目')
    },
    // 通过标签选择题目（直接添加）
    async selectProblemByTag(problem) {
      this.programmingForm.problemId = problem.problemId

      // 获取完整题目信息
      const userInfo = this.$store.getters.userInfo
      const token = localStorage.getItem('token')

      if (!userInfo || !userInfo.username || !token) {
        this.$message.warning('请先登录')
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

        const res = await getJudgeInfo(requestData)

        if (res.code === 200 && res.data) {
          const problemData = res.data

          // 检查是否已经添加过该题目
          const exists = this.selectedQuestions.some(q => q.problemId === this.programmingForm.problemId)
          if (exists) {
            this.$message.warning('该编程题已添加')
            return
          }

          // 直接添加编程题
          const tempQuestion = {
            id: `hoj_${this.programmingForm.problemId}`,
            problemId: this.programmingForm.problemId,
            title: problemData.problem.title,
            type: 'programming',
            difficulty: 5,
            score: this.programmingForm.score,
            content: `BingOJ 题目 ID: ${this.programmingForm.problemId}`
          }

          this.selectedQuestions.push(tempQuestion)
          this.$message.success(`已添加题目：${problem.problemId} - ${problem.title}`)
          // 清空表单以便继续添加，但保持对话框打开
          this.programmingForm.problemId = ''
          this.programmingProblemPreview = null
          this.programmingExamples = []
        } else {
          this.$message.error(res.message || '获取题目信息失败')
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
    // 渲染 Markdown
    renderMarkdown(text) {
      if (!text) return ''
      try {
        return md.render(text)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
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
          const res = await getJudgeInfo({
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
        const params = {
          page: this.paperCurrentPage,
          limit: this.paperPageSize
        }
        if (this.paperSearchKeyword) {
          params.keyword = this.paperSearchKeyword
        }

        const res = await this.$store.dispatch('classroom/getExamPaperList', params)
        if (res.code === 200) {
          this.examPapers = res.data.papers || res.data || []
          this.paperTotal = res.data.total || 0
        }
      } catch (error) {
        console.error('加载试卷列表失败:', error)
        this.$message.error('加载试卷列表失败')
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
        this.$message.warning('请选择要导入的试卷')
        return
      }

      this.importingPaper = true
      try {
        // 确保参数是数字类型，并验证有效性
        const paperId = parseInt(this.selectedPaperId)
        const classroomId = parseInt(this.classroomId)

        // 验证转换结果
        if (isNaN(paperId) || paperId <= 0) {
          this.$message.error('请选择有效的试卷')
          return
        }
        if (isNaN(classroomId) || classroomId <= 0) {
          this.$message.error('班级ID无效')
          return
        }

        const requestData = {
          paperId: paperId,
          classroomId: classroomId
        }

        const res = await this.$store.dispatch('classroom/importExamPaper', requestData)

        // 检查响应数据结构
        if (!res) {
          this.$message.error('导入试卷失败：服务器未返回数据')
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

            this.$message.success(`成功导入 ${questions.length} 道题目`)
            this.showImportPaperDialog = false
          } else {
            this.$message.warning('该试卷中没有题目')
          }
        } else {
          this.$message.error(res.message || '导入试卷失败')
        }
      } catch (error) {
        console.error('导入试卷失败:', error)
        // 检查是否是502错误（网关错误）
        if (error.response && error.response.status === 502) {
          this.$message.error('导入试卷失败：服务器繁忙，请稍后重试')
        } else if (error.response && error.response.status === 404) {
          this.$message.error('导入试卷失败：试卷不存在或已被删除')
        } else if (error.response && error.response.status === 403) {
          this.$message.error('导入试卷失败：您没有权限导入此试卷')
        } else {
          this.$message.error('导入试卷失败：' + (error.message || '参数错误，请检查试卷数据'))
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

.detail-content {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  line-height: 1.8;
  max-height: 400px;
  overflow-y: auto;
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
