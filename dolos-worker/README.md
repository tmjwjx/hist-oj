# Hist-OJ Dolos worker

这是一个与 Java 后端解耦的本地查重 worker。它使用官方 `@dodona/dolos-lib@3.5.1`，在内存中创建 Dolos `File` 对象，不把提交代码写入磁盘，也不需要访问 Docker socket。

## 接口

### `GET /healthz`

返回 worker 和 Dolos 库版本、当前并发数。

### `POST /v1/analyze`

请求体：

```json
{
  "language": "python",
  "submissions": [
    {"id": "submit-1", "filename": "main.py", "code": "print(1)"},
    {"id": "submit-2", "filename": "solution.py", "code": "print(1)"}
  ],
  "options": {
    "kgramLength": 23,
    "kgramsInWindow": 17,
    "minFragmentLength": 1,
    "minSimilarity": 0,
    "includeComments": false
  },
  "fragmentLimit": 200
}
```

`language` 可以省略，让 Dolos 按文件扩展名自动识别。建议调用方先按“题目 + 语言族”分组，不要把不同语言放到同一个分析请求中。常见平台语言别名会被归一化，例如 `Python3` → `python`、`C++17` → `cpp`、`Golang` → `go`。

成功响应中的 `pairs` 包含每一对提交的双向覆盖率（`similarity1to2`、`similarity2to1` 为 0-100 整数）、对称相似度、重叠 k-gram 数、最长匹配片段和左右代码区域。行号和列号使用 Dolos/Tree-sitter 的 0-based 坐标。

## 本地运行

需要 Node.js 22 LTS、Python 3 和 C/C++ 编译器（Tree-sitter 原生模块）。

```bash
npm ci
npm test
npm start
```

默认监听 `127.0.0.1:8091`。生产环境建议设置 `HOST=127.0.0.1` 或仅绑定内网地址，并通过现有后端做鉴权和请求限流。

如果 worker 需要监听非本机地址，请设置 `AUTH_TOKEN`。查重请求需携带
`Authorization: Bearer <AUTH_TOKEN>` 或 `X-Worker-Token: <AUTH_TOKEN>`；`/healthz` 保持免鉴权，便于容器健康检查。

## Docker

```bash
docker build -t hist-oj-dolos-worker:0.1.0 .
docker run --rm -p 127.0.0.1:8091:8091 hist-oj-dolos-worker:0.1.0
```

线上建议只加入后端所在的 Docker 内网，不发布宿主机端口：

```bash
docker run -d --restart unless-stopped \
  --name hoj-dolos-worker \
  --network main_hoj-network \
  -e HOST=0.0.0.0 \
  -e PORT=8091 \
  -e AUTH_TOKEN='<随机长令牌>' \
  hist-oj-dolos-worker:0.1.0
```

Java 后端对应配置为 `HOJ_PLAGIARISM_DOLOS_WORKER_URL=http://hoj-dolos-worker:8091`，并使用同一个令牌。

Docker 镜像不挂载 `/var/run/docker.sock`，分析过程只在 worker 容器内运行。默认限制：单请求最多 1,000 份提交、合计源代码 64 MiB、请求体 128 MiB、单次分析 900 秒、最多 200,000 对结果。批量调用方可以将 `fragmentLimit` 设为 `0`，只返回相似度，不生成片段详情，以控制响应体大小。可用环境变量调整，但不建议在线上无限放大。

## 许可证

worker 代码沿用仓库许可证；Dolos 本身及当前 bundled Tree-sitter parsers 为 MIT。发布镜像时请保留 `node_modules/@dodona/dolos-lib`、`@dodona/dolos-core` 和 `@dodona/dolos-parsers` 中的许可证文件。
