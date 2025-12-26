<template>
  <div class="about-us-page">
    <!-- Banner 区域 -->
    <div class="banner">
      <h1>Bing Online Judge</h1>
      <p>承办校级程序设计竞赛 · 全语言支持的高效评测系统 · 智能QQ机器人助手<br>加入校队，一起冲击ACM/ICPC、百度之星，天梯赛，码蹄杯等高水平竞赛！</p>
      <a href="/" class="banner-btn">进入评测系统</a>
    </div>

    <!-- 数据统计条 -->
    <div class="stats-bar">
      <div class="stats-grid">
        <div class="stat-item">
          <h3>C++/Java/Py/Go/C#/JS</h3>
          <p>全语言支持</p>
        </div>
        <div class="stat-item">
          <h3>School Contest</h3>
          <p>赛事承办</p>
        </div>
        <div class="stat-item">
          <h3>Competition Registration</h3>
          <p>赛事报名</p>
        </div>
        <div class="stat-item">
          <h3>ELO Rating</h3>
          <p>积分系统</p>
        </div>
        <div class="stat-item">
          <h3>QQ Bot</h3>
          <p>智能交互</p>
        </div>
      </div>
    </div>

    <div class="container">
      <!-- 团队介绍 -->
      <div class="team-section">
        <div class="section-title" style="margin-bottom: 10px;">
          <h2>团队成员</h2>
        </div>
        <p style="text-align: center; color: #666; margin-bottom: 40px;">致力于构建专业的算法竞赛训练平台</p>

        <!-- 指导教师 -->
        <div class="sub-title">指导教师</div>
        <div class="teacher-grid">
          <div class="teacher-card">
            <div class="teacher-avatar">李</div>
            <div>
              <h3 style="margin-bottom: 5px;">李学勇</h3>
              <p style="font-size: 13px; color: #666;">校队指导教师 / 竞赛总指导</p>
            </div>
          </div>
          <div class="teacher-card">
            <div class="teacher-avatar">孙</div>
            <div>
              <h3 style="margin-bottom: 5px;">孙甲霞</h3>
              <p style="font-size: 13px; color: #666;">校队指导教师 / 教学管理</p>
            </div>
          </div>
        </div>

        <!-- 核心团队成员 -->
        <div class="sub-title">团队成员</div>
        <div class="student-grid">
          <div class="member-card role-admin">
            <img src="https://ui-avatars.com/api/?name=Zhuang+Qingjia&background=faad14&color=fff&size=128&bold=true" class="m-avatar">
            <div class="m-name">庄青佳</div>
            <span class="m-role">管理员 / 第一任队长</span>
            <div class="m-info">21级 信息工程<br>平台统筹与校队建设</div>
            <div class="m-qq"><i class="fab fa-qq"></i> <span class="qq-text">3311918412</span></div>
          </div>
          <div class="member-card role-capt">
            <img src="https://ui-avatars.com/api/?name=Zhao+Yuming&background=52c41a&color=fff&size=128&bold=true" class="m-avatar">
            <div class="m-name">赵禹铭 <span class="capt-tag">招新对接</span></div>
            <span class="m-role">第二任队长</span>
            <div class="m-info">24级 计算机科学与技术<br>校队管理与纳新，欢迎咨询加入</div>
            <div class="m-qq"><i class="fab fa-qq"></i> <span class="qq-text">3175829058</span></div>
          </div>
          <div class="member-card role-dev">
            <img src="https://ui-avatars.com/api/?name=Ning+Zhaoxing&background=1890ff&color=fff&size=128&bold=true" class="m-avatar">
            <div class="m-name">宁赵星</div>
            <span class="m-role">OJ 技术人员</span>
            <div class="m-info">22级 计算机科学与技术<br>Online Judge 平台的搭建和维护</div>
            <div class="m-qq"><i class="fab fa-qq"></i> <span class="qq-text">1282270052</span></div>
          </div>
          <div class="member-card role-dev">
            <img src="https://ui-avatars.com/api/?name=Tian+Jiajie&background=1890ff&color=fff&size=128&bold=true" class="m-avatar">
            <div class="m-name">田家杰</div>
            <span class="m-role">OJ 技术人员</span>
            <div class="m-info">22级 计算机科学与技术<br>Online Judge 平台的开发和建设</div>
            <div class="m-qq"><i class="fab fa-qq"></i> <span class="qq-text">2937693919</span></div>
          </div>
        </div>
      </div>

      <!-- 平台特色 -->
      <div class="section-title">
        <h2>平台特色</h2>
      </div>
      <div class="feature-grid">
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-code"></i></div>
          <div class="feature-content">
            <h3>全面语言支持</h3>
            <p>打破语言壁垒！平台完美支持 <strong>C/C++、Java、Python、Go</strong> 等主流语言，同时特色支持 <strong>C#、JavaScript</strong>，内置多种编译器配置，满足各类开发者的刷题偏好。</p>
          </div>
        </div>
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-flag-checkered"></i></div>
          <div class="feature-content">
            <h3>专业赛事承办</h3>
            <p>BingoJ 已成功承办多次<strong>校级程序设计竞赛、新生赛及校队选拔赛</strong>，并且定期举办周赛。平台配备<strong>独立的赛事报名系统</strong>，支持选手在线报名、信息审核、报名状态查询等全流程管理。</p>
          </div>
        </div>
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-database"></i></div>
          <div class="feature-content">
            <h3>梯度化题库</h3>
            <p>拥有从语法入门到高难度的海量题库，覆盖<strong>校招算法题专项</strong>：包含主流互联网大厂（字节、腾讯、阿里、华为等）近三年校招/实习笔试、面试高频算法题。</p>
          </div>
        </div>
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-chart-line"></i></div>
          <div class="feature-content">
            <h3>独立 Rating 系统</h3>
            <p>采用类 Codeforces 的 <strong>Elo Rating 积分算法</strong>。系统根据选手在每场比赛中的排名动态计算 Rating 分值，生成专属<strong>积分变化曲线</strong>。</p>
          </div>
        </div>
      </div>

      <!-- QQ机器人助手 -->
      <div class="section-title">
        <h2>QQ机器人助手</h2>
        <h3>加入校队群，体验完整服务</h3>
      </div>
      <div class="feature-grid">
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-robot"></i></div>
          <div class="feature-content">
            <h3>实时编程PK对战</h3>
            <p>发送 <strong>/bat 用户名A 用户名B 难度</strong> 指令发起PK，系统随机匹配指定难度题目，实时监控评测队列，率先AC的用户自动判定获胜。</p>
          </div>
        </div>
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-trophy"></i></div>
          <div class="feature-content">
            <h3>多维度数据查询</h3>
            <p>• <strong>/rank</strong>：查看所有用户的PK胜场排名<br>• <strong>/rating 用户名</strong>：查询用户Rating值<br>• <strong>/contest</strong>：查看比赛信息</p>
          </div>
        </div>
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-random"></i></div>
          <div class="feature-content">
            <h3>随机题目推送</h3>
            <p>发送 <strong>/rand 难度</strong> 指令，快速获取指定难度的随机题目链接，满足日常刷题、赛前热身、PK选题等多样化需求。</p>
          </div>
        </div>
        <div class="feature-card">
          <div class="feature-icon"><i class="fas fa-question-circle"></i></div>
          <div class="feature-content">
            <h3>完整指令与数据保障</h3>
            <p>发送 <strong>/help</strong> 指令查看所有功能的详细使用格式，用户胜场排名数据自动保存，重启Bot后数据不丢失。</p>
          </div>
        </div>
      </div>

      <!-- 校队招募 -->
      <div class="section-title recruit-title">
        <h2>校队招募 · 招贤纳士</h2>
        <h3>只要你热爱算法，愿意坚持，我们都欢迎！</h3>
      </div>
      <div class="feature-grid">
        <div class="feature-card recruit-card">
          <div class="feature-icon recruit-icon"><i class="fas fa-users"></i></div>
          <div class="feature-content recruit-content">
            <h3>关于我们的校队</h3>
            <p>本校程序设计校队是由专业教师指导、学生自主管理的高水平竞赛团队，依托BingoJ平台开展日常训练与赛事备战。<br><br>
            ✅ 定期算法周赛<br>
            ✅ 组队参加ACM/ICPC、蓝桥杯、天梯赛等竞赛<br>
            ✅ 专属训练资源 + 竞赛经验分享<br>
            ✅ 评优评奖、保研加分等政策支持</p>
          </div>
        </div>
        <div class="feature-card recruit-card">
          <div class="feature-icon recruit-icon"><i class="fas fa-handshake"></i></div>
          <div class="feature-content recruit-content">
            <h3>招募要求 & 联系我们</h3>
            <p><strong>招募对象</strong>：全体在校学生（无年级/专业限制）<br>
            <strong>基本要求</strong>：<br>
            • 热爱编程与算法，有学习热情<br>
            • 能保证基础的训练时间<br>
            • 遵守校队训练规则，团队协作意识强<br>
            （零基础也可加入，我们有完整的入门培养体系！）</p>
            <div class="contact-tip">
              <strong>🔔 意向加入请联系：赵禹铭</strong><br>
              QQ：<span class="qq-text">3175829058</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <footer>
      <p>&copy; 2025 BingoJ Online Judge. All Rights Reserved.</p>
      <p>Made with <i class="fas fa-heart" style="color:#ff4d4f"></i> by BingoJ Team | 加入校队，一起成长，冲击更高水平竞赛！</p>
    </footer>
  </div>
</template>

<script>
export default {
  name: 'AboutUs',
  mounted() {
    window.scrollTo(0, 0);
  }
};
</script>

<style scoped>
/* 基础样式 */
* { margin: 0; padding: 0; box-sizing: border-box; }

.about-us-page {
  background-color: #f0f2f5;
  min-height: 100vh;
}

/* Banner 区域 */
.banner {
  background: linear-gradient(135deg, #001529 0%, #003a70 100%);
  padding: 80px 20px;
  text-align: center;
  color: #fff;
  position: relative;
  overflow: hidden;
}

/* ICPC Logo 水印背景 */
.banner::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80%;
  max-width: 500px;
  height: 100%;
  background-image: url('https://upload.wikimedia.org/wikipedia/commons/4/48/ICPC_Foundation_logo.svg');
  background-repeat: no-repeat;
  background-position: center;
  background-size: contain;
  opacity: 0.1;
  z-index: 0;
  filter: grayscale(100%) brightness(500%);
}

/* 确保 Banner 上的文字在 Logo 之上 */
.banner h1, .banner p, .banner-btn {
  position: relative;
  z-index: 1;
}

.banner h1 {
  font-size: 36px;
  margin-bottom: 15px;
  font-weight: 600;
  text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}

.banner p {
  opacity: 0.95;
  font-size: 18px;
  max-width: 800px;
  margin: 0 auto;
  text-shadow: 0 1px 2px rgba(0,0,0,0.2);
}

.banner-btn {
  display: inline-block;
  margin-top: 30px;
  padding: 10px 30px;
  background-color: #1890ff;
  color: #fff;
  border-radius: 4px;
  font-weight: 500;
  transition: background 0.3s;
  box-shadow: 0 4px 6px rgba(0,0,0,0.2);
  text-decoration: none;
}

.banner-btn:hover {
  background-color: #40a9ff;
}

/* 数据统计条 */
.stats-bar {
  background: #fff;
  padding: 25px 0;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  margin-bottom: 50px;
  position: relative;
  z-index: 2;
}

.stats-grid {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-around;
  text-align: center;
  flex-wrap: wrap;
}

.stat-item {
  margin: 10px;
}

.stat-item h3 {
  font-size: 20px;
  color: #1890ff;
  margin-bottom: 2px;
}

.stat-item p {
  font-size: 13px;
  color: #8c8c8c;
  text-transform: uppercase;
}

/* 通用容器 */
.container {
  max-width: 1200px;
  margin: 0 auto 60px;
  padding: 0 20px;
}

.section-title {
  text-align: center;
  margin-bottom: 40px;
  position: relative;
}

.section-title h2 {
  font-size: 28px;
  font-weight: 600;
  display: inline-block;
}

.section-title h3 {
  font-size: 16px;
  font-weight: 500;
  color: #8c8c8c;
  margin-top: 10px;
  font-style: normal;
}

.section-title::after {
  content: '';
  display: block;
  width: 60px;
  height: 3px;
  background: #1890ff;
  margin: 10px auto 0;
}

/* 招募板块标题特殊样式 */
.recruit-title::after {
  background: #52c41a;
}

/* 特色介绍卡片 (2x2 布局) */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 30px;
  margin-bottom: 80px;
}

.feature-card {
  background: #fff;
  border-radius: 12px;
  padding: 30px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  transition: all 0.3s;
  display: flex;
  gap: 20px;
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 20px rgba(0,0,0,0.1);
}

/* 招募卡片特殊样式 */
.recruit-card {
  background: #f6ffed;
  border: 2px solid #b7eb8f;
}

.recruit-card:hover {
  border-color: #52c41a;
}

.feature-icon {
  width: 60px;
  height: 60px;
  background: #e6f7ff;
  color: #1890ff;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
}

/* 招募卡片图标样式 */
.recruit-icon {
  background: rgba(82, 196, 26, 0.1);
  color: #52c41a;
}

.feature-content h3 {
  font-size: 20px;
  margin-bottom: 10px;
  color: #333;
}

.feature-content p {
  font-size: 14px;
  color: #666;
  line-height: 1.7;
}

/* 招募卡片文字样式 */
.recruit-content h3 {
  color: #52c41a;
}

.contact-tip {
  background: #fff;
  padding: 10px;
  border-radius: 6px;
  margin-top: 15px;
  border-left: 3px solid #52c41a;
}

.contact-tip strong {
  color: #52c41a;
}

/* 团队部分样式 */
.team-section {
  background: #fff;
  padding: 60px 20px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
  margin-bottom: 80px;
}

.sub-title {
  font-size: 18px;
  color: #262626;
  border-left: 4px solid #1890ff;
  padding-left: 12px;
  margin: 40px 0 20px;
}

/* 队长标注样式 */
.capt-tag {
  display: inline-block;
  background: #f6ffed;
  color: #52c41a;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  margin-left: 8px;
}

/* 教师卡片 */
.teacher-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.teacher-card {
  background: #f0f2f5;
  padding: 20px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 15px;
}

.teacher-avatar {
  width: 60px;
  height: 60px;
  background: #d6e4ff;
  color: #1890ff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
}

/* 学生卡片 */
.student-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 24px;
}

.member-card {
  background: #fff;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 25px 20px;
  text-align: center;
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}

.member-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 4px;
}

.role-admin::before {
  background: #faad14;
}

.role-capt::before {
  background: #52c41a;
}

.role-dev::before {
  background: #1890ff;
}

.member-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 20px rgba(0,0,0,0.1);
}

.m-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  margin: 0 auto 15px;
  border: 3px solid #f0f2f5;
}

.m-name {
  font-weight: bold;
  font-size: 18px;
  margin-bottom: 5px;
}

.m-role {
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 12px;
  display: inline-block;
  margin-bottom: 10px;
}

.role-admin .m-role {
  background: #fffbe6;
  color: #faad14;
}

.role-capt .m-role {
  background: #f6ffed;
  color: #52c41a;
}

.role-dev .m-role {
  background: #e6f7ff;
  color: #1890ff;
}

.m-info {
  font-size: 13px;
  color: #666;
  margin-bottom: 15px;
  min-height: 40px;
}

.m-qq {
  font-size: 13px;
  color: #888;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  background: #f9f9f9;
  padding: 5px;
  border-radius: 4px;
}

.m-qq i {
  color: #1890ff;
}

.qq-text {
  user-select: text;
}

/* 页脚 */
footer {
  text-align: center;
  padding: 40px 0;
  color: #888;
  font-size: 14px;
  background: #fff;
  border-top: 1px solid #eee;
  margin-top: 60px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
  .banner h1 {
    font-size: 28px;
  }
}
</style>
