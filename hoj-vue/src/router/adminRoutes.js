// 引入 view 组件
const Login = () => import('@/views/admin/Login')
const Home = () => import('@/views/admin/Home')
const Dashboard = () => import('@/views/admin/Dashboard')
const User = () => import('@/views/admin/general/User')
const RatingManagement = () => import('@/views/admin/general/RatingManagement')
const BattleRecordsAdmin = () => import('@/views/admin/toolbox/BattleRecordsAdmin')
const PermissionDocs = () => import('@/views/admin/general/PermissionDocs')
const Announcement = () => import('@/views/admin/general/Announcement')
const SysNotice = () => import('@/views/admin/general/SysNotice')
const SystemConfig = () => import('@/views/admin/general/SystemConfig')
const SysSwitch = () => import('@/views/admin/general/SysSwitch')
const ProblemList = () => import('@/views/admin/problem/ProblemList')
const AdminGroupProblemList = () => import('@/views/admin/problem/GroupProblemList')
const Problem = () => import('@/views/admin/problem/Problem')
const ProblemAIHistory = () => import('@/views/admin/problem/ProblemAIHistory')
const Tag = () => import('@/views/admin/problem/Tag')
const ProblemImportAndExport = () => import('@/views/admin/problem/ImportAndExport')
const Contest = () => import('@/views/admin/contest/Contest')
const ContestList = () => import('@/views/admin/contest/ContestList')
const ContestRegistrationList = () => import('@/views/admin/contest/ContestRegistrationList')
const Training = () => import('@/views/admin/training/Training')
const TrainingList = () => import('@/views/admin/training/TrainingList')
const TrainingProblemList = () => import('@/views/admin/training/TrainingProblemList')
const TrainingParticipants = () => import('@/views/admin/training/TrainingParticipants')
const TrainingCategory = () => import('@/views/admin/training/Category')
const DiscussionList = () => import('@/views/admin/discussion/Discussion')
const UserRoleManagement = () => import('@/views/admin/classroom/UserRoleManagement')
const ClassroomAdmin = () => import('@/views/admin/classroom/ClassroomAdmin')
const ClassroomHomeworkAnalysis = () => import('@/views/admin/classroom/HomeworkAnalysis')
const ClassroomExamMonitoring = () => import('@/views/admin/classroom/ExamMonitoring')
const AdminCreateHomework = () => import('@/views/classroom/teacher/CreateHomework')
const AdminStudentSubmissionDetail = () => import('@/views/classroom/teacher/StudentSubmissionDetail')
const ToolboxAdmin = () => import('@/views/admin/toolbox/ToolboxAdmin')
const QuestionBankAdmin = () => import('@/views/admin/toolbox/QuestionBankAdmin')
const QuestionBankEditor = () => import('@/views/common/QuestionBankEditor')
const ExamPaperAdmin = () => import('@/views/admin/toolbox/ExamPaperAdmin')
const AdminLearningMapList = () => import('@/views/admin/learning-map/AdminLearningMapList')
const AdminLearningMapEditor = () => import('@/views/admin/learning-map/AdminLearningMapEditor')

const adminRoutes = [
  {
    path: '/admin/login',
    name: 'admin-login',
    component: Login,
    meta: { title: 'Login' }
  },
  {
    path: '/admin/',
    component: Home,
    meta: { requireAuth: true, requireAdmin: true },
    children: [
      {
        path: '',
        redirect: 'dashboard'
      },
      {
        path: 'dashboard',
        name: 'admin-dashboard',
        component: Dashboard,
        meta: { title: 'Dashboard' }
      },
      {
        path: 'user',
        name: 'admin-user',
        component: User,
        meta: { requireSuperAdmin: true, title: '用户管理' }
      },
      {
        path: 'rating',
        name: 'admin-rating',
        component: RatingManagement,
        meta: { requireSuperAdmin: true, title: 'Rating管理' }
      },
      {
        path: 'battle-records',
        name: 'admin-battle-records',
        component: BattleRecordsAdmin,
        meta: { requireSuperAdmin: true, title: '对战记录管理' }
      },
      {
        path: 'permission-docs',
        name: 'admin-permission-docs',
        component: PermissionDocs,
        meta: { requireAdmin: true, title: '权限说明' }
      },
      {
        path: 'question-bank',
        name: 'admin-question-bank',
        component: QuestionBankAdmin,
        meta: { requireSuperAdmin: true, title: '客观题题库管理' }
      },
      {
        path: 'question-bank/create',
        name: 'admin-question-bank-create',
        component: QuestionBankEditor,
        props: { scene: 'admin' },
        meta: { requireSuperAdmin: true, title: '创建客观题' }
      },
      {
        path: 'question-bank/edit/:questionId',
        name: 'admin-question-bank-edit',
        component: QuestionBankEditor,
        props: { scene: 'admin' },
        meta: { requireSuperAdmin: true, title: '编辑客观题' }
      },
      {
        path: 'exam-paper',
        name: 'admin-exam-paper',
        component: ExamPaperAdmin,
        meta: { requireAdmin: true, title: '试卷库管理' }
      },
      {
        path: 'announcement',
        name: 'admin-announcement',
        component: Announcement,
        meta: { requireSuperAdmin: true, title: '公告管理' }
      },
      {
        path: 'notice',
        name: 'admin-notice',
        component: SysNotice,
        meta: { requireSuperAdmin: true, title: '通知管理' }
      },
      {
        path: 'conf',
        name: 'admin-conf',
        component: SystemConfig,
        meta: { requireSuperAdmin: true, title: '系统配置' }
      },
      {
        path: 'switch',
        name: 'admin-switch',
        component: SysSwitch,
        meta: { requireSuperAdmin: true, title: '系统开关' }
      },
      {
        path: 'problems',
        name: 'admin-problem-list',
        component: ProblemList,
        meta: { title: '题目列表' }
      },
      {
        path: 'problem/create',
        name: 'admin-create-problem',
        component: Problem,
        meta: { title: '创建题目' }
      },
      {
        path: 'problem/edit/:problemId',
        name: 'admin-edit-problem',
        component: Problem,
        meta: { title: '编辑题目' }
      },
      {
        path: 'problem/edit/:problemId/ai-history',
        name: 'admin-problem-ai-history',
        component: ProblemAIHistory,
        meta: { title: '历史 AI 验题' }
      },
      {
        path: 'problem/tag',
        name: 'admin-problem-tag',
        component: Tag,
        meta: { title: '标签管理' }
      },
      {
        path: 'group-problem/apply',
        name: 'admin-group-apply-problem',
        component: AdminGroupProblemList,
        meta: { title: '团队题目审批' }
      },
      {
        path: 'problem/batch-operation',
        name: 'admin-problem_batch_operation',
        component: ProblemImportAndExport,
        meta: { title: '导入导出题目' }
      },
      {
        path: 'training/create',
        name: 'admin-create-training',
        component: Training,
        meta: { title: '创建训练' }
      },
      {
        path: 'training',
        name: 'admin-training-list',
        component: TrainingList,
        meta: { title: '训练列表' }
      },
      {
        path: 'training/:trainingId/edit',
        name: 'admin-edit-training',
        component: Training,
        meta: { title: '编辑训练' }
      },
      {
        path: 'training/:trainingId/problems',
        name: 'admin-training-problem-list',
        component: TrainingProblemList,
        meta: { title: '训练题目列表' }
      },
      {
        path: 'training/:trainingId/participants',
        name: 'admin-training-participants',
        component: TrainingParticipants,
        meta: { title: '训练参与者列表' }
      },
      {
        path: 'training/category',
        name: 'admin-training-category',
        component: TrainingCategory,
        meta: { title: '分类管理' }
      },
      {
        path: 'contest/create',
        name: 'admin-create-contest',
        component: Contest,
        meta: { title: '创建比赛' }
      },
      {
        path: 'contest',
        name: 'admin-contest-list',
        component: ContestList,
        meta: { title: '比赛列表' }
      },
      {
        path: 'contest/:contestId/edit',
        name: 'admin-edit-contest',
        component: Contest,
        meta: { title: '编辑比赛' }
      },
      {
        path: 'contest/:contestId/announcement',
        name: 'admin-contest-announcement',
        component: Announcement,
        meta: { title: '比赛公告' }
      },
      {
        path: 'contest/:contestId/problems',
        name: 'admin-contest-problem-list',
        component: ProblemList,
        meta: { title: '比赛题目列表' }
      },
      {
        path: 'contest/:contestId/problem/create',
        name: 'admin-create-contest-problem',
        component: Problem,
        meta: { title: '创建题目' }
      },
      {
        path: 'contest/:contestId/problem/:problemId/edit',
        name: 'admin-edit-contest-problem',
        component: Problem,
        meta: { title: '编辑题目' }
      },
      {
        path: 'discussion',
        name: 'admin-discussion-list',
        component: DiscussionList,
        meta: { title: '讨论管理' }
      },
      {
        path: 'user-role-management',
        name: 'admin-user-role-management',
        component: UserRoleManagement,
        meta: { requireAdmin: true, title: '用户角色管理' }
      },
      {
        path: 'classroom',
        name: 'admin-classroom',
        component: ClassroomAdmin,
        meta: { requireSuperAdmin: true, title: '班级管理' }
      },
      {
        path: 'classroom/homework/create/:classroomId',
        name: 'admin-create-homework',
        component: AdminCreateHomework,
        meta: { requireSuperAdmin: true, title: '创建作业' }
      },
      {
        path: 'classroom/homework/edit/:classroomId/:homeworkId',
        name: 'admin-edit-homework',
        component: AdminCreateHomework,
        meta: { requireSuperAdmin: true, title: '编辑作业' }
      },
      {
        path: 'classroom/homework/:homeworkId/submission/:uid',
        name: 'admin-student-submission-detail',
        component: AdminStudentSubmissionDetail,
        meta: { requireSuperAdmin: true, title: '学生提交详情' }
      },
      {
        path: 'classroom/homework/analysis',
        name: 'admin-classroom-homework-analysis',
        component: ClassroomHomeworkAnalysis,
        meta: { requireSuperAdmin: true, title: '作业学情分析' }
      },
      {
        path: 'classroom/homework/exam-monitoring',
        name: 'admin-classroom-exam-monitoring',
        component: ClassroomExamMonitoring,
        meta: { requireSuperAdmin: true, title: '作业考试监控' }
      },
      {
        path: 'toolbox',
        name: 'admin-toolbox',
        component: ToolboxAdmin,
        meta: { title: '工具箱' }
      },
      {
        path: 'learning-map',
        name: 'admin-learning-map-list',
        component: AdminLearningMapList,
        meta: { title: '航海图管理' }
      },
      {
        path: 'learning-map/:mapId',
        name: 'admin-learning-map-editor',
        component: AdminLearningMapEditor,
        meta: { title: '航海图编辑器' }
      },
      {
        path: 'contest/:contestId/registrations',
        name: 'admin-contest-registration-list',
        component: ContestRegistrationList,
        meta: { title: '比赛报名信息' }
      },
    ]
  },
  {
    path: '/admin/*',
    redirect: '/admin/login'
  }
]

export default adminRoutes
