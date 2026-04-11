<template>
  <div class="question-bank-panel">
    <div class="page-header">
      <el-button icon="el-icon-arrow-left" @click="goBack">{{ $t('m.Back') }}</el-button>
      <h3>{{ $t('m.Question_Bank') }}</h3>
    </div>
    <div class="action-bar">
      <el-button type="primary" icon="el-icon-plus" @click="goCreatePage">
        {{ $t('m.Create_Question') }}
      </el-button>
    </div>

    <!-- 筛选条件 -->
    <div class="filter-bar" style="margin-bottom: 20px;">
      <el-row :gutter="15">
        <el-col :span="4">
          <el-select v-model="filters.type" placeholder="题型筛选" clearable size="small" @change="handleFilterChange">
            <el-option label="全部题型" value=""></el-option>
            <el-option label="单选题" value="single_choice"></el-option>
            <el-option label="多选题" value="multiple_choice"></el-option>
            <el-option label="判断题" value="judge"></el-option>
            <el-option label="填空题" value="fill_blank"></el-option>
            <el-option label="主观题" value="subjective"></el-option>
            <el-option label="组合题" value="composite"></el-option>
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.course" placeholder="课程筛选" clearable size="small" filterable @change="handleFilterChange">
            <el-option label="全部课程" value=""></el-option>
            <el-option
              v-for="course in commonCourses"
              :key="course"
              :label="course"
              :value="course"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-input
            v-model="filters.tag"
            placeholder="标签筛选"
            clearable
            size="small"
            @keyup.enter.native="handleFilterChange"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.difficulty" placeholder="难度筛选" clearable size="small" @change="handleFilterChange">
            <el-option label="全部难度" value=""></el-option>
            <el-option label="简单" value="1"></el-option>
            <el-option label="中等" value="2"></el-option>
            <el-option label="困难" value="3"></el-option>
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-input
            v-model="filters.questionId"
            placeholder="搜索题目ID"
            clearable
            size="small"
            @keyup.enter.native="handleFilterChange"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索标题"
            clearable
            size="small"
            @keyup.enter.native="handleFilterChange"
          >
            <el-button slot="append" icon="el-icon-search" @click="handleFilterChange"></el-button>
          </el-input>
        </el-col>
      </el-row>
    </div>

    <!-- 移除 v-loading 避免轮询时闪烁 -->
    <el-table
      :data="questions"
      class="question-bank-table"
      :row-class-name="getQuestionRowClass"
      stripe
    >
      <el-table-column prop="title" :label="$t('m.Question_Title')" min-width="520">
        <template slot-scope="{ row }">
          <div class="inline-question-cell">
            <div class="inline-meta-row">
              <el-tag type="info" size="mini">ID:{{ row.id }}</el-tag>
              <el-tag :type="getQuestionTypeColor(row.type)" size="mini">{{ getQuestionTypeName(row.type) }}</el-tag>
              <el-tag type="danger" size="mini">难度：{{ getDifficultyText(row.difficulty) }}</el-tag>
              <el-tag type="primary" size="mini">分值：{{ Number(row.score || 0) }}分</el-tag>
              <el-tag type="info" size="mini">创建者：{{ row.creator ? row.creator.username : (row.creatorId || '-') }}</el-tag>
              <el-tag type="info" size="mini">创建时间：{{ formatTime(row.createTime || row.createdAt) }}</el-tag>
              <el-tag v-if="row.course" type="warning" size="mini">所属课程：{{ row.course }}</el-tag>
              <el-tag
                v-for="(tag, idx) in parseQuestionTags(row.tags)"
                :key="`meta-tag-${row.id}-${idx}`"
                size="mini"
                type="info"
              >
                标签：{{ tag }}
              </el-tag>
              <el-tag :type="row.isShared ? 'success' : 'info'" size="mini">
                开放权限：{{ row.isShared ? '共享' : '个人' }}
              </el-tag>
            </div>
            <div v-html="renderMarkdown(row.title)" class="markdown-body inline-question-title" v-highlight></div>
            <div
              v-if="row.content"
              v-html="renderMarkdown(row.content)"
              class="markdown-body inline-question-content"
              v-highlight
            ></div>

            <div v-if="row.type === 'single_choice' || row.type === 'multiple_choice'" class="inline-options-list">
              <div
                v-for="(option, index) in parseOptionEntries(row.options)"
                :key="`opt-${row.id}-${index}`"
                class="inline-option-item"
              >
                <span class="inline-option-label">{{ option.letter }}.</span>
                <span class="markdown-body inline-option-text" v-html="renderMarkdown(option.text)" v-highlight></span>
              </div>
            </div>

            <div v-else-if="row.type === 'judge'" class="inline-options-list">
              <div class="inline-option-item">
                <span class="inline-option-label">A.</span>
                <span class="inline-option-text">正确</span>
              </div>
              <div class="inline-option-item">
                <span class="inline-option-label">B.</span>
                <span class="inline-option-text">错误</span>
              </div>
            </div>

            <div v-else-if="row.type === 'composite'" class="inline-composite-list">
              <div
                v-for="(sub, subIndex) in parseCompositeSubQuestions(row.options)"
                :key="`sub-${row.id}-${sub.id || subIndex}`"
                class="inline-composite-item"
              >
                <div class="inline-composite-head">子题 {{ subIndex + 1 }}（{{ Number(sub.score || 0) }}分）</div>
                <div
                  class="markdown-body inline-composite-content"
                  v-html="renderMarkdown(sub.content || '')"
                  v-highlight
                ></div>
                <div class="inline-options-list">
                  <div
                    v-for="(option, optionIndex) in parseOptionEntries(sub.options)"
                    :key="`subopt-${row.id}-${subIndex}-${optionIndex}`"
                    class="inline-option-item"
                  >
                    <span class="inline-option-label">{{ option.letter }}.</span>
                    <span class="markdown-body inline-option-text" v-html="renderMarkdown(option.text)" v-highlight></span>
                  </div>
                </div>
                <div class="answer-info compact-answer-info">
                  <strong>正确答案：</strong>
                  <el-tag type="success">{{ getCompositeAnswerBySubQuestion(row.answer, sub.id, subIndex) || '-' }}</el-tag>
                </div>
              </div>
            </div>

            <div v-if="row.type !== 'composite'" class="inline-answer-row answer-info compact-answer-info">
              <span class="inline-answer-label">{{ row.type === 'subjective' ? '参考答案：' : '正确答案：' }}</span>
              <span v-if="row.type !== 'subjective'" class="inline-answer-text">{{ formatInlineAnswer(row) }}</span>
              <span
                v-else-if="row.answer"
                class="markdown-body inline-answer-text"
                v-html="renderMarkdown(row.answer)"
                v-highlight
              ></span>
              <span v-else class="inline-answer-text">暂无答案</span>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('m.Operation')" width="280">
        <template slot-scope="{ row }">
          <el-button size="small" type="info" @click="handleViewDetail(row)">
            {{ $t('m.View_Detail') || '查看详情' }}
          </el-button>
          <el-button size="small" @click="goEditPage(row)">{{ $t('m.Edit') }}</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">
            {{ $t('m.Delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pagination-container">
      <el-pagination
        @size-change="handlePageSizeChange"
        @current-change="handleCurrentPageChange"
        :current-page="pagination.currentPage"
        :page-sizes="[10, 20, 50, 100]"
        :page-size="pagination.pageSize"
        layout="total, sizes, prev, pager, next, jumper"
        :total="pagination.total"
      >
      </el-pagination>
    </div>

    <el-dialog :title="$t('m.Create_Question')" :visible.sync="showCreateDialog" width="1480px" class="question-edit-dialog">
      <el-row :gutter="20">
        <el-col :span="14" class="question-form-column">
          <el-form :model="createForm" ref="createForm" label-width="110px" class="question-form">
            <el-form-item :label="$t('m.Question_Type')" prop="type">
              <el-select v-model="createForm.type" @change="handleTypeChange">
                <el-option :label="$t('m.Single_Choice')" value="single_choice" />
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice" />
                <el-option :label="$t('m.Judge')" value="judge" />
                <el-option :label="$t('m.Subjective')" value="subjective" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('m.Question_Title')" prop="title">
              <el-input v-model="createForm.title" />
            </el-form-item>
            <el-form-item :label="$t('m.Content')" prop="content">
              <el-input type="textarea" v-model="createForm.content" :rows="7" />
            </el-form-item>

        <!-- 单选题：固定4个选项，单选 -->
        <template v-if="createForm.type === 'single_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="option-item">
                <el-radio v-model="createForm.correctAnswer" :label="index" class="option-radio">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-radio>
                <el-input v-model="createForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
            <div class="answer-tip">
              <i class="el-icon-info"></i>
              {{ $t('m.Select_Correct_Answer_Tip') }}
            </div>
          </el-form-item>
        </template>

        <!-- 多选题：固定4个选项，多选 -->
        <template v-if="createForm.type === 'multiple_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="option-item">
                <el-checkbox v-model="createForm.correctAnswers[index]" :label="index" class="option-checkbox">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-checkbox>
                <el-input v-model="createForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
            <div class="answer-tip">
              <i class="el-icon-info"></i>
              {{ $t('m.Select_Correct_Answers_Tip') }}
            </div>
          </el-form-item>
        </template>

        <!-- 判断题：选择正确/错误 -->
        <template v-if="createForm.type === 'judge'">
          <el-form-item :label="$t('m.Correct_Answer')" required>
            <el-radio-group v-model="createForm.correctAnswer">
              <el-radio label="true">{{ $t('m.True') }}</el-radio>
              <el-radio label="false">{{ $t('m.False') }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>

        <!-- 主观题：需要人工打分 -->
        <template v-if="createForm.type === 'subjective'">
          <el-form-item :label="$t('m.Reference_Answer')">
            <el-input type="textarea" v-model="createForm.referenceAnswer" :rows="3" :placeholder="$t('m.Reference_Answer_Tip')" />
          </el-form-item>
        </template>

        <!-- 题目解析 -->
        <el-form-item label="题目解析">
          <el-input type="textarea" v-model="createForm.analysis" :rows="2" placeholder="请输入题目解析（可选）" />
        </el-form-item>

        <!-- 题目标签 -->
        <el-form-item label="题目标签">
          <div class="tags-input-container">
            <div class="tags-list">
              <el-tag
                v-for="(tag, index) in createForm.tags"
                :key="index"
                closable
                @close="removeCreateTag(index)"
                style="margin-right: 5px; margin-bottom: 5px;"
              >
                {{ tag }}
              </el-tag>
            </div>
            <el-input
              v-model="createTagInput"
              placeholder="输入标签名称，按回车添加"
              @keyup.enter.native="addCreateTag"
              style="width: 100%;"
            />
          </div>
          <div class="form-tip">
            <i class="el-icon-info"></i>
            输入标签名称后按回车添加，可添加多个标签
          </div>
        </el-form-item>

        <el-row :gutter="12" class="compact-form-row">
          <el-col :span="12">
            <el-form-item label="所属课程">
              <el-select
                v-model="createForm.course"
                placeholder="请选择课程"
                style="width: 100%"
              >
                <el-option
                  v-for="course in commonCourses"
                  :key="course"
                  :label="course"
                  :value="course"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('m.Difficulty')" prop="difficulty">
              <el-rate v-model="createForm.difficulty" :max="3" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12" class="compact-form-row">
          <el-col :span="12">
            <el-form-item :label="$t('m.Score')" prop="score">
              <el-input-number v-model="createForm.score" :min="1" :max="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('m.Share_To_Question_Pool')">
              <el-switch v-model="createForm.isShared" />
              <div class="form-tip compact-tip">
                <i class="el-icon-info"></i>
                {{ $t('m.Share_To_Question_Pool_Tip') }}
              </div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
        </el-col>
        <el-col :span="10" class="preview-column">
          <el-card class="preview-card markdown-preview">
            <div slot="header">
              <i class="el-icon-view"></i> 实时预览
            </div>
            <div class="preview-content">
              <div v-if="createForm.title" v-html="renderMarkdown(createForm.title)" class="markdown-body preview-title" v-highlight></div>
              <p v-else class="preview-placeholder">题目标题预览</p>

              <div v-if="createForm.content" v-html="renderMarkdown(createForm.content)" class="markdown-body preview-content-text" v-highlight></div>
              <p v-else class="preview-placeholder">题目内容预览</p>

              <!-- 单选题选项预览 -->
              <!-- 单选题选项预览 -->
              <div v-if="createForm.type === 'single_choice'" class="preview-options">
                <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="createForm.correctAnswer === index ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body" v-highlight></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 多选题选项预览 -->
              <div v-if="createForm.type === 'multiple_choice'" class="preview-options">
                <div v-for="(option, index) in createForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="createForm.correctAnswers[index] ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body" v-highlight></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 判断题预览 -->
              <div v-if="createForm.type === 'judge'" class="preview-options">
                <div class="preview-option-item">
                  <el-tag :type="createForm.correctAnswer === 'true' ? 'success' : 'info'" size="small">✓</el-tag>
                  <span>{{ $t('m.True') }}</span>
                </div>
                <div class="preview-option-item">
                  <el-tag :type="createForm.correctAnswer === 'false' ? 'success' : 'info'" size="small">✗</el-tag>
                  <span>{{ $t('m.False') }}</span>
                </div>
              </div>

              <!-- 主观题预览 -->
              <div v-if="createForm.type === 'subjective'" class="preview-subjective">
                <el-alert type="info" :closable="false">
                  <i class="el-icon-edit"></i> 主观题，学生需要输入文字答案
                </el-alert>
              </div>

              <!-- 题目解析预览 -->
              <div v-if="createForm.analysis" class="preview-analysis">
                <el-divider content-position="left">
                  <i class="el-icon-document" style="color: #E6A23C;"></i>
                  <span style="color: #E6A23C; font-weight: bold;">题目解析</span>
                </el-divider>
                <div v-html="renderMarkdown(createForm.analysis)" class="markdown-body preview-analysis-content" v-highlight></div>
              </div>
              <p v-else class="preview-placeholder" style="margin-top: 15px;">题目解析预览</p>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <span slot="footer">
        <el-button @click="showCreateDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="createQuestion">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 编辑题目对话框 -->
    <el-dialog :title="$t('m.Edit_Question')" :visible.sync="showEditDialog" width="1480px" class="question-edit-dialog">
      <el-row :gutter="20">
        <el-col :span="14" class="question-form-column">
          <el-form :model="editForm" ref="editForm" label-width="110px" class="question-form">
            <el-form-item :label="$t('m.Question_Type')" prop="type">
              <el-select v-model="editForm.type" @change="handleEditTypeChange">
                <el-option :label="$t('m.Single_Choice')" value="single_choice" />
                <el-option :label="$t('m.Multiple_Choice')" value="multiple_choice" />
                <el-option :label="$t('m.Judge')" value="judge" />
                <el-option :label="$t('m.Subjective')" value="subjective" />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('m.Question_Title')" prop="title">
              <el-input v-model="editForm.title" />
            </el-form-item>
            <el-form-item :label="$t('m.Content')" prop="content">
              <el-input type="textarea" v-model="editForm.content" :rows="7" />
            </el-form-item>

        <!-- 单选题 -->
        <template v-if="editForm.type === 'single_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="option-item">
                <el-radio v-model="editForm.correctAnswer" :label="index" class="option-radio">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-radio>
                <el-input v-model="editForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
          </el-form-item>
        </template>

        <!-- 多选题 -->
        <template v-if="editForm.type === 'multiple_choice'">
          <el-form-item :label="$t('m.Options')" required>
            <div class="options-container">
              <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="option-item">
                <el-checkbox v-model="editForm.correctAnswers[index]" :label="index" class="option-checkbox">
                  {{ ['A', 'B', 'C', 'D'][index] }}
                </el-checkbox>
                <el-input v-model="editForm.choiceOptions[index]" :placeholder="`${['A', 'B', 'C', 'D'][index]}. ${$t('m.Option_Content')}`" />
              </div>
            </div>
          </el-form-item>
        </template>

        <!-- 判断题 -->
        <template v-if="editForm.type === 'judge'">
          <el-form-item :label="$t('m.Correct_Answer')" required>
            <el-radio-group v-model="editForm.correctAnswer">
              <el-radio label="true">{{ $t('m.True') }}</el-radio>
              <el-radio label="false">{{ $t('m.False') }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>

        <!-- 主观题 -->
        <template v-if="editForm.type === 'subjective'">
          <el-form-item :label="$t('m.Reference_Answer')">
            <el-input type="textarea" v-model="editForm.referenceAnswer" :rows="3" :placeholder="$t('m.Reference_Answer_Tip')" />
          </el-form-item>
        </template>

        <!-- 题目解析 -->
        <el-form-item label="题目解析">
          <el-input type="textarea" v-model="editForm.analysis" :rows="2" placeholder="请输入题目解析（可选）" />
        </el-form-item>

        <!-- 题目标签 -->
        <el-form-item label="题目标签">
          <div class="tags-input-container">
            <div class="tags-list">
              <el-tag
                v-for="(tag, index) in editForm.tags"
                :key="index"
                closable
                @close="removeEditTag(index)"
                style="margin-right: 5px; margin-bottom: 5px;"
              >
                {{ tag }}
              </el-tag>
            </div>
            <el-input
              v-model="editTagInput"
              placeholder="输入标签名称，按回车添加"
              @keyup.enter.native="addEditTag"
              style="width: 100%;"
            />
          </div>
          <div class="form-tip">
            <i class="el-icon-info"></i>
            输入标签名称后按回车添加，可添加多个标签
          </div>
        </el-form-item>

        <el-row :gutter="12" class="compact-form-row">
          <el-col :span="12">
            <el-form-item label="所属课程">
              <el-select
                v-model="editForm.course"
                placeholder="请选择课程"
                style="width: 100%"
              >
                <el-option
                  v-for="course in commonCourses"
                  :key="course"
                  :label="course"
                  :value="course"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('m.Difficulty')" prop="difficulty">
              <el-rate v-model="editForm.difficulty" :max="3" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12" class="compact-form-row">
          <el-col :span="12">
            <el-form-item :label="$t('m.Score')" prop="score">
              <el-input-number v-model="editForm.score" :min="1" :max="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('m.Share_To_Question_Pool')">
              <el-switch v-model="editForm.isShared" />
              <div class="form-tip compact-tip">
                <i class="el-icon-info"></i>
                {{ $t('m.Share_To_Question_Pool_Tip') }}
              </div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
        </el-col>
        <el-col :span="10" class="preview-column">
          <el-card class="preview-card markdown-preview">
            <div slot="header">
              <i class="el-icon-view"></i> 实时预览
            </div>
            <div class="preview-content">
              <div v-if="editForm.title" v-html="renderMarkdown(editForm.title)" class="markdown-body preview-title" v-highlight></div>
              <p v-else class="preview-placeholder">题目标题预览</p>

              <div v-if="editForm.content" v-html="renderMarkdown(editForm.content)" class="markdown-body preview-content-text" v-highlight></div>
              <p v-else class="preview-placeholder">题目内容预览</p>

              <!-- 单选题选项预览 -->
              <div v-if="editForm.type === 'single_choice'" class="preview-options">
                <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === index ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body" v-highlight></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 多选题选项预览 -->
              <div v-if="editForm.type === 'multiple_choice'" class="preview-options">
                <div v-for="(option, index) in editForm.choiceOptions" :key="index" class="preview-option-item">
                  <el-tag :type="editForm.correctAnswers[index] ? 'success' : 'info'" size="small">
                    {{ ['A', 'B', 'C', 'D'][index] }}
                  </el-tag>
                  <div v-if="option" v-html="renderMarkdown(option)" class="markdown-body" v-highlight></div>
                  <div v-else class="preview-placeholder">选项内容</div>
                </div>
              </div>

              <!-- 判断题预览 -->
              <div v-if="editForm.type === 'judge'" class="preview-options">
                <div class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === 'true' ? 'success' : 'info'" size="small">✓</el-tag>
                  <span>{{ $t('m.True') }}</span>
                </div>
                <div class="preview-option-item">
                  <el-tag :type="editForm.correctAnswer === 'false' ? 'success' : 'info'" size="small">✗</el-tag>
                  <span>{{ $t('m.False') }}</span>
                </div>
              </div>

              <!-- 主观题预览 -->
              <div v-if="editForm.type === 'subjective'" class="preview-subjective">
                <el-alert type="info" :closable="false">
                  <i class="el-icon-edit"></i> 主观题，学生需要输入文字答案
                </el-alert>
              </div>

              <!-- 题目解析预览 -->
              <div v-if="editForm.analysis" class="preview-analysis">
                <el-divider content-position="left">
                  <i class="el-icon-document" style="color: #E6A23C;"></i>
                  <span style="color: #E6A23C; font-weight: bold;">题目解析</span>
                </el-divider>
                <div v-html="renderMarkdown(editForm.analysis)" class="markdown-body preview-analysis-content" v-highlight></div>
              </div>
              <p v-else class="preview-placeholder" style="margin-top: 15px;">题目解析预览</p>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <span slot="footer">
        <el-button @click="showEditDialog = false">{{ $t('m.Cancel') }}</el-button>
        <el-button type="primary" @click="updateQuestion">{{ $t('m.Confirm') }}</el-button>
      </span>
    </el-dialog>

    <!-- 查看题目详情对话框 -->
    <el-dialog :title="$t('m.Question_Detail') || '题目详情'" :visible.sync="showViewDialog" width="900px">
      <div v-if="viewQuestion" class="question-detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item :label="$t('m.Question_Type')">
            <el-tag :type="getQuestionTypeColor(viewQuestion.type)">
              {{ getQuestionTypeName(viewQuestion.type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Difficulty')">
            <el-rate :value="getDifficultyStars(viewQuestion.difficulty)" :max="3" disabled />
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Score')">
            {{ viewQuestion.score }} 分
          </el-descriptions-item>
          <el-descriptions-item :label="$t('m.Shared')">
            <el-tag :type="viewQuestion.isShared ? 'success' : 'info'">
              {{ viewQuestion.isShared ? $t('m.Yes') : $t('m.No') }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="所属课程" :span="2">
            <el-tag v-if="viewQuestion.course" type="warning" size="small">{{ viewQuestion.course }}</el-tag>
            <span v-else style="color: #909399;">-</span>
          </el-descriptions-item>
          <el-descriptions-item label="标签" :span="2">
            <el-tag
              v-for="(tag, idx) in parseQuestionTags(viewQuestion.tags)"
              :key="idx"
              size="small"
              type="info"
              style="margin-right: 5px;"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!viewQuestion.tags || viewQuestion.tags === '[]'" style="color: #909399;">-</span>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">
          <i class="el-icon-document"></i>
          <span>题目内容</span>
        </el-divider>

        <!-- 题目标题 -->
        <div class="detail-section">
          <h4 class="detail-label">题目标题：</h4>
          <div v-html="renderMarkdown(viewQuestion.title)" class="markdown-body detail-content" v-highlight></div>
        </div>

        <!-- 题目描述 -->
        <div v-if="viewQuestion.content" class="detail-section">
          <h4 class="detail-label">题目描述：</h4>
          <div v-html="renderMarkdown(viewQuestion.content)" class="markdown-body detail-content" v-highlight></div>
        </div>

        <!-- 选择题选项 -->
        <div v-if="viewQuestion.type === 'single_choice' || viewQuestion.type === 'multiple_choice'" class="detail-section">
          <h4 class="detail-label">选项：</h4>
          <div class="detail-content">
            <div v-if="parseOptionEntries(viewQuestion.options).length > 0">
              <div v-for="(option, index) in parseOptionEntries(viewQuestion.options)" :key="index" class="option-item-detail">
                <span class="option-label">{{ option.letter }}.</span>
                <span v-html="renderMarkdown(option.text)" class="option-text markdown-body" v-highlight></span>
              </div>
            </div>
            <div v-else class="no-options">
              <el-alert type="info" :closable="false">暂无选项数据</el-alert>
            </div>
          </div>
        </div>

        <!-- 组合题子题 -->
        <div v-if="viewQuestion.type === 'composite'" class="detail-section">
          <h4 class="detail-label">组合题子题：</h4>
          <div class="composite-detail-list">
            <div
              v-for="(subQuestion, subIndex) in parseCompositeSubQuestions(viewQuestion.options)"
              :key="subQuestion.id || subIndex"
              class="composite-detail-item"
            >
              <div class="composite-detail-header">
                <span>子题 {{ subIndex + 1 }}</span>
                <span>{{ Number(subQuestion.score || 0) }}分</span>
              </div>
              <div v-html="renderMarkdown(subQuestion.content || '')" class="markdown-body detail-content" v-highlight></div>
              <div class="detail-content composite-detail-options">
                <div v-if="parseOptionEntries(subQuestion.options).length > 0">
                  <div
                    v-for="(option, optionIndex) in parseOptionEntries(subQuestion.options)"
                    :key="`sub-${subIndex}-${optionIndex}`"
                    class="option-item-detail"
                  >
                    <span class="option-label">{{ option.letter }}.</span>
                    <span v-html="renderMarkdown(option.text)" class="option-text markdown-body" v-highlight></span>
                  </div>
                </div>
                <div v-else class="no-options">
                  <el-alert type="info" :closable="false">暂无子题选项数据</el-alert>
                </div>
              </div>
              <div class="composite-answer-row answer-info compact-answer-info">
                <el-tag type="success">
                  正确答案：{{ getCompositeAnswerBySubQuestion(viewQuestion.answer, subQuestion.id, subIndex) || '-' }}
                </el-tag>
              </div>
            </div>
            <el-empty
              v-if="parseCompositeSubQuestions(viewQuestion.options).length === 0"
              description="暂无组合题子题数据"
              :image-size="90"
            />
          </div>
        </div>

        <!-- 正确答案 -->
        <div v-if="viewQuestion.type !== 'composite'" class="detail-section">
          <h4 class="detail-label">正确答案：</h4>
          <div class="detail-content answer-info compact-answer-info">
            <el-tag v-if="viewQuestion.type === 'single_choice'" type="success">
              {{ parseSingleChoiceAnswer(viewQuestion) }}
            </el-tag>
            <el-tag v-else-if="viewQuestion.type === 'multiple_choice'" type="success">
              {{ parseMultipleChoiceAnswer(viewQuestion) }}
            </el-tag>
            <el-tag v-else-if="viewQuestion.type === 'judge'" :type="isJudgeTrue(viewQuestion.answer) ? 'success' : 'danger'">
              {{ isJudgeTrue(viewQuestion.answer) ? $t('m.True') || '正确' : $t('m.False') || '错误' }}
            </el-tag>
            <el-tag v-else-if="viewQuestion.type === 'fill_blank'" type="success">
              {{ parseFillBlankAnswer(viewQuestion.answer) }}
            </el-tag>
            <div v-else-if="viewQuestion.type === 'subjective'" class="subjective-answer">
              <div v-if="viewQuestion.answer" v-html="renderMarkdown(viewQuestion.answer)" class="markdown-body" v-highlight></div>
              <span v-else style="color: #909399;">暂无参考答案</span>
            </div>
          </div>
        </div>

        <!-- 题目解析 -->
        <div v-if="viewQuestion.analysis" class="detail-section">
          <el-divider content-position="left">
            <i class="el-icon-document" style="color: #E6A23C;"></i>
            <span style="color: #E6A23C; font-weight: bold;">题目解析</span>
          </el-divider>
          <div v-html="renderMarkdown(viewQuestion.analysis)" class="markdown-body detail-content" v-highlight></div>
        </div>

        <!-- 创建和更新时间 -->
        <div class="detail-meta">
          <el-tag size="mini" type="info">
            创建时间：{{ formatTime(viewQuestion.createdAt) }}
          </el-tag>
          <el-tag size="mini" type="info" style="margin-left: 10px;">
            更新时间：{{ formatTime(viewQuestion.updatedAt) }}
          </el-tag>
        </div>
      </div>
      <span slot="footer">
        <el-button @click="showViewDialog = false">{{ $t('m.Close') || '关闭' }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import MarkdownIt from 'markdown-it'
import katex from '@iktakahiro/markdown-it-katex'
import 'katex/dist/katex.min.css'
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

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'QuestionBank',
  mixins: [realtimeSync, teacherAuth],
  props: {
    classroomId: {
      type: [String, Number],
      required: false
    }
  },
  data() {
    return {
      loading: false,
      requesting: false,
      questions: [],
      showCreateDialog: false,
      showEditDialog: false,
      currentEditId: null,
      createTagInput: '', // 创建题目时的标签输入
      editTagInput: '', // 编辑题目时的标签输入
      filters: {
        type: '',
        course: '',
        tag: '',
        difficulty: '',
        questionId: '',
        keyword: ''
      },
      pagination: {
        currentPage: 1,
        pageSize: 20,
        total: 0
      },
      createForm: {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''], // 4个选项的内容
        correctAnswer: 0, // 单选题正确答案（0-3）
        correctAnswers: [false, false, false, false], // 多选题正确答案（数组）
        referenceAnswer: '', // 主观题参考答案
        analysis: '', // 题目解析
        tags: [], // 题目标签（数组）
        course: '', // 题目所属课程
        difficulty: 1,
        score: 2, // 单选题默认2分
        isShared: false
      },
      editForm: {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        analysis: '', // 题目解析
        tags: [], // 题目标签
        course: '', // 题目所属课程
        difficulty: 1,
        score: 2, // 单选题默认2分
        isShared: false
      },
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000,
        syncFunction: 'loadQuestions',
        immediate: true
      },
      // 常用标签
      commonTags: [
        '基础概念',
        '逻辑推理',
        '计算题',
        '应用题',
        '综合分析',
        '易错题',
        '重点',
        '难点'
      ],
      // 常用课程
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
      // 查看详情
      showViewDialog: false,
      viewQuestion: null
    }
  },
  mounted() {
    // 由 realtimeSync mixin 自动启动同步
  },
  methods: {
    handleTypeChange() {
      this.handleFormTypeChange(this.createForm)
    },
    handleEditTypeChange() {
      this.handleFormTypeChange(this.editForm)
    },
    handleFormTypeChange(form) {
      // 切换题目类型时重置答案相关字段和默认分数
      if (form.type === 'single_choice') {
        form.correctAnswer = 0
        form.correctAnswers = [false, false, false, false]
        form.score = 2 // 单选题默认2分
      } else if (form.type === 'multiple_choice') {
        form.correctAnswer = 0
        form.correctAnswers = [false, false, false, false]
        form.score = 5 // 多选题默认5分
      } else if (form.type === 'judge') {
        form.score = 1 // 判断题默认1分
        form.correctAnswer = 'true'
        form.correctAnswers = [false, false, false, false]
      } else if (form.type === 'subjective') {
        form.score = 5 // 主观题默认5分
        form.correctAnswer = ''
        form.correctAnswers = [false, false, false, false]
      } else {
        form.correctAnswer = 0
        form.correctAnswers = [false, false, false, false]
      }
      form.referenceAnswer = ''
    },
    handleFilterChange() {
      this.pagination.currentPage = 1
      this.loadQuestions()
    },
    handlePageSizeChange(size) {
      this.pagination.pageSize = size
      this.pagination.currentPage = 1
      this.loadQuestions()
    },
    handleCurrentPageChange(page) {
      this.pagination.currentPage = page
      this.loadQuestions()
    },
    getQuestionRowClass({ row }) {
      const normalizedType = String(row && row.type ? row.type : '').replace(/_/g, '-')
      return `question-bank-row type-${normalizedType}`
    },
    async loadQuestions() {
      // 避免重复请求
      if (this.requesting) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.questions.length === 0 && this.pagination.total === 0
      if (isFirstLoad) {
        this.loading = true
      }
      this.requesting = true

      try {
        const questionId = String(this.filters.questionId || '').trim()
        const keyword = String(this.filters.keyword || '').trim()

        const res = await this.$store.dispatch('classroom/getQuestionBank', {
          classroomId: this.classroomId,
          page: this.pagination.currentPage,
          limit: this.pagination.pageSize,
          type: this.filters.type || undefined,
          course: this.filters.course || undefined,
          tag: this.filters.tag || undefined,
          difficulty: this.filters.difficulty || undefined,
          questionId: questionId || undefined,
          keyword: questionId ? undefined : keyword || undefined
        })
        if (res.code === 200) {
          const newQuestions = res.data.questions || res.data || []
          const total = Number(res.data.total || 0)

          // 深度对比：使用 JSON.stringify 检查数据是否真的变化
          const currentDataString = JSON.stringify(this.questions)
          const newDataString = JSON.stringify(newQuestions)

          if (currentDataString !== newDataString) {
            // 数据真的变化了，才更新
            this.questions = newQuestions
          }
          if (this.pagination.total !== total) {
            this.pagination.total = total
          }
        }
      } catch (error) {
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        this.requesting = false
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    async createQuestion() {
      // 验证必填字段
      if (!this.createForm.title || !this.createForm.title.trim()) {
        this.$message.warning('请输入题目标题')
        return
      }
      if (!this.createForm.content || !this.createForm.content.trim()) {
        this.$message.warning('请输入题目内容')
        return
      }

      // 构建提交数据
      const submitData = {
        classroomId: this.classroomId,
        type: this.createForm.type,
        title: this.createForm.title,
        content: this.createForm.content,
        analysis: this.createForm.analysis || '', // 题目解析
        tags: JSON.stringify(this.createForm.tags || []), // 题目标签（JSON格式）
        course: this.createForm.course || '', // 题目所属课程
        difficulty: this.createForm.difficulty || 1,
        score: this.createForm.score || 2,
        isShared: this.createForm.isShared ? 1 : 0
      }

      // 根据题型设置答案格式
      if (this.createForm.type === 'single_choice') {
        // 单选题
        const optionsArray = this.createForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        submitData.answer = ['A', 'B', 'C', 'D'][this.createForm.correctAnswer]
      } else if (this.createForm.type === 'multiple_choice') {
        // 多选题
        const optionsArray = this.createForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        // 多选题答案：保存为 JSON 数组格式，如 ["A","B","C"]
        const selectedAnswers = this.createForm.correctAnswers
          .map((selected, idx) => selected ? ['A', 'B', 'C', 'D'][idx] : null)
          .filter(Boolean)
        submitData.answer = JSON.stringify(selectedAnswers)
      } else if (this.createForm.type === 'judge') {
        // 判断题 - 确保 correctAnswer 是字符串类型
        const answerValue = String(this.createForm.correctAnswer)
        submitData.answer = answerValue === 'true' ? 'true' : 'false'
        submitData.options = null // 判断题不需要选项
      } else if (this.createForm.type === 'subjective') {
        // 主观题
        submitData.answer = this.createForm.referenceAnswer || '需人工评分'
        submitData.options = null // 主观题不需要选项
      }

      try {
        const res = await this.$store.dispatch('classroom/createQuestion', submitData)
        if (res.code === 200) {
          this.$message.success(this.$t('m.Create_Success'))
          this.showCreateDialog = false
          this.resetForm()
          this.loadQuestions()
        } else {
          this.$message.error(res.message || this.$t('m.Create_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Create_Failed'))
      }
    },
    resetForm() {
      this.createForm = {
        type: 'single_choice',
        title: '',
        content: '',
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        analysis: '', // 题目解析
        tags: [], // 题目标签
        course: '', // 题目所属课程
        difficulty: 1,
        score: 10,
        isShared: false
      }
      this.createTagInput = '' // 清空标签输入
    },
    // 添加创建题目的标签
    addCreateTag() {
      const tag = this.createTagInput.trim()
      if (tag && !this.createForm.tags.includes(tag)) {
        this.createForm.tags.push(tag)
      }
      this.createTagInput = '' // 清空输入
    },
    // 删除创建题目的标签
    removeCreateTag(index) {
      this.createForm.tags.splice(index, 1)
    },
    // 添加编辑题目的标签
    addEditTag() {
      const tag = this.editTagInput.trim()
      if (tag && !this.editForm.tags.includes(tag)) {
        this.editForm.tags.push(tag)
      }
      this.editTagInput = '' // 清空输入
    },
    // 删除编辑题目的标签
    removeEditTag(index) {
      this.editForm.tags.splice(index, 1)
    },
    // 查看题目详情
    handleViewDetail(question) {
      this.viewQuestion = question
      this.showViewDialog = true
    },
    goCreatePage() {
      const query = {}
      if (this.classroomId) {
        query.classroomId = this.classroomId
      }
      this.$router.push({ name: 'QuestionBankCreate', query })
    },
    goEditPage(question) {
      if (!question || !question.id) {
        this.$message.warning('题目ID无效')
        return
      }
      this.$router.push({ name: 'QuestionBankEdit', params: { questionId: String(question.id) } })
    },
    handleEdit(question) {
      this.currentEditId = question.id
      // 解析题目数据并填充到编辑表单
      this.editForm = {
        type: question.type,
        title: question.title,
        content: question.content || '',
        difficulty: question.difficulty,
        score: question.score,
        isShared: question.isShared === 1,
        choiceOptions: ['', '', '', ''],
        correctAnswer: 0,
        correctAnswers: [false, false, false, false],
        referenceAnswer: '',
        analysis: question.analysis || '', // 题目解析
        tags: [], // 题目标签
        course: question.course || '' // 题目所属课程
      }

      // 解析标签（从JSON字符串转为数组）
      if (question.tags) {
        try {
          this.editForm.tags = JSON.parse(question.tags)
        } catch (e) {
          this.editForm.tags = []
        }
      }

      // 解析选项和答案
      if (question.type === 'single_choice' || question.type === 'multiple_choice') {
        // 解析选项 JSON
        if (question.options) {
          try {
            const optionsArray = JSON.parse(question.options)
            this.editForm.choiceOptions = optionsArray.map(opt => {
              // 去掉 "A. " 这样的前缀
              return opt.replace(/^[A-D]\.\s*/, '')
            })
          } catch (e) {
            console.error('解析选项失败', e)
          }
        }

        // 解析答案
        if (question.type === 'single_choice') {
          const answerMap = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 }
          this.editForm.correctAnswer = answerMap[question.answer] || 0
        } else {
          // 多选题
          const answerMap = { 'A': 0, 'B': 1, 'C': 2, 'D': 3 }
          this.editForm.correctAnswers = [false, false, false, false]

          // 尝试解析 JSON 数组格式
          try {
            const answers = JSON.parse(question.answer)
            if (Array.isArray(answers)) {
              answers.forEach(ans => {
                const idx = answerMap[ans]
                if (idx !== undefined) {
                  this.editForm.correctAnswers[idx] = true
                }
              })
            }
          } catch (e) {
            // 兼容旧的逗号分隔格式
            const answers = question.answer.split(',')
            answers.forEach(ans => {
              const idx = answerMap[ans.trim()]
              if (idx !== undefined) {
                this.editForm.correctAnswers[idx] = true
              }
            })
          }
        }
      } else if (question.type === 'judge') {
        this.editForm.correctAnswer = this.isJudgeTrue(question.answer) ? 'true' : 'false'
      } else if (question.type === 'subjective') {
        this.editForm.referenceAnswer = question.answer || ''
      }

      this.showEditDialog = true
    },
    async updateQuestion() {
      // 验证必填字段
      if (!this.editForm.title || !this.editForm.title.trim()) {
        this.$message.warning('请输入题目标题')
        return
      }
      if (!this.editForm.content || !this.editForm.content.trim()) {
        this.$message.warning('请输入题目内容')
        return
      }

      const submitData = {
        type: this.editForm.type,
        title: this.editForm.title,
        content: this.editForm.content,
        analysis: this.editForm.analysis || '', // 题目解析
        tags: JSON.stringify(this.editForm.tags || []), // 题目标签（JSON格式）
        course: this.editForm.course || '', // 题目所属课程
        difficulty: this.editForm.difficulty || 1,
        score: this.editForm.score || 2,
        isShared: this.editForm.isShared ? 1 : 0
      }

      // 根据题型设置答案格式
      if (this.editForm.type === 'single_choice') {
        const optionsArray = this.editForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        submitData.answer = ['A', 'B', 'C', 'D'][this.editForm.correctAnswer]
      } else if (this.editForm.type === 'multiple_choice') {
        const optionsArray = this.editForm.choiceOptions.map((opt, idx) =>
          `${['A', 'B', 'C', 'D'][idx]}. ${opt}`
        )
        submitData.options = JSON.stringify(optionsArray)
        // 多选题答案：保存为 JSON 数组格式，如 ["A","B","C"]
        const selectedAnswers = this.editForm.correctAnswers
          .map((selected, idx) => selected ? ['A', 'B', 'C', 'D'][idx] : null)
          .filter(Boolean)
        submitData.answer = JSON.stringify(selectedAnswers)
      } else if (this.editForm.type === 'judge') {
        // 判断题 - 确保 correctAnswer 是字符串类型
        const answerValue = String(this.editForm.correctAnswer)
        submitData.answer = answerValue === 'true' ? 'true' : 'false'
        submitData.options = null // 判断题不需要选项
      } else if (this.editForm.type === 'subjective') {
        submitData.answer = this.editForm.referenceAnswer || '需人工评分'
        submitData.options = null // 主观题不需要选项
      }

      try {
        const res = await this.$store.dispatch('classroom/updateQuestion', {
          questionId: this.currentEditId,
          data: submitData
        })
        if (res.code === 200) {
          this.$message.success(this.$t('m.Update_Success'))
          this.showEditDialog = false
          this.loadQuestions()
        } else {
          this.$message.error(res.message || this.$t('m.Update_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Update_Failed'))
      }
    },
    async handleDelete(question) {
      this.$confirm(this.$t('m.Confirm_Delete_Question'), this.$t('m.Warning'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await this.$store.dispatch('classroom/deleteQuestion', question.id)
          if (res.code === 200) {
            this.$message.success(this.$t('m.Delete_Success'))
            if (this.questions.length === 1 && this.pagination.currentPage > 1) {
              this.pagination.currentPage -= 1
            }
            this.loadQuestions()
          } else {
            this.$message.error(res.message || this.$t('m.Delete_Failed'))
          }
        } catch (error) {
          this.$message.error(this.$t('m.Delete_Failed'))
        }
      })
    },
    getQuestionTypeName(type) {
      const map = {
        single_choice: this.$t('m.Single_Choice'),
        multiple_choice: this.$t('m.Multiple_Choice'),
        judge: this.$t('m.Judge'),
        fill_blank: '填空题',
        subjective: this.$t('m.Subjective'),
        composite: '组合题'
      }
      return map[type] || type
    },
    getDifficultyStars(difficulty) {
      return parseInt(difficulty) || 1
    },
    getDifficultyText(difficulty) {
      const level = Number(difficulty)
      if (level === 1) return '简单'
      if (level === 2) return '中等'
      if (level === 3) return '困难'
      return `等级${difficulty}`
    },
    // 解析题目标签
    parseQuestionTags(tags) {
      if (!tags) return []
      try {
        const parsed = JSON.parse(tags)
        return Array.isArray(parsed) ? parsed : []
      } catch (e) {
        return []
      }
    },
    getQuestionTypeColor(type) {
      const map = {
        single_choice: 'primary',
        multiple_choice: 'success',
        judge: 'warning',
        fill_blank: 'success',
        subjective: 'info',
        composite: 'danger'
      }
      return map[type] || ''
    },
    goBack() {
      this.$router.push({ name: 'TeacherDashboard' })
    },
    renderMarkdown(content) {
      if (!content) return ''
      try {
        return md.render(content)
      } catch (e) {
        console.error('Markdown渲染失败:', e)
        return content
      }
    },
    stripOptionPrefix(optionText, letterHint = '') {
      const normalizedText = String(optionText || '').trim()
      if (!normalizedText) return ''
      const escapedHint = letterHint ? letterHint.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') : '[A-Za-z]'
      const prefixRegex = new RegExp(`^\\s*(?:${escapedHint}|[A-Za-z])\\s*[\\.\\)、:：]\\s*`)
      return normalizedText.replace(prefixRegex, '').trim()
    },
    parseOptionEntries(optionsInput) {
      if (!optionsInput) return []
      let parsed = optionsInput
      if (typeof parsed === 'string') {
        try {
          parsed = JSON.parse(parsed)
        } catch (e) {
          return []
        }
      }
      if (!Array.isArray(parsed)) return []
      return parsed.map((opt, idx) => {
        if (opt && typeof opt === 'object' && !Array.isArray(opt)) {
          const normalizedLetter = String(opt.letter || opt.label || String.fromCharCode(65 + idx))
            .trim()
            .replace(/[^A-Za-z0-9]/g, '')
            .toUpperCase()
          const letter = normalizedLetter || String.fromCharCode(65 + idx)
          const rawText = opt.text !== undefined ? opt.text : (opt.content !== undefined ? opt.content : '')
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
      let parsed = optionsInput
      if (typeof parsed === 'string') {
        try {
          parsed = JSON.parse(parsed)
        } catch (e) {
          return []
        }
      }
      if (!Array.isArray(parsed)) return []
      return parsed.map((subQuestion, index) => {
        const optionSource = subQuestion && subQuestion.options !== undefined
          ? subQuestion.options
          : (subQuestion && subQuestion.choiceOptions !== undefined ? subQuestion.choiceOptions : [])
        return {
          id: String((subQuestion && subQuestion.id) || `sq_${index + 1}`),
          content: (subQuestion && subQuestion.content) || '',
          options: optionSource,
          score: Number((subQuestion && (subQuestion.score || subQuestion.subScore)) || 0)
        }
      })
    },
    parseAnswerArray(answerInput, { allowCommaSplit = true } = {}) {
      if (!answerInput) return []
      let parsed = answerInput
      if (typeof parsed === 'string') {
        try {
          parsed = JSON.parse(parsed)
        } catch (e) {
          parsed = answerInput
        }
      }
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
      if (!answerInput) return {}
      let parsed = answerInput
      if (typeof parsed === 'string') {
        try {
          parsed = JSON.parse(parsed)
        } catch (e) {
          return {}
        }
      }
      return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
    },
    getCompositeAnswerBySubQuestion(answerInput, subQuestionId, subIndex) {
      const answerMap = this.parseCompositeAnswerMap(answerInput)
      const candidates = [
        String(subQuestionId || ''),
        String(subIndex + 1),
        String(subIndex),
        `sub_${subIndex + 1}`,
        `sq_${subIndex + 1}`
      ]
      for (const key of candidates) {
        if (!key) continue
        if (Object.prototype.hasOwnProperty.call(answerMap, key)) {
          const value = String(answerMap[key] || '').trim()
          if (value) return value
        }
      }
      return ''
    },
    formatCompositeAnswerSummary(answerInput, optionsInput, emptyText = '暂无答案') {
      const subQuestions = this.parseCompositeSubQuestions(optionsInput)
      if (subQuestions.length === 0) {
        const answerMap = this.parseCompositeAnswerMap(answerInput)
        const entries = Object.keys(answerMap)
          .map(key => `${key}: ${String(answerMap[key] || '').trim()}`)
          .filter(item => item && !item.endsWith(': '))
        return entries.length > 0 ? entries.join('； ') : emptyText
      }
      const segments = []
      subQuestions.forEach((subQuestion, index) => {
        const value = this.getCompositeAnswerBySubQuestion(answerInput, subQuestion.id, index)
        if (value) {
          segments.push(`子题${index + 1}: ${value}`)
        }
      })
      return segments.length > 0 ? segments.join('； ') : emptyText
    },
    formatInlineAnswer(question) {
      if (!question) return '暂无答案'
      switch (question.type) {
        case 'single_choice': {
          const value = String(question.answer || '').trim()
          return value || '暂无答案'
        }
        case 'multiple_choice': {
          const values = this.parseAnswerArray(question.answer)
          return values.length > 0 ? values.join('、') : '暂无答案'
        }
        case 'judge': {
          const raw = String(question.answer || '').trim()
          if (!raw) return '暂无答案'
          return this.isJudgeTrue(raw) ? '正确' : '错误'
        }
        case 'fill_blank': {
          const values = this.parseAnswerArray(question.answer, { allowCommaSplit: false })
          return values.length > 0 ? values.join(' / ') : '暂无答案'
        }
        case 'composite':
          return this.formatCompositeAnswerSummary(question.answer, question.options)
        default: {
          const value = String(question.answer || '').trim()
          return value || '暂无答案'
        }
      }
    },
    // 解析选择题选项
    parseQuestionOptions(options) {
      if (!options) return []
      try {
        // options 字段存储的是 JSON 字符串数组，格式如 ["A. 选项1", "B. 选项2", ...]
        const optionsArray = typeof options === 'string' ? JSON.parse(options) : options
        // 去掉 "A. "、"B. " 这样的前缀，只保留选项内容
        return optionsArray.map(opt => {
          if (typeof opt === 'string') {
            return opt.replace(/^[A-D]\.\s*/, '').trim()
          }
          return opt
        })
      } catch (e) {
        console.error('解析选项失败:', e, options)
        return []
      }
    },
    // 解析单选题答案
    parseSingleChoiceAnswer(question) {
      if (question.answer) {
        return `答案：${question.answer}`
      }
      return '暂无答案'
    },
    // 解析多选题答案
    parseMultipleChoiceAnswer(question) {
      if (question.answer) {
        try {
          const answers = typeof question.answer === 'string'
            ? JSON.parse(question.answer)
            : question.answer
          return `答案：${answers.join('、')}`
        } catch (e) {
          return `答案：${question.answer}`
        }
      }
      return '暂无答案'
    },
    isJudgeTrue(answer) {
      const raw = String(answer || '').trim()
      const lowered = raw.toLowerCase()
      return lowered === 'true'
    },
    parseFillBlankAnswer(answer) {
      if (!answer) return '暂无答案'
      try {
        const parsed = typeof answer === 'string' ? JSON.parse(answer) : answer
        if (Array.isArray(parsed)) {
          const values = parsed.map(item => String(item || '').trim()).filter(Boolean)
          return values.length > 0 ? `答案：${values.join(' / ')}` : '暂无答案'
        }
      } catch (e) {
        // fall through
      }
      const normalized = String(answer || '').trim()
      return normalized ? `答案：${normalized}` : '暂无答案'
    },
    // 格式化时间
    formatTime(time) {
      if (!time) return '--'
      try {
        const date = new Date(time)
        return date.toLocaleString('zh-CN')
      } catch (e) {
        return '--'
      }
    }
  }
}
</script>

<style scoped>
.question-bank-panel {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.page-header h3 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.action-bar {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.question-bank-panel >>> .question-bank-table.el-table::before {
  height: 0;
}

.question-bank-panel >>> .question-bank-table .el-table__body-wrapper table {
  border-collapse: separate;
  border-spacing: 0 10px;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row > td {
  background: #fff;
  border-top: 1px solid #ebeef5;
  border-bottom: 1px solid #ebeef5;
  vertical-align: top;
  transition: background-color 0.2s ease;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row > td:first-child {
  border-left: 4px solid #dcdfe6;
  border-radius: 10px 0 0 10px;
  padding-left: 12px;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row > td:last-child {
  border-right: 1px solid #ebeef5;
  border-radius: 0 10px 10px 0;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row:hover > td {
  background: #f8fbff;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row.type-single-choice > td:first-child {
  border-left-color: #409eff;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row.type-multiple-choice > td:first-child {
  border-left-color: #67c23a;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row.type-judge > td:first-child {
  border-left-color: #e6a23c;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row.type-fill-blank > td:first-child {
  border-left-color: #67c23a;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row.type-subjective > td:first-child {
  border-left-color: #909399;
}

.question-bank-panel >>> .question-bank-table .el-table__body tr.question-bank-row.type-composite > td:first-child {
  border-left-color: #f56c6c;
}

.inline-question-cell {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.inline-meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: nowrap;
  overflow-x: auto;
  white-space: nowrap;
  padding-bottom: 2px;
}

.inline-meta-row::-webkit-scrollbar {
  height: 4px;
}

.inline-meta-row::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 2px;
}

.inline-question-title {
  font-weight: 600;
  color: #303133;
}

.inline-question-content {
  color: #606266;
}

.inline-options-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px 12px;
}

.inline-option-item {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  min-width: 0;
  padding: 4px 6px;
  border-radius: 4px;
  background: #f8fafc;
}

.inline-option-label {
  flex-shrink: 0;
  width: 20px;
  color: #409eff;
  font-weight: 600;
  line-height: 1.6;
}

.inline-option-text {
  flex: 1;
  min-width: 0;
  line-height: 1.6;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.inline-question-cell >>> .inline-option-text p,
.inline-question-cell >>> .inline-question-content p,
.inline-question-cell >>> .inline-answer-text p {
  margin: 0;
}

.inline-question-cell >>> .inline-option-text pre,
.inline-question-cell >>> .inline-question-content pre,
.inline-question-cell >>> .inline-answer-text pre {
  margin: 4px 0 0;
  max-width: 100%;
  overflow-x: auto;
}

.inline-composite-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.inline-composite-item {
  padding: 8px;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
}

.inline-composite-head {
  font-weight: 600;
  color: #606266;
  margin-bottom: 6px;
}

.inline-composite-content {
  margin-bottom: 6px;
}

.compact-answer-info {
  margin-top: 8px;
  padding: 8px 10px;
}

.answer-info {
  margin-top: 15px;
  padding: 10px;
  background: #f0f9ff;
  border-left: 3px solid #409eff;
  border-radius: 4px;
}

.inline-answer-row {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 4px;
  background: #f0f9ff;
  border: 1px solid #d9ecff;
}

.inline-answer-label {
  flex-shrink: 0;
  color: #409eff;
  font-weight: 600;
}

.inline-answer-text {
  flex: 1;
  min-width: 0;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.question-form-column,
.preview-column {
  max-height: 72vh;
  overflow-y: auto;
}

.question-form-column {
  padding-right: 6px;
}

.preview-column {
  padding-left: 6px;
}

.question-form .el-form-item {
  margin-bottom: 14px;
}

.compact-form-row .el-form-item {
  margin-bottom: 10px;
}

.compact-tip {
  margin-top: 4px;
}

.options-container {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 14px;
}

.option-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.option-radio {
  margin-right: 8px;
  font-weight: bold;
  font-size: 14px;
  min-width: 20px;
}

.option-checkbox {
  margin-right: 8px;
  font-weight: bold;
  font-size: 14px;
  min-width: 20px;
}

.answer-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.answer-tip i {
  font-size: 14px;
}

.form-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.form-tip i {
  font-size: 14px;
}

/* 预览卡片样式 */
.preview-card {
  min-height: 100%;
  border: 2px solid #E4E7ED;
}

.preview-content {
  padding: 10px;
}

.preview-title {
  color: #303133;
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 15px;
  border-bottom: 2px solid #E4E7ED;
  padding-bottom: 10px;
}

.preview-content-text {
  color: #606266;
  line-height: 1.8;
  margin-bottom: 20px;
  font-size: 14px;
}

.preview-placeholder {
  color: #C0C4CC;
  font-style: italic;
}

.preview-options {
  margin-top: 15px;
}

.preview-option-item {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  margin-bottom: 10px;
  background: #F5F7FA;
  border-radius: 4px;
  gap: 8px;
}

.preview-option-item .el-tag {
  flex-shrink: 0;
  margin-right: 0;
  padding: 0 6px !important;
  height: 20px !important;
  line-height: 20px !important;
  font-size: 12px !important;
  width: 24px !important;
  min-width: 24px !important;
  max-width: 24px !important;
  text-align: center;
  display: inline-block;
  box-sizing: border-box;
}

.preview-option-item .markdown-body {
  flex: 1;
  line-height: 1.6;
  word-wrap: break-word;
  word-break: break-word;
}

.preview-option-item span {
  flex: 1;
  line-height: 1.6;
}

.preview-subjective {
  margin-top: 15px;
}

/* 标签输入容器样式 */
.tags-input-container {
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 5px;
  min-height: 80px;
}

.tags-list {
  margin-bottom: 8px;
}

/* 题目解析预览样式 */
.preview-analysis {
  margin-top: 15px;
  padding: 10px;
  background: #fff9e6;
  border-left: 3px solid #E6A23C;
  border-radius: 4px;
}

.preview-analysis-content {
  margin-top: 10px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  line-height: 1.8;
}

/* 题目详情样式 */
.question-detail-content {
  padding: 10px;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-label {
  color: #409EFF;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
}

.detail-content {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
  line-height: 1.8;
}

.option-item-detail {
  display: block;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.option-item-detail:last-child {
  border-bottom: none;
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
  color: #303133;
  font-weight: 600;
}

.composite-detail-options {
  margin-top: 8px;
}

.composite-answer-row {
  margin-top: 8px;
}

.option-label {
  font-weight: bold;
  color: #409EFF;
  display: block;
  margin-bottom: 6px;
}

.option-text {
  display: block;
  width: 100%;
  word-wrap: break-word;
  overflow-wrap: anywhere;
}

.option-text p {
  margin: 0;
}

.option-text pre {
  margin: 0;
  max-width: 100%;
  overflow-x: auto;
}

.no-options {
  padding: 10px;
}

.subjective-answer {
  padding: 10px;
  background: white;
  border-radius: 4px;
}

.detail-meta {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

@media (max-width: 1280px) {
  .inline-options-list {
    grid-template-columns: 1fr;
  }

  .options-container {
    grid-template-columns: 1fr;
  }
}
</style>

<!-- 非scoped样式，确保markdown-body样式生效 -->
<style>
/* 预览卡片中的 Markdown 样式 - 高优先级 */
.markdown-preview .markdown-body,
.el-card.preview-card .markdown-body,
.preview-card .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.markdown-preview .markdown-body h1,
.markdown-preview .markdown-body h2,
.markdown-preview .markdown-body h3,
.markdown-preview .markdown-body h4,
.markdown-preview .markdown-body h5,
.markdown-preview .markdown-body h6,
.el-card.preview-card .markdown-body h1,
.el-card.preview-card .markdown-body h2,
.el-card.preview-card .markdown-body h3,
.el-card.preview-card .markdown-body h4,
.el-card.preview-card .markdown-body h5,
.el-card.preview-card .markdown-body h6,
.preview-card .markdown-body h1,
.preview-card .markdown-body h2,
.preview-card .markdown-body h3,
.preview-card .markdown-body h4,
.preview-card .markdown-body h5,
.preview-card .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.markdown-preview .markdown-body h1,
.el-card.preview-card .markdown-body h1,
.preview-card .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.markdown-preview .markdown-body h2,
.el-card.preview-card .markdown-body h2,
.preview-card .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.markdown-preview .markdown-body h3,
.el-card.preview-card .markdown-body h3,
.preview-card .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.markdown-preview .markdown-body h3:before,
.el-card.preview-card .markdown-body h3:before,
.preview-card .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.markdown-preview .markdown-body p,
.el-card.preview-card .markdown-body p,
.preview-card .markdown-body p {
  margin-bottom: 16px !important;
}

.markdown-preview .markdown-body strong,
.el-card.preview-card .markdown-body strong,
.preview-card .markdown-body strong {
  font-weight: bold !important;
}

.markdown-preview .markdown-body em,
.el-card.preview-card .markdown-body em,
.preview-card .markdown-body em {
  font-style: italic !important;
}

.markdown-preview .markdown-body code,
.el-card.preview-card .markdown-body code,
.preview-card .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.markdown-preview .markdown-body pre,
.el-card.preview-card .markdown-body pre,
.preview-card .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

/* 通用样式 - 兼容没有 .el-card 前缀的情况 */
.preview-card .markdown-body {
  font-size: 15px !important;
  word-wrap: break-word !important;
  word-break: break-word !important;
  line-height: 1.8 !important;
  color: #606266 !important;
}

.preview-card .markdown-body h1,
.preview-card .markdown-body h2,
.preview-card .markdown-body h3,
.preview-card .markdown-body h4,
.preview-card .markdown-body h5,
.preview-card .markdown-body h6 {
  position: relative !important;
  margin-top: 1em !important;
  margin-bottom: 16px !important;
  font-weight: bold !important;
  line-height: 1.4 !important;
}

.preview-card .markdown-body h1 {
  padding-bottom: 0.3em !important;
  font-size: 1.86em !important;
  line-height: 1.2 !important;
  border-bottom: 1px solid #eee !important;
}

.preview-card .markdown-body h2 {
  font-size: 1.45em !important;
  line-height: 1.425 !important;
  border-bottom: 1px solid #eee !important;
  background: #cce5ff !important;
  padding: 8px 10px !important;
  color: #545857 !important;
  border-radius: 3px !important;
}

.preview-card .markdown-body h3 {
  font-size: 1.3em !important;
  line-height: 1.43 !important;
}

.preview-card .markdown-body h3:before {
  content: "" !important;
  border-left: 4px solid #03a9f4 !important;
  padding-left: 6px !important;
}

.preview-card .markdown-body p {
  margin-bottom: 16px !important;
}

.preview-card .markdown-body strong {
  font-weight: bold !important;
}

.preview-card .markdown-body em {
  font-style: italic !important;
}

.preview-card .markdown-body code {
  background: #f8f8f9 !important;
  padding: 2px 6px !important;
  border-radius: 3px !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace !important;
}

.preview-card .markdown-body pre {
  padding: 5px 10px !important;
  white-space: pre-wrap !important;
  margin-top: 15px !important;
  margin-bottom: 15px !important;
  background: #f8f8f9 !important;
  border: 1px dashed #e9eaec !important;
  border-radius: 3px !important;
}

/* 教师题库：代码块取消默认左缩进 */
.question-bank-panel .markdown-body pre,
.question-bank-panel .inline-question-cell .markdown-body pre,
.question-bank-panel .question-detail-content .markdown-body pre {
  padding-left: 0 !important;
  text-indent: 0 !important;
}

.question-bank-panel .markdown-body pre ol.pre-numbering,
.question-bank-panel .inline-question-cell .markdown-body pre ol.pre-numbering,
.question-bank-panel .question-detail-content .markdown-body pre ol.pre-numbering {
  display: none !important;
}

.question-bank-panel .markdown-body pre code,
.question-bank-panel .markdown-body code.hljs,
.question-bank-panel .inline-question-cell .markdown-body pre code,
.question-bank-panel .inline-question-cell .markdown-body code.hljs,
.question-bank-panel .question-detail-content .markdown-body pre code,
.question-bank-panel .question-detail-content .markdown-body code.hljs {
  margin-left: 0 !important;
  padding-left: 0 !important;
  text-indent: 0 !important;
}
</style>
