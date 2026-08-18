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
            :lg="6"
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
          title: '套卷练习',
          description: '公开套卷在线练习，支持按题查看标准答案',
          icon: 'el-icon-reading',
          iconColor: '#1A73E8',
          url: '/practice'
        },
        {
          title: '算法知识点航海图',
          description: '知识点 + 题目技能树，按前置依赖解锁学习路线',
          icon: 'fa fa-sitemap',
          iconColor: '#36B37E',
          url: '/learning-map'
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
  width: 100%;
  box-sizing: border-box;
  padding: 20px;
  margin: 0;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.toolbox-card {
  width: 100%;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.toolbox-header {
  background: #fff;
  border-bottom: 2px solid #409eff;
  padding: 20px;
}

.toolbox-title-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.toolbox-logo {
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

.toolbox-logo i {
  text-shadow: none;
  filter: none;
}

.toolbox-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.toolbox-content {
  padding: 24px 20px;
  background: #fff;
}

.toolbox-content /deep/ .el-row {
  display: flex;
  flex-wrap: wrap;
}

.tool-item-col {
  display: flex;
}

.tool-item-col /deep/ .el-card {
  width: 100%;
}

.tool-item-col {
  margin-bottom: 20px;
}

.tool-card {
  cursor: pointer;
  transition: all 0.2s;
  border-radius: 4px;
  height: 100%;
  border: 1px solid #e4e7ed;
}

.tool-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.tool-icon {
  text-align: center;
  font-size: 48px;
  margin-bottom: 12px;
  transition: transform 0.2s ease;
}

.tool-icon i {
  font-size: 48px;
  text-shadow: none;
  filter: none;
}

.tool-card:hover .tool-icon {
  transform: scale(1.05);
}

.tool-info {
  text-align: center;
}

.tool-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.tool-description {
  font-size: 13px;
  color: #606266;
  margin: 0;
  line-height: 1.6;
}

@media screen and (max-width: 768px) {
  .toolbox-container {
    padding: 10px;
  }

  .toolbox-title {
    font-size: 20px;
  }

  .tool-icon {
    font-size: 36px;
    margin-bottom: 10px;
  }

  .tool-icon i {
    font-size: 36px;
  }

  .tool-title {
    font-size: 15px;
  }

  .tool-description {
    font-size: 12px;
  }
}
</style>
