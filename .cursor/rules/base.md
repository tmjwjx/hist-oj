# HOJ 项目结构和大纲

## 项目概述

HOJ (Hcode Online Judge) 是一个基于 Vue 和 Spring Boot、Spring Cloud Alibaba 构建的前后端分离、分布式架构的在线判题系统。

### 技术栈

- **后端**: Java 1.8, Spring Boot 2.2.6.RELEASE, Spring Cloud Alibaba 2.2.1.RELEASE
- **前端**: Vue 2.6.11, Element UI 2.15.3
- **数据库**: MySQL 8.0.19
- **缓存**: Redis 5.0.9
- **服务注册**: Nacos 1.4.2
- **判题沙箱**: Sandbox (amd64/arm64)
- **Rating服务**: Go (hist-oj)

## 项目结构

```
hoj2/
├── hoj-springboot/          # 后端服务（Spring Boot）
│   ├── api/                 # API服务模块
│   │   └── src/main/java/top/hcode/hoj/api/
│   ├── DataBackup/          # 数据备份模块
│   │   └── src/main/java/top/hcode/hoj/
│   ├── JudgeServer/         # 判题服务器模块
│   │   └── src/main/java/top/hcode/hoj/judge/
│   └── pom.xml              # Maven父POM
│
├── hoj-vue/                 # 前端项目（Vue.js）
│   ├── src/
│   │   ├── components/      # 组件
│   │   │   ├── admin/       # 管理后台组件
│   │   │   └── oj/          # OJ前端组件
│   │   ├── views/           # 页面视图
│   │   │   ├── admin/       # 管理后台页面
│   │   │   └── oj/          # OJ前端页面
│   │   ├── router/          # 路由配置
│   │   ├── store/           # Vuex状态管理
│   │   ├── common/          # 公共工具
│   │   └── i18n/            # 国际化
│   ├── public/              # 静态资源
│   └── package.json
│
├── hist-oj/                 # Rating计算服务（Go）
│   ├── cmd/server/          # 服务入口
│   ├── internal/
│   │   ├── api/             # API处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── model/           # 数据模型
│   │   ├── client/          # 客户端（数据库、HOJ API）
│   │   ├── config/          # 配置
│   │   ├── middleware/      # 中间件
│   │   ├── schedule/        # 定时任务
│   │   └── utils/           # 工具类
│   ├── migrations/          # 数据库迁移
│   ├── configs/            # 配置文件
│   └── go.mod
│
├── hoj-scrollBoard/         # 滚动排行榜（静态页面）
│   ├── css/                 # 样式文件
│   ├── js/                  # JavaScript文件
│   └── index.html
│
├── docs/                    # 项目文档
│   └── docs/
│       ├── deploy/          # 部署文档
│       ├── develop/         # 开发文档
│       ├── introducition/   # 介绍文档
│       ├── monomer/         # 单体部署文档
│       └── use/             # 使用文档
│
├── sandbox/                 # 判题沙箱
│   ├── Sandbox-amd64-v1.8.0
│   └── Sandbox-arm64-v1.8.0
│
└── sqlAndsetting/          # SQL脚本和配置
    ├── hoj.sql             # 数据库初始化脚本
    ├── hoj-update.sql      # 数据库更新脚本
    └── nacos.sql           # Nacos配置脚本
```

## 核心模块说明

### 1. hoj-springboot (后端服务)

#### 1.1 api 模块
- **功能**: 提供主要的API接口服务
- **技术**: Spring Boot, MyBatis Plus, Shiro
- **主要功能**:
  - 用户认证与授权
  - 题目管理
  - 比赛管理
  - 提交记录管理
  - 讨论区管理
  - 团队管理
  - 训练管理

#### 1.2 JudgeServer 模块
- **功能**: 判题服务器，负责代码评测
- **主要功能**:
  - 普通评测
  - 特殊评测（SPJ）
  - 交互评测
  - 在线自测
  - 子任务分组评测
  - 文件IO评测
  - Remote Judge（HDU、POJ、Codeforces、AtCoder、SPOJ、LIBRE）

#### 1.3 DataBackup 模块
- **功能**: 数据备份服务
- **主要功能**: 定期备份数据库数据

### 2. hoj-vue (前端项目)

#### 2.1 目录结构
- `src/components/`: 可复用组件
  - `admin/`: 管理后台组件
  - `oj/`: OJ前端组件
- `src/views/`: 页面视图
  - `admin/`: 管理后台页面（20个页面）
  - `oj/`: OJ前端页面（49个页面）
- `src/router/`: 路由配置
  - `adminRoutes.js`: 管理后台路由
  - `ojRoutes.js`: OJ前端路由
- `src/store/`: Vuex状态管理
  - `user.js`: 用户状态
  - `contest.js`: 比赛状态
  - `group.js`: 团队状态
  - `training.js`: 训练状态
- `src/common/`: 公共工具
  - `api.js`: API请求封装
  - `utils.js`: 工具函数
  - `message.js`: 消息提示
  - `storage.js`: 本地存储
  - `time.js`: 时间处理
  - `codeblock.js`: 代码块处理
  - `highlight.js`: 代码高亮
  - `katex.js`: 数学公式渲染

#### 2.2 国际化
- 支持语言: 中文简体、中文繁体、英文、日文、韩文
- 路径: `src/i18n/`

### 3. hist-oj (Rating计算服务)

#### 3.1 功能
- 基于 Elo 算法的 Rating 计算
- 新手保护（前3场掉分减半）
- 参与激励（过题+5分）
- 比赛结束自动计算
- RESTful API

#### 3.2 主要模块
- `internal/api/`: API处理器和路由
- `internal/service/`: 业务逻辑（Rating计算、查询）
- `internal/model/`: 数据模型
- `internal/client/`: 客户端（数据库、HOJ API）
- `internal/schedule/`: 定时任务调度器

### 4. hoj-scrollBoard (滚动排行榜)

- **功能**: 比赛滚榜展示页面
- **技术**: Bootstrap, jQuery
- **用途**: 比赛结束后展示滚榜动画

## 核心功能

### 1. 评测功能
- ✅ 普通测评
- ✅ 特殊测评（SPJ）
- ✅ 交互测评
- ✅ 在线自测
- ✅ 子任务分组评测
- ✅ 文件IO评测
- ✅ Remote Judge（HDU、POJ、Codeforces、AtCoder、SPOJ、LIBRE）

### 2. 支持的编程语言
C、C++、C#、Python、PyPy2、PyPy3、Go、Java、JavaScript V8、JavaScript Node、PHP、Ruby、Rust

### 3. 比赛功能
- ACM赛制
- OI赛制
- 打星队伍
- 关注队伍
- 外榜
- 滚榜
- 比赛打印
- 允许比赛结束后提交

### 4. 训练功能
- 私有训练
- 公开训练（题单）
- 自定义难度

### 5. 团队功能
- 团队创建与管理
- 团队题目
- 团队比赛

### 6. 讨论功能
- 公共讨论区
- 题目讨论区
- 比赛评论
- 评论回复
- 点赞功能
- 举报功能

### 7. 消息系统
- 站内消息
- 评论通知
- 回复通知
- 点赞通知
- 系统通知

### 8. 管理功能
- 用户管理
- 题目管理
- 比赛管理
- 讨论管理
- 系统配置
- 数据导入导出

## 数据库

### 主要数据表（推测）
- 用户表（user）
- 题目表（problem）
- 比赛表（contest）
- 提交记录表（judge）
- 讨论表（discussion）
- 团队表（group）
- 训练表（training）
- Rating表（rating）

## 部署架构

### 微服务架构
- **服务注册中心**: Nacos
- **API网关**: Spring Cloud Gateway（推测）
- **配置中心**: Nacos Config
- **服务发现**: Nacos Discovery

### 部署方式
- Docker Compose 一键部署
- 支持多判题服务器部署
- 支持分布式部署

## 开发规范

### 后端（Java）
- Java 8
- Spring Boot 2.2.6.RELEASE
- MyBatis Plus 3.2.0
- 包名: `top.hcode.hoj`

### 前端（Vue）
- Vue 2.6.11
- Element UI 2.15.3
- Vue Router 3.2.0
- Vuex 3.4.0

### Rating服务（Go）
- Go Modules
- RESTful API
- 数据库迁移支持

## 文档资源

- **在线文档**: https://docs.hdoi.cn
- **部署文档**: `docs/docs/deploy/`
- **开发文档**: `docs/docs/develop/`
- **使用文档**: `docs/docs/use/`

## 重要配置文件

- `hoj-springboot/*/src/main/resources/application.yml`: Spring Boot配置
- `hist-oj/configs/config.yaml`: Rating服务配置
- `hoj-vue/vue.config.js`: Vue构建配置
- `sqlAndsetting/hoj.sql`: 数据库初始化脚本

## 注意事项

1. **系统要求**: CentOS 8+ 或 Ubuntu 16.04+
2. **服务器配置**: 建议 2核4G 以上
3. **避免使用**: 突发性能或共享型云服务器实例
4. **判题沙箱**: 需要根据系统架构选择 amd64 或 arm64 版本

## 版本信息

- **当前版本**: 4.6
- **最后更新**: 2024-03-13（支持LibreOJ的远程评测）

