<template>
  <span class="rating-badge" :style="{ color: ratingColor }">
    <i v-if="showIcon" class="el-icon-trophy"></i>
    {{ displayText }}
  </span>
</template>

<script>
import { getRatingColor, getRatingTitle, formatRating } from '@/common/rating-utils'

export default {
  name: 'RatingBadge',
  props: {
    rating: {
      type: Number,
      default: null
    },
    showTitle: {
      type: Boolean,
      default: false
    },
    showIcon: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    ratingColor() {
      if (this.rating === null) return '#808080'
      return getRatingColor(this.rating)
    },
    displayText() {
      if (this.rating === null) return 'Unrated'
      if (this.showTitle) {
        return `${getRatingTitle(this.rating)} (${this.rating})`
      }
      return formatRating(this.rating)
    }
  }
}
</script>

<style scoped>
.rating-badge {
  font-weight: bold;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
</style>
