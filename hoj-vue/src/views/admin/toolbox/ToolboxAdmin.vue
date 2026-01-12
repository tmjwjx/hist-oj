<template>
  <div class="toolbox-admin-container">
    <el-card class="toolbox-card">
      <div slot="header" class="toolbox-header">
        <div class="toolbox-title-wrapper">
          <div class="toolbox-logo">
            <i class="fa fa-briefcase fa-lg"></i>
          </div>
          <span class="toolbox-title">管理员工具箱</span>
        </div>
      </div>
      <div class="toolbox-content">
        <el-row :gutter="20">
          <el-col
            :md="8"
            :sm="12"
            :xs="24"
            v-for="(tool, index) in tools"
            :key="index"
            class="tool-item-col"
          >
            <el-card
              class="tool-card"
              shadow="hover"
              @click.native="openTool(tool)"
            >
              <div class="tool-icon" :style="{ color: tool.iconColor || '#409eff' }">
                <i :class="tool.icon"></i>
              </div>
              <div class="tool-info">
                <h3 class="tool-title">{{ tool.title }}</h3>
                <p class="tool-description">{{ tool.description }}</p>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'ToolboxAdmin',
  components: {},
  data() {
    return {
      tools: [
        {
          title: '赛事报名系统管理',
          description: '管理赛事报名系统和报名信息',
          icon: 'fa fa-trophy',
          iconColor: '#FFD700',
          url: '/registration/admin.html'
        },
        {
          title: '班级管理',
          description: '查看和管理所有班级信息',
          icon: 'el-icon-s-operation',
          iconColor: '#67C23A',
          action: 'openClassroomManagement'
        },
        {
          title: '用户角色管理',
          description: '为用户分配教师/学生角色',
          icon: 'fa fa-user-plus',
          iconColor: '#409EFF',
          action: 'openUserRoleManagement'
        },
        {
          title: 'Rating 管理',
          description: '手动调整用户 Rating 和查看历史记录',
          icon: 'fa fa-line-chart',
          iconColor: '#67C23A',
          action: 'openRatingManagement'
        },
        {
          title: '判题终端',
          description: '代码自测与提交判题',
          icon: 'fa fa-terminal',
          iconColor: '#409EFF',
          action: 'openJudgeTerminal'
        },
        {
          title: '对战记录查询',
          description: '查询所有用户的对战记录',
          icon: 'fa fa-gamepad',
          iconColor: '#E6A23C',
          action: 'openBattleRecords'
        }
      ]
    };
  },
  methods: {
    openTool(tool) {
      if (tool.action) {
        // 执行特定动作
        if (tool.action === 'openClassroomManagement') {
          // 跳转到班级管理页面
          this.$router.push({ name: 'admin-classroom' });
        } else if (tool.action === 'openUserRoleManagement') {
          // 跳转到用户角色管理页面
          this.$router.push({ name: 'admin-user-role-management' });
        } else if (tool.action === 'openRatingManagement') {
          // 动态加载 RatingManagement 组件
          this.$router.push({ name: 'admin-rating' });
        } else if (tool.action === 'openJudgeTerminal') {
          // 跳转到判题终端页面
          this.$router.push({ name: 'admin-judge-terminal-page' });
        } else if (tool.action === 'openBattleRecords') {
          // 跳转到对战记录查询页面
          this.$router.push({ name: 'admin-battle-records' });
        }
      } else if (tool.url) {
        if (tool.url.startsWith('http')) {
          window.open(tool.url, '_blank');
        } else {
          window.location.href = tool.url;
        }
      }
    }
  }
};
</script>

<style scoped>
.toolbox-admin-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.toolbox-card {
  border-radius: 8px;
}

.toolbox-header {
  text-align: center;
}

.toolbox-title-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 15px;
}

.toolbox-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 50px;
  height: 50px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border-radius: 12px;
  box-shadow: 0 4px 15px rgba(245, 87, 108, 0.4);
  color: white;
  font-size: 24px;
}

.toolbox-title {
  font-size: 28px;
  font-weight: 600;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.toolbox-content {
  padding: 20px 0;
}

.tool-item-col {
  margin-bottom: 20px;
}

.tool-card {
  cursor: pointer;
  transition: all 0.3s;
  border-radius: 8px;
  height: 100%;
}

.tool-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 20px rgba(64, 158, 255, 0.3);
}

.tool-icon {
  text-align: center;
  font-size: 56px;
  color: #409eff;
  margin-bottom: 15px;
  transition: all 0.3s ease;
}

.tool-icon i {
  font-size: 56px;
  filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.15));
}

.tool-card:hover .tool-icon {
  transform: scale(1.1);
}

.tool-info {
  text-align: center;
}

.tool-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 10px 0;
}

.tool-description {
  font-size: 14px;
  color: #909399;
  margin: 0;
  line-height: 1.5;
}

@media screen and (max-width: 768px) {
  .toolbox-admin-container {
    padding: 10px;
  }

  .toolbox-title {
    font-size: 24px;
  }

  .tool-icon {
    font-size: 36px;
    margin-bottom: 10px;
  }

  .tool-icon i {
    font-size: 36px;
  }

  .tool-title {
    font-size: 16px;
  }

  .tool-description {
    font-size: 12px;
  }
}
</style>
