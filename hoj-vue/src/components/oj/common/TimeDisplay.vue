<template>
  <el-card class="time-card card-top">
    <div slot="header" class="clearfix">
      <span class="panel-title home-title">
        <i class="el-icon-time"></i> {{ $t('m.Current_Time') }}
      </span>
    </div>
    <div class="time-body">
      <div class="time-text">{{ formattedTime }}</div>
      <div class="date-text">{{ formattedDate }}</div>
    </div>
  </el-card>
</template>

<script>
export default {
  name: 'TimeDisplay',
  data() {
    return {
      currentTime: new Date(),
      timer: null
    };
  },
  computed: {
    formattedTime() {
      const hours = String(this.currentTime.getHours()).padStart(2, '0');
      const minutes = String(this.currentTime.getMinutes()).padStart(2, '0');
      const seconds = String(this.currentTime.getSeconds()).padStart(2, '0');
      return `${hours}:${minutes}:${seconds}`;
    },
    formattedDate() {
      const localeMap = {
        'zh-CN': 'zh-CN',
        'zh-TW': 'zh-TW',
        'en-US': 'en-US',
        'ja-JP': 'ja-JP',
        'ko-KR': 'ko-KR',
      };
      return new Intl.DateTimeFormat(localeMap[this.$i18n.locale] || 'en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
      }).format(this.currentTime);
    }
  },
  mounted() {
    this.updateTime();
    this.timer = setInterval(this.updateTime, 1000);
  },
  beforeDestroy() {
    if (this.timer) {
      clearInterval(this.timer);
    }
  },
  methods: {
    updateTime() {
      this.currentTime = new Date();
    }
  }
};
</script>

<style scoped>
.time-card {
  margin-bottom: 20px;
}

/* 与首页其他卡片标题保持一致（panel-title/home-title 在 Home.vue 中为 scoped 样式） */
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

.home-title i {
  color: #409eff;
}

.time-body {
  text-align: center;
  padding: 8px 0 4px;
}

.time-text {
  font-size: 40px;
  font-weight: 600;
  color: #409eff;
  font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
  font-variant-numeric: tabular-nums;
  letter-spacing: 2px;
  line-height: 1.2;
}

.date-text {
  margin-top: 6px;
  font-size: 14px;
  color: #909399;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .time-text {
    font-size: 30px;
    letter-spacing: 1px;
  }

  .date-text {
    font-size: 13px;
  }
}
</style>
