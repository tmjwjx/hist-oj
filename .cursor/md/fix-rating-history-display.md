# Rating 历史显示问题修复方案

## 问题描述

个人主页的 Rating 历史图表不显示，经过分析发现以下问题：

1. **字段名不一致**：代码使用 `profile.uid`，但 HOJ 后端返回的是 `profile.uuid`
2. **参数传递不明确**：RatingChart 组件可能接收到 username 而不是 uuid，导致 API 查询失败
3. **条件判断冗余**：RatingChart 组件上有重复的 v-if 判断
4. **错误处理不完善**：当获取 rating 失败时，用户看不到具体原因

## 解决方案

### 1. 统一用户标识获取逻辑

在 UserHome.vue 中添加一个计算属性，统一处理用户标识的获取：

```javascript
computed: {
  // 获取用户的唯一标识（优先使用 uuid）
  userIdentifier() {
    return this.profile.uuid || this.profile.uid || this.$route.query.uid || this.$route.query.username;
  },
  // 其他计算属性...
}
```

### 2. 修改 RatingChart 组件的调用

简化 RatingChart 的调用，移除冗余的 v-if：

```vue
<!-- 修改前 -->
<RatingChart :uid="profile.uid || $route.query.uid || $route.query.username"
             v-if="profile.uid || $route.query.uid || $route.query.username" />

<!-- 修改后 -->
<RatingChart :uid="userIdentifier" />
```

### 3. 优化显示逻辑

将 Rating 历史卡片的显示逻辑改为：
- 如果正在加载 rating 数据，显示加载中
- 如果加载失败或用户没有 rating，不显示 Rating 历史卡片
- 如果加载成功且有 rating，显示 Rating 历史卡片（即使没有历史记录，也让 RatingChart 组件自己处理）

```vue
<!-- 修改前 -->
<el-card style="margin-top:1rem;" v-if="userRating !== null">
  <div class="card-title">
    <i class="el-icon-data-line" style="color:#409eff"></i>
    Rating 变化历史
  </div>
  <RatingChart :uid="profile.uid || $route.query.uid || $route.query.username"
               v-if="profile.uid || $route.query.uid || $route.query.username" />
</el-card>
<el-card style="margin-top:1rem;" v-else>
  <div style="text-align: center; padding: 20px; color: #999;">
    {{ userRating === null ? '加载 Rating 数据中...' : '暂无 Rating 数据' }}
  </div>
</el-card>

<!-- 修改后 -->
<el-card style="margin-top:1rem;" v-if="ratingLoading">
  <div style="text-align: center; padding: 20px; color: #999;">
    <i class="el-icon-loading"></i> 加载 Rating 数据中...
  </div>
</el-card>
<el-card style="margin-top:1rem;" v-else-if="userRating !== null && userIdentifier">
  <div class="card-title">
    <i class="el-icon-data-line" style="color:#409eff"></i>
    Rating 变化历史
  </div>
  <RatingChart :uid="userIdentifier" />
</el-card>
```

### 4. 添加加载状态

在 data 中添加 ratingLoading 状态：

```javascript
data() {
  return {
    // ... 其他数据
    userRating: null,
    maxRating: null,
    contestCount: 0,
    ratingLoading: false,  // 新增：rating 加载状态
  };
}
```

### 5. 修改 fetchUserRating 方法

```javascript
async fetchUserRating(uid) {
  if (!uid) {
    console.warn('fetchUserRating: uid 为空，跳过获取 rating');
    return;
  }

  console.log('fetchUserRating called with uid:', uid);
  this.ratingLoading = true;

  try {
    const data = await ratingApi.getUserRating(uid);
    console.log('getUserRating response:', data);
    this.userRating = data.rating;
    this.maxRating = data.maxRating;
    console.log('userRating set to:', this.userRating);

    // 获取参赛次数
    const historyData = await ratingApi.getRatingHistory(uid, 1, 1);
    console.log('getRatingHistory response:', historyData);
    this.contestCount = historyData.total || 0;
  } catch (error) {
    console.error('获取用户 Rating 失败:', error);
    this.userRating = null;
    this.maxRating = null;
    this.contestCount = 0;
  } finally {
    this.ratingLoading = false;
  }
}
```

### 6. 修改 init 方法中的调用

```javascript
init() {
  const uid = this.$route.query.uid;
  const username = this.$route.query.username;
  this.loading = true;

  api.getUserInfo(uid, username).then((res) => {
    const userData = res.data.data || res.data;
    this.changeDomTitle({ title: userData.username });
    this.profile = userData;
    this.$nextTick((_) => {
      addCodeBtn();
    });
    this.loading = false;

    // 修改：使用计算属性 userIdentifier
    if (this.userIdentifier) {
      this.fetchUserRating(this.userIdentifier);
    } else {
      console.warn('无法获取用户标识，跳过 rating 数据加载');
    }
  }, (_) => {
    this.loading = false;
  });
}
```

### 7. RatingChart 组件优化（可选）

在 RatingChart 组件中添加更友好的错误提示：

```vue
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
```

## 实现步骤

1. 在 UserHome.vue 的 computed 中添加 `userIdentifier` 计算属性
2. 在 data 中添加 `ratingLoading: false`
3. 修改 `fetchUserRating` 方法，添加参数校验和 loading 状态管理
4. 修改 `init` 方法，使用 `userIdentifier` 计算属性
5. 修改模板中的 Rating 历史卡片显示逻辑
6. （可选）优化 RatingChart 组件的空状态提示

## 注意事项

1. **向后兼容**：同时支持 `uuid` 和 `uid` 字段，确保与不同版本的后端兼容
2. **参数优先级**：`profile.uuid` > `profile.uid` > `$route.query.uid` > `$route.query.username`
3. **错误处理**：所有异步操作都有 try-catch，确保不会因为 rating 服务异常影响页面其他功能
4. **用户体验**：明确区分"加载中"、"加载失败"、"无数据"三种状态
5. **调试信息**：保留 console.log，方便后续排查问题

## 测试场景

修改完成后，需要测试以下场景：

1. ✅ 用户有 rating 且有历史记录 → 显示图表
2. ✅ 用户有 rating 但无历史记录 → 显示"暂无 Rating 历史记录"
3. ✅ 用户没有 rating → 不显示 Rating 历史卡片
4. ✅ rating 服务异常 → 不显示 Rating 历史卡片，不影响页面其他功能
5. ✅ 通过不同 URL 参数访问（uid、username）→ 都能正确加载
6. ✅ profile 返回 uuid 或 uid 字段 → 都能正确识别
