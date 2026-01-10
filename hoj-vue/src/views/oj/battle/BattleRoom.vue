<template>
  <div class="battle-room-container">
    <!-- 等待状态 -->
    <el-card v-if="room.status === 0" class="waiting-card">
      <div slot="header" class="card-header">
        <span>房间号：{{ roomId }}</span>
        <div class="header-actions">
          <el-button type="text" @click="copyRoomId">
            <i class="fa fa-copy"></i> 复制
          </el-button>
          <el-button v-if="isHost" type="text" style="color: #F56C6C;" @click="handleDissolveRoom">
            <i class="fa fa-times"></i> 解散房间
          </el-button>
        </div>
      </div>

      <div class="waiting-content">
        <div class="player-section">
          <div class="player-card host">
            <div class="player-avatar">
              <i class="fa fa-user-circle fa-5x"></i>
            </div>
            <div class="player-info">
              <h3 :style="{ color: getRatingColor(room.hostRating) }">{{ room.hostUsername }}</h3>
              <el-tag type="success">房主</el-tag>
            </div>
          </div>

          <div class="vs-badge">
            <span>VS</span>
          </div>

          <div class="player-card challenger">
            <div v-if="room.challengerUsername" class="player-avatar">
              <i class="fa fa-user-circle fa-5x"></i>
            </div>
            <div v-else class="player-avatar empty">
              <i class="fa fa-user fa-3x"></i>
            </div>
            <div class="player-info">
              <h3 v-if="room.challengerUsername" :style="{ color: getRatingColor(room.challengerRating) }">{{ room.challengerUsername }}</h3>
              <h3 v-else>等待玩家加入...</h3>
              <el-tag v-if="room.challengerUsername" type="warning">挑战者</el-tag>
            </div>
          </div>
        </div>

        <div v-if="isHost" class="action-section">
          <el-button
            type="primary"
            size="large"
            :disabled="!room.challengerUsername"
            @click="handleStartBattle"
          >
            <i class="fa fa-play"></i> 开始对战
          </el-button>
          <el-button
            type="info"
            size="large"
            style="margin-left: 10px;"
            @click="handleLeaveRoom"
          >
            <i class="fa fa-sign-out-alt"></i> 退出房间
          </el-button>
        </div>
        <div v-else class="action-section">
          <el-alert
            title="等待房主开始对战"
            type="info"
            :closable="false"
            show-icon
          ></el-alert>
          <el-button
            type="danger"
            size="large"
            style="margin-top: 15px;"
            @click="handleLeaveRoom"
          >
            <i class="fa fa-sign-out-alt"></i> 退出房间
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 对战状态 -->
    <el-card v-else-if="room.status === 1" class="battle-card">
      <div slot="header" class="card-header">
        <span>对战进行中</span>
        <div class="header-actions">
          <el-button v-if="isHost" type="text" style="color: #F56C6C;" @click="handleDissolveRoom">
            <i class="fa fa-times"></i> 解散房间
          </el-button>
          <el-button type="danger" size="small" @click="handleGiveup">
            <i class="fa fa-flag"></i> 放弃对战
          </el-button>
        </div>
      </div>

      <div class="battle-content">
        <div class="problem-section">
          <el-alert
            :title="`对战题目：${problem.title}`"
            type="success"
            :closable="false"
            show-icon
          ></el-alert>
          <el-button
            type="primary"
            style="margin-top: 15px;"
            @click="goToProblem"
          >
            <i class="fa fa-code"></i> 前往题目页面
          </el-button>
        </div>

        <el-divider></el-divider>

        <div class="players-status">
          <div class="status-card" :class="{ winner: battleResult && battleResult.winnerId === room.hostId }">
            <div class="status-header">
              <i class="fa fa-user-circle fa-3x player-avatar-icon"></i>
              <div class="player-name">{{ room.hostUsername }}</div>
              <el-tag v-if="battleResult && battleResult.winnerId === room.hostId" type="success">获胜</el-tag>
            </div>
            <div class="status-body">
              <div class="submit-count">提交次数: {{ hostSubmitCount }}</div>
              <div class="latest-status">
                最新状态:
                <el-tag :type="getStatusTagType(hostLatestStatus)" size="small">
                  {{ getStatusText(hostLatestStatus) }}
                </el-tag>
              </div>
            </div>
          </div>

          <div class="status-card" :class="{ winner: battleResult && battleResult.winnerId === room.challengerId }">
            <div class="status-header">
              <i class="fa fa-user-circle fa-3x player-avatar-icon"></i>
              <div class="player-name">{{ room.challengerUsername }}</div>
              <el-tag v-if="battleResult && battleResult.winnerId === room.challengerId" type="success">获胜</el-tag>
            </div>
            <div class="status-body">
              <div class="submit-count">提交次数: {{ challengerSubmitCount }}</div>
              <div class="latest-status">
                最新状态:
                <el-tag :type="getStatusTagType(challengerLatestStatus)" size="small">
                  {{ getStatusText(challengerLatestStatus) }}
                </el-tag>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 结束状态 -->
    <el-card v-else-if="room.status === 2" class="result-card">
      <div class="result-content">
        <div v-if="isWinner" class="winner-section">
          <i class="fa fa-trophy fa-5x" style="color: #FFD700;"></i>
          <h2>恭喜获胜！</h2>
          <p>{{ endReasonText }}</p>
        </div>
        <div v-else class="loser-section">
          <i class="fa fa-heart-broken fa-5x" style="color: #909399;"></i>
          <h2>对战失败</h2>
          <p>{{ endReasonText }}</p>
        </div>

        <div class="result-actions">
          <el-button type="success" @click="rematch">
            <i class="fa fa-redo"></i> 再来一局
          </el-button>
          <el-button v-if="isHost" type="danger" @click="handleDissolveRoom">
            <i class="fa fa-times"></i> 解散房间
          </el-button>
          <el-button type="danger" @click="handleLeaveRoom">
            <i class="fa fa-sign-out"></i> 退出房间
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import { getRoomInfo, startBattle, giveupBattle, dissolveRoom, leaveRoom, resetRoom } from '@/api/battle';
import { getRatingColor } from '@/common/rating-utils';

export default {
  name: 'BattleRoom',
  data() {
    return {
      roomId: '',
      room: {
        status: 0,
        hostId: '',
        hostUsername: '',
        challengerId: '',
        challengerUsername: '',
        problemId: null
      },
      problem: {},
      isHost: false,
      battleResult: null,
      hostSubmitCount: 0,
      challengerSubmitCount: 0,
      hostLatestStatus: null,
      challengerLatestStatus: null,
      pollingInterval: null,
      hasRedirectedToProblem: false // 标记是否已经跳转过题目页面
    };
  },
  computed: {
    isWinner() {
      if (!this.battleResult) return false;
      const userId = this.$store.getters.userInfo?.uid;
      return this.battleResult.winnerId === userId;
    },
    endReasonText() {
      if (!this.battleResult) return '';
      const isLoser = !this.isWinner;

      // 根据结束原因和是否获胜来显示不同的文本
      if (this.battleResult.endReason === 'giveup') {
        // 放弃对战：放弃者看到"你放弃了对战"，获胜者看到"对方放弃了对战"
        return isLoser ? '你放弃了对战' : '对方放弃了对战';
      }

      const reasonMap = {
        'ac': isLoser ? '对方通过了题目' : '你通过了题目',
        'timeout': isLoser ? '对方超时' : '你超时了'
      };
      return reasonMap[this.battleResult.endReason] || '对战结束';
    }
  },
  mounted() {
    this.initRoom();
  },
  watch: {
    '$route.params.roomId': {
      handler(newRoomId) {
        if (newRoomId && newRoomId !== this.roomId) {
          console.log('[BattleRoom] 房间ID变化:', this.roomId, '->', newRoomId);
          this.initRoom();
        }
      },
      immediate: false
    }
  },
  beforeDestroy() {
    this.stopPolling();
    // 清除房间记录（如果是对战结束或退出房间）
    const userId = this.$store.getters.userInfo?.uid;
    if (userId && this.battleResult) {
      // 对战结束了，清除记录
      sessionStorage.removeItem(`battle_room_${userId}`);
    }
  },
  methods: {
    initRoom() {
      // 停止旧的轮询
      this.stopPolling();

      // 获取新的房间ID
      this.roomId = this.$route.params.roomId;
      console.log('[BattleRoom] 初始化房间, roomId:', this.roomId);

      if (!this.roomId) {
        console.error('[BattleRoom] roomId 为空！路由参数可能有问题');
        this.$message.error('房间号获取失败，请重新进入');
        return;
      }

      // 保存用户当前房间ID到 sessionStorage
      const userId = this.$store.getters.userInfo?.uid;
      if (userId) {
        sessionStorage.setItem(`battle_room_${userId}`, this.roomId);
        console.log('[BattleRoom] 更新sessionStorage:', `battle_room_${userId} =`, this.roomId);
      }

      // 重置对战结果
      this.battleResult = null;
      this.hasRedirectedToProblem = false;

      this.loadRoomInfo();
      this.startPolling();
    },
    getRatingColor(rating) {
      return getRatingColor(rating);
    },
    async loadRoomInfo() {
      try {
        console.log('[BattleRoom] 开始加载房间信息, roomId:', this.roomId);
        const res = await getRoomInfo(this.roomId);

        console.log('[BattleRoom] 房间信息响应:', res.data);

        // 房间不存在（被解散或删除）
        if (res.data.code === 1) {
          this.stopPolling();
          const errorMsg = res.data.msg || '未知错误';
          console.error('[BattleRoom] 加载房间信息失败:', errorMsg);

          // 提示用户房间已解散
          this.$message.warning(errorMsg);
          // 清除 sessionStorage
          const userId = this.$store.getters.userInfo?.uid;
          if (userId) {
            sessionStorage.removeItem(`battle_room_${userId}`);
          }
          this.$router.push({ name: 'BattleHome' });
          return;
        }

        if (res.data.code === 0) {
          const data = res.data.data;
          const newRoom = data.room;
          const userId = this.$store.getters.userInfo?.uid;
          this.isHost = newRoom.hostId === userId;

          // 如果响应中包含题目信息，总是更新题目信息
          if (data.problem) {
            this.problem = data.problem;
          }

          // 检查挑战者是否退出（用于调试）
          if (this.room.challengerUsername && !newRoom.challengerUsername) {
            console.log('挑战者已退出房间');
            this.$message.warning('挑战者已退出房间');
          }

          // 检查房间状态变化
          if (this.room.status !== newRoom.status) {
            // 房间状态发生变化，更新状态
            if (newRoom.status === 2) {
              // 对战结束，设置结果
              this.battleResult = {
                winnerId: newRoom.winnerId,
                endReason: newRoom.endReason
              };
              // 停止轮询
              this.stopPolling();
              // 重置跳转标志
              this.hasRedirectedToProblem = false;
            } else if (newRoom.status === 1 && this.room.status === 0) {
              // 从等待变为对战中

              // 先更新房间信息（包含problemId）
              this.room = { ...this.room, ...newRoom };

              // 不再自动跳转，让用户手动点击"进入题目"按钮
              // 房主和挑战者都留在对战房间页面
              this.stopPolling();
              this.startPolling();

              return; // 提前返回，避免下面的重复更新
            }
          }

          // 更新房间信息
          this.room = { ...this.room, ...newRoom };

          // 注意：题目信息已经在响应中获取，不需要再次请求
        }
      } catch (error) {
        // 处理网络错误 - 轮询期间完全静默处理，避免控制台报错
        if (error.response) {
          const status = error.response.status;
          if (status === 404 || status === 400) {
            // 房间不存在或已被解散
            this.stopPolling();
            // 提示用户房间已解散
            this.$message.warning('房间已被房主解散');
            // 清除 sessionStorage
            const userId = this.$store.getters.userInfo?.uid;
            if (userId) {
              sessionStorage.removeItem(`battle_room_${userId}`);
            }
            this.$router.push({ name: 'BattleHome' });
          }
          // 其他错误完全静默处理，不输出日志
        }
        // 网络错误完全静默处理，不输出日志
      }
    },

    async loadProblemInfo() {
      // 从HOJ加载题目信息
      if (!this.room.problemId) {
        return;
      }

      try {
        const res = await this.$http.get(`/api/problem/${this.room.problemId}`);
        if (res.data.code === 0) {
          this.problem = res.data.data;
        }
      } catch (error) {
        // 静默处理错误，不影响对战功能
        // 题目信息加载失败不影响对战进行
      }
    },

    async handleStartBattle() {
      try {
        const res = await startBattle({ roomId: this.roomId });
        if (res.data.code === 0) {
          this.$message.success('对战开始！');
          // 从响应中获取房间和题目信息
          this.room = res.data.data.room;
          this.problem = res.data.data.problem;
          this.hasRedirectedToProblem = true; // 标记已跳转

          // 不再自动跳转到题目页面，让用户手动点击"进入题目"按钮
        } else {
          this.$message.error(res.data.msg || '开始对战失败');
        }
      } catch (error) {
        this.$message.error('开始对战失败');
      }
    },

    async handleGiveup() {
      this.$confirm('确定要放弃对战吗？这将判定对方获胜', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await giveupBattle({ roomId: this.roomId });
          if (res.data.code === 0) {
            this.battleResult = res.data.data;
            this.room.status = 2;
          }
        } catch (error) {
          this.$message.error('操作失败');
        }
      });
    },

    async handleDissolveRoom() {
      this.$confirm('确定要解散房间吗？这将删除房间且无法恢复', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await dissolveRoom({ roomId: this.roomId });
          if (res.data.code === 0) {
            this.$message.success('房间已解散');
            // 清除 sessionStorage 中的房间记录
            const userId = this.$store.getters.userInfo?.uid;
            if (userId) {
              sessionStorage.removeItem(`battle_room_${userId}`);
            }
            this.stopPolling();
            this.$router.push({ name: 'BattleHome' });
          } else {
            this.$message.error(res.data.msg || '解散房间失败');
          }
        } catch (error) {
          this.$message.error('解散房间失败');
        }
      });
    },

    async handleLeaveRoom() {
      // 根据是否是房主显示不同的提示
      let confirmMessage = '确定要退出房间吗？';
      if (this.isHost) {
        if (this.room.challengerUsername) {
          confirmMessage = '你是房主，退出后房主身份将转移给对方。确定要退出吗？';
        } else {
          confirmMessage = '你是房主，退出后房间将被解散。确定要退出吗？';
        }
      }

      this.$confirm(confirmMessage, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await leaveRoom({ roomId: this.roomId });
          if (res.data.code === 0) {
            this.$message.success('已退出房间');
            // 清除 sessionStorage 中的房间记录
            const userId = this.$store.getters.userInfo?.uid;
            if (userId) {
              sessionStorage.removeItem(`battle_room_${userId}`);
            }
            this.stopPolling();
            this.$router.push({ name: 'BattleHome' });
          } else {
            this.$message.error(res.data.msg || '退出房间失败');
          }
        } catch (error) {
          this.$message.error('退出房间失败');
        }
      });
    },

    startPolling() {
      // 根据房间状态设置不同的轮询频率
      // 等待中：5秒轮询一次
      // 对战中：3秒轮询一次（需要及时检测AC）
      this.pollingInterval = setInterval(() => {
        this.loadRoomInfo();
      }, this.room.status === 1 ? 3000 : 5000);
    },

    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval);
        this.pollingInterval = null;
      }
    },

    copyRoomId() {
      // 兼容多种复制方式
      const textToCopy = this.roomId;

      // 方式1: 使用 Clipboard API (现代浏览器)
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(textToCopy).then(() => {
          this.$message.success('房间号已复制');
        }).catch(() => {
          this.fallbackCopy(textToCopy);
        });
      } else {
        // 方式2: 使用传统方法
        this.fallbackCopy(textToCopy);
      }
    },

    fallbackCopy(text) {
      // 创建临时文本域
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      try {
        // 执行复制命令
        const successful = document.execCommand('copy');
        if (successful) {
          this.$message.success('房间号已复制');
        } else {
          this.$message.error('复制失败，请手动复制');
        }
      } catch (err) {
        this.$message.error('复制失败，请手动复制');
      }

      // 移除临时元素
      document.body.removeChild(textArea);
    },

    goToProblem() {
      this.$router.push({
        name: 'ProblemDetails',
        params: { problemID: this.room.problemId },
        query: { battle: this.roomId }
      });
    },

    backToHome() {
      this.$router.push({ name: 'BattleHome' });
    },

    async rematch() {
      try {
        const userId = this.$store.getters.userInfo?.uid;
        if (!userId) {
          this.$message.error('获取用户信息失败');
          return;
        }

        const res = await resetRoom({ roomId: this.roomId, userId: userId });
        if (res.data.code === 0) {
          this.$message.success('房间已重置，可以开始新一局');
          // 重置本地状态
          this.battleResult = null;
          this.room.status = 0;
          this.problem = {};
          this.hostSubmitCount = 0;
          this.challengerSubmitCount = 0;
          this.hostLatestStatus = null;
          this.challengerLatestStatus = null;
          // 重新开始轮询
          this.startPolling();
        } else {
          this.$message.error(res.data.msg || '重置房间失败');
        }
      } catch (error) {
        this.$message.error('重置房间失败');
      }
    },

    getDifficultyText(difficulty) {
      const map = { 1: '入门', 2: '入门', 3: '中等', 4: '困难', 5: '困难' };
      return map[difficulty] || '未知';
    },

    getStatusText(status) {
      const map = {
        0: 'Pending',
        1: 'Accepted',
        '-2': 'Compilation Error',
        '-3': 'Wrong Answer',
        '-4': 'Runtime Error',
        '-5': 'Time Limit Exceeded'
      };
      return map[status] || 'Unknown';
    },

    getStatusTagType(status) {
      if (status === 1) return 'success';
      if (status === 0) return 'info';
      return 'danger';
    }
  }
};
</script>

<style scoped>
.battle-room-container {
  padding: 20px;
  max-width: 1000px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.waiting-content {
  padding: 30px 0;
}

.player-section {
  display: flex;
  justify-content: space-around;
  align-items: center;
  margin-bottom: 40px;
}

.player-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  background: #f5f7fa;
}

.player-avatar {
  margin-bottom: 15px;
  color: #909399;
}

.player-avatar.empty {
  color: #c0c4cc;
}

.player-info h3 {
  margin: 10px 0;
  font-size: 18px;
}

.vs-badge {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  padding: 0 30px;
}

.action-section {
  text-align: center;
  margin-top: 30px;
}

.battle-content {
  padding: 20px 0;
}

.problem-section {
  text-align: center;
  padding: 20px;
  background: #f0f9ff;
  border-radius: 8px;
}

.players-status {
  display: flex;
  justify-content: space-around;
  margin-top: 20px;
}

.status-card {
  flex: 1;
  padding: 20px;
  border-radius: 8px;
  background: #f5f7fa;
  margin: 0 10px;
  border: 2px solid transparent;
}

.status-card.winner {
  border-color: #67C23A;
  background: #f0f9ff;
}

.status-header {
  text-align: center;
  margin-bottom: 15px;
}

.player-name {
  margin: 10px 0;
  font-size: 16px;
  font-weight: 600;
}

.player-avatar-icon {
  color: #909399;
}

.status-body {
  text-align: center;
}

.submit-count,
.latest-status {
  margin: 5px 0;
  font-size: 14px;
}

.result-content {
  text-align: center;
  padding: 50px 20px;
}

.result-actions {
  margin-top: 30px;
  display: flex;
  justify-content: center;
  gap: 15px;
}

.winner-section i,
.loser-section i {
  margin-bottom: 20px;
}

.winner-section h2 {
  color: #67C23A;
  margin-bottom: 10px;
}

.loser-section h2 {
  color: #909399;
  margin-bottom: 10px;
}

@media screen and (max-width: 768px) {
  .player-section {
    flex-direction: column;
    gap: 20px;
  }

  .vs-badge {
    padding: 10px 0;
  }

  .players-status {
    flex-direction: column;
    gap: 15px;
  }

  .status-card {
    margin: 0;
  }
}
</style>
