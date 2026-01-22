<template>
  <div class="problem-list">
    <!-- 筛选和排序工具栏 -->
    <div class="filter-toolbar">
      <el-radio-group v-model="statusFilter" size="small" @change="applyFilter">
        <el-radio-button label="all">全部</el-radio-button>
        <el-radio-button label="completed">已完成</el-radio-button>
        <el-radio-button label="incomplete">未完成</el-radio-button>
      </el-radio-group>

      <div class="sort-buttons">
        <span class="sort-label">排序:</span>
        <el-button-group size="small">
          <el-button
            :type="sortBy === 'problemId' ? 'primary' : ''"
            @click="toggleSort('problemId')"
          >
            题目ID
            <i v-if="sortBy === 'problemId'" :class="getSortIcon()"></i>
          </el-button>
          <el-button
            :type="sortBy === 'difficulty' ? 'primary' : ''"
            @click="toggleSort('difficulty')"
          >
            难度
            <i v-if="sortBy === 'difficulty'" :class="getSortIcon()"></i>
          </el-button>
        </el-button-group>
      </div>
    </div>

    <vxe-table
      border="inner"
      stripe
      auto-resize
      highlight-hover-row
      :data="problemList"
      align="center"
      @cell-click="goTrainingProblem"
    >
      <vxe-table-column
        field="status"
        title=""
        width="50"
        v-if="isAuthenticated"
      >
        <template v-slot="{ row }">
          <template v-if="isGetStatusOk">
            <el-tooltip
              :content="JUDGE_STATUS[row.myStatus]['name']"
              placement="top"
            >
              <template v-if="row.myStatus == 0">
                <i
                  class="el-icon-check"
                  :style="getIconColor(row.myStatus)"
                ></i>
              </template>

              <template v-else-if="row.myStatus != -10">
                <i
                  class="el-icon-minus"
                  :style="getIconColor(row.myStatus)"
                ></i>
              </template>
            </el-tooltip>
          </template>
        </template>
      </vxe-table-column>
      <vxe-table-column
        field="problemId"
        :title="$t('m.Problem_ID')"
        width="150"
        show-overflow
      >
      </vxe-table-column>
      <vxe-table-column
        field="title"
        :title="$t('m.Title')"
        min-width="150"
        show-overflow
      ></vxe-table-column>

      <vxe-table-column
        field="difficulty"
        :title="$t('m.Level')"
        min-width="100"
      >
        <template v-slot="{ row }">
          <span
            class="el-tag el-tag--small"
            :style="getLevelColor(row.difficulty)"
            >{{ getLevelName(row.difficulty) }}</span
          >
        </template>
      </vxe-table-column>

      <vxe-table-column field="tag" min-width="100">
        <template v-slot:header
          ><el-link
            type="primary"
            v-if="!showTags"
            :underline="false"
            @click="showTags = !showTags"
            >{{ $t('m.Show_Tags') }}</el-link
          >
          <el-link
            type="danger"
            v-else
            @click="showTags = !showTags"
            :underline="false"
            >{{ $t('m.Hide_Tags') }}</el-link
          >
        </template>
        <template v-slot="{ row }">
          <div v-if="showTags">
            <span
              class="el-tag el-tag--small"
              :style="
                'margin-right:7px;color:#FFF;background-color:' +
                  (tag.color ? tag.color : '#409eff')
              "
              v-for="tag in row.tags"
              :key="tag.id"
              >{{ tag.name }}</span
            >
          </div>
        </template>
      </vxe-table-column>

      <vxe-table-column field="ac" :title="$t('m.AC_Rate')" min-width="120">
        <template v-slot="{ row }">
          <span>
            <el-tooltip
              effect="dark"
              :content="row.ac + '/' + row.total"
              placement="top"
            >
              <el-progress
                :text-inside="true"
                :stroke-width="20"
                :percentage="getPassingRate(row.ac, row.total)"
              ></el-progress>
            </el-tooltip>
          </span>
        </template>
      </vxe-table-column>
    </vxe-table>
  </div>
</template>

<script>
import { mapState, mapGetters } from 'vuex';
import utils from '@/common/utils';
import { JUDGE_STATUS } from '@/common/constants';
import api from '@/common/api';
export default {
  name: 'TrainingProblemList',
  data() {
    return {
      JUDGE_STATUS: {},
      isGetStatusOk: false,
      testcolor: 'rgba(0, 206, 209, 1)',
      showTags: false,
      groupID:null,
      // 筛选和排序
      statusFilter: 'all', // all, completed, incomplete
      sortBy: 'problemId', // problemId, difficulty
      sortOrder: 'asc', // asc, desc
    };
  },
  created(){
    let gid = this.$route.params.groupID;
    if(gid){
      this.groupID = gid;
    }
  },
  mounted() {
    this.JUDGE_STATUS = Object.assign({}, JUDGE_STATUS);
    this.getTrainingProblemList();
  },
  methods: {
    getTrainingProblemList() {
      this.$store.dispatch('getTrainingProblemList').then((res) => {
        if (this.isAuthenticated) {
          // 如果已登录，则需要查询对当前页面题目列表中各个题目的提交情况
          let pidList = [];
          // 使用 originalProblemList 而不是 problemList
          if (this.originalProblemList && this.originalProblemList.length > 0) {
            for (let index = 0; index < this.originalProblemList.length; index++) {
              pidList.push(this.originalProblemList[index].pid);
            }
            this.isGetStatusOk = false;
            api.getUserProblemStatus(pidList, false,null,this.groupID).then((res) => {
              let result = res.data.data;
              for (let index = 0; index < this.originalProblemList.length; index++) {
                this.originalProblemList[index]['myStatus'] =
                  result[this.originalProblemList[index].pid]['status'];
              }
              this.isGetStatusOk = true;
            });
          }
        }
      });
    },
    goTrainingProblem(event) {
      if(this.groupID){
        this.$router.push({
          name: 'GroupTrainingProblemDetails',
          params: {
            trainingID: this.$route.params.trainingID,
            problemID: event.row.problemId,
            groupID: this.groupID
          },
        });
      }else{
        this.$router.push({
          name: 'TrainingProblemDetails',
          params: {
            trainingID: this.$route.params.trainingID,
            problemID: event.row.problemId,
          },
        });
      }
    },
    getACRate(ACCount, TotalCount) {
      return utils.getACRate(ACCount, TotalCount);
    },
    getIconColor(status) {
      return (
        'font-weight: 600;font-size: 16px;color:' + JUDGE_STATUS[status].rgb
      );
    },
    getLevelColor(difficulty) {
      return utils.getLevelColor(difficulty);
    },
    getLevelName(difficulty) {
      return utils.getLevelName(difficulty);
    },
    getPassingRate(ac, total) {
      if (!total) {
        return 0;
      }
      return ((ac / total) * 100).toFixed(2);
    },
    // 切换排序字段
    toggleSort(field) {
      if (this.sortBy === field) {
        // 如果点击的是当前排序字段，切换排序顺序
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
      } else {
        // 如果点击的是新字段，设置为该字段并默认升序
        this.sortBy = field;
        this.sortOrder = 'asc';
      }
    },
    // 获取排序图标
    getSortIcon() {
      return this.sortOrder === 'asc' ? 'el-icon-top' : 'el-icon-bottom';
    },
    // 应用筛选（其实筛选是通过computed自动响应的，这个方法可以保留用于未来扩展）
    applyFilter() {
      // 筛选逻辑在computed problemList中自动处理
    },
  },
  computed: {
    ...mapState({
      originalProblemList: (state) => state.training.trainingProblemList
    }),
    ...mapGetters(['isAuthenticated']),
    // 计算属性：应用筛选和排序
    problemList() {
      // 确保有数据
      if (!this.originalProblemList || !Array.isArray(this.originalProblemList)) {
        return [];
      }

      let filtered = [...this.originalProblemList];

      // 1. 状态筛选
      if (this.statusFilter === 'completed') {
        // 已完成：myStatus === 0 (AC)
        filtered = filtered.filter(p => p.myStatus === 0);
      } else if (this.statusFilter === 'incomplete') {
        // 未完成：myStatus !== 0 (包括未提交的 -10 和提交但没AC的其他状态)
        filtered = filtered.filter(p => p.myStatus !== 0);
      }

      // 2. 排序
      filtered.sort((a, b) => {
        let aVal, bVal;

        if (this.sortBy === 'problemId') {
          // 按题目ID字典序排序
          aVal = a.problemId || '';
          bVal = b.problemId || '';
        } else if (this.sortBy === 'difficulty') {
          // 按难度排序
          aVal = a.difficulty || 0;
          bVal = b.difficulty || 0;
        }

        if (this.sortOrder === 'asc') {
          return aVal > bVal ? 1 : -1;
        } else {
          return aVal < bVal ? 1 : -1;
        }
      });

      return filtered;
    }
  },
};
</script>

<style scoped>
.filter-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 16px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.sort-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sort-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

@media screen and (min-width: 1050px) {
  /deep/ .vxe-table--body-wrapper {
    overflow-x: hidden !important;
  }
}
</style>
