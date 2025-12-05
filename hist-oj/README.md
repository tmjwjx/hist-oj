# Hist-OJ

HOJ 扩展服务，提供 Rating 计算等功能。

## 功能

- 基于 Elo 算法的 Rating 计算
- 新手保护（前 3 场掉分减半）
- 参与激励（过题 +5 分）
- 比赛结束自动计算
- RESTful API

## 快速部署

### 1. 初始化数据库

```bash
mysql -h <---数据库地址---> -u <---用户名---> -p hoj < migrations/init.sql
```

### 2. 修改配置

编辑 `docker-compose.yml`：

```yaml
environment:
  - DATABASE_HOST=<---HOJ数据库地址，同网络填容器名，否则填IP--->
  - DATABASE_USER=<---数据库用户名--->
  - DATABASE_PASSWORD=<---数据库密码--->
  - HOJ_API_BASE_URL=<---HOJ后端地址，如 http://hoj-backend:6688--->
  - JWT_SECRET=<---与HOJ一致的JWT密钥--->

networks:
  - <---HOJ的docker网络名--->
```

### 3. 部署

```bash
make deploy
```

### 4. 验证

```bash
curl http://localhost:8080/health
```

## API 接口

| 方法 | 路径 | 权限 | 说明 |
|-----|------|------|------|
| GET | `/api/rating/user/:uid` | 公开 | 获取用户 Rating |
| GET | `/api/rating/history/:uid` | 公开 | 获取 Rating 历史 |
| GET | `/api/rating/color/:rating` | 公开 | 获取 Rating 颜色 |
| GET | `/api/rating/contest/:contestId` | 公开 | 获取比赛参赛者 Rating |
| POST | `/api/rating/calculate/:contestId` | 管理员 | 手动触发计算 |
| GET | `/health` | 公开 | 健康检查 |

## 设置计分比赛

```sql
INSERT INTO contest_rating_status (contest_id, is_rated, rating_calculated)
VALUES (<---比赛ID--->, 1, 0);
```

比赛结束后会自动计算 Rating。

## 常用命令

```bash
make deploy         # 一键部署
make docker-logs    # 查看日志
make docker-restart # 重启
make docker-stop    # 停止
make status         # 查看状态
```

## 文档

- [架构设计](docs/架构设计.md) - 系统架构和模块说明
- [部署文档](docs/部署文档.md) - 详细的部署说明和配置指南
- [Rating 计算规则](docs/Rating计算规则.md) - Rating 算法规则说明
