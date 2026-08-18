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
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.battle-home-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.battle-header {
  background: #fff;
  border-bottom: 2px solid #409eff;
  padding: 20px;
}

.battle-title-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.battle-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: #409eff;
  border-radius: 4px;
  color: white;
  font-size: 20px;
}

.battle-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.battle-content {
  padding: 24px 20px;
  background: #fff;
}

.modules-section {
  margin-bottom: 24px;
}

.module-card {
  cursor: pointer;
  transition: all 0.2s;
  border-radius: 4px;
  text-align: center;
  padding: 30px 20px;
  background: #fff;
  border: 1px solid #e4e7ed;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.module-card > * {
  width: 100%;
}

.module-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.module-icon {
  margin: 0 auto 16px;
  transition: transform 0.2s;
  width: 60px;
  height: 60px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #ecf5ff;
}

.module-card:hover .module-icon {
  transform: scale(1.1);
}

/* 创建房间 - 绿色主题 */
.create-card .module-icon {
  background: #f0f9ff;
  color: #67C23A;
}

.create-card:hover {
  border-color: #67C23A;
}

/* 加入房间 - 蓝色主题 */
.join-card .module-icon {
  background: #ecf5ff;
  color: #409EFF;
}

.join-card:hover {
  border-color: #409EFF;
}

/* 排行榜 - 橙色主题 */
.rank-card .module-icon {
  background: #fdf6ec;
  color: #e6a23c;
}

.rank-card:hover {
  border-color: #e6a23c;
}

/* 我的战绩 - 紫色主题 */
.records-card .module-icon {
  background: #f4f4f5;
  color: #909399;
}

.records-card:hover {
  border-color: #909399;
}

.module-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.module-desc {
  font-size: 13px;
  color: #606266;
  margin: 0;
  line-height: 1.6;
}

.rules-card {
  margin-top: 24px;
  border-radius: 4px;
  background: #fef0f0;
  border: 1px solid #fbc4c4;
}

.rules-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
  color: #f56c6c;
  padding: 12px 20px;
  background: #fff;
  border-bottom: 1px solid #fbc4c4;
}

.rules-content {
  padding: 16px 20px;
}

.rules-list {
  margin: 0;
  padding-left: 20px;
  line-height: 1.8;
}

.rules-list li {
  color: #606266;
  font-size: 14px;
  margin-bottom: 8px;
}

.rules-list li::marker {
  color: #f56c6c;
}

@media screen and (max-width: 768px) {
  .battle-home-container {
    padding: 10px;
  }

  .battle-title {
    font-size: 20px;
  }

  .battle-logo {
    width: 36px;
    height: 36px;
    font-size: 18px;
  }

  .module-card {
    padding: 20px 15px;
    margin-bottom: 15px;
  }

  .module-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 12px;
  }

  .module-icon i {
    font-size: 2rem !important;
  }

  .module-title {
    font-size: 15px;
  }

  .module-desc {
    font-size: 12px;
  }
}
</style>
