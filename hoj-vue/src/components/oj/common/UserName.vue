<template>
  <span
    :style="{ color: userColor, cursor: clickable ? 'pointer' : 'inherit', fontWeight: bold ? 'bold' : 'normal' }"
    @click="handleClick"
    :title="tooltip"
    class="username-colored"
  >
    <slot>{{ displayName }}</slot>
  </span>
</template>

<script>
import ratingApi from '@/common/rating-api';
import { getRatingColor } from '@/common/rating-utils';

// 全局 Rating 缓存 - 会话级别
const ratingCache = new Map();
// 正在进行的请求，避免重复请求
const pendingRequests = new Map();

export default {
  name: 'UserName',
  props: {
    username: {
      type: String,
      required: true
    },
    clickable: {
      type: Boolean,
      default: false
    },
    showTooltip: {
      type: Boolean,
      default: false
    },
    bold: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      rating: null,
      loading: false
    }
  },
  computed: {
    displayName() {
      return this.username;
    },
    userColor() {
      if (this.rating !== null) {
        return getRatingColor(this.rating);
      }
      // 加载中或失败时使用默认颜色
      return 'inherit';
    },
    tooltip() {
      if (!this.showTooltip || this.rating === null) return '';
      return `Rating: ${this.rating}`;
    }
  },
  mounted() {
    this.fetchRating();
  },
  watch: {
    username(newVal, oldVal) {
      if (newVal !== oldVal) {
        this.fetchRating();
      }
    }
  },
  methods: {
    async fetchRating() {
      if (!this.username) return;

      // 检查缓存
      if (ratingCache.has(this.username)) {
        this.rating = ratingCache.get(this.username);
        return;
      }

      // 检查是否有正在进行的请求
      if (pendingRequests.has(this.username)) {
        try {
          const rating = await pendingRequests.get(this.username);
          this.rating = rating;
        } catch (error) {
          // 请求失败，使用默认值
          this.rating = 1500;
        }
        return;
      }

      // 发起新请求
      this.loading = true;
      const promise = ratingApi.getUserRating(this.username)
        .then(data => {
          const rating = data ? data.rating : 1500;
          ratingCache.set(this.username, rating);
          pendingRequests.delete(this.username);
          return rating;
        })
        .catch(() => {
          // 失败时使用默认值
          const defaultRating = 1500;
          ratingCache.set(this.username, defaultRating);
          pendingRequests.delete(this.username);
          return defaultRating;
        })
        .finally(() => {
          this.loading = false;
        });

      pendingRequests.set(this.username, promise);
      this.rating = await promise;
    },
    handleClick() {
      if (this.clickable) {
        this.$emit('click', this.username);
      }
    }
  }
}
</script>

<style scoped>
.username-colored {
  transition: color 0.3s ease;
}
</style>
