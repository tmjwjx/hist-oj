<template>
  <div :class="['problem-submit-page', { 'is-embedded': embedded }]">
    <el-card
      v-loading="loading"
      :shadow="embedded ? 'never' : 'always'"
      class="submit-card"
    >
      <div v-if="!embedded" slot="header" class="submit-card-header">
        <div>
          <h2>提交代码</h2>
          <p v-if="problemTitle" class="problem-title">{{ problemTitle }}</p>
        </div>
        <el-button size="small" icon="el-icon-back" @click="returnToProblem">
          返回题目
        </el-button>
      </div>

      <el-form label-position="left" label-width="92px" class="submit-form">
        <el-form-item v-if="isContestSubmit" label="题目" required>
          <el-select
            :value="selectedProblemID"
            filterable
            placeholder="请选择要提交的题目"
            class="form-control"
            @change="onContestProblemChange"
          >
            <el-option
              v-for="problem in contestProblems"
              :key="problem.displayId"
              :label="contestProblemLabel(problem)"
              :value="String(problem.displayId)"
            ></el-option>
          </el-select>
          <span v-if="loadingProblem" class="problem-loading">
            <i class="el-icon-loading"></i> 正在加载题目信息
          </span>
        </el-form-item>

        <el-form-item v-else label="题目">
          <el-input :value="problemLabel" class="form-control" disabled></el-input>
        </el-form-item>

        <el-form-item label="语言" required>
          <el-select
            :value="language"
            :disabled="!selectedProblemID"
            class="form-control"
            placeholder="请选择语言"
            @change="onLanguageChange"
          >
            <el-option
              v-for="item in problemData.languages"
              :key="item"
              :label="item"
              :value="item"
            ></el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="源代码" required>
          <div class="source-toolbar">
            <el-button
              size="small"
              icon="el-icon-upload2"
              :disabled="!selectedProblemID || loadingProblem"
              @click="chooseSourceFile"
            >选择文件</el-button>
            <el-button
              size="small"
              icon="el-icon-refresh-left"
              :disabled="!selectedProblemID || !language || loadingProblem"
              @click="resetToTemplate"
            >恢复代码模板</el-button>
            <span v-if="sourceFileName" class="source-file-name">
              已读取：{{ sourceFileName }}
            </span>
            <input
              ref="sourceFileInput"
              type="file"
              class="source-file-input"
              @change="readSourceFile"
            />
          </div>
          <el-input
            v-model="code"
            type="textarea"
            :autosize="{ minRows: 22, maxRows: 36 }"
            :disabled="!selectedProblemID"
            resize="vertical"
            class="source-input"
            placeholder="请在这里粘贴源代码，或者选择本地源码文件"
            @keydown.native="handleSourceKeydown"
          ></el-input>
          <div class="source-hint">
            支持直接粘贴代码或读取本地源码文件，代码长度不能超过 65535 个字符。
          </div>
        </el-form-item>

        <el-form-item class="submit-actions">
          <el-button
            type="primary"
            icon="el-icon-upload"
            :loading="submitting"
            :disabled="submitDisabled"
            @click="submitCode"
          >{{ submitting ? '提交中...' : '提交代码' }}</el-button>
          <el-button @click="returnToProblem">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { mapGetters } from "vuex";
import api from "@/common/api";
import storage from "@/common/storage";
import myMessage from "@/common/message";
import {
  JUDGE_STATUS_RESERVE,
  RULE_TYPE,
  buildIndividualLanguageAndSettingKey,
  buildProblemCodeAndSettingKey,
} from "@/common/constants";

const EMPTY_PROBLEM_DATA = () => ({
  problem: {},
  languages: [],
  codeTemplate: {},
});

export default {
  name: "ProblemSubmit",
  props: {
    embedded: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      loading: false,
      loadingProblem: false,
      submitting: false,
      contestProblems: [],
      selectedProblemID: "",
      currentDraftProblemID: "",
      problemData: EMPTY_PROBLEM_DATA(),
      language: "",
      code: "",
      sourceFileName: "",
      submissionExists: false,
      problemLoadRequestId: 0,
      submissionStatusPromise: null,
      submissionStatusLoading: false,
    };
  },
  computed: {
    ...mapGetters([
      "isAuthenticated",
      "canSubmit",
      "problemSubmitDisabled",
      "contestRuleType",
      "ContestRealTimePermission",
    ]),
    contestID() {
      return this.$route.params.contestID || null;
    },
    trainingID() {
      return this.$route.params.trainingID || null;
    },
    groupID() {
      return this.$route.params.groupID || null;
    },
    isContestSubmit() {
      return Boolean(this.contestID);
    },
    problemTitle() {
      return this.problemData.problem.title || "";
    },
    problemLabel() {
      if (!this.selectedProblemID) return "";
      return this.problemTitle
        ? `${this.selectedProblemID} - ${this.problemTitle}`
        : this.selectedProblemID;
    },
    submitDisabled() {
      return (
        this.loadingProblem ||
        this.submissionStatusLoading ||
        !this.selectedProblemID ||
        !this.language ||
        (this.isAuthenticated && this.isContestSubmit &&
          (this.problemSubmitDisabled || !this.canSubmit))
      );
    },
  },
  created() {
    this.initializePage();
  },
  beforeDestroy() {
    this.saveDraft();
  },
  methods: {
    async initializePage() {
      this.loading = true;
      try {
        if (this.isContestSubmit) {
          await this.ensureContestLoaded();
          const response = await api.getContestProblemList(this.contestID, true);
          this.contestProblems = response.data.data || [];

          const requestedProblemID = String(
            this.$route.query.problemID || this.$route.params.problemID || ""
          );
          const requestedProblemExists = this.contestProblems.some(
            (problem) => String(problem.displayId) === requestedProblemID
          );

          if (requestedProblemExists) {
            this.selectedProblemID = requestedProblemID;
            await this.loadProblem(requestedProblemID);
          }
        } else {
          this.selectedProblemID = String(
            this.$route.params.problemID || this.$route.query.problemID || ""
          );
          if (this.selectedProblemID) {
            await this.loadProblem(this.selectedProblemID);
          }
        }
      } catch (error) {
        // 请求层会展示具体错误；保留已成功加载的比赛题目列表，便于重新选择。
      } finally {
        this.loading = false;
      }
    },
    async ensureContestLoaded() {
      // ContestSubmit 是独立路由。离开 ContestDetails 时父组件会清空
      // contest store，因此这里必须主动重新加载，不能依赖旧页面缓存。
      await this.$store.dispatch("getContest");
    },
    contestProblemLabel(problem) {
      const title = problem.displayTitle || problem.title || "";
      return title ? `${problem.displayId} - ${title}` : String(problem.displayId);
    },
    onContestProblemChange(problemID) {
      const nextProblemID = String(problemID);
      if (nextProblemID === this.selectedProblemID) return;
      this.saveDraft();
      this.selectedProblemID = nextProblemID;
      this.loadProblem(nextProblemID);
    },
    async loadProblem(problemID) {
      if (!problemID) return;

      const requestId = ++this.problemLoadRequestId;
      this.loadingProblem = true;
      this.submissionStatusPromise = null;
      this.submissionStatusLoading = false;

      try {
        const response = this.isContestSubmit
          ? await api.getContestProblem(
              problemID,
              this.contestID,
              this.groupID,
              true
            )
          : await api.getProblem(problemID, null, this.groupID, true);
        if (requestId !== this.problemLoadRequestId) return;

        const result = response.data.data;
        const nextProblemData = {
          ...EMPTY_PROBLEM_DATA(),
          ...result,
          languages: result.languages || [],
          codeTemplate: result.codeTemplate || {},
        };
        this.problemData = nextProblemData;
        this.currentDraftProblemID = problemID;
        this.sourceFileName = "";
        this.submissionExists = false;
        this.restoreDraft();
        // 题目详情就绪后立即恢复表单；提交状态在后台更新，不再阻塞选题。
        this.loadingProblem = false;
        const needsSubmissionStatus =
          this.isContestSubmit &&
          this.contestRuleType === RULE_TYPE.OI &&
          !this.ContestRealTimePermission;
        this.submissionStatusLoading = needsSubmissionStatus;
        this.submissionStatusPromise = this.loadSubmissionStatus(
          this.problemData.problem.id,
          requestId
        ).finally(() => {
          if (requestId === this.problemLoadRequestId) {
            this.submissionStatusLoading = false;
          }
        });
      } catch (error) {
        if (requestId === this.problemLoadRequestId) {
          this.selectedProblemID = this.currentDraftProblemID || "";
        }
      } finally {
        if (requestId === this.problemLoadRequestId) {
          this.loadingProblem = false;
        }
      }
    },
    restoreDraft() {
      const draft = storage.get(
        buildProblemCodeAndSettingKey(
          this.currentDraftProblemID,
          this.contestID
        )
      );
      const individualSetting = storage.get(
        buildIndividualLanguageAndSettingKey()
      );
      const languages = this.problemData.languages;
      const preferredLanguage =
        (draft && draft.language) ||
        (individualSetting && individualSetting.language);

      if (preferredLanguage && languages.includes(preferredLanguage)) {
        this.language = preferredLanguage;
      } else {
        this.language = languages[0] || "";
      }

      if (draft && typeof draft.code === "string") {
        this.code = draft.code;
      } else {
        this.code = this.getCodeTemplate(this.language);
      }
    },
    saveDraft() {
      if (!this.currentDraftProblemID) return;
      const draftKey = buildProblemCodeAndSettingKey(
        this.currentDraftProblemID,
        this.contestID
      );
      const previousDraft = storage.get(draftKey) || {};
      storage.set(
        draftKey,
        {
          ...previousDraft,
          code: this.code,
          language: this.language,
        }
      );
      const individualKey = buildIndividualLanguageAndSettingKey();
      const previousIndividualSetting = storage.get(individualKey) || {};
      storage.set(individualKey, {
        ...previousIndividualSetting,
        language: this.language,
      });
    },
    getCodeTemplate(language) {
      return (this.problemData.codeTemplate || {})[language] || "";
    },
    onLanguageChange(newLanguage) {
      const oldTemplate = this.getCodeTemplate(this.language);
      const shouldReplaceCode = !this.code || this.code === oldTemplate;
      this.language = newLanguage;
      if (shouldReplaceCode) {
        this.code = this.getCodeTemplate(newLanguage);
      }
      this.saveDraft();
    },
    resetToTemplate() {
      const template = this.getCodeTemplate(this.language);
      if (!this.code || this.code === template) {
        this.code = template;
        return;
      }
      this.$confirm("确定要用当前语言的代码模板覆盖源码吗？", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      })
        .then(() => {
          this.code = template;
          this.sourceFileName = "";
          this.saveDraft();
        })
        .catch(() => {});
    },
    chooseSourceFile() {
      if (this.$refs.sourceFileInput) {
        this.$refs.sourceFileInput.click();
      }
    },
    readSourceFile(event) {
      const input = event.target;
      const file = input.files && input.files[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = (loadEvent) => {
        const text = String(loadEvent.target.result || "");
        if (text.length > 65535) {
          myMessage.error(this.$i18n.t("m.Code_Length_can_not_exceed_65535"));
        } else {
          this.code = text;
          this.sourceFileName = file.name;
          this.saveDraft();
        }
        input.value = "";
      };
      reader.onerror = () => {
        myMessage.error("源码文件读取失败");
        input.value = "";
      };
      reader.readAsText(file, "UTF-8");
    },
    handleSourceKeydown(event) {
      if (event.key !== "Tab") return;
      event.preventDefault();
      const textarea = event.target;
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const indentation = "    ";
      this.code =
        this.code.slice(0, start) + indentation + this.code.slice(end);
      this.$nextTick(() => {
        textarea.selectionStart = textarea.selectionEnd =
          start + indentation.length;
      });
    },
    async loadSubmissionStatus(problemId, requestId) {
      if (!this.isAuthenticated || !problemId) return;
      try {
        const response = await api.getUserProblemStatus(
          [problemId],
          this.isContestSubmit,
          this.contestID,
          this.groupID,
          true
        );
        if (requestId !== this.problemLoadRequestId) return;
        const status = response.data.data[problemId];
        this.submissionExists =
          status && status.status !== JUDGE_STATUS_RESERVE["ns"];
      } catch (error) {
        if (requestId === this.problemLoadRequestId) {
          this.submissionExists = false;
        }
      }
    },
    async submitCode() {
      if (!this.isAuthenticated) {
        myMessage.warning(this.$i18n.t("m.Please_login_first"));
        this.$store.dispatch("changeModalStatus", { visible: true });
        return;
      }
      if (this.isContestSubmit && !this.canSubmit) {
        myMessage.warning("请先完成比赛报名或验证后再提交");
        return;
      }
      if (!this.selectedProblemID) {
        myMessage.warning("请选择要提交的题目");
        return;
      }
      if (!this.language) {
        myMessage.warning("请选择提交语言");
        return;
      }
      if (!this.code.trim()) {
        myMessage.error(this.$i18n.t("m.Code_can_not_be_empty"));
        return;
      }
      if (this.code.length > 65535) {
        myMessage.error(this.$i18n.t("m.Code_Length_can_not_exceed_65535"));
        return;
      }

      if (
        this.isContestSubmit &&
        this.contestRuleType === RULE_TYPE.OI &&
        !this.ContestRealTimePermission &&
        this.submissionStatusPromise
      ) {
        await this.submissionStatusPromise;
      }

      if (
        this.isContestSubmit &&
        this.contestRuleType === RULE_TYPE.OI &&
        !this.ContestRealTimePermission &&
        this.submissionExists
      ) {
        try {
          await this.$confirm(
            this.$i18n.t(
              "m.You_have_submission_in_this_problem_sure_to_cover_it"
            ),
            "Warning",
            {
              confirmButtonText: this.$i18n.t("m.OK"),
              cancelButtonText: this.$i18n.t("m.Cancel"),
              type: "warning",
            }
          );
        } catch (error) {
          return;
        }
      }

      this.submitting = true;
      try {
        const response = await api.submitCode({
          pid: this.selectedProblemID,
          language: this.language,
          code: this.code,
          // 后端约定非比赛提交也必须显式传 0，不能省略或传 null。
          cid: this.contestID || 0,
          tid: this.trainingID,
          gid: this.groupID,
          isRemote: Boolean(this.problemData.problem.isRemote),
        });
        const submitID =
          response.data.data && response.data.data.submitId
            ? response.data.data.submitId
            : null;
        this.submissionExists = true;
        this.saveDraft();
        myMessage.success(this.$i18n.t("m.Submit_code_successfully"));
        this.goToMySubmission(submitID);
      } catch (error) {
        // 请求拦截器负责展示后端错误；这里显式吞掉异常，避免组件产生未处理 Promise。
        return;
      } finally {
        this.submitting = false;
      }
    },
    problemRoute() {
      const problemID = this.selectedProblemID;
      const sourceRoute = this.$route.query && this.$route.query.sourceRoute;
      const fullScreenRoutes = {
        TrainingFullProblemDetails: {
          name: "TrainingFullProblemDetails",
          params: { trainingID: this.trainingID, problemID },
        },
        ContestFullProblemDetails: {
          name: "ContestFullProblemDetails",
          params: { contestID: this.contestID, problemID },
        },
        GroupFullProblemDetails: {
          name: "GroupFullProblemDetails",
          params: { groupID: this.groupID, problemID },
        },
        GroupTrainingFullProblemDetails: {
          name: "GroupTrainingFullProblemDetails",
          params: {
            groupID: this.groupID,
            trainingID: this.trainingID,
            problemID,
          },
        },
      };
      if (sourceRoute && fullScreenRoutes[sourceRoute]) {
        return fullScreenRoutes[sourceRoute];
      }

      const routes = {
        ProblemSubmit: {
          name: "ProblemDetails",
          params: { problemID },
        },
        TrainingProblemSubmit: {
          name: "TrainingProblemDetails",
          params: { trainingID: this.trainingID, problemID },
        },
        ContestSubmit: {
          name: "ContestProblemDetails",
          params: { contestID: this.contestID, problemID },
        },
        GroupProblemSubmit: {
          name: "GroupProblemDetails",
          params: { groupID: this.groupID, problemID },
        },
        GroupTrainingProblemSubmit: {
          name: "GroupTrainingProblemDetails",
          params: {
            groupID: this.groupID,
            trainingID: this.trainingID,
            problemID,
          },
        },
      };
      return routes[this.$route.name] || routes.ProblemSubmit;
    },
    returnToProblem() {
      if (this.embedded) {
        this.$emit("cancel");
        return;
      }
      if (!this.selectedProblemID) {
        this.$router.back();
        return;
      }
      this.saveDraft();
      this.$router.push({
        ...this.problemRoute(),
        query: this.problemContextQuery(),
      });
    },
    goToMySubmission(submitID) {
      if (this.embedded) {
        this.$emit("show-my-submission", {
          submitID,
          problemID: this.selectedProblemID,
        });
        return;
      }
      const route = this.problemRoute();
      this.$router.push({
        ...route,
        query: this.problemContextQuery(true, submitID),
      });
    },
    problemContextQuery(showMySubmission = false, submitID = null) {
      const query = {};
      if (this.$route.query && this.$route.query.battle) {
        query.battle = this.$route.query.battle;
        if (submitID) {
          query.submitID = submitID;
        }
      }
      if (this.$route.query && this.$route.query.sourceRoute) {
        query.sourceRoute = this.$route.query.sourceRoute;
      }
      if (showMySubmission) {
        query.tab = "mySubmission";
      }
      return query;
    },
  },
};
</script>

<style scoped>
.problem-submit-page {
  max-width: 1180px;
  margin: 20px auto;
  padding: 0 18px 28px;
}

.submit-card {
  border-radius: 8px;
}

.problem-submit-page.is-embedded {
  max-width: none;
  margin: 0;
  padding: 24px 28px 32px;
}

.problem-submit-page.is-embedded .submit-card {
  border: 0;
  border-radius: 0;
}

.submit-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.submit-card-header h2 {
  margin: 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.problem-title {
  margin: 8px 0 0;
  color: #606266;
  font-size: 15px;
}

.submit-form {
  max-width: 1040px;
  margin: 12px auto 0;
}

.form-control {
  width: min(100%, 520px);
}

.problem-loading {
  margin-left: 12px;
  color: #909399;
  font-size: 13px;
  white-space: nowrap;
}

.source-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 12px;
}

.source-file-input {
  display: none;
}

.source-file-name {
  color: #67c23a;
  font-size: 13px;
}

.source-input /deep/ textarea {
  min-height: 440px !important;
  padding: 14px 16px;
  color: #202124;
  background: #fff;
  font-family: Menlo, Monaco, Consolas, "Courier New", monospace;
  font-size: 14px;
  line-height: 1.6;
  tab-size: 4;
}

.source-hint {
  margin-top: 8px;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}

.submit-actions {
  margin-top: 24px;
}

@media screen and (max-width: 768px) {
  .problem-submit-page {
    margin-top: 10px;
    padding: 0 8px 18px;
  }

  .submit-card-header {
    align-items: flex-start;
  }

  .submit-form {
    margin-top: 4px;
  }

  .source-input /deep/ textarea {
    min-height: 360px !important;
  }
}
</style>
