# Logo 更换指南

本文档说明如何更换 BingOJ 网站的 Logo（前端Logo、管理员Logo、网站图标）。

---

## 📋 Logo 文件说明

BingOJ 系统中有 **3个主要的 Logo**：

1. **前端 Logo** (`logo.png`) - 显示在导航栏左侧，用于学生端和公共页面
2. **管理员 Logo** (`backstage.png`) - 显示在管理后台侧边栏顶部
3. **网站图标** (`favicon.ico`) - 显示在浏览器标签页

---

## 📂 文件位置结构

```
hist-oj/                              # 项目根目录
├── logo.png                          # ⭐ Docker映射源文件（前端Logo）
├── backstage.png                      # ⭐ Docker映射源文件（管理员Logo）
├── favicon.ico                       # ⭐ Docker映射源文件（网站图标）
├── hoj-vue/
│   ├── src/assets/
│   │   ├── logo.png                 # 开发时引用（前端）
│   │   └── backstage.png           # 开发时引用（管理员）
│   └── public/
│       ├── logo.png                 # 构建时复制到 dist
│       └── favicon.ico              # 构建时复制到 dist
└── dist/                             # 构建生成（自动生成）
    └── assets/img/
        ├── logo.a0feb5a8.png        # 前端实际使用的文件
        ├── backstage.a0feb5a8.png    # 管理员实际使用的文件
        └── favicon.ico               # 网站图标
```

---

## 🛠️ 如何更换 Logo

### 方法一：直接替换文件（推荐）

#### 1. 准备新的 Logo 图片

- 图片格式：PNG 或 ICO
- 建议尺寸：
  - 前端 Logo：139px × 50px（高 × 宽）
  - 管理员 Logo：根据实际需求调整
  - 网站图标：32px × 32px 或 64px × 64px

#### 2. 备份当前 Logo（可选）

```bash
cp logo.png logo.png.backup
cp backstage.png backstage.png.backup
cp favicon.ico favicon.ico.backup
```

#### 3. 替换 Logo 文件

将新的 Logo 图片重命名并放置在项目根目录：

```bash
# 假设新Logo.png 是您的新 Logo
cp 新Logo.png logo.png          # 前端 Logo
cp 新Logo.png backstage.png     # 管理员 Logo
cp 新Logo.png favicon.ico       # 网站图标
```

如果新 Logo 是 PNG 格式，需要转换为 ICO 格式作为网站图标，可以使用在线工具或：

```bash
# 使用 ImageMagick 转换（需要安装）
convert logo.png -resize 32x32 favicon.ico
```

#### 4. 更新源文件（保持一致性）

```bash
# 更新源文件
cp logo.png hoj-vue/src/assets/logo.png
cp logo.png hoj-vue/public/logo.png

cp backstage.png hoj-vue/src/assets/backstage.png
cp favicon.ico hoj-vue/public/favicon.ico
```

#### 5. 重新构建前端（如果需要）

如果修改了 `src/assets/` 或 `public/` 下的文件，需要重新构建：

```bash
cd hoj-vue
npm run build
```

#### 6. 重启 Docker 容器

```bash
cd /Users/zhuangqingjia/vscode/histoj/hist-oj
docker-compose up -d hoj-frontend
```

#### 7. 清除浏览器缓存查看效果

- **Chrome/Edge**: `Ctrl + Shift + R` (Windows) 或 `Cmd + Shift + R` (Mac)
- **Firefox**: `Ctrl + Shift + Delete` (Windows) 或 `Cmd + Shift + Delete` (Mac)
- 或者使用无痕模式/隐私模式验证

---

### 方法二：仅替换 Docker 映射文件（快速更新）

如果只想更新 Logo 而不重新构建，只需要替换项目根目录的 3 个文件：

```bash
# 1. 替换 Logo 文件
cp 新Logo.png logo.png
cp 新Logo.png backstage.png
cp 新Logo.png favicon.ico

# 2. 重启容器
docker-compose up -d hoj-frontend
```

**注意**：这种方法只更新了 Docker 映射的文件，不会影响构建后的 dist 文件。但如果 Docker volumes 配置正确，容器内的文件会被覆盖。

---

## ⚙️ Docker 配置说明

### Docker Compose 配置 (`docker-compose.yml`)

```yaml
hoj-frontend:
  volumes:
    # Logo 映射（前端logo、后台logo、网站图标）
    - ./logo.png:/usr/share/nginx/html/assets/img/logo.a0feb5a8.png
    - ./backstage.png:/usr/share/nginx/html/assets/img/backstage.a0feb5a8.png
    - ./favicon.ico:/usr/share/nginx/html/favicon.ico
```

**工作原理**：
- Docker 会将项目根目录的 `logo.png`、`backstage.png`、`favicon.ico` 映射到容器内
- 容器内的文件会被主机的文件覆盖
- 修改主机文件并重启容器后，Logo 立即更新

---

## 🔍 验证 Logo 是否更新成功

### 1. 检查容器内的文件

```bash
docker exec hoj-frontend ls -lh /usr/share/nginx/html/assets/img/logo.a0feb5a8.png
docker exec hojfrontend ls -lh /usr/share/nginx/html/assets/img/backstage.a0feb5a8.png
docker exec hoj-frontend ls -lh /usr/share/nginx/html/favicon.ico
```

### 2. 在浏览器中检查

- **前端页面**：访问 `https://bingoj.cn/`，查看导航栏左侧的 Logo
- **管理后台**：访问 `https://bingoj.cn/admin` 或其他管理员界面，查看侧边栏顶部的 Logo
- **浏览器标签页**：查看标签页上的图标

---

## 📝 注意事项

### 1. 文件大小

- 建议单个 Logo 文件大小不超过 **100KB**
- 过大的 Logo 会影响页面加载速度

### 2. 图片格式

- **前端 Logo**：推荐使用 PNG 格式（支持透明背景）
- **网站图标**：必须是 ICO 格式（或 PNG 格式，浏览器会自动转换）
- 也可以使用 SVG 格式，但需要修改配置

### 3. 图片尺寸

建议保持宽高比，避免 Logo 变形：

| Logo 类型 | 建议尺寸 | 宽高比 |
|-----------|---------|--------|
| 前端 Logo | 139px × 50px | 约 2.8:1 |
| 管理员 Logo | 根据实际需求 | 自定义 |
| 网站图标 | 32px × 32px 或 64px × 64px | 1:1 |

### 4. 文件权限

确保文件具有正确的读取权限：

```bash
chmod 644 logo.png backstage.png favicon.ico
```

---

## 🐛 常见问题

### Q1: 更换了 Logo 但浏览器还显示旧的？

**A**: 浏览器缓存问题，请：
1. 强制刷新浏览器（`Ctrl + Shift + R` 或 `Cmd + Shift + R`）
2. 清除浏览器缓存
3. 使用无痕模式验证

### Q2: 管理后台的 Logo 没有更新？

**A**: 请检查：
1. `backstage.png` 文件是否已更新
2. Docker 容器是否已重启
3. 浏览器缓存是否已清除

### Q3: 如何查看当前使用的 Logo 文件？

**A**: 在浏览器开发者工具中：
1. 打开开发者工具（F12）
2. 查看 Network 面板
3. 找到 `logo.a0feb5a8.png` 或 `backstage.a0feb5a8.png`
4. 预览图片查看内容

### Q4: Logo 显示变形或模糊？

**A**: 检查：
1. Logo 图片的原始尺寸和宽高比
2. CSS 中是否有 `object-fit` 或 `width/height` 设置
3. 图片分辨率是否足够高

### Q5: 更换 Logo 后容器启动失败？

**A**: 检查：
1. 文件路径是否正确
2. 文件是否存在
3. 文件权限是否正确

---

## 📚 相关文档

- [Vue CLI 静态资源处理](https://cli.vuejs.org/guide/html-and-static-assets.html)
- [Docker Volumes 文件映射](https://docs.docker.com/storage/volumes/)
- [Favicon 生成工具](https://www.favicon-generator.org/)

---

## 📅 更新记录

- **2025-03-05**: 初始化文档，说明 Logo 更换流程
- **作者**: BingOJ 开发团队

---

## 📞 技术支持

如有问题，请联系 BingOJ 技术支持团队。
