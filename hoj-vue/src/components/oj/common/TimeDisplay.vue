<template>
  <div class="time-widget">
    <div class="time-container">
      <!-- 时分秒部分 -->
      <div class="time-main">
        <div class="time-display">
          <i class="el-icon-time time-icon"></i>
          <span class="time-text">{{ formattedTime }}</span>
        </div>
      </div>

      <!-- 年月日部分 -->
      <div class="date-display">
        <i class="el-icon-date date-icon"></i>
        <span class="date-text">{{ formattedDate }}</span>
      </div>

      <!-- 装饰元素 -->
      <div class="decoration-line"></div>
    </div>
  </div>
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
        month: '2-digit',
        day: '2-digit',
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
.time-widget {
  width: 100%;
  margin-bottom: 20px;
}

.time-container {
  background: linear-gradient(135deg, #e0f7fa 0%, #e8f5e9 100%);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.time-container::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  animation: rotate 20s linear infinite;
}

@keyframes rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.time-container:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 28px rgba(0, 0, 0, 0.12);
}

/* 时分秒部分 */
.time-main {
  position: relative;
  z-index: 1;
  margin-bottom: 16px;
}

.time-display {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.time-icon {
  font-size: 32px;
  color: #26a69a;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.05);
  }
}

.time-text {
  font-size: 48px;
  font-weight: 700;
  background: linear-gradient(135deg, #26a69a 0%, #66bb6a 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-family: 'Arial', 'Helvetica', monospace;
  letter-spacing: 3px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

/* 年月日部分 */
.date-display {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding-top: 8px;
}

.date-icon {
  font-size: 20px;
  color: #66bb6a;
}

.date-text {
  font-size: 18px;
  font-weight: 500;
  color: #4db6ac;
  font-family: 'Arial', sans-serif;
  letter-spacing: 1px;
}

/* 装饰线 */
.decoration-line {
  position: relative;
  z-index: 1;
  height: 3px;
  background: linear-gradient(90deg,
    transparent 0%,
    #26a69a 20%,
    #66bb6a 50%,
    #26a69a 80%,
    transparent 100%
  );
  margin-top: 16px;
  border-radius: 2px;
}

.decoration-line::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 8px;
  height: 8px;
  background: #fff;
  border: 2px solid #26a69a;
  border-radius: 50%;
  box-shadow: 0 0 10px rgba(38, 166, 154, 0.5);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .time-container {
    padding: 16px;
  }

  .time-icon {
    font-size: 24px;
  }

  .time-text {
    font-size: 36px;
    letter-spacing: 2px;
  }

  .date-icon {
    font-size: 18px;
  }

  .date-text {
    font-size: 16px;
  }
}
</style>
