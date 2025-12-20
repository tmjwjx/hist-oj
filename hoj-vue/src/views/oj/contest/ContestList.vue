<template>
  <div>
      <!-- <ContestListAttention></ContestListAttention> -->
      <el-row type="flex" justify="space-around">
        <el-col :span="24">
          <el-card shadow>
            <div slot="header">
              <span class="panel-title"
                >{{
                  query.type === '' ? $t('m.All') : parseContestType(query.type)
                }}
                {{ $t('m.Contests') }}</span
              >
              <div class="filter-row">
                <span>
                  <el-dropdown
                    @command="onRuleChange"
                    placement="bottom"
                    trigger="hover"
                    class="drop-menu"
                  >
                    <span class="el-dropdown-link">
                      {{
                        query.type === ''
                          ? $t('m.Contest_Rule')
                          : parseContestType(query.type)
                      }}
                      <i class="el-icon-caret-bottom"></i>
                    </span>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item command="">{{
                        $t('m.All')
                      }}</el-dropdown-item>
                      <el-dropdown-item command="0">ACM</el-dropdown-item>
                      <el-dropdown-item command="1">OI</el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </span>

                <span>
                  <el-dropdown
                    @command="onStatusChange"
                    placement="bottom"
                    trigger="hover"
                    class="drop-menu"
                  >
                    <span class="el-dropdown-link">
                      {{
                        query.status === ''
                          ? $t('m.Status')
                          : $t('m.' + CONTEST_STATUS_REVERSE[query.status]['name'])
                      }}
                      <i class="el-icon-caret-bottom"></i>
                    </span>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item command="">{{
                        $t('m.All')
                      }}</el-dropdown-item>
                      <el-dropdown-item command="-1">{{
                        $t('m.Scheduled')
                      }}</el-dropdown-item>
                      <el-dropdown-item command="0">{{
                        $t('m.Running')
                      }}</el-dropdown-item>
                      <el-dropdown-item command="1">{{
                        $t('m.Ended')
                      }}</el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </span>

                <span>
                  <vxe-input
                    v-model="query.keyword"
                    :placeholder="$t('m.Enter_keyword')"
                    type="search"
                    size="medium"
                    @keyup.enter.native="onKeywordChange"
                    @search-click="onKeywordChange"
                  ></vxe-input>
                </span>
              </div>
            </div>
            <div v-loading="loading">
              <p id="no-contest" v-show="contests.length == 0">
                <el-empty :description="$t('m.No_contest')"></el-empty>
              </p>
              <ol id="contest-list">
                <li
                  v-for="contest in contests"
                  :key="contest.title"
                  :style="getborderColor(contest)"
                >
                  <el-row type="flex" justify="space-between" align="middle">
                    <el-col :xs="10" :sm="4" :md="3" :lg="2">
                      <template v-if="contest.type == 0">
                        <el-image 
                        :src="acmSrc" 
                        class="trophy"
                        style="width: 100px;"
                        :preview-src-list="[acmSrc]">
                        </el-image>
                      </template>
                      <template v-else>
                        <el-image 
                        :src="oiSrc" 
                        class="trophy"
                        style="width: 100px;"
                        :preview-src-list="[oiSrc]">
                        </el-image>
                      </template>
                    </el-col>
                    <el-col
                      :xs="10"
                      :sm="16"
                      :md="19"
                      :lg="20"
                      class="contest-main"
                    >
                      <p class="title">
                        <a class="entry" @click.stop="toContest(contest)">
                          {{ contest.title }}
                        </a>
                        <template v-if="contest.auth == 1">
                          <i
                            class="el-icon-lock"
                            size="20"
                            style="color:#d9534f"
                          ></i>
                        </template>
                        <template v-if="contest.auth == 2">
                          <i
                            class="el-icon-lock"
                            size="20"
                            style="color:#f0ad4e"
                          ></i>
                        </template>
                      </p>
                      <ul class="detail">
                        <li>
                          <i
                            class="fa fa-calendar"
                            aria-hidden="true"
                            style="color: #3091f2"
                          ></i>
                          {{ contest.startTime | localtime }}
                        </li>
                        <li>
                          <i
                            class="fa fa-clock-o"
                            aria-hidden="true"
                            style="color: #3091f2"
                          ></i>
                          {{ getDuration(contest.startTime, contest.endTime) }}
                        </li>
                        <li>
                          <template v-if="contest.type == 0">
                            <el-button
                              size="mini"
                              round
                              :type="'primary'"
                              @click="onRuleChange(contest.type)"
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
                                size="mini"
                                round
                                :type="'warning'"
                                @click="onRuleChange(contest.type)"
                                ><i class="fa fa-trophy"></i>
                                {{ contest.type | parseContestType }}
                              </el-button>
                            </el-tooltip>
                          </template>
                        </li>
                        <li v-if="isRatingContest(contest.id)">
                          <el-tooltip
                            content="Rating 比赛"
                            placement="top"
                            effect="dark"
                          >
                            <el-tag
                              type="danger"
                              effect="plain"
                              size="small"
                            >
                              <i class="fa fa-star"></i> Rating
                            </el-tag>
                          </el-tooltip>
                        </li>
                        <li>
                          <el-tooltip
                            :content="
                              $t('m.' + CONTEST_TYPE_REVERSE[contest.auth].tips)
                            "
                            placement="top"
                            effect="light"
                          >
                            <el-tag
                              :type="CONTEST_TYPE_REVERSE[contest.auth]['color']"
                              effect="plain"
                            >
                              {{
                                $t(
                                  'm.' + CONTEST_TYPE_REVERSE[contest.auth]['name']
                                )
                              }}
                            </el-tag>
                          </el-tooltip>
                        </li>
                        <li v-if="contest.count != null">
                          <i
                            class="el-icon-user-solid"
                            style="color:rgb(48, 145, 242);"
                          ></i
                          >x{{ contest.count }}
                        </li>
                        <li v-if="contest.openRank">
                          <el-tooltip
                            :content="$t('m.Contest_Outside_ScoreBoard')"
                            placement="top"
                            effect="dark"
                          >
                            <el-button
                              circle
                              size="small"
                              type="primary"
                              :disabled="contest.status == CONTEST_STATUS.SCHEDULED"
                              icon="el-icon-data-analysis"
                              @click="
                                toContestOutsideScoreBoard(contest.id, contest.type)
                              "
                            ></el-button>
                          </el-tooltip>
                        </li>
                      </ul>
                    </el-col>
                    <el-col
                      :xs="4"
                      :sm="4"
                      :md="2"
                      :lg="2"
                      style="text-align: center"
                    >
                      <el-tag
                        effect="dark"
                        :color="CONTEST_STATUS_REVERSE[contest.status]['color']"
                        size="medium"
                      >
                        <i class="fa fa-circle" aria-hidden="true"></i>
                        {{
                          $t('m.' + CONTEST_STATUS_REVERSE[contest.status]['name'])
                        }}
                      </el-tag>
                    </el-col>
                  </el-row>
                </li>
              </ol>
            </div>
          </el-card>
          <Pagination
            :total="total"
            :pageSize="limit"
            @on-change="onCurrentPageChange"
            :current.sync="currentPage"
          ></Pagination>
        </el-col>
      </el-row>
  </div>
</template>

<script>
import api from '@/common/api';
import ratingApi from '@/common/rating-api';
import { mapGetters } from 'vuex';
import utils from '@/common/utils';
import time from '@/common/time';
import {
  CONTEST_STATUS_REVERSE,
  CONTEST_TYPE_REVERSE,
  CONTEST_STATUS,
} from '@/common/constants';
import myMessage from '@/common/message';
const Pagination = () => import('@/components/oj/common/Pagination');
// const ContestListAttention = () => import('@/components/oj/contest/ContestListAttention');
const limit = 10;

export default {
  name: 'contest-list',
  components: {
    Pagination,
    // ContestListAttention
  },
  data() {
    return {
      currentPage: 1,
      query: {
        status: '',
        keyword: '',
        type: '',
      },
      limit: limit,
      total: 0,
      rows: '',
      contests: [],
      ratingContests: new Set(), // 存储 Rating 比赛的 ID
      CONTEST_STATUS_REVERSE: {},
      CONTEST_STATUS: {},
      CONTEST_TYPE_REVERSE: {},
      acmSrc: require('@/assets/acm.jpg'),
      oiSrc: require('@/assets/oi.jpg'),
      loading: true,
    };
  },
  created() {
    let route = this.$route.query;
    this.currentPage = parseInt(route.currentPage) || 1;
  },
  mounted() {
    this.CONTEST_STATUS_REVERSE = Object.assign({}, CONTEST_STATUS_REVERSE);
    this.CONTEST_TYPE_REVERSE = Object.assign({}, CONTEST_TYPE_REVERSE);
    this.CONTEST_STATUS = Object.assign({}, CONTEST_STATUS);
    this.init();
  },
  methods: {
    init() {
      let route = this.$route.query;
      this.query.status = route.status || '';
      if(route.type === 0 || route.type === '0'){
        this.query.type = 0;
      }else if(route.type === 1 || route.type === '1'){
        this.query.type = 1;
      }else{
        this.query.type = '';
      }
      this.query.keyword = route.keyword || '';
      this.currentPage = parseInt(route.currentPage) || 1;
      this.getContestList();
    },
    getContestList() {
      this.loading = true;
      api.getContestList(this.currentPage, this.limit, this.query).then(
        (res) => {
          this.contests = res.data.data.records;
          this.total = res.data.data.total;
          this.loading = false;

          // 异步获取 Rating 比赛信息
          this.fetchRatingContests();
        },
        (err) => {
          this.loading = false;
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

    filterByChange() {
      let query = Object.assign({}, this.query);
      query.currentPage = this.currentPage;
      this.$router.push({
        name: 'ContestList',
        query: utils.filterEmptyValue(query),
      });
    },

    parseContestType(type) {
      if (type == 0) {
        return 'ACM';
      } else if (type == 1) {
        return 'OI';
      }
    },

    onCurrentPageChange(page) {
      this.currentPage = page;
      this.filterByChange();
    },
    onRuleChange(rule) {
      this.query.type = rule;
      this.currentPage = 1;
      this.filterByChange();
    },
    onStatusChange(status) {
      this.query.status = status;
      this.currentPage = 1;
      this.filterByChange();
    },
    onKeywordChange() {
      this.currentPage = 1;
      this.filterByChange();
    },
    toContest(contest) {
      if (!this.isAuthenticated) {
        myMessage.warning(this.$i18n.t('m.Please_login_first'));
        this.$store.dispatch('changeModalStatus', { visible: true });
      } else {
        this.$router.push({
          name: 'ContestDetails',
          params: { contestID: contest.id },
        });
      }
    },
    toContestOutsideScoreBoard(cid, type) {
      if (type == 0) {
        this.$router.push({
          name: 'ACMScoreBoard',
          params: { contestID: cid },
        });
      } else if (type == 1) {
        this.$router.push({
          name: 'OIScoreBoard',
          params: { contestID: cid },
        });
      }
    },
    getDuration(startTime, endTime) {
      return time.formatSpecificDuration(startTime, endTime);
    },
    getborderColor(contest) {
      return (
        'border-left: 4px solid ' +
        CONTEST_STATUS_REVERSE[contest.status]['color']
      );
    },
  },
  computed: {
    ...mapGetters(['isAuthenticated', 'userInfo']),
  },
  watch: {
    $route(newVal, oldVal) {
      if (newVal !== oldVal) {
        this.init();
      }
    },
  },
};
</script>
<style scoped>
#no-contest {
  text-align: center;
  font-size: 16px;
  padding: 20px;
}
.filter-row {
  float: right;
}
@media screen and (max-width: 768px) {
  .filter-row span {
    margin-right: 2px;
  }
  ol {
    padding-inline-start: 5px;
  }
  /deep/ .el-card__header {
    margin-bottom: 5px;
  }
}
@media screen and (min-width: 768px) {
  .filter-row span {
    margin-right: 20px;
  }
}
/deep/ .el-card__header {
  border-bottom: 0px;
}

#contest-list > li {
  padding: 5px;
  margin-left: -20px;
  margin-top: 10px;
  width: 100%;
  border-bottom: 1px solid rgba(187, 187, 187, 0.5);
  list-style: none;
}
#contest-list .trophy {
  height: 70px;
  margin-left: 10px;
  margin-right: -20px;
}
@media screen and (max-width: 1500px) and (min-width: 1200px){
  #contest-list .trophy {
    width: 100% !important;
  }
  #contest-list .contest-main{
    margin-left: 20px;
  }
}
#contest-list .contest-main {
  text-align: left;
}
#contest-list .contest-main .title {
  font-size: 1.25rem;
  padding-left: 8px;
  margin-bottom: 0;
}
#contest-list .contest-main .title a.entry {
  color: #495060;
}
#contest-list .contest-main .title a:hover {
  color: #2d8cf0;
  border-bottom: 1px solid #2d8cf0;
}
#contest-list .contest-main .detail {
  font-size: 0.875rem;
  padding-left: 0;
  padding-bottom: 10px;
}
#contest-list .contest-main li {
  display: inline-block;
  padding: 10px 0 0 10px;
}
</style>
