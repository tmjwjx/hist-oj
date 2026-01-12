<template>
  <div class="student-discussion">
    <div class="discussion-header">
      <h3>{{ $t('m.Classroom_Discussion') }}</h3>
      <el-button type="danger" icon="el-icon-delete" @click="clearMessages">
        清屏
      </el-button>
    </div>

    <div class="picked-student" v-if="pickedStudent">
      <el-alert
        :title="$t('m.Picked_Student') + ': ' + (pickedStudent.pickedUser?.nickname || '-')"
        type="success"
        :closable="false"
      />
    </div>

    <div class="message-list" ref="messageList">
      <div v-for="msg in messages" :key="msg.id" class="message-item">
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

export default {
  name: 'Discussion',
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
      imageUploadProgress: 0
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
    this.startPolling()
  },
  beforeDestroy() {
    this.stopPolling()
  },
  methods: {
    async loadMessages() {
      this.loading = true
      try {
        const res = await this.$store.dispatch('classroom/getMessages', {
          classroomId: this.classroomId,
          params: { page: 1, limit: 50 }
        })
        if (res.code === 200) {
          this.messages = res.data || []
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        }
      } catch (error) {
        this.$message.error(this.$t('m.Load_Failed'))
      } finally {
        this.loading = false
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
    startPolling() {
      this.pollingTimer = setInterval(() => {
        this.loadMessages()
      }, 5000)
    },
    stopPolling() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
      }
    },
    getSenderName(msg) {
      if (msg.sender && msg.sender.nickname) {
        return msg.sender.nickname
      }
      if (msg.sender && msg.sender.username) {
        return msg.sender.username
      }
      if (msg.studentInfo && msg.studentInfo.realName) {
        return msg.studentInfo.realName
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
      const token = localStorage.getItem('token')
      if (token) {
        this.$store.dispatch('user/getInfo').then(res => {
          if (res && res.data && res.data.data) {
            this.currentUserId = res.data.data.uuid
          }
        }).catch(() => {
          const userInfo = localStorage.getItem('user')
          if (userInfo) {
            try {
              this.currentUserId = JSON.parse(userInfo).uuid
            } catch (e) {
              console.error('解析用户信息失败', e)
            }
          }
        })
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
    },
    async clearMessages() {
      try {
        await this.$confirm('确定要清屏吗？这将删除您在该班级的所有消息记录！', '警告', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })

        const res = await this.$store.dispatch('classroom/clearMessages', this.classroomId)
        if (res.code === 200) {
          this.$message.success('清屏成功')
          this.loadMessages()
        } else {
          this.$message.error(res.message || '清屏失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('清屏失败')
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
  transition: background-color 0.2s;
}

.emoji-item:hover {
  background-color: #F5F7FA;
}
</style>
