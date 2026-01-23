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
      ref="problemTable"
      border="inner"
      stripe
      auto-resize
      highlight-hover-row
      :data="problemList"
      align="center"
      :row-class-name="getRowClassName"
      :max-height="700"
      show-overflow
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

    <!-- 浮动按钮组 -->
    <div class="float-buttons">
      <el-tooltip content="回到顶部" placement="left">
        <el-button
          type="primary"
          circle
          icon="el-icon-top"
          class="float-button"
          @click="scrollToTop"
        ></el-button>
      </el-tooltip>
      <el-tooltip content="上次作答题目" placement="left">
        <el-button
          type="success"
          circle
          icon="el-icon-bottom"
          class="float-button"
          :disabled="!lastClickedProblemId"
          @click="scrollToLastProblem"
        ></el-button>
      </el-tooltip>
    </div>
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
      // 上次点击的题目ID
      lastClickedProblemId: null,
      // 滚动位置存储key
      scrollPositionKey: 'training_problem_list_scroll',
      // 上次点击题目ID的存储key
      lastProblemKey: 'training_last_clicked_problem',
    };
  },
  created(){
    let gid = this.$route.params.groupID;
    if(gid){
      this.groupID = gid;
    }
    // 加载上次点击的题目ID
    this.loadLastClickedProblem();
  },
  mounted() {
    this.JUDGE_STATUS = Object.assign({}, JUDGE_STATUS);
    this.getTrainingProblemList();
  },
  beforeDestroy() {
    // 移除滚动监听（使用 try-catch 避免销毁时的潜在错误）
    try {
      this.removeScrollListener();
    } catch (error) {
      // 忽略销毁时的错误
    }
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
              // 数据加载完成后恢复滚动位置和添加监听器
              this.$nextTick(() => {
                this.restoreScrollPosition();
                this.addScrollListener();
              });
            });
          } else {
            // 没有数据时也要添加监听器
            this.$nextTick(() => {
              this.addScrollListener();
            });
          }
        } else {
          // 未登录时直接添加监听器
          this.$nextTick(() => {
            this.restoreScrollPosition();
            this.addScrollListener();
          });
        }
      });
    },
    goTrainingProblem(event) {
      // 保存当前点击的题目ID
      this.saveLastClickedProblem(event.row.problemId);

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
    // 获取行的自定义类名（用于高亮上次点击的题目）
    getRowClassName({ row }) {
      if (row.problemId === this.lastClickedProblemId) {
        return 'last-clicked-row';
      }
      return '';
    },
    // 保存上次点击的题目ID到 localStorage
    saveLastClickedProblem(problemId) {
      const key = `${this.lastProblemKey}_${this.$route.params.trainingID}`;
      localStorage.setItem(key, problemId);
      this.lastClickedProblemId = problemId;
    },
    // 从 localStorage 加载上次点击的题目ID
    loadLastClickedProblem() {
      const key = `${this.lastProblemKey}_${this.$route.params.trainingID}`;
      const problemId = localStorage.getItem(key);
      if (problemId) {
        this.lastClickedProblemId = problemId;
      }
    },
    // 获取表格滚动容器
    getTableWrapper() {
      if (!this.$refs.problemTable || !this.$refs.problemTable.$el) return null;

      // 尝试多种方式获取滚动容器
      let wrapper = this.$refs.problemTable.$el?.querySelector('.vxe-table--body-wrapper');

      // 如果找不到，尝试其他可能的选择器
      if (!wrapper) {
        wrapper = this.$refs.problemTable.$el?.querySelector('.vxe-table .body--wrapper');
      }

      // 如果还是找不到，直接查找 body-wrapper
      if (!wrapper) {
        wrapper = this.$refs.problemTable.$el?.querySelector('.body--wrapper');
      }

      return wrapper;
    },
    // 添加滚动监听器
    addScrollListener() {
      // 使用多次尝试确保 DOM 已渲染
      const tryAddListener = (attempt = 0) => {
        if (attempt > 5) return; // 最多尝试5次

        this.$nextTick(() => {
          const tableWrapper = this.getTableWrapper();
          if (tableWrapper) {
            // 先移除可能存在的监听器，避免重复添加
            tableWrapper.removeEventListener('scroll', this.handleScroll);
            tableWrapper.addEventListener('scroll', this.handleScroll);
          } else {
            // 如果还没找到，延迟后重试
            setTimeout(() => tryAddListener(attempt + 1), 100);
          }
        });
      };

      tryAddListener();
    },
    // 移除滚动监听器
    removeScrollListener() {
      const tableWrapper = this.getTableWrapper();
      if (tableWrapper) {
        tableWrapper.removeEventListener('scroll', this.handleScroll);
      }
    },
    // 处理滚动事件
    handleScroll(event) {
      const scrollTop = event.target.scrollTop;
      const key = `${this.scrollPositionKey}_${this.$route.params.trainingID}`;
      sessionStorage.setItem(key, scrollTop.toString());
    },
    // 恢复滚动位置
    restoreScrollPosition() {
      const key = `${this.scrollPositionKey}_${this.$route.params.trainingID}`;
      const scrollTop = sessionStorage.getItem(key);
      if (scrollTop) {
        const tableWrapper = this.getTableWrapper();
        if (tableWrapper) {
          tableWrapper.scrollTop = parseInt(scrollTop);
        }
      }
    },
    // 滚动到顶部
    scrollToTop() {
      const tableWrapper = this.getTableWrapper();
      if (tableWrapper) {
        tableWrapper.scrollTo({
          top: 0,
          behavior: 'smooth'
        });
      }
    },
    // 滚动到上次点击的题目
    scrollToLastProblem() {
      if (!this.lastClickedProblemId) return;

      // 在当前显示的列表中查找目标题目
      const targetRow = this.problemList.find(p => p.problemId === this.lastClickedProblemId);

      if (!targetRow) {
        // 检查是否因为筛选导致题目不可见
        const existsInOriginal = this.originalProblemList &&
          this.originalProblemList.some(p => p.problemId === this.lastClickedProblemId);

        if (existsInOriginal) {
          // 题目存在但被筛选过滤了
          const filterText = this.statusFilter === 'completed' ? '已完成' :
                           this.statusFilter === 'incomplete' ? '未完成' : '当前筛选';
          this.$message.warning(`上次作答的题目不在"${filterText}"列表中，请切换筛选条件查看`);
        } else {
          this.$message.warning('未找到上次作答的题目');
        }
        return;
      }

      // 使用 vxe-table 的 scrollTo 方法滚动到指定行
      this.$nextTick(() => {
        if (this.$refs.problemTable) {
          try {
            // vxe-table v3/v4 的 scrollToRow 方法
            this.$refs.problemTable.scrollToRow(targetRow);
          } catch (e) {
            // 如果失败，尝试使用索引方式
            const targetIndex = this.problemList.indexOf(targetRow);
            if (targetIndex !== -1) {
              this.$refs.problemTable.scrollTo(targetIndex);
            }
          }
        }
      });
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

/* 高亮上次点击的题目行 */
/deep/ .vxe-table .last-clicked-row {
  background-color: #e6f7ff !important;
  animation: highlight-pulse 2s ease-in-out;
}

@keyframes highlight-pulse {
  0% {
    background-color: #91d5ff;
  }
  100% {
    background-color: #e6f7ff;
  }
}

/* 浮动按钮组 */
.float-buttons {
  position: fixed;
  right: 60px;
  bottom: 100px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 9999;
  align-items: center;
}

.float-button {
  width: 45px;
  height: 45px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  padding: 0;
}

.float-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
}

.float-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

/* 响应式调整 */
@media screen and (max-width: 768px) {
  .float-buttons {
    right: 20px;
    bottom: 70px;
  }

  .float-button {
    width: 40px;
    height: 40px;
  }
}

@media screen and (min-width: 1050px) {
  /deep/ .vxe-table--body-wrapper {
    overflow-x: hidden !important;
  }
}
</style>
