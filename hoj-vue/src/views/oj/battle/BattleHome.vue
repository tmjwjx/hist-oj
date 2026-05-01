<template>
  <div class="battle-home-container">
    <el-card class="battle-home-card">
      <div slot="header" class="battle-header">
        <div class="battle-title-wrapper">
          <div class="battle-logo">
            <i class="fa fa-gamepad fa-lg"></i>
          </div>
          <span class="battle-title">代码对战</span>
        </div>
      </div>

      <div class="battle-content">
        <!-- 四个模块统一布局 -->
        <el-row :gutter="20" class="modules-section">
          <el-col :xs="24" :sm="12" :md="12">
            <el-card class="module-card create-card" shadow="hover" @click.native="createRoom">
              <div class="module-icon">
                <i class="fa fa-plus-circle fa-3x"></i>
              </div>
              <h3 class="module-title">创建对战房间</h3>
              <p class="module-desc">创建房间并邀请好友对战</p>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="12">
            <el-card class="module-card join-card" shadow="hover" @click.native="showJoinDialog = true">
              <div class="module-icon">
                <i class="fa fa-sign-in fa-3x"></i>
              </div>
              <h3 class="module-title">加入对战房间</h3>
              <p class="module-desc">输入房间号加入对战</p>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="12">
            <el-card class="module-card rank-card" shadow="hover" @click.native="goToRank">
              <div class="module-icon">
                <i class="fa fa-trophy fa-3x"></i>
              </div>
              <h3 class="module-title">对战排行榜</h3>
              <p class="module-desc">查看高手排名</p>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12" :md="12">
            <el-card class="module-card records-card" shadow="hover" @click.native="goToMyRecords">
              <div class="module-icon">
                <i class="fa fa-history fa-3x"></i>
              </div>
              <h3 class="module-title">我的战绩</h3>
              <p class="module-desc">查看对战记录</p>
            </el-card>
          </el-col>
        </el-row>

        <!-- 规则说明 -->
        <el-card class="rules-card" shadow="never">
          <div slot="header" class="rules-header">
            <i class="fa fa-book"></i>
            <span>对战规则</span>
          </div>
          <div class="rules-content">
            <ol class="rules-list">
              <li>创建房间后，分享房间号给好友，等待好友加入</li>
              <li>双方进入房间后，房主点击"开始对战"按钮</li>
              <li>系统会从双方都未AC过的题目中随机选择一题</li>
              <li>最先AC的一方获胜，若一方放弃则另一方获胜</li>
              <li>对战结束后自动记录战绩，并更新排行榜</li>
            </ol>
          </div>
        </el-card>
      </div>
    </el-card>

    <!-- 加入房间对话框 -->
    <el-dialog
      title="加入对战房间"
      :visible.sync="showJoinDialog"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form :model="joinForm" :rules="joinRules" ref="joinForm" label-width="80px">
        <el-form-item label="房间号" prop="roomId">
          <el-input
            v-model="joinForm.roomId"
            placeholder="请输入6位房间号"
            maxlength="6"
            style="text-transform: uppercase;"
            @input="joinForm.roomId = joinForm.roomId.toUpperCase()"
          ></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="showJoinDialog = false">取 消</el-button>
        <el-button type="primary" @click="joinRoom" :loading="joinLoading">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { createRoom, joinRoom, getRoomInfo } from '@/api/battle';

export default {
  name: 'BattleHome',
  data() {
    return {
      showJoinDialog: false,
      joinLoading: false,
      joinForm: {
        roomId: ''
      },
      joinRules: {
        roomId: [
          { required: true, message: '请输入房间号', trigger: 'blur' },
          { pattern: /^[A-Z0-9]{6}$/, message: '房间号为6位数字或大写字母', trigger: 'blur' }
        ]
      }
    };
  },
  mounted() {
    // 检查用户是否在房间中，如果在则自动跳转回房间
    this.checkExistingRoom();
  },
  methods: {
    // 检查用户是否已在房间中
    async checkExistingRoom() {
      // 从 sessionStorage 获取当前用户的房间ID
      const userId = this.$store.getters.userInfo?.uid;
      if (!userId) return;

      const currentRoomId = sessionStorage.getItem(`battle_room_${userId}`);
      if (currentRoomId) {
        // 先检查房间是否还存在
        try {
          const res = await getRoomInfo(currentRoomId);
          // 房间已被解散或不存在
          if (res.data.code === 1 || !res.data.data) {
            // 清除无效的房间记录
            sessionStorage.removeItem(`battle_room_${userId}`);
            return;
          }

          // 房间存在，询问用户是否返回
          this.$confirm(`检测到您正在房间 ${currentRoomId} 中，是否返回？`, '提示', {
            confirmButtonText: '返回房间',
            cancelButtonText: '退出房间',
            type: 'info'
          }).then(() => {
            // 用户选择返回房间
            this.$router.push({
              name: 'BattleRoom',
              params: { roomId: currentRoomId }
            });
          }).catch(() => {
            // 用户选择退出房间，清除记录
            sessionStorage.removeItem(`battle_room_${userId}`);
          });
        } catch (error) {
          // 房间检查失败（可能已解散），清除记录
          sessionStorage.removeItem(`battle_room_${userId}`);
        }
      }
    },

    // 创建房间
    async createRoom() {
      try {
        const res = await createRoom();

        console.log('[BattleHome] 创建房间响应:', res.data);

        // 检查业务状态码
        if (res.data.code === 0) {
          // 成功
          const roomId = res.data.data.roomId;
          console.log('[BattleHome] 房间创建成功, roomId:', roomId);
          this.$message.success('房间创建成功');
          this.$router.push({
            name: 'BattleRoom',
            params: { roomId: roomId }
          });
        } else {
          // 业务逻辑错误
          const errorMsg = res.data.msg || '创建房间失败';

          // 检查是否是"你已在房间 XXXXX 中"的错误
          const match = errorMsg.match(/你已在房间\s+([A-Z0-9]{6})\s+中/);
          if (match) {
            // 用户已在其他房间中，提供返回房间的选项
            this.$confirm(errorMsg, '提示', {
              confirmButtonText: '返回房间',
              cancelButtonText: '取消',
              type: 'warning'
            }).then(() => {
              // 用户点击"返回房间"
              this.$router.push({
                name: 'BattleRoom',
                params: { roomId: match[1] }
              });
            }).catch(() => {
              // 用户取消，不操作
            });
          } else {
            // 其他错误，直接显示消息
            this.$message.error(errorMsg);
          }
        }
      } catch (error) {
        // 网络错误或其他异常
        this.$message.error('网络错误，请检查网络连接');
      }
    },

    // 加入房间
    joinRoom() {
      this.$refs.joinForm.validate(async (valid) => {
        if (valid) {
          this.joinLoading = true;
          try {
            const res = await joinRoom({ roomId: this.joinForm.roomId });

            // 检查业务状态码
            if (res.data.code === 0) {
              // 成功
              this.$message.success('加入房间成功');
              const roomId = this.joinForm.roomId; // 保存房间号
              this.showJoinDialog = false;
              this.joinForm.roomId = '';
              this.$router.push({
                name: 'BattleRoom',
                params: { roomId: roomId }
              });
            } else {
              // 业务逻辑错误
              const errorMsg = res.data.msg || '加入房间失败';

              // 检查是否是"你已在房间 XXXXX 中"的错误
              const match = errorMsg.match(/你已在房间\s+([A-Z0-9]{6})\s+中/);
              if (match) {
                // 用户已在其他房间中，先检查该房间是否还存在
                const existingRoomId = match[1];
                try {
                  const roomCheck = await getRoomInfo(existingRoomId);
                  // 如果房间已解散(code===1)或数据不存在，直接清除sessionStorage并继续加入新房间
                  if (roomCheck.data.code === 1 || !roomCheck.data.data) {
                    // 房间已解散，清除本地存储
                    const userId = this.$store.getters.userInfo?.uid;
                    if (userId) {
                      sessionStorage.removeItem(`battle_room_${userId}`);
                    }
                    // 重新尝试加入新房间
                    const retryRes = await joinRoom({ roomId: this.joinForm.roomId });
                    if (retryRes.data.code === 0) {
                      this.$message.success('加入房间成功');
                      const roomId = this.joinForm.roomId;
                      this.showJoinDialog = false;
                      this.joinForm.roomId = '';
                      this.$router.push({
                        name: 'BattleRoom',
                        params: { roomId: roomId }
                      });
                    } else {
                      this.$message.error(retryRes.data.msg || '加入房间失败');
                    }
                  } else {
                    // 房间仍存在，提供返回房间的选项
                    this.$confirm(errorMsg, '提示', {
                      confirmButtonText: '返回房间',
                      cancelButtonText: '取消',
                      type: 'warning'
                    }).then(() => {
                      this.showJoinDialog = false;
                      this.joinForm.roomId = '';
                      this.$router.push({
                        name: 'BattleRoom',
                        params: { roomId: existingRoomId }
                      });
                    }).catch(() => {
                      // 用户取消
                    });
                  }
                } catch (error) {
                  // 房间检查失败，说明房间已解散，清除本地存储并提示
                  const userId = this.$store.getters.userInfo?.uid;
                  if (userId) {
                    sessionStorage.removeItem(`battle_room_${userId}`);
                  }
                  this.$message.warning('你所在的房间已解散，正在加入新房间...');
                  // 重新尝试加入
                  const retryRes = await joinRoom({ roomId: this.joinForm.roomId });
                  if (retryRes.data.code === 0) {
                    this.$message.success('加入房间成功');
                    const roomId = this.joinForm.roomId;
                    this.showJoinDialog = false;
                    this.joinForm.roomId = '';
                    this.$router.push({
                      name: 'BattleRoom',
                      params: { roomId: roomId }
                    });
                  } else {
                    this.$message.error(retryRes.data.msg || '加入房间失败');
                  }
                }
              } else {
                // 其他错误，直接显示消息
                this.$message.error(errorMsg);
              }
            }
          } catch (error) {
            // 网络错误或其他异常
            this.$message.error('网络错误，请检查网络连接');
          } finally {
            this.joinLoading = false;
          }
        }
      });
    },

    // 跳转到排行榜
    goToRank() {
      this.$router.push({ name: 'BattleRank' });
    },

    // 跳转到我的战绩
    goToMyRecords() {
      this.$router.push({ name: 'BattleMyRecords' });
    }
  }
};
</script>

<style scoped>
.battle-home-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.battle-home-card {
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  animation: fadeInUp 0.6s ease;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.battle-header {
  text-align: center;
  padding: 30px 20px;
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  border-radius: 16px 16px 0 0;
  position: relative;
  overflow: hidden;
}

.battle-header::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255,255,255,0.1) 1px, transparent 1px);
  background-size: 20px 20px;
  animation: bgMove 20s linear infinite;
}

@keyframes bgMove {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(20px, 20px);
  }
}

.battle-title-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20px;
  position: relative;
  z-index: 1;
}

.battle-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 70px;
  height: 70px;
  background: linear-gradient(135deg, #FF6B6B 0%, #FF8E53 100%);
  border-radius: 20px;
  box-shadow: 0 8px 25px rgba(255, 107, 107, 0.5);
  color: white;
  font-size: 32px;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 8px 25px rgba(255, 107, 107, 0.5);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 12px 35px rgba(255, 107, 107, 0.7);
  }
}

.battle-title {
  font-size: 36px;
  font-weight: 700;
  background: linear-gradient(135deg, #ffffff 0%, #f0f0f0 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 2px 10px rgba(255, 255, 255, 0.3);
}

.battle-content {
  padding: 30px 20px;
}

.modules-section {
  margin-bottom: 30px;
}

.module-card {
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  border-radius: 16px;
  text-align: center;
  padding: 35px 20px;
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.module-card > * {
  width: 100%;
}

.module-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0;
  transition: opacity 0.3s;
  background: linear-gradient(135deg, rgba(79, 172, 254, 0.08) 0%, rgba(0, 242, 254, 0.08) 100%);
}

.module-card:hover::before {
  opacity: 1;
}

.module-card:hover {
  transform: translateY(-10px) scale(1.03);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.12);
  border-color: #4facfe;
}

.module-icon {
  margin: 0 auto 20px;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  position: relative;
  z-index: 1;
  width: 80px;
  height: 80px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: linear-gradient(135deg, #f0f9ff 0%, #e1f3ff 100%);
}

.module-card:hover .module-icon {
  transform: scale(1.15) rotate(5deg);
}

/* 创建房间 - 绿色主题 */
.create-card .module-icon {
  background: linear-gradient(135deg, #d4edda 0%, #c3e6cb 100%);
  color: #67C23A;
}

.create-card:hover {
  border-color: #67C23A;
}

/* 加入房间 - 蓝色主题 */
.join-card .module-icon {
  background: linear-gradient(135deg, #d1ecf1 0%, #bee5eb 100%);
  color: #409EFF;
}

.join-card:hover {
  border-color: #409EFF;
}

/* 排行榜 - 橙色主题 */
.rank-card .module-icon {
  background: linear-gradient(135deg, #FFE57F 0%, #FFB300 100%);
  color: #ffffff;
  box-shadow: 0 4px 15px rgba(255, 179, 0, 0.4);
}

.rank-card:hover {
  border-color: #FFB300;
}

.rank-card .module-icon i {
  animation: none;
}

.rank-card:hover .module-icon i {
  animation: bounce 0.6s ease;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

/* 我的战绩 - 紫色主题 */
.records-card .module-icon {
  background: linear-gradient(135deg, #e2d9f3 0%, #d3c7e8 100%);
  color: #9b59b6;
}

.records-card:hover {
  border-color: #9b59b6;
}

.module-title {
  font-size: 20px;
  font-weight: 700;
  color: #303133;
  margin: 0 0 12px 0;
  position: relative;
  z-index: 1;
  transition: all 0.3s ease;
}

.module-card:hover .module-title {
  transform: scale(1.05);
}

.module-desc {
  font-size: 14px;
  color: #909399;
  margin: 0;
  position: relative;
  z-index: 1;
}

.rules-card {
  margin-top: 30px;
  border-radius: 12px;
  background: linear-gradient(135deg, #fff9e6 0%, #fff3cd 100%);
  border: 2px solid #ffc107;
}

.rules-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 18px;
  color: #856404;
}

.rules-list {
  margin: 0;
  padding-left: 25px;
  line-height: 2.2;
}

.rules-list li {
  color: #856404;
  font-size: 15px;
  position: relative;
  padding-left: 10px;
}

.rules-list li::marker {
  color: #ffc107;
  font-size: 1.2em;
}

@media screen and (max-width: 768px) {
  .battle-home-container {
    padding: 10px;
  }

  .battle-title {
    font-size: 28px;
  }

  .battle-logo {
    width: 60px;
    height: 60px;
    font-size: 28px;
  }

  .module-card {
    padding: 25px 20px;
    margin-bottom: 15px;
  }

  .module-icon {
    width: 60px;
    height: 60px;
    margin: 0 auto 15px;
  }

  .module-icon i {
    font-size: 2rem !important;
  }

  .module-title {
    font-size: 18px;
  }

  .module-desc {
    font-size: 13px;
  }
}
</style>
