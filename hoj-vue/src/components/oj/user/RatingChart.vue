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
        // API 返回的数据是倒序的（最新在前），需要反转为正序（最早在前）
        const records = data.records || data || []
        this.chartData = Array.isArray(records) ? [...records].reverse() : []
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
      if (!this.$refs.chart || this.chartData.length === 0) {
        return
      }

      this.chart = echarts.init(this.$refs.chart)

      // 统一格式化日期为 YYYY-MM-DD
      const dates = this.chartData.map(item => {
        // 兼容两种字段命名方式
        const isManual = item.is_manual || item.isManual
        const contestTime = item.contest_time || item.contestTime
        const createdAt = item.created_at || item.createdAt

        // 手动调整记录使用创建时间
        if (isManual) {
          const date = new Date(createdAt)
          const year = date.getFullYear()
          const month = String(date.getMonth() + 1).padStart(2, '0')
          const day = String(date.getDate()).padStart(2, '0')
          return `${year}-${month}-${day}`
        }
        // 比赛记录使用比赛时间
        if (!contestTime) return '未知日期'
        const date = new Date(contestTime)
        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        return `${year}-${month}-${day}`
      })
      const ratings = this.chartData.map(item => item.new_rating ?? item.newRating)

      // 区分比赛记录、Skip 记录和手动调整记录
      const itemColors = this.chartData.map(item => {
        if (item.is_skip || item.isSkip) return '#F56C6C'  // Skip 用户用红色
        if (item.is_manual || item.isManual) return '#FF4D4F'  // 手动调整用红色
        return '#409EFF'  // 比赛记录用蓝色
      })

      const option = {
        title: {
          text: 'Rating 变化历史',
          left: 'center'
        },
        tooltip: {
          trigger: 'axis',
          formatter: (params) => {
            const data = this.chartData[params[0].dataIndex]

            // 兼容两种字段命名方式
            const isManual = data.is_manual || data.isManual
            const isSkip = data.is_skip || data.isSkip
            const skipReason = data.skip_reason || data.skipReason
            const oldRating = data.old_rating ?? data.oldRating
            const newRating = data.new_rating ?? data.newRating
            // 优先使用 rating_change 字段，如果不存在则从 newRating 和 oldRating 计算
            let ratingChange = data.rating_change ?? data.ratingChange ?? null
            if (ratingChange === null || ratingChange === undefined) {
              if (oldRating !== null && oldRating !== undefined && newRating !== null && newRating !== undefined) {
                ratingChange = newRating - oldRating
              } else {
                ratingChange = 0
              }
            }
            const contestTitle = data.contest_title || data.contestTitle
            const rank = data.rank
            const participants = data.participants
            const createdAt = data.created_at || data.createdAt
            const manualAdjustReason = data.manualAdjustReason || data.manual_adjust_reason
            const manualAdjustDelta = data.manualAdjustDelta ?? data.manual_adjust_delta
            const rankDisplay = rank && participants ? `${rank} / ${participants}` : '未排名'

            // 手动调整记录的 tooltip
            if (isManual) {
              return `
                <div style="text-align: left;">
                  <strong style="color: #FF4D4F;">⚠️ ${data.reason || '手动调整'}</strong><br/>
                  Rating: ${oldRating} → ${newRating}<br/>
                  变化: <span style="color: ${ratingChange > 0 ? '#67C23A' : (ratingChange < 0 ? '#F56C6C' : '#909399')};">${ratingChange > 0 ? '+' : ''}${ratingChange}</span><br/>
                  <span style="color: #999; font-size: 12px;">${new Date(createdAt).toLocaleString('zh-CN')}</span>
                </div>
              `
            }

            // 比赛记录的 tooltip
            const skipInfo = isSkip
              ? `<br/><strong style="color: #F56C6C;">⚠️ Skip: ${skipReason || '该比赛不计入 Rating'}</strong>`
              : ''
            const manualInfo = manualAdjustReason
              ? `<br/><strong style="color: #E6A23C;">📝 个人调整: ${manualAdjustReason}</strong>${
                  manualAdjustDelta !== null && manualAdjustDelta !== undefined
                    ? `<br/><span style="color: #E6A23C;">调整变化: ${manualAdjustDelta > 0 ? '+' : ''}${manualAdjustDelta}</span>`
                    : ''
                }`
              : ''

            return `
              <div style="text-align: left;">
                <strong>${contestTitle || '比赛'}</strong><br/>
                Rating: ${oldRating} → ${newRating}<br/>
                变化: <span style="color: ${ratingChange > 0 ? '#67C23A' : (ratingChange < 0 ? '#F56C6C' : '#909399')};">${ratingChange > 0 ? '+' : ''}${ratingChange}</span><br/>
                排名: ${rankDisplay}${skipInfo}${manualInfo}
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
            color: (params) => {
              return itemColors[params.dataIndex]
            }
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
            ],
            itemStyle: {
              color: '#409EFF'
            }
          },
          // 标记手动调整和 Skip 的点
          markPoint: {
            data: this.chartData
              .map((item, index) => {
                // 兼容两种字段命名方式
                const isSkip = item.is_skip || item.isSkip
                const isManual = item.is_manual || item.isManual
                const skipReason = item.skip_reason || item.skipReason
                const reason = item.reason
                const newRating = item.new_rating ?? item.newRating

                // Skip 用户标记
                if (isSkip) {
                  return {
                    name: skipReason || 'Skip',
                    coord: [index, newRating],
                    itemStyle: {
                      color: '#F56C6C'
                    },
                    label: {
                      show: true,
                      formatter: '⛔',
                      fontSize: 20
                    }
                  }
                }
                // 手动调整标记
                if (isManual) {
                  return {
                    name: reason || '手动调整',
                    coord: [index, newRating],
                    itemStyle: {
                      color: '#FF4D4F'
                    },
                    label: {
                      show: true,
                      formatter: '⚠️',
                      fontSize: 20
                    }
                  }
                }
                return null
              })
              .filter(item => item !== null)
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
