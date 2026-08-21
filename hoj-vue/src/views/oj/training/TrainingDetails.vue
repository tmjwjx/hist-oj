<template>
  <div>
    <el-card shadow class="training-header">
      <div slot="header">
        <span class="panel-title">{{ training.title }}</span>
      </div>
      <div class="training-progress" v-if="isAuthenticated">
        <el-tooltip
          effect="dark"
          :content="`已完成 ${training.acCount || 0} / ${training.problemCount || 0} 题目`"
          placement="top"
        >
          <el-progress
            :text-inside="true"
            :stroke-width="20"
            :percentage="parseFloat(getAcProblemPercent())"
          ></el-progress>
        </el-tooltip>
      </div>
    </el-card>
    <div class="card-top">
      <el-tabs @tab-click="tabClick" v-model="route_name">
        <el-tab-pane :name="groupID?'GroupTrainingDetails':'TrainingDetails'" lazy>
          <span slot="label"
            ><i class="el-icon-s-home"></i>&nbsp;{{
              $t('m.Training_Introduction')
            }}</span
          >
          <el-row :gutter="30">
            <el-col :sm="24" :md="7">
              <el-card
                v-if="trainingPasswordFormVisible"
                class="password-form-card"
              >
                <div slot="header">
                  <span class="panel-title" style="color: #e6a23c;"
                    ><i class="el-icon-warning">
                      {{ $t('m.Password_Required') }}</i
                    ></span
                  >
                </div>
                <h3>
                  {{ $t('m.To_Enter_Training_Need_Password') }}
                </h3>
                <el-form>
                  <el-input
                    v-model="trainingPassword"
                    type="password"
                    :placeholder="$t('m.Enter_the_training_password')"
                    @keydown.enter.native="checkPassword"
                    style="width:70%"
                  />
                  <el-button
                    type="primary"
                    @click="checkPassword"
                    :loading="btnLoading"
                    style="margin:5px"
                    >{{ $t('m.OK') }}</el-button
                  >
                </el-form>
              </el-card>
              <el-card>
                <div class="info-rows">
                  <div>
                    <span>
                      <span>{{ $t('m.Training_Number') }}</span>
                    </span>
                    <span>
                      <span>{{ training.rank }}</span>
                    </span>
                  </div>
                  <div>
                    <span>
                      <span>{{ $t('m.Training_Auth') }}</span>
                    </span>
                    <span v-if="training.auth">
                      <el-tag
                        :type="TRAINING_TYPE[training.auth]['color']"
                        effect="dark"
                      >
                        {{ $t('m.Training_' + training.auth) }}
                      </el-tag>
                    </span>
                  </div>
                  <div>
                    <span>
                      <span>{{ $t('m.Training_Category') }}</span>
                    </span>
                    <span>
                      <span
                        ><el-tag
                          size="medium"
                          class="category-item"
                          :style="
                            'color: #fff;background-color: ' +
                              training.categoryName +
                              ';background-color: ' +
                              training.categoryColor
                          "
                          >{{ training.categoryName }}</el-tag
                        ></span
                      >
                    </span>
                  </div>

                  <div>
                    <span>
                      <span>{{ $t('m.Training_Total_Problems') }}</span>
                    </span>
                    <span>
                      <span>{{ training.problemCount }}</span>
                    </span>
                  </div>
                  <div>
                    <span>
                      <span>{{ $t('m.Author') }}</span>
                    </span>
                    <span>
                      <span
                        ><el-link
                          type="info"
                          @click="goUserHome(training.author)"
                          >{{ training.author }}</el-link
                        ></span
                      >
                    </span>
                  </div>
                  <div>
                    <span>
                      <span>{{ $t('m.Recent_Update') }}</span>
                    </span>
                    <span>
                      <span>{{ training.gmtModified | localtime }}</span>
                    </span>
                  </div>
                  <div v-if="isAuthenticated">
                    <span>
                      <span>训练状态</span>
                    </span>
                    <span>
                      <template v-if="trainingPasswordFormVisible">
                        <el-tag type="warning">需要密码验证</el-tag>
                      </template>
                      <template v-else-if="!hasJoined">
                        <el-button
                          type="primary"
                          size="small"
                          @click="handleJoinTraining"
                          :loading="joinLoading"
                        >
                          参加训练
                        </el-button>
                      </template>
                      <template v-else>
                        <el-tag type="success">已参加</el-tag>
                      </template>
                    </span>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :sm="24" :md="17">
              <el-card>
                <div slot="header">
                  <span class="panel-title">{{
                    $t('m.Training_Introduction')
                  }}</span>
                </div>
                <Markdown 
                  :isAvoidXss="groupID" 
                  :content="training.description">
                </Markdown>
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>

        <el-tab-pane
          :name="groupID?'GroupTrainingProblemList':'TrainingProblemList'"
          lazy
          :disabled="trainingMenuDisabled || !hasJoined"
        >
          <span slot="label"
            ><i class="fa fa-list" aria-hidden="true"></i>&nbsp;{{
              $t('m.Problem_List')
            }}</span
          >
          <transition name="el-zoom-in-bottom">
            <router-view
              v-if="(route_name === 'TrainingProblemList' || route_name === 'GroupTrainingProblemList') && hasJoined"
            ></router-view>
          </transition>
        </el-tab-pane>

        <el-tab-pane
          :name="groupID?'GroupTrainingRank':'TrainingRank'"
          lazy
          :disabled="trainingMenuDisabled || !hasJoined"
          v-if="isPrivateTraining"
        >
          <span slot="label"
            ><i class="fa fa-bar-chart" aria-hidden="true"></i>&nbsp;{{
              $t('m.Record_List')
            }}</span
          >
          <transition name="el-zoom-in-bottom">
            <router-view v-if="(route_name === 'TrainingRank' || route_name === 'GroupTrainingRank') && hasJoined"></router-view>
          </transition>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
import { TRAINING_TYPE } from '@/common/constants';
import { mapState, mapGetters, mapActions } from 'vuex';
import myMessage from '@/common/message';
import api from '@/common/api';
import Markdown from '@/components/oj/common/Markdown';
import { joinTraining, getMyTrainingRecord } from '@/api/training';
export default {
  components: {
    Markdown
  },
  data() {
    return {
      route_name: 'TrainingDetails',
      TRAINING_TYPE: {},
      trainingPassword: '',
      btnLoading: false,
      joinLoading: false,
      groupID:null,
      hasJoined: false,
    };
  },
  created() {
    this.route_name = this.$route.name;
    let gid = this.$route.params.groupID;
    if(gid){
      this.groupID = gid;
      if (this.route_name == 'GroupTrainingProblemDetails') {
        this.route_name = 'GroupTrainingProblemList';
      }
    }else{
      if (this.route_name == 'TrainingProblemDetails') {
        this.route_name = 'TrainingProblemList';
      }
    }
    this.TRAINING_TYPE = Object.assign({}, TRAINING_TYPE);
    this.$store.dispatch('getTraining').then((res) => {
      this.changeDomTitle({ title: res.data.data.title });
      // 检查用户是否已参加训练
      if (this.isAuthenticated) {
        this.checkJoinStatus();
      }
    });
  },
  methods: {
    ...mapActions(['changeDomTitle']),
    // 查询不到个人记录表示尚未参加，是正常页面状态。
    checkJoinStatus() {
      // 如果需要密码验证，不检查参加状态
      if (this.trainingPasswordFormVisible) {
        this.hasJoined = false;
        return;
      }

      // 调用 API 查询 training_participant 表，检查用户是否已报名
      const trainingId = this.$route.params.trainingID;
      getMyTrainingRecord(trainingId).then(
        (res) => {
          this.hasJoined = Boolean(
            res.data && res.data.code === 200 && res.data.data
          );
        },
        (err) => {
          // 404 错误，用户不在 training_participant 表中（未报名）
          this.hasJoined = false;
        }
      );
    },
    handleJoinTraining() {
      // 如果需要密码验证，提示用户
      if (this.trainingPasswordFormVisible) {
        myMessage.warning('请先输入训练密码');
        // 滚动到密码输入框
        this.$nextTick(() => {
          const passwordCard = document.querySelector('.password-form-card');
          if (passwordCard) {
            passwordCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        });
        return;
      }
      
      this.joinLoading = true;
      const trainingId = this.$route.params.trainingID;
      joinTraining(trainingId).then(
        (res) => {
          this.joinLoading = false;
          if (res.data.code === 200) {
            myMessage.success('参加训练成功！');
            this.hasJoined = true;
            // 刷新训练信息以更新进度
            this.$store.dispatch('getTraining');
          } else {
            myMessage.error(res.data.message || '参加训练失败');
          }
        },
        (err) => {
          this.joinLoading = false;
          myMessage.error(err.response?.data?.message || '参加训练失败');
        }
      );
    },
    tabClick(tab) {
      let name = tab.name;
      if (name !== this.$route.name) {
        this.$router.push({ name: name });
      }
    },
    checkPassword() {
      if (this.trainingPassword === '') {
        myMessage.warning(this.$i18n.t('m.Enter_the_training_password'));
        return;
      }
      this.btnLoading = true;
      const trainingId = this.$route.params.trainingID;
      
      // 先验证密码
      api.registerTraining(this.training.id + '', this.trainingPassword).then(
        (res) => {
          // 密码验证成功，直接完成报名
          myMessage.success('密码验证成功，正在加入训练...');
          this.$store.commit('trainingIntoAccess', { intoAccess: true });
          
          // 自动调用 joinTraining 完成报名
          return joinTraining(trainingId).then(
            (joinRes) => {
              this.btnLoading = false;
              if (joinRes.data.code === 200) {
                myMessage.success('加入训练成功！');
                this.hasJoined = true;
                this.trainingPassword = ''; // 清空密码
                // 刷新训练信息以更新进度
                this.$store.dispatch('getTraining');
              } else {
                myMessage.error(joinRes.data.message || '加入训练失败');
              }
            },
            (joinErr) => {
              this.btnLoading = false;
              myMessage.error(joinErr.response?.data?.message || '加入训练失败');
            }
          );
        },
        (err) => {
          this.btnLoading = false;
          // 密码验证失败
          if (err.response?.data?.message) {
            myMessage.error(err.response.data.message);
          } else {
            myMessage.error('密码验证失败');
          }
        }
      );
    },
    goUserHome(username) {
      this.$router.push({
        name: 'UserHome',
        query: { username: username },
      });
    },
    getAcProblemPercent() {
      if (!this.training.problemCount) {
        return 100;
      }
      if (this.training.acCount == null) {
        this.training.acCount = 0;
      }
      return (
        (this.training.acCount / this.training.problemCount) *
        100
      ).toFixed(2);
    },
  },
  computed: {
    ...mapState({
      training: (state) => state.training.training,
    }),
    ...mapGetters([
      'trainingPasswordFormVisible',
      'trainingMenuDisabled',
      'isPrivateTraining',
      'isAuthenticated',
      'isTrainingAdmin',
    ]),
  },
  watch: {
    $route(newVal) {
      this.route_name = newVal.name;
      if(this.groupID){
         if (newVal.name == 'GroupTrainingProblemDetails') {
          this.route_name = 'GroupTrainingProblemList';
        }
      }else{
        if (newVal.name == 'TrainingProblemDetails') {
          this.route_name = 'TrainingProblemList';
        }
      }
      this.changeDomTitle({ title: this.training.title });
    },
  },
  beforeDestroy() {
    this.$store.commit('clearTraining');
  },
};
</script>

<style scoped>
.card-top {
  margin-top: 15px;
}
.training-header {
  text-align: center;
}
.training-progress {
  margin-top: 15px;
  padding: 0 20px;
}
.count {
  margin-top: 10px;
  font-size: 18px;
  font-weight: 700;
}
.join-hint {
  margin-top: 10px;
  font-size: 14px;
  color: #909399;
}
.password-form-card {
  text-align: center;
  margin-bottom: 15px;
}

.info-rows > * {
  margin-bottom: var(--info-row-margin-bottom, 1em);
  display: flex;
  align-items: center;
  font-size: 16px;
  line-height: 1.5;
  color: rgba(0, 0, 0, 0.75);
}
.info-rows > * > *:first-child {
  flex: 1 0 auto;
  text-align: left;
}
.info-rows > :last-child {
  margin-bottom: 0;
}
/deep/ .el-card__header {
  border-bottom: 0px;
  padding-bottom: 0px;
}
/deep/.el-tabs__nav-wrap {
  background: #fff;
  border-radius: 3px;
}
/deep/.el-tabs--top .el-tabs__item.is-top:nth-child(2) {
  padding-left: 20px;
}
</style>
