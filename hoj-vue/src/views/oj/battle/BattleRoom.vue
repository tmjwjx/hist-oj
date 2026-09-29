<template>
  <div class="battle-room-container">
    <!-- 等待状态 -->
    <el-card v-if="room.status === 0" class="waiting-card">
      <div slot="header" class="card-header">
        <span>{{ $t('m.Battle_Room_No') }}{{ roomId }}</span>
        <div class="header-actions">
          <el-button type="text" @click="copyRoomId">
            <i class="fa fa-copy"></i> {{ $t('m.Copy') }}
          </el-button>
          <el-button v-if="isHost" type="text" style="color: #F56C6C;" @click="handleDissolveRoom">
            <i class="fa fa-times"></i> {{ $t('m.Battle_Dissolve_Room') }}
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
              <el-tag type="success">{{ $t('m.Battle_Host') }}</el-tag>
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
              <h3 v-else>{{ $t('m.Battle_Waiting_Player') }}</h3>
              <div v-if="room.challengerUsername">
                <el-tag type="warning">{{ $t('m.Battle_Challenger') }}</el-tag>
                <el-tag v-if="room.challengerReady" type="success" style="margin-left: 5px;">{{ $t('m.Battle_Ready') }}</el-tag>
                <el-tag v-else type="info" style="margin-left: 5px;">{{ $t('m.Battle_Not_Ready') }}</el-tag>
              </div>
            </div>
          </div>
        </div>

        <div v-if="isHost" class="action-section">
          <el-button
            type="primary"
            size="large"
            :disabled="!room.challengerUsername || !room.challengerReady"
            @click="handleStartBattle"
          >
            <i class="fa fa-play"></i> {{ $t('m.Battle_Start') }}
          </el-button>
          <el-alert
            v-if="room.challengerUsername && !room.challengerReady"
            :title="$t('m.Battle_Waiting_Ready')"
            type="warning"
            :closable="false"
            show-icon
            style="margin-top: 10px;"
          ></el-alert>
          <el-button
            type="info"
            size="large"
            style="margin-left: 10px;"
            @click="handleLeaveRoom"
          >
            <i class="fa fa-sign-out-alt"></i> {{ $t('m.Battle_Leave_Room') }}
          </el-button>
        </div>
        <div v-else class="action-section">
          <el-button
            v-if="room.challengerUsername"
            :type="room.challengerReady ? 'warning' : 'success'"
            size="large"
            @click="handleReady"
          >
            <i class="fa" :class="room.challengerReady ? 'fa-times' : 'fa-check'"></i>
            {{ room.challengerReady ? $t('m.Battle_Cancel_Ready') : $t('m.Battle_Ready') }}
          </el-button>
          <el-alert
            v-if="room.challengerReady"
            :title="$t('m.Battle_Ready_Waiting_Host')"
            type="success"
            :closable="false"
            show-icon
            style="margin-top: 10px;"
          ></el-alert>
          <el-button
            type="danger"
            size="large"
            style="margin-top: 15px;"
            @click="handleLeaveRoom"
          >
            <i class="fa fa-sign-out-alt"></i> {{ $t('m.Battle_Leave_Room') }}
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 对战状态 -->
    <el-card v-else-if="room.status === 1" class="battle-card">
      <div slot="header" class="card-header">
        <span>{{ $t('m.Battle_In_Progress') }}</span>
        <div class="header-actions">
          <el-button v-if="isHost" type="text" style="color: #F56C6C;" @click="handleDissolveRoom">
            <i class="fa fa-times"></i> {{ $t('m.Battle_Dissolve_Room') }}
          </el-button>
          <el-button type="danger" size="small" @click="handleGiveup">
            <i class="fa fa-flag"></i> {{ $t('m.Battle_Give_Up') }}
          </el-button>
        </div>
      </div>

      <div class="battle-content">
        <div class="problem-section">
          <el-alert
            :title="$t('m.Battle_Problem_Is') + problem.title"
            type="success"
            :closable="false"
            show-icon
          ></el-alert>
          <el-button
            type="primary"
            style="margin-top: 15px;"
            @click="goToProblem"
          >
            <i class="fa fa-code"></i> {{ $t('m.Battle_Go_To_Problem') }}
          </el-button>
        </div>

        <el-divider></el-divider>

        <div class="players-status">
          <div class="status-card" :class="{ winner: battleResult && battleResult.winnerId === room.hostId }">
            <div class="status-header">
              <i class="fa fa-user-circle fa-3x player-avatar-icon"></i>
              <div class="player-name">{{ room.hostUsername }}</div>
              <el-tag v-if="battleResult && battleResult.winnerId === room.hostId" type="success">{{ $t('m.Battle_Win') }}</el-tag>
            </div>
            <div class="status-body">
              <div class="submit-count">{{ $t('m.Battle_Submit_Count') }}: {{ hostSubmitCount }}</div>
              <div class="latest-status">
                {{ $t('m.Battle_Latest_Status') }}
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
              <el-tag v-if="battleResult && battleResult.winnerId === room.challengerId" type="success">{{ $t('m.Battle_Win') }}</el-tag>
            </div>
            <div class="status-body">
              <div class="submit-count">{{ $t('m.Battle_Submit_Count') }}: {{ challengerSubmitCount }}</div>
              <div class="latest-status">
                {{ $t('m.Battle_Latest_Status') }}
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
          <h2>{{ $t('m.Battle_Congrats_Win') }}</h2>
          <p>{{ endReasonText }}</p>
        </div>
        <div v-else class="loser-section">
          <i class="fa fa-heart-broken fa-5x" style="color: #909399;"></i>
          <h2>{{ $t('m.Battle_Lost') }}</h2>
          <p>{{ endReasonText }}</p>
        </div>

        <div class="result-actions">
          <el-button type="success" @click="rematch">
            <i class="fa fa-redo"></i> {{ $t('m.Battle_Rematch') }}
          </el-button>
          <el-button v-if="isHost" type="danger" @click="handleDissolveRoom">
            <i class="fa fa-times"></i> {{ $t('m.Battle_Dissolve_Room') }}
          </el-button>
          <el-button type="danger" @click="handleLeaveRoom">
            <i class="fa fa-sign-out"></i> {{ $t('m.Battle_Leave_Room') }}
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import { getRoomInfo, startBattle, giveupBattle, dissolveRoom, leaveRoom, resetRoom, readyBattle } from '@/api/battle';
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
        challengerReady: false, // 挑战者准备状态
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
        return isLoser ? this.$t('m.Battle_You_Gave_Up_Battle') : this.$t('m.Battle_Opponent_Gave_Up_Battle');
      }

      const reasonMap = {
        'ac': isLoser ? this.$t('m.Battle_Opponent_Solved') : this.$t('m.Battle_You_Solved'),
        'timeout': isLoser ? this.$t('m.Battle_Opponent_Timeout') : this.$t('m.Battle_You_Timeout')
      };
      return reasonMap[this.battleResult.endReason] || this.$t('m.Battle_Finished');
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
    console.log('[BattleRoom] 组件销毁,停止轮询');
    this.stopPolling();
    // 清除房间记录（如果是对战结束或退出房间）
    const userId = this.$store.getters.userInfo?.uid;
    if (userId) {
      // 无论如何都清除记录，防止残留
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
        this.$message.error(this.$t('m.Battle_Room_Load_Failed'));
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
      // 如果组件已经销毁或正在销毁过程中，不要执行请求
      if (this._isDestroyed) {
        console.log('[BattleRoom] 组件已销毁，跳过房间信息加载');
        return;
      }

      try {
        console.log('[BattleRoom] 开始加载房间信息, roomId:', this.roomId);
        const res = await getRoomInfo(this.roomId);

        // 如果在请求过程中组件被销毁了，直接返回
        if (this._isDestroyed) {
          console.log('[BattleRoom] 请求返回时组件已销毁');
          return;
        }

        console.log('[BattleRoom] 房间信息响应:', res.data);

        // 房间不存在（被解散或删除）
        if (res.data.code === 1) {
          this.stopPolling();
          const errorMsg = res.data.msg || this.$t('m.Unknown');
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
            this.$message.warning(this.$t('m.Battle_Challenger_Left'));
          }

          // 检查挑战者状态变化（加入或退出）- 需要调整轮询频率
          const challengerChanged = this.room.challengerUsername !== newRoom.challengerUsername;
          // 检查房间状态变化
          const statusChanged = this.room.status !== newRoom.status;

          if (statusChanged) {
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
          } else if (challengerChanged && this.room.status === 0) {
            // 挑战者状态变化（加入或退出）且处于等待状态 - 重新启动轮询以应用新的频率
            console.log('[BattleRoom] 挑战者状态变化，重新启动轮询以调整频率');
            this.stopPolling();
            this.startPolling();
          }

          // 更新房间信息
          this.room = { ...this.room, ...newRoom };

          // 注意：题目信息已经在响应中获取，不需要再次请求
        }
      } catch (error) {
        // 如果组件已经销毁，不处理错误
        if (this._isDestroyed) {
          console.log('[BattleRoom] 请求出错时组件已销毁');
          return;
        }

        // 处理网络错误 - 轮询期间完全静默处理，避免控制台报错
        if (error.response) {
          const status = error.response.status;
          if (status === 404 || status === 400) {
            // 房间不存在或已被解散
            this.stopPolling();
            // 提示用户房间已解散
            this.$message.warning(this.$t('m.Battle_Room_Dissolved_By_Host'));
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
          this.$message.success(this.$t('m.Battle_Started'));
          // 从响应中获取房间和题目信息
          this.room = res.data.data.room;
          this.problem = res.data.data.problem;
          this.hasRedirectedToProblem = true; // 标记已跳转

          // 立即重新启动轮询,使用对战的频率(3秒)
          this.stopPolling();
          this.startPolling();
        } else {
          this.$message.error(res.data.msg || this.$t('m.Battle_Start_Failed'));
        }
      } catch (error) {
        this.$message.error(this.$t('m.Battle_Start_Failed'));
      }
    },

    async handleReady() {
      // 挑战者准备/取消准备 (乐观更新 - 立即响应)
      try {
        const userId = this.$store.getters.userInfo?.uid;
        const oldReadyState = this.room.challengerReady;
        const newReadyState = !oldReadyState;

        // 乐观更新: 立即更新本地状态,提供即时反馈
        this.room.challengerReady = newReadyState;
        this.$message.success(newReadyState ? this.$t('m.Battle_Ready') : this.$t('m.Battle_Ready_Cancelled'));

        // 异步调用 API
        const res = await readyBattle({
          roomId: this.roomId,
          userId: userId,
          ready: newReadyState
        });

        // 检查后端响应
        if (res.data.code !== 0) {
          // 失败时回滚状态
          this.room.challengerReady = oldReadyState;
          this.$message.error(res.data.msg || this.$t('m.Battle_Operate_Failed'));
        }
        // 成功则保持乐观更新的状态,无需额外操作
      } catch (error) {
        // 失败时回滚状态
        const oldReadyState = this.room.challengerReady;
        this.room.challengerReady = !oldReadyState;
        this.$message.error(this.$t('m.Network_Error'));
      }
    },

    async handleGiveup() {
      this.$confirm(this.$t('m.Battle_Give_Up_Confirm'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await giveupBattle({ roomId: this.roomId });
          if (res.data.code === 0) {
            this.battleResult = res.data.data;
            this.room.status = 2;
          }
        } catch (error) {
          this.$message.error(this.$t('m.Battle_Operate_Failed'));
        }
      });
    },

    async handleDissolveRoom() {
      this.$confirm(this.$t('m.Battle_Dissolve_Confirm'), this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await dissolveRoom({ roomId: this.roomId });
          if (res.data.code === 0) {
            this.$message.success(this.$t('m.Battle_Room_Dissolved'));
            // 清除 sessionStorage 中的房间记录
            const userId = this.$store.getters.userInfo?.uid;
            if (userId) {
              sessionStorage.removeItem(`battle_room_${userId}`);
            }
            this.stopPolling();
            this.$router.push({ name: 'BattleHome' });
          } else {
            this.$message.error(res.data.msg || this.$t('m.Battle_Dissolve_Failed'));
          }
        } catch (error) {
          this.$message.error(this.$t('m.Battle_Dissolve_Failed'));
        }
      });
    },

    async handleLeaveRoom() {
      // 根据是否是房主显示不同的提示
      let confirmMessage = this.$t('m.Battle_Leave_Confirm');
      if (this.isHost) {
        if (this.room.challengerUsername) {
          confirmMessage = this.$t('m.Battle_Leave_Host_Transfer');
        } else {
          confirmMessage = this.$t('m.Battle_Leave_Host_Dissolve');
        }
      }

      this.$confirm(confirmMessage, this.$t('m.Prompt'), {
        confirmButtonText: this.$t('m.Confirm'),
        cancelButtonText: this.$t('m.Cancel'),
        type: 'warning'
      }).then(async () => {
        try {
          const res = await leaveRoom({ roomId: this.roomId });
          if (res.data.code === 0) {
            this.$message.success(this.$t('m.Battle_Left_Room'));
            // 清除 sessionStorage 中的房间记录
            const userId = this.$store.getters.userInfo?.uid;
            if (userId) {
              sessionStorage.removeItem(`battle_room_${userId}`);
            }
            this.stopPolling();
            this.$router.push({ name: 'BattleHome' });
          } else {
            this.$message.error(res.data.msg || this.$t('m.Battle_Leave_Failed'));
          }
        } catch (error) {
          this.$message.error(this.$t('m.Battle_Leave_Failed'));
        }
      });
    },

    startPolling() {
      // 如果组件已销毁，不要启动轮询
      if (this._isDestroyed) {
        console.log('[BattleRoom] 组件已销毁，不启动轮询');
        return;
      }

      // 先停止旧的轮询
      this.stopPolling();

      // 根据房间状态和挑战者状态设置动态轮询频率
      // 等待中且有挑战者(status=0 && challenger存在)：500ms - 快速检测挑战者进出/准备状态变化
      // 等待中但无挑战者(status=0 && challenger不存在)：1秒 - 正常等待加入
      // 对战中(status=1)：2秒 - 需要及时检测AC
      // 已结束(status=2)：不轮询 - 不需要更新
      if (this.room.status === 2) {
        console.log('[BattleRoom] 对战已结束，不启动轮询');
        return;
      }

      // 动态轮询频率：等待状态且有挑战者时使用更快的轮询来快速检测进出变化
      let interval;
      if (this.room.status === 1) {
        interval = 2000; // 对战中：2秒
      } else if (this.room.status === 0 && this.room.challengerUsername) {
        interval = 500;  // 等待中且有挑战者：500ms - 快速检测进出/准备状态
      } else {
        interval = 1000; // 等待中但无挑战者：1秒
      }
      this.pollingInterval = setInterval(() => {
        // 在执行前再次检查组件是否已销毁
        if (this._isDestroyed) {
          console.log('[BattleRoom] 轮询执行时发现组件已销毁，停止轮询');
          this.stopPolling();
          return;
        }
        this.loadRoomInfo();
      }, interval);

      console.log('[BattleRoom] 启动轮询,间隔:', interval + 'ms, 状态:', this.getRoomStatusText(this.room.status));
    },

    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval);
        this.pollingInterval = null;
        console.log('[BattleRoom] 停止轮询');
      }
    },

    copyRoomId() {
      // 兼容多种复制方式
      const textToCopy = this.roomId;

      // 方式1: 使用 Clipboard API (现代浏览器)
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(textToCopy).then(() => {
          this.$message.success(this.$t('m.Battle_Room_ID_Copied'));
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
          this.$message.success(this.$t('m.Battle_Room_ID_Copied'));
        } else {
          this.$message.error(this.$t('m.Battle_Copy_Failed'));
        }
      } catch (err) {
        this.$message.error(this.$t('m.Battle_Copy_Failed'));
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
          this.$message.error(this.$t('m.Battle_User_Info_Failed'));
          return;
        }

        const res = await resetRoom({ roomId: this.roomId, userId: userId });
        if (res.data.code === 0) {
          this.$message.success(this.$t('m.Battle_Room_Reset'));
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
          this.$message.error(res.data.msg || this.$t('m.Battle_Reset_Failed'));
        }
      } catch (error) {
        this.$message.error(this.$t('m.Battle_Reset_Failed'));
      }
    },

    getDifficultyText(difficulty) {
      const map = { 1: this.$t('m.Difficulty_Entry'), 2: this.$t('m.Difficulty_Entry'), 3: this.$t('m.Difficulty_Medium'), 4: this.$t('m.Difficulty_Hard'), 5: this.$t('m.Difficulty_Hard') };
      return map[difficulty] || this.$t('m.Unknown');
    },

    getRoomStatusText(status) {
      const map = {
        0: this.$t('m.Battle_Status_Waiting'),
        1: this.$t('m.Battle_Status_Battling'),
        2: this.$t('m.Battle_Status_Finished')
      };
      return map[status] || this.$t('m.Unknown');
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
  max-width: 1200px;
  margin: 0 auto;
  min-height: calc(100vh - 60px);
}

.battle-room-container /deep/ .el-card {
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  background: #fff;
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
  border-radius: 4px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
}

.player-avatar {
  margin-bottom: 15px;
  color: #606266;
}

.player-avatar.empty {
  color: #c0c4cc;
}

.player-info h3 {
  margin: 10px 0;
  font-size: 18px;
  font-weight: 600;
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
  background: #ecf5ff;
  border-radius: 4px;
  border: 1px solid #d9ecff;
}

.players-status {
  display: flex;
  justify-content: space-around;
  margin-top: 20px;
}

.status-card {
  flex: 1;
  padding: 20px;
  border-radius: 4px;
  background: #f5f7fa;
  margin: 0 10px;
  border: 1px solid #e4e7ed;
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
  .battle-room-container {
    padding: 10px;
  }

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
