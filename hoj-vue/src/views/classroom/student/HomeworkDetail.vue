<template>
  <div class="homework-detail student-homework-page">
    <div class="header">
      <h3>{{ homework.title || '作业详情' }}</h3>
      <div class="header-actions">
        <el-button @click="goBack">{{ $t('m.Back') }}</el-button>
      </div>
    </div>

    <!-- 自动保存提示 -->
    <div v-if="autoSaving" class="auto-save-tip">
      <i class="el-icon-loading"></i> 正在自动保存...
    </div>
    <div v-else-if="lastSaveTime" class="auto-save-tip success">
      <i class="el-icon-check"></i> 已于 {{ lastSaveTime }} 自动保存
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-card>
      <div v-if="homework.id">
        <p><strong>{{ $t('m.Homework_Title') }}:</strong> {{ homework.title }}</p>
        <p><strong>{{ $t('m.Description') }}:</strong> {{ homework.description || '-' }}</p>
        <p><strong>{{ $t('m.Start_Time') }}:</strong> {{ formatTime(homework.startTime) }}</p>
        <p><strong>{{ $t('m.End_Time') }}:</strong> {{ formatTime(homework.endTime) }}</p>

        <el-divider></el-divider>

        <!-- 考试模式确认信息（在题目列表上方显示） -->
        <div v-if="isExamMode && !examStarted && !isSubmitted && !dataLoading" class="exam-confirm-inline">
          <div class="exam-confirm-header">
            <i class="el-icon-warning-outline" style="color: #E6A23C; font-size: 32px; margin-right: 10px;"></i>
            <h2>当前为考试模式</h2>
          </div>

          <el-alert
            type="warning"
            :closable="false"
            show-icon
            style="margin: 15px 0;"
          >
            <template slot="title">
              请认真阅读以下考试规则，开始后将无法修改
            </template>
          </el-alert>

          <div class="exam-rules-inline">
            <h3>⏱️ 时间规则</h3>
            <ul>
              <li><strong>考试时长：</strong>{{ examConfig.examDuration }} 分钟</li>
              <li>
                <strong>开始时间：</strong>点击"开始答题"后立即开始计时
              </li>
              <li v-if="examConfig.allowSubmitAfterMinutes > 0">
                <strong>最早交卷时间：</strong>开考后 {{ examConfig.allowSubmitAfterMinutes }} 分钟
              </li>
              <li class="warning">
                <strong>重要：</strong>考试时间到后系统将<strong>强制收卷</strong>，未保存的答案将丢失
              </li>
            </ul>

            <h3>🔒 防作弊规则</h3>
            <ul>
              <li v-if="examConfig.disableCopyPaste">
                ✅ 禁止复制、粘贴任何内容
              </li>
              <li v-if="examConfig.requireFullscreen">
                ✅ 必须保持全屏模式，退出全屏将被记录
              </li>
              <li v-if="examConfig.disallowTabSwitch">
                ✅ 禁止切换浏览器标签页，将被记录
              </li>
              <li>
                ✅ 禁止同时打开其他软件或窗口（系统会检测）
              </li>
            </ul>

            <h3>💡 答题建议</h3>
            <ul>
              <li>请确保网络连接稳定</li>
              <li>建议使用 Chrome 或 Edge 浏览器</li>
              <li>系统会自动保存您的答题进度</li>
              <li>考试结束前可随时修改已答题目</li>
            </ul>
          </div>

          <div class="checkbox-group">
            <el-checkbox v-model="hasReadRules">
              我已仔细阅读并理解以上考试规则，保证遵守考试纪律
            </el-checkbox>
          </div>

          <div class="exam-start-actions">
            <el-button
              size="large"
              @click="goBack"
            >
              取消
            </el-button>
            <el-button
              type="primary"
              size="large"
              :disabled="!canStartExam"
              :loading="submitting"
              @click="startExam"
            >
              <i class="el-icon-edit"></i>
              开始答题
            </el-button>
          </div>
        </div>

        <!-- 试卷列表 - 根据教师设置控制是否显示，数据加载完成前不显示 -->
        <div v-if="canViewHomework && !dataLoading" class="questions-container">
          <h4>{{ $t('m.Questions') }}</h4>
          <div v-for="(item, index) in homework.questions" :key="item.id" class="question-item">
            <!-- 编程题：使用 problemId 判断 -->
            <div v-if="item.problemId" class="programming-question-wrapper">
              <div class="question-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <span class="question-type">(编程题)</span>
                <span class="question-score">{{ item.score }}分</span>
                <!-- 未提交标签（学生提交后老师新增的题目） -->
                <el-tag v-if="isQuestionUnsubmitted(item)" type="warning" size="small" style="margin-left: 10px">
                  未提交
                </el-tag>
                <!-- 编程题作答状态 -->
                <el-tag
                  v-else
                  :type="programmingStatus[item.problemId] === 'submitted' ? 'success' : 'info'"
                  size="small"
                  style="margin-left: 10px"
                >
                  <i v-if="programmingStatus[item.problemId] === 'checking'" class="el-icon-loading"></i>
                  {{ programmingStatusText[item.problemId] || '未作答' }}
                </el-tag>
                <!-- 已提交后显示得分 -->
                <span v-if="canViewScore && questionScores[item.problemId] !== undefined" class="question-score-earned">
                  得分: <span :style="{ color: questionScores[item.problemId] === 0 ? '#F56C6C' : '#67C23A', fontWeight: 'bold' }">{{ questionScores[item.problemId] }}</span>
                  <span v-if="!questionIsScored[item.problemId]" style="color: #909399; font-size: 12px; margin-left: 5px;">(未评分)</span>
                </span>
              </div>
              <ProgrammingQuestion
                :problem-id="item.problemId"
                :question-id="item.id"
                :homework-id="homework.id"
                :can-view-homework="canViewHomework"
                :can-view-score="canViewScore"
                :can-view-answer="canViewAnswer"
                :is-submitted="isSubmitted"
              />
            </div>

            <!-- 普通题目：使用 question 判断 -->
            <div v-else-if="item.question">
              <div class="question-header">
                <span class="question-number">{{ index + 1 }}.</span>
                <span class="question-type">({{ getQuestionTypeText(item.question.type) }})</span>
                <span class="question-score">{{ item.score }}分</span>
                <!-- 未提交标签（学生提交后老师新增的题目） -->
                <el-tag v-if="isQuestionUnsubmitted(item)" type="warning" size="small" style="margin-left: 10px">
                  未提交
                </el-tag>
                <!-- 已提交后显示得分 -->
                <span v-if="canViewScore && questionScores[item.question.id] !== undefined" class="question-score-earned">
                  得分: <span :style="{ color: questionScores[item.question.id] === 0 ? '#F56C6C' : '#67C23A', fontWeight: 'bold' }">{{ questionScores[item.question.id] }}</span>
                  <span v-if="!questionIsScored[item.question.id]" style="color: #909399; font-size: 12px; margin-left: 5px;">(未评分)</span>
                </span>
              </div>
              <div class="question-title markdown-body" v-html="formatContent(item.question.title)" v-highlight></div>
              <div class="question-content markdown-body" v-html="formatContent(item.question.content)" v-highlight></div>

              <!-- 单选题选项 -->
              <div v-if="item.question.type === 'single_choice'" class="question-options">
                <div
                  v-for="(option, idx) in parseOptions(item.question.options)"
                  :key="idx"
                  class="option-item option-item-clickable"
                  @click="handleSingleChoiceSelect(item.question.id, option.letter)"
                >
                  <el-radio
                    v-model="answers[item.question.id]"
                    :label="option.letter"
                    @change="handleAnswerChange"
                    @click.native.stop
                    :disabled="isSubmitted"
                  >
                    <span class="option-rich-text">
                      <span class="option-letter option-head">{{ option.letter }}.</span>
                      <span class="option-content markdown-body" v-html="formatContent(option.text)" v-highlight></span>
                    </span>
                  </el-radio>
                </div>
                <!-- 显示学生已选择的选项 -->
                <div v-if="answers[item.question.id]" class="student-answer">
                  <el-tag type="info">已选: {{ answers[item.question.id] }}</el-tag>
                </div>
                <!-- 显示正确答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="correct-answer">
                  <el-tag type="success">正确答案: {{ item.question.answer }}</el-tag>
                </div>
                <!-- 显示题目解析（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted && item.question.analysis" class="question-analysis">
                  <div class="analysis-title">
                    <i class="el-icon-info" style="color: #409EFF;"></i> 题目解析：
                  </div>
                  <div class="analysis-content markdown-body" v-html="formatContent(item.question.analysis)" v-highlight></div>
                </div>
              </div>

              <!-- 多选题选项 -->
              <div v-if="item.question.type === 'multiple_choice'" class="question-options">
                <div
                  v-for="(option, idx) in parseOptions(item.question.options)"
                  :key="idx"
                  class="option-item option-item-clickable"
                  @click="handleMultipleChoiceSelect(item.question.id, option.letter)"
                >
                  <el-checkbox
                    v-model="multipleAnswers[item.question.id]"
                    :label="option.letter"
                    @change="handleMultipleChoiceChange(item.question.id)"
                    @click.native.stop
                    :disabled="isSubmitted"
                  >
                    <span class="option-rich-text">
                      <span class="option-letter option-head">{{ option.letter }}.</span>
                      <span class="option-content markdown-body" v-html="formatContent(option.text)" v-highlight></span>
                    </span>
                  </el-checkbox>
                </div>
                <!-- 显示学生已选择的选项（按字典序排列） -->
                <div v-if="multipleAnswers[item.question.id] && multipleAnswers[item.question.id].length > 0" class="student-answer">
                  <el-tag type="info">已选: {{ [...multipleAnswers[item.question.id]].sort().join(', ') }}</el-tag>
                </div>
                <!-- 显示正确答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="correct-answer">
                  <el-tag type="success">正确答案: {{ formatMultipleChoiceAnswer(item.question.answer) }}</el-tag>
                </div>
                <!-- 显示题目解析（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted && item.question.analysis" class="question-analysis">
                  <div class="analysis-title">
                    <i class="el-icon-info" style="color: #409EFF;"></i> 题目解析：
                  </div>
                  <div class="analysis-content markdown-body" v-html="formatContent(item.question.analysis)" v-highlight></div>
                </div>
              </div>

              <!-- 判断题 -->
              <div v-if="item.question.type === 'judge'" class="question-options">
                <el-radio
                  v-model="answers[item.question.id]"
                  label="true"
                  @change="handleAnswerChange"
                  :disabled="isSubmitted"
                >正确</el-radio>
                <el-radio
                  v-model="answers[item.question.id]"
                  label="false"
                  @change="handleAnswerChange"
                  :disabled="isSubmitted"
                >错误</el-radio>
                <!-- 显示学生已选择的选项 -->
                <div v-if="answers[item.question.id]" class="student-answer">
                  <el-tag type="info">已选: {{ answers[item.question.id] === 'true' ? '正确' : '错误' }}</el-tag>
                </div>
                <!-- 显示正确答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="correct-answer">
                  <el-tag type="success">正确答案: {{ isJudgeTrue(item.question.answer) ? '正确' : '错误' }}</el-tag>
                </div>
                <!-- 显示题目解析（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted && item.question.analysis" class="question-analysis">
                  <div class="analysis-title">
                    <i class="el-icon-info" style="color: #409EFF;"></i> 题目解析：
                  </div>
                  <div class="analysis-content markdown-body" v-html="formatContent(item.question.analysis)" v-highlight></div>
                </div>
              </div>

              <!-- 主观题 -->
              <div v-if="item.question.type === 'subjective'" class="subjective-answer">
                <el-input
                  v-model="answers[item.question.id]"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入你的答案"
                  @blur="handleAnswerChange"
                  :disabled="isSubmitted"
                />
                <!-- 图片上传区域 -->
                <div v-if="!isSubmitted" class="image-upload-area">
                  <div class="upload-tip">
                    <i class="el-icon-picture-outline"></i> 支持上传图片作为答案（可选）
                  </div>
                  <el-upload
                    :action="uploadUrl"
                    :headers="uploadHeaders"
                    :on-success="(response, file, fileList) => handleUploadSuccess(response, file, fileList, item.question.id)"
                    :on-error="handleUploadError"
                    :on-remove="(file, fileList) => handleRemoveFile(file, fileList, item.question.id)"
                    :file-list="getImageList(item.question.id)"
                    :limit="5"
                    :on-exceed="handleExceed"
                    accept="image/*"
                    list-type="picture-card"
                    :class="{ 'hide-upload-btn': getImageList(item.question.id).length >= 5 }"
                  >
                    <i class="el-icon-plus"></i>
                  </el-upload>
                  <div class="upload-limit-tip">
                    <i class="el-icon-info"></i> 最多上传5张图片，每张图片不超过10MB
                  </div>
                </div>
                <!-- 已提交的图片显示 -->
                <div v-else-if="getSubmittedImages(item.question.id).length > 0" class="submitted-images">
                  <div class="submitted-images-title">
                    <i class="el-icon-picture"></i> 已提交的图片：
                  </div>
                  <div class="submitted-images-list">
                    <el-image
                      v-for="(img, idx) in getSubmittedImages(item.question.id)"
                      :key="idx"
                      :src="img"
                      :preview-src-list="getSubmittedImages(item.question.id)"
                      fit="cover"
                      style="width: 100px; height: 100px; margin: 5px;"
                    >
                    </el-image>
                  </div>
                </div>
                <!-- 显示参考答案（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted" class="reference-answer">
                  <div class="reference-answer-title">
                    <i class="el-icon-document"></i> 参考答案：
                  </div>
                  <div v-if="item.question.answer" class="reference-answer-content markdown-body" v-html="formatContent(item.question.answer)" v-highlight></div>
                  <div v-else class="reference-answer-empty">教师未设置答案</div>
                </div>
                <!-- 显示题目解析（仅在已提交且允许查看答案时） -->
                <div v-if="canViewAnswer && isSubmitted && item.question.analysis" class="question-analysis">
                  <div class="analysis-title">
                    <i class="el-icon-info" style="color: #409EFF;"></i> 题目解析：
                  </div>
                  <div class="analysis-content markdown-body" v-html="formatContent(item.question.analysis)" v-highlight></div>
                </div>
              </div>

              <!-- 组合题 -->
              <div v-if="item.question.type === 'composite'" class="question-options composite-answer-area">
                <div
                  v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(item.question.options)"
                  :key="subQuestion.id || subIndex"
                  class="composite-sub-question-card"
                >
                  <div class="composite-sub-header">
                    <span>子题 {{ subIndex + 1 }}</span>
                    <span class="question-score">{{ Number(subQuestion.score || 0) }}分</span>
                  </div>
                  <div class="markdown-body" v-html="formatContent(subQuestion.content || '')" v-highlight></div>

                  <div
                    v-for="(option, optionIndex) in parseOptions(JSON.stringify(subQuestion.options || []))"
                    :key="optionIndex"
                    class="option-item option-item-clickable"
                    @click="handleCompositeChoiceSelect(item.question.id, subQuestion.id, option.letter)"
                  >
                    <el-tag
                      :type="getCompositeSelectedAnswer(item.question.id, subQuestion.id) === option.letter ? 'primary' : 'info'"
                      size="small"
                    >
                      {{ option.letter }}
                    </el-tag>
                    <span class="option-rich-text">
                      <span class="option-content markdown-body" v-html="formatContent(option.text)" v-highlight></span>
                    </span>
                  </div>

                  <div class="student-answer" v-if="getCompositeSelectedAnswer(item.question.id, subQuestion.id)">
                    <el-tag type="info">已选: {{ getCompositeSelectedAnswer(item.question.id, subQuestion.id) }}</el-tag>
                  </div>
                  <div class="correct-answer" v-if="canViewAnswer && isSubmitted">
                    <el-tag type="success">正确答案: {{ getCompositeCorrectAnswer(item.question.answer, subQuestion.id) || '-' }}</el-tag>
                  </div>
                </div>

                <div v-if="canViewAnswer && isSubmitted && item.question.analysis" class="question-analysis">
                  <div class="analysis-title">
                    <i class="el-icon-info" style="color: #409EFF;"></i> 题目解析：
                  </div>
                  <div class="analysis-content markdown-body" v-html="formatContent(item.question.analysis)" v-highlight></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <el-divider></el-divider>

        <!-- 操作区域 -->
        <div v-if="canSubmit" class="action-area">
          <el-button type="primary" @click="showSubmitConfirm" :loading="submitting">
            <i class="el-icon-upload"></i> 提交作业
          </el-button>
        </div>

        <!-- 已提交信息（只有正式提交后才显示） -->
        <div v-if="isSubmitted && submission">
          <!-- 当可以查看作业时显示详细信息 -->
          <el-alert v-if="canViewHomework" type="success" :closable="false">
            <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(submission.submitTime) }}</p>
            <!-- 根据教师设置决定是否显示分数 -->
            <div v-if="canViewScore">
              <!-- 有未评分题目时显示详细分数信息 -->
              <div v-if="submission.hasUngraded">
                <p><strong>{{ $t('m.Graded_Score') }}:</strong> {{ submission.gradedScore !== undefined ? submission.gradedScore : 0 }}</p>
                <p style="color: #909399; font-size: 12px;">
                  <i class="el-icon-info"></i>
                  {{ $t('m.Has_Ungraded_Questions_Tip') }}
                </p>
              </div>
              <!-- 所有题目都已评分 -->
              <p v-else><strong>{{ $t('m.Score') }}:</strong> <span :style="{ color: submission.score === 0 ? '#F56C6C' : '' }">{{ submission.score !== undefined ? submission.score : $t('m.Not_Graded') }}</span></p>
            </div>
            <p v-else><strong>{{ $t('m.Score') }}:</strong> {{ $t('m.Score_Hidden_Tip') }}</p>
          </el-alert>

          <!-- 当不能查看作业时，显示简化提示 -->
          <el-alert v-else type="info" :closable="false">
            <p><strong>{{ $t('m.Submit_Time') }}:</strong> {{ formatTime(submission.submitTime) }}</p>
            <div v-if="canViewScore">
              <!-- 有未评分题目时显示详细分数信息 -->
              <div v-if="submission.hasUngraded">
                <p><strong>{{ $t('m.Graded_Score') }}:</strong> {{ submission.gradedScore !== undefined ? submission.gradedScore : 0 }}</p>
                <p style="color: #909399; font-size: 12px;">
                  <i class="el-icon-info"></i>
                  {{ $t('m.Has_Ungraded_Questions_Tip') }}
                </p>
              </div>
              <!-- 所有题目都已评分 -->
              <p v-else><strong>{{ $t('m.Score') }}:</strong> <span :style="{ color: submission.score === 0 ? '#F56C6C' : '' }">{{ submission.score !== undefined ? submission.score : $t('m.Not_Graded') }}</span></p>
            </div>
            <p v-else><strong>{{ $t('m.Score') }}:</strong> {{ $t('m.Score_Hidden_Tip') }}</p>
            <p style="margin-top: 10px; color: #909399;">
              <i class="el-icon-info"></i>
              {{ $t('m.Homework_Content_Hidden_Tip') }}
            </p>
          </el-alert>
        </div>
      </div>
    </el-card>

    <!-- 提交确认对话框 -->
    <el-dialog
      title="确认提交"
      :visible.sync="showConfirmDialog"
      width="400px"
    >
      <p>确定要提交作业吗？提交后将无法修改答案。</p>
      <span slot="footer">
        <el-button @click="showConfirmDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmSubmit" :loading="submitting">确认提交</el-button>
      </span>
    </el-dialog>

    <!-- 考试模式顶部标识栏 -->
    <div v-if="isExamMode && examStarted" class="exam-header-bar">
      <div class="exam-badge">
        <i class="el-icon-warning-outline"></i>
        考试模式进行中
      </div>

      <div class="exam-timer">
        <i class="el-icon-time"></i>
        <span class="timer-label">剩余时间：</span>
        <span class="timer-value" :style="{ color: timerColor }">{{ formattedTime }}</span>
      </div>

      <div class="exam-actions">
        <el-button
          type="danger"
          size="small"
          @click="handleExamSubmit"
          :loading="submitting"
        >
          <i class="el-icon-check"></i>
          交卷
        </el-button>
      </div>
    </div>

    <!-- 考试模式题目导航 -->
    <div v-if="isExamMode && examStarted && canViewHomework" class="question-navigator">
      <div class="navigator-title">题目导航</div>
      <div class="navigator-grid">
        <div
          v-for="(item, index) in homework.questions"
          :key="item.id"
          class="question-nav-item"
          :class="getQuestionNavClass(item)"
          @click="scrollToQuestion(index)"
        >
          {{ index + 1 }}
        </div>
      </div>
      <div class="navigator-legend">
        <span class="legend-item"><span class="legend-color unanswered"></span>未答</span>
        <span class="legend-item"><span class="legend-color answered"></span>已答</span>
      </div>
    </div>

    <!-- 考试模式右下角计时器 -->
    <div v-if="isExamMode && examStarted && canViewHomework" class="exam-corner-timer">
      <i class="el-icon-time"></i>
      <span class="timer-label">剩余时间：</span>
      <span class="timer-value" :style="{ color: timerColor }">{{ formattedTime }}</span>
    </div>
  </div>
</template>

<script>
import moment from 'moment'
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
import ProgrammingQuestion from '@/components/classroom/ProgrammingQuestion.vue'
import realtimeSync from '@/mixins/realtimeSync'

// 配置 markdown-it 支持 KaTeX
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

import studentAuth from '@/mixins/studentAuth'
export default {
  name: 'HomeworkDetail',
  mixins: [realtimeSync, studentAuth],
  components: {
    ProgrammingQuestion
  },
  data() {
    return {
      loading: false,
      dataLoading: true, // 数据加载状态，用于控制题目显示时机
      homework: {},
      submission: null,
      submitting: false,
      saving: false,
      showConfirmDialog: false,
      autoSaving: false,
      lastSaveTime: '',
      hasUnsavedChanges: false,
      autoSaveTimer: null,
      isSubmitted: false, // 是否已正式提交（区别于草稿）
      answers: {}, // 单选题、判断题、主观题答案
      multipleAnswers: {}, // 多选题答案数组 { questionId: ['A', 'B'] }
      compositeAnswers: {}, // 组合题答案 { questionId: { subId: 'A' } }
      questionScores: {}, // 每题得分 { questionId: score }
      questionIsScored: {}, // 每题是否已评分 { questionId: boolean }
      programmingStatus: {}, // 编程题状态 { problemId: 'not_started' | 'checking' | 'submitted' }
      programmingStatusText: {}, // 编程题状态文本 { problemId: '未作答' | '检测中...' | '已作答' }
      pollingTimer: null, // 轮询定时器
      isPolling: false, // 是否正在进行轮询请求（防止重复请求）
      isPageVisible: true, // 页面是否可见
      pollingConfig: {
        enabled: true, // 是否启用轮询
        interval: 10000, // 轮询间隔（毫秒）- 增加到10秒
        fastInterval: 10000, // 检测到提交前的快速轮询间隔（从5秒改为10秒）
        slowInterval: 30000, // 所有题目都提交后的慢速轮询间隔（从15秒改为30秒）
        consecutiveErrors: 0, // 连续错误次数
        maxErrors: 3, // 最大连续错误次数（从1改为3，给服务器更多恢复机会）
        pauseDuration: 60000, // 暂停时长（毫秒）
        lastErrorTime: 0, // 上次错误时间
        retryCount: 0, // 重试计数器
        isAllSubmitted: false, // 是否所有编程题都已提交
        requestDelay: 300 // 每个编程题请求之间的延迟（毫秒，从200改为300）
      },
      attachments: {}, // 每题的图片URL列表 { questionId: [url1, url2, ...] }
      uploadedImages: {}, // 已上传的图片文件列表 { questionId: [{name, url, uid}, ...] }
      submittedQuestionIds: new Set(), // 已提交的题目ID集合（用于识别后添加的题目）
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadHomeworkDetail',
        immediate: true
      },
      // 考试模式相关
      examConfirmDialogVisible: false, // 考试确认弹窗
      hasReadRules: false, // 是否已阅读考试规则
      examConfig: {}, // 考试配置
      examStartTime: null, // 考试开始时间
      remainingSeconds: 0, // 剩余秒数
      timeState: 'normal', // 时间状态: normal, warning, critical, overtime
      canSubmitInExam: false, // 考试模式下是否允许交卷
      examTimerInterval: null, // 考试计时器
      isExamMode: false, // 是否是考试模式
      examStarted: false, // 考试是否已开始
      fullscreenWarned: false, // 是否已经警告过退出全屏
      hasShownForceSubmitMessage: false, // 是否已显示过强制收卷消息
      fullscreenChangeTime: 0, // 全屏变化的时间戳，用于避免 visibilitychange 误报
      isFullscreenChanging: false, // 标志：是否正在处理全屏变化（用于过滤visibilitychange事件）
      isWindowFocused: true, // 窗口是否有焦点
      lastFullscreenState: false // 记录上一次的全屏状态，用于判断是进入还是退出全屏
    }
  },
  computed: {
    canSubmit() {
      // 普通模式：只检查状态和是否已提交
      if (!this.isExamMode) {
        return this.homework.status === 2 && !this.isSubmitted
      }
      // 考试模式：需要额外检查是否满足交卷时间限制
      return this.homework.status === 2 && !this.isSubmitted && this.canSubmitInExam
    },
    // 是否可以查看作业（题目和答案）
    canViewHomework() {
      // 已提交后，根据教师设置判断
      if (this.isSubmitted) {
        // 如果教师允许查看作业内容（showHomework === 1），则可以查看
        return this.homework.showHomework === 1
      }

      // 考试模式下，只有考试已开始才能查看题目
      if (this.isExamMode && !this.examStarted) {
        return false
      }

      // 未提交时，作业进行中可以查看
      return this.homework.status === 2
    },
    // 是否可以查看分数
    canViewScore() {
      // 必须已提交且教师允许查看分数
      return this.isSubmitted && this.homework.showScore === 1
    },
    // 是否可以查看答案
    canViewAnswer() {
      // 必须已提交且教师允许查看答案
      return this.isSubmitted && this.homework.showAnswer === 1
    },
    // 图片上传URL
    uploadUrl() {
      return '/rating-api/api/classroom/homework/upload-attachment'
    },
    // 上传请求头
    uploadHeaders() {
      const token = this.$store.getters.token
      return {
        'Authorization': 'Bearer ' + token
      }
    },
    // 考试模式：是否可以开始考试
    canStartExam() {
      return this.hasReadRules
    },
    // 格式化剩余时间
    formattedTime() {
      // 如果已提交或时间已到，显示"考试结束"
      if (this.isSubmitted || this.remainingSeconds <= 0) {
        return '考试结束'
      }
      const hours = Math.floor(this.remainingSeconds / 3600)
      const minutes = Math.floor((this.remainingSeconds % 3600) / 60)
      const seconds = this.remainingSeconds % 60
      return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
    },
    // 计算时间颜色
    timerColor() {
      // 如果已提交或时间已到，显示为红色
      if (this.isSubmitted || this.remainingSeconds <= 0) {
        return '#F56C6C'
      }
      // 剩余时间 <= 5分钟，显示为红色
      if (this.remainingSeconds <= 300) {
        return '#F56C6C'
      }
      // 剩余时间 > 5分钟且 <= 10分钟，显示为黄色
      if (this.remainingSeconds <= 600) {
        return '#E6A23C'
      }
      // 剩余时间 > 10分钟，显示为绿色
      return '#67C23A'
    }
  },
  mounted() {
    this.loadHomeworkDetail()

    // 添加页面可见性监听
    document.addEventListener('visibilitychange', this.handleVisibilityChange)
  },
  beforeDestroy() {
    // 移除页面可见性监听
    document.removeEventListener('visibilitychange', this.handleVisibilityChange)

    // 清除自动保存定时器
    if (this.autoSaveTimer) {
      clearTimeout(this.autoSaveTimer)
    }
    // 清除轮询定时器
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer)
      this.pollingTimer = null
    }
    // 清除考试计时器
    if (this.examTimerInterval) {
      clearInterval(this.examTimerInterval)
    }
    // 清理防作弊监听
    this.cleanupAntiCheat()
    // 页面退出前总是保存所有答案（确保不丢失数据）
    this.saveDraftSync()
  },
  methods: {
    async loadHomeworkDetail() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = !this.homework || !this.homework.id
      if (isFirstLoad) {
        this.loading = true
        this.dataLoading = true
      }

      try {
        const homeworkId = this.$route.params.homeworkId

        // 如果作业已提交，强制刷新数据，不使用缓存
        // 这样可以立即显示最新的答案和分数
        const forceRefresh = this.isSubmitted
        const res = await this.$store.dispatch('classroom/getHomeworkDetail', homeworkId, forceRefresh)

        if (res.code === 200) {
          // 安全措施：清空所有题目的答案和难度字段（防止前端泄露）
          // 注意：学生始终看不到难度，只有在已提交且教师允许时才能看到答案
          if (res.data.questions && Array.isArray(res.data.questions)) {
            res.data.questions.forEach(item => {
              if (item.question) {
                // 学生始终看不到难度
                if (item.question.difficulty) {
                  item.question.difficulty = 0
                }
                // 只有在未提交或教师不允许查看答案时，才清空答案字段
                if (!this.canViewAnswer && item.question.answer) {
                  item.question.answer = ''
                }
              }
            })
          }

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentHomeworkString = JSON.stringify(this.homework)
          const newHomeworkString = JSON.stringify(res.data)

          if (currentHomeworkString !== newHomeworkString) {
            // 数据真的变化了，才更新
            this.homework = res.data

            // 只在首次加载时初始化答案对象，避免覆盖用户正在输入的答案
            if (isFirstLoad) {
              // 初始化所有题目的答案对象
              this.homework.questions.forEach(item => {
                // 编程题跳过
                if (item.problemId) return

                // 普通题目处理
                if (!item.question) return

                const qid = item.question.id
                if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
                  // 使用 $set 确保响应式
                  this.$set(this.answers, qid, '')
                } else if (item.question.type === 'multiple_choice') {
                  this.$set(this.multipleAnswers, qid, [])
                } else if (item.question.type === 'composite') {
                  this.$set(this.compositeAnswers, qid, {})
                }
              })
            }
          }

          // 检查是否是考试模式（需要在加载答案之前检查，以更新题目顺序）
          let needUpdateQuestions = false
          let examQuestions = null

          if (res.data.isExamMode === 1) {
            this.isExamMode = true
            this.examConfig = {
              examDuration: res.data.examDuration,
              allowSubmitAfterMinutes: res.data.allowSubmitAfterMinutes,
              disableCopyPaste: res.data.disableCopyPaste === 1,
              requireFullscreen: res.data.requireFullscreen === 1,
              disallowTabSwitch: res.data.disallowTabSwitch === 1
            }

            // 每次加载都检查考试状态（不仅仅是首次）
            const homeworkId = this.homework.id || this.$route.params.homeworkId
            const examStatusRes = await this.$store.dispatch('classroom/getExamStatus', homeworkId)

            if (examStatusRes.code === 200 && examStatusRes.data && examStatusRes.data.hasStarted) {
              // 已开始，标记需要更新题目列表
              if (examStatusRes.data.questions && Array.isArray(examStatusRes.data.questions) && examStatusRes.data.questions.length > 0) {
                needUpdateQuestions = true
                examQuestions = examStatusRes.data.questions
              }

              // 恢复考试状态
              this.examStarted = true
              this.examStartTime = new Date(examStatusRes.data.examStartTime)
              this.remainingSeconds = examStatusRes.data.remainingSeconds || 0
              this.canSubmitInExam = examStatusRes.data.canSubmit || false

              // 检查是否已经被强制收卷
              if (examStatusRes.data.isSubmitted) {
                this.isSubmitted = true

                // 只有在作业进行中才显示强制收卷提示
                // 作业结束后查看不显示提示
                if (examStatusRes.data.isForcedSubmit &&
                    !this.hasShownForceMessage &&
                    this.homework.status === 2) {
                  // 判断是否是超时收卷（remainingSeconds <= 0）
                  if (this.remainingSeconds <= 0) {
                    this.$message.warning({
                      message: '考试时间已到，系统已自动收卷',
                      duration: 5000,
                      showClose: true
                    })
                  } else {
                    // 还有剩余时间但被收卷，说明是老师强制收卷
                    this.$message.warning({
                      message: '您已被老师强制收卷，请停止答题',
                      duration: 5000,
                      showClose: true
                    })
                  }
                  this.hasShownForceSubmitMessage = true
                }
                // 如果已提交，停止之前的计时器和清理防作弊
                if (this.examTimerInterval) {
                  clearInterval(this.examTimerInterval)
                  this.examTimerInterval = null
                }
                this.cleanupAntiCheat()
              } else {
                // 考试还在进行中，重新启动计时器和防作弊
                // 先清除旧的计时器
                if (this.examTimerInterval) {
                  clearInterval(this.examTimerInterval)
                  this.examTimerInterval = null
                }
                this.startExamTimer()
                this.initAntiCheat()
              }
            } else {
              // 未开始，会在页面中显示考试确认信息（不需要弹窗）
            }
          }

          // 如果需要更新题目列表，在加载答案之前更新
          if (needUpdateQuestions && examQuestions) {
            this.homework.questions = examQuestions
          }

          // 只在首次加载时加载提交记录和初始化答案
          // realtimeSync 时不需要重新加载，避免延迟
          if (isFirstLoad) {
            // 加载已提交的答案（在题目列表更新后加载）
            await this.loadSubmission(true)

            // 初始化答案（确保所有题目都有记录）
            await this.initializeAnswers()
          }

          // 初始化编程题状态检测（内部会自动启动轮询）
          this.initializeProgrammingStatus()
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
          this.dataLoading = false
        }
      }
    },
    async loadSubmission(shouldRestoreAnswers = true) {
      try {
        const homeworkId = this.homework.id || this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getStudentHomeworkDetail', homeworkId)

        if (res.code === 200 && res.data) {
          this.submission = res.data

          // 根据后端返回的 isOfficiallySubmitted 字段设置提交状态
          // 只要有正式提交（包括编程题提交），就设置为已提交
          if (res.data.isOfficiallySubmitted) {
            this.isSubmitted = true

            // 记录已提交的题目ID（用于识别后添加的题目）
            this.submittedQuestionIds.clear()

            // 记录普通题目的ID
            if (res.data.answers) {
              const answersData = JSON.parse(res.data.answers || '{}')
              Object.keys(answersData).forEach(qid => {
                this.submittedQuestionIds.add(qid)
              })
            }

            // 记录编程题的problemId
            if (this.homework.questions) {
              this.homework.questions.forEach(item => {
                if (item.problemId) {
                  // 编程题通过轮询检查状态，如果有提交记录也算已提交
                  this.submittedQuestionIds.add(`programming_${item.problemId}`)
                }
              })
            }
          }

          // 只在首次加载或明确要求时恢复答案
          // 避免在用户正在输入时覆盖其输入
          if (shouldRestoreAnswers && res.data.answers) {
            const answersData = JSON.parse(res.data.answers || '{}')

            // 统计已作答的题目数
            let answeredCount = 0

            Object.keys(answersData).forEach(qid => {
              answeredCount++
              const questionType = this.getQuestionTypeByQuestionId(qid)

              // 按题型恢复答案（多选/组合题为JSON结构）
              try {
                const parsed = JSON.parse(answersData[qid])
                if (questionType === 'multiple_choice' && Array.isArray(parsed)) {
                  // 多选题
                  this.$set(this.multipleAnswers, qid, parsed)
                } else if (questionType === 'composite' && parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
                  this.$set(this.compositeAnswers, qid, parsed)
                } else {
                  // 单选题、判断题、主观题
                  this.$set(this.answers, qid, answersData[qid])
                }
              } catch (e) {
                // 解析失败，说明是普通字符串
                this.$set(this.answers, qid, answersData[qid])
              }
            })
          }

          // 恢复每题得分
          if (res.data.scores) {
            const scoresData = JSON.parse(res.data.scores || '{}')
            Object.keys(scoresData).forEach(qid => {
              this.$set(this.questionScores, qid, scoresData[qid])
            })
          }

          // 恢复每题评分状态
          if (res.data.isScoredMap) {
            const isScoredData = JSON.parse(res.data.isScoredMap || '{}')
            Object.keys(isScoredData).forEach(qid => {
              this.$set(this.questionIsScored, qid, isScoredData[qid])
            })
          }

          // 恢复每题附件
          if (res.data.attachments && shouldRestoreAnswers) {
            const attachmentsData = JSON.parse(res.data.attachments || '{}')
            Object.keys(attachmentsData).forEach(qid => {
              const attachmentStr = attachmentsData[qid]
              if (attachmentStr) {
                // 将逗号分隔的URL字符串转换为数组
                const urls = attachmentStr.split(',').filter(url => url.trim())
                this.$set(this.attachments, qid, urls)
                // 构建上传文件列表（用于el-upload显示）
                const fileList = urls.map((url, idx) => ({
                  name: `image_${idx + 1}`,
                  url: url,
                  uid: Date.now() + idx,
                  response: { data: { url: url } }
                }))
                this.$set(this.uploadedImages, qid, fileList)
              }
            })
          }
        }

        // 重要：加载提交记录后，异步检查编程题状态
        // 不使用 await 阻塞，让学生答案能立即显示
        // 编程题状态会在后台异步更新
        if (this.homework && this.homework.questions && this.homework.questions.some(q => q.problemId)) {
          // 如果作业已提交，只需要检查一次状态，不需要持续轮询
          // 使用一个标志位确保只检查一次
          if (!this._hasCheckedProgrammingStatus) {
            // 不使用 await，让状态检查异步执行
            // 这样答案可以立即显示，不会阻塞
            this.checkProgrammingStatus(true).then(() => {
              this._hasCheckedProgrammingStatus = true
            })
          }
        }
      } catch (error) {
        // 忽略错误，可能还没有提交
      }
    },
    // 处理答案变化 - 触发自动保存
    handleAnswerChange() {
      this.hasUnsavedChanges = true
      this.scheduleAutoSave()
    },
    // 单选题：支持点击整行选项
    handleSingleChoiceSelect(questionId, optionLetter) {
      if (this.isSubmitted) return
      if (this.answers[questionId] === optionLetter) return
      this.$set(this.answers, questionId, optionLetter)
      this.handleAnswerChange()
    },
    // 多选题：支持点击整行选项
    handleMultipleChoiceSelect(questionId, optionLetter) {
      if (this.isSubmitted) return

      const currentAnswers = Array.isArray(this.multipleAnswers[questionId])
        ? [...this.multipleAnswers[questionId]]
        : []

      const existingIndex = currentAnswers.indexOf(optionLetter)
      if (existingIndex >= 0) {
        currentAnswers.splice(existingIndex, 1)
      } else {
        currentAnswers.push(optionLetter)
      }

      currentAnswers.sort()
      this.$set(this.multipleAnswers, questionId, currentAnswers)
      this.handleAnswerChange()
    },
    handleCompositeChoiceSelect(questionId, subId, optionLetter) {
      if (this.isSubmitted) return
      if (!questionId || !subId) return

      const current = this.compositeAnswers[questionId] && typeof this.compositeAnswers[questionId] === 'object'
        ? { ...this.compositeAnswers[questionId] }
        : {}
      if (current[subId] === optionLetter) return
      current[subId] = optionLetter
      this.$set(this.compositeAnswers, questionId, current)
      this.handleAnswerChange()
    },
    getCompositeSelectedAnswer(questionId, subId) {
      const subAnswers = this.compositeAnswers[questionId]
      if (!subAnswers || typeof subAnswers !== 'object') return ''
      return subAnswers[subId] || ''
    },
    getCompositeCorrectAnswer(answerStr, subId) {
      try {
        const parsed = typeof answerStr === 'string' ? JSON.parse(answerStr || '{}') : answerStr
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
          return parsed[subId] || ''
        }
      } catch (e) {
        return ''
      }
      return ''
    },
    getQuestionTypeByQuestionId(questionId) {
      const questionIdStr = String(questionId)
      const target = (this.homework.questions || []).find(item => item && item.question && String(item.question.id) === questionIdStr)
      return target && target.question ? target.question.type : ''
    },
    // 页面初始化后立即保存一次（确保所有题目都有记录）
    async initializeAnswers() {
      try {
        const homeworkId = this.homework.id || parseInt(this.$route.params.homeworkId)

        // 初始化所有题目的答案（即使是空答案）
        const answersData = {}
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            // 如果已有答案就用已有的，否则用空字符串
            answersData[qid] = this.answers[qid] || ''
          } else if (item.question.type === 'composite') {
            answersData[qid] = JSON.stringify(this.compositeAnswers[qid] || {})
          }
        })

        // 初始化多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 调用草稿保存 API（不判分，不标记为已提交）
        const res = await this.$store.dispatch('classroom/saveHomeworkDraft', {
          homeworkId: homeworkId,
          answers: answersData
        })
      } catch (error) {
        // Silently handle error
      }
    },
    // 处理多选题选项变化
    handleMultipleChoiceChange(questionId) {
      // 确保 multipleAnswers[questionId] 是数组
      if (!Array.isArray(this.multipleAnswers[questionId])) {
        this.$set(this.multipleAnswers, questionId, [])
      }
      // 多选题也要触发自动保存
      this.handleAnswerChange()
    },
    // 安排自动保存（延迟执行，避免频繁保存）
    scheduleAutoSave() {
      if (this.autoSaveTimer) {
        clearTimeout(this.autoSaveTimer)
      }
      // 3秒后自动保存
      this.autoSaveTimer = setTimeout(() => {
        this.saveDraft()
      }, 3000)
    },
    // 保存草稿（批量提交）
    async saveDraft() {
      this.autoSaving = true
      try {
        const homeworkId = this.homework.id || parseInt(this.$route.params.homeworkId)

        // 检查题目数据是否存在
        if (!this.homework.questions || this.homework.questions.length === 0) {
          this.$message.warning('题目数据加载中，请稍后再试')
          return
        }

        // 合并所有答案
        const answersData = {}

        // 单选、判断、主观题答案（包括空值，确保保存所有题目的状态）
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            answersData[qid] = this.answers[qid] || ''
          } else if (item.question.type === 'composite') {
            answersData[qid] = JSON.stringify(this.compositeAnswers[qid] || {})
          }
        })

        // 多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 调用草稿保存 API（不判分，不标记为已提交）
        const res = await this.$store.dispatch('classroom/saveHomeworkDraft', {
          homeworkId: homeworkId,
          answers: answersData
        })

        if (res.code === 200) {
          this.hasUnsavedChanges = false
          this.lastSaveTime = moment().format('HH:mm:ss')
          this.$message.success('草稿保存成功')
          // 草稿保存后不需要重新加载提交状态（因为不会改变 isSubmitted）
        } else {
          console.error('保存草稿失败:', res.message)
          this.$message.error(res.message || '保存失败')
        }
      } catch (error) {
        console.error('保存草稿异常:', error.message)
        this.$message.error('保存失败：' + (error.message || '未知错误'))
      } finally {
        this.autoSaving = false
      }
    },
    // 同步保存草稿（用于页面退出时，不显示提示）
    saveDraftSync() {
      try {
        const homeworkId = this.homework.id || parseInt(this.$route.params.homeworkId)

        // 合并所有答案
        const answersData = {}

        // 单选、判断、主观题答案（包括空值）
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            answersData[qid] = this.answers[qid] || ''
          } else if (item.question.type === 'composite') {
            answersData[qid] = JSON.stringify(this.compositeAnswers[qid] || {})
          }
        })

        // 多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 准备附件数据（将URL数组用逗号连接）
        const attachmentsData = {}
        Object.keys(this.attachments).forEach(qid => {
          if (Array.isArray(this.attachments[qid]) && this.attachments[qid].length > 0) {
            attachmentsData[qid] = this.attachments[qid].join(',')
          }
        })

        // 调试日志：打印要发送的附件数据

        // 使用 fetch 的 keepalive 选项来确保页面关闭时也能发送请求
        // 调用草稿保存 API（不判分，不标记为已提交）
        fetch('/api/classroom/homework/draft', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': localStorage.getItem('token') || ''
          },
          body: JSON.stringify({
            homeworkId: homeworkId,
            answers: answersData,
            attachments: attachmentsData
          }),
          keepalive: true
        }).then(response => {
          // Silently handle response
        }).catch(error => {
          // Silently handle error
        })
      } catch (error) {
        // Silently handle error
      }
    },
    // 显示提交确认对话框
    showSubmitConfirm() {
      // 检查是否有未完成的题目
      const unanswered = this.getUnansweredQuestions()
      if (unanswered.length > 0) {
        this.$message.warning(`请完成所有题目后再提交`)
        return
      }
      this.showConfirmDialog = true
    },
    // 获取未完成的题目列表
    getUnansweredQuestions() {
      const unanswered = []
      this.homework.questions.forEach(item => {
        // 编程题处理
        if (item.problemId) {
          // 检查编程题是否已提交
          const status = this.programmingStatus[item.problemId]
          if (status !== 'submitted') {
            unanswered.push(`第${this.getQuestionIndex(item)}题 (编程题)`)
          }
          return
        }

        // 普通题目处理
        if (!item.question) return

        const qid = item.question.id
        if (item.question.type === 'multiple_choice') {
          // 多选题：检查是否至少选择了一个选项
          if (!this.multipleAnswers[qid] || this.multipleAnswers[qid].length === 0) {
            unanswered.push(`第${this.getQuestionIndex(item)}题`)
          }
        } else if (item.question.type === 'composite') {
          const subQuestions = this.parseCompositeSubQuestions(item.question.options)
          const subAnswers = this.compositeAnswers[qid] || {}
          const hasMissing = subQuestions.some(subQuestion => !subAnswers[subQuestion.id])
          if (hasMissing) {
            unanswered.push(`第${this.getQuestionIndex(item)}题`)
          }
        } else {
          // 其他题型：检查是否有答案
          if (!this.answers[qid]) {
            unanswered.push(`第${this.getQuestionIndex(item)}题`)
          }
        }
      })
      return unanswered
    },
    // 获取题目序号
    getQuestionIndex(item) {
      const index = this.homework.questions.findIndex(q => q.id === item.id)
      return index + 1
    },
    // 确认提交
    async confirmSubmit() {
      this.submitting = true
      try {
        const homeworkId = this.homework.id || parseInt(this.$route.params.homeworkId)

        // 合并所有答案
        const answersData = {}

        // 单选、判断、主观题答案
        this.homework.questions.forEach(item => {
          // 编程题跳过
          if (item.problemId) return

          // 普通题目处理
          if (!item.question) return

          const qid = item.question.id
          if (item.question.type === 'single_choice' || item.question.type === 'judge' || item.question.type === 'subjective') {
            answersData[qid] = this.answers[qid] || ''
          } else if (item.question.type === 'composite') {
            answersData[qid] = JSON.stringify(this.compositeAnswers[qid] || {})
          }
        })

        // 多选题答案
        Object.keys(this.multipleAnswers).forEach(qid => {
          if (Array.isArray(this.multipleAnswers[qid])) {
            answersData[qid] = JSON.stringify(this.multipleAnswers[qid])
          }
        })

        // 准备附件数据（将URL数组用逗号连接）

        const attachmentsData = {}
        Object.keys(this.attachments).forEach(qid => {
          if (Array.isArray(this.attachments[qid]) && this.attachments[qid].length > 0) {
            attachmentsData[qid] = this.attachments[qid].join(',')
          } else {
          }
        })

        // 调试日志：打印要发送的附件数据

        // 检查是否有答案
        if (Object.keys(answersData).length === 0) {
          this.$message.warning('请至少作答一道题目')
          return
        }

        // 批量提交
        const res = await this.$store.dispatch('classroom/submitHomework', {
          homeworkId: homeworkId,
          answers: answersData,
          attachments: attachmentsData
        })

        if (res.code === 200) {
          this.$message.success(this.$t('m.Submit_Success'))
          this.showConfirmDialog = false
          this.hasUnsavedChanges = false
          this.isSubmitted = true // 标记为已正式提交

          // 重要：提交成功后立即停止轮询
          this.stopPollingProgrammingStatus()

          // 重新加载提交记录和作业详情
          await this.loadSubmission()
          // 清除 loading 状态，允许重新加载作业详情
          this.loading = false
          // 重新加载作业详情以获取最新数据（包括可能允许查看的答案）
          await this.loadHomeworkDetail()
        } else {
          this.$message.error(res.message || this.$t('m.Submit_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Submit_Failed'))
      } finally {
        this.submitting = false
      }
    },
    parseOptions(optionsStr) {
      try {
        const options = JSON.parse(optionsStr || '[]')
        return options.map((opt, idx) => ({
          letter: ['A', 'B', 'C', 'D'][idx],
          text: opt.replace(/^[A-D]\.\s*/, '')
        }))
      } catch {
        return []
      }
    },
    parseCompositeSubQuestions(optionsStr) {
      try {
        const parsed = typeof optionsStr === 'string' ? JSON.parse(optionsStr || '[]') : optionsStr
        if (!Array.isArray(parsed)) return []
        return parsed.map((sub, index) => ({
          id: String(sub.id || `sq_${index + 1}`),
          content: sub.content || '',
          options: Array.isArray(sub.options) ? sub.options : [],
          score: Number(sub.score || sub.subScore || 0)
        }))
      } catch {
        return []
      }
    },
    formatMultipleChoiceAnswer(answer) {
      if (!answer) return ''
      try {
        const parsed = typeof answer === 'string' ? JSON.parse(answer) : answer
        if (Array.isArray(parsed)) {
          return parsed.join(', ')
        }
      } catch (e) {
        // fall through
      }
      return answer
    },
    isJudgeTrue(answer) {
      const raw = String(answer || '').trim()
      const lowered = raw.toLowerCase()
      return lowered === 'true'
    },
    formatContent(content) {
      if (!content) return ''
      try {
        return md.render(content)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return content
      }
    },
    goToProblem(problemId) {
      window.open(`/problem/${problemId}`, '_blank')
    },
    goBack() {
      this.$router.back()
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm')
    },
    getQuestionTypeText(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        subjective: this.$t('m.Subjective'),
        composite: '组合题',
        programming: this.$t('m.Programming')
      }
      return map[type] || type
    },
    // 检查是否需要轮询（有编程题正在评测中）
    needsPolling() {
      // 如果作业已提交，不需要轮询
      if (this.isSubmitted) {
        return false
      }

      // 如果没有编程题，不需要轮询
      const programmingQuestions = this.homework.questions?.filter(q => q.problemId) || []
      if (programmingQuestions.length === 0) {
        return false
      }

      // 检查是否至少有一个编程题还没有评测结果
      // 条件：状态不是 'completed'（即还没有评测结果）
      const hasJudgingQuestions = programmingQuestions.some(q => {
        const status = this.programmingStatus[q.problemId]
        return status !== 'completed'  // 不是完成状态，说明还在评测中或未提交
      })

      return hasJudgingQuestions
    },
    // 初始化编程题状态
    async initializeProgrammingStatus() {
      // 重要：如果作业已提交，不需要启动轮询
      if (this.isSubmitted) {
        return
      }

      // 重置轮询配置
      this.pollingConfig.enabled = true
      this.pollingConfig.consecutiveErrors = 0
      this.pollingConfig.isAllSubmitted = false
      this.isPageVisible = !document.hidden // 初始化页面可见性

      // 为所有编程题初始化状态
      this.homework.questions.forEach(item => {
        if (item.problemId) {
          this.$set(this.programmingStatus, item.problemId, 'not_started')
          this.$set(this.programmingStatusText, item.problemId, '未作答')
        }
      })

      // 2秒后开始第一次检测并启动轮询（仅在页面可见且未提交时）
      setTimeout(async () => {
        // 再次检查是否已提交（防止竞态条件）
        if (this.isPageVisible && !this.isSubmitted) {
          // 先执行一次检测，更新状态
          await this.checkProgrammingStatus()

          // 检查是否需要继续轮询
          if (this.needsPolling()) {
            this.startPollingProgrammingStatus()
          }
        }
      }, 2000)
    },
    // 启动轮询检测编程题提交状态
    startPollingProgrammingStatus() {
      // 检查页面可见性
      if (!this.isPageVisible) {
        return
      }

      // 重要：检查是否需要轮询
      if (!this.needsPolling()) {
        return
      }

      // 清除之前的定时器（如果存在）
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
      }

      // 根据状态动态调整轮询间隔
      const interval = this.pollingConfig.isAllSubmitted
        ? this.pollingConfig.slowInterval // 所有题已提交，使用慢速轮询（30秒）
        : this.pollingConfig.fastInterval // 还有题目未提交，使用快速轮询（10秒）

      this.pollingTimer = setInterval(() => {
        // 重要：每次轮询时再次检查是否需要轮询
        if (!this.needsPolling()) {
          // 如果不需要轮询了，立即停止
          this.stopPollingProgrammingStatus()
        } else if (this.pollingConfig.enabled && this.isPageVisible && !this.isPolling) {
          this.checkProgrammingStatus()
        }
      }, interval)
    },
    // 检测编程题提交状态
    async checkProgrammingStatus(forceCheck = false) {
      // 如果轮询被禁用或页面不可见，直接返回（除非强制检查）
      if (!forceCheck && (!this.pollingConfig.enabled || !this.isPageVisible)) {
        return
      }

      // 重要：如果作业已提交且非强制检查，停止轮询
      if (!forceCheck && this.isSubmitted) {
        this.stopPollingProgrammingStatus()
        return
      }

      // 防止重复请求：如果已经在轮询中，直接返回
      if (this.isPolling) {
        return
      }

      try {
        this.isPolling = true // 标记开始轮询

        // 使用 this.homework.id 而不是路由参数，因为路由参数可能为 undefined
        const homeworkId = this.homework.id

        // 如果 homeworkId 不存在，跳过检测
        if (!homeworkId) {
          return
        }

        // 获取所有编程题
        const programmingQuestions = this.homework.questions.filter(q => q.problemId)

        if (programmingQuestions.length === 0) {
          // 没有编程题，停止轮询
          this.stopPollingProgrammingStatus()
          return
        }

        let statusChanged = false
        let allCompleted = true  // 改名：所有编程题都已完成评测
        let hasError = false

        // 逐个查询编程题提交状态
        for (const question of programmingQuestions) {
          try {
            // 调用后端API查询该题目的提交记录
            const res = await this.$store.dispatch('classroom/getProgrammingSubmissions', {
              homeworkId: homeworkId,
              questionId: question.id
            })

            const hasSubmission = res.code === 200 && res.data && res.data.length > 0
            const currentStatus = this.programmingStatus[question.problemId]

            if (hasSubmission) {
              // 有提交记录
              const latestSubmission = res.data[0]  // 获取最新的提交记录

              // 后端API返回的字段名是 "result" 而不是 "judgeResult"
              const judgeResult = latestSubmission.result

              // 检查是否有评测结果
              const hasCompleted = judgeResult &&
                                   judgeResult !== '' &&
                                   judgeResult !== 'Pending' &&
                                   judgeResult !== 'Judging' &&
                                   judgeResult !== 'Submitting'

              if (hasCompleted) {
                // 评测完成，更新状态
                if (currentStatus !== 'completed') {
                  this.$set(this.programmingStatus, question.problemId, 'completed')
                  statusChanged = true
                }
                // 状态文本显示"已完成"而不是具体的评测结果
                // 具体的评测结果在下面的"提交历史"中显示
                this.$set(this.programmingStatusText, question.problemId, '已完成')
              } else {
                // 评测中，继续等待
                // 始终更新为"评测中"状态，避免状态回退
                if (currentStatus !== 'judging' && currentStatus !== 'completed') {
                  this.$set(this.programmingStatus, question.problemId, 'judging')
                  this.$set(this.programmingStatusText, question.problemId, '评测中...')
                  statusChanged = true
                }
                allCompleted = false
              }
            } else {
              // 没有提交记录
              // 重要：如果当前状态已经是"评测中"或"已完成"，不要回退到"未作答"
              // 这可能是后端数据延迟导致的，保持当前状态不变
              const currentStatus = this.programmingStatus[question.problemId]
              if (currentStatus === 'judging' || currentStatus === 'completed') {
                // 保持当前状态，不回退
                allCompleted = (currentStatus === 'completed') ? allCompleted : false
              } else {
                // 只有当前不是"评测中"或"已完成"时，才设置为"未作答"
                this.$set(this.programmingStatus, question.problemId, 'not_started')
                this.$set(this.programmingStatusText, question.problemId, '未作答')
                allCompleted = false
              }
            }

            // 添加小延迟，避免瞬间发送多个请求
            if (programmingQuestions.length > 1) {
              await new Promise(resolve => setTimeout(resolve, this.pollingConfig.requestDelay))
            }
          } catch (error) {
            // 单个题目查询失败，继续查询其他题目
            const statusCode = error.response ? error.response.status : 'unknown'

            // 502/503/504错误也要计入错误计数器
            if (statusCode === 502 || statusCode === 503 || statusCode === 504) {
              hasError = true
            } else if (statusCode !== 'unknown') {
              hasError = true
            }
          }
        }

        // 如果有错误发生，增加错误计数
        if (hasError) {
          this.pollingConfig.consecutiveErrors++
          this.pollingConfig.lastErrorTime = Date.now()
        } else {
          // 请求成功，重置错误计数器
          if (this.pollingConfig.consecutiveErrors > 0) {
            this.pollingConfig.consecutiveErrors = 0
            this.pollingConfig.retryCount = 0
          }
        }

        // 如果连续错误超过阈值，暂停轮询并使用指数退避
        if (this.pollingConfig.consecutiveErrors >= this.pollingConfig.maxErrors) {
          this.pollingConfig.enabled = false
          this.pollingConfig.retryCount++

          // 指数退避：60秒、120秒、240秒...
          const backoffDuration = Math.min(
            this.pollingConfig.pauseDuration * Math.pow(2, this.pollingConfig.retryCount - 1),
            300000 // 最大5分钟
          )

          // 指定时间后重新启用轮询
          setTimeout(() => {
            if (this.isPageVisible) {
              this.pollingConfig.enabled = true
              this.pollingConfig.consecutiveErrors = 0
              this.startPollingProgrammingStatus()
            }
          }, backoffDuration)

          return
        }

        // 检查是否所有题目都已完成评测
        if (allCompleted) {
          // 所有编程题都已完成评测，停止轮询
          this.stopPollingProgrammingStatus()
        }

      } catch (error) {
        // 整体请求失败（例如homeworkId不存在等）
        const statusCode = error.response ? error.response.status : 'unknown'

        // 增加连续错误计数（包括502错误）
        this.pollingConfig.consecutiveErrors++

        // 如果连续错误超过阈值，暂停轮询
        if (this.pollingConfig.consecutiveErrors >= this.pollingConfig.maxErrors) {
          this.pollingConfig.enabled = false

          // 60秒后重新启用轮询
          setTimeout(() => {
            if (this.isPageVisible) {
              this.pollingConfig.enabled = true
              this.pollingConfig.consecutiveErrors = 0
              this.startPollingProgrammingStatus()
            }
          }, 60000)
        }
      } finally {
        this.isPolling = false // 标记轮询完成
      }
    },
    // 停止轮询
    stopPollingProgrammingStatus() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
      }
      this.isPolling = false
    },
    // 处理页面可见性变化
    handleVisibilityChange() {
      if (document.hidden) {
        // 页面隐藏，暂停轮询
        this.isPageVisible = false
        this.stopPollingProgrammingStatus()
      } else {
        // 页面显示，恢复轮询（但仅在需要时）
        this.isPageVisible = true
        if (this.pollingConfig.enabled && this.homework && this.homework.questions) {
          // 使用 needsPolling() 检查是否真的需要轮询
          if (this.needsPolling()) {
            this.startPollingProgrammingStatus()
          }
        }
      }
    },
    // 图片上传成功回调
    handleUploadSuccess(response, file, fileList, questionId) {
      if (response.code === 200 && response.data && response.data.url) {
        const url = response.data.url
        // 统一转换为字符串类型作为键
        const qid = String(questionId)
        // 初始化该题目的图片数组
        if (!this.attachments[qid]) {
          this.$set(this.attachments, qid, [])
        }
        // 添加URL到数组
        this.attachments[qid].push(url)
        // 保存文件列表
        this.$set(this.uploadedImages, qid, fileList)
        // 触发自动保存
        this.handleAnswerChange()
      } else {
        this.$message.error(response.message || '图片上传失败')
      }
    },
    // 图片上传失败回调
    handleUploadError(err, file, fileList) {
      this.$message.error('图片上传失败，请重试')
    },
    // 删除图片回调
    handleRemoveFile(file, fileList, questionId) {
      // 统一转换为字符串类型作为键
      const qid = String(questionId)
      // 从attachments中移除该图片URL
      if (file.response && file.response.data && file.response.data.url) {
        const url = file.response.data.url
        if (this.attachments[qid]) {
          const index = this.attachments[qid].indexOf(url)
          if (index > -1) {
            this.attachments[qid].splice(index, 1)
          }
        }
      }
      // 更新文件列表
      this.$set(this.uploadedImages, qid, fileList)
      // 触发自动保存
      this.handleAnswerChange()
    },
    // 超出上传数量限制
    handleExceed(files, fileList) {
      this.$message.warning('最多只能上传5张图片')
    },
    // 获取某题目的图片列表（用于el-upload）
    getImageList(questionId) {
      return this.uploadedImages[questionId] || []
    },
    // 获取已提交的图片列表（用于显示）
    getSubmittedImages(questionId) {
      try {
        const qid = String(questionId)

        // 优先从 this.attachments 获取（loadSubmission 时已恢复）
        if (this.attachments && this.attachments[qid]) {
          return this.attachments[qid]
        }

        // 如果 this.attachments 没有，尝试从 this.submission.attachments 获取
        if (this.submission && this.submission.attachments) {
          const attachmentsData = JSON.parse(this.submission.attachments || '{}')
          const attachmentStr = attachmentsData[qid]
          if (attachmentStr) {
            const urls = attachmentStr.split(',').filter(url => url.trim())
            return urls
          }
        }

        return []
      } catch (e) {
        console.error('解析attachments失败:', e)
        return []
      }
    },
    // ==================== 考试模式相关方法 ====================
    // 检查考试是否已开始
    async checkExamStarted() {
      try {
        const homeworkId = this.homework.id || this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getExamStatus', homeworkId)
        if (res.code === 200 && res.data) {
          return res.data.hasStarted
        }
        return false
      } catch (error) {
        console.error('检查考试状态失败:', error)
        return false
      }
    },
    // 开始考试
    async startExam() {
      try {
        const homeworkId = this.homework.id || this.$route.params.homeworkId
        const deviceInfo = this.getDeviceInfo()
        const browserInfo = this.getBrowserInfo()

        const res = await this.$store.dispatch('classroom/startExam', {
          homeworkId,
          deviceInfo: JSON.stringify(deviceInfo),
          browserInfo: JSON.stringify(browserInfo)
        })

        if (res.code === 200) {
          this.examConfirmDialogVisible = false
          this.showExamStartDialog = false
          this.examStarted = true
          this.examStartTime = new Date(res.data.examStartTime)
          this.remainingSeconds = res.data.remainingSeconds
          this.canSubmitInExam = res.data.canSubmit || false // 修复：设置是否允许交卷

          // 如果返回了乱序的题目，更新homework.questions
          if (res.data.questions && Array.isArray(res.data.questions)) {
            // 保留原有的答案，只更新题目顺序
            const oldQuestions = this.homework.questions || []
            const newQuestions = res.data.questions

            // 创建问题ID到答案的映射
            const answerMap = new Map()
            oldQuestions.forEach(item => {
              if (item.question) {
                const qid = item.question.id
                if (this.answers[qid] !== undefined) {
                  answerMap.set(qid, this.answers[qid])
                }
                if (this.multipleAnswers[qid] !== undefined) {
                  answerMap.set(qid, this.multipleAnswers[qid])
                }
              }
            })

            // 更新题目列表
            this.homework.questions = newQuestions

            // 恢复答案
            newQuestions.forEach(item => {
              if (item.question) {
                const qid = item.question.id
                if (answerMap.has(qid)) {
                  const answer = answerMap.get(qid)
                  if (Array.isArray(answer)) {
                    this.$set(this.multipleAnswers, qid, answer)
                  } else {
                    this.$set(this.answers, qid, answer)
                  }
                }
              }
            })
          }

          // 启动计时器
          this.startExamTimer()

          // 初始化防作弊
          this.initAntiCheat()

          this.$message.success('考试开始，请认真答题')
        } else {
          this.$message.error(res.message || '开始考试失败')
        }
      } catch (error) {
        console.error('开始考试失败:', error)
        this.$message.error('开始考试失败')
      }
    },
    // 启动考试计时器
    startExamTimer() {
      if (this.examTimerInterval) {
        clearInterval(this.examTimerInterval)
      }

      this.examTimerInterval = setInterval(async () => {
        this.remainingSeconds--

        // 更新时间状态
        if (this.remainingSeconds <= 0) {
          // 超时，强制交卷
          this.timeState = 'overtime'
          this.handleOvertime()
        } else if (this.remainingSeconds <= 300) {
          // 剩余5分钟
          this.timeState = 'critical'
        } else if (this.remainingSeconds <= 600) {
          // 剩余10分钟
          this.timeState = 'warning'
        }

        // 每30秒轮询一次考试状态（包括canSubmit状态）
        if (this.remainingSeconds % 30 === 0) {
          await this.pollExamStatus()
        }
      }, 1000)
    },
    // 处理考试模式下的交卷按钮点击
    handleExamSubmit() {
      if (!this.canSubmitInExam) {
        const elapsedMinutes = Math.floor((new Date() - this.examStartTime) / 1000 / 60)
        const remainingMinutes = this.examConfig.allowSubmitAfterMinutes - elapsedMinutes
        this.$message.warning({
          message: `开考后${this.examConfig.allowSubmitAfterMinutes}分钟内不允许交卷，请再等待${remainingMinutes}分钟`,
          duration: 3000,
          showClose: true
        })
        return
      }
      this.showConfirmDialog = true
    },
    // 轮询考试状态
    async pollExamStatus() {
      try {
        const homeworkId = this.homework.id || this.$route.params.homeworkId
        const res = await this.$store.dispatch('classroom/getExamStatus', homeworkId)
        if (res.code === 200 && res.data) {
          this.remainingSeconds = res.data.remainingSeconds
          this.canSubmitInExam = res.data.canSubmit

          // 如果后端已经自动强制收卷，停止计时器并提示
          if (res.data.isSubmitted && !this.isSubmitted) {
            this.isSubmitted = true

            // 停止计时器
            if (this.examTimerInterval) {
              clearInterval(this.examTimerInterval)
              this.examTimerInterval = null
            }

            // 清理防作弊监听
            this.cleanupAntiCheat()

            // 显示提示（区分超时和老师强制收卷）
            // 只有在作业进行中才显示提示
            if (res.data.isForcedSubmit &&
                !this.hasShownForceSubmitMessage &&
                this.homework.status === 2) {
              if (this.remainingSeconds <= 0) {
                this.$message.warning({
                  message: '考试时间已到，系统已自动收卷',
                  duration: 5000,
                  showClose: true
                })
              } else {
                this.$message.warning({
                  message: '您已被老师强制收卷，请停止答题',
                  duration: 5000,
                  showClose: true
                })
              }
              this.hasShownForceSubmitMessage = true
            }

            // 刷新页面显示最终状态
            setTimeout(async () => {
              await this.loadHomeworkDetail()
            }, 1000)
          }
        }
      } catch (error) {
        console.error('轮询考试状态失败:', error)
      }
    },
    // 超时处理
    async handleOvertime() {
      this.$message.warning('考试时间已到，系统正在自动收卷...')

      // 停止计时器
      if (this.examTimerInterval) {
        clearInterval(this.examTimerInterval)
      }

      // 自动提交
      setTimeout(async () => {
        await this.confirmSubmit(true)
      }, 3000)
    },
    // 初始化防作弊
    initAntiCheat() {
      // 先清理旧的监听器，防止重复添加
      this.cleanupAntiCheat()

      this.$nextTick(() => {
        // 禁止右键
        document.addEventListener('contextmenu', this.handleContextMenu)

        // 禁止复制粘贴
        if (this.examConfig.disableCopyPaste) {
          document.addEventListener('copy', this.handleCopy)
          document.addEventListener('paste', this.handlePaste)
          document.addEventListener('cut', this.handleCut)
        }

        // 监听全屏变化
        if (this.examConfig.requireFullscreen) {
          // 初始化全屏状态
          this.lastFullscreenState = !!(document.fullscreenElement || document.webkitFullscreenElement)
          document.addEventListener('fullscreenchange', this.handleFullscreenChange)
          document.addEventListener('webkitfullscreenchange', this.handleFullscreenChange)
          // 进入全屏
          this.enterFullscreen()
          // 延迟检查全屏状态（给浏览器一点时间处理全屏请求）
          setTimeout(() => {
            this.checkFullscreenStatus()
          }, 1000)
        }

        // 监听标签页切换
        if (this.examConfig.disallowTabSwitch) {
          document.addEventListener('visibilitychange', this.handleVisibilityChange)
        }

        // 监听窗口焦点变化（检测切换到其他软件）
        window.addEventListener('blur', this.handleWindowBlur)
        window.addEventListener('focus', this.handleWindowFocus)

        // 禁用常用快捷键
        document.addEventListener('keydown', this.handleKeyDown)

        // 防止页面刷新或关闭
        window.addEventListener('beforeunload', this.handleBeforeUnload)
      })
    },
    // 清理防作弊监听
    cleanupAntiCheat() {
      document.removeEventListener('contextmenu', this.handleContextMenu)
      document.removeEventListener('copy', this.handleCopy)
      document.removeEventListener('paste', this.handlePaste)
      document.removeEventListener('cut', this.handleCut)
      document.removeEventListener('fullscreenchange', this.handleFullscreenChange)
      document.removeEventListener('webkitfullscreenchange', this.handleFullscreenChange)
      document.removeEventListener('visibilitychange', this.handleVisibilityChange)
      window.removeEventListener('blur', this.handleWindowBlur)
      window.removeEventListener('focus', this.handleWindowFocus)
      document.removeEventListener('keydown', this.handleKeyDown)
      window.removeEventListener('beforeunload', this.handleBeforeUnload)
    },
    // 禁止右键
    handleContextMenu(e) {
      // 如果已经提交，不再检测
      if (this.isSubmitted) {
        return
      }
      e.preventDefault()
      this.$message.warning('考试模式下禁止使用右键菜单')
      this.logViolation('context_menu', '尝试打开右键菜单')
    },
    // 禁止复制
    handleCopy(e) {
      // 如果已经提交，不再检测
      if (this.isSubmitted) {
        return
      }
      e.preventDefault()
      this.$message.warning('考试模式下禁止复制')
      this.logViolation('copy_attempt', '尝试复制内容')
    },
    // 禁止粘贴
    handlePaste(e) {
      // 如果已经提交，不再检测
      if (this.isSubmitted) {
        return
      }
      e.preventDefault()
      this.$message.warning('考试模式下禁止粘贴')
      this.logViolation('paste_attempt', '尝试粘贴内容')
    },
    // 禁止剪切
    handleCut(e) {
      // 如果已经提交，不再检测
      if (this.isSubmitted) {
        return
      }
      e.preventDefault()
      this.$message.warning('考试模式下禁止剪切')
      this.logViolation('cut_attempt', '尝试剪切内容')
    },
    // 进入全屏
    async enterFullscreen() {
      try {
        if (document.documentElement.requestFullscreen) {
          await document.documentElement.requestFullscreen()
        } else if (document.documentElement.webkitRequestFullscreen) {
          await document.documentElement.webkitRequestFullscreen()
        }
      } catch (error) {
        // 全屏请求被拒绝或失败，静默处理
        // checkFullscreenStatus会检查并提示用户
      }
    },
    // 检查全屏状态
    checkFullscreenStatus() {
      // 如果已经提交，不再检查全屏状态
      if (this.isSubmitted) {
        return
      }
      // 防止重复显示消息
      if (this.fullscreenWarned) {
        return
      }
      const isFullscreen = document.fullscreenElement || document.webkitFullscreenElement
      if (!isFullscreen && this.examConfig.requireFullscreen) {
        this.fullscreenWarned = true
        this.$message.warning({
          message: '请进入全屏模式参加考试（按F11或点击页面）',
          duration: 5000,
          showClose: true,
          onClose: () => {
            // 用户关闭提示后，重置标志，允许再次提示
            this.fullscreenWarned = false
            // 用户关闭提示后，再次尝试进入全屏
            this.enterFullscreen()
          }
        })
      }
    },
    // 全屏变化监听
    handleFullscreenChange() {
      // 如果已经提交，不再检测全屏变化
      if (this.isSubmitted) {
        return
      }

      const currentFullscreenState = !!(document.fullscreenElement || document.webkitFullscreenElement)

      // 检查全屏状态是否真的发生了变化
      if (currentFullscreenState === this.lastFullscreenState) {
        // 状态没有变化，忽略这个事件（避免重复触发）
        return
      }

      // 设置标志：正在处理全屏变化
      this.isFullscreenChanging = true
      this.fullscreenChangeTime = Date.now()

      if (!currentFullscreenState && this.examConfig.requireFullscreen) {
        // 从全屏变为非全屏：记录违规
        this.logViolation('fullscreen_exit', '退出全屏')

        // 只警告一次，不强制弹窗或自动进入全屏
        if (!this.fullscreenWarned) {
          this.fullscreenWarned = true
          this.$message.warning({
            message: '检测到退出全屏，请立即返回全屏！按 F11 或在页面右键选择"进入全屏"',
            duration: 5000,
            showClose: true
          })
        }
      } else if (currentFullscreenState) {
        // 从非全屏变为全屏：重置警告标志
        this.fullscreenWarned = false
      }

      // 更新上一次的全屏状态
      this.lastFullscreenState = currentFullscreenState

      // 2秒后重置标志，允许正常的blur和visibilitychange检测
      setTimeout(() => {
        this.isFullscreenChanging = false
      }, 2000)
    },
    // 标签页切换监听
    handleVisibilityChange() {
      // 如果已经提交，不再检测标签页切换
      if (this.isSubmitted) {
        return
      }

      // 检查当前全屏状态
      const isFullscreen = !!(document.fullscreenElement || document.webkitFullscreenElement)

      // 如果当前在全屏状态，但页面隐藏了（可能是用户切换到了其他标签页）
      if (document.hidden) {
        if (isFullscreen && this.examConfig.disallowTabSwitch) {
          this.$message.warning('检测到切换标签页，请专注于考试！')
          this.logViolation('tab_switch', '切换标签页')
        } else if (!isFullscreen && this.examConfig.requireFullscreen) {
          // 如果不在全屏状态且页面隐藏了，说明用户可能退出了全屏并切换到了其他应用
          this.$message.warning('检测到退出全屏，请立即返回考试界面！')
          this.logViolation('fullscreen_exit', '退出全屏')
        }
      }
    },
    // 窗口失去焦点
    handleWindowBlur() {
      // 如果已经提交，不再检测窗口焦点变化
      if (this.isSubmitted) {
        return
      }

      // 更新窗口焦点状态
      this.isWindowFocused = false

      // 检查当前全屏状态
      const isFullscreen = !!(document.fullscreenElement || document.webkitFullscreenElement)

      // 窗口失焦：点击了浏览器窗口以外的界面
      // 如果页面还是可见的，但窗口失焦了，说明是点击了其他软件
      if (!document.hidden && isFullscreen) {
        this.$notify({
          title: '警告',
          message: '检测到切换到其他窗口，请专注于考试！',
          type: 'warning',
          duration: 3000
        })
        this.logViolation('window_blur', '切换到其他软件')
      }
    },
    // 窗口获得焦点
    handleWindowFocus() {
      // 如果已经提交，不再检查
      if (this.isSubmitted) {
        return
      }

      // 更新窗口焦点状态
      this.isWindowFocused = true

      if (this.examConfig.requireFullscreen) {
        // 延迟检查全屏状态
        setTimeout(() => {
          this.checkFullscreenStatus()
        }, 500)
      }
    },
    // 禁用快捷键
    handleKeyDown(e) {
      // 禁用 Ctrl+C, Ctrl+V, Ctrl+X
      if ((e.ctrlKey || e.metaKey) && ['c', 'v', 'x'].includes(e.key.toLowerCase())) {
        if (this.examConfig.disableCopyPaste) {
          e.preventDefault()
          return
        }
      }

      // 禁用 F12
      if (e.key === 'F12') {
        e.preventDefault()
        this.logViolation('devtools_attempt', '尝试打开开发者工具')
        return
      }

      // 禁用 Ctrl+Shift+I
      if (e.ctrlKey && e.shiftKey && e.key === 'I') {
        e.preventDefault()
        this.logViolation('devtools_attempt', '尝试打开开发者工具')
        return
      }

      // 禁用 Ctrl+U
      if (e.ctrlKey && e.key === 'u') {
        e.preventDefault()
        return
      }

      // 禁用 Ctrl+S
      if (e.ctrlKey && e.key === 's') {
        e.preventDefault()
        return
      }

      // 禁用 Esc（如果要求全屏）
      if (e.key === 'Escape' && this.examConfig.requireFullscreen) {
        e.preventDefault()
        this.$message.warning('考试期间禁止退出全屏')
        return
      }
    },
    // 记录违规（带重试机制）
    async logViolation(type, description) {
      const homeworkId = this.homework.id || parseInt(this.$route.params.homeworkId)
      const maxRetries = 3
      let lastError = null

      for (let i = 0; i < maxRetries; i++) {
        try {
          await this.$store.dispatch('classroom/logViolation', {
            homeworkId,
            violationType: type,
            description
          })
          // 上报成功，不再重试
          return
        } catch (error) {
          lastError = error
          // 等待一小段时间后重试（指数退避）
          if (i < maxRetries - 1) {
            await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 1000))
          }
        }
      }

      // 所有重试都失败，记录错误
      console.error('违规记录上报最终失败:', type, description, lastError.message)
    },
    // 获取设备信息
    getDeviceInfo() {
      return {
        userAgent: navigator.userAgent,
        platform: navigator.platform,
        screen: {
          width: screen.width,
          height: screen.height
        },
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
      }
    },
    // 获取浏览器信息
    getBrowserInfo() {
      const ua = navigator.userAgent
      let browser = 'Unknown'
      if (ua.includes('Chrome')) browser = 'Chrome'
      else if (ua.includes('Firefox')) browser = 'Firefox'
      else if (ua.includes('Safari')) browser = 'Safari'
      else if (ua.includes('Edge')) browser = 'Edge'

      return {
        name: browser,
        version: navigator.appVersion,
        language: navigator.language
      }
    },
    // 防止页面刷新或关闭
    handleBeforeUnload(e) {
      if (this.isExamMode && this.examStarted && !this.isSubmitted) {
        e.preventDefault()
        e.returnValue = '考试正在进行中，离开会导致答案丢失！确定要离开吗？'
        return e.returnValue
      }
    },
    // ==================== 考试模式题目导航 ====================
    // 获取题目导航样式类
    getQuestionNavClass(question) {
      const classes = []

      // 判断是否已答题
      let isAnswered = false
      if (question.problemId) {
        // 编程题
        isAnswered = this.programmingStatus[question.problemId] === 'submitted'
      } else if (question.question) {
        // 普通题目
        const qid = question.question.id
        if (question.question.type === 'multiple_choice') {
          isAnswered = this.multipleAnswers[qid] && this.multipleAnswers[qid].length > 0
        } else if (question.question.type === 'composite') {
          const subQuestions = this.parseCompositeSubQuestions(question.question.options)
          const subAnswers = this.compositeAnswers[qid] || {}
          isAnswered = subQuestions.length > 0 && subQuestions.every(sub => !!subAnswers[sub.id])
        } else {
          isAnswered = this.answers[qid] && this.answers[qid].toString().trim() !== ''
        }
      }

      classes.push(isAnswered ? 'answered' : 'unanswered')

      return classes.join(' ')
    },
    // 判断题目是否是提交后新增的（未提交）
    isQuestionUnsubmitted(question) {
      if (!this.isSubmitted) {
        return false
      }

      if (question.problemId) {
        // 编程题：检查是否在已提交的编程题列表中
        return !this.submittedQuestionIds.has(`programming_${question.problemId}`)
      } else if (question.question) {
        // 普通题目：检查是否在已提交的题目列表中
        const qid = question.question.id.toString()
        return !this.submittedQuestionIds.has(qid)
      }

      return false
    },
    // 滚动到指定题目
    scrollToQuestion(index) {
      const questions = document.querySelectorAll('.question-item')
      if (questions && questions[index]) {
        questions[index].scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    }
  },
  // 路由导航守卫：离开考试页面时给出警告
  beforeRouteLeave(to, from, next) {
    // 考试模式下且未提交时，给出警告但允许离开
    if (this.isExamMode && this.examStarted && !this.isSubmitted) {
      const answer = confirm('⚠️ 警告：考试正在进行中！\n\n离开后您可以再次进入继续考试，但请确保不会超时。确定要离开吗？')
      if (answer) {
        // 用户确认离开，不清理资源（组件销毁时会自动清理）
        next()
      } else {
        // 用户取消，阻止导航
        next(false)
      }
    } else {
      next()
    }
  }
}
</script>

<style scoped>
.homework-detail {
  padding: 20px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.questions-container {
  margin: 20px 0;
}
.question-item {
  margin-bottom: 30px;
  padding: 20px;
  border: 1px solid #eee;
  border-radius: 8px;
  background: #fafafa;
}
.question-header {
  margin-bottom: 15px;
  font-weight: bold;
}
.question-number {
  font-size: 18px;
  color: #409EFF;
}
.question-type {
  color: #909399;
  margin: 0 10px;
  font-size: 14px;
}
.question-score {
  color: #67C23A;
  font-size: 14px;
}
.question-score-earned {
  color: #409EFF;
  font-size: 14px;
  margin-left: auto;
}
.question-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 10px;
}
.question-content {
  margin-bottom: 15px;
  line-height: 1.6;
}
.question-options {
  margin-top: 15px;
}
.question-options /deep/ .el-radio,
.question-options /deep/ .el-checkbox {
  display: flex;
  align-items: flex-start;
  white-space: normal;
  line-height: 1.7;
  height: auto;
}
.question-options /deep/ .el-radio__label,
.question-options /deep/ .el-checkbox__label {
  display: block;
  flex: 1;
  width: 100%;
  white-space: normal;
  line-height: 1.7;
  padding-left: 0;
}
.question-options /deep/ .el-radio__input,
.question-options /deep/ .el-checkbox__input {
  margin-top: 3px;
}
.option-item {
  margin: 10px 0;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
  /* 移除 transition 避免轮询时闪烁 */
}
.option-item-clickable {
  cursor: pointer;
}
.option-item:hover {
  background: #f5f7fa;
}
.option-rich-text {
  display: block;
  width: 100%;
}
.option-letter {
  font-weight: 600;
  color: #606266;
}
.option-head {
  display: block;
  margin-bottom: 6px;
}
.option-content {
  display: block;
  word-break: break-word;
  overflow-wrap: anywhere;
}
.question-options /deep/ .option-content p {
  margin: 0;
}
.question-options /deep/ .option-content pre {
  margin: 0;
  max-width: 100%;
  overflow-x: auto;
}
.student-answer {
  margin-top: 15px;
  padding: 10px;
  background: #e6f7ff;
  border-left: 3px solid #409EFF;
  border-radius: 4px;
}
.action-area {
  text-align: center;
  padding: 20px 0;
}
.submit-area {
  text-align: center;
  padding: 20px 0;
}
.subjective-answer {
  margin-top: 15px;
}
.composite-answer-area {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.composite-sub-question-card {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  background: #fff;
}
.composite-sub-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
}
.programming-answer {
  margin-top: 15px;
  text-align: center;
  padding: 20px;
  background: #f0f9ff;
  border-radius: 4px;
}
</style>

<!-- 非scoped样式，确保markdown-body样式生效 -->
<style>
.homework-detail .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.homework-detail .markdown-body h1,
.homework-detail .markdown-body h2,
.homework-detail .markdown-body h3,
.homework-detail .markdown-body h4,
.homework-detail .markdown-body h5,
.homework-detail .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.homework-detail .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.homework-detail .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.homework-detail .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.homework-detail .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.homework-detail .markdown-body p {
  margin-bottom: 16px !important;
}

.homework-detail .markdown-body strong {
  font-weight: bold !important;
}

.homework-detail .markdown-body em {
  font-style: italic !important;
}

.homework-detail .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.homework-detail .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

/* 答案显示样式 */
.correct-answer {
  margin-top: 15px;
  padding: 10px;
  background: #f0f9ff;
  border-left: 3px solid #67C23A;
  border-radius: 4px;
}

.reference-answer {
  margin-top: 15px;
  padding: 15px;
  background: #f5f7fa;
  border-left: 3px solid #409EFF;
  border-radius: 4px;
}

.reference-answer-title {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
  font-size: 14px;
}

.reference-answer-content {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  line-height: 1.8;
}

.question-analysis {
  margin-top: 15px;
  padding: 15px;
  background: #fff9e6;
  border-left: 3px solid #E6A23C;
  border-radius: 4px;
}

.analysis-title {
  font-weight: bold;
  color: #E6A23C;
  margin-bottom: 10px;
  font-size: 14px;
}

.analysis-content {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  line-height: 1.8;
}

.reference-answer-empty {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  color: #909399;
  font-style: italic;
}

/* 图片上传样式 */
.image-upload-area {
  margin-top: 15px;
  padding: 15px;
  background: #f9f9f9;
  border-radius: 4px;
  border: 1px dashed #d9d9d9;
}

.upload-tip {
  margin-bottom: 10px;
  color: #606266;
  font-size: 14px;
}

.upload-limit-tip {
  margin-top: 10px;
  color: #909399;
  font-size: 12px;
}

.hide-upload-btn .el-upload--picture-card {
  display: none;
}

.submitted-images {
  margin-top: 15px;
  padding: 15px;
  background: #f0f9ff;
  border-radius: 4px;
  border: 1px solid #b3d8ff;
}

.submitted-images-title {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
  font-size: 14px;
}

.submitted-images-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

/* ==================== 考试模式样式 ==================== */
/* 考试模式确认区域（内联显示） */
.exam-confirm-inline {
  padding: 30px;
  background: white;
  border-radius: 12px;
  border: 2px solid #E6A23C;
  margin-bottom: 30px;
  box-shadow: 0 4px 12px rgba(230, 162, 60, 0.2);
}

.exam-confirm-header {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.exam-confirm-header h2 {
  margin: 0;
  color: #E6A23C;
  font-size: 24px;
  font-weight: bold;
}

.exam-rules-inline {
  margin: 20px 0;
  padding: 20px;
  background: white;
  border-radius: 8px;
  border: 1px solid #E6A23C;
}

.exam-rules-inline h3 {
  color: #333;
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: bold;
  padding-bottom: 8px;
  border-bottom: 2px solid #E6A23C;
}

.exam-rules-inline ul {
  list-style: none;
  padding: 0;
  margin: 10px 0;
}

.exam-rules-inline li {
  padding: 8px 0;
  line-height: 1.8;
  color: #606266;
}

.exam-rules-inline li.warning {
  color: #F56C6C;
  font-weight: bold;
}

.exam-start-actions {
  display: flex;
  justify-content: center;
  gap: 15px;
  margin-top: 25px;
  padding-top: 20px;
  border-top: 1px solid #E6A23C;
}

.checkbox-group {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 15px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 8px;
}

/* 考试模式顶部标识栏 */
.exam-header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: linear-gradient(135deg, #FF6B35 0%, #F7931E 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30px;
  box-shadow: 0 2px 8px rgba(255, 107, 53, 0.3);
  z-index: 1000;
  animation: pulse-border 2s infinite;
}

@keyframes pulse-border {
  0%, 100% { box-shadow: 0 2px 8px rgba(255, 107, 53, 0.3); }
  50% { box-shadow: 0 2px 12px rgba(255, 107, 53, 0.6); }
}

.exam-badge {
  display: flex;
  align-items: center;
  font-size: 18px;
  font-weight: bold;
}

.exam-badge i {
  margin-right: 8px;
  font-size: 24px;
}

/* 考试计时器 */
.exam-timer {
  display: flex;
  align-items: center;
  font-size: 20px;
  font-weight: bold;
  padding: 8px 20px;
  border-radius: 20px;
  color: white;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

/* 考试模式下页面边框 */
.homework-detail {
  transition: all 0.3s ease;
}

.homework-detail >>> .el-card {
  margin-top: 80px; /* 为顶部栏留出空间 */
}

/* ==================== 考试模式题目导航 ==================== */
.question-navigator {
  position: fixed;
  right: 20px;
  bottom: 80px;
  max-width: 280px;
  width: auto;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  padding: 15px;
  z-index: 998;
  animation: slideInFromBottom 0.3s ease-out;
}

@keyframes slideInFromBottom {
  from {
    transform: translateY(100px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.navigator-timer {
  text-align: center;
  padding: 10px;
  margin-bottom: 15px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: white;
  white-space: nowrap;
}

.navigator-title {
  font-size: 14px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 10px;
  text-align: center;
}

.navigator-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
  justify-content: center;
}

.question-nav-item {
  min-width: 36px;
  width: auto;
  height: 36px;
  padding: 0 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #DCDFE6;
}

.question-nav-item.unanswered {
  background: #F5F7FA;
  color: #909399;
}

.question-nav-item.answered {
  background: #67C23A;
  color: white;
  border-color: #67C23A;
}

.question-nav-item:hover {
  transform: scale(1.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.navigator-legend {
  display: flex;
  justify-content: center;
  gap: 12px;
  font-size: 11px;
  color: #606266;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  border: 1px solid #DCDFE6;
}

.legend-color.unanswered {
  background: #F5F7FA;
}

.legend-color.answered {
  background: #67C23A;
  border-color: #67C23A;
}

/* 考试模式右下角计时器 */
.exam-corner-timer {
  position: fixed;
  right: 20px;
  bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 50px;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  font-size: 16px;
  font-weight: 600;
  z-index: 999;
  animation: slideIn 0.3s ease-out;
}

.exam-corner-timer i {
  font-size: 18px;
}

.exam-corner-timer .timer-label {
  font-size: 14px;
  opacity: 0.9;
}

.exam-corner-timer .timer-value {
  font-size: 18px;
  font-weight: 700;
  font-family: 'Monaco', 'Consolas', monospace;
}

@keyframes slideIn {
  from {
    transform: translateY(100px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* 学生端专用样式：代码显示 - 只影响学生端 */
.student-homework-page .code-display-wrapper .markdown-body pre {
  padding: 0 16px 0 40px !important;  /* 左侧40px给行号留空间 */
  position: relative !important;
}

.student-homework-page .code-display-wrapper .markdown-body pre code {
  padding: 0px 16px 0px 0px !important;  /* code不添加额外缩进，总缩进保持40px */
  line-height: 26px !important;
}

.student-homework-page .code-display-wrapper .markdown-body pre ol.pre-numbering {
  line-height: 26px !important;
  font-size: 1rem !important;
}

.student-homework-page .code-display-wrapper .markdown-body pre ol.pre-numbering li {
  line-height: 26px !important;
  margin: 0 !important;
  padding: 0 !important;
}

.student-homework-page .code-display-wrapper .markdown-body pre ol.pre-numbering li:before {
  width: 40px !important;
  font-size: 1rem !important;
  line-height: 26px !important;
  vertical-align: top !important;
}
</style>
