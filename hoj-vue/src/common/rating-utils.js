// Rating 等级配置（参考 Codeforces）
export const RATING_LEVELS = [
  { min: 0, max: 1199, name: 'Newbie', color: '#808080', title: '新手' },
  { min: 1200, max: 1399, name: 'Pupil', color: '#008000', title: '学徒' },
  { min: 1400, max: 1599, name: 'Specialist', color: '#03A89E', title: '专家' },
  { min: 1600, max: 1899, name: 'Expert', color: '#0000FF', title: '大师' },
  { min: 1900, max: 2099, name: 'Candidate Master', color: '#AA00AA', title: '候选大师' },
  { min: 2100, max: 2299, name: 'Master', color: '#FF8C00', title: '宗师' },
  { min: 2300, max: 2399, name: 'International Master', color: '#FF8C00', title: '国际大师' },
  { min: 2400, max: 2599, name: 'Grandmaster', color: '#FF0000', title: '特级大师' },
  { min: 2600, max: 2999, name: 'International Grandmaster', color: '#FF0000', title: '国际特级大师' },
  { min: 3000, max: 9999, name: 'Legendary Grandmaster', color: '#FF0000', title: '传奇大师' }
]

/**
 * 根据 Rating 获取等级信息
 * @param {number} rating - Rating 分数
 * @returns {object} 等级信息对象
 */
export function getRatingLevel(rating) {
  if (rating === null || rating === undefined) {
    return RATING_LEVELS[0]
  }

  for (let level of RATING_LEVELS) {
    if (rating >= level.min && rating <= level.max) {
      return level
    }
  }
  return RATING_LEVELS[0]
}

/**
 * 根据 Rating 获取颜色
 * @param {number} rating - Rating 分数
 * @returns {string} 颜色值
 */
export function getRatingColor(rating) {
  if (rating === null || rating === undefined) {
    return '#808080'
  }
  return getRatingLevel(rating).color
}

/**
 * 根据 Rating 获取称号
 * @param {number} rating - Rating 分数
 * @returns {string} 称号
 */
export function getRatingTitle(rating) {
  if (rating === null || rating === undefined) {
    return '未定级'
  }
  return getRatingLevel(rating).title
}

/**
 * 根据 Rating 获取英文名称
 * @param {number} rating - Rating 分数
 * @returns {string} 英文名称
 */
export function getRatingName(rating) {
  if (rating === null || rating === undefined) {
    return 'Unrated'
  }
  return getRatingLevel(rating).name
}

/**
 * 格式化 Rating 显示
 * @param {number} rating - Rating 分数
 * @returns {string} 格式化后的字符串
 */
export function formatRating(rating) {
  if (rating === null || rating === undefined) {
    return 'Unrated'
  }
  return rating.toString()
}

/**
 * 格式化 Rating 变化
 * @param {number} change - Rating 变化值
 * @returns {string} 格式化后的字符串（带正负号）
 */
export function formatRatingChange(change) {
  if (change === null || change === undefined || change === 0) {
    return '0'
  }
  return change > 0 ? `+${change}` : change.toString()
}
