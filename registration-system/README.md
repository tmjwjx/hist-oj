# 报名系统集成到HOJ项目部署指南

## 项目概述

报名系统已成功集成到HOJ项目中,包括前端用户界面、管理员界面和Go后端服务。系统实现了基于HOJ用户认证的自动登录功能。

## 项目结构

```
histoj/hist-oj/
├── hoj-vue/                              # HOJ前端Vue项目
│   ├── public/
│   │   └── registration/                 # 报名系统前端文件
│   │       ├── registration.html         # 用户报名界面
│   │       ├── admin.html                # 管理员界面
│   │       ├── admin-registrations.html  # 报名管理界面
│   │       ├── admin-edit-competition.html # 比赛编辑界面
│   │       └── competition-detail.html   # 比赛详情界面
│   └── src/
│       ├── components/oj/common/
│       │   └── NavBar.vue                # 主导航栏(已添加报名系统入口)
│       ├── router/
│       │   ├── ojRoutes.js               # 用户路由(已添加/registration)
│       │   └── adminRoutes.js            # 管理员路由(已添加/admin/registration)
│       ├── views/oj/registration/
│       │   └── Registration.vue          # 报名系统Vue组件
│       └── views/admin/registration/
│           └── RegistrationAdmin.vue     # 管理员后台Vue组件
└── registration-system/                  # 报名系统Go后端
    ├── backend/
    │   └── main.go                       # Go后端服务器
    └── frontend/                         # 原始前端文件(备份)
```

## 部署配置

### 1. 数据库配置

Go后端已配置为使用远程MySQL数据库:

- **服务器IP**: 43.143.133.62
- **端口**: 3306
- **数据库名**: hoj
- **用户名**: root
- **密码**: hist2025

数据库表结构会在Go后端首次启动时自动创建。

### 2. HOJ后端配置

HOJ后端配置为:
- **服务器IP**: 43.143.133.62
- **端口**: 6688
- **域名**: bingoj.cn

### 3. Go后端配置

Go后端服务器配置:
- **端口**: 8080
- **数据库**: MySQL (远程)
- **表前缀**: histcontest_register_

### 4. 前端配置

HOJ Vue前端已配置代理:
- 开发环境: `/registration-api/api` -> `http://localhost:8080/api`
- 生产环境: 需要Nginx配置反向代理

## 部署步骤

### 第一步: 启动Go后端服务

```bash
cd /Users/zhuangqingjia/vscode/histoj/hist-oj/registration-system/backend

# 安装Go依赖
go mod tidy

# 启动Go后端服务(前台运行)
go run main.go

# 或使用启动脚本(后台运行)
cd ../
./start.sh -d
```

Go后端会自动:
1. 连接到MySQL数据库 43.143.133.62:3306
2. 创建所需的数据表(如果不存在)
3. 监听8080端口
4. 提供RESTful API服务

### 第二步: 启动HOJ前端

```bash
cd /Users/zhuangqingjia/vscode/histoj/hist-oj/hoj-vue

# 安装依赖
npm install

# 启动开发服务器
npm run serve
```

前端会运行在 `http://localhost:8066`

### 第三步: 访问系统

1. **用户界面**: 访问 `http://localhost:8066/registration`
   - 需要先登录HOJ账号
   - 系统会自动使用HOJ的用户信息登录报名系统

2. **管理员界面**: 访问 `http://localhost:8066/admin/registration`
   - 需要HOJ管理员权限
   - 可以管理比赛和审核报名

## 功能特性

### 1. 基于HOJ的自动登录

- 用户登录HOJ后,访问报名系统会自动登录
- 使用HOJ的token和username进行身份验证
- 如果用户在报名系统中不存在,会自动创建

### 2. 用户界面功能

- 查看可用比赛列表
- 报名参加比赛
- 填写报名信息(姓名、班级、学院、学号等)
- 查看报名状态(待审核、已通过、已拒绝)
- 与管理员聊天沟通

### 3. 管理员功能

- 创建和管理比赛
- 设置报名字段(可自定义需要填写的字段)
- 审核报名(通过/拒绝)
- 与报名用户聊天
- 导出报名数据
- 上传比赛Logo

## API端点

Go后端提供的API:

- `GET /registration-api/api/competitions` - 获取比赛列表
- `POST /registration-api/api/competitions` - 创建比赛
- `GET /registration-api/api/competitions/{id}` - 获取比赛详情
- `PUT /registration-api/api/competitions/{id}` - 更新比赛
- `DELETE /registration-api/api/competitions/{id}` - 删除比赛
- `GET /registration-api/api/competitions/{competitionId}/registrations` - 获取报名列表
- `POST /registration-api/api/registrations` - 创建报名
- `GET /registration-api/api/competitions/{competitionId}/user-registration` - 获取用户报名
- `PUT /registration-api/api/registrations/{id}` - 更新报名状态
- `POST /registration-api/api/user/login` - 用户登录(传统方式)
- `GET /registration-api/api/user/hoj-auto-login` - HOJ自动登录
- `POST /registration-api/api/upload/image` - 上传图片

## 生产环境部署

### Nginx配置示例

```nginx
server {
    listen 80;
    server_name bingoj.cn;

    # HOJ前端
    location / {
        root /path/to/hoj-vue/dist;
        try_files $uri $uri/ /index.html;
    }

    # HOJ后端API代理
    location /api {
        proxy_pass http://43.143.133.62:6688;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 报名系统Go后端代理
    location /registration-api/api {
        proxy_pass http://localhost:8080/api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 报名系统静态文件
    location /registration {
        alias /path/to/hoj-vue/public/registration;
        index registration.html;
    }
}
```

### 使用systemd管理Go后端

创建服务文件 `/etc/systemd/system/registration.service`:

```ini
[Unit]
Description=Contest Registration System Go Backend
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/registration-system/backend
ExecStart=/usr/local/go/bin/go run /path/to/registration-system/backend/main.go
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启动服务:
```bash
sudo systemctl daemon-reload
sudo systemctl start registration
sudo systemctl enable registration
sudo systemctl status registration
```

## 数据库表结构

### histcontest_register_competitions (比赛表)
- id: 比赛ID
- name: 比赛名称
- start_time: 开始时间
- end_time: 结束时间
- fields: 字段配置(JSON)
- logo_url: Logo图片URL
- description: 比赛说明(HTML)
- visible: 是否可见
- created_at: 创建时间

### histcontest_register_registrations (报名表)
- id: 报名ID
- competition_id: 比赛ID
- user_uuid: 用户UUID
- name: 姓名
- class: 班级
- college: 学院
- student_id: 学号
- gender: 性别
- shirt_size: 衬衫大小
- team_name: 队名
- qq: QQ号
- status: 状态(pending/approved/rejected)
- remark: 备注(JSON格式,存储聊天记录)
- created_at: 创建时间

### user_info (用户表)
- uuid: 用户UUID
- username: 用户名
- password: 密码(MD5加密,HOJ用户为空)

## 故障排查

### 1. Go后端无法启动

检查数据库连接:
```bash
mysql -h 43.143.133.62 -u root -p hist2025 hoj
```

### 2. 前端无法连接后端

检查代理配置是否正确,查看浏览器Network标签中的请求URL。

### 3. 自动登录失败

检查:
- HOJ是否正常登录
- token是否正确传递
- 浏览器控制台是否有错误

## 注意事项

1. **不要删除或移动报名系统原始文件** - `报名系统/` 目录作为备份保留
2. **数据库连接信息** - 已配置为线上部署,使用MySQL远程数据库
3. **端口冲突** - 确保8080端口没有被其他服务占用
4. **CORS配置** - Go后端已配置全局CORS,允许跨域请求
5. **文件上传** - 上传的图片存储在 `backend/uploads/` 目录

## 更新日志

- 已集成HOJ用户认证自动登录
- 已配置线上数据库部署
- 已添加前端用户和管理员入口
- 已配置Vue代理支持Go后端
- 已删除传统登录界面,使用HOJ统一认证

## 联系方式

如有问题,请联系系统管理员。
