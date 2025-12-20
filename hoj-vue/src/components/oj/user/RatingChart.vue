<template>
  <div class="rating-chart">
    <div v-if="loading" class="loading">
      <i class="el-icon-loading"></i> 加载中...
    </div>
    <div v-else-if="error" class="error">
      <i class="el-icon-warning"></i> {{ error }}
    </div>
    <div v-else-if="chartData.length === 0" class="empty">
      <i class="el-icon-info"></i> 暂无 Rating 历史记录
      <p style="font-size: 12px; margin-top: 8px; color: #999;">
        参加 Rating 比赛后，这里将显示您的 Rating 变化历史
      </p>
    </div>
    <div v-else ref="chart" style="width: 100%; height: 400px;"></div>
  </div>
</template>

<script>
import * as echarts from 'echarts'
import ratingApi from '@/common/rating-api'
import { getRatingColor } from '@/common/rating-utils'

export default {
  name: 'RatingChart',
  props: {
    uid: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      loading: false,
      error: null,
      chartData: [],
      chart: null
    }
  },
  mounted() {
    if (this.uid && this.uid !== 'undefined') {
      this.fetchData()
    } else {
      this.error = '用户ID无效'
    }
  },
  watch: {
    uid(newVal, oldVal) {
      if (newVal && newVal !== 'undefined' && newVal !== oldVal) {
        this.fetchData()
      }
    }
  },
  beforeDestroy() {
    if (this.chart) {
      this.chart.dispose()
    }
  },
  methods: {
    async fetchData() {
      if (!this.uid || this.uid === 'undefined') {
        this.error = '用户ID无效'
        return
      }

      this.loading = true
      this.error = null
      try {
        const data = await ratingApi.getRatingHistory(this.uid, 1, 100)
        console.log('Rating history response:', data)
        // API 返回的数据是倒序的（最新在前），需要反转为正序（最早在前）
        const records = data.records || data || []
        this.chartData = Array.isArray(records) ? [...records].reverse() : []
        console.log('Chart data:', this.chartData)
      } catch (error) {
        console.error('获取 Rating 历史失败:', error)
        this.error = '加载失败: ' + (error.message || '未知错误')
      } finally {
        this.loading = false
        // 确保 DOM 更新后再渲染图表
        this.$nextTick(() => {
          this.$nextTick(() => {
            this.renderChart()
          })
        })
      }
    },
    renderChart() {
      console.log('[RatingChart] renderChart called')
      console.log('[RatingChart] this.$refs.chart:', this.$refs.chart)
      console.log('[RatingChart] this.chartData.length:', this.chartData.length)

      if (!this.$refs.chart || this.chartData.length === 0) {
        console.warn('[RatingChart] renderChart early return:', {
          hasRef: !!this.$refs.chart,
          dataLength: this.chartData.length
        })
        return
      }

      console.log('[RatingChart] Initializing echarts...')
      this.chart = echarts.init(this.$refs.chart)
      console.log('[RatingChart] echarts initialized:', this.chart)

      const dates = this.chartData.map(item =>
        new Date(item.contest_time).toLocaleDateString()
      )
      const ratings = this.chartData.map(item => item.new_rating)

      const option = {
        title: {
          text: 'Rating 变化历史',
          left: 'center'
        },
        tooltip: {
          trigger: 'axis',
          formatter: (params) => {
            const data = this.chartData[params[0].dataIndex]
            return `
              <div style="text-align: left;">
                <strong>${data.contest_title}</strong><br/>
                Rating: ${data.old_rating} → ${data.new_rating}<br/>
                变化: ${data.rating_change > 0 ? '+' : ''}${data.rating_change}<br/>
                排名: ${data.rank}
              </div>
            `
          }
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true
        },
        xAxis: {
          type: 'category',
          data: dates,
          boundaryGap: false
        },
        yAxis: {
          type: 'value',
          name: 'Rating'
        },
        series: [{
          data: ratings,
          type: 'line',
          smooth: true,
          itemStyle: {
            color: '#409EFF'
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
              { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
            ])
          },
          markPoint: {
            data: [
              { type: 'max', name: '最高' },
              { type: 'min', name: '最低' }
            ]
          }
        }]
      }

      this.chart.setOption(option)

      // 响应式调整
      window.addEventListener('resize', () => {
        if (this.chart) {
          this.chart.resize()
        }
      })
    }
  }
}
</script>

<style scoped>
.rating-chart {
  padding: 20px;
  background: #fff;
  border-radius: 4px;
}
.loading, .error, .empty {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 14px;
}
.error {
  color: #f56c6c;
}
.loading i {
  font-size: 24px;
  margin-right: 8px;
}
</style>
