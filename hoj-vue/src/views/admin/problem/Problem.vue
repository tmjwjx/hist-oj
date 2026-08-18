<template>
  <div class="problem">
    <el-card>
      <div slot="header">
        <span class="panel-title home-title">{{ title }}</span>
      </div>
      <el-steps :active="editorStep" finish-status="success" align-center class="problem-steps">
        <el-step title="题面与配置" description="题目 ID、描述、输入输出与判题配置" />
        <el-step title="样例与测试数据" description="导入 ZIP 或手工测试点，仅数据变化时同步" />
        <el-step title="验证题目" description="AI 验题与创建者标准程序验题" />
      </el-steps>
      <el-form
        v-show="editorStep < 2"
        ref="form"
        :model="problem"
        :rules="rules"
        label-position="top"
        label-width="70px"
        class="problem-editor-form"
      >
        <div v-show="editorStep === 0">
        <div class="problem-header-grid">
          <div class="problem-overview-grid">
            <el-form-item class="overview-field overview-id" prop="problemId" :label="$t('m.Problem_Display_ID')" required>
              <el-input :placeholder="$t('m.Problem_Display_ID')" v-model="problem.problemId" :disabled="problem.isRemote" />
            </el-form-item>
            <el-form-item class="overview-field overview-title" prop="title" :label="$t('m.Title')" required>
              <el-input :placeholder="$t('m.Title')" v-model="problem.title" />
            </el-form-item>
            <el-form-item class="overview-field" :label="$t('m.Time_Limit') + '(ms)'" required>
              <el-input type="Number" v-model="problem.timeLimit" :disabled="problem.isRemote" />
            </el-form-item>
            <el-form-item class="overview-field" :label="$t('m.Memory_Limit') + '(mb)'" required>
              <el-input type="Number" v-model="problem.memoryLimit" :disabled="problem.isRemote" />
            </el-form-item>
            <el-form-item class="overview-field" :label="$t('m.Stack_Limit') + '(mb)'" required>
              <el-input type="Number" v-model="problem.stackLimit" :disabled="problem.isRemote" />
            </el-form-item>
            <el-form-item class="overview-field" :label="$t('m.Level')" required>
              <el-select class="difficulty-select" v-model="problem.difficulty">
                <el-option :label="getLevelName(key)" :value="parseInt(key)" v-for="(value, key, index) in PROBLEM_LEVEL" :key="index" />
              </el-select>
            </el-form-item>
          </div>
          <div class="problem-quick-settings">
            <el-form-item class="quick-field quick-auth" :label="$t('m.Auth')">
              <el-select v-model="problem.auth" size="small">
                <el-option :label="$t('m.Public_Problem')" :value="1" />
                <el-option :label="$t('m.Private_Problem')" :value="2" />
                <el-option :label="$t('m.Contest_Problem')" :value="3" />
              </el-select>
            </el-form-item>
            <el-form-item class="quick-field quick-type" :label="$t('m.Type')">
              <el-select v-model="problem.type" :disabled="disableRuleType || problem.isRemote" @change="problemTypeChange">
                <el-option label="ACM" :value="0" />
                <el-option label="OI" :value="1" />
              </el-select>
            </el-form-item>
          </div>
        </div>

        <div class="problem-secondary-settings">
          <el-form-item class="source-setting secondary-middle-field" :label="$t('m.Source')">
            <el-input :placeholder="$t('m.Source')" v-model="problem.source" />
          </el-form-item>
          <el-form-item class="secondary-middle-field language-setting" :label="$t('m.Languages')" :error="error.languages" required>
            <el-select v-model="problemLanguages" multiple collapse-tags filterable placeholder="选择允许运行的语言" class="language-select">
              <el-option v-for="lang in allLanguage" :key="lang.name" :label="lang.name" :value="lang.name" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="!problem.isRemote" class="secondary-toggle-field" :label="$t('m.Auto_Remove_the_Blank_at_the_End_of_Code')">
            <el-switch v-model="problem.isRemoveEndBlank" />
          </el-form-item>
          <el-form-item class="secondary-toggle-field" :label="$t('m.Publish_the_Judging_Result_of_Test_Data')">
            <el-switch v-model="problem.openCaseResult" />
          </el-form-item>
          <el-form-item class="secondary-toggle-field" :label="$t('m.Code_Shareable')">
            <el-switch v-model="problem.codeShare" active-text="" inactive-text="" />
          </el-form-item>
          <el-form-item class="tags-setting secondary-toggle-field" :label="$t('m.Tags')">
            <el-tag v-for="tag in problemTags" :key="tag.name" closable :close-transition="false" size="small"
              @close="closeTag(tag.name)">{{ tag.name }}</el-tag>
            <el-autocomplete v-if="inputVisible" size="mini" class="input-new-tag" v-model="tagInput"
              :trigger-on-focus="true" @keyup.enter.native="addTag" @click="selectTag" @select="addTag"
              :fetch-suggestions="querySearch" />
            <el-tooltip v-else effect="dark" :content="$t('m.Add')" placement="top">
              <el-button class="button-new-tag" size="small" @click="inputVisible = true" icon="el-icon-plus" />
            </el-tooltip>
          </el-form-item>
        </div>

        <el-row
          :gutter="20"
          v-if="contestID"
        >
          <el-col
            :md="12"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Display_Title')"
              required
            >
              <el-input
                :placeholder="$t('m.Contest_Display_Title')"
                v-model="contestProblem.displayTitle"
              ></el-input>
            </el-form-item>
          </el-col>

          <el-col
            :md="12"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Display_ID')"
              required
            >
              <el-input
                :placeholder="$t('m.Contest_Display_ID')"
                v-model="contestProblem.displayId"
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item
              class="compact-description"
              prop="description"
              :label="$t('m.Description')"
              required
            >
              <Editor :value.sync="problem.description"></Editor>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :md="12" :xs="24">
            <el-form-item
              class="compact-description"
              prop="input_description"
              :label="$t('m.Input')"
              required
            >
              <Editor :value.sync="problem.input"></Editor>
            </el-form-item>
          </el-col>
          <el-col :md="12" :xs="24">
            <el-form-item
              class="compact-description"
              prop="output_description"
              :label="$t('m.Output')"
              required
            >
              <Editor :value.sync="problem.output"></Editor>
            </el-form-item>
          </el-col>
        </el-row>

        </div>

        <div v-show="editorStep === 1">
          <el-alert class="test-data-tip" title="测试数据按变更导入"
            description="上传新 ZIP 或修改手工测试点时才会同步判题机；普通保存不会重复传输已同步的大数据包。"
            type="info" :closable="false" show-icon />
          <div class="panel-title home-title">
            {{ $t('m.Problem_Examples') }}
            <el-popover
              placement="right"
              trigger="hover"
            >
              <p>
                {{ $t('m.Problem_Examples_Desc') }}
              </p>
              <i
                slot="reference"
                class="el-icon-question"
              ></i>
            </el-popover>
          </div>
          <el-form-item
            v-for="(example, index) in problem.examples"
            :key="'example' + index"
          >
            <Accordion
              :title="$t('m.Problem_Example') + (index + 1)"
              :isOpen="example.isOpen ? true : false"
              :index="index"
              @changeVisible="changeExampleVisible"
            >
              <el-button
                type="danger"
                size="small"
                icon="el-icon-delete"
                slot="header"
                @click="deleteExample(index)"
              >
                {{ $t('m.Delete') }}
              </el-button>
              <el-row :gutter="20">
                <el-col
                  :xs="24"
                  :md="12"
                >
                  <el-form-item
                    :label="$t('m.Example_Input')"
                    required
                  >
                    <el-input
                      :rows="5"
                      type="textarea"
                      :placeholder="$t('m.Example_Input')"
                      v-model="example.input"
                      style="white-space: pre-line"
                    >
                    </el-input>
                  </el-form-item>
                </el-col>
                <el-col
                  :xs="24"
                  :md="12"
                >
                  <el-form-item
                    :label="$t('m.Example_Output')"
                    required
                  >
                    <el-input
                      :rows="5"
                      type="textarea"
                      :placeholder="$t('m.Example_Output')"
                      v-model="example.output"
                    >
                    </el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="24">
                  <el-form-item label="样例解释（可选）">
                    <SampleExplanationEditor v-model="example.explanation" />
                  </el-form-item>
                </el-col>
              </el-row>
            </Accordion>
          </el-form-item>
        </div>

        <div v-show="editorStep === 1" class="add-example-btn">
          <el-button
            class="add-examples"
            @click="addExample()"
            icon="el-icon-plus"
            type="small"
          >{{ $t('m.Add_Example') }}
          </el-button>
        </div>

        <el-collapse v-show="editorStep === 0" v-model="advancedConfigOpen" class="advanced-config">
          <el-collapse-item title="高级判题与代码配置" name="judge">
        <div v-show="editorStep === 0" class="advanced-config-body">
          <div class="advanced-option-grid">
            <div v-if="!problem.isRemote" class="advanced-option extra-file-option">
              <div class="advanced-label">
                {{ $t('m.Judge_Extra_File') }}
                <el-popover placement="right" trigger="hover">
                  <p>{{ $t('m.Judge_Extra_File_Tips1') }}</p>
                  <p>{{ $t('m.Judge_Extra_File_Tips2') }}</p>
                  <i slot="reference" class="el-icon-question"></i>
                </el-popover>
              </div>
              <div class="inline-checks">
                <el-checkbox v-model="addUserExtraFile">{{ $t('m.User_Program') }}</el-checkbox>
                <el-checkbox v-model="addJudgeExtraFile">{{ $t('m.SPJ_Or_Interactive_Program') }}</el-checkbox>
              </div>
            </div>

            <div v-if="!problem.isRemote" class="advanced-option">
              <div class="advanced-label">{{ $t('m.Read_Write_Mode') }}</div>
              <el-select v-model="problem.isFileIO">
                <el-option :label="$t('m.Standard_IO')" :value="false" />
                <el-option :label="$t('m.File_IO')" :value="true" />
              </el-select>
            </div>

            <div v-if="!problem.isRemote" class="advanced-option">
              <div class="advanced-label">
                {{ $t('m.Judge_Mode') }}
                <el-popover placement="right" trigger="hover">
                  <p>1. {{ $t('m.General_Judge_Mode_Tips') }}</p>
                  <p>2. {{ $t('m.Special_Judge_Mode_Tips') }}</p>
                  <p>3. {{ $t('m.Interactive_Judge_Mode_Tips') }}</p>
                  <i slot="reference" class="el-icon-question"></i>
                </el-popover>
              </div>
              <el-select v-model="problem.judgeMode" @change="switchMode">
                <el-option label="普通判题" value="default" />
                <el-option label="特殊判题（SPJ）" value="spj" />
                <el-option label="交互判题" value="interactive" />
              </el-select>
              <div v-if="error.spj" class="advanced-error">{{ error.spj }}</div>
            </div>

            <div class="advanced-option template-option">
              <div class="advanced-label">{{ $t('m.Code_Template') }}</div>
              <el-select v-model="selectedTemplateLanguages" multiple collapse-tags filterable
                placeholder="选择模板语言">
                <el-option v-for="(v, k) in codeTemplate" :key="'template-option' + k"
                  :label="k" :value="k" />
              </el-select>
            </div>
          </div>

          <el-row v-if="!problem.isRemote && (addUserExtraFile || addJudgeExtraFile)" :gutter="16" class="advanced-details">
            <el-col v-if="addUserExtraFile" :md="12" :xs="24">
              <AddExtraFile :files.sync="userExtraFile" type="user"
                @upsertFile="upsertFile" @deleteFile="deleteFile" />
            </el-col>
            <el-col v-if="addJudgeExtraFile" :md="12" :xs="24">
              <AddExtraFile :files.sync="judgeExtraFile" type="judge"
                @upsertFile="upsertFile" @deleteFile="deleteFile" />
            </el-col>
          </el-row>

          <el-row v-if="!problem.isRemote && problem.isFileIO" :gutter="16" class="advanced-details">
            <el-col :md="12" :xs="24">
              <el-input v-model="problem.ioReadFileName" size="small">
                <template slot="prepend">{{ $t('m.Input_File_Name') }}</template>
              </el-input>
            </el-col>
            <el-col :md="12" :xs="24">
              <el-input v-model="problem.ioWriteFileName" size="small">
                <template slot="prepend">{{ $t('m.Output_File_Name') }}</template>
              </el-input>
            </el-col>
          </el-row>

          <el-form-item v-if="!problem.isRemote && problem.judgeMode != 'default'" class="advanced-details">
            <Accordion :title="problem.judgeMode == 'spj' ? $t('m.Special_Judge_Code') : $t('m.Interactive_Judge_Code')">
              <template slot="header">
                <span class="spj-header-label">{{ problem.judgeMode == 'spj' ? $t('m.SPJ_Language') : $t('m.Interactive_Language') }}：</span>
                <el-select v-model="problem.spjLanguage" class="spj-language-select">
                  <el-option v-for="lang in allSpjLanguage" :key="lang.name" :label="lang.name" :value="lang.name" />
                </el-select>
                <el-button type="primary" size="small" icon="el-icon-folder-checked"
                  @click="compileSPJ" :loading="loadingCompile" class="compile-button">{{ $t('m.Compile') }}</el-button>
              </template>
              <code-mirror v-model="problem.spjCode" :mode="spjMode"></code-mirror>
            </Accordion>
          </el-form-item>

          <div v-if="selectedTemplateLanguages.length" class="advanced-details">
            <div v-for="(v, k) in codeTemplate" :key="'template' + k" v-show="v.status" class="template-editor">
              <div class="template-title">{{ k }}</div>
              <code-mirror v-model="v.code" :mode="v.mode" />
            </div>
          </div>
        </div>
          </el-collapse-item>
        </el-collapse>

        <el-row
          :gutter="20"
          v-if="!problem.isRemote"
          v-show="editorStep === 1"
        >
          <div class="panel-title home-title">
            {{ $t('m.Judge_Samples') }}
            <el-popover
              placement="right"
              trigger="hover"
            >
              <p>{{ $t('m.Sample_Tips') }}</p>
              <i
                slot="reference"
                class="el-icon-question"
              ></i>
            </el-popover>
          </div>

          <el-form-item
            required
          >
            <el-select v-model="problem.judgeCaseMode" @change="switchJudgeCaseMode" class="mode-select">
              <el-option :label="problem.type == 1 ? $t('m.OI_Judge_Case_Default_Mode') : $t('m.ACM_Judge_Case_Default_Mode')" :value="JUDGE_CASE_MODE.DEFAULT" />
              <el-option v-if="problem.type == 1" :label="$t('m.Judge_Case_Subtask_Lowest_Mode')" :value="JUDGE_CASE_MODE.SUBTASK_LOWEST" />
              <el-option v-if="problem.type == 1" :label="$t('m.Judge_Case_Subtask_Average_Mode')" :value="JUDGE_CASE_MODE.SUBTASK_AVERAGE" />
              <el-option v-if="problem.type != 1" :label="$t('m.Judge_Case_Ergodic_Without_Error_Mode')" :value="JUDGE_CASE_MODE.ERGODIC_WITHOUT_ERROR" />
            </el-select>
          </el-form-item>

          <el-form-item required>
            <el-switch
              v-model="problem.isUploadCase"
              :active-text="$t('m.Use_Upload_File')"
              :inactive-text="$t('m.Use_Manual_Input')"
              style="margin: 10px 0"
            >
            </el-switch>
          </el-form-item>

          <div v-show="problem.isUploadCase">
            <el-col :span="24">
              <el-form-item :error="error.testcase">
                <el-upload
                  :action="uploadFileUrl+'?mode='+problem.judgeCaseMode"
                  name="file"
                  :show-file-list="true"
                  :on-success="uploadSucceeded"
                  :on-error="uploadFailed"
                >
                  <el-button
                    size="small"
                    type="primary"
                    icon="el-icon-upload"
                  >{{ $t('m.Choose_File') }}</el-button>
                </el-upload>
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <vxe-table
                ref="xTable"
                stripe
                auto-resize
                :data="problem.testCaseScore"
                :sort-config="{trigger: 'cell', 
                defaultSort: {field: 'groupNum', order: 'asc'}, 
                orders: ['desc', 'asc', null],
                sortMethod: customSortMethod}"
                align="center"
              >
                <vxe-table-column
                  field="index"
                  title="#"
                  width="60"
                ></vxe-table-column>
                <vxe-table-column
                  field="input"
                  :title="$t('m.Sample_Input_File')"
                  min-width="100"
                >
                </vxe-table-column>
                <vxe-table-column
                  field="output"
                  :title="$t('m.Sample_Output_File')"
                  min-width="100"
                >
                </vxe-table-column>
                <vxe-table-column
                  v-if="problem.judgeCaseMode == JUDGE_CASE_MODE.SUBTASK_LOWEST 
                    || problem.judgeCaseMode == JUDGE_CASE_MODE.SUBTASK_AVERAGE"
                  field="groupNum"
                  :title="$t('m.Sample_Group_Num')"
                  sortable
                  min-width="100"
                >
                  <template v-slot="{ row }">
                    <el-input
                      size="small"
                      v-model="row.groupNum"
                      @change="sortTestCaseList"
                      type="number"
                    ></el-input>
                  </template>
                </vxe-table-column>
                <vxe-table-column
                  field="score"
                  :title="$t('m.Score')"
                  v-if="problem.type == 1"
                  min-width="100"
                >
                  <template v-slot="{ row }">
                    <el-input
                      size="small"
                      :placeholder="$t('m.Score')"
                      v-model="row.score"
                      :disabled="problem.type != 1"
                      type="number"
                    >
                    </el-input>
                  </template>
                </vxe-table-column>
              </vxe-table>
            </el-col>
          </div>

          <div v-show="!problem.isUploadCase">
            <el-form-item
              v-for="(sample, index) in problemSamples"
              :key="'sample' + index"
            >
              <Accordion
                :title="$t('m.Problem_Sample') + (sample.index)"
                :isOpen="sample.isOpen ? true : false"
                :index="index"
                @changeVisible="changeSampleVisible"
              >
                <el-button
                  type="danger"
                  size="small"
                  icon="el-icon-delete"
                  slot="header"
                  @click="deleteSample(index)"
                >
                  {{ $t('m.Delete') }}
                </el-button>
                <el-row :gutter="20">
                  <el-col
                    :xs="24"
                    :md="12"
                  >
                    <el-form-item
                      :label="$t('m.Sample_Input')"
                      required
                    >
                      <el-input
                        :rows="5"
                        type="textarea"
                        :placeholder="$t('m.Sample_Input')"
                        v-model="sample.input"
                      >
                      </el-input>
                    </el-form-item>
                  </el-col>
                  <el-col
                    :xs="24"
                    :md="12"
                  >
                    <el-form-item
                      :label="$t('m.Sample_Output')"
                      required
                    >
                      <el-input
                        :rows="5"
                        type="textarea"
                        :placeholder="$t('m.Sample_Output')"
                        v-model="sample.output"
                      >
                      </el-input>
                    </el-form-item>
                  </el-col>
                  <el-col
                    :span="24"
                    v-if="problem.type == 1"
                  >
                    <el-form-item :label="$t('m.Score')">
                      <el-input
                        type="number"
                        size="small"
                        :placeholder="$t('m.Score')"
                        v-model="sample.score"
                      >
                      </el-input>
                    </el-form-item>
                  </el-col>
                  <el-col
                    :span="24"
                    v-show="problem.judgeCaseMode == JUDGE_CASE_MODE.SUBTASK_LOWEST 
                    || problem.judgeCaseMode == JUDGE_CASE_MODE.SUBTASK_AVERAGE"
                  >
                    <el-form-item :label="$t('m.Sample_Group_Num')">
                      <el-input
                        type="number"
                        size="small"
                        :placeholder="$t('m.Sample_Group_Num')"
                        v-model="sample.groupNum"
                        @change="sortManualProblemSampleList"
                      >
                      </el-input>
                    </el-form-item>
                  </el-col>
                </el-row>
              </Accordion>
            </el-form-item>

            <div class="add-sample-btn">
              <el-button
                class="add-samples"
                @click="addSample()"
                icon="el-icon-plus"
                type="small"
              >{{ $t('m.Add_Sample') }}
              </el-button>
            </div>
          </div>
        </el-row>

        <div class="wizard-actions" :class="{ 'single-action': editorStep === 0 }">
          <el-button v-if="editorStep === 0" type="primary" @click="nextToDataStep">下一步：样例与测试数据</el-button>
          <template v-else>
            <el-button @click="editorStep = 0">返回题面配置</el-button>
            <el-button type="primary" :loading="saving" @click.native="submit()">保存并进入验题</el-button>
          </template>
        </div>
      </el-form>
      <ProblemValidationStep
        v-if="editorStep === 2 && verificationPid"
        :pid="verificationPid"
        :languages="problemLanguages"
        @back="editorStep = 1"
        @passed="finishVerification"
      />
    </el-card>
  </div>
</template>

<script>
import utils from "@/common/utils";
import { mapGetters } from "vuex";
import api from "@/common/api";
import myMessage from "@/common/message";
import { PROBLEM_LEVEL, JUDGE_CASE_MODE } from "@/common/constants";
const Editor = () => import("@/components/admin/Editor.vue");
const Accordion = () => import("@/components/admin/Accordion.vue");
const AddExtraFile = () => import("@/components/admin/AddExtraFile.vue");
const CodeMirror = () => import("@/components/admin/CodeMirror.vue");
const SampleExplanationEditor = () =>
  import("@/components/admin/SampleExplanationEditor.vue");
const ProblemValidationStep = () =>
  import("@/components/admin/ProblemValidationStep.vue");
export default {
  name: "Problem",
  components: {
    Accordion,
    AddExtraFile,
    CodeMirror,
    Editor,
    SampleExplanationEditor,
    ProblemValidationStep,
  },
  data() {
    return {
      rules: {
        title: {
          required: true,
          message: "Title is required",
          trigger: "blur",
        },
        input_description: {
          required: true,
          message: "Input Description is required",
          trigger: "blur",
        },
        output_description: {
          required: true,
          message: "Output Description is required",
          trigger: "blur",
        },
      },
      backPath: "",
      loadingCompile: false,
      mode: "", // 该题目是编辑或者创建
      contest: {},
      codeTemplate: {},
      pid: null, // 题目id，如果为创建模式则为null
      contestID: null, // 比赛id
      contestProblem: {
        displayId: null,
        displayTitle: null,
        cid: null,
        pid: null,
      }, // 比赛题目的相关属性
      problem: {
        id: null,
        title: "",
        problemId: "",
        description: "",
        input: "",
        output: "",
        timeLimit: 1000,
        memoryLimit: 256,
        stackLimit: 128,
        difficulty: 0,
        auth: 1,
        codeShare: true,
        examples: [], // 题面上的样例输入输出
        spjLanguage: "",
        spjCode: "",
        spjCompileOk: false,
        uploadTestcaseDir: "",
        testCaseScore: [],
        isRemote: false,
        isUploadCase: true,
        type: 0,
        hint: "",
        source: "",
        cid: null,
        isRemoveEndBlank:false,
        openCaseResult: true,
        judgeMode: "default",
        judgeCaseMode: "default",
        userExtraFile: "",
        judgeExtraFile: "",
        isFileIO: false,
        ioReadFileName: null,
        ioWriteFileName: null
      },
      problemTags: [], //指定问题的标签列表
      problemLanguages: [], //指定问题的编程语言列表
      problemSamples: [], // 判题机使用的样例
      problemCodeTemplate: [],
      reProblem: {},
      testCaseUploaded: false,
      allLanguage: [], //所有编程语言
      allSpjLanguage: [], // 所有可以用特殊判题的语言
      allTags: [],
      allTagsTmp: [],
      inputVisible: false,
      tagInput: "",
      title: "",
      spjMode: "",
      disableRuleType: false,
      routeName: "",
      uploadTestcaseDir: "",
      uploadFileUrl: "",
      error: {
        tags: "",
        spj: "",
        languages: "",
        testCase: "",
      },
      PROBLEM_LEVEL: {},
      JUDGE_CASE_MODE: {},
      spjRecord: {
        spjCode: "",
        spjLanguage: "",
      },
      addUserExtraFile: false,
      addJudgeExtraFile: false,
      userExtraFile: null,
      judgeExtraFile: null,
      judgeExtraFileRecord: null,
      judgeCaseModeRecord: "default",
      sampleIndex: 1,
      verificationPid: null,
      needsVerification: false,
      editorStep: 0,
      saving: false,
      savedPid: null,
      advancedConfigOpen: ['judge'],
    };
  },
  mounted() {
    this.PROBLEM_LEVEL = Object.assign({}, PROBLEM_LEVEL);
    this.JUDGE_CASE_MODE = Object.assign({}, JUDGE_CASE_MODE);
    this.routeName = this.$route.name;
    let contestID = this.$route.params.contestId;
    this.uploadFileUrl = "/api/file/upload-testcase-zip";
    if (
      this.routeName === "admin-edit-problem" ||
      this.routeName === "admin-edit-contest-problem"
    ) {
      this.mode = "edit";
    } else {
      this.mode = "add";
    }
    this.applyRequestedStep();
    api
      .admin_getAllProblemTagList("ALL")
      .then((res) => {
        this.allTags = res.data.data;
        for (let tag of res.data.data) {
          this.allTagsTmp.push({ value: tag.name, oj: tag.oj });
        }
      })
      .catch(() => {});
    api.getLanguages(this.$route.params.problemId, false).then((res) => {
      let allLanguage = res.data.data;
      this.allLanguage = allLanguage;
      for (let i = 0; i < allLanguage.length; i++) {
        if (allLanguage[i].isSpj == true) {
          this.allSpjLanguage.push(allLanguage[i]);
        }
      }
      this.problem = this.reProblem = {
        id: null,
        problemId: "",
        title: "",
        description: "",
        input: "",
        output: "",
        timeLimit: 1000,
        memoryLimit: 256,
        stackLimit: 128,
        difficulty: 0,
        auth: 1,
        codeShare: true,
        examples: [],
        spjLanguage: "",
        spjCode: "",
        spjCompileOk: false,
        isUploadCase: true,
        uploadTestcaseDir: "",
        testCaseScore: [],
        contestProblem: {},
        type: 0,
        hint: "",
        source: "",
        cid: null,
        isRemoveEndBlank:false,
        openCaseResult: true,
        judgeMode: "default",
        judgeCaseMode: "default",
        userExtraFile: null,
        judgeExtraFile: null,
        isFileIO: false,
        ioReadFileName: null,
        ioWriteFileName: null
      };

      this.contestID = contestID;
      if (contestID) {
        this.problem.cid = this.reProblem.cid = contestID;
        this.problem.auth = this.reProblem.auth = 3;
        this.disableRuleType = true;
        api.admin_getContest(contestID).then((res) => {
          this.problem.type = this.reProblem.type = res.data.data.type;
          this.contest = res.data.data;
        });
      }
      this.problem.spjLanguage = "C";
      this.init();
    });
  },
  watch: {
    $route() {
      this.routeName = this.$route.name;
      if (
        this.routeName === "admin-edit-problem" ||
        this.routeName === "admin-edit-contest-problem"
      ) {
        this.mode = "edit";
      } else {
        this.mode = "add";
      }
      this.$refs.form.resetFields();
      this.problem = this.reProblem;
      this.problemTags = []; //指定问题的标签列表
      this.problemLanguages = []; //指定问题的编程语言列表
      this.problemSamples = [];
      this.problemCodeTemplate = [];
      this.codeTemplate = [];
      this.editorStep = 0;
      this.applyRequestedStep();
      this.savedPid = null;
      this.init();
    },

    problemLanguages(newVal) {
      let data = {};
      // use deep copy to avoid infinite loop
      let languages = JSON.parse(JSON.stringify(newVal)).sort();
      for (let item of languages) {
        if (this.codeTemplate[item] === undefined) {
          let langConfig = this.allLanguage.find((lang) => {
            return lang.name === item;
          });
          let codeTemp;
          let problemCodeTemplate = this.problemCodeTemplate;
          if (problemCodeTemplate) {
            codeTemp = problemCodeTemplate.find((temp) => {
              return temp.lid == langConfig.id;
            });
          }
          if (codeTemp == undefined) {
            data[item] = {
              id: null,
              status: false,
              code: langConfig.codeTemplate,
              mode: langConfig.contentType,
            };
          } else {
            data[item] = {
              id: codeTemp.id,
              status: true,
              code: codeTemp.code,
              mode: langConfig.contentType,
            };
          }
        } else {
          data[item] = this.codeTemplate[item];
        }
      }
      this.codeTemplate = data;
    },

    "problem.spjLanguage"(newVal) {
      if (this.allSpjLanguage.length && this.problem.judgeMode != "default") {
        this.spjMode = this.allSpjLanguage.find((item) => {
          return item.name == this.problem.spjLanguage && item.isSpj == true;
        })["contentType"];
      }
    },
  },
  methods: {
    applyRequestedStep() {
      const requestedStep = Number(this.$route.query && this.$route.query.step);
      if (requestedStep !== 2 || this.mode !== "edit") return;
      this.editorStep = 2;
      this.verificationPid = this.$route.params.problemId;
    },
    nextToDataStep() {
      const required = [
        [this.problem.problemId, this.$i18n.t("m.Problem_Display_ID")],
        [this.problem.title, this.$i18n.t("m.Title")],
        [this.problem.description, this.$i18n.t("m.Description")],
        [this.problem.input, this.$i18n.t("m.Input")],
        [this.problem.output, this.$i18n.t("m.Output")],
      ];
      const missing = required.find(([value]) => !value || !String(value).trim());
      if (missing) return myMessage.error(missing[1] + " " + this.$i18n.t("m.is_required"));
      if (!this.problemLanguages.length) return myMessage.error(this.$i18n.t("m.Language") + " " + this.$i18n.t("m.is_required"));
      this.editorStep = 1;
      this.$nextTick(() => window.scrollTo(0, 0));
    },
    init() {
      this.sampleIndex = 1;
      if (this.mode === "edit") {
        this.pid = this.$route.params.problemId;
        this.backPath = this.$route.query.back;
        this.title = this.$i18n.t("m.Edit_Problem");
        let funcName = {
          "admin-edit-problem": "admin_getProblem",
          "admin-edit-contest-problem": "admin_getContestProblem",
        }[this.routeName];
        api[funcName](this.pid).then((problemRes) => {
          let data = problemRes.data.data;
          data.spjCompileOk = false;
          data.uploadTestcaseDir = "";
          data.testCaseScore = [];
          if (!data.spjCode) {
            data.spjCode = "";
          }
          data.spjLanguage = data.spjLanguage || "C";
          this.spjRecord.spjLanguage = data.spjLanguage;
          this.spjRecord.spjCode = data.spjCode;
          this.judgeExtraFileRecord = data.judgeExtraFile || null;
          this.judgeCaseModeRecord = data.judgeCaseMode;
          this.problem = data;
          this.problem["examples"] = utils.stringToExamples(data.examples);
          if (this.problem["examples"].length > 0) {
            this.problem["examples"][0]["isOpen"] = true;
          }
          this.testCaseUploaded = true;
          if (this.problem.userExtraFile) {
            this.addUserExtraFile = true;
            this.userExtraFile = JSON.parse(this.problem.userExtraFile);
          }
          if (this.problem.judgeExtraFile) {
            this.addJudgeExtraFile = true;
            this.judgeExtraFile = JSON.parse(this.problem.judgeExtraFile);
          }
          api
            .admin_getProblemCases(this.pid, this.problem.isUploadCase)
            .then((res) => {
              if (this.problem.isUploadCase) {
                this.problem.testCaseScore = res.data.data;
                this.problem.testCaseScore.forEach((item, index) => {
                  item.index = index + 1;
                });
                if(this.$refs.xTable != undefined){
                  this.$refs.xTable.sort("groupNum", "asc");
                }
              } else {
                this.problemSamples = res.data.data;
                if (
                  this.problemSamples != null &&
                  this.problemSamples.length > 0
                ) {
                  this.problemSamples[0]["isOpen"] = true;
                  this.problemSamples.forEach((item, index) => {
                    item.index = index + 1;
                  });
                  this.sampleIndex = this.problemSamples.length + 1;
                }
              }
            });
        });
        if (funcName === "admin_getContestProblem") {
          api
            .admin_getContestProblemInfo(this.pid, this.contestID)
            .then((res) => {
              this.contestProblem = res.data.data;
            });
        }
        this.getProblemCodeTemplateAndLanguage();

        api.admin_getProblemTags(this.pid).then((res) => {
          this.problemTags = res.data.data;
        });
      } else {
        this.addExample();
        this.testCaseUploaded = false;
        this.title = this.$i18n.t("m.Create_Problem");
        for (let item of this.allLanguage) {
          this.problemLanguages.push(item.name);
        }
      }
    },

    async getProblemCodeTemplateAndLanguage() {
      const that = this;
      await api.getProblemCodeTemplate(that.pid).then((res) => {
        that.problemCodeTemplate = res.data.data;
      });
      api.getProblemLanguages(that.pid).then((res) => {
        let Languages = res.data.data;
        for (let i = 0; i < Languages.length; i++) {
          that.problemLanguages.push(Languages[i].name);
        }
      });
    },

    switchMode(mode) {
      let modeName = "General_Judge";
      let modeTips = "General_Judge_Mode_Tips";
      if (mode == "spj") {
        modeName = "Special_Judge";
        modeTips = "Special_Judge_Mode_Tips";
      } else if (mode == "interactive") {
        modeName = "Interactive_Judge";
        modeTips = "Interactive_Judge_Mode_Tips";
      }
      const h = this.$createElement;
      this.$msgbox({
        title: this.$i18n.t("m." + modeName),
        message: h("div", null, [
          h(
            "p",
            { style: "text-align: center;font-weight:bolder;color:red" },
            this.$i18n.t("m.Change_Judge_Mode")
          ),
          h("br", null, null),
          h(
            "p",
            { style: "font-weight:bolder" },
            this.$i18n.t("m." + modeTips)
          ),
        ]),
      });
    },
    querySearch(queryString, cb) {
      var ojName = "ME";
      if (this.problem.isRemote) {
        ojName = this.problem.problemId.split("-")[0];
      }
      var restaurants = this.allTagsTmp.filter((item) => item.oj == ojName);
      var results = queryString
        ? restaurants.filter(
            (item) =>
              item.value.toLowerCase().indexOf(queryString.toLowerCase()) >= 0
          )
        : restaurants;
      cb(results);
    },
    changeContent(newVal) {
      this.announcement.content = newVal;
    },
    getLevelName(difficulty) {
      return utils.getLevelName(difficulty);
    },

    selectTag(item) {
      for (var i = 0; i < this.problemTags.length; i++) {
        if (this.problemTags[i].name == item.value) {
          myMessage.warning(this.$i18n.t("m.Add_Tag_Error"));
          this.tagInput = "";
          return;
        }
      }
      this.tagInput = item.value;
    },
    addTag(item) {
      let newTag = {
        name: this.tagInput,
      };
      if (this.tagInput) {
        for (var i = 0; i < this.problemTags.length; i++) {
          if (this.problemTags[i].name == this.tagInput) {
            myMessage.warning(this.$i18n.t("m.Add_Tag_Error"));
            this.tagInput = "";
            return;
          }
        }
        this.problemTags.push(newTag);
        this.inputVisible = false;
        this.tagInput = "";
      }
    },

    // 根据tag name从题目的tags列表中移除
    closeTag(tag) {
      this.problemTags.splice(
        this.problemTags.map((item) => item.name).indexOf(tag),
        1
      );
    },

    deleteFile(type, name) {
      if (type == "user") {
        this.$delete(this.userExtraFile, name);
      } else {
        this.$delete(this.judgeExtraFile, name);
      }
    },

    upsertFile(type, name, oldname, content) {
      if (type == "user") {
        if (oldname && oldname != name) {
          this.$delete(this.userExtraFile, oldname);
        }
        if (!this.userExtraFile) {
          this.userExtraFile = {};
        }
        this.userExtraFile[name] = content;
      } else {
        if (oldname && oldname != name) {
          this.$delete(this.judgeExtraFile, name);
        }
        if (!this.judgeExtraFile) {
          this.judgeExtraFile = {};
        }
        this.judgeExtraFile[name] = content;
      }
    },

    problemTypeChange(type) {
      if (type == 1) {
        let length = this.problemSamples.length;
        let aver = parseInt(100 / length);
        let add_1_num = 100 - aver * length;
        for (let i = 0; i < length; i++) {
          if (i >= length - add_1_num) {
            this.problemSamples[i].score = aver + 1;
          } else {
            this.problemSamples[i].score = aver;
          }
        }
      }
    },

    // 添加题目样例
    addExample() {
      this.problem.examples.push({ input: "", output: "", explanation: "", isOpen: true });
    },
    changeExampleVisible(index, isOpen) {
      this.problem.examples[index]["isOpen"] = isOpen;
    },
    // 添加判题机的测试样例
    addSample() {
      let len = this.sampleIndex;
      if (this.mode === "edit") {
        this.problemSamples.push({
          input: "",
          output: "",
          score: this.problem.type == 0 ? null : 0,
          groupNum: this.problem.type == 0 ? null : len,
          pid: this.pid,
          isOpen: true,
          index: len
        });
      } else {
        this.problemSamples.push({
          input: "",
          output: "",
          score: this.problem.type == 0 ? null : 0,
          groupNum: this.problem.type == 0 ? null : len,
          pid: this.pid,
          isOpen: true,
          index: len
        });
      }
      this.sampleIndex = len + 1;
      this.sortManualProblemSampleList();
    },
    //根据下标删除特定的题目样例
    deleteExample(index) {
      this.problem.examples.splice(index, 1);
    },
    //根据下标删除特定的判题机测试样例
    deleteSample(index) {
      this.problemSamples.splice(index, 1);
    },
    changeSampleVisible(index, isOpen) {
      this.problemSamples[index]["isOpen"] = isOpen;
    },
    uploadSucceeded(response) {
      if (response.status != 200) {
        myMessage.error(response.msg);
        this.testCaseUploaded = false;
        return;
      }
      myMessage.success(this.$i18n.t("m.Upload_Testcase_Successfully"));
      let fileList = response.data.fileList;
      let averSorce = parseInt(100 / fileList.length);
      let add_1_num = 100 - averSorce * fileList.length;
      for (let i = 0; i < fileList.length; i++) {
        if (averSorce) {
          if (i >= fileList.length - add_1_num) {
            fileList[i].score = averSorce + 1;
          } else {
            fileList[i].score = averSorce;
          }
        }
        if (!fileList[i].output) {
          fileList[i].output = "-";
        }
        fileList[i].pid = this.problem.id;
      }
      this.problem.testCaseScore = fileList;
      this.problem.testCaseScore.forEach((item, index) => {
        item.index = index + 1;
      });
      this.testCaseUploaded = true;
      this.problem.uploadTestcaseDir = response.data.fileListDir;
    },
    uploadFailed() {
      myMessage.error(this.$i18n.t("m.Upload_Testcase_Failed"));
    },

    compileSPJ() {
      let data = {
        pid: this.problem.id,
        code: this.problem.spjCode,
        language: this.problem.spjLanguage,
        extraFiles: this.judgeExtraFile,
      };
      this.loadingCompile = true;
      let apiMethodName = "compileSPJ";
      if (this.problem.judgeMode == "interactive") {
        apiMethodName = "compileInteractive";
      }
      api[apiMethodName](data).then(
        (res) => {
          this.loadingCompile = false;
          this.problem.spjCompileOk = true;
          this.error.spj = "";
          myMessage.success(this.$i18n.t("m.Compiled_Successfully"));
        },
        (err) => {
          this.loadingCompile = false;
          this.problem.spjCompileOk = false;
          const h = this.$createElement;
          this.$msgbox({
            title: "Compile Error",
            type: "error",
            message: h("pre", err.data.msg),
            showCancelButton: false,
            closeOnClickModal: false,
            customClass: "dialog-compile-error",
          });
        }
      );
    },
    sortTestCaseList() {
      this.$refs.xTable.clearSort();
      this.$refs.xTable.sort("groupNum", "asc");
    },
    customSortMethod({ data, sortList }) {
      const sortItem = sortList[0];
      const { property, order } = sortItem;
      let list = [];
      list = data.sort(function (a, b) {
        var value1 = a.groupNum,
          value2 = b.groupNum;
        if (value1 === value2) {
          return a.index - b.index;
        }
        if (order == "desc") {
          return value2 - value1;
        } else {
          return value1 - value2;
        }
      });
      return list;
    },
    sortManualProblemSampleList() {
      this.problemSamples = this.problemSamples.sort(function (a, b) {
        var value1 = a.groupNum,
          value2 = b.groupNum;
        if (value1 === value2) {
          return a.index - b.index;
        }
        return value1 - value2;
      });
    },
    submit() {
      if (!this.problem.problemId) {
        myMessage.error(
          this.$i18n.t("m.Problem_Display_ID") +
            " " +
            this.$i18n.t("m.is_required")
        );
        return;
      }

      if (this.contestID) {
        if (!this.contestProblem.displayId) {
          myMessage.error(
            this.$i18n.t("m.Contest_Display_ID") +
              " " +
              this.$i18n.t("m.is_required")
          );
          return;
        }
        if (!this.contestProblem.displayTitle) {
          myMessage.error(
            this.$i18n.t("m.Contest_Display_Title") +
              " " +
              this.$i18n.t("m.is_required")
          );
          return;
        }
      }

      if(this.problem.isFileIO && (!this.problem.ioReadFileName || !this.problem.ioWriteFileName)){
        myMessage.error(this.$i18n.t("m.When_the_read_write_mode_is_File_IO_the_input_file_name_or_output_file_name_cannot_be_empty"));
        return;
      }

      // // 不强制校验题目样例不能为空
      // if (!this.problem.examples.length && !this.problem.isRemote) {
      //   myMessage.error(
      //     this.$i18n.t('m.Problem_Examples') +
      //       ' ' +
      //       this.$i18n.t('m.is_required')
      //   );
      //   return;
      // }

      if (!this.problem.isRemote) {
        // 选择手动输入
        if (!this.problem.isUploadCase) {
          if (!this.problemSamples.length) {
            myMessage.error(
              this.$i18n.t("m.Judge_Samples") +
                " " +
                this.$i18n.t("m.is_required")
            );
            return;
          }

          for (let sample of this.problemSamples) {
            if (!sample.input && !sample.output) {
              myMessage.error(
                this.$i18n.t("m.Sample_Input") +
                  " or " +
                  this.$i18n.t("m.Sample_Output") +
                  " " +
                  this.$i18n.t("m.is_required")
              );
              return;
            }
          }

          // 同时是oi题目，则对应的每个测试样例的io得分不能为空或小于0
          if (this.problem.type == 1) {
            for (let i = 0; i < this.problemSamples.length; i++) {
              if (this.problemSamples[i].score == "") {
                myMessage.error(
                  this.$i18n.t("m.Problem_Sample") +
                    (this.problemSamples[i].index) +
                    " " +
                    this.$i18n.t("m.Score_must_be_an_integer")
                );
                return;
              }
              try {
                if (parseInt(this.problemSamples[i].score) < 0) {
                  myMessage.error(
                    this.$i18n.t("m.Problem_Sample") +
                      (this.problemSamples[i].index) +
                      " " +
                      this.$i18n.t("m.Score_must_be_greater_than_or_equal_to_0")
                  );
                  return;
                }
              } catch (e) {
                myMessage.error(this.$i18n.t("m.Score_must_be_an_integer"));
                return;
              }
              if (
                (this.problem.judgeCaseMode == this.JUDGE_CASE_MODE.SUBTASK_LOWEST 
                  || this.problem.judgeCaseMode == this.JUDGE_CASE_MODE.SUBTASK_AVERAGE
                ) && this.problemSamples[i].groupNum == ""
              ) {
                myMessage.error(
                  this.$i18n.t("m.Problem_Sample") +
                    (this.problemSamples[i].index) +
                    "：" +
                    this.$i18n.t(
                      "m.Non_Default_Judge_Case_Mode_And_Group_Num_IS_NULL"
                    )
                );
                return;
              }
            }
          }
        } else {
          // 选择上传文件
          // 两种情况：create模式是需要校验是否上传成功了，edit模式获取题目数据已经默认为true了，若是edit又重新上传数据，需要校验
          if (!this.testCaseUploaded) {
            this.error.testCase =
              this.$i18n.t("m.Judge_Samples") +
              " " +
              this.$i18n.t("m.is_required");
            myMessage.error(this.error.testCase);
            return;
          }

          // 如果是oi题目，需要检查上传的数据的得分
          if (this.problem.type == 1) {
            let problemSamples = this.problem.testCaseScore;
            for (let i = 0; i < problemSamples.length; i++) {
              if (problemSamples[i].score == "") {
                myMessage.error(
                  this.$i18n.t("m.Problem_Sample") +
                    (i + 1) +
                    " " +
                    this.$i18n.t("m.Score_must_be_an_integer")
                );
                return;
              }
              try {
                if (parseInt(problemSamples[i].score) < 0) {
                  myMessage.error(
                    this.$i18n.t("m.Problem_Sample") +
                      (i + 1) +
                      " " +
                      this.$i18n.t("m.Score_must_be_greater_than_or_equal_to_0")
                  );
                  return;
                }
              } catch (e) {
                myMessage.error(this.$i18n.t("m.Score_must_be_an_integer"));
                return;
              }
              if (
                (this.problem.judgeCaseMode == this.JUDGE_CASE_MODE.SUBTASK_LOWEST 
                  || this.problem.judgeCaseMode == this.JUDGE_CASE_MODE.SUBTASK_AVERAGE
                ) && problemSamples[i].groupNum == ""
              ) {
                myMessage.error(
                  this.$i18n.t("m.Problem_Sample") +
                    (i + 1) +
                    "：" +
                    this.$i18n.t(
                      "m.Non_Default_Judge_Case_Mode_And_Group_Num_IS_NULL"
                    )
                );
                return;
              }
            }
          }
        }
      }
      // 运行题目标签为空
      // if (!this.problemTags.length) {
      //   this.error.tags =
      //     this.$i18n.t('m.Tags') + ' ' + this.$i18n.t('m.is_required');
      //   myMessage.error(this.error.tags);
      //   return;
      // }
      let isChangeModeCode =
        this.spjRecord.spjLanguage != this.problem.spjLanguage ||
        this.spjRecord.spjCode != this.problem.spjCode ||
        this.judgeExtraFileRecord !=
          (Object.keys(this.judgeExtraFile || {}).length
            ? JSON.stringify(this.judgeExtraFile)
            : null);
      if (!this.problem.isRemote) {
        if (this.problem.judgeMode != "default") {
          if (!this.problem.spjCode) {
            this.error.spj =
              this.$i18n.t("m.Spj_Or_Interactive_Code") +
              " " +
              this.$i18n.t("m.is_required");
            myMessage.error(this.error.spj);
          } else if (!this.problem.spjCompileOk && isChangeModeCode) {
            this.error.spj = this.$i18n.t(
              "m.Spj_Or_Interactive_Code_not_Compile_Success"
            );
          }
          if (this.error.spj) {
            myMessage.error(this.error.spj);
            return;
          }
        }
      }

      if (!this.problemLanguages.length) {
        this.error.languages =
          this.$i18n.t("m.Language") + " " + this.$i18n.t("m.is_required");
        myMessage.error(this.error.languages);
        return;
      }

      let funcName = this.savedPid
        ? (this.contestID ? "admin_editContestProblem" : "admin_editProblem")
        : ({
        "admin-create-problem": "admin_createProblem",
        "admin-edit-problem": "admin_editProblem",
        "admin-create-contest-problem": "admin_createContestProblem",
        "admin-edit-contest-problem": "admin_editContestProblem",
      }[this.routeName]);
      // edit contest problem 时, contest_id会被后来的请求覆盖掉
      if (funcName === "editContestProblem") {
        this.problem.cid = this.contest.id;
      }
      if (
        funcName === "admin_createProblem" ||
        funcName === "admin_createContestProblem"
      ) {
        this.problem.author = this.userInfo.username;
      }

      var ojName = "ME";
      if (this.problem.isRemote) {
        ojName = this.problem.problemId.split("-")[0];
      }

      let problemTagList = [];
      if (this.problemTags.length > 0) {
        problemTagList = Object.assign([], this.problemTags);
        for (let i = 0; i < problemTagList.length; i++) {
          //避免后台插入违反唯一性
          for (let tag2 of this.allTags) {
            if (problemTagList[i].name == tag2.name && tag2.oj == ojName) {
              problemTagList[i] = tag2;
              break;
            }
          }
        }
      }
      this.problemCodeTemplate = [];
      let problemLanguageList = Object.assign([], this.problemLanguages); // 深克隆 防止影响
      for (let i = 0; i < problemLanguageList.length; i++) {
        problemLanguageList[i] = { name: problemLanguageList[i] };
        for (let lang of this.allLanguage) {
          if (problemLanguageList[i].name == lang.name) {
            problemLanguageList[i] = lang;
            if (this.codeTemplate[lang.name].status) {
              if(this.codeTemplate[lang.name].code == null 
                || this.codeTemplate[lang.name].code.length == 0){
                  myMessage.error(
                    lang.name +
                      "：" +
                      this.$i18n.t("m.Code_template_of_the_language_cannot_be_empty")
                  );
                  return;
              }
              this.problemCodeTemplate.push({
                id: this.codeTemplate[lang.name].id,
                pid: this.pid,
                code: this.codeTemplate[lang.name].code,
                lid: lang.id,
                status: this.codeTemplate[lang.name].status,
              });
            }
            break;
          }
        }
      }
      let problemDto = {}; // 上传给后台的数据
      if (!this.problem.isRemote) {
        if (this.problem.judgeMode != "default") {
          if (isChangeModeCode) {
            problemDto["changeModeCode"] = true;
          }
        } else {
          // 原本是spj或交互，但现在关闭了
          if (!this.spjRecord.spjCode) {
            problemDto["changeModeCode"] = true;
            this.problem.spjCode = null;
            this.problem.spjLanguage = null;
          }
        }

        if (this.userExtraFile && Object.keys(this.userExtraFile).length != 0) {
          this.problem.userExtraFile = JSON.stringify(this.userExtraFile);
        } else {
          this.problem.userExtraFile = null;
        }

        if (
          this.judgeExtraFile &&
          Object.keys(this.judgeExtraFile).length != 0
        ) {
          this.problem.judgeExtraFile = JSON.stringify(this.judgeExtraFile);
        } else {
          this.problem.judgeExtraFile = null;
        }
      }

      problemDto["problem"] = Object.assign({}, this.problem); // 深克隆
      problemDto.problem.examples = utils.examplesToString(
        this.problem.examples
      ); // 需要转换格式

      problemDto["codeTemplates"] = this.problemCodeTemplate;
      problemDto["tags"] = problemTagList;
      problemDto["languages"] = problemLanguageList;
      problemDto["isUploadTestCase"] = this.problem.isUploadCase;
      problemDto["uploadTestcaseDir"] = this.problem.uploadTestcaseDir;
      problemDto["judgeMode"] = this.problem.judgeMode;

      // 如果选择上传文件，则使用上传后的结果
      if (this.problem.isUploadCase) {
        problemDto["samples"] = this.problem.testCaseScore;
      } else {
        problemDto["samples"] = this.problemSamples;
      }

      if (this.judgeCaseModeRecord != this.problem.judgeCaseMode) {
        problemDto["changeJudgeCaseMode"] = true;
      } else {
        problemDto["changeJudgeCaseMode"] = false;
      }
      this.needsVerification = !this.problem.isRemote;
      this.saving = true;

      api[funcName](problemDto)
        .then((res) => {
          if (
            this.routeName === "admin-create-contest-problem" ||
            this.routeName === "admin-edit-contest-problem"
          ) {
            if (res.data.data) {
              // 新增题目操作 需要使用返回来的pid
              this.contestProblem["pid"] = res.data.data.pid;
              this.contestProblem["cid"] = this.$route.params.contestId;
            }
            api.admin_setContestProblemInfo(this.contestProblem).then(() => {
              this.afterSave(this.contestProblem.pid);
            });
          } else {
            this.afterSave((res.data.data && res.data.data.pid) || this.problem.id);
          }
        })
        .catch(() => {})
        .finally(() => { this.saving = false; });
    },
    afterSave(pid) {
      if (this.needsVerification && pid) {
        this.savedPid = pid;
        this.pid = pid;
        this.problem.id = pid;
        this.verificationPid = pid;
        this.editorStep = 2;
        this.$nextTick(() => window.scrollTo(0, 0));
        return;
      }
      this.finishVerification();
    },
    finishVerification() {
      myMessage.success("success");
      if (this.routeName === "admin-create-contest-problem" || this.routeName === "admin-edit-contest-problem") {
        this.$router.push({
          name: "admin-contest-problem-list",
          params: { contestId: this.$route.params.contestId },
        });
      } else if (this.backPath) {
        this.$router.push({ path: this.backPath });
      } else {
        this.$router.push({ name: "admin-problem-list" });
      }
    },
  },
  computed: {
    ...mapGetters(["userInfo"]),
    selectedTemplateLanguages: {
      get() {
        return Object.keys(this.codeTemplate).filter((key) => this.codeTemplate[key].status);
      },
      set(values) {
        const selected = values || [];
        Object.keys(this.codeTemplate).forEach((key) => {
          this.codeTemplate[key].status = selected.indexOf(key) !== -1;
        });
      },
    },
  },
};
</script>

<style scoped>
/* 页面整体 */
.problem {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.problem /deep/ .el-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.problem /deep/ .el-card__header {
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  padding: 16px 20px;
}

.problem /deep/ .el-card__body {
  padding: 24px;
}

/* 表单项 */
.el-form-item {
  margin-bottom: 18px !important;
}

.problem /deep/ .el-form-item__label {
  padding: 0 0 8px 0 !important;
  line-height: 1.5;
  font-weight: 500;
  color: #606266;
}

/* 步骤条 */
.problem-steps {
  margin-bottom: 28px;
}

/* 表单变量 */
.problem-editor-form {
  --editor-border: #e4e7ed;
}

/* 头部网格：题目ID、标题、限制条件等 */
.problem-header-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 14px 16px;
  margin-bottom: 20px;
}

.problem-overview-grid {
  display: contents;
}

.problem-overview-grid .el-form-item,
.problem-quick-settings .el-form-item {
  margin-bottom: 0 !important;
}

.problem-overview-grid .overview-field:nth-child(1) { grid-column: 1; }
.problem-overview-grid .overview-field:nth-child(2) { grid-column: 2; }
.problem-overview-grid .overview-field:nth-child(3) { grid-column: 3; }
.problem-overview-grid .overview-field:nth-child(4) { grid-column: 4; }
.problem-overview-grid .overview-field:nth-child(5) { grid-column: 5; }
.problem-overview-grid .overview-field:nth-child(6) { grid-column: 6; }

/* 快速设置区域：权限、类型 */
.problem-quick-settings {
  display: contents;
}

.problem-quick-settings .el-select {
  width: 100%;
}

.problem-quick-settings .quick-auth { grid-column: 7; }
.problem-quick-settings .quick-type { grid-column: 8; }

/* 次要设置：来源、开关、标签 */
.problem-secondary-settings {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px 16px;
  align-items: end;
  padding: 16px;
  margin-bottom: 20px;
  background: #f9fafb;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

.problem-secondary-settings .el-form-item {
  margin-bottom: 0 !important;
}

.problem-secondary-settings .el-select,
.problem-secondary-settings .el-input {
  width: 100%;
}

.problem-secondary-settings .secondary-middle-field { grid-column: span 2; }
.problem-secondary-settings .secondary-toggle-field { grid-column: span 1; }

.problem-secondary-settings .el-tag {
  margin-right: 7px;
  margin-top: 4px;
}

/* 编辑器 */
.compact-description .mavon-editor,
.compact-description /deep/ .v-note-wrapper {
  min-height: 300px;
  height: 340px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.compact-description {
  position: relative;
  z-index: 1;
  overflow: hidden;
}

.compact-description /deep/ .v-note-panel,
.compact-description /deep/ .v-note-edit.divarea {
  min-height: 286px;
}

/* 高级配置 */
.advanced-config {
  margin-top: 20px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fff;
}

.advanced-config /deep/ .el-collapse-item__header {
  height: 48px;
  line-height: 48px;
  padding: 0 20px;
  color: #606266;
  font-weight: 600;
  background: #f5f7fa;
}

.advanced-config /deep/ .el-collapse-item__content {
  padding: 20px;
}

.advanced-config-body {
  position: relative;
  z-index: 2;
  background: #fff;
}

.advanced-option-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  align-items: end;
}

.advanced-option {
  min-width: 0;
  position: relative;
}

.advanced-label {
  min-height: 24px;
  margin-bottom: 8px;
  color: #606266;
  font-weight: 500;
  line-height: 1.5;
}

.advanced-option .el-select {
  width: 100%;
}

.advanced-option /deep/ .el-input__inner {
  height: 40px;
  line-height: 40px;
}

.advanced-error {
  position: absolute;
  left: 0;
  top: 69px;
  color: #f56c6c;
  font-size: 12px;
  line-height: 1;
}

.inline-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  min-height: 40px;
  align-items: center;
}

.inline-checks .el-checkbox {
  margin-right: 0;
}

.advanced-details {
  position: relative;
  z-index: 3;
  margin-top: 16px;
}

/* 其他元素 */
.spj-header-label {
  margin-right: 8px;
}

.compile-button {
  margin-left: 10px;
}

.difficulty-select {
  width: 100%;
}

.language-select {
  width: 100% !important;
}

.mode-select {
  width: 100%;
}

.spj-language-select {
  width: 100%;
}

.template-editor {
  margin-top: 16px;
}

.template-title {
  margin-bottom: 8px;
  color: #606266;
  font-weight: 600;
}

.problem-editor-form /deep/ .el-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
}

.problem-editor-form /deep/ .el-checkbox {
  margin-right: 0;
  line-height: 24px;
}

/* 标签输入 */
.input-new-tag {
  width: 120px;
}

.button-new-tag {
  height: 24px;
  line-height: 22px;
  padding-top: 0;
  padding-bottom: 0;
}

/* 向导操作 */
.wizard-actions {
  margin-top: 24px;
  display: flex;
  justify-content: space-between;
}

.wizard-actions.single-action {
  justify-content: flex-end;
}

.test-data-tip {
  margin-bottom: 16px;
}

/* 示例添加按钮 */
.add-examples,
.add-samples {
  width: 100%;
  background-color: #fff;
  border: 1px dashed #409eff;
  outline: none;
  cursor: pointer;
  color: #409eff;
  height: 40px;
  font-size: 14px;
  border-radius: 4px;
  transition: all 0.2s;
}

.add-examples i,
.add-samples i {
  margin-right: 8px;
}

.add-examples:hover {
  border: 1px solid #409eff;
  background-color: #409eff !important;
  color: #fff;
}

.add-samples {
  border-color: #67c23a;
  color: #67c23a;
}

.add-samples:hover {
  border: 1px solid #67c23a;
  background-color: #67c23a !important;
  color: #fff;
}

.add-example-btn,
.add-sample-btn {
  margin-bottom: 10px;
}

.accordion {
  margin-bottom: 10px;
}

/* 编译错误对话框 */
.dialog-compile-error {
  width: auto;
  max-width: 80%;
  overflow-x: auto;
}

/* 响应式 */
@media (max-width: 1200px) {
  .problem-overview-grid {
    display: contents;
  }

  .problem-quick-settings {
    display: contents;
  }

  .problem-secondary-settings {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .problem-header-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .problem-overview-grid .overview-field:nth-child(n),
  .problem-quick-settings .quick-field {
    grid-column: auto !important;
  }

  .advanced-option-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .problem {
    padding: 10px;
  }

  .problem /deep/ .el-card__body {
    padding: 16px;
  }

  .problem-overview-grid,
  .problem-quick-settings,
  .problem-secondary-settings,
  .advanced-option-grid {
    grid-template-columns: 1fr;
    padding: 16px;
  }

  .problem-header-grid {
    grid-template-columns: 1fr;
  }

  .problem-secondary-settings .secondary-middle-field,
  .problem-secondary-settings .secondary-toggle-field {
    grid-column: auto;
  }
}
</style>
