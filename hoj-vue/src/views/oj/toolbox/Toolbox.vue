<template>
  <div class="toolbox-container">
    <el-card class="toolbox-card">
      <div slot="header" class="toolbox-header">
        <div class="toolbox-title-wrapper">
          <div class="toolbox-logo">
            <i class="fa fa-briefcase fa-lg"></i>
          </div>
          <span class="toolbox-title">工具箱</span>
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
  name: 'Toolbox',
  data() {
    return {
      tools: [
        {
          title: '代码对战',
          description: '1v1实时代码对战，挑战对手',
          icon: 'fa fa-gamepad',
          iconColor: '#FF6B6B',
          url: '/battle'
        },
        {
          title: '班级管理',
          description: '教师班级管理，学生作业提交',
          icon: 'el-icon-s-operation',
          iconColor: '#67C23A',
          url: '/classroom'
        },
        {
          title: '赛事报名系统',
          description: '查看和管理赛事报名信息',
          icon: 'fa fa-trophy',
          iconColor: '#FFD700',
          url: '/registration/list'
        },
        {
          title: '算法知识点航海图',
          description: '知识点 + 题目技能树，按前置依赖解锁学习路线',
          icon: 'fa fa-sitemap',
          iconColor: '#36B37E',
          url: '/toolbox/learning-map'
        }
        // 未来可以在这里添加更多工具
        // {
        //   title: '工具名称',
        //   description: '工具描述',
        //   icon: 'fa fa-icon-name',
        //   iconColor: '#409eff',
        //   url: '/tool-path'
        // }
      ]
    };
  },
  methods: {
    openTool(tool) {
      if (tool.url) {
        if (tool.url.startsWith('http')) {
          window.open(tool.url, '_blank');
        } else {
          this.$router.push(tool.url);
        }
      }
    }
  }
};
</script>

<style scoped>
.toolbox-container {
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
  font-size: 24px;
}

.toolbox-logo i {
  text-shadow: none;
  filter: none;
}

.toolbox-title {
  font-size: 28px;
  font-weight: 600;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  text-shadow: none;
  filter: none;
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
  .toolbox-container {
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
