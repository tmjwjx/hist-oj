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
        // 手动调整记录使用创建时间
        if (item.is_manual) {
          const date = new Date(item.created_at)
          const year = date.getFullYear()
          const month = String(date.getMonth() + 1).padStart(2, '0')
          const day = String(date.getDate()).padStart(2, '0')
          return `${year}-${month}-${day}`
        }
        // 比赛记录使用比赛时间
        if (!item.contest_time) return '未知日期'
        const date = new Date(item.contest_time)
        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        return `${year}-${month}-${day}`
      })
      const ratings = this.chartData.map(item => item.new_rating)

      // 区分比赛记录和手动调整记录
      const itemColors = this.chartData.map(item => {
        return item.is_manual ? '#FF4D4F' : '#409EFF' // 手动调整用红色，比赛记录用蓝色
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

            // 手动调整记录的 tooltip
            if (data.is_manual) {
              return `
                <div style="text-align: left;">
                  <strong style="color: #FF4D4F;">⚠️ ${data.reason || '手动调整'}</strong><br/>
                  Rating: ${data.old_rating} → ${data.new_rating}<br/>
                  变化: <span style="color: ${data.rating_change > 0 ? '#67C23A' : '#F56C6C'}">${data.rating_change > 0 ? '+' : ''}${data.rating_change}</span><br/>
                  <span style="color: #999; font-size: 12px;">${new Date(data.created_at).toLocaleString('zh-CN')}</span>
                </div>
              `
            }

            // 比赛记录的 tooltip
            return `
              <div style="text-align: left;">
                <strong>${data.contest_title || '比赛'}</strong><br/>
                Rating: ${data.old_rating} → ${data.new_rating}<br/>
                变化: ${data.rating_change > 0 ? '+' : ''}${data.rating_change}<br/>
                排名: ${data.rank} / ${data.participants}
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
          // 标记手动调整的点
          markPoint: {
            data: this.chartData
              .map((item, index) => {
                if (item.is_manual) {
                  return {
                    name: item.reason || '手动调整',
                    coord: [index, item.new_rating],
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
