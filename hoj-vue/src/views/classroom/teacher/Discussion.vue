<template>
  <div class="student-discussion">
    <div class="discussion-header">
      <h3>{{ $t('m.Classroom_Discussion') }}</h3>
    </div>

    <div class="picked-student" v-if="pickedStudent">
      <el-alert
        :title="$t('m.Picked_Student') + ': ' + (pickedStudent.pickedUser?.nickname || '-')"
        type="success"
        :closable="false"
      />
    </div>

    <div class="message-list" ref="messageList">
      <div v-for="msg in messages" :key="msg.id" class="message-item" :class="{ 'system-message': msg.msgType === 'system' }">
        <div v-if="msg.msgType === 'system'" class="system-message-content">
          {{ msg.content }}
        </div>
        <template v-else>
          <div class="message-header">
            <span class="sender">{{ getSenderName(msg) }}</span>
            <div class="message-actions">
              <span class="time">{{ formatTime(msg.createdAt) }}</span>
              <el-button v-if="canRecallMessage(msg)" type="text" size="mini" icon="el-icon-back" @click="recallMessage(msg)">
                撤回
              </el-button>
            </div>
          </div>
          <div v-if="msg.msgType === 'text'" class="message-content" v-html="renderContent(msg.content)"></div>
          <div v-else class="message-image">
            <img :src="getImageUrl(msg.imageUrl)" alt="image" />
          </div>
        </template>
      </div>
    </div>

    <div class="message-input">
      <el-popover
        v-model="showEmojiPicker"
        placement="top-start"
        width="400"
        trigger="click"
      >
        <div class="emoji-picker">
          <span
            v-for="(emoji, index) in emojiList"
            :key="index"
            class="emoji-item"
            @click="insertEmoji(emoji)"
          >{{ emoji }}</span>
        </div>
        <el-button slot="reference" icon="el-icon-star-off" size="small">
          表情
        </el-button>
      </el-popover>

      <el-input
        v-model="newMessage"
        type="textarea"
        :rows="3"
        :placeholder="$t('m.Enter_Message')"
        @keyup.enter.native="handleEnterKey"
        @compositionstart.native="handleCompositionStart"
        @compositionend.native="handleCompositionEnd"
      />
      <div class="actions">
        <el-upload
          :action="uploadUrl"
          :headers="uploadHeaders"
          :data="{ classroomId: classroomId }"
          :show-file-list="false"
          :on-success="handleImageUploadSuccess"
          :on-error="handleImageUploadError"
          :on-progress="handleImageUploadProgress"
          :before-upload="beforeImageUpload"
          accept="image/*"
        >
          <el-button icon="el-icon-picture" :loading="imageUploading">
            {{ imageUploading ? `${$t('m.Uploading')} (${imageUploadProgress}%)` : $t('m.Upload_Image') }}
          </el-button>
        </el-upload>
        <el-button type="primary" @click="sendMessage" :loading="sending">
          {{ $t('m.Send') }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import moment from 'moment'
import realtimeSync from '@/mixins/realtimeSync'

import teacherAuth from '@/mixins/teacherAuth'
export default {
  name: 'Discussion',
  mixins: [realtimeSync, teacherAuth],
  props: {
    classroomId: [String, Number]
  },
  data() {
    return {
      loading: false,
      messages: [],
      newMessage: '',
      sending: false,
      pickedStudent: null,
      uploadUrl: '/rating-api/api/classroom/message/upload-image',
      uploadHeaders: {
        Authorization: localStorage.getItem('token') || ''
      },
      showEmojiPicker: false,
      emojiList: ['😀', '😁', '😂', '🤣', '😃', '😄', '😅', '😆', '😉', '😊', '😋', '😎', '😍', '😘', '🥰', '😗', '😙', '😚', '🙂', '🤗', '🤩', '🤔', '🤨', '😐', '😑', '😶', '🙄', '😏', '😣', '😥', '😮', '🤐', '😯', '😪', '😫', '😴', '😌', '😛', '😜', '😝', '🤤', '😒', '😓', '😔', '😕', '🙃', '🤑', '😲', '☹️', '🙁', '😖', '😞', '😟', '😤', '😢', '😭', '😦', '😧', '😨', '😩', '🤯', '😬', '😰', '😱', '🥵', '🥶', '😳', '🤪', '😵', '😡', '😠', '🤬', '😷', '🤒', '🤕', '🤢', '🤮', '🤧', '😇', '🤠', '🤡', '🥳', '🥴', '🥺', '🤥', '🤫', '🤭', '🧐', '🤓', '😈', '👍', '👎', '👏', '🙏', '💪', '🎉', '❤️', '💕', '💖', '💗', '💙', '💚', '💛', '🧡', '💜', '🖤', '💯', '✨', '⭐', '🌟', '💫', '🔥', '💥', '💢', '💦', '💨'],
      isComposing: false,
      justFinishedComposing: false,
      compositionEndTime: 0,
      currentUserId: null,
      imageUploading: false,
      imageUploadProgress: 0,
      // 实时同步配置
      realtimeSyncConfig: {
        enabled: true,
        interval: 3000, // 3秒刷新一次，减少闪烁
        syncFunction: 'loadMessages',
        immediate: true
      }
    }
  },
  watch: {
    classroomId: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.loadMessages()
        }
      }
    }
  },
  mounted() {
    this.getCurrentUserId()
    // realtimeSync mixin 会自动启动轮询
  },
  methods: {
    async loadMessages() {
      // 避免重复请求
      if (this.loading) return

      // 只在首次加载时显示 loading，轮询时不显示
      const isFirstLoad = this.messages.length === 0
      if (isFirstLoad) {
        this.loading = true
      }

      try {
        const res = await this.$store.dispatch('classroom/getMessages', {
          classroomId: this.classroomId,
          params: { page: 1, limit: 50 }
        })
        if (res.code === 200) {
          const newMessages = res.data || []

          // 检查数量是否变化
          if (newMessages.length !== this.messages.length) {
            const hasNewMessages = newMessages.length > this.messages.length
            this.messages = newMessages
            if (hasNewMessages) {
              this.$nextTick(() => {
                this.scrollToBottom()
              })
            }
            return
          }

          // 检查每条消息的 ID 是否都相同(避免深度对比整个对象)
          const currentIds = this.messages.map(m => m.id).join(',')
          const newIds = newMessages.map(m => m.id).join(',')

          if (currentIds !== newIds) {
            // ID 列表不同，说明有消息变化，需要更新
            this.messages = newMessages
          }
          // 如果 ID 列表相同，不更新数据，避免闪烁
        }
      } catch (error) {
        // 只在首次加载失败时提示错误
        if (isFirstLoad) {
          this.$message.error(this.$t('m.Load_Failed'))
        }
      } finally {
        if (isFirstLoad) {
          this.loading = false
        }
      }
    },
    async sendMessage() {
      if (!this.newMessage.trim()) {
        return
      }
      this.sending = true
      try {
        const res = await this.$store.dispatch('classroom/sendMessage', {
          classroomId: Number(this.classroomId),
          content: this.newMessage
        })
        if (res.code === 200) {
          this.newMessage = ''
          this.messages.push(res.data)
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        } else {
          this.$message.error(res.message || this.$t('m.Send_Failed'))
        }
      } catch (error) {
        this.$message.error(this.$t('m.Send_Failed'))
      } finally {
        this.sending = false
      }
    },
    handleEnterKey(e) {
      const now = Date.now()
      if (this.isComposing || (this.justFinishedComposing && now - this.compositionEndTime < 100)) {
        e.preventDefault()
        return
      }
      e.preventDefault()
      this.sendMessage()
    },
    handleCompositionStart() {
      this.isComposing = true
      this.justFinishedComposing = false
    },
    handleCompositionEnd() {
      this.isComposing = false
      this.justFinishedComposing = true
      this.compositionEndTime = Date.now()
    },
    handleImageUploadSuccess(response) {
      this.imageUploading = false
      this.imageUploadProgress = 0
      if (response.code === 200) {
        this.messages.push(response.data)
        this.$nextTick(() => {
          this.scrollToBottom()
        })
        this.$message.success(this.$t('m.Upload_Success'))
      } else {
        this.$message.error(response.message || this.$t('m.Upload_Failed'))
      }
    },
    handleImageUploadError(error) {
      this.imageUploading = false
      this.imageUploadProgress = 0
      let errorMessage = this.$t('m.Upload_Failed')

      if (error.message) {
        if (error.message.includes('Network Error') || error.message.includes('timeout')) {
          errorMessage = this.$t('m.Network_Error') || '网络错误，请检查网络连接'
        } else if (error.message.includes('413')) {
          errorMessage = this.$t('m.File_Too_Large') || '文件大小超出限制'
        } else if (error.message.includes('500')) {
          errorMessage = this.$t('m.Server_Error') || '服务器错误，请稍后重试'
        } else if (error.message.includes('401')) {
          errorMessage = this.$t('m.Unauthorized') || '未授权，请重新登录'
        }
      }

      this.$message.error(errorMessage)
    },
    handleImageUploadProgress(event) {
      this.imageUploading = true
      this.imageUploadProgress = Math.floor(event.percent)
    },
    beforeImageUpload(file) {
      const isImage = file.type.startsWith('image/')
      const isLt5M = file.size / 1024 / 1024 < 5

      if (!isImage) {
        this.$message.error(this.$t('m.Upload_Image_Only'))
        return false
      }
      if (!isLt5M) {
        this.$message.error(this.$t('m.Image_Size_Limit'))
        return false
      }
      return true
    },
    insertEmoji(emoji) {
      this.newMessage += emoji
      this.showEmojiPicker = false
    },
    renderContent(content) {
      if (!content) return ''
      // 将图片 markdown 转换为 img 标签，去掉 alt 文本
      return content.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" style="max-width: 300px; max-height: 300px; border-radius: 4px;" />')
    },
    scrollToBottom() {
      const container = this.$refs.messageList
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    },
    getSenderName(msg) {
      // 优先使用班级内的真实姓名
      if (msg.studentInfo && msg.studentInfo.realName) {
        return msg.studentInfo.realName
      }
      // 其次使用 HOJ 昵称
      if (msg.sender && msg.sender.nickname) {
        return msg.sender.nickname
      }
      // 最后使用 HOJ 用户名
      if (msg.sender && msg.sender.username) {
        return msg.sender.username
      }
      return this.$t('m.Unknown')
    },
    getImageUrl(url) {
      if (!url) return ''
      if (url.startsWith('/')) {
        return window.location.origin + url
      }
      return url
    },
    formatTime(time) {
      return moment(time).format('YYYY-MM-DD HH:mm:ss')
    },
    getCurrentUserId() {
      // 优先从 localStorage 获取用户信息
      const userInfo = localStorage.getItem('user')
      if (userInfo) {
        try {
          const user = JSON.parse(userInfo)
          this.currentUserId = user.uuid || user.userId
          return
        } catch (e) {
          // 解析失败，继续尝试其他方式
        }
      }

      // 备选方案：从 store 获取
      const storeState = this.$store.state
      if (storeState.user && storeState.user.userInfo) {
        this.currentUserId = storeState.user.userInfo.uuid || storeState.user.userInfo.userId
      }
    },
    canRecallMessage(msg) {
      return msg.senderId === this.currentUserId
    },
    async recallMessage(msg) {
      try {
        await this.$confirm('确定要撤回这条消息吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })

        const res = await this.$store.dispatch('classroom/recallMessage', msg.id)
        if (res.code === 200) {
          this.$message.success('撤回成功')
          this.loadMessages()
        } else {
          this.$message.error(res.message || '撤回失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('撤回失败')
        }
      }
    }
  }
}
</script>

<style scoped>
.student-discussion {
  padding: 20px;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 200px);
}

.discussion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.discussion-header h3 {
  font-size: 20px;
  color: #409EFF;
  margin: 0;
}

.picked-student {
  margin-bottom: 20px;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
  background-color: #F5F7FA;
}

.message-item {
  margin-bottom: 15px;
}

.message-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 5px;
}

.sender {
  font-weight: bold;
  color: #409EFF;
}

.time {
  font-size: 12px;
  color: #909399;
}

.message-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.message-content {
  padding: 10px;
  background-color: #FFF;
  border-radius: 4px;
  white-space: pre-wrap;
  word-break: break-word;
}

.message-image img {
  max-width: 300px;
  max-height: 300px;
  border-radius: 4px;
}

.message-input {
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 10px;
  background-color: #FFF;
}

.message-input .actions {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
}

.emoji-picker {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 5px;
  max-height: 200px;
  overflow-y: auto;
}

.emoji-item {
  font-size: 24px;
  cursor: pointer;
  text-align: center;
  padding: 5px;
  border-radius: 4px;
  /* 移除 transition 避免轮询时闪烁 */
}

.emoji-item:hover {
  background-color: #F5F7FA;
}

.system-message {
  display: flex;
  justify-content: center;
  margin: 10px 0;
}

.system-message-content {
  background-color: #fff3cd;
  color: #856404;
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  border: 1px solid #ffeaa7;
}
</style>
