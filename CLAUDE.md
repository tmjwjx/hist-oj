# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

HOJ (Hcode Online Judge) 是一个基于 Vue 和 Spring Boot、Spring Cloud Alibaba 构建的前后端分离、分布式架构的在线判题系统。

### 项目背景

**本项目是基于开源项目 HOJ 改写的定制版本**：
- **后端（hoj-springboot）**：大部分直接复用原 HOJ 项目的代码
- **前端（hoj-vue）**：已经重写和更改，与原版有较大差异
- **新增模块（hist-oj）**：为了扩展新功能而新建的独立 Go 服务

### 当前开发状态

**hist-oj 模块（Rating 系统）**：
- **目标**：实现 Rating 计算功能，并根据不同的 Rating 分数让用户显示不同的称号/等级
- **后端状态**：已有雏形，基于 Elo 算法的 Rating 计算逻辑已实现
- **前端状态**：尚未开始改动，需要在 hoj-vue 中集成 Rating 显示功能
- **下一步**：需要在前端实现用户 Rating 显示、称号系统、Rating 历史图表等功能

### 本地调试方案

**当前调试策略**：
- **前端（hoj-vue）**：本地运行，连接线上后端（`http://43.143.133.62:6688`） `npm run serve`
- **hist-oj 服务**：本地运行（默认端口 9527），需要配置本地数据库连接 `make run`
- **hoj-springboot**：暂不本地运行，直接使用线上服务

**调试目标**：
1. 修改前端代码，添加 Rating 显示相关的 UI 组件和页面
2. 修改 hist-oj 代码，完善 Rating 计算和查询接口
3. 打通前端 → hist-oj 的数据流，实现 Rating 功能的端到端展示

**代理配置**：
- 前端已配置 hist-oj 服务代理：`/rating-api` → `http://localhost:9527/api`
- 配置位置：[vue.config.js](hoj-vue/vue.config.js#L100-L106)
- 前端调用示例：`axios.get('/rating-api/user/rating')` 会被代理到 `http://localhost:9527/api/user/rating`

**技术栈**：
- 后端：Java 8, Spring Boot 2.2.6, Spring Cloud Alibaba 2.2.1
- 前端：Vue 2.6.11, Element UI 2.15.3
- 数据库：MySQL 8.0.19, Redis 5.0.9
- 服务注册：Nacos 1.4.2
- Rating 服务：Go 1.21

## 常用命令

### 前端开发 (hoj-vue)

```bash
# 安装依赖
cd hoj-vue
npm install

# 启动开发服务器（端口 8066）
npm run serve

# 构建生产版本
npm run build
```

**注意**：开发服务器会将 `/api` 请求代理到 `http://43.143.133.62:6688`（配置在 [vue.config.js](hoj-vue/vue.config.js#L97)）

### 后端开发 (hoj-springboot)

```bash
# 在项目根目录构建所有模块
cd hoj-springboot
mvn clean package -DskipTests

# 构建单个模块
cd hoj-springboot/api
mvn clean package -DskipTests

# 运行测试
mvn test
```

**模块说明**：
- `api/`：主要 API 服务
- `JudgeServer/`：判题服务器
- `DataBackup/`：数据备份服务

### Rating 服务开发 (hist-oj)

```bash
cd hist-oj

# 安装依赖
make deps

# 本地运行
make run

# 构建二进制文件
make build

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# Docker 部署
make docker-build
make docker-run

# 查看日志
make docker-logs

# 停止服务
make docker-stop

# 重置 Rating 数据
make reset-rating

# 查看服务状态
make status
```

**hist-oj 配置文件**：
- 配置文件位于 `hist-oj/configs/config.yaml`
- 需要配置数据库连接信息（MySQL）
- 默认服务端口：9527

## 项目架构

### 整体结构

```
hoj2/
├── hoj-springboot/      # 后端服务（多模块 Maven 项目）
│   ├── api/             # API 服务模块
│   ├── JudgeServer/     # 判题服务器模块
│   └── DataBackup/      # 数据备份模块
├── hoj-vue/             # 前端项目（Vue 2）
├── hist-oj/             # Rating 计算服务（Go）
├── hoj-scrollBoard/     # 滚动排行榜（静态页面）
├── sandbox/             # 判题沙箱（amd64/arm64）
└── sqlAndsetting/       # SQL 脚本和配置
```

### 后端架构 (hoj-springboot)

**微服务架构**：
- 使用 Spring Cloud Alibaba 构建分布式系统
- Nacos 作为服务注册中心和配置中心
- 三个独立的 Spring Boot 模块：api、JudgeServer、DataBackup

**核心依赖**：
- Spring Boot 2.2.6.RELEASE
- Spring Cloud Hoxton.SR1
- Spring Cloud Alibaba 2.2.1.RELEASE
- MyBatis Plus 3.2.0
- Shiro Redis 3.2.1
- Hutool 5.8.8

**包结构**：所有 Java 代码位于 `top.hcode.hoj` 包下

### 前端架构 (hoj-vue)

**目录结构**：
- `src/components/`：可复用组件
  - `admin/`：管理后台组件
  - `oj/`：OJ 前端组件
- `src/views/`：页面视图
  - `admin/`：管理后台页面
  - `oj/`：OJ 前端页面
- `src/router/`：路由配置
- `src/store/`：Vuex 状态管理
- `src/common/`：公共工具（API 封装、工具函数等）
- `src/i18n/`：国际化（支持中文、英文、日文、韩文）

**核心依赖**：
- Vue 2.6.11
- Element UI 2.15.3
- Vue Router 3.2.0
- Vuex 3.4.0
- Axios 0.21.0
- Mavon Editor 2.9.1（Markdown 编辑器）
- Highlight.js 10.3.2（代码高亮）
- ECharts 4.9.0（图表）

### Rating 服务架构 (hist-oj)

**功能**：基于 Elo 算法的比赛 Rating 计算服务

**目录结构**：
- `cmd/server/`：服务入口
- `internal/api/`：API 处理器和路由
- `internal/service/`：业务逻辑（Rating 计算、查询）
- `internal/model/`：数据模型
- `internal/client/`：客户端（数据库、HOJ API）
- `internal/schedule/`：定时任务调度器
- `migrations/`：数据库迁移脚本
- `configs/`：配置文件

**核心依赖**：
- Gin v1.9.1（Web 框架）
- GORM v1.25.5（ORM）
- Viper v1.18.2（配置管理）
- Zap v1.26.0（日志）
- Cron v3.0.1（定时任务）

## 核心功能模块

### 1. 评测系统
- 支持多种评测模式：普通测评、特殊测评（SPJ）、交互测评、子任务分组评测、文件 IO
- 支持 Remote Judge：HDU、POJ、Codeforces、AtCoder、SPOJ、LIBRE
- 支持 13 种编程语言：C、C++、C#、Python、PyPy、Go、Java、JavaScript、PHP、Ruby、Rust

### 2. 比赛系统
- 支持 ACM 和 OI 两种赛制
- 打星队伍、关注队伍、外榜、滚榜功能
- 比赛打印、允许比赛结束后提交

### 3. 训练系统
- 私有训练和公开训练（题单）
- 自定义难度

### 4. 团队系统
- 团队创建与管理
- 团队题目、团队比赛

### 5. 讨论系统
- 公共讨论区、题目讨论区、比赛评论
- 评论回复、点赞、举报功能

### 6. 消息系统
- 站内消息、评论通知、回复通知、点赞通知、系统通知

## 开发注意事项

### 后端开发

1. **Java 版本**：必须使用 Java 8
2. **编码**：所有文件使用 UTF-8 编码
3. **测试**：Maven 配置中默认跳过测试（`skipTests=true`）
4. **配置文件**：各模块的配置文件位于 `src/main/resources/application.yml`

### 前端开发

1. **代理配置**：开发环境的 API 代理配置在 [vue.config.js](hoj-vue/vue.config.js#L95-L100)
2. **CDN 配置**：生产环境可选择使用 CDN 加载依赖（当前 `isProduction=false`）
3. **构建优化**：已配置 UglifyJS 和 Gzip 压缩
4. **国际化**：修改文本时需同步更新 `src/i18n/` 下的所有语言文件

### Rating 服务开发

1. **Go 版本**：使用 Go 1.21
2. **配置文件**：配置文件位于 `configs/config.yaml`
3. **数据库迁移**：迁移脚本位于 `migrations/` 目录
4. **定时任务**：比赛结束后自动触发 Rating 计算

## 部署相关

### 系统要求
- 操作系统：CentOS 8+ 或 Ubuntu 16.04+
- 服务器配置：建议 2 核 4G 以上
- 避免使用突发性能或共享型云服务器实例

### 部署方式
- 推荐使用 Docker Compose 一键部署
- 部署文档：https://docs.hdoi.cn/deploy/docker
- 部署仓库：https://gitee.com/himitzh0730/hoj-deploy

### 更新方式
```bash
# 在 docker-compose.yml 所在目录执行
docker-compose pull
docker-compose up -d
```

## 数据库

### 初始化脚本
- `sqlAndsetting/hoj.sql`：数据库初始化脚本
- `sqlAndsetting/hoj-update.sql`：数据库更新脚本
- `sqlAndsetting/nacos.sql`：Nacos 配置脚本

## 文档资源

- **在线文档**：https://docs.hdoi.cn
- **在线 Demo**：https://hdoi.cn
- **GitHub 仓库**：https://github.com/HimitZH/HOJ
- **Gitee 仓库**：https://gitee.com/himitzh0730/hoj

## 项目规范（来自 .cursor/rules/base.md）

### 设计文档规范

**需要创建设计文档的场景**（在 `.cursor/md/` 目录）：
- 新功能开发（涉及 3 个以上文件或模块）
- 架构调整或重构
- 复杂业务逻辑实现（需要多步骤协调）
- 涉及数据库 Schema 变更
- 第三方服务集成
- 性能优化方案
- 安全相关改动

**无需创建文档的场景**：
- 简单 Bug 修复（1-2 个文件）
- 代码格式化、注释补充
- 配置项调整
- 简单的增删改查功能
- 日志添加、错误处理优化
- 单一函数或方法的小改动

**设计文档结构**：
- 问题/需求描述
- 解决方案
- 实现步骤（复杂场景必需）
- 注意事项（可选）

**确认流程**：
- 复杂需求：设计文档必须完成并经用户确认后，才能进行代码修改
- 简单需求：直接实现，无需文档和确认流程
