const ClassroomRoutes = {
  path: '/classroom',
  component: () => import('@/views/classroom/ClassroomHome'),
  redirect: '/classroom/index',
  children: [
    {
      path: 'index',
      name: 'ClassroomIndex',
      component: () => import('@/views/classroom/ClassroomIndex'),
      meta: { title: '班级' }
    },
    // 题库（独立页面，不绑定到班级）
    {
      path: 'question-bank',
      name: 'QuestionBank',
      component: () => import('@/views/classroom/teacher/QuestionBank'),
      meta: { title: '题库' }
    },
    // 教师端路由
    {
      path: 'teacher',
      name: 'TeacherDashboard',
      component: () => import('@/views/classroom/teacher/Dashboard'),
      meta: { title: '教师工作台' }
    },
    // 创建作业和作业详情 - 独立页面，不在 ClassroomDetail 下
    {
      path: 'teacher/homework/create/:classroomId',
      name: 'CreateHomework',
      component: () => import('@/views/classroom/teacher/CreateHomework'),
      meta: { title: '创建作业' }
    },
    {
      path: 'teacher/homework/:classroomId/:homeworkId',
      name: 'TeacherHomeworkDetail',
      component: () => import('@/views/classroom/teacher/HomeworkDetail'),
      meta: { title: '作业详情' }
    },
    {
      path: 'teacher/homework/:homeworkId/submission/:uid',
      name: 'StudentSubmissionDetail',
      component: () => import('@/views/classroom/teacher/StudentSubmissionDetail'),
      meta: { title: '学生提交详情' }
    },
    {
      path: 'teacher/classroom/:classroomId',
      name: 'TeacherClassroomDetail',
      component: () => import('@/views/classroom/teacher/ClassroomDetail'),
      meta: { title: '班级详情' },
      redirect: { name: 'TeacherHomework' },
      children: [
        {
          path: 'students',
          name: 'TeacherStudents',
          component: () => import('@/views/classroom/teacher/Students'),
          meta: { title: '学生管理' }
        },
        {
          path: 'checkin',
          name: 'TeacherCheckin',
          component: () => import('@/views/classroom/teacher/Checkin'),
          meta: { title: '签到管理' }
        },
        {
          path: 'homework',
          name: 'TeacherHomework',
          component: () => import('@/views/classroom/teacher/Homework'),
          meta: { title: '作业管理' }
        },
        {
          path: 'materials',
          name: 'TeacherMaterials',
          component: () => import('@/views/classroom/teacher/Materials'),
          meta: { title: '资料库' }
        },
        {
          path: 'discussion',
          name: 'TeacherDiscussion',
          component: () => import('@/views/classroom/teacher/Discussion'),
          meta: { title: '班级讨论' }
        }
      ]
    },
    // 学生端路由
    {
      path: 'student',
      name: 'StudentDashboard',
      component: () => import('@/views/classroom/student/Dashboard'),
      meta: { title: '我的班级' }
    },
    {
      path: 'student/classroom/:classroomId',
      name: 'StudentClassroomDetail',
      component: () => import('@/views/classroom/student/ClassroomDetail'),
      meta: { title: '班级详情' },
      redirect: { name: 'StudentHomework' },
      children: [
        {
          path: 'homework',
          name: 'StudentHomework',
          component: () => import('@/views/classroom/student/Homework'),
          meta: { title: '作业列表' }
        },
        {
          path: 'checkin',
          name: 'StudentCheckin',
          component: () => import('@/views/classroom/student/Checkin'),
          meta: { title: '签到' }
        },
        {
          path: 'materials',
          name: 'StudentMaterials',
          component: () => import('@/views/classroom/student/Materials'),
          meta: { title: '学习资料' }
        },
        {
          path: 'discussion',
          name: 'StudentDiscussion',
          component: () => import('@/views/classroom/student/Discussion'),
          meta: { title: '班级讨论' }
        }
      ]
    },
    {
      path: 'student/classroom/:classroomId/homework/:homeworkId',
      name: 'StudentHomeworkDetail',
      component: () => import('@/views/classroom/student/HomeworkDetail'),
      meta: { title: '作业详情' }
    }
  ]
}

export default ClassroomRoutes
