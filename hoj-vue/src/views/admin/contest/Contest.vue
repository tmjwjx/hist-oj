<template>
  <div class="view">
    <el-card>
      <div slot="header">
        <span class="panel-title home-title">
          {{ title }}
        </span>
      </div>
      <el-form label-position="top">
        <el-row :gutter="20">
          <el-col :md="10" :xs="24">
            <el-form-item
              :label="$t('m.Contest_Title')"
              required
            >
              <el-input
                v-model="contest.title"
                :placeholder="$t('m.Contest_Title')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item
              class="contest-description-editor"
              :label="$t('m.Contest_Description')"
              required
            >
              <Editor :value.sync="contest.description"></Editor>
            </el-form-item>
          </el-col>
          <el-col
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Start_Time')"
              required
            >
              <el-date-picker
                v-model="contest.startTime"
                @change="changeDuration"
                type="datetime"
                :placeholder="$t('m.Contest_Start_Time')"
              >
              </el-date-picker>
            </el-form-item>
          </el-col>
          <el-col
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_End_Time')"
              required
            >
              <el-date-picker
                v-model="contest.endTime"
                @change="changeDuration"
                type="datetime"
                :placeholder="$t('m.Contest_End_Time')"
              >
              </el-date-picker>
            </el-form-item>
          </el-col>

          <el-col
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Duration')"
              required
            >
              <el-input
                v-model="durationText"
                disabled
              > </el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <ContestRegistrationConfig :contest="contest" />

        <section class="contest-options">
          <div class="contest-options-title">赛制、榜单、权限与奖项</div>

        <el-row>
          <el-col
            class="contest-grid-rule"
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Rule_Type')"
              required
            >
              <el-select v-model="contest.type" @change="setSealRankTimeDefaultValue"
                :disabled="disableRuleType" class="compact-select">
                <el-option label="ACM" :value="0" />
                <el-option label="OI" :value="1" />
              </el-select>
            </el-form-item>
          </el-col>

          <el-col
            class="contest-grid-conditional"
            :md="8"
            :xs="24"
            v-show="contest.type == 1"
          >
            <el-form-item
              :label="$t('m.OI_Rank_Score_Type')"
            >
              <el-select v-model="contest.oiRankScoreType" class="compact-select">
                <el-option :label="$t('m.OI_Rank_Score_Type_Recent')" value="Recent" />
                <el-option :label="$t('m.OI_Rank_Score_Type_Highest')" value="Highest" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row>
          <el-col
            class="contest-grid-seal"
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Timeliness_Of_Rank')"
              required
            >
              <el-switch
                v-model="contest.sealRank"
                active-color="#13ce66"
                :active-text="$t('m.Seal_Time_Rank')"
                :inactive-text="$t('m.Real_Time_Rank')"
                @change="setSealRankTimeDefaultValue"
              >
              </el-switch>
            </el-form-item>
          </el-col>

          <el-col
            class="contest-grid-conditional"
            :md="8"
            :xs="24"
            v-show="contest.sealRank"
          >
            <el-form-item
              :label="$t('m.Seal_Rank_Time')"
              :required="contest.sealRank"
            >
              <el-select v-model="seal_rank_time">
                <el-option
                  :label="$t('m.Contest_Seal_Half_Hour')"
                  :value="0"
                  :disabled="contest.duration < 1800"
                ></el-option>
                <el-option
                  :label="$t('m.Contest_Seal_An_Hour')"
                  :value="1"
                  :disabled="contest.duration < 3600"
                ></el-option>
                <el-option
                  :label="$t('m.Contest_Seal_All_Hour')"
                  :value="2"
                ></el-option>
              </el-select>
            </el-form-item>
          </el-col>

          <el-col
            class="contest-grid-conditional"
            :md="8"
            :xs="24"
            v-show="contest.sealRank"
          >
            <el-form-item
              :label="$t('m.Auto_Real_Rank')"
              required
            >
              <el-switch
                v-model="contest.autoRealRank"
                :active-text="$t('m.Real_Rank_After_Contest')"
                :inactive-text="$t('m.Seal_Rank_After_Contest')"
              >
              </el-switch>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row>
          <el-col
            class="contest-grid-outside"
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Outside_ScoreBoard')"
              required
            >
              <el-switch
                v-model="contest.openRank"
                :active-text="$t('m.Open')"
                :inactive-text="$t('m.Close')"
              >
              </el-switch>
            </el-form-item>
          </el-col>

          <el-col
            class="contest-grid-end-submit"
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Allow_Submission_After_The_Contest_Ends')"
              required
            >
              <el-switch
                v-model="contest.allowEndSubmit"
                :active-text="$t('m.Open')"
                :inactive-text="$t('m.Close')"
              >
              </el-switch>
            </el-form-item>
          </el-col>

          <el-col
            class="contest-grid-print"
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Print_Func')"
              required
            >
              <el-switch
                v-model="contest.openPrint"
                :active-text="$t('m.Support_Offline_Print')"
                :inactive-text="$t('m.Not_Support_Print')"
              >
              </el-switch>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row>
          <el-col class="contest-grid-rank-name">
            <el-form-item :label="$t('m.Rank_Show_Name')" required>
              <el-select v-model="contest.rankShowName" class="compact-select">
                <el-option label="用户名" value="username" />
                <el-option label="昵称" value="nickname" />
                <el-option label="真实姓名" value="realname" />
              </el-select>
            </el-form-item>
          </el-col>

          <el-col class="star-user-setting contest-grid-star">
            <el-form-item
              :label="$t('m.Star_User_UserName')"
              required
            >
              <el-tag
                v-for="username in contest.starAccount"
                closable
                :close-transition="false"
                :key="username"
                type="warning"
                size="medium"
                @close="removeStarUser(username)"
                style="margin-right: 7px;margin-top:4px"
              >{{ username }}</el-tag>
              <el-input
                v-if="inputVisible"
                size="medium"
                class="input-new-star-user"
                v-model="starUserInput"
                :trigger-on-focus="true"
                @keyup.enter.native="addStarUser"
                @blur="addStarUser"
              >
              </el-input>
              <el-tooltip
                effect="dark"
                :content="$t('m.Add')"
                placement="top"
                v-else
              >
                <el-button
                  class="button-new-tag"
                  size="small"
                  @click="inputVisible = true"
                  icon="el-icon-plus"
                ></el-button>
              </el-tooltip>
            </el-form-item>
          </el-col>

          <el-col
            class="contest-grid-auth"
            :md="8"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Auth')"
              required
            >
              <el-select v-model="contest.auth">
                <el-option
                  :label="$t('m.Public')"
                  :value="0"
                ></el-option>
                <el-option
                  :label="$t('m.Private')"
                  :value="1"
                ></el-option>
                <el-option
                  :label="$t('m.Protected')"
                  :value="2"
                ></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col
            class="contest-grid-conditional"
            :md="8"
            :xs="24"
            v-show="contest.auth != 0"
          >
            <el-form-item
              :label="$t('m.Contest_Password')"
              :required="contest.auth != 0"
            >
              <el-input
                v-model="contest.pwd"
                :placeholder="$t('m.Contest_Password')"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col
            class="contest-grid-conditional"
            :md="8"
            :xs="24"
            v-show="contest.auth != 0"
          >
            <el-form-item
              :label="$t('m.Account_Limit')"
              :required="contest.auth != 0"
            >
              <el-switch v-model="contest.openAccountLimit"> </el-switch>
            </el-form-item>
          </el-col>

          <template v-if="contest.openAccountLimit">
            <el-form :model="formRule" class="account-rule-form contest-full-row">
              <el-col
                :md="6"
                :xs="24"
              >
                <el-form-item
                  :label="$t('m.Prefix')"
                  prop="prefix"
                >
                  <el-input
                    v-model="formRule.prefix"
                    placeholder="Prefix"
                  ></el-input>
                </el-form-item>
              </el-col>
              <el-col
                :md="6"
                :xs="24"
              >
                <el-form-item
                  :label="$t('m.Suffix')"
                  prop="suffix"
                >
                  <el-input
                    v-model="formRule.suffix"
                    placeholder="Suffix"
                  ></el-input>
                </el-form-item>
              </el-col>
              <el-col
                :md="6"
                :xs="24"
              >
                <el-form-item
                  :label="$t('m.Start_Number')"
                  prop="number_from"
                >
                  <el-input-number
                    v-model="formRule.number_from"
                    style="width: 100%"
                  ></el-input-number>
                </el-form-item>
              </el-col>
              <el-col
                :md="6"
                :xs="24"
              >
                <el-form-item
                  :label="$t('m.End_Number')"
                  prop="number_to"
                >
                  <el-input-number
                    v-model="formRule.number_to"
                    style="width: 100%"
                  ></el-input-number>
                </el-form-item>
              </el-col>

              <div
                class="userPreview"
                v-if="formRule.number_from <= formRule.number_to"
              >
                {{ $t('m.The_allowed_account_will_be') }}
                {{ formRule.prefix + formRule.number_from + formRule.suffix }},
                <span v-if="formRule.number_from + 1 < formRule.number_to">
                  {{
                    formRule.prefix +
                      (formRule.number_from + 1) +
                      formRule.suffix +
                      '...'
                  }}
                </span>
                <span v-if="formRule.number_from + 1 <= formRule.number_to">
                  {{ formRule.prefix + formRule.number_to + formRule.suffix }}
                </span>
              </div>

              <el-col
                :md="24"
                :xs="24"
                class="extra-account-field"
              >
                <el-form-item
                  :label="$t('m.Extra_Account')"
                  prop="prefix"
                >
                  <el-input
                    type="textarea"
                    :placeholder="$t('m.Extra_Account_Tips')"
                    :rows="8"
                    v-model="formRule.extra_account"
                  >
                  </el-input>
                </el-form-item>
              </el-col>
            </el-form>
          </template>

          <el-col
            class="contest-grid-award"
            :md="24"
            :xs="24"
          >
            <el-form-item
              :label="$t('m.Contest_Award')"
              required
            >
              <el-select
                v-model="contest.awardType"
                @change="contestAwardTypeChange"
              >
                <el-option
                  :label="$t('m.Contest_Award_Null')"
                  :value="0"
                ></el-option>
                <el-option
                  :label="$t('m.Contest_Award_Set_Proportion')"
                  :value="1"
                ></el-option>
                <el-option
                  :label="$t('m.Contest_Award_Set_Number')"
                  :value="2"
                ></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col
            :span="24"
            v-if="contest.awardType != 0"
            class="contest-full-row award-config-table"
          >
            <div style="margin-bottom:10px">
              <el-button
                type="primary"
                icon="el-icon-plus"
                circle
                @click="insertEvent(-1)"
                size="small"
              ></el-button>
              <el-button
                type="danger"
                icon="el-icon-delete"
                circle
                @click="removeEvent()"
                size="small"
              ></el-button>
            </div>
            <vxe-table
              border
              ref="xAwardTable"
              :data="contest.awardConfigList"
              :edit-config="{trigger: 'click', mode: 'cell'}"
              :sort-config="{trigger: 'cell', defaultSort: {field: 'priority', order: 'asc'}, orders: ['desc', 'asc', null]}"
              align="center"
              @edit-closed="editClosedEvent"
              style="margin-bottom:15px"
            >
              <vxe-table-column
                type="checkbox"
                width="60"
              ></vxe-table-column>
              <vxe-table-column
                field="priority"
                width="100"
                :title="$t('m.Contest_Award_Priority')"
                :edit-render="{name: 'input', attrs: {type: 'number'}}"
                sortable
              >
              </vxe-table-column>
              <vxe-table-column
                field="name"
                min-width="150"
                :title="$t('m.Contest_Award_Name')"
                :edit-render="{name: 'input', attrs: {type: 'text'}}"
              >
              </vxe-table-column>
              <vxe-table-column
                field="background"
                min-width="150"
                :title="$t('m.Contest_Award_Background')"
              >
                <template v-slot="{ row }">
                  <el-color-picker
                    v-model="row.background"
                    size="small"
                  ></el-color-picker>
                </template>
              </vxe-table-column>
              <vxe-table-column
                field="color"
                min-width="150"
                :title="$t('m.Contest_Award_Color')"
              >
                <template v-slot="{ row }">
                  <el-color-picker
                    v-model="row.color"
                    size="small"
                  ></el-color-picker>
                </template>
              </vxe-table-column>

              <vxe-table-column
                field="show"
                min-width="150"
                :title="$t('m.Contest_Award_Show')"
              >
                <template v-slot="{ row }">
                  <RankBox
                    :name="row.name"
                    :background="row.background"
                    :color="row.color"
                    :num="1"
                  ></RankBox>
                </template>
              </vxe-table-column>

              <vxe-table-column
                field="num"
                min-width="150"
                :title="contest.awardType == 1?$t('m.Contest_Award_Proportion'):$t('m.Contest_Award_Number')"
              >
                <template v-slot="{ row }">

                  <el-input
                    :placeholder="$t('m.Contest_Award_Proportion')"
                    v-model="row.num"
                    size="small"
                    v-if="contest.awardType == 1"
                    type="number"
                  >
                    <template slot="append">%</template>
                  </el-input>
                  <el-input
                    :placeholder="$t('m.Contest_Award_Number')"
                    v-model="row.num"
                    size="small"
                    v-else
                  >
                  </el-input>
                </template>
              </vxe-table-column>
            </vxe-table>
          </el-col>
        </el-row>
        </section>
      </el-form>
      <el-button
        type="primary"
        @click.native="saveContest"
      >{{
        $t('m.Save')
      }}</el-button>
    </el-card>
  </div>
</template>

<script>
import api from "@/common/api";
import time from "@/common/time";
import moment from "moment";
import { mapGetters } from "vuex";
import myMessage from "@/common/message";
const Editor = () => import("@/components/admin/Editor.vue");
const RankBox = () => import("@/components/oj/common/RankBox");
const ContestRegistrationConfig = () =>
  import("@/components/admin/ContestRegistrationConfig.vue");
export default {
  name: "CreateContest",
  components: {
    Editor,
    RankBox,
    ContestRegistrationConfig,
  },
  data() {
    return {
      title: "Create Contest",
      disableRuleType: false,
      durationText: "", // 比赛时长文本表示
      seal_rank_time: 2, // 当开启封榜模式，即实时榜单关闭时，可选择前半小时，前一小时，全程封榜,默认全程封榜
      contest: {
        title: "",
        description: "",
        startTime: "",
        endTime: "",
        duration: 0,
        type: 0,
        pwd: "",
        sealRank: false,
        sealRankTime: "", //封榜时间
        autoRealRank: true,
        auth: 0,
        openPrint: false,
        rankShowName: "username",
        openAccountLimit: false,
        allowEndSubmit: false,
        openRegistration: false,
        registrationFields: [],
        useRegistrationName: false,
        registrationNameFields: [],
        visible: false,
        accountLimitRule: "",
        starAccount: [],
        oiRankScoreType: "Recent",
        awardType: 0,
        awardConfigList: [
          {
            priority: 1,
            name: "金牌",
            background: "#e6bf25",
            color: "#fff",
            num: 10,
          },
          {
            priority: 2,
            name: "银牌",
            background: "#b4c0c7",
            color: "#fff",
            num: 20,
          },
          {
            priority: 3,
            name: "铜牌",
            background: "#CD7F32",
            color: "#fff",
            num: 30,
          },
        ],
      },
      formRule: {
        prefix: "",
        suffix: "",
        number_from: 0,
        number_to: 10,
        extra_account: "",
      },
      starUserInput: "",
      inputVisible: false,
    };
  },
  mounted() {
    if (this.$route.name === "admin-edit-contest") {
      this.title = this.$i18n.t("m.Edit_Contest");
      this.disableRuleType = true;
      this.getContestByCid();
    } else {
      this.title = this.$i18n.t("m.Create_Contest");
      this.disableRuleType = false;
    }
  },
  watch: {
    $route() {
      if (this.$route.name === "admin-edit-contest") {
        this.title = this.$i18n.t("m.Edit_Contest");
        this.disableRuleType = true;
        this.getContestByCid();
      } else {
        this.title = this.$i18n.t("m.Create_Contest");
        this.disableRuleType = false;
        this.contest = {};
      }
    },
  },
  computed: {
    ...mapGetters(["userInfo"]),
  },
  methods: {
    getContestByCid() {
      api
        .admin_getContest(this.$route.params.contestId)
        .then((res) => {
          let data = res.data.data;
      this.contest = data;
          this.contest.registrationFields = this.contest.registrationFields || [];
          this.contest.registrationNameFields = this.contest.registrationNameFields || [];
          this.contest.openRegistration = Boolean(this.contest.openRegistration);
          this.contest.useRegistrationName = Boolean(this.contest.useRegistrationName);
          this.changeDuration();
          // 封榜时间转换
          let halfHour = moment(this.contest.endTime)
            .subtract(1800, "seconds")
            .toString();
          let oneHour = moment(this.contest.endTime)
            .subtract(3600, "seconds")
            .toString();
          let allHour = moment(this.contest.startTime).toString();
          let sealRankTime = moment(this.contest.sealRankTime).toString();
          switch (sealRankTime) {
            case halfHour:
              this.seal_rank_time = 0;
              break;
            case oneHour:
              this.seal_rank_time = 1;
              break;
            case allHour:
              this.seal_rank_time = 2;
              break;
          }
          if (this.contest.accountLimitRule) {
            this.formRule = this.changeStrToAccountRule(
              this.contest.accountLimitRule
            );
          }
        })
        .catch(() => {});
    },

    saveContest() {
      if (!this.contest.title) {
        myMessage.error(
          this.$i18n.t("m.Contest_Title") + " " + this.$i18n.t("m.is_required")
        );
        return;
      }
      if (!this.contest.description) {
        myMessage.error(
          this.$i18n.t("m.Contest_Description") +
            " " +
            this.$i18n.t("m.is_required")
        );
        return;
      }
      if (!this.contest.startTime) {
        myMessage.error(
          this.$i18n.t("m.Contest_Start_Time") +
            " " +
            this.$i18n.t("m.is_required")
        );
        return;
      }
      if (!this.contest.endTime) {
        myMessage.error(
          this.$i18n.t("m.Contest_End_Time") +
            " " +
            this.$i18n.t("m.is_required")
        );
        return;
      }
      if (!this.contest.duration || this.contest.duration <= 0) {
        myMessage.error(this.$i18n.t("m.Contest_Duration_Check"));
        return;
      }
      if (this.contest.auth != 0 && !this.contest.pwd) {
        myMessage.error(
          this.$i18n.t("m.Contest_Password") +
            " " +
            this.$i18n.t("m.is_required")
        );
        return;
      }
      if (this.contest.openRegistration && !this.contest.registrationFields.length) {
        myMessage.error("开启比赛报名后，至少选择一个报名字段");
        return;
      }
      if (
        this.contest.useRegistrationName &&
        !this.contest.registrationNameFields.length
      ) {
        myMessage.error("请选择比赛内名称的组合字段");
        return;
      }

      if (this.contest.openAccountLimit) {
        this.contest.accountLimitRule = this.changeAccountRuleToStr(
          this.formRule
        );
      }

      let funcName =
        this.$route.name === "admin-edit-contest"
          ? "admin_editContest"
          : "admin_createContest";

      switch (this.seal_rank_time) {
        case 0: // 结束前半小时
          this.contest.sealRankTime = moment(this.contest.endTime).subtract(
            1800,
            "seconds"
          );
          break;
        case 1: // 结束前一小时
          this.contest.sealRankTime = moment(this.contest.endTime).subtract(
            3600,
            "seconds"
          );
          break;
        case 2: // 全程
          this.contest.sealRankTime = moment(this.contest.startTime);
      }
      let data = Object.assign({}, this.contest);
      if (funcName === "admin_createContest") {
        data["uid"] = this.userInfo.uid;
        data["author"] = this.userInfo.username;
      }

      api[funcName](data)
        .then((res) => {
          myMessage.success("success");
          this.$router.push({
            name: "admin-contest-list",
            query: { refresh: "true" },
          });
        })
        .catch(() => {});
    },
    changeDuration() {
      let start = this.contest.startTime;
      let end = this.contest.endTime;
      let durationMS = time.durationMs(start, end);
      if (durationMS < 0) {
        this.durationText = this.$i18n.t("m.Contets_Time_Check");
        this.contest.duration = 0;
        return;
      }
      if (start != "" && end != "") {
        this.durationText = time.formatSpecificDuration(start, end);
        this.contest.duration = durationMS;
      }
    },
    changeAccountRuleToStr(formRule) {
      let result =
        "<prefix>" +
        formRule.prefix +
        "</prefix><suffix>" +
        formRule.suffix +
        "</suffix><start>" +
        formRule.number_from +
        "</start><end>" +
        formRule.number_to +
        "</end><extra>" +
        formRule.extra_account +
        "</extra>";
      return result;
    },
    changeStrToAccountRule(value) {
      let reg =
        "<prefix>([\\s\\S]*?)</prefix><suffix>([\\s\\S]*?)</suffix><start>([\\s\\S]*?)</start><end>([\\s\\S]*?)</end><extra>([\\s\\S]*?)</extra>";
      let re = RegExp(reg, "g");
      let tmp = re.exec(value);
      return {
        prefix: tmp[1],
        suffix: tmp[2],
        number_from: tmp[3],
        number_to: tmp[4],
        extra_account: tmp[5],
      };
    },

    addStarUser() {
      this.starUserInput = this.starUserInput.replace(/(^\s*)|(\s*$)/g, "");
      if (this.starUserInput) {
        for (var i = 0; i < this.contest.starAccount.length; i++) {
          if (this.contest.starAccount[i] == this.starUserInput) {
            myMessage.warning(this.$i18n.t("m.Add_Star_User_Error"));
            this.starUserInput = "";
            return;
          }
        }
        this.contest.starAccount.push(this.starUserInput);
        this.inputVisible = false;
        this.starUserInput = "";
      }
    },

    // 根据UserName 从打星用户列表中移除
    removeStarUser(username) {
      this.contest.starAccount.splice(
        this.contest.starAccount.map((item) => item.name).indexOf(username),
        1
      );
    },

    setSealRankTimeDefaultValue() {
      if (this.contest.sealRank == true) {
        if (this.contest.type == 0) {
          // ACM比赛开启封榜 默认为一小时,如果比赛时长小于一小时，则默认为全程
          if (this.contest.duration < 3600) {
            this.seal_rank_time = 2;
          } else {
            this.seal_rank_time = 1;
          }
        } else {
          // OI比赛开启封榜 默认全程
          this.seal_rank_time = 2;
        }
      }
    },

    contestAwardTypeChange() {
      if (this.contest.awardType != 0 && !this.contest.awardConfigList.length) {
        this.contest.awardConfigList = [
          {
            priority: 1,
            name: "金牌",
            background: "#e6bf25",
            color: "#fff",
            num: 10,
          },
          {
            priority: 2,
            name: "银牌",
            background: "#b4c0c7",
            color: "#fff",
            num: 20,
          },
          {
            priority: 3,
            name: "铜牌",
            background: "#CD7F32",
            color: "#fff",
            num: 30,
          },
        ];
      }
    },

    async insertEvent(row) {
      let record = {
        priority: this.contest.awardConfigList.length + 1,
        name: "name",
        background: "#ededed",
        color: "#666",
        num: 0,
      };
      let { row: newRow } = await this.$refs.xAwardTable.insertAt(record, row);
      const { insertRecords } = this.$refs.xAwardTable.getRecordset();
      this.contest.awardConfigList =
        this.contest.awardConfigList.concat(insertRecords);
      await this.$refs.xAwardTable.setActiveCell(newRow, "name");
    },

    async removeEvent() {
      this.$refs.xAwardTable.removeCheckboxRow();
      let removeRecords = this.$refs.xAwardTable.getRemoveRecords();
      function getDifferenceSetB(arr1, arr2, typeName) {
        return Object.values(
          arr1.concat(arr2).reduce((acc, cur) => {
            if (
              acc[cur[typeName]] &&
              acc[cur[typeName]][typeName] === cur[typeName]
            ) {
              delete acc[cur[typeName]];
            } else {
              acc[cur[typeName]] = cur;
            }
            return acc;
          }, {})
        );
      }
      this.contest.awardConfigList = getDifferenceSetB(
        this.contest.awardConfigList,
        removeRecords,
        "_XID"
      );
    },

    editClosedEvent({ row, column }) {
      let xTable = this.$refs.xAwardTable;
      let field = column.property;
      // 判断单元格值是否被修改
      if (xTable.isUpdateByRow(row, field)) {
        setTimeout(() => {
          // 局部更新单元格为已保存状态
          this.$refs.xAwardTable.reloadRow(row, null, field);
        }, 300);
      }
    },
  },
};
</script>
<style scoped>
/* 页面整体 */
.view {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.view /deep/ .el-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  margin-bottom: 20px;
}

.view /deep/ .el-card__header {
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  padding: 16px 20px;
}

.view /deep/ .el-card__body {
  padding: 24px;
}

/* 表单项 */
.view /deep/ .el-form-item {
  margin-bottom: 18px;
}

.view /deep/ .el-form-item__label {
  padding-bottom: 8px;
  line-height: 1.5;
  font-weight: 500;
  color: #606266;
}

/* 编辑器 */
.view /deep/ .mavon-editor,
.view /deep/ .contest-description-editor .v-note-wrapper {
  height: 300px;
  min-height: 280px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.view /deep/ .v-note-panel,
.view /deep/ .v-note-edit.divarea {
  min-height: 250px;
}

/* 日期选择器 */
.view /deep/ .el-date-editor.el-input,
.view /deep/ .el-date-editor.el-input__inner {
  width: 100%;
}

/* 分组标题 */
.contest-options {
  margin-top: 24px;
  padding: 20px;
  background: #f9fafb;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

.contest-options-title {
  margin-bottom: 20px;
  color: #303133;
  font-weight: 600;
  font-size: 16px;
  line-height: 1.5;
  padding-bottom: 12px;
  border-bottom: 2px solid #409eff;
}

/* 核心设置固定为四列两行，条件设置继续在下方按需显示。 */
.contest-options {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px 20px;
  align-items: start;
}

.contest-options-title {
  grid-column: 1 / -1;
  margin-bottom: 2px;
}

.contest-options > .el-row {
  display: contents;
}

.contest-options .el-col {
  float: none;
  width: auto !important;
  min-width: 0;
  padding: 0 !important;
}

.contest-options .contest-grid-rule,
.contest-options .contest-grid-rank-name,
.contest-options .contest-grid-auth,
.contest-options .contest-grid-award,
.contest-options .contest-grid-seal,
.contest-options .contest-grid-outside,
.contest-options .contest-grid-end-submit,
.contest-options .contest-grid-print,
.contest-options .contest-grid-star {
  grid-row: auto;
}

.contest-options .contest-grid-rule { grid-column: 1; grid-row: 2; }
.contest-options .contest-grid-rank-name { grid-column: 2; grid-row: 2; }
.contest-options .contest-grid-auth { grid-column: 3; grid-row: 2; }
.contest-options .contest-grid-award { grid-column: 4; grid-row: 2; }
.contest-options .contest-grid-seal { grid-column: 1; grid-row: 3; }
.contest-options .contest-grid-outside { grid-column: 2; grid-row: 3; }
.contest-options .contest-grid-end-submit { grid-column: 3; grid-row: 3; }
.contest-options .contest-grid-print { grid-column: 4; grid-row: 3; }
.contest-options .contest-grid-star { grid-column: 1 / -1; grid-row: 4; }
.contest-options .contest-grid-conditional { grid-column: 1 / -1; }

.contest-options .contest-full-row,
.contest-options .account-rule-form,
.contest-options .award-config-table {
  grid-column: 1 / -1;
  min-width: 0 !important;
}

.contest-options .el-form-item {
  margin-bottom: 4px;
}

.contest-options .el-select,
.contest-options .el-input {
  width: 100%;
}

.contest-options .el-switch {
  white-space: nowrap;
}

/* 账号规则单独成一行，并在行内继续保持四列紧凑布局。 */
.contest-options .account-rule-form {
  display: grid !important;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0 16px;
  margin: 0;
  padding: 12px 14px 4px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.contest-options .account-rule-form > .el-col {
  float: none;
  width: auto !important;
  min-width: 0;
  padding: 0 !important;
}

.contest-options .account-rule-form .userPreview,
.contest-options .account-rule-form .extra-account-field {
  grid-column: 1 / -1;
}

.contest-options .account-rule-form .userPreview {
  margin: 2px 0 10px;
}

/* 跨列元素 */
.contest-options .userPreview {
  width: 100%;
}

.contest-options .award-config-table {
  overflow-x: auto;
}

.contest-options .star-user-setting {
  min-width: 0;
}

/* 紧凑选择器 */
.compact-select {
  width: 100%;
}

/* 用户预览 */
.userPreview {
  padding: 12px 16px;
  margin-top: 12px;
  background: #fef0f0;
  border: 1px solid #fbc4c4;
  border-radius: 4px;
  color: #f56c6c;
  font-size: 13px;
  line-height: 1.6;
}

/* 星标用户输入 */
.input-new-star-user {
  width: 200px;
}

/* 响应式 */
@media (max-width: 1200px) {
  .contest-options {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .contest-options .account-rule-form {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .contest-options .contest-grid-rule,
  .contest-options .contest-grid-rank-name,
  .contest-options .contest-grid-auth,
  .contest-options .contest-grid-award,
  .contest-options .contest-grid-seal,
  .contest-options .contest-grid-outside,
  .contest-options .contest-grid-end-submit,
  .contest-options .contest-grid-print,
  .contest-options .contest-grid-star,
  .contest-options .contest-grid-conditional {
    grid-column: auto;
    grid-row: auto;
  }
}

@media (max-width: 768px) {
  .view {
    padding: 10px;
  }

  .view /deep/ .el-card__body {
    padding: 16px;
  }

  .contest-options {
    padding: 16px;
    grid-template-columns: 1fr;
  }

  .contest-options .contest-full-row,
  .contest-options .account-rule-form,
  .contest-options .award-config-table,
  .contest-options-title {
    grid-column: 1;
  }

  .contest-options .account-rule-form {
    grid-template-columns: 1fr;
  }
}

/* 按钮样式优化 */
.view /deep/ .el-button {
  border-radius: 4px;
}

/* 输入框优化 */
.view /deep/ .el-input__inner,
.view /deep/ .el-textarea__inner {
  border-radius: 4px;
}

/* 标签优化 */
.view /deep/ .el-tag {
  border-radius: 2px;
}
</style>
