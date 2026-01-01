import os
import sys
import time
import json
import warnings
import pymysql
import requests
import re
import subprocess
import tempfile
from flask import Flask, render_template_string, request, Response, stream_with_context, jsonify
from datetime import datetime

# 屏蔽无关警告
os.environ["PYTHONWARNINGS"] = "ignore"
warnings.filterwarnings("ignore")

app = Flask(__name__)

# ==========================================
# 1. 数据库配置
# ==========================================
DB_CONFIG = {
    "host": "43.143.133.62",
    "port": 3306,
    "user": "root",
    "password": "hist2025",
    "db": "hoj",
    "charset": "utf8mb4",
    "cursorclass": pymysql.cursors.DictCursor,
    "connect_timeout": 5
}

# ==========================================
# 2. API 配置
# ==========================================
BASE_URL = "http://bingoj.cn"
API_LOGIN = f"{BASE_URL}/api/login"
API_SUBMIT = f"{BASE_URL}/api/submit-problem-judge"
API_RESULT = f"{BASE_URL}/api/get-submission-detail"
API_PROBLEM_NORMAL = f"{BASE_URL}/api/get-problem-detail"
API_PROBLEM_CONTEST = f"{BASE_URL}/api/get-contest-problem-details"

# 状态映射表（中文化）
STATUS_MAP = {
    0: "答案正确", -1: "答案错误", -2: "编译错误",
    -3: "格式错误", 1: "时间超限",
    2: "内存超限", 3: "运行错误",
    4: "NO", 5: "系统错误", 6: "等待中", 7: "判题中",
    8: "部分正确", 9: "提交中", 10: "提交失败"
}

# ==========================================
# 3. 前端页面代码
# ==========================================
FRONTEND_HTML = """
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BingoJ 智能判题终端</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <script>MathJax={tex:{inlineMath:[['$','$'],['\\\\(','\\\\)']]},svg:{fontCache:'global'}};</script>
    <script id="MathJax-script" async src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"></script>

    <style>
        body { background-color: #f4f6f9; font-family: 'Segoe UI', Roboto, "Microsoft YaHei", sans-serif; padding-top: 15px; }
        .main-container { max-width: 1600px; margin: 0 auto; padding-bottom: 50px; }

        /* 卡片通用 */
        .card { border: none; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border-radius: 8px; margin-bottom: 15px; background: #fff; }
        .card-header { background: #fff; border-bottom: 1px solid #f0f0f0; font-weight: 600; padding: 10px 15px; color: #444; }
        .card-header i { color: #0d6efd; margin-right: 6px; }

        /* 状态文字样式 */
        .status-text { font-weight: bold; font-size: 11px; }

        /* 结果面板 */
        #result-panel { display: none; margin-top: 20px; }
        .result-banner { padding: 20px; text-align: center; border-radius: 8px; margin-bottom: 15px; color: #fff; font-size: 1.6rem; font-weight: bold; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .bg-ac { background: linear-gradient(135deg, #28a745, #20c997); }
        .bg-wa { background: linear-gradient(135deg, #dc3545, #ff6b6b); }
        .bg-pending { background: linear-gradient(135deg, #6c757d, #adb5bd); }

        /* 样例列表 */
        .sample-item { border: 1px solid #eee; border-radius: 6px; margin-bottom: 8px; background: #fff; overflow: hidden; }
        .sample-header { padding: 8px 15px; display: flex; justify-content: space-between; align-items: center; cursor: pointer; background: #fdfdfd; }
        .sample-header:hover { background: #f8f9fa; }
        .sample-detail { display: none; padding: 10px 15px; background: #fafafa; border-top: 1px solid #eee; font-size: 0.9rem; }
        .code-block { background: #2d2d2d; color: #f8f8f2; padding: 8px; border-radius: 4px; font-family: Consolas, monospace; white-space: pre-wrap; margin: 4px 0; }

        /* 左侧日志 */
        #log-area { background-color: #212529; color: #00ff00; font-family: 'Consolas', monospace; padding: 10px; border-radius: 4px; height: 120px; overflow-y: auto; font-size: 11px; margin-bottom: 0; }

        /* 历史表格优化 */
        .history-table { table-layout: fixed; }
        .history-table th { font-size: 11px; color: #666; font-weight: 600; padding: 8px 4px; }
        .history-table td { font-size: 11px; padding: 6px 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; vertical-align: middle; }
        .badge-custom { font-size: 0.85rem; font-weight: normal; }

        /* 题目内容 */
        .prob-content { font-size: 0.95rem; line-height: 1.6; color: #333; }
        .prob-content pre { background: #f8f9fa; padding: 10px; border: 1px solid #e9ecef; border-radius: 4px; font-family: Consolas; }
    </style>
</head>
<body>

<div class="container-fluid main-container">
    <div class="row">
        <!-- 左侧栏 -->
        <div class="col-lg-4">
            <h5 class="mb-3 fw-bold text-primary"><i class="fas fa-terminal"></i> BingoJ 验题终端</h5>

            <form id="mainForm">
                <!-- 1. 设置 -->
                <div class="card">
                    <div class="card-header py-2 small"><i class="fas fa-sliders-h"></i> 配置</div>
                    <div class="card-body py-2">
                        <div class="row g-2 mb-2">
                            <div class="col-6"><input type="text" class="form-control form-control-sm" id="username" value="root"></div>
                            <div class="col-6"><input type="password" class="form-control form-control-sm" id="password" value="hist2025"></div>
                        </div>
                        <div class="d-flex align-items-center gap-2">
                             <div class="form-check form-check-inline m-0">
                                <input class="form-check-input" type="radio" name="mode" value="normal" checked onchange="toggleMode()">
                                <label class="form-check-label small">普通</label>
                            </div>
                            <div class="form-check form-check-inline m-0">
                                <input class="form-check-input" type="radio" name="mode" value="contest" onchange="toggleMode()">
                                <label class="form-check-label small">比赛</label>
                            </div>
                            <button type="button" class="btn btn-sm btn-outline-primary ms-auto" onclick="fetchInfo()">获取题目</button>
                        </div>
                    </div>
                </div>

                <!-- 2. 代码 -->
                <div class="card">
                    <div class="card-header py-2 small"><i class="fas fa-code"></i> 代码编辑器</div>
                    <div class="card-body p-2">
                        <div class="input-group input-group-sm mb-2">
                            <input type="text" class="form-control" id="pid" placeholder="ID (如 1001)">
                            <input type="text" class="form-control" id="cid" placeholder="比赛ID" style="display:none;">
                            <select class="form-select" id="language" style="max-width: 120px;">
                                <option value="C++ 17 With O2">C++ 17</option>
                                <option value="C With O2">C</option>
                                <option value="Python3">Python3</option>
                                <option value="Java">Java</option>
                            </select>
                        </div>
                        <textarea class="form-control" id="code" rows="12" placeholder="// 在此粘贴代码..."
                                  style="font-family: Consolas, monospace; font-size: 13px; background:#f8f9fa; border:1px solid #dee2e6;"></textarea>

                        <button type="submit" class="btn btn-primary w-100 mt-2 fw-bold" id="runBtn">
                            <i class="fas fa-play"></i> 运行自测并提交
                        </button>
                    </div>
                </div>
            </form>

            <!-- 3. 日志 -->
            <div class="card bg-dark">
                <div class="card-header bg-secondary text-white py-1 small border-0">系统日志</div>
                <div class="card-body p-0">
                    <div id="log-area"></div>
                </div>
            </div>

            <!-- 4. 数据库历史 -->
            <div class="card">
                <div class="card-header py-2 small"><i class="fas fa-history"></i> 本地记录</div>
                <div class="table-responsive">
                    <table class="table table-hover table-striped history-table mb-0 text-center align-middle">
                        <thead class="table-light">
                            <tr>
                                <th style="width:12%">用户</th>
                                <th style="width:19%">远程结果</th>
                                <th style="width:19%">本地自测</th>
                                <th style="width:18%">耗时/内存</th>
                                <th style="width:11%">语言</th>
                                <th style="width:13%">时间</th>
                                <th style="width:8%">代码</th>
                            </tr>
                        </thead>
                        <tbody id="dbHistoryBody">
                            <tr><td colspan="7" class="text-muted">暂无数据</td></tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <!-- 右侧栏 -->
        <div class="col-lg-8">

            <!-- A. 题目详情 (置顶) -->
            <div class="card">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <span><i class="fas fa-book-open"></i> 题目详情</span>
                    <span id="probHeaderInfo" class="badge bg-secondary">未加载</span>
                </div>
                <div class="card-body" id="problemContent" style="min-height: 200px;">
                    <div class="text-center text-muted py-5">
                        <i class="fas fa-arrow-left"></i> 请先在左侧输入ID并获取题目
                    </div>
                </div>
            </div>

            <!-- B. 结果面板 (置底) -->
            <div id="result-panel">
                <div class="card border-primary">
                    <div class="card-header bg-primary text-white"><i class="fas fa-poll-h"></i> 判题结果报告</div>
                    <div class="card-body">
                        <!-- 远程大Banner -->
                        <div id="remote-banner" class="result-banner bg-pending">
                            <i class="fas fa-hourglass-start"></i> 等待开始...
                        </div>

                        <!-- 样例列表 -->
                        <div class="d-flex justify-content-between align-items-center mb-2 mt-4 border-bottom pb-2">
                            <h6 class="m-0 fw-bold text-secondary">本地样例自测详情</h6>
                            <span id="sample-summary" class="badge bg-light text-dark border">等待中</span>
                        </div>
                        <div id="sample-list"></div>
                    </div>
                </div>
            </div>

        </div>
    </div>
</div>

<!-- 代码模态框 -->
<div class="modal fade" id="codeModal" tabindex="-1">
    <div class="modal-dialog modal-lg">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title">提交代码预览</h5>
                <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
                <pre id="codeModalBody" class="code-block" style="max-height:600px;overflow-y:auto;"></pre>
            </div>
        </div>
    </div>
</div>

<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
<script>
    let currentHistoryData = [];

    function toggleMode() {
        const mode = document.querySelector('input[name="mode"]:checked').value;
        const cidInput = document.getElementById('cid');
        const pidInput = document.getElementById('pid');
        if (mode === 'contest') {
            cidInput.style.display = 'block';
            pidInput.placeholder = '题号 (如 A,B)';
        } else {
            cidInput.style.display = 'none';
            pidInput.placeholder = '题号 (如 1001)';
        }
    }

    function appendLog(msg) {
        const box = document.getElementById('log-area');
        const time = new Date().toLocaleTimeString('en-GB', { hour12: false });
        box.innerHTML += `<div class="log-line"><span>[${time}]</span> ${msg}</div>`;
        box.scrollTop = box.scrollHeight;
    }

    window.toggleSample = function(id) {
        const el = document.getElementById(id);
        el.style.display = (el.style.display === 'none') ? 'block' : 'none';
    }

    window.showCode = function(index) {
        if (currentHistoryData && currentHistoryData[index]) {
            document.getElementById('codeModalBody').innerText = currentHistoryData[index].code;
            new bootstrap.Modal(document.getElementById('codeModal')).show();
        }
    }

    // 获取信息
    async function fetchInfo() {
        const pid = document.getElementById('pid').value;
        const cid = document.getElementById('cid').value;
        const mode = document.querySelector('input[name="mode"]:checked').value;
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        if (!pid) return alert("请输入题目ID");
        appendLog(`正在获取 ${pid}...`);

        try {
            const res = await fetch('/api/get-info', {
                method: 'POST', headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({ pid, cid, mode, username, password })
            });
            const data = await res.json();

            if (data.status === 'success') {
                const p = data.problem;
                document.getElementById('probHeaderInfo').innerText = data.display_id;
                document.getElementById('probHeaderInfo').className = "badge bg-primary";

                let html = `<h4 class="text-center mb-4">${data.display_id} - ${p.title}</h4>`;
                const secs = [{t:'描述',c:p.description},{t:'输入',c:p.input},{t:'输出',c:p.output},{t:'提示',c:p.hint}];
                secs.forEach(s => {
                    if(s.c) html += `<div class="mb-3"><h6 class="fw-bold text-primary"><i class="fas fa-caret-right"></i> ${s.t}</h6><div class="prob-content ms-2">${marked.parse(s.c)}</div></div>`;
                });
                if (p.examples) {
                    const regex = /<input>([\s\S]*?)<\/input><output>([\s\S]*?)<\/output>/g;
                    let m, i=1;
                    while ((m = regex.exec(p.examples)) !== null) {
                        html += `<div class="row mb-2 ms-1"><div class="col-md-6"><small class="fw-bold">样例 ${i} 输入:</small><pre class="p-2 border bg-light">${m[1].trim()}</pre></div><div class="col-md-6"><small class="fw-bold">样例 ${i} 输出:</small><pre class="p-2 border bg-light">${m[2].trim()}</pre></div></div>`;
                        i++;
                    }
                }
                document.getElementById('problemContent').innerHTML = html;
                if(window.MathJax) MathJax.typesetPromise();

                // 渲染历史
                currentHistoryData = data.history;
                renderHistoryTable();
                appendLog("获取成功.");
            } else { appendLog("错误: " + data.message); }
        } catch(e) { appendLog("网络错误"); }
    }

    function renderHistoryTable() {
        const tbody = document.getElementById('dbHistoryBody');
        tbody.innerHTML = '';
        if (!currentHistoryData.length) {
            tbody.innerHTML = '<tr><td colspan="7" class="text-muted">无历史记录</td></tr>';
            return;
        }
        currentHistoryData.forEach((h, index) => {
            // 1. OJ Result (适配中文 Badge)
            let ojHtml = h.result;
            if (h.result === '答案正确') {
                ojHtml = '<span class="badge bg-success badge-custom"><i class="fas fa-check"></i> 答案正确</span>';
            } else if (['答案错误', '时间超限', '内存超限', '运行错误', '编译错误', '格式错误'].some(s => h.result.includes(s))) {
                ojHtml = `<span class="badge bg-danger badge-custom">${h.result}</span>`;
            } else if (['等待中', '判题中', '提交中'].includes(h.result)) {
                ojHtml = `<span class="badge bg-warning text-dark badge-custom">${h.result}</span>`;
            } else {
                ojHtml = `<span class="badge bg-secondary badge-custom">${h.result}</span>`;
            }

            // 2. Local Result (适配中文 Badge)
            let localText = h.local_info || '-';
            let localHtml = localText;

            if (localText.includes('All Pass') || localText.includes('全部通过')) {
                localHtml = '<span class="badge bg-success badge-custom"><i class="fas fa-check"></i> 全部通过</span>';
            } else if (localText.includes('Pass') && !localText.includes('All')) {
                localHtml = `<span class="badge bg-danger badge-custom">${localText.replace('Pass', '通过')}</span>`;
            } else if (localText.includes('Skip') || localText.includes('跳过')) {
                localHtml = '<span class="badge bg-secondary badge-custom">跳过</span>';
            } else if (localText !== '-' && localText !== '') {
                 localHtml = `<span class="badge bg-danger badge-custom">${localText}</span>`;
            }

            // 3. 简化语言
            let simpleLang = h.language;
            if (simpleLang.includes('C++')) simpleLang = 'C++';
            else if (simpleLang.includes('C With')) simpleLang = 'C';
            else if (simpleLang.includes('Python')) simpleLang = 'Py3';

            // 4. 时间格式
            const shortTime = h.submit_time ? h.submit_time.split('T')[1].split('.')[0] : '';

            // 5. 耗时/内存 (新增)
            const timeMem = `${h.time_used || '--'} / ${h.memory_used || '--'}`;

            tbody.innerHTML += `
                <tr>
                    <td title="${h.username}">${h.username}</td>
                    <td>${ojHtml}</td>
                    <td>${localHtml}</td>
                    <td class="text-secondary" style="font-size:11px;">${timeMem}</td>
                    <td>${simpleLang}</td>
                    <td style="color:#888;">${shortTime}</td>
                    <td><a href="javascript:showCode(${index})" class="text-dark"><i class="fas fa-code"></i></a></td>
                </tr>`;
        });
    }

    // 提交运行
    document.getElementById('mainForm').addEventListener('submit', async function(e) {
        e.preventDefault();
        const btn = document.getElementById('runBtn');
        const payload = {
            mode: document.querySelector('input[name="mode"]:checked').value,
            pid: document.getElementById('pid').value,
            cid: document.getElementById('cid').value,
            username: document.getElementById('username').value,
            password: document.getElementById('password').value,
            language: document.getElementById('language').value,
            code: document.getElementById('code').value
        };
        if(!payload.code) return alert("代码不能为空!");

        btn.disabled = true;
        btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> 处理中...';
        document.getElementById('result-panel').style.display = 'block';
        document.getElementById('log-area').innerHTML = '';

        setTimeout(() => document.getElementById('result-panel').scrollIntoView({behavior: "smooth"}), 100);

        const remoteBanner = document.getElementById('remote-banner');
        remoteBanner.className = 'result-banner bg-pending';
        remoteBanner.innerHTML = '<i class="fas fa-running"></i> 本地测试中...';
        document.getElementById('sample-list').innerHTML = '';
        let passCount = 0, totalCount = 0;

        try {
            const response = await fetch('/run-combined', {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            const reader = response.body.getReader();
            const decoder = new TextDecoder();

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                decoder.decode(value).split('\\n\\n').forEach(line => {
                    if (line.startsWith('data: ')) {
                        try {
                            const d = JSON.parse(line.substring(6));
                            if (d.type === 'log') appendLog(d.msg);
                            else if (d.type === 'sample_res') {
                                totalCount++;
                                if(d.is_ok) passCount++;
                                const icon = d.is_ok ? '<i class="fas fa-check text-success"></i>' : '<i class="fas fa-times text-danger"></i>';
                                const detailId = 'sd-' + Math.random().toString(36).substr(2);

                                document.getElementById('sample-list').innerHTML += `
                                <div class="sample-item">
                                    <div class="sample-header" onclick="toggleSample('${detailId}')">
                                        <span>样例 ${d.id}</span>
                                        <span>${icon} ${d.is_ok?'通过':'失败'}</span>
                                    </div>
                                    <div id="${detailId}" class="sample-detail">
                                        <div><strong>输入:</strong><div class="code-block">${d.input}</div></div>
                                        <div><strong>预期:</strong><div class="code-block">${d.expected}</div></div>
                                        <div><strong>输出:</strong><div class="code-block">${d.output}</div></div>
                                    </div>
                                </div>`;
                                document.getElementById('sample-summary').innerText = `通过: ${passCount}/${totalCount}`;
                            }
                            else if (d.type === 'compile_error') {
                                document.getElementById('sample-list').innerHTML = `<div class="alert alert-danger p-2"><pre class="m-0">${d.msg}</pre></div>`;
                                remoteBanner.className = 'result-banner bg-wa'; remoteBanner.innerHTML = '本地编译错误';
                            }
                            else if (d.type === 'remote_status') {
                                if (['等待中','判题中','提交中'].includes(d.status)) {
                                    remoteBanner.innerHTML = `<i class="fas fa-sync fa-spin"></i> ${d.status}...`;
                                } else {
                                    const isAc = d.status === '答案正确';
                                    remoteBanner.className = isAc ? 'result-banner bg-ac' : 'result-banner bg-wa';
                                    remoteBanner.innerHTML = isAc ? `<i class="fas fa-trophy"></i> 答案正确` : `<i class="fas fa-times-circle"></i> ${d.status}`;
                                    fetchInfo(); // 刷新表格
                                }
                            }
                        } catch(e) {}
                    }
                });
            }
        } catch (err) { appendLog("错误: " + err); }
        finally {
            btn.disabled = false;
            btn.innerHTML = '<i class="fas fa-play"></i> 运行自测并提交';
        }
    });
</script>
</body>
</html>
"""

# ==========================================
# 4. 后端逻辑
# ==========================================
def get_db_connection():
    try: return pymysql.connect(**DB_CONFIG)
    except: return None

def init_db():
    conn = get_db_connection()
    if not conn: return
    try:
        with conn.cursor() as cursor:
            cursor.execute("""
            CREATE TABLE IF NOT EXISTS submission_history (
                id INT AUTO_INCREMENT PRIMARY KEY,
                submit_id VARCHAR(50), pid VARCHAR(50), cid VARCHAR(50) DEFAULT '0',
                username VARCHAR(100), result VARCHAR(50), time_used VARCHAR(20),
                memory_used VARCHAR(20), language VARCHAR(50), code MEDIUMTEXT,
                local_info VARCHAR(100) DEFAULT '',
                submit_time DATETIME DEFAULT CURRENT_TIMESTAMP
            ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
            """)
            cursor.execute("SHOW COLUMNS FROM submission_history LIKE 'code'")
            if not cursor.fetchone(): cursor.execute("ALTER TABLE submission_history ADD COLUMN code MEDIUMTEXT")
            cursor.execute("SHOW COLUMNS FROM submission_history LIKE 'local_info'")
            if not cursor.fetchone(): cursor.execute("ALTER TABLE submission_history ADD COLUMN local_info VARCHAR(100) DEFAULT ''")
        conn.commit()
    finally: conn.close()

def save_to_db(data):
    conn = get_db_connection()
    if not conn: return
    try:
        with conn.cursor() as cursor:
            cursor.execute("""
                INSERT INTO submission_history
                (submit_id, pid, cid, username, result, time_used, memory_used, language, code, local_info)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, (data['submit_id'], data['pid'], data['cid'], data['username'],
                data['result'], data['time'], data['memory'], data['language'], data.get('code',''), data.get('local_info', '')))
        conn.commit()
    finally: conn.close()

# 【修复】查询函数现在接收 cid，并过滤 cid
def query_history(pid, cid):
    conn = get_db_connection()
    if not conn: return []
    try:
        with conn.cursor() as cursor:
            # 增加 AND cid=%s 条件，防止不同比赛的同号题目（如A题）混淆
            cursor.execute("SELECT * FROM submission_history WHERE pid=%s AND cid=%s ORDER BY id DESC LIMIT 10", (str(pid), str(cid)))
            return cursor.fetchall()
    finally: conn.close()

def get_oj_session(username, password):
    s = requests.Session()
    try:
        res = s.post(API_LOGIN, json={"username": username, "password": password}, timeout=5)
        if res.status_code != 200: return None, "登录接口异常"
        token = res.headers.get("Authorization")
        if not token: return None, "Token获取失败"
        s.headers.update({"Authorization": token, "User-Agent": "Mozilla/5.0"})
        return s, "OK"
    except Exception as e: return None, str(e)

def compile_local(language, code, temp_dir):
    if "C++" in language or "C With" in language:
        src = os.path.join(temp_dir, "main.cpp")
        exe = os.path.join(temp_dir, "main")
        if os.name == 'nt': exe += ".exe"
        with open(src, 'w', encoding='utf-8') as f: f.write(code)
        compiler = "gcc" if language.startswith("C ") else "g++"
        cmd = [compiler, src, "-o", exe, "-O2"]
        if compiler == "g++": cmd.append("-std=c++17")
        try:
            subprocess.check_output(cmd, stderr=subprocess.STDOUT)
            return exe, None
        except subprocess.CalledProcessError as e:
            return None, e.output.decode('utf-8', errors='ignore')
    elif language == "Java":
        code = re.sub(r'public\s+class\s+\w+', 'public class Main', code)
        src = os.path.join(temp_dir, "Main.java")
        with open(src, 'w', encoding='utf-8') as f: f.write(code)
        try:
            subprocess.check_output(["javac", src], stderr=subprocess.STDOUT)
            return "java", None
        except subprocess.CalledProcessError as e:
            return None, e.output.decode('utf-8', errors='ignore')
    elif "Python" in language:
        src = os.path.join(temp_dir, "main.py")
        with open(src, 'w', encoding='utf-8') as f: f.write(code)
        return sys.executable, None
    return None, "不支持该语言本地测试"

def run_once(exec_cmd, inp_str, temp_dir, language):
    cmd = []
    if "C++" in language or "C With" in language: cmd = [exec_cmd]
    elif language == "Java": cmd = ["java", "-cp", temp_dir, "Main"]
    elif "Python" in language: cmd = [exec_cmd, os.path.join(temp_dir, "main.py")]
    try:
        proc = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        stdout, stderr = proc.communicate(input=inp_str, timeout=2)
        return stdout, stderr, None
    except subprocess.TimeoutExpired:
        proc.kill()
        return "", "", "时间超限"
    except Exception as e:
        return "", "", str(e)

@app.route('/')
def index(): return render_template_string(FRONTEND_HTML)

@app.route('/api/get-info', methods=['POST'])
def get_info():
    d = request.json
    username, password = d.get('username'), d.get('password')
    pid, cid, mode = d.get('pid'), d.get('cid', '0'), d.get('mode')

    # 【修复】确定查询数据库时使用的 cid
    if mode == 'contest':
        target_url, params = API_PROBLEM_CONTEST, {"displayId": str(pid).upper(), "cid": cid}
        db_pid = str(pid).upper()
        db_cid = str(cid) # 比赛模式使用具体 CID
    else:
        target_url, params = API_PROBLEM_NORMAL, {"problemId": pid}
        db_pid = str(pid)
        db_cid = '0' # 普通模式 CID 默认为 '0'

    session, msg = get_oj_session(username, password)
    if not session: return jsonify({"status": "error", "message": msg})

    res_data = {"status": "success", "problem": None, "history": [], "display_id": db_pid}
    try:
        r = session.get(target_url, params=params, timeout=8)
        dt = r.json().get('data', {})
        prob = dt.get('problem', dt) if 'problem' in dt else dt
        if prob:
            res_data['problem'] = {
                "title": prob.get('title'), "description": prob.get('description'),
                "input": prob.get('input'), "output": prob.get('output'),
                "examples": prob.get('examples'), "hint": prob.get('hint')
            }

        # 【修复】调用历史记录查询时传入 cid
        res_data['history'] = query_history(db_pid, db_cid)

        for h in res_data['history']:
             if isinstance(h['submit_time'], datetime): h['submit_time'] = h['submit_time'].isoformat()
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)})

    return jsonify(res_data)

@app.route('/run-combined', methods=['POST'])
def run_combined():
    d = request.json
    username, password = d.get('username'), d.get('password')
    pid, cid, mode = d.get('pid'), d.get('cid', '0'), d.get('mode')
    language, code = d.get('language'), d.get('code')

    def stream():
        def send(type_, **kwargs): return f"data: {json.dumps({'type': type_, **kwargs})}\n\n"

        yield send("log", msg="正在登录 OJ...")
        session, msg = get_oj_session(username, password)
        if not session:
            yield send("log", msg=f"登录失败: {msg}")
            return

        if mode == 'contest':
            url, params = API_PROBLEM_CONTEST, {"displayId": str(pid).upper(), "cid": cid}
            real_pid = str(pid).upper()
        else:
            url, params = API_PROBLEM_NORMAL, {"problemId": pid}
            real_pid = str(pid)

        # 1. 提取样例
        yield send("log", msg="正在获取样例...")
        examples = ""
        try:
            r = session.get(url, params=params, timeout=5)
            dt = r.json().get('data', {})
            prob = dt.get('problem', dt) if 'problem' in dt else dt
            examples = prob.get('examples', '')
        except:
            examples = ""

        local_info = "跳过 (无样例)"

        # 2. 本地自测
        if examples:
            matches = re.findall(r'<input>(.*?)</input><output>(.*?)</output>', examples, re.DOTALL)
            if matches:
                yield send("log", msg=f"开始自测 {len(matches)} 组样例...")
                with tempfile.TemporaryDirectory() as temp_dir:
                    exe, err = compile_local(language, code, temp_dir)
                    if err:
                        yield send("compile_error", msg=err)
                        yield send("log", msg="本地编译失败")
                        return

                    pass_cnt = 0
                    for idx, (inp, exp) in enumerate(matches, 1):
                        inp, exp = inp.strip(), exp.strip()
                        out, _, run_err = run_once(exe, inp, temp_dir, language)
                        out = out.strip()
                        is_ok = (out == exp) and (not run_err)
                        if is_ok: pass_cnt += 1
                        yield send("sample_res", id=idx, is_ok=is_ok, input=inp, expected=exp, output=run_err if run_err else out)

                    local_info = "全部通过" if pass_cnt == len(matches) else f"{pass_cnt}/{len(matches)} 通过"
                    yield send("log", msg=f"自测完成: {local_info}")
            else:
                yield send("log", msg="无有效样例")

        # 3. 提交远程
        yield send("log", msg="正在提交远程 OJ...")
        yield send("remote_status", status="提交中")

        # 提交参数修正逻辑
        submit_pid = str(pid).upper() if mode == 'contest' else str(pid)
        # 确保 submit_cid 在普通模式下为 0，比赛模式下为具体ID
        submit_cid = int(cid) if (mode == 'contest' and cid) else 0

        payload = {
            "pid": submit_pid,
            "cid": submit_cid,
            "language": language,
            "code": code,
            "tid": None, "gid": None, "isRemote": False
        }

        try:
            r_sub = session.post(API_SUBMIT, json=payload, timeout=10)
            res_json = r_sub.json()
            submit_id = res_json.get("data", {}).get("submitId")
            if not submit_id:
                yield send("log", msg=f"提交失败: {res_json.get('msg')}")
                yield send("remote_status", status="提交失败")
                return

            yield send("log", msg=f"提交ID: {submit_id}，判题中...")
            for _ in range(20):
                time.sleep(1)
                r_res = session.get(f"{API_RESULT}?submitId={submit_id}")
                sub_info = r_res.json().get("data", {}).get("submission", {})
                status_code = sub_info.get("status", 6)
                status_text = STATUS_MAP.get(status_code, "未知状态")

                if status_code in [6, 7, 9]:
                    yield send("remote_status", status=status_text)
                    continue

                yield send("remote_status", status=status_text)
                yield send("log", msg=f"最终结果: {status_text}")

                # 保存到数据库
                save_data = {
                    "submit_id": submit_id, "pid": real_pid, "cid": str(submit_cid), # 确保cid存为字符串
                    "username": username, "result": status_text,
                    "time": f"{sub_info.get('time')}ms", "memory": f"{sub_info.get('memory')}KB",
                    "language": language, "code": code,
                    "local_info": local_info
                }
                save_to_db(save_data)
                return
            yield send("log", msg="查询超时")
            yield send("remote_status", status="超时")
        except Exception as e: yield send("log", msg=f"异常: {e}")

    return Response(stream_with_context(stream()), mimetype='text/event-stream')

if __name__ == '__main__':
    init_db()
    print("Starting on http://127.0.0.1:5000")
    app.run(debug=True, port=5000)