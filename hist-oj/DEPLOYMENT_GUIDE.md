# 📋 对战系统准备功能 - 部署操作指南

## 🎯 目标
在服务器的 `hoj` 数据库的 `battle_room` 表中添加 `challenger_ready` 字段

## 📝 操作步骤

### 方式 1: 通过 SSH 连接服务器执行

#### 1. 连接到服务器
```bash
ssh root@43.143.133.62
# 输入密码后登录
```

#### 2. 连接到 MySQL 数据库
```bash
mysql -uroot -p'jia13579..' hoj
```

#### 3. 执行以下 SQL 语句
```sql
-- 添加 challenger_ready 字段
ALTER TABLE `battle_room`
ADD COLUMN `challenger_ready` TINYINT(1) NOT NULL DEFAULT 0
COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)'
AFTER `challenger_username`;

-- 添加索引
ALTER TABLE `battle_room`
ADD INDEX `idx_challenger_ready` (`challenger_ready`);
```

#### 4. 验证字段是否添加成功
```sql
-- 查看表结构
DESCRIBE battle_room;

-- 或者只查看新字段
SELECT
    COLUMN_NAME AS '字段名',
    COLUMN_TYPE AS '类型',
    IS_NULLABLE AS '可空',
    COLUMN_DEFAULT AS '默认值',
    COLUMN_COMMENT AS '注释'
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
AND TABLE_NAME = 'battle_room'
AND COLUMN_NAME = 'challenger_ready';
```

#### 5. 退出 MySQL
```sql
exit;
```

#### 6. 退出服务器
```bash
exit;
```

---

### 方式 2: 使用 phpMyAdmin (如果已安装)

1. 打开浏览器访问: `http://43.143.133.62/phpmyadmin` (或对应端口)
2. 登录到 phpMyAdmin
3. 选择 `hoj` 数据库
4. 找到 `battle_room` 表
5. 点击 "SQL" 标签
6. 粘贴以下 SQL 并执行:

```sql
ALTER TABLE `battle_room`
ADD COLUMN `challenger_ready` TINYINT(1) NOT NULL DEFAULT 0
COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)'
AFTER `challenger_username`;

ALTER TABLE `battle_room`
ADD INDEX `idx_challenger_ready` (`challenger_ready`);
```

---

### 方式 3: 使用 MySQL 客户端工具 (Navicat/DataGrip 等)

1. 创建新的数据库连接:
   - 主机: `43.143.133.62`
   - 端口: `3306`
   - 用户名: `root`
   - 密码: `jia13579..`
   - 数据库: `hoj`

2. 执行上述 SQL 语句

---

## ✅ 验证步骤

### 1. 检查字段是否存在

执行以下查询:
```sql
DESCRIBE battle_room;
```

**预期输出** (应该包含以下行):
```
+------------------+--------------+------+-----+---------+----------------+
| Field            | Type         | Null | Key | Default | Extra          |
+------------------+--------------+------+-----+---------+----------------+
| challenger_ready | tinyint(1)    | NO   | MUL | 0       |                |
+------------------+--------------+------+-----+---------+----------------+
```

### 2. 检查现有数据

```sql
-- 查看所有房间
SELECT room_id, host_username, challenger_username, challenger_ready, status
FROM battle_room
ORDER BY gmt_create DESC
LIMIT 10;
```

### 3. 测试新字段

```sql
-- 测试更新准备状态
UPDATE battle_room
SET challenger_ready = 1
WHERE challenger_id IS NOT NULL
LIMIT 1;

-- 查看更新结果
SELECT room_id, challenger_username, challenger_ready
FROM battle_room
WHERE challenger_ready = 1;

-- 恢复测试数据
UPDATE battle_room
SET challenger_ready = 0
WHERE challenger_id IS NOT NULL;
```

---

## 🔍 常见问题排查

### 问题 1: 字段已存在
**错误信息**: `Duplicate column name 'challenger_ready'`

**原因**: 字段已经添加过了

**解决**: 跳过此步骤,继续部署后端

---

### 问题 2: 权限不足
**错误信息**: `ERROR 1227 (42000): Access denied; you need (at least one of) the SUPER privilege(s)`

**解决**: 使用有足够权限的账号 (root)

---

### 问题 3: 表不存在
**错误信息**: `Table 'hoj.battle_room' doesn't exist`

**解决**: 检查数据库名称是否正确

---

## 📊 迁移后的数据影响

### 现有数据
- ✅ 所有现有记录的 `challenger_ready` 自动设置为 `0` (未准备)
- ✅ 不影响现有功能
- ✅ 向后兼容

### 新功能
- 挑战者加入房间后,需要手动点击"准备"
- 房主看到挑战者准备后才能开始对战
- 退出房间或重置房间会清除准备状态

---

## 🎉 完成后的后续步骤

1. ✅ 数据库迁移完成
2. ⏭️ 重启后端服务
   ```bash
   cd /path/to/hist-oj
   make run
   # 或 Docker 环境
   make docker-build && make docker-run
   ```
3. ⏭️ 清除浏览器缓存 (Ctrl+Shift+R)
4. ⏭️ 测试准备功能

---

## 📞 需要帮助?

如果遇到问题,请提供以下信息:
1. 执行的 SQL 语句
2. 错误信息截图
3. MySQL 版本: `SELECT VERSION();`
4. 当前表结构: `DESCRIBE battle_room;`

---

## 🔄 回滚方案 (如果需要)

如果发现问题,可以回滚:

```sql
-- 删除索引
ALTER TABLE `battle_room` DROP INDEX `idx_challenger_ready`;

-- 删除字段
ALTER TABLE `battle_room` DROP COLUMN `challenger_ready`;
```

然后回滚代码即可。

---

**部署完成后,请执行验证步骤并告知结果!** ✅
