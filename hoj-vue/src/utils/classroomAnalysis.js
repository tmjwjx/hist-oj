export function firstDefined(source, keys, fallback) {
  for (const key of keys) {
    if (source && source[key] !== undefined && source[key] !== null) {
      return source[key]
    }
  }
  return fallback
}

export function toNumber(value, fallback = 0) {
  const number = Number(value)
  return Number.isFinite(number) ? number : fallback
}

function toArray(value) {
  if (Array.isArray(value)) return value
  if (typeof value !== 'string' || value.trim() === '') return []
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch (error) {
    return []
  }
}

function toAnswerArray(value) {
  if (Array.isArray(value)) return value
  if (value === undefined || value === null || value === '') return []
  const parsed = toArray(value)
  return parsed.length > 0 ? parsed : [value]
}

function normalizeStudent(student) {
  const source = student || {}
  return {
    ...source,
    realName: firstDefined(source, ['realName', 'real_name', 'realname'], ''),
    username: firstDefined(source, ['username', 'userName', 'user_name'], ''),
    score: toNumber(firstDefined(source, ['score', 'totalScore', 'total_score'], 0)),
    rank: firstDefined(source, ['rank', 'ranking'], null),
    answer: toAnswerArray(firstDefined(source, ['answer', 'answers'], []))
  }
}

function normalizeOption(option) {
  const source = typeof option === 'object' && option !== null ? option : { content: option }
  const selectedBy = toArray(firstDefined(source, ['selectedBy', 'selected_by', 'students'], []))
    .map(normalizeStudent)
  return {
    ...source,
    content: firstDefined(source, ['content', 'optionContent', 'option_content', 'label'], ''),
    label: firstDefined(source, ['label', 'content'], ''),
    selectedCount: toNumber(firstDefined(source, ['selectedCount', 'selected_count', 'count'], selectedBy.length)),
    percentage: toNumber(firstDefined(source, ['percentage', 'percent', 'rate'], 0)),
    selectedBy
  }
}

function normalizeQuestion(question, index) {
  const source = question || {}
  const submittedBy = toArray(firstDefined(source, ['submittedBy', 'submitted_by', 'submittedStudents', 'submitted_students'], []))
    .map(normalizeStudent)
  const unsubmittedBy = toArray(firstDefined(source, ['unsubmittedBy', 'unsubmitted_by', 'unsubmittedStudents', 'unsubmitted_students'], []))
    .map(normalizeStudent)
  const options = toArray(firstDefined(source, ['options', 'optionAnalysis', 'option_analysis', 'distribution'], []))
    .map(normalizeOption)
  const optionTotal = options.reduce((sum, option) => sum + option.selectedCount, 0)
  options.forEach(option => {
    if (!option.percentage && optionTotal > 0) {
      option.percentage = option.selectedCount / optionTotal * 100
    }
  })

  return {
    ...source,
    id: firstDefined(source, ['id', 'homeworkQuestionId', 'homework_question_id'], index),
    homeworkQuestionId: firstDefined(source, ['homeworkQuestionId', 'homework_question_id', 'id'], index),
    questionOrder: firstDefined(source, ['questionOrder', 'question_order', 'order'], index + 1),
    title: firstDefined(source, ['title', 'questionTitle', 'question_title'], ''),
    type: firstDefined(source, ['type', 'questionType', 'question_type'], ''),
    score: toNumber(firstDefined(source, ['score', 'fullScore', 'full_score'], 0)),
    avgScore: toNumber(firstDefined(source, ['avgScore', 'averageScore', 'avg_score', 'average_score'], 0)),
    submittedCount: toNumber(firstDefined(source, ['submittedCount', 'submissionCount', 'submitted_count', 'submission_count'], submittedBy.length)),
    unsubmittedCount: toNumber(firstDefined(source, ['unsubmittedCount', 'unsubmitted_count'], unsubmittedBy.length)),
    submittedBy,
    unsubmittedBy,
    options,
    answer: firstDefined(source, ['answer', 'correctAnswer', 'correct_answer'], '')
  }
}

export function normalizeHomeworkAnalysis(payload) {
  const source = payload && typeof payload === 'object' ? payload : {}
  const submittedStudents = toArray(firstDefined(source, ['submittedStudents', 'submitted_students'], []))
    .map(normalizeStudent)
  const unsubmittedStudents = toArray(firstDefined(source, ['unsubmittedStudents', 'unsubmitted_students'], []))
    .map(normalizeStudent)
  const studentRankings = toArray(firstDefined(source, ['studentRankings', 'student_rankings', 'rankings'], []))
    .map(normalizeStudent)

  // 当前最小版接口中 studentCount 是交卷学生数，submissionCount 是题目作答记录数。
  const responseStudentCount = toNumber(firstDefined(source, ['studentCount', 'student_count'], 0))
  const submittedCount = toNumber(firstDefined(
    source,
    ['submittedCount', 'submitted_count', 'submittedStudentCount', 'submitted_student_count'],
    submittedStudents.length || responseStudentCount
  ))
  const explicitTotal = firstDefined(
    source,
    ['totalStudentCount', 'total_student_count', 'totalStudents', 'total_students', 'classStudentCount', 'class_student_count'],
    null
  )
  const totalStudentCount = explicitTotal === null
    ? Math.max(submittedCount + unsubmittedStudents.length, responseStudentCount)
    : toNumber(explicitTotal)
  const unsubmittedCount = toNumber(firstDefined(
    source,
    ['unsubmittedCount', 'unsubmitted_count', 'unsubmittedStudentCount', 'unsubmitted_student_count'],
    Math.max(totalStudentCount - submittedCount, 0)
  ))
  const questionAnalysis = toArray(firstDefined(
    source,
    ['questionAnalysis', 'question_analysis', 'questions'],
    []
  )).map(normalizeQuestion)

  return {
    ...source,
    totalStudentCount,
    submittedCount,
    unsubmittedCount,
    submissionCount: toNumber(firstDefined(source, ['submissionCount', 'submission_count'], 0)),
    averageScore: toNumber(firstDefined(source, ['averageScore', 'average_score', 'avgScore', 'avg_score'], 0)),
    submittedStudents,
    unsubmittedStudents,
    studentRankings,
    questionAnalysis
  }
}
