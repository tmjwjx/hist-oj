<template>
  <div class="home-container">
    <el-row :gutter="20">
      <!-- 左侧：公告 - 扩大宽度 -->
      <el-col :md="16" :sm="24">
        <Announcements default-expanded></Announcements>
      </el-col>

      <!-- 右侧：近期比赛、时间 + Rating排行榜 -->
      <el-col :md="8" :sm="24" class="phone-margin">
        <!-- 将完整的近期比赛卡片放在时间组件上方 -->
        <template v-if="contests.length">
          <el-card class="recent-contests-sidebar">
            <div slot="header" class="clearfix title content-center">
              <div class="home-title home-contest">
                <i class="el-icon-trophy"></i> {{ $t('m.Recent_Contest') }}
              </div>
            </div>
            <el-row :gutter="12">
              <el-col
                :span="24"
                v-for="(contest, index) in contests"
                :key="index"
                class="contest-col"
              >
                <el-card
                  shadow="hover"
                  class="contest-card"
                  :class="
                    contest.status == 0
                      ? 'contest-card-running'
                      : 'contest-card-schedule'
                  "
                >
                  <div slot="header" class="clearfix contest-header">
                    <a class="contest-title" @click="goContest(contest.id)">{{
                      contest.title
                    }}</a>
                    <div class="contest-status">
                      <el-tag
                        effect="dark"
                        size="medium"
                        :color="CONTEST_STATUS_REVERSE[contest.status]['color']"
                      >
                        <i class="fa fa-circle" aria-hidden="true"></i>
                        {{
                          $t('m.' + CONTEST_STATUS_REVERSE[contest.status]['name'])
                        }}
                      </el-tag>
                    </div>
                  </div>
                  <div class="contest-type-auth">
                    <template v-if="contest.type == 0">
                      <el-button
                        :type="'primary'"
                        round
                        @click="goContestList(contest.type)"
                        size="mini"
                      >
                        <i class="fa fa-trophy"></i>
                        {{ contest.type | parseContestType }}
                      </el-button>
                    </template>
                    <template v-else>
                      <el-tooltip
                        :content="
                          $t('m.Contest_Rank') +
                            '：' +
                            (contest.oiRankScoreType == 'Recent'
                              ? $t(
                                  'm.Based_on_The_Recent_Score_Submitted_Of_Each_Problem'
                                )
                              : $t(
                                  'm.Based_on_The_Highest_Score_Submitted_For_Each_Problem'
                                ))
                        "
                        placement="top"
                      >
                        <el-button
                          :type="'warning'"
                          round
                          @click="goContestList(contest.type)"
                          size="mini"
                        >
                          <i class="fa fa-trophy"></i>
                          {{ contest.type | parseContestType }}
                        </el-button>
                      </el-tooltip>
                    </template>
                    <el-tooltip
                      v-if="isRatingContest(contest.id)"
                      :content="$t('m.Rating_Contest')"
                      placement="top"
                      effect="dark"
                    >
                      <el-tag type="danger" effect="plain" size="medium">
                        <i class="fa fa-star"></i> Rating
                      </el-tag>
                    </el-tooltip>
                    <el-tooltip
                      :content="$t('m.' + CONTEST_TYPE_REVERSE[contest.auth].tips)"
                      placement="top"
                      effect="light"
                    >
                      <el-tag
                        :type="CONTEST_TYPE_REVERSE[contest.auth]['color']"
                        size="medium"
                        effect="plain"
                      >
                        {{ $t('m.' + CONTEST_TYPE_REVERSE[contest.auth]['name']) }}
                      </el-tag>
                    </el-tooltip>
                  </div>
                  <ul class="contest-info">
                    <li>
                      <el-button type="primary" round size="mini">
                        <i class="fa fa-calendar"></i>
                        {{ contest.startTime | localtime((format = 'MM-DD HH:mm')) }}
                      </el-button>
                    </li>
                    <li>
                      <el-button type="success" round size="mini">
                        <i class="fa fa-clock-o"></i>
                        {{ getDuration(contest.startTime, contest.endTime) }}
                      </el-button>
                    </li>
                    <li>
                      <el-button size="mini" round plain v-if="contest.count != null">
                        <i class="el-icon-user-solid" style="color:rgb(48, 145, 242);"></i>
                        {{ $t('m.Registered_Count') }}: {{ contest.count }}
                      </el-button>
                    </li>
                    <li>
                      <el-button size="mini" round plain>
                        <i class="el-icon-user"></i>
                        {{ $t('m.Problem_Setter') }}: {{ contest.author }}
                      </el-button>
                    </li>
                  </ul>
                </el-card>
              </el-col>
            </el-row>
          </el-card>
        </template>

        <!-- 时间显示组件 -->
        <TimeDisplay></TimeDisplay>

        <!-- Rating 排行榜 -->
        <el-card class="card-top">
          <div slot="header" class="clearfix">
            <span class="panel-title home-title">
              <i class="el-icon-trophy"></i> {{ $t('m.Rating_Rank') }}
            </span>
            <el-button
              type="text"
              size="small"
              @click="goRatingRank"
              style="float: right; padding: 3px 0; color: #409eff;"
            >
              {{ $t('m.View_All') }} <i class="el-icon-d-arrow-right"></i>
            </el-button>
          </div>
          <vxe-table
            border="inner"
            stripe
            auto-resize
            align="center"
            :data="ratingRankList"
            max-height="500px"
            :loading="loading.ratingRankLoading"
          >
            <vxe-table-column type="seq" width="50">
              <template v-slot="{ rowIndex }">
                <span :class="getRankTagClass(rowIndex)">{{ rowIndex + 1 }}</span>
                <span :class="'cite no' + rowIndex"></span>
              </template>
            </vxe-table-column>
            <vxe-table-column
              field="username"
              :title="$t('m.Username')"
              min-width="120"
              align="left"
            >
              <template v-slot="{ row }">
                <avatar
                  :username="row.username"
                  :inline="true"
                  :size="25"
                  color="#FFF"
                  :src="row.avatar"
                  class="user-avatar"
                ></avatar>
                <a
                  @click="goUserHome(row.username, row.uid)"
                  :style="{ color: row.color, fontWeight: 'bold' }"
                  class="username-link"
                >{{ row.username }}</a>
                <span style="margin-left:2px" v-if="row.level">
                  <el-tag effect="dark" size="small" :color="row.color">
                    {{ row.level }}
                  </el-tag>
                </span>
              </template>
            </vxe-table-column>
            <vxe-table-column field="rating" title="Rating" width="70" align="center">
              <template v-slot="{ row }">
                <span :style="{ color: row.color, fontWeight: 'bold' }">{{ row.rating }}</span>
              </template>
            </vxe-table-column>
          </vxe-table>
        </el-card>
      </el-col>
    </el-row>

  </div>
</template>

<script>
import time from "@/common/time";
import api from "@/common/api";
import ratingApi from "@/common/rating-api";
import {
  CONTEST_STATUS_REVERSE,
  CONTEST_TYPE_REVERSE,
} from "@/common/constants";
import myMessage from "@/common/message";
import { mapState, mapGetters } from "vuex";
import Avatar from "vue-avatar";
const Announcements = () => import("@/components/oj/common/Announcements.vue");
const TimeDisplay = () =>
  import("@/components/oj/common/TimeDisplay.vue");
export default {
  name: "home",
  components: {
    Announcements,
    TimeDisplay,
    Avatar,
  },
  data() {
    return {
      interval: 5000,
      ratingRankList: [],
      CONTEST_STATUS_REVERSE: {},
      CONTEST_TYPE_REVERSE: {},
      contests: [],
      ratingContests: new Set(), // 存储 Rating 比赛的 ID
      loading: {
        ratingRankLoading: false,
        recentContests: false,
      },
      carouselImgList: [],
      srcHight: "440px",
    };
  },
  mounted() {
    let screenWidth = window.screen.width;
    if (screenWidth < 768) {
      this.srcHight = "200px";
    } else {
      this.srcHight = "360px";
    }
    this.CONTEST_STATUS_REVERSE = Object.assign({}, CONTEST_STATUS_REVERSE);
    this.CONTEST_TYPE_REVERSE = Object.assign({}, CONTEST_TYPE_REVERSE);
    this.getHomeCarousel();
    this.getRecentContests();
    // 使用 nextTick 确保 DOM 渲染后再异步加载 Rating 数据
    this.$nextTick(() => {
      this.getRatingRank();
    });
  },
  methods: {
    getHomeCarousel() {
      api.getHomeCarousel().then((res) => {
        if (res.data.data != null && res.data.data.length > 0) {
          this.carouselImgList = res.data.data;
        }
      });
    },

    getRecentContests() {
      this.loading.recentContests = true;
      api.getRecentContests().then(
        (res) => {
          this.contests = Array.isArray(res.data.data) ? res.data.data : [];
          this.loading.recentContests = false;

          // 异步获取 Rating 比赛信息
          this.fetchRatingContests();
        },
        (err) => {
          this.loading.recentContests = false;
        }
      );
    },

    async fetchRatingContests() {
      if (!this.contests || this.contests.length === 0) return;

      try {
        // 提取所有比赛 ID
        const contestIds = this.contests.map(c => c.id);

        // 检查本地缓存
        const cachedData = this.getRatingContestsFromCache(contestIds);
        if (cachedData.allCached) {
          // 所有数据都在缓存中，直接使用
          this.ratingContests = cachedData.ratingContests;
          this.$forceUpdate();
          return;
        }

        // 使用批量查询接口（一次请求获取所有比赛的 Rating 状态）
        const results = await ratingApi.getBatchContestInfo(contestIds);

        // 更新 ratingContests Set 和缓存
        this.ratingContests.clear();
        Object.entries(results).forEach(([contestId, info]) => {
          const id = parseInt(contestId);
          if (info.isRating) {
            this.ratingContests.add(id);
          }
          // 缓存到 localStorage（有效期 1 小时）
          this.cacheRatingContest(id, info.isRating);
        });

        // 强制更新视图
        this.$forceUpdate();
      } catch (error) {
        console.error('获取 Rating 比赛信息失败:', error);
        // 失败时尝试使用缓存数据
        const cachedData = this.getRatingContestsFromCache(this.contests.map(c => c.id));
        if (cachedData.ratingContests.size > 0) {
          this.ratingContests = cachedData.ratingContests;
          this.$forceUpdate();
        }
      }
    },

    // 从缓存中获取 Rating 比赛信息
    getRatingContestsFromCache(contestIds) {
      const ratingContests = new Set();
      let cachedCount = 0;

      contestIds.forEach(id => {
        const cacheKey = `rating_contest_${id}`;
        const cached = localStorage.getItem(cacheKey);

        if (cached) {
          try {
            const data = JSON.parse(cached);
            // 检查缓存是否过期（1 小时）
            if (Date.now() - data.timestamp < 3600000) {
              cachedCount++;
              if (data.isRating) {
                ratingContests.add(id);
              }
            } else {
              // 缓存过期，删除
              localStorage.removeItem(cacheKey);
            }
          } catch (e) {
            localStorage.removeItem(cacheKey);
          }
        }
      });

      return {
        ratingContests,
        allCached: cachedCount === contestIds.length
      };
    },

    // 缓存 Rating 比赛信息
    cacheRatingContest(contestId, isRating) {
      const cacheKey = `rating_contest_${contestId}`;
      const data = {
        isRating,
        timestamp: Date.now()
      };
      try {
        localStorage.setItem(cacheKey, JSON.stringify(data));
      } catch (e) {
        // localStorage 可能已满，忽略错误
        console.warn('缓存 Rating 信息失败:', e);
      }
    },

    isRatingContest(contestId) {
      return this.ratingContests.has(contestId);
    },
    getRatingRank() {
      this.loading.ratingRankLoading = true;
      ratingApi.getRatingRank(1, 10).then(
        (res) => {
          this.ratingRankList = res.records || [];
          this.loading.ratingRankLoading = false;
        },
        (err) => {
          console.error('获取 Rating 排名失败:', err);
          // 静默处理错误，不显示提示
          this.ratingRankList = [];
          this.loading.ratingRankLoading = false;
        }
      );
    },
    goContest(cid) {
      if (!this.isAuthenticated) {
        myMessage.warning(this.$i18n.t("m.Please_login_first"));
        this.$store.dispatch("changeModalStatus", { visible: true });
      } else {
        this.$router.push({
          name: "ContestDetails",
          params: { contestID: cid },
        });
      }
    },
    goContestList(type) {
      this.$router.push({
        name: "ContestList",
        query: {
          type,
        },
      });
    },
    goUserHome(username, uid) {
      this.$router.push({
        path: "/user-home",
        query: { uid, username },
      });
    },
    goRatingRank() {
      this.$router.push({
        path: "/rating-rank"
      });
    },
    getDuration(startTime, endTime) {
      return time.formatSpecificDuration(startTime, endTime);
    },
    getRankTagClass(rowIndex) {
      return "rank-tag no" + (rowIndex + 1);
    },
  },
  computed: {
    ...mapState(["websiteConfig"]),
    ...mapGetters(["isAuthenticated"]),
  },
};
</script>
<style>
.contest-card-running {
  border-left: 3px solid #19be6b;
}
.contest-card-schedule {
  border-left: 3px solid #ff9900;
}
</style>
<style scoped>
/* 整体布局 */
.home-container {
  width: 100%;
  max-width: none;
  margin: 0;
  padding: 20px;
  box-sizing: border-box;
  overflow-x: hidden;
  background: #f5f7fa;
}

.phone-margin,
.recent-contests-sidebar {
  min-width: 0;
}

/* 卡片通用样式 */
/deep/.el-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  background: #fff;
  margin-bottom: 20px;
}

/deep/.el-card__header {
  padding: 12px 16px !important;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.card-top {
  margin-top: 0;
}

.recent-contests-sidebar {
  margin-bottom: 20px;
}

/* 标题样式 */
.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.home-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.home-contest {
  text-align: left;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.5;
}

.welcome-title {
  font-weight: 600;
  font-size: 18px;
}

/* 轮播图样式 */
.img-carousel {
  height: 360px;
  border-radius: 4px;
  overflow: hidden;
}

.img-carousel /deep/ .el-carousel__indicator {
  background-color: rgba(255, 255, 255, 0.5);
}

.img-carousel /deep/ .el-carousel__indicator.is-active {
  background-color: #409eff;
}

/* 比赛卡片样式 */
.contest-card {
  margin-bottom: 16px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  transition: all 0.2s;
}

.contest-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.contest-card /deep/ .el-card__header {
  padding: 12px 16px !important;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.contest-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.contest-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  cursor: pointer;
  transition: color 0.2s;
}

.contest-title:hover {
  color: #409eff;
}

.contest-card-running .contest-title {
  color: #19be6b;
}

.contest-card-schedule .contest-title {
  color: #ff9900;
}

.contest-status {
  float: none;
}

.contest-type-auth {
  text-align: left;
  margin: 12px 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.contest-info {
  text-align: left;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.contest-info li {
  display: inline-block;
  padding-right: 0;
}

/* Rating 排行榜样式 */
.user-avatar {
  margin-right: 8px !important;
  vertical-align: middle;
}

span.rank-tag {
  display: inline-block;
  min-width: 24px;
  height: 24px;
  line-height: 24px;
  padding: 0 6px;
  text-align: center;
  color: #fff;
  font-weight: 600;
  font-size: 13px;
  background: #909399;
  border-radius: 2px;
}

span.rank-tag.no1 {
  background: #FFD700;
  color: #8b6914;
}

span.rank-tag.no2 {
  background: #C0C0C0;
  color: #4a4a4a;
}

span.rank-tag.no3 {
  background: #CD7F32;
  color: #5c3a1a;
}

.cite {
  display: none;
}

/* 响应式布局 */
@media screen and (max-width: 768px) {
  .home-container {
    padding: 10px;
  }

  .img-carousel {
    height: 200px;
  }

  .phone-margin {
    margin-top: 0;
  }

  .contest-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .contest-status {
    margin-top: 8px;
  }

  .contest-info {
    flex-direction: column;
  }

  .welcome-title {
    font-size: 16px;
  }
}

/* 表格优化 */
/deep/ .vxe-table {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

/deep/ .vxe-table--header-wrapper {
  background: #f5f7fa;
}

/deep/ .vxe-table .vxe-header--column {
  background: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

/deep/ .vxe-table .vxe-body--row:hover {
  background: #f5f7fa;
}

/deep/ .vxe-table--border-line {
  border-color: #e4e7ed;
}

@media screen and (min-width: 1050px) {
  /deep/ .vxe-table--body-wrapper {
    overflow-x: hidden !important;
  }
}

/deep/.el-image {
  height: 100%;
  width: 100%;
}

/* 清除旧样式 */
.content-center {
  text-align: center;
}

.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}

.clearfix:after {
  clear: both;
}

ul,
li {
  padding: 0;
  margin: 0;
  list-style: none;
}

/* 标签优化 */
/deep/ .el-tag {
  border-radius: 2px;
  font-size: 12px;
  height: 24px;
  line-height: 22px;
  padding: 0 8px;
}

/deep/ .el-button--mini {
  font-size: 12px;
  padding: 5px 10px;
}

/deep/ .el-button--small {
  font-size: 13px;
  padding: 7px 15px;
}
</style>
