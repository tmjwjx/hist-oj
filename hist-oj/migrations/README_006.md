# 数据库迁移指南 - 添加挑战者准备状态

## 📋 迁移信息
- **版本**: 006
- **日期**: 2025-01-10
- **功能**: 为 battle_room 表添加 challenger_ready 字段,实现挑战者准备机制

## 🎯 功能说明
添加挑战者准备功能,确保对战公平性:
- 挑战者必须点击"准备"后,房主才能开始对战
- 解决"再来一局"后挑战者被强制拉回但没有防备的问题

## 📝 数据库变更

### SQL 脚本
```sql
-- 添加 challenger_ready 字段
ALTER TABLE `battle_room`
ADD COLUMN `challenger_ready` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)' AFTER `challenger_username`;

-- 添加索引以提高查询性能
ALTER TABLE `battle_room`
ADD INDEX `idx_challenger_ready` (`challenger_ready`);
```

### 执行方式
在服务器上执行以下命令:

```bash
mysql -uroot -p'jia13579..' hoj < migrations/006_add_challenger_ready.sql
```

或者直接执行 SQL:
```bash
mysql -uroot -p'jia13579..' hoj -e "
ALTER TABLE \`battle_room\`
ADD COLUMN \`challenger_ready\` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '挑战者是否已准备' AFTER \`challenger_username\`;

ALTER TABLE \`battle_room\`
ADD INDEX \`idx_challenger_ready\` (\`challenger_ready\`);
"
```

## ✅ 兼容性检查

### 已检查的功能点
✅ **加入房间 (JoinRoom)**: 新加入的挑战者默认 `challengerReady = false`
✅ **退出房间 (LeaveRoom)**: 挑战者退出时清除准备状态
✅ **房主退出**: 转移房主身份时清除准备状态
✅ **重置房间 (ResetRoom)**: 重置时清除准备状态
✅ **开始对战 (StartBattle)**: 校验挑战者准备状态
✅ **准备对战 (ReadyBattle)**: 新增功能,更新准备状态

### GORM Save 操作
所有 `Save(&room)` 操作都会正确处理 `ChallengerReady` 字段:
- 使用 `room.ChallengerReady = false` 明确设置
- 数据库默认值为 0 (false)
- 不会意外覆盖其他字段

## 🔍 影响范围分析

### 后端 (hist-oj)
1. **Model 层**: ✅ 添加 `ChallengerReady` 字段
2. **Service 层**: ✅ 5 个方法更新
3. **API 层**: ✅ 新增 `ReadyBattle` 接口
4. **路由**: ✅ 注册 `/api/battle/ready`

### 前端 (hoj-vue)
1. **API 调用**: ✅ 新增 `readyBattle` 方法
2. **UI 组件**: ✅ 准备按钮和状态显示
3. **数据字段**: ✅ `room.challengerReady` 绑定

### 数据库
- ✅ 添加字段不影响现有数据
- ✅ 默认值为 false,向后兼容
- ✅ 添加索引优化查询

## 🚀 部署步骤

### 1. 备份数据库 (推荐)
```bash
mysqldump -uroot -p'jia13579..' hoj > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. 执行迁移
```bash
mysql -uroot -p'jia13579..' hoj < migrations/006_add_challenger_ready.sql
```

### 3. 验证字段
```bash
mysql -uroot -p'jia13579..' hoj -e "DESCRIBE battle_room;" | grep challenger_ready
```

预期输出:
```
challenger_ready tinyint(1)  NO      0
```

### 4. 重启后端服务
```bash
cd hist-oj
make run
# 或
make docker-build && make docker-run
```

### 5. 前端无需重启
前端代码会自动生效,如有缓存可刷新浏览器。

## 🧪 测试验证

### 场景 1: 挑战者加入房间
- [ ] 挑战者加入后显示"未准备"状态
- [ ] 准备按钮显示"准备"
- [ ] 房主看到挑战者"未准备"标签

### 场景 2: 挑战者准备
- [ ] 点击"准备"按钮
- [ ] 按钮变为"取消准备"
- [ ] 显示"已准备,等待房主开始对战"
- [ ] 房主看到挑战者"已准备"标签
- [ ] 房主"开始对战"按钮可点击

### 场景 3: 挑战者取消准备
- [ ] 点击"取消准备"按钮
- [ ] 按钮变回"准备"
- [ ] 准备提示消失
- [ ] 房主看到挑战者"未准备"标签
- [ ] 房主"开始对战"按钮禁用

### 场景 4: 房主开始对战
- [ ] 挑战者未准备时,点击开始显示错误提示
- [ ] 挑战者准备后,点击开始成功

### 场景 5: 退出房间
- [ ] 挑战者退出,准备状态清除
- [ ] 新挑战者加入,准备状态重置为未准备

### 场景 6: 再来一局
- [ ] 点击"再来一局",准备状态清除
- [ ] 挑战者需要重新准备

## 🔄 回滚方案

如果需要回滚,执行以下 SQL:
```sql
ALTER TABLE `battle_room` DROP INDEX `idx_challenger_ready`;
ALTER TABLE `battle_room` DROP COLUMN `challenger_ready`;
```

然后回滚代码到迁移前的版本。

## 📊 性能影响

- **索引**: 添加 `idx_challenger_ready` 索引,查询性能无影响
- **存储**: 每行增加 1 字节 (TINYINT)
- **查询**: 准备状态校验增加约 0.1ms

## ✅ 部署检查清单

- [ ] 数据库备份完成
- [ ] SQL 迁移脚本执行成功
- [ ] 字段验证通过
- [ ] 后端服务重启成功
- [ ] 准备功能测试通过
- [ ] 退出房间测试通过
- [ ] 再来一局测试通过
- [ ] 开始对战测试通过

## 📞 问题排查

### 问题 1: 字段不存在
**错误**: `Error 1054: Unknown column 'challenger_ready'`
**解决**: 执行数据库迁移脚本

### 问题 2: 准备状态不更新
**原因**: 前端缓存或后端未重启
**解决**:
1. 重启后端服务
2. 清除浏览器缓存 (Ctrl+Shift+R)

### 问题 3: 准备后仍无法开始对战
**原因**: 后端代码未更新
**解决**: 确保后端代码包含 `StartBattle` 的准备状态校验

## 📝 相关文件

### 后端
- `hist-oj/internal/model/battle.go` - Model 定义
- `hist-oj/internal/service/battle_service.go` - 业务逻辑
- `hist-oj/internal/api/battle_api.go` - API 接口
- `hist-oj/migrations/006_add_challenger_ready.sql` - 数据库迁移

### 前端
- `hoj-vue/src/api/battle.js` - API 调用
- `hoj-vue/src/views/oj/battle/BattleRoom.vue` - 房间组件

## 🎉 完成

迁移完成后,你的对战系统将具备以下功能:
- ✅ 挑战者准备机制
- ✅ 防止突然开始对战
- ✅ 更好的游戏体验
- ✅ 公平的对战环境
