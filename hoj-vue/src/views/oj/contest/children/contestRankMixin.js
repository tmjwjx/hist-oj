import api from '@/common/api'
import { mapGetters, mapState } from 'vuex'
import { CONTEST_STATUS, buildContestRankConcernedKey } from '@/common/constants'
import storage from '@/common/storage';
import ratingApi from '@/common/rating-api';
import { getRatingColor } from '@/common/rating-utils';

export default {
  data() {
    return {
      userRatings: new Map(), // 存储用户当前 Rating
      contestRatings: new Map(), // 存储比赛 Rating 变化
      isRatingContest: false, // 是否为 Rating 比赛
      contestInfoFetched: false, // 是否已获取比赛信息
      ratingDataLoaded: false, // Rating 数据是否已加载完成
      ratingDataRetryCount: 0, // Rating 数据重试次数
      ratingCache: {}, // 缓存计算结果，避免重复查询
    }
  },
  methods: {
    initConcernedList(){
      let key = buildContestRankConcernedKey(this.$route.params.contestID);
      this.concernedList = storage.get(key) || [];
    },
    getContestRankData (page = 1, refresh = false) {
      // 首次加载时获取比赛信息
      if (!this.contestInfoFetched) {
        this.fetchContestInfo()
          .then(() => {
            this.contestInfoFetched = true;
            this.loadRankData(page, refresh);
          })
          .catch((error) => {
            console.error('获取比赛信息失败，但继续加载排名数据:', error);
            this.contestInfoFetched = true;
            this.isRatingContest = false;
            this.loadRankData(page, refresh);
          });
        return;
      }

      this.loadRankData(page, refresh);
    },
    loadRankData(page, refresh) {
      if (this.showChart && !refresh && this.$refs.chart) {
        this.$refs.chart.showLoading({maskColor: 'rgba(250, 250, 250, 0.8)'})
      }
      let data = {
        currentPage:page,
        limit: this.limit,
        cid: this.$route.params.contestID,
        forceRefresh: this.forceUpdate ? true: false,
        removeStar: !this.showStarUser,
        concernedList:this.concernedList,
        keyword: this.keyword == null? null: this.keyword.trim(),
        containsEnd: this.isContainsAfterContestJudge,
      }
      api.getContestRank(data).then(res => {
        if (this.showChart && !refresh && this.$refs.chart) {
          this.$refs.chart.hideLoading()
        }

        if (!res.data || !res.data.data) {
          console.error('排名数据格式错误:', res);
          return;
        }

        this.total = res.data.data.total
        const records = res.data.data.records || [];

        // 先立即渲染表格，不等待 Rating 数据
        if (page === 1 && this.showChart && this.$refs.chart) {
          try {
            this.applyToChart(records)
          } catch (error) {
            console.error('渲染图表失败:', error);
          }
        }

        try {
          this.applyToTable(records)
        } catch (error) {
          console.error('渲染表格失败:', error);
        }

        // 异步获取 Rating 数据（所有比赛都需要获取用户当前 Rating 来显示颜色）
        this.fetchRatingData(records).catch(error => {
          console.error('获取 Rating 数据失败:', error);
        });
      }).catch(error => {
        console.error('获取排名数据失败:', error);
        if (this.showChart && !refresh && this.$refs.chart) {
          this.$refs.chart.hideLoading()
        }
      })
    },
    handleAutoRefresh (status) {
      if (status == true) {
        this.refreshFunc = setInterval(() => {
          this.$store.dispatch('getContestProblems');
          this.getContestRankData(this.page, true);
        }, 10000)
      } else {
        clearInterval(this.refreshFunc)
      }
    },
    updateConcernedList(uid, isConcerned){
      if(isConcerned){
        this.concernedList.push(uid);
      }else{
        var index = this.concernedList.indexOf(uid);
        if (index > -1) {
        this.concernedList.splice(index, 1);
        }
      }
      let key = buildContestRankConcernedKey(this.contestID);
      storage.set(key, this.concernedList);
      this.getContestRankData(this.page, true);
    },
    getRankShowName(rankShowName, username){
      let finalShowName = rankShowName;
      if(rankShowName == null || rankShowName == '' || rankShowName.trim().length == 0){
        finalShowName = username;
      }
      return finalShowName;
    },
    async fetchContestInfo() {
      try {
        const contestId = this.$route.params.contestID;
        const contestInfo = await ratingApi.getContestInfo(contestId);
        this.isRatingContest = contestInfo.isRating || false;
      } catch (error) {
        console.warn('获取比赛信息失败，默认为非 Rating 比赛:', error);
        this.isRatingContest = false;
      }
    },
    async fetchRatingData(records, retryCount = 0) {
      if (!records || records.length === 0) return;

      const MAX_RETRY = 1; // 减少重试次数到1次
      let hasError = false;

      try {
        const contestId = this.$route.params.contestID;

        // 1. 如果是 Rating 比赛，获取比赛参赛者的 Rating 变化
        console.log(`[Rating Debug] isRatingContest=${this.isRatingContest}, contestId=${contestId}`)
        if (this.isRatingContest) {
          try {
            const contestRatingData = await ratingApi.getContestParticipantsRating(contestId);
            const ratingRecords = contestRatingData.records || [];
            console.log(`[Rating Debug] Loaded ${ratingRecords.length} rating records for contest ${contestId}`)

            // 清空旧数据和缓存
            this.contestRatings.clear();
            this.ratingCache = {};

            ratingRecords.forEach(item => {
              if (item && item.uid) {
                const ratingInfo = {
                  oldRating: item.oldRating,
                  newRating: item.newRating,
                  ratingChange: item.ratingChange,
                  isSkip: item.isSkip || false,
                  skipReason: item.skipReason || '',
                };
                this.contestRatings.set(item.uid, ratingInfo);

                // 预计算并缓存常用数据
                this.ratingCache[`change_${item.uid}`] = item.ratingChange;
                this.ratingCache[`rating_${item.uid}`] = item.newRating;
                this.ratingCache[`color_${item.uid}`] = getRatingColor(item.newRating);
                this.ratingCache[`isskip_${item.uid}`] = item.isSkip || false;
                this.ratingCache[`skipreason_${item.uid}`] = item.skipReason || '';

                if (item.isSkip) {
                  console.log(`[Rating Debug] Found skipped user: uid=${item.uid}, reason=${item.skipReason}`)
                }
              }
            });
            console.log(`[Rating Debug] Total skipped users: ${ratingRecords.filter(r => r.isSkip).length}`)
          } catch (error) {
            console.error('获取比赛 Rating 数据失败:', error);
            hasError = true;
          }
        } else {
          console.log(`[Rating Debug] Contest ${contestId} is NOT a rating contest, skipping rating data load`)
        }

        // 2. 批量获取用户当前 Rating（用于用户名着色）- 所有比赛都需要
        try {
          const uids = records.map(r => r.uid).filter(uid => uid);
          if (uids.length > 0) {
            const batchRatingsData = await ratingApi.getBatchUserRating(uids);

            // 清空旧数据（如果不是 Rating 比赛，这里才清空）
            if (!this.isRatingContest) {
              this.userRatings.clear();
              this.ratingCache = {};
            }

            // 存储用户当前 Rating
            Object.entries(batchRatingsData).forEach(([uid, data]) => {
              if (data) {
                const username = records.find(r => r.uid === uid)?.username;
                const ratingInfo = {
                  rating: data.rating,
                  color: data.color,
                };
                // 同时用 uid 和 username 作为 key 存储
                this.userRatings.set(uid, ratingInfo);
                if (username) {
                  this.userRatings.set(username, ratingInfo);
                }

                // 缓存用户颜色
                this.ratingCache[`usercolor_${username}`] = data.color;
              }
            });
          }
        } catch (error) {
          console.warn('获取用户 Rating 数据失败:', error);
          hasError = true;
        }

        // 标记 Rating 数据已加载完成
        this.ratingDataLoaded = true;

        // 强制更新视图
        this.$forceUpdate();

        // 只在第一次失败时重试
        if (hasError && retryCount === 0) {
          setTimeout(() => {
            this.fetchRatingData(records, 1);
          }, 1500);
        }
      } catch (error) {
        console.error('获取 Rating 数据时发生错误:', error);

        // 只重试一次
        if (retryCount === 0) {
          setTimeout(() => {
            this.fetchRatingData(records, 1);
          }, 2000);
        }
      }
    },
    getUserRatingColor(username) {
      // 优先使用缓存
      const cacheKey = `usercolor_${username}`;
      if (this.ratingCache[cacheKey]) {
        return this.ratingCache[cacheKey];
      }

      // 缓存未命中，查询 Map
      const ratingInfo = this.userRatings.get(username);
      if (ratingInfo) {
        this.ratingCache[cacheKey] = ratingInfo.color;
        return ratingInfo.color;
      }

      // 默认颜色（黑色，表示未评级）
      return '#000000';
    },
    getRatingChange(uid) {
      // Skip 用户没有 Rating 变化，返回 null
      const skipped = this.isUserSkipped(uid);
      console.log(`[Rating Debug] getRatingChange for uid=${uid}, skipped=${skipped}, isRatingContest=${this.isRatingContest}`)
      if (skipped) {
        return null;
      }

      // 优先使用缓存
      const cacheKey = `change_${uid}`;
      if (this.ratingCache[cacheKey] !== undefined) {
        console.log(`[Rating Debug] getRatingChange from cache: uid=${uid}, change=${this.ratingCache[cacheKey]}`)
        return this.ratingCache[cacheKey];
      }

      // 缓存未命中，查询 Map
      const contestRating = this.contestRatings.get(uid);
      if (contestRating && contestRating.ratingChange !== undefined) {
        this.ratingCache[cacheKey] = contestRating.ratingChange;
        console.log(`[Rating Debug] getRatingChange from Map: uid=${uid}, change=${contestRating.ratingChange}`)
        return contestRating.ratingChange;
      }
      console.log(`[Rating Debug] getRatingChange: uid=${uid}, not found, returning null`)
      return null;
    },
    // 获取比赛结束时的 Rating（newRating）
    getContestRating(uid) {
      // 优先使用缓存
      const cacheKey = `rating_${uid}`;
      if (this.ratingCache[cacheKey] !== undefined) {
        return this.ratingCache[cacheKey];
      }

      // 缓存未命中，查询 Map
      const contestRating = this.contestRatings.get(uid);
      if (contestRating && contestRating.newRating !== undefined && contestRating.newRating !== null) {
        this.ratingCache[cacheKey] = contestRating.newRating;
        return contestRating.newRating;
      }
      return null;
    },
    // 获取比赛结束时 Rating 的颜色
    getContestRatingColor(uid) {
      // 优先使用缓存
      const cacheKey = `color_${uid}`;
      if (this.ratingCache[cacheKey]) {
        return this.ratingCache[cacheKey];
      }

      // 缓存未命中，计算颜色
      const newRating = this.getContestRating(uid);
      if (newRating !== null) {
        const color = getRatingColor(newRating);
        this.ratingCache[cacheKey] = color;
        return color;
      }
      return '#000000'; // 默认黑色（未评级）
    },
    // 检查用户是否被 Skip
    isUserSkipped(uid) {
      // 优先使用缓存
      const cacheKey = `isskip_${uid}`;
      if (this.ratingCache[cacheKey] !== undefined) {
        console.log(`[Rating Debug] isUserSkipped from cache: uid=${uid}, skipped=${this.ratingCache[cacheKey]}`)
        return this.ratingCache[cacheKey];
      }

      // 缓存未命中，查询 Map
      const contestRating = this.contestRatings.get(uid);
      if (contestRating && contestRating.isSkip !== undefined) {
        this.ratingCache[cacheKey] = contestRating.isSkip;
        console.log(`[Rating Debug] isUserSkipped from Map: uid=${uid}, skipped=${contestRating.isSkip}, reason=${contestRating.skipReason}`)
        return contestRating.isSkip;
      }
      console.log(`[Rating Debug] isUserSkipped: uid=${uid}, not found in contestRatings (map size=${this.contestRatings.size}), returning false`)
      return false;
    },
    // 获取用户 Skip 原因
    getUserSkipReason(uid) {
      // 优先使用缓存
      const cacheKey = `skipreason_${uid}`;
      if (this.ratingCache[cacheKey] !== undefined) {
        return this.ratingCache[cacheKey];
      }

      // 缓存未命中，查询 Map
      const contestRating = this.contestRatings.get(uid);
      if (contestRating && contestRating.skipReason) {
        this.ratingCache[cacheKey] = contestRating.skipReason;
        return contestRating.skipReason;
      }
      return '';
    }
  },
  computed: {
    ...mapGetters(['isContestAdmin','userInfo', "isContainsAfterContestJudge"]),
    ...mapState({
      'contest': state => state.contest.contest,
      'contestProblems': state => state.contest.contestProblems
    }),
    showChart: {
      get () {
        return this.$store.state.contest.itemVisible.chart
      },
      set (value) {
        this.$store.commit('changeContestItemVisible', {chart: value})
      }
    },
    showStarUser:{
      get () {
        return !this.$store.state.contest.removeStar
      },
      set (value) {
        this.$store.commit('changeRankRemoveStar', {value: !value})
      }
    },
    showTable: {
      get () {
        return this.$store.state.contest.itemVisible.table
      },
      set (value) {
        this.$store.commit('changeContestItemVisible', {table: value})
      }
    },
    forceUpdate: {
      get () {
        return this.$store.state.contest.forceUpdate
      },
      set (value) {
        this.$store.commit('changeRankForceUpdate', {value: value})
      }
    },
    concernedList:{
      get () {
        return this.$store.state.contest.concernedList
      },
      set (value) {
        this.$store.commit('changeConcernedList', {value: value})
      }
    },
    refreshDisabled () {
      return this.contest.status == CONTEST_STATUS.ENDED
    }
  },
  beforeDestroy () {
    clearInterval(this.refreshFunc)
  }
}
