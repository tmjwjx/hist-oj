<template>
  <div :class="bodyClass">
    <div id="problem-main">
      <!--problem main-->
      <el-row class="problem-box" 
        :id="'problem-box' + '-' + $route.name">
        <el-col
          :sm="24"
          :md="24"
          :lg="24"
          class="problem-left"
          :id="'problem-left'+'-'+ $route.name"
        >
          <div class="problem-top-actions">
            <span v-if="isBattleMode">
              <el-link
                type="warning"
                :underline="false"
                @click="returnToBattleRoom"
              ><i class="fa fa-arrow-left" aria-hidden="true"></i>
                {{ $t('m.Back_To_Battle_Room') }}</el-link>
            </span>
            <span v-if="isShowProblemDiscussion">
              <el-link
                type="primary"
                :underline="false"
                @click="goProblemDiscussion"
              ><i class="fa fa-comments" aria-hidden="true"></i>
                {{ $t('m.Problem_Discussion') }}</el-link>
            </span>
            <span>
              <el-link
                type="primary"
                :underline="false"
                @click="graphVisible = !graphVisible"
              ><i class="fa fa-pie-chart" aria-hidden="true"></i>
                {{ $t('m.Statistic') }}</el-link>
            </span>
            <span>
              <el-link
                type="primary"
                :underline="false"
                @click="goProblemSubmission"
              ><i class="fa fa-bars" aria-hidden="true"></i>
                {{ $t('m.Solutions') }}</el-link>
            </span>
          </div>
          <el-tabs
            v-model="activeName"
            type="border-card"
            @tab-click="handleClickTab"
          >
            <el-tab-pane
              name="problemDetail"
              v-loading="loading"
            >
              <span slot="label">
                <i class="fa fa-list-alt">
                  {{ $t('m.Problem_Description') }}
                </i>
              </span>
              <div
                :padding="10"
                shadow
                :id="'js-left'+'-'+ $route.name"
                class="js-left"
              >
                <div
                  slot="header"
                  class="panel-title"
                >
                  <h1 class="problem-title-text">{{ problemData.problem.title }}</h1>
                  <div class="problem-header-actions">
                    <div class="problem-actions-left">
                      <div class="problem-tag">
                    <span v-if="problemData.problem.isFileIO"
                      style="padding-right: 10px">
                      <el-popover
                        placement="bottom"
                        trigger="hover"
                      >
                      <el-tag
                          slot="reference"
                          size="medium"
                          type="warning"
                          style="cursor: pointer;"
                          effect="dark"
                      ><i class="el-icon-document"> {{ $t('m.File_IO') }}</i>
                      </el-tag>
                      <table style="white-space: nowrap;">
                        <tbody>
                          <tr>
                            <td align="right" style="padding-right: 10px">
                            <strong>{{ $t('m.Input_File') }}</strong>
                            </td>
                            <td>{{ problemData.problem.ioReadFileName }}</td>
                          </tr>
                          <tr>
                            <td align="right" style="padding-right: 10px">
                              <strong>{{ $t('m.Output_File') }}</strong>
                            </td>
                            <td>{{ problemData.problem.ioWriteFileName }}</td>
                          </tr>
                        </tbody>
                      </table>
                      </el-popover>
                    </span>
                    <span v-if="contestID && !contestEnded">
                      <el-tag
                        effect="plain"
                        size="medium"
                      >{{
                        $t('m.Contest_Problem')
                      }}</el-tag>
                    </span>
                    <span
                      v-else-if="problemData.tags.length > 0"
                    >
                      <el-popover
                        placement="right-start"
                        width="60"
                        trigger="hover"
                      >
                        <el-tag
                          slot="reference"
                          size="medium"
                          type="primary"
                          style="cursor: pointer;"
                          effect="plain"
                        >{{ $t('m.Show_Tags') }} <i class="el-icon-caret-bottom"></i></el-tag>
                        <el-tag
                          v-for="(tag, index) in problemData.tags"
                          :key="index"
                          size="small"
                          :color="tag.color ? tag.color : '#409eff'"
                          effect="dark"
                          style="margin-right:5px;margin-top:2px"
                        >{{ tag.name }}</el-tag>
                      </el-popover>
                    </span>
                    <span
                      v-else-if="problemData.tags.length == 0"
                    >
                      <el-tag
                        effect="plain"
                        size="medium"
                      >{{
                        $t('m.No_tag')
                      }}</el-tag>
                    </span>
                      </div>

                    </div>
                  </div>
                  <div class="problem-summary">
                    <template v-if="!isCFProblem">
                      <span>{{ $t('m.Time_Limit') }}：C/C++
                        {{ problemData.problem.timeLimit }}MS，{{
                          $t('m.Other')
                        }}
                        {{ problemData.problem.timeLimit * 2 }}MS</span><br />
                      <span>{{ $t('m.Memory_Limit') }}：C/C++
                        {{ problemData.problem.memoryLimit }}MB，{{
                          $t('m.Other')
                        }}
                        {{ problemData.problem.memoryLimit * 2 }}MB</span><br />
                    </template>

                    <template v-else>
                      <span>{{ $t('m.Time_Limit') }}：{{
                          problemData.problem.timeLimit
                        }}MS</span>
                      <br />
                      <span>{{ $t('m.Memory_Limit') }}：{{
                          problemData.problem.memoryLimit
                        }}MB</span><br />
                    </template>
                    <div class="problem-meta-row">
                      <span>{{ $t('m.Judge_Mode') }}: {{ getJudgeModeText(problemData.problem.judgeMode) }}</span>
                      <template v-if="problemData.problem.difficulty != null">
                        <span>{{ $t('m.Level') }}：<span
                          class="el-tag el-tag--small"
                          :style="getLevelColor(problemData.problem.difficulty)"
                        >{{
                            getLevelName(problemData.problem.difficulty)
                          }}</span></span>
                      </template>
                      <template v-if="problemData.problem.author">
                        <span>{{ $t('m.Created') }}：<el-link
                            type="info"
                            class="author-name"
                            @click="goUserHome(problemData.problem.author)"
                          >{{ problemData.problem.author }}</el-link></span>
                      </template>
                    </div>
                    <template v-if="problemData.problem.type == 1">
                      <span>{{ $t('m.Score') }}：{{ problemData.problem.ioScore }}
                      </span>
                      <span
                        v-if="!contestID"
                        style="margin-left:5px;"
                      >
                        {{ $t('m.OI_Rank_Score') }}：{{
                          calcOIRankScore(
                            problemData.problem.ioScore,
                            problemData.problem.difficulty
                          )
                        }}(0.1*{{ $t('m.Score') }}+2*{{ $t('m.Level') }})
                      </span>
                      <br />
                    </template>

                  </div>
                </div>

                <div id="problem-content">
                  <template v-if="problemData.problem.description">
                    <p class="title">{{ $t('m.Description') }}</p>
                    <Markdown 
                      class="md-content"
                      :isAvoidXss="problemData.problem.gid != null" 
                      :content="problemData.problem.description">
                    </Markdown>
                  </template>

                  <template v-if="problemData.problem.input">
                    <p class="title">{{ $t('m.Input') }}</p>
                    <Markdown 
                      class="md-content"
                      :isAvoidXss="problemData.problem.gid != null" 
                      :content="problemData.problem.input">
                    </Markdown>
                  </template>

                  <template v-if="problemData.problem.output">
                    <p class="title">{{ $t('m.Output') }}</p>
                    <Markdown 
                      class="md-content"
                      :isAvoidXss="problemData.problem.gid != null" 
                      :content="problemData.problem.output">
                    </Markdown>
                  </template>

                  <template v-if="problemData.problem.examples">
                    <div
                      v-for="(example, index) of problemData.problem.examples"
                      :key="index"
                    >
                      <div class="flex-container example">
                        <div class="example-input">
                          <p class="title">
                            {{ $t('m.Sample_Input') }} {{ index + 1 }}
                            <a
                              class="copy"
                              v-clipboard:copy="example.input"
                              v-clipboard:success="onCopy"
                              v-clipboard:error="onCopyError"
                            >
                              <i class="el-icon-document-copy"></i>
                            </a>
                          </p>
                          <pre>{{ example.input }}</pre>
                        </div>
                        <div class="example-output">
                          <p class="title">
                            {{ $t('m.Sample_Output') }} {{ index + 1 }}
                            <a
                              class="copy"
                              v-clipboard:copy="example.output"
                              v-clipboard:success="onCopy"
                              v-clipboard:error="onCopyError"
                            >
                              <i class="el-icon-document-copy"></i>
                            </a>
                          </p>
                          <pre>{{ example.output }}</pre>
                        </div>
                      </div>
                      <div v-if="example.explanation" class="example-explanation">
                        <p class="title">{{ $t('m.Sample_Explanation') }} {{ index + 1 }}</p>
                        <Markdown
                          class="md-content"
                          :isAvoidXss="true"
                          :content="example.explanation">
                        </Markdown>
                      </div>
                    </div>
                  </template>

                  <template v-if="problemData.problem.source && !contestID">
                    <p class="title">{{ $t('m.Source') }}</p>
                    <template v-if="problemData.problem.gid != null">
                      <p
                      class="md-content"
                      v-dompurify-html="problemData.problem.source"
                      ></p>
                    </template>
                    <template v-else>
                      <p
                      class="md-content"
                      v-html="problemData.problem.source"
                      ></p>
                    </template>
                  </template>
                </div>
              </div>
            </el-tab-pane>
            <el-tab-pane name="submitCode" lazy>
              <span slot="label">
                <i class="el-icon-edit-outline"></i> {{ $t('m.Submit_Code') }}
              </span>
              <ProblemSubmit
                v-if="activeName === 'submitCode'"
                embedded
                @cancel="activeName = 'problemDetail'"
                @show-my-submission="showMySubmissionAfterSubmit"
              />
            </el-tab-pane>
            <el-tab-pane name="mySubmission">
              <span slot="label"><i class="el-icon-time"></i> {{ $t('m.My_Submission') }}</span>
              <template v-if="!isAuthenticated">
                <div
                  style="margin:20px 0px;margin-left:-20px;"
                  id="js-submission"
                >
                  <el-alert
                    :title="$t('m.Please_login_first')"
                    type="warning"
                    center
                    :closable="false"
                    :description="$t('m.Login_to_view_your_submission_history')"
                    show-icon
                  >
                  </el-alert>
                </div>
              </template>
              <template v-else>
                <div
                  style="margin-right:10px;"
                  id="js-submission"
                >
                  <vxe-table
                    align="center"
                    :data="mySubmissions"
                    stripe
                    auto-resize
                    border="inner"
                    :loading="loadingTable"
                  >
                    <vxe-table-column
                      :title="$t('m.Submit_Time')"
                      min-width="96"
                    >
                      <template v-slot="{ row }">
                        <span>
                          <el-tooltip
                            :content="row.submitTime | localtime"
                            placement="top"
                          >
                            <span>{{ row.submitTime | fromNow }}</span>
                          </el-tooltip>
                        </span>
                      </template>
                    </vxe-table-column>
                    <vxe-table-column
                      field="status"
                      :title="$t('m.Status')"
                      min-width="160"
                    >
                      <template v-slot="{ row }">
                        <span :class="getStatusColor(row.status)">{{
                          row.statusText || JUDGE_STATUS[row.status].name
                        }}</span>
                      </template>
                    </vxe-table-column>
                    <vxe-table-column
                      :title="$t('m.Time')"
                      min-width="96"
                    >
                      <template v-slot="{ row }">
                        <span>{{ submissionTimeFormat(row.time) }}</span>
                      </template>
                    </vxe-table-column>
                    <vxe-table-column
                      :title="$t('m.Memory')"
                      min-width="96"
                    >
                      <template v-slot="{ row }">
                        <span>{{ submissionMemoryFormat(row.memory) }}</span>
                      </template>
                    </vxe-table-column>
                    <vxe-table-column
                      :title="$t('m.Score')"
                      min-width="64"
                      v-if="problemData.problem.type == 1"
                    >
                      <template v-slot="{ row }">
                        <template v-if="contestID && row.score != null">
                          <el-tag
                            effect="plain"
                            size="medium"
                            :type="JUDGE_STATUS[row.status]['type']"
                          >{{ row.score }}</el-tag>
                        </template>
                        <template v-else-if="row.score != null">
                          <el-tooltip placement="top">
                            <div slot="content">
                              {{ $t('m.Problem_Score') }}：{{
                                row.score != null ? row.score : $t('m.Unknown')
                              }}<br />{{ $t('m.OI_Rank_Score') }}：{{
                                row.oiRankScore != null
                                  ? row.oiRankScore
                                  : $t('m.Unknown')
                              }}<br />
                              {{
                                $t('m.OI_Rank_Calculation_Rule')
                              }}：(score*0.1+difficulty*2)
                            </div>
                            <el-tag
                              effect="plain"
                              size="medium"
                              :type="JUDGE_STATUS[row.status]['type']"
                            >{{ row.score }}</el-tag>
                          </el-tooltip>
                        </template>
                        <template v-else-if="
                            row.status == JUDGE_STATUS_RESERVE['Pending'] ||
                              row.status == JUDGE_STATUS_RESERVE['Compiling'] ||
                              row.status == JUDGE_STATUS_RESERVE['Judging']
                          ">
                          <el-tag
                            effect="plain"
                            size="medium"
                            :type="JUDGE_STATUS[row.status]['type']"
                          >
                            <i class="el-icon-loading"></i>
                          </el-tag>
                        </template>
                        <template v-else>
                          <el-tag
                            effect="plain"
                            size="medium"
                            :type="JUDGE_STATUS[row.status]['type']"
                          >--</el-tag>
                        </template>
                      </template>
                    </vxe-table-column>
                    <vxe-table-column
                      field="language"
                      :title="$t('m.Language')"
                      show-overflow
                      min-width="130"
                    >
                      <template v-slot="{ row }">
                        <el-tooltip
                          class="item"
                          effect="dark"
                          :content="$t('m.View_submission_details')"
                          placement="top"
                        >
                          <el-button
                            type="text"
                            @click="showSubmitDetail(row)"
                          >{{ row.language }}</el-button>
                        </el-tooltip>
                      </template>
                    </vxe-table-column>
                  </vxe-table>
                  <Pagination
                    :total="mySubmission_total"
                    :page-size="mySubmission_limit"
                    @on-change="getMySubmission"
                    :current.sync="mySubmission_currentPage"
                  ></Pagination>
                </div>
              </template>
            </el-tab-pane>

            <el-tab-pane
              name="extraFile"
              v-if="userExtraFile"
            >
              <span slot="label"><i class="fa fa-file-code-o"> {{ $t('m.Problem_Annex') }}</i>
              </span>
              <div id="js-extraFile">
                <el-divider></el-divider>
                <div>
                  <el-tag
                    :key="index"
                    v-for="(value, key, index) in userExtraFile"
                    class="extra-file"
                    :disable-transitions="false"
                    @click="showExtraFileContent(key, value)"
                  >
                    <i class="fa fa-file-code-o"> {{ key }}</i>
                  </el-tag>
                </div>
                <el-divider></el-divider>

                <div
                  class="markdown-body"
                  v-if="fileContent"
                >
                  <h3>
                    {{ fileName }}
                    <el-button
                      type="primary"
                      icon="el-icon-download"
                      size="small"
                      circle
                      @click="downloadExtraFile"
                      class="file-download"
                    ></el-button>
                  </h3>
                  <pre v-highlight="fileContent"><code class="c++"></code></pre>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-col>
      </el-row>
    </div>
    <ProblemHorizontalMenu
      v-if="showProblemHorizontalMenu"
      :pid.sync="problemData.problem.id" 
      :cid="contestID"
      :tid="trainingID"
      ref="problemHorizontalMenu"
      :gid="groupID">
    </ProblemHorizontalMenu>

    <el-dialog
      :visible.sync="graphVisible"
      width="400px"
    >
      <div id="pieChart-detail">
        <ECharts
          :options="largePie"
          :initOptions="largePieInitOpts"
        ></ECharts>
      </div>
      <div slot="footer">
        <el-button
          type="ghost"
          @click="graphVisible = false"
          size="small"
        >{{
          $t('m.Close')
        }}</el-button>
      </div>
    </el-dialog>

  </div>
</template>

<script>
import { mapGetters, mapActions } from "vuex";
import utils from "@/common/utils";
import {
  JUDGE_STATUS,
  CONTEST_STATUS,
  JUDGE_STATUS_RESERVE,
} from "@/common/constants";
import { pie, largePie } from "./chartData";
import api from "@/common/api";
import myMessage from "@/common/message";
import { addCodeBtn } from "@/common/codeblock";
import { getRoomInfo, submitAC as battleSubmitAC, leaveRoom } from "@/api/battle";
import Pagination from "@/components/oj/common/Pagination";
import ProblemHorizontalMenu from "@/components/oj/common/ProblemHorizontalMenu";
import ProblemSubmit from "@/views/oj/problem/ProblemSubmit.vue";
import Markdown from "@/components/oj/common/Markdown";
// 只显示这些状态的图形占用
const filtedStatus = ["wa", "ce", "ac", "pa", "tle", "mle", "re", "pe"];

export default {
  name: "ProblemDetails",
  components: {
    Pagination,
    ProblemHorizontalMenu,
    ProblemSubmit,
    Markdown
  },
  data() {
    return {
      graphVisible: false,
      contestID: 0,
      groupID: null,
      problemID: "",
      trainingID: null,
      problemData: {
        problem: {
          difficulty: 0,
        },
        problemCount: {},
        tags: [],
      },
      pie: pie,
      largePie: largePie,
      // echarts 无法获取隐藏dom的大小，需手动指定
      largePieInitOpts: {
        width: "380",
        height: "380",
      },
      JUDGE_STATUS_RESERVE: {},
      JUDGE_STATUS: {},
      activeName:
        this.$route.query.tab === "mySubmission"
          ? "mySubmission"
          : "problemDetail",
      loadingTable: false,
      mySubmission_total: 0,
      mySubmission_limit: 10,
      mySubmission_currentPage: 1,
      submissionProblemID: "",
      mySubmissions: [],
      refreshStatus: null,
      judgePollAttempts: 0,
      submissionId: "",
      result: { status: 9 },
      activeBattleSubmissionId: "",
      submissionListTimer: null,
      submissionListLoading: false,
      loading: false,
      bodyClass: "",
      userExtraFile: null,
      fileContent: "",
      fileName: "",
      showProblemHorizontalMenu: false,
      // 对战相关
      isBattleMode: false,
      battleRoomId: null,
      battlePollingTimer: null,
      battleEnded: false,
    };
  },
  created() {
    this.JUDGE_STATUS_RESERVE = Object.assign({}, JUDGE_STATUS_RESERVE);
    this.JUDGE_STATUS = Object.assign({}, JUDGE_STATUS);
    this.syncRouteLayout();
    if (utils.isFocusModePage(this.$route.name)) {
      this.contestID = this.$route.params.contestID || 0;
      this.trainingID = this.$route.params.trainingID || null;
      this.groupID = this.$route.params.groupID || null;
      this.showProblemHorizontalMenu = true;
    }
  },

  mounted() {
    this.syncActiveNameFromRoute();
    this.init();
    this.initializeBattleSubmission();
  },
  beforeDestroy() {
    this.stopSubmissionPolling();
    // 清理对战轮询定时器
    this.stopBattlePolling();
  },
  methods: {
    ...mapActions(["changeDomTitle"]),
    syncRouteLayout() {
      const isFocusModePage = utils.isFocusModePage(this.$route.name);
      this.bodyClass = "problem-body";
      this.showProblemHorizontalMenu = isFocusModePage;
    },
    syncActiveNameFromRoute() {
      this.activeName =
        this.$route.query.tab === "mySubmission"
          ? "mySubmission"
          : "problemDetail";
    },
    async initializeBattleSubmission() {
      await this.checkBattleMode();
      const submitID =
        this.$route.query.battle && this.$route.query.submitID
          ? String(this.$route.query.submitID)
          : "";
      if (!submitID) return;
      if (this.activeBattleSubmissionId === submitID) return;
      if (this.refreshStatus) clearTimeout(this.refreshStatus);
      this.refreshStatus = null;
      this.activeBattleSubmissionId = submitID;
      this.submissionId = submitID;
      this.result = { status: 9 };
      this.checkSubmissionStatus();
    },
    handleClickTab({ name }) {
      if (name == "mySubmission" && this.isAuthenticated) {
        this.submissionProblemID = "";
        this.getMySubmission();
      } else {
        if (name === "problemDetail") {
          this.submissionProblemID = "";
        }
        this.stopMySubmissionPolling();
      }
    },
    showMySubmissionAfterSubmit(payload = {}) {
      const submitID = payload && payload.submitID;
      this.submissionProblemID =
        payload && payload.problemID ? String(payload.problemID) : "";
      this.activeName = "mySubmission";
      this.mySubmission_currentPage = 1;
      this.getMySubmission();

      if (submitID && this.isBattleMode) {
        if (this.refreshStatus) clearTimeout(this.refreshStatus);
        this.refreshStatus = null;
        this.activeBattleSubmissionId = String(submitID);
        this.submissionId = String(submitID);
        this.result = { status: 9 };
        this.checkSubmissionStatus();
      }
    },
    getMySubmission(options = {}) {
      const silent = options && typeof options === "object" && options.silent;
      if (this.submissionListLoading) return Promise.resolve();
      let params = {
        onlyMine: true,
        currentPage: this.mySubmission_currentPage,
        problemID: this.submissionProblemID || this.problemID,
        contestID: this.contestID,
        completeProblemID: true,
        gid: this.groupID,
        limit: this.mySubmission_limit,
      };
      if (this.contestID) {
        if (this.contestStatus == CONTEST_STATUS.SCHEDULED) {
          params.beforeContestSubmit = true;
        } else {
          params.beforeContestSubmit = false;
        }
        params.containsEnd = true;
      }
      let func = this.contestID
        ? "getContestSubmissionList"
        : "getSubmissionList";
      this.submissionListLoading = true;
      if (!silent) this.loadingTable = true;
      return api[func](this.mySubmission_limit, utils.filterEmptyValue(params))
        .then(
          (res) => {
            let data = res.data.data;
            this.mySubmissions = data.records;
            this.mySubmission_total = data.total;
            this.scheduleMySubmissionPolling();
          },
          () => this.stopMySubmissionPolling()
        )
        .finally(() => {
          this.submissionListLoading = false;
          this.loadingTable = false;
        });
    },
    isSubmissionRunning(status) {
      return [
        JUDGE_STATUS_RESERVE["Pending"],
        JUDGE_STATUS_RESERVE["Compiling"],
        JUDGE_STATUS_RESERVE["Judging"],
        JUDGE_STATUS_RESERVE["Submitting"],
      ].includes(status);
    },
    syncSubmissionRow(submission) {
      if (!submission || !submission.submitId) return;
      const index = this.mySubmissions.findIndex(
        (item) => item.submitId === submission.submitId
      );
      if (index >= 0) {
        this.$set(this.mySubmissions, index, {
          ...this.mySubmissions[index],
          ...submission,
        });
      }
    },
    scheduleMySubmissionPolling() {
      this.stopMySubmissionPolling();
      if (
        this.activeName !== "mySubmission" ||
        !this.isAuthenticated ||
        !this.mySubmissions.some((item) => this.isSubmissionRunning(item.status))
      ) {
        return;
      }
      this.submissionListTimer = setTimeout(
        () => this.getMySubmission({ silent: true }),
        1500
      );
    },
    stopMySubmissionPolling() {
      if (this.submissionListTimer) clearTimeout(this.submissionListTimer);
      this.submissionListTimer = null;
    },
    stopSubmissionPolling() {
      if (this.refreshStatus) clearTimeout(this.refreshStatus);
      this.refreshStatus = null;
      this.stopMySubmissionPolling();
    },
    getStatusColor(status) {
      return "el-tag el-tag--medium status-" + JUDGE_STATUS[status].color;
    },
    submissionTimeFormat(time) {
      return utils.submissionTimeFormat(time);
    },

    submissionMemoryFormat(memory) {
      return utils.submissionMemoryFormat(memory);
    },

    showSubmitDetail(row) {
      if (row.cid != 0) {
        // 比赛提交详情
        this.$router.push({
          name: "ContestSubmissionDetails",
          params: {
            contestID: this.$route.params.contestID,
            problemID: row.displayId,
            submitID: row.submitId,
          },
        });
      } else if (this.groupID) {
        this.$router.push({
          name: "GroupSubmissionDetails",
          params: {
            groupID: this.groupID,
            submitID: row.submitId,
          },
        });
      } else {
        this.$router.push({
          name: "SubmissionDetails",
          params: { submitID: row.submitId },
        });
      }
    },

    init() {
      if(this.$route.name === "ContestFullProblemDetails"){
        this.$store.dispatch('getContest');
      }
      this.contestID = this.$route.params.contestID || 0;
      this.groupID = this.$route.params.groupID || null;
      this.problemID = this.$route.params.problemID;
      this.trainingID = this.$route.params.trainingID || null;
      let func =
        this.$route.name === "ContestProblemDetails" ||
        this.$route.name === "ContestFullProblemDetails"
          ? "getContestProblem"
          : "getProblem";
      this.loading = true;
      api[func](this.problemID, this.contestID, this.groupID, true).then(
        (res) => {
          let result = res.data.data;
          this.changeDomTitle({ title: result.problem.title });
          result["myStatus"] = -10; // 设置默认值

          result.problem.examples = utils.stringToExamples(
            result.problem.examples
          );
          if (result.problem.userExtraFile) {
            this.userExtraFile = JSON.parse(result.problem.userExtraFile);
          }

          this.problemData = result;

          this.loading = false;

          this.changePie(result.problemCount);
          this.$nextTick((_) => {
            addCodeBtn();
          });
        },
        (err) => {
          this.loading = false;
        }
      );
      
      if(this.activeName == "mySubmission"){
        this.getMySubmission();
      }
    },
    changePie(problemData) {
      let total = problemData.total;
      let acNum = problemData.ac;
      // 该状态结果数为0的不显示,同时一些无关参数也排除
      for (let k in problemData) {
        if (problemData[k] == 0 || filtedStatus.indexOf(k) === -1) {
          delete problemData[k];
        }
      }

      let data = [
        { name: "WA", value: total - acNum },
        { name: "AC", value: acNum },
      ];
      this.pie.series[0].data = data;
      // 只把大图的AC selected下，这里需要做一下deepcopy
      let data2 = JSON.parse(JSON.stringify(data));
      data2[1].selected = true;
      this.largePie.series[1].data = data2;

      // 根据结果设置legend,没有提交过的legend不显示
      let legend = Object.keys(problemData).map((ele) =>
        (ele + "").toUpperCase()
      );
      if (legend.length === 0) {
        legend.push("AC", "WA");
      }
      this.largePie.legend.data = legend;

      // 把ac的数据提取出来放在最后
      let acCount = problemData.ac;
      delete problemData.ac;

      let largePieData = [];
      Object.keys(problemData).forEach((ele) => {
        largePieData.push({
          name: (ele + "").toUpperCase(),
          value: problemData[ele],
        });
      });
      largePieData.push({ name: "AC", value: acCount });
      this.largePie.series[0].data = largePieData;
    },

    goProblemSubmission() {
      if (this.contestID) {
        this.$router.push({
          name: "ContestSubmissionList",
          params: { contestID: this.contestID },
          query: { problemID: this.problemID, completeProblemID: true },
        });
      } else if (this.groupID) {
        this.$router.push({
          name: "GroupSubmissionList",
          params: { groupID: this.groupID },
          query: {
            problemID: this.problemID,
            completeProblemID: true,
          },
        });
      } else {
        this.$router.push({
          name: "SubmissionList",
          query: {
            problemID: this.problemID,
            completeProblemID: true,
          },
        });
      }
    },
    goProblemDiscussion() {
      if (this.groupID) {
        this.$router.push({
          name: "GroupProblemDiscussion",
          params: { problemID: this.problemID, groupID: this.groupID },
        });
      } else {
        this.$router.push({
          name: "ProblemDiscussion",
          params: { problemID: this.problemID },
        });
      }
    },

    checkSubmissionStatus() {
      if (!this.submissionId) return;
      // 使用setTimeout避免一些问题
      if (this.refreshStatus) {
        // 如果之前的提交状态检查还没有停止,则停止,否则将会失去timeout的引用造成无限请求
        clearTimeout(this.refreshStatus);
      }
      this.judgePollAttempts = 0;
      const checkStatus = () => {
        let submitId = this.submissionId;
        api.getSubmission(submitId).then(
          (res) => {
            const submission = res.data.data.submission;
            this.result.status = submission.status;
            this.result.statusText = submission.statusText;
            this.syncSubmissionRow(submission);
            if (Object.keys(submission).length !== 0) {
              if (!this.isSubmissionRunning(submission.status)) {
                clearTimeout(this.refreshStatus);
                this.refreshStatus = null;
                if (this.activeName === "mySubmission") {
                  this.getMySubmission({ silent: true });
                }

                // 检查是否AC，如果是对战模式则处理
                if (this.result.status === 0 && this.isBattleMode) {
                  this.handleBattleAC();
                }

                this.init();
                if(this.showProblemHorizontalMenu){
                  this.$refs.problemHorizontalMenu.getFullScreenProblemList();
                }
              } else {
                this.refreshStatus = setTimeout(checkStatus, this.nextJudgePollDelay());
              }
            } else {
              this.refreshStatus = setTimeout(checkStatus, this.nextJudgePollDelay());
            }
          },
          (res) => {
            clearTimeout(this.refreshStatus);
            this.refreshStatus = null;
          }
        );
      };
      // 设置每2秒检查一下该题的提交结果
      this.refreshStatus = setTimeout(checkStatus, 2000);
    },

    // 判题结果轮询退避：前2次2秒（正常判题1-3秒出结果，体感不变），
    // 之后逐步拉长到10秒，避免比赛高峰期 400 人同时轮询打满后端
    nextJudgePollDelay() {
      const n = ++this.judgePollAttempts;
      if (n <= 2) return 2000;
      if (n <= 4) return 4000;
      if (n <= 6) return 6000;
      return 10000;
    },

    showExtraFileContent(name, content) {
      this.fileName = name;
      this.fileContent = content;
      this.$nextTick((_) => {
        addCodeBtn();
      });
    },
    downloadExtraFile() {
      utils.downloadFileByText(this.fileName, this.fileContent);
    },

    getLevelColor(difficulty) {
      return utils.getLevelColor(difficulty);
    },
    getLevelName(difficulty) {
      return utils.getLevelName(difficulty);
    },
    getJudgeModeText(mode) {
      const modeMap = {
        'default': this.$t('m.Judge_Mode_Default'),
        'spj': this.$t('m.Judge_Mode_Spj'),
        'interactive': this.$t('m.Judge_Mode_Interactive'),
        'subtask': this.$t('m.Judge_Mode_Subtask')
      }
      return modeMap[mode] || mode || this.$t('m.Judge_Mode_Default')
    },
    goUserHome(username) {
      this.$router.push({
        path: "/user-home",
        query: { username },
      });
    },
    calcOIRankScore(score, difficulty) {
      return Math.round(0.1 * score + 2 * difficulty);
    },

    onCopy(event) {
      myMessage.success(this.$i18n.t("m.Copied_successfully"));
    },
    onCopyError(e) {
      myMessage.success(this.$i18n.t("m.Copied_failed"));
    },
    // 对战相关方法
    async checkBattleMode() {
      // 检查URL参数中是否包含battle参数
      const battleRoomId = this.$route.query.battle;
      if (!battleRoomId) {
        this.stopBattlePolling();
        this.isBattleMode = false;
        this.battleRoomId = null;
        this.battleEnded = false;
        return;
      }
      if (battleRoomId) {
        // 先验证房间是否存在
        try {
          const res = await getRoomInfo(battleRoomId);
          // 房间不存在或已被解散
          if (res.data.code === 1 || !res.data.data) {
            // 清除URL中的battle参数
            this.$router.replace({ query: {} });
            return;
          }
          // 房间存在，设置对战模式
          this.isBattleMode = true;
          this.battleRoomId = battleRoomId;
          // 保存到sessionStorage
          const userId = this.$store.getters.userInfo?.uid;
          if (userId) {
            sessionStorage.setItem(`battle_room_${userId}`, battleRoomId);
          }
          // 开始轮询对战状态
          this.startBattlePolling();
        } catch (error) {
          // 房间不存在或请求失败，清除URL中的battle参数
          this.$router.replace({ query: {} });
        }
      }
    },
    startBattlePolling() {
      if (this.battlePollingTimer) {
        clearInterval(this.battlePollingTimer);
      }
      // 每2秒轮询一次对战状态 - 及时检测对战结束
      this.battlePollingTimer = setInterval(() => {
        this.checkBattleStatus();
      }, 2000);
    },
    stopBattlePolling() {
      if (this.battlePollingTimer) {
        clearInterval(this.battlePollingTimer);
        this.battlePollingTimer = null;
      }
    },
    async checkBattleStatus() {
      if (!this.battleRoomId || this.battleEnded) {
        return;
      }
      try {
        const res = await getRoomInfo(this.battleRoomId);

        // 房间不存在（被解散）
        if (res.data.code === 1 || !res.data.data) {
          this.battleEnded = true;
          this.stopBattlePolling();
          // 清除对战模式和房间ID
          this.isBattleMode = false;
          this.battleRoomId = null;
          // 清除 sessionStorage 中的房间记录
          const userId = this.$store.getters.userInfo?.uid;
          if (userId) {
            sessionStorage.removeItem(`battle_room_${userId}`);
          }
          // 房间解散后不做任何提示和跳转，保持用户在当前页面
          return;
        }

        if (res.data.code === 0 && res.data.data) {
          const room = res.data.data.room;
          // 检查对战是否已结束（status = 2）
          if (room.status === 2 && room.winnerId) {
            this.battleEnded = true;
            this.stopBattlePolling();
            const currentUserId = this.$store.getters.userInfo?.uid;
            const isWinner = room.winnerId === currentUserId;
            const endReason = room.endReason || 'ac'; // 默认为ac

            // 根据胜负和结束原因显示不同的消息
            let message, title, type;
            if (isWinner) {
              title = this.$t('m.Battle_Win_Title');
              if (endReason === 'ac') {
                message = this.$t('m.Battle_Win_AC');
              } else if (endReason === 'giveup') {
                message = this.$t('m.Battle_Win_Giveup');
              } else if (endReason === 'timeout') {
                message = this.$t('m.Battle_Win_Timeout');
              } else {
                message = this.$t('m.Battle_Win');
              }
              type = 'success';
            } else {
              title = this.$t('m.Battle_Lose_Title');
              if (endReason === 'ac') {
                message = this.$t('m.Battle_Lose_AC');
              } else if (endReason === 'giveup') {
                message = this.$t('m.Battle_Lose_Giveup');
              } else if (endReason === 'timeout') {
                message = this.$t('m.Battle_Lose_Timeout');
              } else {
                message = this.$t('m.Battle_Lose');
              }
              type = 'warning';
            }

            // 显示通知
            this.$notify({
              title: title,
              message: message,
              type: type,
              duration: 2000,
              position: 'top-right'
            });

            // 2秒后自动返回对战房间
            setTimeout(() => {
              this.returnToBattleRoom();
            }, 2000);
          }
        }
      } catch (error) {
        // 房间不存在或已被解散
        if (error.response && (error.response.status === 404 || error.response.status === 400)) {
          this.battleEnded = true;
          this.stopBattlePolling();
          // 清除对战模式和房间ID
          this.isBattleMode = false;
          this.battleRoomId = null;
          // 清除 sessionStorage 中的房间记录
          const userId = this.$store.getters.userInfo?.uid;
          if (userId) {
            sessionStorage.removeItem(`battle_room_${userId}`);
          }
          // 房间解散后不做任何提示和跳转，保持用户在当前页面
        }
        // 其他错误静默处理
      }
    },
    async returnToBattleRoom() {
      if (!this.battleRoomId) {
        return;
      }

      try {
        // 先检查房间是否存在
        const res = await getRoomInfo(this.battleRoomId);

        // 房间不存在或已被解散
        if (res.data.code === 1 || !res.data.data) {
          // 立即隐藏"返回对战房间"链接
          this.isBattleMode = false;
          this.battleEnded = true;
          this.battleRoomId = null;
          this.stopBattlePolling();
          // 清除 sessionStorage
          const userId = this.$store.getters.userInfo?.uid;
          if (userId) {
            sessionStorage.removeItem(`battle_room_${userId}`);
          }
          // 房间解散后不做任何提示，静默返回大厅
          this.$router.push({ name: 'BattleHome' });
          return;
        }

        // 房间存在，跳转回房间
        this.$router.push({
          name: 'BattleRoom',
          params: { roomId: this.battleRoomId }
        });
      } catch (error) {
        // 房间不存在或已被解散 - 立即隐藏链接
        this.isBattleMode = false;
        this.battleEnded = true;
        this.battleRoomId = null;
        this.stopBattlePolling();
        // 清除 sessionStorage
        const userId = this.$store.getters.userInfo?.uid;
        if (userId) {
          sessionStorage.removeItem(`battle_room_${userId}`);
        }
        // 房间解散后不做任何提示，静默返回大厅
        this.$router.push({ name: 'BattleHome' });
      }
    },
    async leaveBattleRoomAndReturnHome() {
      try {
        // 调用退出房间API
        const res = await leaveRoom({ roomId: this.battleRoomId });
        if (res.data.code === 0) {
          this.$message.success(this.$t('m.Left_Room'));
        }
      } catch (error) {
        // 退出失败也静默处理,不影响返回大厅
      } finally {
        // 无论API调用成功与否,都清除本地状态并返回大厅
        this.isBattleMode = false;
        this.battleEnded = true;
        this.battleRoomId = null;
        this.stopBattlePolling();

        // 清除 sessionStorage
        const userId = this.$store.getters.userInfo?.uid;
        if (userId) {
          sessionStorage.removeItem(`battle_room_${userId}`);
        }

        // 返回对战大厅
        this.$router.push({ name: 'BattleHome' });
      }
    },
    async handleBattleAC() {
      if (!this.isBattleMode || !this.battleRoomId || this.battleEnded) {
        return;
      }
      try {
        const userId = this.$store.getters.userInfo?.uid;
        await battleSubmitAC({
          roomId: this.battleRoomId,
          problemId: this.problemID
        });
        // AC提交成功，继续轮询以获取对战结果
      } catch (error) {
        // 静默处理错误
      }
    }
  },
  computed: {
    ...mapGetters([
      "contestStatus",
      "isAuthenticated",
      "websiteConfig"
    ]),
    contestEnded() {
      return this.contestStatus === CONTEST_STATUS.ENDED;
    },
    isCFProblem() {
      if (
        this.problemID.indexOf("CF-") == 0 ||
        this.problemID.indexOf("GYM-") == 0
      ) {
        return true;
      } else {
        return false;
      }
    },
    isShowProblemDiscussion() {
      if (!this.contestID) {
        if (this.groupID) {
          if (this.websiteConfig.openGroupDiscussion) {
            return true;
          }
        } else {
          if (this.websiteConfig.openPublicDiscussion) {
            return true;
          }
        }
      }
      return false;
    },
  },
  beforeRouteLeave(to, from, next) {
    this.stopSubmissionPolling();
    if(this.$route.name === "ContestFullProblemDetails"){
      this.$store.commit('clearContest');
    }
    next();
  },
  beforeRouteUpdate (to, from, next) {
    next();
  },
  watch: {
    $route(to, from) {
      const routeContextKeys = [
        "problemID",
        "contestID",
        "trainingID",
        "groupID",
      ];
      const contextChanged =
        to.name !== from.name ||
        routeContextKeys.some(
          (key) => String(to.params[key] || "") !== String(from.params[key] || "")
        );
      const battleSubmissionChanged =
        String(to.query.battle || "") !== String(from.query.battle || "") ||
        String(to.query.submitID || "") !== String(from.query.submitID || "");

      this.syncRouteLayout();
      this.syncActiveNameFromRoute();
      if (contextChanged) {
        this.stopSubmissionPolling();
        this.submissionProblemID = "";
        this.activeBattleSubmissionId = "";
        this.submissionId = "";
        this.result = { status: 9 };
        this.init();
      } else if (this.activeName === "mySubmission" && this.isAuthenticated) {
        this.getMySubmission();
      } else {
        this.stopMySubmissionPolling();
      }

      if (contextChanged || battleSubmissionChanged) {
        if (!contextChanged) {
          if (this.refreshStatus) clearTimeout(this.refreshStatus);
          this.refreshStatus = null;
          this.activeBattleSubmissionId = "";
          this.submissionId = "";
          this.result = { status: 9 };
        }
        this.initializeBattleSubmission();
      }
    },
    isAuthenticated(newVal) {
      if (newVal === true) {
        this.init();
      }
    },
  },
};
</script>
<style>
.katex .katex-mathml {
  display: none;
}
</style>

<style scoped>
.problem-menu {
  float: left;
}
a {
  color: #3091f2 !important ;
}
.problem-menu span {
  margin-left: 5px;
}
.problem-top-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 48px;
  margin-bottom: -1px;
  padding: 10px 20px;
  border: 1px solid #dcdfe6;
  border-bottom: 0;
  background: #fff;
  justify-content: flex-start;
}
.problem-top-actions .el-link {
  white-space: nowrap;
}
.el-link {
  font-size: 16px !important;
}
.author-name {
  font-size: 14px !important;
  color: #909399 !important;
}
.panel-title {
  padding: 24px 28px 0;
}

.problem-title-text {
  margin: 0 0 22px;
  color: #303133;
  font-size: 30px;
  font-weight: 600;
  line-height: 1.35;
  text-align: center;
}

.problem-header-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.problem-actions-left {
  display: flex;
  align-items: center;
}

.problem-summary {
  margin: 24px auto 18px;
  padding: 0;
  color: #495060;
  font-size: 15px;
  line-height: 1.9;
  text-align: center;
}

.problem-meta-row {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 22px;
}

.extra-file {
  margin: 10px;
  cursor: pointer;
}
.file-download {
  vertical-align: bottom;
  float: right;
  margin-right: 5px;
}

/deep/.el-tabs--border-card > .el-tabs__content {
  padding-top: 0px;
  padding-right: 0px;
  padding-bottom: 0px;
}

.js-left {
  padding-right: 15px;
}
@media screen and (min-width: 992px) {
  .problem-body {
    width: 82%;
    max-width: 2000px;
    margin: 0 auto;
  }

  .problem-tag {
    display: inline;
  }

  .problem-menu {
    float: right;
  }

  .problem-menu span {
    margin-left: 10px;
  }

}

.problem-box,
.problem-left,
#problem-main {
  width: 100%;
  height: auto;
  min-height: 0;
  overflow: visible;
}

.js-left,
#js-submission,
#js-extraFile {
  height: auto !important;
  max-height: none;
  overflow: visible;
}
/deep/ .el-card__header {
  border-bottom: 0px;
  padding-bottom: 0px;
}
/deep/ .el-card__body{
  padding-bottom: 5px !important;
}
#problem-content {
  margin: 28px auto 0;
  padding: 0 28px 40px;
  text-align: left;
}
#problem-content .title {
  font-size: 16px;
  font-weight: 600;
  margin: 25px 0 8px 0;
  color: #3091f2;
}
#problem-content .copy {
  padding-left: 8px;
}

.example-explanation {
  margin: 8px 0 18px;
  padding: 10px 14px;
  border-left: 3px solid #67c23a;
  background: #f0f9eb;
}

.example-explanation .title {
  margin: 0 0 4px !important;
  color: #67c23a !important;
}

.md-content {
  margin: 1em 0;
  font-size: 16px;
  line-height: 1.8;
}
.flex-container {
  display: flex;
  width: 100%;
  max-width: 100%;
  justify-content: space-around;
  align-items: flex-start;
  flex-flow: row nowrap;
}

.example {
  align-items: stretch;
}
.example-input,
.example-output {
  width: 50%;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
}
.example pre {
  flex: 1 1 auto;
  align-self: stretch;
  border-style: solid;
  background: transparent;
  padding: 5px 10px;
  white-space: pre;
  margin-top: 10px;
  margin-bottom: 10px;
  background: #f1f1f1;
  border: 1px dashed #e9eaec;
  overflow: auto;
  font-size: 1.1em;
  margin-right: 7%;
}
/deep/.el-dialog__body {
  padding: 10px 10px !important;
}
#pieChart .echarts {
  height: 250px;
  width: 210px;
}
#pieChart #detail {
  position: absolute;
  right: 10px;
  top: 10px;
}
/deep/.echarts {
  width: 350px;
  height: 350px;
}
#pieChart-detail {
  /* margin-top: 20px; */
  height: 350px;
}

@media screen and (max-width: 991px) {
  .problem-body {
    width: 100%;
  }

  .panel-title,
  #problem-content {
    padding-left: 16px;
    padding-right: 16px;
  }

  .problem-title-text {
    font-size: 24px;
  }

  .problem-header-actions {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
  }

  .problem-top-actions {
    flex-wrap: wrap;
    gap: 10px;
    padding-left: 16px;
    padding-right: 16px;
  }

  .problem-menu {
    float: none;
  }
}
</style>
