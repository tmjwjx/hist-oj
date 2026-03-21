<template>
  <div>
    <el-row :gutter="20">
      <el-col
        :md="15"
        :sm="24"
      >
        <el-card>
          <div
            slot="header"
            class="content-center"
          >
            <span class="panel-title home-title welcome-title">{{ $t('m.Welcome_to')
              }}{{ websiteConfig.shortName }}</span>
          </div>
          <el-carousel
            :interval="interval"
            :height="srcHight"
            class="img-carousel"
            arrow="always"
            indicator-position="outside"
          >
            <el-carousel-item
              v-for="(item, index) in carouselImgList"
              :key="index"
            >
              <el-image
                :src="item.url"
                fit="fill"
              >
                <div
                  slot="error"
                  class="image-slot"
                >
                  <i class="el-icon-picture-outline"></i>
                </div>
              </el-image>
            </el-carousel-item>
          </el-carousel>
        </el-card>
        <Announcements class="card-top"></Announcements>
        <SubmissionStatistic class="card-top"></SubmissionStatistic>
      </el-col>
      <el-col
        :md="9"
        :sm="24"
        class="phone-margin"
      >
        <template v-if="contests.length">
          <el-card>
            <div
              slot="header"
              class="clearfix title content-center"
            >
              <div class="home-title home-contest">
                <i class="el-icon-trophy"></i> {{ $t('m.Recent_Contest') }}
              </div>
            </div>
            <el-card
              shadow="hover"
              v-for="(contest, index) in contests"
              :key="index"
              class="contest-card"
              :class="
                contest.status == 0
                  ? 'contest-card-running'
                  : 'contest-card-schedule'
              "
            >
              <div
                slot="header"
                class="clearfix contest-header"
              >
                <a
                  class="contest-title"
                  @click="goContest(contest.id)"
                >{{
                  contest.title
                }}</a>
                <div class="contest-status">
                  <el-tag
                    effect="dark"
                    size="medium"
                    :color="CONTEST_STATUS_REVERSE[contest.status]['color']"
                  >
                    <i
                      class="fa fa-circle"
                      aria-hidden="true"
                    ></i>
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
                    style="margin-right: 10px;"
                  ><i class="fa fa-trophy"></i>
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
                      style="margin-right: 10px;"
                    ><i class="fa fa-trophy"></i>
                      {{ contest.type | parseContestType }}
                    </el-button>
                  </el-tooltip>
                </template>
                <el-tooltip
                  v-if="isRatingContest(contest.id)"
                  content="Rating 比赛"
                  placement="top"
                  effect="dark"
                >
                  <el-tag
                    type="danger"
                    effect="plain"
                    size="medium"
                    style="margin-right: 10px;"
                  >
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
                  <el-button
                    type="primary"
                    round
                    size="mini"
                    style="margin-top: 4px;"
                  ><i class="fa fa-calendar"></i>
                    {{
                      contest.startTime | localtime((format = 'MM-DD HH:mm'))
                    }}
                  </el-button>
                </li>
                <li>
                  <el-button
                    type="success"
                    round
                    size="mini"
                    style="margin-top: 4px;"
                  ><i class="fa fa-clock-o"></i>
                    {{ getDuration(contest.startTime, contest.endTime) }}
                  </el-button>
                </li>
                <li>
                  <el-button
                    size="mini"
                    round
                    plain
                    v-if="contest.count != null"
                  >
                    <i
                      class="el-icon-user-solid"
                      style="color:rgb(48, 145, 242);"
                    ></i>x{{ contest.count }}
                  </el-button>
                </li>
              </ul>
            </el-card>
          </el-card>
        </template>
        <!-- 时间显示组件 -->
        <TimeDisplay :class="contests.length ? 'card-top' : ''"></TimeDisplay>
        <el-card :class="contests.length ? 'card-top' : ''">
          <div
            slot="header"
            class="clearfix"
          >
            <span class="panel-title home-title">
              <i class="el-icon-trophy"></i> Rating 排行榜
            </span>
            <el-button
              type="text"
              size="small"
              @click="goRatingRank"
              style="float: right; padding: 3px 0; color: #409eff;"
            >
              查看全部 <i class="el-icon-d-arrow-right"></i>
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
            <vxe-table-column
              type="seq"
              min-width="50"
            >
              <template v-slot="{ rowIndex }">
                <span :class="getRankTagClass(rowIndex)">{{ rowIndex + 1 }}
                </span>
                <span :class="'cite no' + rowIndex"></span>
              </template>
            </vxe-table-column>
            <vxe-table-column
              field="username"
              :title="$t('m.Username')"
              min-width="200"
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
                >{{ row.username }}</a>
                <span
                  style="margin-left:2px"
                  v-if="row.level"
                >
                  <el-tag
                    effect="dark"
                    size="small"
                    :color="row.color"
                  >
                    {{ row.level }}
                  </el-tag>
                </span>
              </template>
            </vxe-table-column>
            <vxe-table-column
              field="rating"
              title="Rating"
              min-width="80"
              align="left"
            >
              <template v-slot="{ row }">
                <span :style="{ color: row.color, fontWeight: 'bold' }">{{ row.rating }}</span>
              </template>
            </vxe-table-column>
          </vxe-table>
        </el-card>

        <el-card class="card-top">
          <div
            slot="header"
            class="clearfix"
          >
            <span class="panel-title home-title">
              <i class="el-icon-magic-stick"></i> {{
              $t('m.Latest_Problem')
            }}</span>
          </div>
          <vxe-table
            border="inner"
            highlight-hover-row
            stripe
            :loading="loading.recentUpdatedProblemsLoading"
            auto-resize
            :data="recentUpdatedProblems"
            @cell-click="goProblem"
          >
            <vxe-table-column
              field="problemId"
              :title="$t('m.Problem_ID')"
              min-width="100"
              show-overflow
              align="center"
            >
            </vxe-table-column>
            <vxe-table-column
              field="title"
              :title="$t('m.Title')"
              show-overflow
              min-width="130"
              align="center"
            >
            </vxe-table-column>
            <vxe-table-column
              field="gmtModified"
              :title="$t('m.Recent_Update')"
              show-overflow
              min-width="96"
              align="center"
            >
              <template v-slot="{ row }">
                <el-tooltip
                  :content="row.gmtModified | localtime"
                  placement="top"
                >
                  <span>{{ row.gmtModified | fromNow }}</span>
                </el-tooltip>
              </template>
            </vxe-table-column>

          </vxe-table>
        </el-card>
        <el-card class="card-top">
          <div
            slot="header"
            class="clearfix title"
          >
            <span class="home-title panel-title">
              <i class="el-icon-monitor"></i> {{ $t('m.Supported_Remote_Online_Judge') }}
            </span>
          </div>
          <el-row :gutter="20">
            <el-col
              :md="8"
              :sm="24"
              v-for="(oj, index) in remoteJudgeList"
              :key="index"
            >
              <a
                :href="oj.url"
                target="_blank"
              >
                <el-tooltip
                  :content="oj.name"
                  placement="top"
                >
                  <el-image
                    :src="oj.logo"
                    fit="fill"
                    class="oj-logo"
                    :class="
                      oj.status ? 'oj-normal ' + oj.name : 'oj-error ' + oj.name
                    "
                  >
                    <div
                      slot="error"
                      class="image-slot"
                    >
                      <i class="el-icon-picture-outline"></i>
                    </div>
                  </el-image>
                </el-tooltip>
              </a>
            </el-col>
          </el-row>
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
const SubmissionStatistic = () =>
  import("@/components/oj/home/SubmissionStatistic.vue");
const TimeDisplay = () =>
  import("@/components/oj/common/TimeDisplay.vue");
export default {
  name: "home",
  components: {
    Announcements,
    SubmissionStatistic,
    TimeDisplay,
    Avatar,
  },
  data() {
    return {
      interval: 5000,
      recentUpdatedProblems: [],
      recentUserACRecord: [],
      ratingRankList: [],
      CONTEST_STATUS_REVERSE: {},
      CONTEST_TYPE_REVERSE: {},
      contests: [],
      ratingContests: new Set(), // 存储 Rating 比赛的 ID
      loading: {
        recent7ACRankLoading: false,
        ratingRankLoading: false,
        recentUpdatedProblemsLoading: false,
        recentContests: false,
      },
      carouselImgList: [],
      srcHight: "440px",
      remoteJudgeList: [
        {
          url: "http://acm.hdu.edu.cn",
          name: "HDU",
          logo: require("@/assets/hdu-logo.png"),
          status: true,
        },
        {
          url: "http://poj.org",
          name: "POJ",
          logo: require("@/assets/poj-logo.png"),
          status: true,
        },
        {
          url: "https://codeforces.com",
          name: "Codeforces",
          logo: require("@/assets/codeforces-logo.png"),
          status: true,
        },
        {
          url: "https://codeforces.com/gyms",
          name: "GYM",
          logo: require("@/assets/gym-logo.png"),
          status: true,
        },
        {
          url: "https://atcoder.jp",
          name: "AtCoder",
          logo: require("@/assets/atcoder-logo.png"),
          status: true,
        },
        {
          url: "https://www.spoj.com",
          name: "SPOJ",
          logo: require("@/assets/spoj-logo.png"),
          status: true,
        },
        {
          url: "https://loj.ac/",
          name: "LibreOJ",
          logo: require("@/assets/libre-logo.png"),
          status: true,
        },
      ],
    };
  },
  mounted() {
    let screenWidth = window.screen.width;
    if (screenWidth < 768) {
      this.srcHight = "200px";
    } else {
      this.srcHight = "440px";
    }
    this.CONTEST_STATUS_REVERSE = Object.assign({}, CONTEST_STATUS_REVERSE);
    this.CONTEST_TYPE_REVERSE = Object.assign({}, CONTEST_TYPE_REVERSE);
    this.getHomeCarousel();
    this.getRecentContests();
    // 使用 nextTick 确保 DOM 渲染后再异步加载 Rating 数据
    this.$nextTick(() => {
      this.getRatingRank();
    });
    this.getRecentUpdatedProblemList();
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
          this.contests = res.data.data;
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
    getRecentUpdatedProblemList() {
      this.loading.recentUpdatedProblemsLoading = true;
      api.getRecentUpdatedProblemList().then(
        (res) => {
          this.recentUpdatedProblems = res.data.data;
          this.loading.recentUpdatedProblemsLoading = false;
        },
        (err) => {
          this.loading.recentUpdatedProblemsLoading = false;
        }
      );
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
    goProblem(event) {
      this.$router.push({
        name: "ProblemDetails",
        params: {
          problemID: event.row.problemId,
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
  border-color: rgb(25, 190, 107);
}
.contest-card-schedule {
  border-color: #f90;
}
</style>
<style scoped>
/deep/.el-card__header {
  padding: 0.6rem 1.25rem !important;
}
.card-top {
  margin-top: 20px;
}
.home-contest {
  text-align: left;
  font-size: 21px;
  font-weight: 500;
  line-height: 30px;
}
.oj-logo {
  border: 1px solid rgba(0, 0, 0, 0.15);
  border-radius: 4px;
  margin-bottom: 1rem;
  padding: 0.5rem 1rem;
  background: rgb(255, 255, 255);
  min-height: 47px;
}
.oj-normal {
  border-color: #409eff;
}
.oj-error {
  border-color: #e65c47;
}

.el-carousel__item h3 {
  color: #475669;
  font-size: 14px;
  opacity: 0.75;
  line-height: 200px;
  margin: 0;
}

.contest-card {
  margin-bottom: 20px;
}
.contest-title {
  font-size: 1.15rem;
  font-weight: 600;
}
.contest-type-auth {
  text-align: center;
  margin-top: -10px;
  margin-bottom: 5px;
}
ul,
li {
  padding: 0;
  margin: 0;
  list-style: none;
}
.contest-info {
  text-align: center;
}
.contest-info li {
  display: inline-block;
  padding-right: 10px;
}

/deep/.contest-card-running .el-card__header {
  border-color: rgb(25, 190, 107);
  background-color: rgba(94, 185, 94, 0.15);
}
.contest-card-running .contest-title {
  color: #5eb95e;
}

/deep/.contest-card-schedule .el-card__header {
  border-color: #f90;
  background-color: rgba(243, 123, 29, 0.15);
}

.contest-card-schedule .contest-title {
  color: #f37b1d;
}

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
.welcome-title {
  font-weight: 600;
  font-size: 25px;
  font-family: "Raleway";
}
.contest-status {
  float: right;
}
.img-carousel {
  height: 490px;
}

@media screen and (max-width: 768px) {
  .contest-status {
    text-align: center;
    float: none;
    margin-top: 5px;
  }
  .contest-header {
    text-align: center;
  }
  .img-carousel {
    height: 220px;
    overflow: hidden;
  }
  .phone-margin {
    margin-top: 20px;
  }
}
.title .el-link {
  font-size: 21px;
  font-weight: 500;
  color: #444;
}
.clearfix h2 {
  color: #409eff;
}
.el-link.el-link--default:hover {
  color: #409eff;
  transition: all 0.28s ease;
}
.contest .content-info {
  padding: 0 70px 40px 70px;
}
.contest .contest-description {
  margin-top: 25px;
}
span.rank-tag.no1 {
  line-height: 24px;
  background: #bf2c24;
}

span.rank-tag.no2 {
  line-height: 24px;
  background: #e67225;
}

span.rank-tag.no3 {
  line-height: 24px;
  background: #e6bf25;
}

span.rank-tag {
  font: 16px/22px FZZCYSK;
  min-width: 14px;
  height: 22px;
  padding: 0 4px;
  text-align: center;
  color: #fff;
  background: #000;
  background: rgba(0, 0, 0, 0.6);
}
.user-avatar {
  margin-right: 5px !important;
  vertical-align: middle;
}
.cite {
  display: block;
  width: 14px;
  height: 0;
  margin: 0 auto;
  margin-top: -3px;
  border-right: 11px solid transparent;
  border-bottom: 0 none;
  border-left: 11px solid transparent;
}
.cite.no0 {
  border-top: 5px solid #bf2c24;
}
.cite.no1 {
  border-top: 5px solid #e67225;
}
.cite.no2 {
  border-top: 5px solid #e6bf25;
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
</style>
