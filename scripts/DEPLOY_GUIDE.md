# 部署脚本使用指南

## 快速开始

### 完整部署（首次部署或重大更新）
```bash
./scripts/deploy.sh
```

### 快速部署（跳过镜像构建）
```bash
./scripts/deploy.sh --skip-build
```

## 高级选项

### 1. 跳过数据库迁移（已部署环境）
对于已经部署过的系统，数据库字段已存在，可以跳过迁移检查：
```bash
./scripts/deploy.sh --skip-db
```

**注意：** 实际上脚本会自动检测字段是否存在，只有在字段不存在时才执行迁移。所以这个选项主要用于完全确定不需要检查时。

### 2. 部署特定服务

#### 仅更新前端
```bash
./scripts/deploy.sh --only-frontend
```

#### 仅更新后端服务（hist-oj Rating 系统）
```bash
./scripts/deploy.sh --only-backend
```

#### 仅更新报名系统
```bash
./scripts/deploy.sh --only-registration
```

### 3. 组合选项

#### 快速更新前端（不构建镜像）
```bash
./scripts/deploy.sh --skip-build --only-frontend
```

#### 快速更新报名系统
```bash
./scripts/deploy.sh --skip-build --only-registration
```

### 4. 查看帮助
```bash
./scripts/deploy.sh -h
# 或
./scripts/deploy.sh --help
```

## 部署流程说明

### 首次部署
```bash
./scripts/deploy.sh
```

**执行步骤：**
1. ✅ 检查必要命令（docker, sshpass）
2. ✅ 构建 Docker 镜像（hist-oj, registration-backend, hoj-frontend）
3. ✅ 保存镜像为 tar.gz 文件
4. ✅ 上传镜像和数据库迁移脚本到服务器
5. ✅ 在服务器上：
   - 检查数据库字段是否存在
   - 如果不存在，执行数据库迁移
   - 停止旧容器
   - 启动新容器
6. ✅ 清理临时文件
7. ✅ 验证部署结果

### 后续更新（推荐）
```bash
./scripts/deploy.sh --skip-build
```

**执行步骤：**
1. ✅ 检查必要命令
2. ⏭️  **跳过**镜像构建（使用本地已有镜像）
3. ✅ 保存镜像
4. ✅ 上传镜像（不上传迁移脚本）
5. ✅ 在服务器上：
   - 检查数据库字段（已存在，跳过迁移）
   - 重启容器
6. ✅ 清理临时文件
7. ✅ 验证部署结果

### 部分更新
```bash
./scripts/deploy.sh --skip-build --only-frontend
```

**执行步骤：**
1. ✅ 检查必要命令
2. ⏭️  **跳过**镜像构建
3. ✅ 仅保存前端镜像
4. ✅ 仅上传前端镜像
5. ✅ 在服务器上：
   - **跳过**数据库迁移
   - 仅重启前端容器
6. ✅ 清理临时文件

## 优化点总结

### 1. **智能数据库迁移**
- 自动检测 `last_view_time` 和 `admin_last_view_time` 字段是否存在
- 只在字段不存在时执行迁移
- 避免重复执行迁移脚本

### 2. **部分部署支持**
- 支持单独部署前端、后端或报名系统
- 节省时间和带宽

### 3. **镜像构建优化**
- `--skip-build` 选项跳过耗时的镜像构建步骤
- 适用于代码无变化，仅配置更新的场景

### 4. **选择性上传**
- 部分部署时只上传需要的镜像
- 减少网络传输时间

## 常见场景

### 场景 1：修改了前端代码
```bash
./scripts/deploy.sh --skip-build --only-frontend
```

### 场景 2：修改了报名系统代码
```bash
./scripts/deploy.sh --skip-build --only-registration
```

### 场景 3：修改了 hist-oj 代码
```bash
./scripts/deploy.sh --skip-build --only-backend
```

### 场景 4：添加了新功能（需要数据库迁移）
```bash
./scripts/deploy.sh
```

### 场景 5：仅配置更新（无代码变化）
```bash
./scripts/deploy.sh --skip-build
```

## 验证部署

部署完成后，访问以下地址验证：

- 前端: http://43.143.133.62
- 报名页面: http://43.143.133.62/registration
- 管理后台: http://43.143.133.62/admin/registration
- hist-oj API: http://43.143.133.62:9527

### 验证数据库字段
```bash
mysql -h43.143.133.62 -uroot -phist2025 -e \
  'SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS \
   WHERE TABLE_NAME="histcontest_register_registrations" \
   AND COLUMN_NAME IN ("last_view_time", "admin_last_view_time")' hoj
```

预期输出：
```
COLUMN_NAME
last_view_time
admin_last_view_time
```

## 故障排查

### 1. 镜像构建失败
```bash
# 手动构建单个镜像
cd hist-oj
docker build --platform linux/amd64 -t hist-oj:latest .
```

### 2. 上传失败
```bash
# 检查服务器连接
ping 43.143.133.62

# 检查 sshpass 是否安装
brew install sshpass
```

### 3. 容器启动失败
```bash
# 登录服务器查看日志
ssh root@43.143.133.62
docker logs hist-oj
docker logs registration-backend
docker logs hoj-frontend
```

### 4. 数据库迁移失败
```bash
# 手动执行迁移
mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/001_add_last_view_time_fields.sql
```
