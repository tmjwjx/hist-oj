#!/usr/bin/env node
/**
 * Node >= 18，无依赖。与 loadtest.html 共用同一请求模拟引擎，不执行 Vue。
 * node scripts/loadtest.mjs --users 25 --ramp 10 --host https://bingoj.cn
 * node scripts/loadtest.mjs --self-test            # 仅本机 HTTP fixture
 * 参数：--users N --ramp 秒 --max-queue-wait 秒 --load-timeout 秒
 *       --timeout 资源请求超时秒 --poll 毫秒 --sla 秒 --prefetch --use-cache --out 文件
 * SIGINT/SIGTERM 中止请求和未到达会话；等待独立清理请求结束后退出。
 */
import fs from 'node:fs';
import vm from 'node:vm';
import { webcrypto } from 'node:crypto';

if (!globalThis.crypto) globalThis.crypto = webcrypto;
const html = fs.readFileSync(new URL('./loadtest.html', import.meta.url), 'utf8');
const engine = html.match(/<script id="loadtest-engine">([\s\S]*?)<\/script>/);
if (!engine) throw new Error('loadtest.html 缺少共享模拟引擎');
vm.runInThisContext(engine[1], { filename: 'loadtest.html:loadtest-engine' });
const { createRun, resources, homeResources } = globalThis.HojLoadtest;

function argumentsFor(argv) {
  const values = {}, flags = new Set(['self-test','prefetch','no-prefetch','use-cache','help']);
  const names = new Set(['host','users','ramp','max-queue-wait','load-timeout','timeout','poll','sla','out']);
  for (let i = 0; i < argv.length; i++) {
    const name = argv[i].replace(/^--/, '');
    if (!argv[i].startsWith('--') || (!flags.has(name) && !names.has(name))) throw new Error('未知参数：' + argv[i]);
    if (flags.has(name)) values[name] = true;
    else { if (!argv[i + 1] || argv[i + 1].startsWith('--')) throw new Error('参数缺少值：--' + name); values[name] = argv[++i]; }
  }
  return values;
}
function percentile(values, p) {
  if (!values.length) return null;
  const sorted = values.slice().sort((a,b) => a-b);
  return +(sorted[Math.min(sorted.length - 1, Math.ceil(sorted.length * p) - 1)] / 1000).toFixed(3);
}
function report(run, sla) {
  const counts = {};
  for (const u of run.users) counts[u.state] = (counts[u.state] || 0) + 1;
  const successes = run.users.filter(u => u.state === 'success');
  const failedRequests = run.trace.filter(r => r.error);
  const p95 = percentile(successes.map(u => u.firstScreenMs), .95);
  const pass = successes.length === run.users.length && p95 !== null && p95 <= sla;
  const rows = run.users.map(u => `| ${u.index + 1} | ${u.state} | ${u.arrivedAt ? ((u.arrivedAt-run.startedAt)/1000).toFixed(3) : '未到达'} | ${u.queueMs === undefined ? '—' : (u.queueMs/1000).toFixed(3)} | ${u.loadMs === undefined ? '—' : (u.loadMs/1000).toFixed(3)} | ${(u.error || u.cleanupState).replaceAll('|','/').replaceAll('\n',' ')} |`);
  return [
    `# BingOJ 资源请求模拟 ${new Date().toISOString()}`,
    '',
    `目标 ${run.options.host}；会话 ${run.users.length}；到达窗口 ${run.options.rampMs/1000} 秒；总耗时 ${((run.endedAt-run.startedAt)/1000).toFixed(2)} 秒。`,
    '单 Node 进程并发请求，不执行 Vue；不是独立浏览器或独立进程。',
    '幂等 GET 网络中断最多重试 2 次；成功可能包含重试恢复，失败请求仍累计。一次性票据不重试。',
    `状态：${JSON.stringify(counts)}。累计失败请求 ${run.errorCount}/${run.requestCount}；已取消请求 ${run.cancelledRequests}；预期匿名身份响应 ${run.expectedAnonymous}；门禁503降级 ${run.degradedGate}；解压后响应体 ${(run.bytes/1048576).toFixed(2)} MiB；保留最近 ${run.trace.length} 条明细（舍弃 ${run.droppedTrace} 条，累计计数不受影响）。`,
    `每会话最多并发 3 个资源请求；无全局发送队列，控制请求使用高优先级。最长 fetch 耗时 ${(run.maxFetchMs/1000).toFixed(3)} 秒（含浏览器排队），最大成功心跳间隔 ${(run.maxHeartbeatGapMs/1000).toFixed(3)} 秒。`,
    `缓存：${run.options.noCache ? 'cache:no-store；仅文档/资源随机查询参数，API保留真实参数及服务端缓存行为' : '允许缓存'}。`,
    `成功会话关键资源和首页接口完成 p95：${p95 ?? '无'} 秒；要求全部成功且 p95 ≤ ${sla} 秒：${pass ? 'PASS' : 'FAIL'}。`,
    '', '| 会话 | 状态 | 实际到达秒 | 排队秒 | 加载秒 | 原因/清理 |', '|---|---|---|---|---|---|', ...rows,
    ...(run.errorSummary.size ? ['', '全部失败累计分类：', ...[...run.errorSummary].map(([key, value]) => `- ${key}：${value.count} 次；最近原因：${value.message}`)] : []),
    ...(failedRequests.length ? ['', '保留的失败请求：', ...failedRequests.map(r => `- 会话 ${r.user} ${r.url}：${r.error}`)] : []),
  ].join('\n');
}

async function selfTest() {
  const { createServer } = await import('node:http');
  const assert = (await import('node:assert/strict')).default;
  let mode = 'ok', revision = 0, retryCount = 0;
  const members = new Map(), events = [];
  const wait = ms => new Promise(resolve => setTimeout(resolve, ms));
  const queueHTML = '<!doctype html><meta name="hoj-queue-page" content="true"><div id="position"></div>';
  const appHTML = '<!doctype html><base href="/assets/"><link rel=stylesheet href=style.css><script src=app.js></script><link rel=preload as=script href="app.js"><link rel=icon href=/favicon.ico><link rel=prefetch href="lazy.js">';
  function snapshot(id) {
    const all = [...members.values()];
    return { myId:id, state:members.get(id)?.state || 'closed', admitted:['reserved','loading'].includes(members.get(id)?.state), position:[...members.keys()].filter(k=>members.get(k).state==='waiting').indexOf(id)+1, waiting:all.filter(m=>m.state==='waiting').length, reserved:all.filter(m=>m.state==='reserved').length, loading:all.filter(m=>m.state==='loading').length, active:all.filter(m=>['reserved','loading'].includes(m.state)).length, total:all.filter(m=>m.state!=='closed').length, maxSlots:1, enabled:true, revision, retryAfterMs:10 };
  }
  const server = createServer(async (req,res) => {
    const url = new URL(req.url, 'http://fixture'), id=url.searchParams.get('qid');
    events.push({ path:url.pathname, search:url.search, id, method:req.method, at:Date.now(), headers:req.headers });
    function send(body, type='application/json', status=200) { if (!res.destroyed) { res.writeHead(status, {'Content-Type':type}); res.end(body); } }
    const json = data => send(JSON.stringify({status:200,data}));
    if (url.pathname === '/api/queue/status') {
      const retryStatuses = [503, 502, 504, 429];
      if (mode === 'retry-status' && retryCount < retryStatuses.length) return send('{}', 'application/json', retryStatuses[retryCount++]);
      if (mode === 'slow-status' && retryCount++ === 0) { await wait(4500); if (res.destroyed) return; }
      if (mode === 'disabled') return json({ ...snapshot(id), state: 'disabled', admitted: true, enabled: false, disabled: true });
      if (mode === 'late-status') await wait(90);
      if (!members.has(id)) { members.set(id,{state:'waiting'}); revision++; }
      const m=members.get(id);
      if (mode === 'expired-identity' && retryCount++ > 0) { m.state='closed'; revision++; }
      if (m.state==='waiting' && mode!=='queue-timeout' && mode!=='late-status' && mode!=='refresh' && mode!=='expired-identity' && ![...members.values()].some(x=>['reserved','loading'].includes(x.state))) { m.state='reserved'; revision++; }
      return json(snapshot(id));
    }
    if (url.pathname === '/api/queue/release' || url.pathname === '/api/queue/leave') {
      assert.equal(req.method,'POST'); members.set(id,{state:'closed'}); revision++; return json(snapshot(id));
    }
    if (url.pathname === '/api/queue/heartbeat') {
      assert.equal(req.method,'POST');
      if (mode === 'heartbeat-transient' && retryCount++ === 0) return send('{}', 'application/json', 503);
      if(mode==='heartbeat-inflight')await wait(150);
      return json(snapshot(id));
    }
    if (url.pathname === '/home') {
      const ticket=url.searchParams.get('hoj_ticket');
      if (mode === 'gate-network-once' && !ticket && retryCount++ === 0) return res.destroy();
      if (mode === 'disabled' && ticket) return send(appHTML,'text/html');
      if (!ticket || members.get(ticket)?.state!=='reserved') return send(queueHTML,'text/html',mode==='gate-degraded'?503:200);
      members.get(ticket).state='loading'; revision++;
      if (mode === 'claim-network-failure') return res.destroy();
      return send(appHTML,'text/html');
    }
    if (url.pathname === '/favicon.ico' && mode === 'gate-icon-network-once' && retryCount++ === 0) return res.destroy();
    if (url.pathname === '/api/classroom/user/roles') return send(JSON.stringify({ status:401, msg:'请先登录' }), 'application/json', 401);
    if (url.pathname === '/api/rating/contest/batch') {
      let body='';for await (const chunk of req) body+=chunk;
      assert.deepEqual(JSON.parse(body),{contestIds:[42,71]});
      if (mode === 'post-network-failure') return res.destroy();
      return send(JSON.stringify({code:200,data:{42:{isRating:true},71:{isRating:false}}}));
    }
    if (url.pathname.startsWith('/api/')) {
      let data={};
      if(mode==='css-api-images' || mode==='post-network-failure') {
        if(url.pathname==='/api/rating/rank') data={records:[{avatar:'/avatar.png'}]};
        if(url.pathname==='/api/get-common-announcement') data={records:[{content:'![公告](/announcement.png) <img src=/publicimg/fixture>'}]};
        if(url.pathname==='/api/get-website-config') data={description:'<img src=/footer.png>'};
        if(url.pathname==='/api/home-carousel') data=[{url:'/banner.png'}];
        if(url.pathname==='/api/get-recent-contest') data=[{id:42},{id:71}];
      }
      return send(JSON.stringify({[url.pathname.startsWith('/api/rating/')?'code':'status']:mode==='business-failure'?500:200,data}));
    }
    if (url.pathname.endsWith('app.js')) {
      if (['resource-network-once', 'resource-network-persistent', 'resource-retry-cancel'].includes(mode) && (retryCount++ === 0 || mode !== 'resource-network-once')) {
        res.writeHead(200, { 'Content-Type': 'application/javascript', 'Content-Length': '1000' });
        res.write('window.fixture='); await wait(10); return res.destroy();
      }
      if (mode==='http-failure') return send('unavailable','text/plain',503);
      if (mode==='html-resource') return send(queueHTML,'text/html');
      if (mode==='load-timeout') await wait(250);
      if (mode==='heartbeat' || mode==='heartbeat-transient') await wait(70);
      if (mode==='heartbeat-inflight') { while(!events.some(e=>e.path==='/api/queue/heartbeat')) await wait(2); await wait(5); }
      if (mode==='cancel-loading') await wait(180);
      return send('window.fixture=true;','application/javascript');
    }
    if (url.pathname.endsWith('.js')) return send('void 0;','application/javascript');
    if (url.pathname.endsWith('.css')) return send(mode==='css-api-images' ? (url.pathname.endsWith('style.css') ? '@import "nested.css";a{background:url(../bg.png)}@font-face{font-family:element-icons;src:url(font.woff2),url(unused.ttf)}@font-face{font-family:KaTeX_Main;src:url(unused.woff2)}' : '@import "style.css";b{background:url(../bg.png)}') : 'body{}','text/css');
    if (url.pathname.endsWith('.woff2')) return send('fixture-font','font/woff2');
    return send('image','image/x-icon');
  });
  await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
  const host='http://127.0.0.1:'+server.address().port;
  const options={host,users:1,rampMs:0,maxQueueMs:500,loadTimeoutMs:500,requestTimeoutMs:200,pollMs:10,heartbeatMs:20,cleanupTimeoutMs:100};
  async function check(name, action) {
    members.clear(); events.length=0; mode=name; revision=0; retryCount=0;
    await action(); console.log('PASS '+name);
  }
  try {
    const bundle='function js(){return "assets/js/"+{"chunk-clock":"12345678"}}function css(){return "assets/css/"+{"chunk-clock":"abcdef12"}}const ann=()=>Promise.resolve(),clock=()=>s.e("chunk-clock");var home={name:"home",components:{Announcements:ann,TimeDisplay:clock}};var logo="assets/img/logo.12345678.png";';
    assert.deepEqual(homeResources(bundle,host+'/assets/js/app.12345678.js').map(job=>job.url),[host+'/assets/js/chunk-clock.12345678.js',host+'/assets/css/chunk-clock.abcdef12.css',host+'/assets/img/logo.12345678.png']);
    assert.throws(()=>homeResources('unrecognized app',host+'/'),/无法解析线上 Home/);
    const parsed=resources(appHTML,host+'/home');
    assert.equal(parsed.length,4); assert.equal(parsed.filter(x=>x.url===host+'/assets/app.js').length,1); assert.equal(parsed.filter(x=>x.prefetch).length,1);
    await check('ok',async()=>{ const run=createRun({...options,prefetch:true}); await run.done; assert.equal(run.users[0].state,'success'); assert.equal(members.get(run.users[0].qid).state,'closed'); assert(events.some(e=>e.path==='/assets/lazy.js'));assert.equal(events.filter(e=>e.path==='/favicon.ico').length,2); });
    await check('gate-degraded',async()=>{const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');assert.equal(run.degradedGate,1);assert.equal(run.errorCount,0);assert(run.trace.some(e=>e.degradedGate&&e.status===503));});
    await check('disabled',async()=>{const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');assert.equal(run.users[0].disabled,true);assert.equal(events.filter(e=>e.path==='/home').length,2);});
    await check('retry-status',async()=>{const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');assert.equal(events.filter(e=>e.path==='/api/queue/status').length,5);assert.equal(run.errorCount,4);assert.equal(run.requestCount,run.trace.length);});
    await check('slow-status',async()=>{
      const run=createRun({...options,requestTimeoutMs:1800000,maxQueueMs:6000});await run.done;
      assert.equal(run.users[0].state,'success');assert.equal(run.options.requestTimeoutMs,1800000);
      const polls=events.filter(e=>e.path==='/api/queue/status');assert.equal(polls.length,2);
      assert(polls[1].at-polls[0].at>=3900 && polls[1].at-polls[0].at<4500);
      assert(run.trace.some(e=>e.url.includes('/api/queue/status')&&e.error.includes('请求超时')));
    });
    await check('expired-identity',async()=>{
      const run=createRun(options);await run.done;const u=run.users[0];
      assert.equal(u.state,'failed');assert.equal(u.requestsDone,0);assert.equal(u.cleanupState,'已清理');
      assert.match(u.error,/上次状态：排队等待/);assert.match(u.error,/距上次成功轮询/);
    });
    await check('gate-network-once', async () => {
      const run = createRun({ ...options, maxQueueMs: 2000 }); await run.done;
      assert.equal(run.users[0].state, 'success'); assert.equal(run.errorCount, 1); assert.equal(run.users[0].requestsDone, 9);
      assert.equal(events.filter(event => event.path === '/home' && !event.search.includes('hoj_ticket=')).length, 2);
      assert.equal(events.filter(event => event.path === '/api/queue/status').length, 1); assert.equal(run.users[0].cleanupState, '已清理');
    });
    await check('gate-icon-network-once', async () => {
      const run = createRun({ ...options, loadTimeoutMs: 2000 }); await run.done;
      assert.equal(run.users[0].state, 'success'); assert.equal(run.errorCount, 1); assert.equal(run.users[0].requestsDone, 9);
      assert.equal(events.filter(event => event.path === '/favicon.ico').length, 3); assert.equal(run.users[0].cleanupState, '已清理');
    });
    await check('claim-network-failure', async () => {
      const run = createRun(options); await run.done;
      assert.equal(run.users[0].state, 'failed'); assert.equal(run.errorCount, 1); assert.equal(run.users[0].requestsDone, 0);
      assert.equal(events.filter(event => event.path === '/home' && event.search.includes('hoj_ticket=')).length, 1);
      assert.equal(run.users[0].cleanupState, '已清理'); assert.equal(members.get(run.users[0].qid).state, 'closed');
    });
    await check('resource-network-once', async () => {
      const run = createRun({ ...options, loadTimeoutMs: 2000 }); await run.done;
      assert.equal(run.users[0].state, 'success'); assert.equal(run.users[0].requestsDone, 9); assert.equal(run.errorCount, 1);
      const attempts = run.trace.filter(entry => new URL(entry.url).pathname === '/assets/app.js');
      assert.equal(attempts.length, 2); assert.equal(attempts[0].status, 200); assert.match(attempts[0].error, /网络错误/); assert.equal(attempts[1].error, '');
      assert.equal(run.users[0].cleanupState, '已清理');
    });
    await check('resource-network-persistent', async () => {
      const run = createRun({ ...options, loadTimeoutMs: 2000 }); await run.done;
      assert.equal(run.users[0].state, 'failed'); assert.equal(run.errorCount, 3);
      assert.equal(events.filter(event => event.path === '/assets/app.js').length, 3);
      assert.equal(run.users[0].cleanupState, '已清理'); assert.equal(members.get(run.users[0].qid).state, 'closed');
    });
    await check('resource-retry-cancel', async () => {
      const run = createRun({ ...options, loadTimeoutMs: 2000 });
      while (!run.users[0].lastError?.startsWith('请求重试')) await wait(2);
      const cancelledAt = Date.now(); run.cancel(0); await run.done;
      assert(Date.now() - cancelledAt < 200); assert.equal(run.users[0].state, 'cancelled'); assert.equal(run.errorCount, 1);
      assert.equal(events.filter(event => event.path === '/assets/app.js').length, 1); assert.equal(run.users[0].cleanupState, '已清理');
    });
    await check('post-network-failure', async () => {
      const run = createRun(options); await run.done;
      assert.equal(run.users[0].state, 'failed'); assert.equal(run.errorCount, 1);
      assert.equal(events.filter(event => event.path === '/api/rating/contest/batch').length, 1); assert.equal(run.users[0].cleanupState, '已清理');
    });
    await check('css-api-images',async()=>{const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');for(const path of ['/assets/nested.css','/assets/font.woff2','/avatar.png','/announcement.png','/footer.png','/publicimg/fixture','/api/rating/contest/batch'])assert.equal(events.filter(e=>e.path===path).length,1,path);assert.equal(events.some(e=>['/banner.png','/assets/unused.ttf','/assets/unused.woff2','/bg.png'].includes(e.path)),false);assert.equal(run.expectedAnonymous,1);assert(events.filter(e=>e.path.startsWith('/api/')).every(e=>!e.search.includes('_cb=')));assert(events.some(e=>e.path==='/api/rating/rank'&&e.search==='?page=1&limit=10&keyword='));assert(events.some(e=>e.path==='/api/get-common-announcement'&&e.search==='?currentPage=1&limit=5'));assert(events.some(e=>e.path==='/api/get-website-config'&&e.headers['url-type']==='general'));});
    for (const name of ['http-failure','html-resource','business-failure']) await check(name,async()=>{const run=createRun(options); await run.done; assert.equal(run.users[0].state,'failed'); assert.equal(members.get(run.users[0].qid).state,'closed');});
    await check('ramp',async()=>{const run=createRun({...options,users:3,rampMs:120}); await wait(35); assert.equal(run.users[1].arrived,false); assert.equal(run.users[2].arrived,false); await run.done; assert(run.users[1].arrivedAt-run.startedAt>=50); assert(run.users[2].arrivedAt-run.startedAt>=110); assert(run.users.every(u=>u.state==='success'));});
    await check('refresh',async()=>{const run=createRun({...options,maxQueueMs:90});while(run.users[0].position===null)await wait(2);const u=run.users[0],arrival=u.arrivedAt,qid=u.qid;await wait(25);assert.equal(run.refresh(0),true);assert.equal(u.qid,qid);assert.equal(u.arrivedAt,arrival);assert.equal(u.refreshEvidence.position,1);assert.equal(events.some(e=>e.path==='/api/queue/leave'),false);await run.done;assert.equal(u.state,'timeout');assert.equal(u.refreshCount,1);assert(u.finishedAt-arrival<125);assert(events.filter(e=>e.path==='/api/queue/status').every(e=>e.id===qid));});
    await check('queue-timeout',async()=>{const run=createRun({...options,maxQueueMs:55});await run.done;assert.equal(run.users[0].state,'timeout');assert.equal(members.get(run.users[0].qid).state,'closed');});
    await check('load-timeout',async()=>{const run=createRun({...options,loadTimeoutMs:50});await run.done;assert.equal(run.users[0].state,'timeout');assert.equal(members.get(run.users[0].qid).state,'closed');});
    await check('heartbeat',async()=>{const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');assert(events.some(e=>e.path==='/api/queue/heartbeat'));});
    await check('heartbeat-transient',async()=>{
      const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');
      const heartbeats=run.trace.filter(e=>e.url.includes('/api/queue/heartbeat'));
      assert(heartbeats.some(e=>e.status===503));assert(heartbeats.some(e=>e.status===200));assert.equal(run.errorCount,1);
    });
    await check('heartbeat-inflight',async()=>{const run=createRun(options);await run.done;assert.equal(run.users[0].state,'success');assert.equal(run.errorCount,0);assert.equal(run.cancelledRequests,1);const cancelled=run.trace.filter(e=>e.cancelled);assert.equal(cancelled.length,1);assert(cancelled[0].url.includes('/api/queue/heartbeat'));assert.equal(cancelled[0].note,'心跳结束');assert.equal(cancelled[0].error,'');});
    await check('late-status',async()=>{
      const run=createRun({...options,users:2,rampMs:300});
      while(!events.some(e=>e.path==='/api/queue/status')) await wait(2);
      run.stop(); await run.done; await wait(100);
      assert(run.users.every(u=>u.state==='cancelled')); assert.equal(run.users[1].arrived,false);
      assert.equal(members.get(run.users[0].qid).state,'closed'); assert.equal(members.has(run.users[1].qid),false);
      assert.equal(events.filter(e=>e.path==='/api/queue/status').length,1);
    });
    await check('cancel-loading',async()=>{const run=createRun(options);while(run.users[0].state!=='loading')await wait(2);run.cancel(0);await run.done;assert.equal(run.users[0].state,'cancelled');assert.equal(members.get(run.users[0].qid).state,'closed');});
    await check('restart',async()=>{
      const old=createRun({...options,users:2,rampMs:300});old.stop();
      const fresh=createRun(options);await Promise.all([old.done,fresh.done]);
      assert(old.users.every(u=>u.state==='cancelled'));assert.equal(fresh.users[0].state,'success');assert.notEqual(old.users[0].qid,fresh.users[0].qid);
      await wait(320);assert.equal(events.filter(e=>e.path==='/api/queue/status').length,1);
    });
    await dispatchTest();
    console.log('All local HTTP fixture checks passed; no online requests.');
  } finally { server.closeAllConnections(); await new Promise(resolve=>server.close(resolve)); }
}

async function dispatchTest() {
  const { createServer } = await import('node:http');
  const assert = (await import('node:assert/strict')).default;
  // 固定旧版的 6/2 发送队列只用于复现；产品引擎不再设全局并发预算。
  const legacyQueue = `
    let activeRequests = 0, activeResources = 0;
    const controlQueue = [], resourceQueue = [];
    function drainRequests() {
      while (activeRequests < 6) {
        const next = activeResources < 2 && resourceQueue.length ? resourceQueue.shift() : controlQueue.shift();
        if (!next) return;
        next.start();
      }
    }
    function acquireRequest(control, signal) {
      return new Promise((resolve, reject) => {
        if (signal.aborted) return reject(abortError(signal));
        const queue = control ? controlQueue : resourceQueue;
        const cancel = () => { const index = queue.indexOf(job); if (index >= 0) queue.splice(index, 1); signal.removeEventListener('abort', cancel); reject(abortError(signal)); drainRequests(); };
        const job = { start() { signal.removeEventListener('abort', cancel); activeRequests++; if (!control) activeResources++; resolve(() => { activeRequests--; if (!control) activeResources--; drainRequests(); }); } };
        signal.addEventListener('abort', cancel, { once: true }); queue.push(job); drainRequests();
      });
    }
  `;
  const legacyEngine = engine[1].replace('  function createRun(options)', legacyQueue + '\n  function createRun(options)')
    .replace('let cancelTimeout;', 'let cancelTimeout, release;')
    .replace("(kind === 'html' && url.searchParams.has('hoj_ticket'))", "kind === 'html'")
    .replace('      try {\n        cancelTimeout =', '      try {\n        release = await acquireRequest(control, ctrl.signal);\n        if (ctrl.signal.aborted) throw abortError(ctrl.signal);\n        cancelTimeout =')
    .replace('        run.activeRequests--;', '        release?.();\n        run.activeRequests--;');
  const cases = [
    { label: 'legacy-6-2', rtt: 200, jitter: 0, ramp: 5000, legacy: true },
    { label: 'current-200ms', rtt: 200, jitter: 0, ramp: 20000 },
    { label: 'current-500ms', rtt: 500, jitter: 0, ramp: 20000 },
    { label: 'current-1000ms-jitter', rtt: 1000, jitter: 400, ramp: 20000 },
    { label: 'current-5s-burst', rtt: 1000, jitter: 400, ramp: 5000 }
  ];
  for (const scenario of cases) {
    // 同比缩放客户端时钟、网络、下载和20/15/60秒服务端租约；每组观测120秒。
    const scale = 0.1, clock = () => Date.now() / scale;
    const wait = ms => new Promise(resolve => setTimeout(resolve, ms * scale));
    const sandbox = { AbortController, TextDecoder, URL, crypto, fetch, performance: { now: () => performance.now() / scale }, Date: class extends Date { static now() { return clock(); } }, setTimeout: (callback, ms) => setTimeout(callback, ms * scale), clearTimeout };
    vm.runInNewContext(scenario.legacy ? legacyEngine : engine[1], sandbox);
    const members = new Map(), expired = [], gate = '<!doctype html><meta name="hoj-queue-page">';
    let revision = 0, maxHeartbeatMs = 0, maxLoadingHeartbeatMs = 0, seed = 12345;
    const jitter = () => { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return (seed / 4294967296 * 2 - 1) * scenario.jitter; };
    const sweep = () => {
      const now = clock();
      for (const [qid, member] of members) {
        const deadline = member.state === 'waiting' ? member.seen + 20000 : member.state === 'reserved' ? Math.min(member.seen + 20000, member.reservedUntil) : member.state === 'loading' ? member.loadingUntil : Infinity;
        if (deadline <= now) { expired.push({ qid, state: member.state }); member.state = 'closed'; }
      }
    };
    const promote = () => {
      let free = 5 - [...members.values()].filter(member => ['reserved', 'loading'].includes(member.state)).length;
      for (const member of members.values()) if (free > 0 && member.state === 'waiting') { member.state = 'reserved'; member.reservedUntil = Math.min(clock() + 15000, member.seen + 20000); free--; }
    };
    const server = createServer(async (req, res) => {
      const url = new URL(req.url, 'http://fixture'), id = url.searchParams.get('qid');
      const send = (body, type = 'application/json') => { if (!res.destroyed) { res.writeHead(200, { 'Content-Type': type }); res.end(body); } };
      const json = () => {
        const position = [...members.keys()].filter(qid => members.get(qid).state === 'waiting').indexOf(id) + 1;
        send(JSON.stringify({ status: 200, data: { myId: id, state: members.get(id)?.state || 'closed', enabled: true, position, revision: ++revision, retryAfterMs: position > 100 ? 5000 : position > 20 ? 3000 : 1000 } }));
      };
      const control = url.pathname.startsWith('/api/queue/') || url.searchParams.has('hoj_ticket');
      await wait(control ? Math.max(1, scenario.rtt + jitter()) : 100);
      sweep();
      if (url.pathname === '/api/queue/status') {
        if (!members.has(id)) members.set(id, { state: 'waiting', seen: clock() });
        const member = members.get(id);
        if (['waiting', 'reserved'].includes(member.state)) {
          maxHeartbeatMs = Math.max(maxHeartbeatMs, clock() - member.seen); member.seen = clock();
          if (member.state === 'reserved') member.reservedUntil = clock() + 15000;
        }
        promote(); return json();
      }
      if (url.pathname === '/api/queue/release' || url.pathname === '/api/queue/leave') { if (members.has(id)) members.get(id).state = 'closed'; else members.set(id, { state: 'closed' }); promote(); return json(); }
      if (url.pathname === '/api/queue/heartbeat') {
        const member = members.get(id);
        if (member?.state === 'loading') { maxLoadingHeartbeatMs = Math.max(maxLoadingHeartbeatMs, clock() - member.loadingSeen); member.loadingSeen = clock(); member.loadingUntil = clock() + 60000; }
        promote(); return json();
      }
      if (url.pathname === '/home') {
        const member = members.get(url.searchParams.get('hoj_ticket'));
        promote();
        if (member?.state === 'reserved') { member.state = 'loading'; member.loadingSeen = clock(); member.loadingUntil = clock() + 60000; return send('<!doctype html><script src="/assets/app1.js"></script><script src="/assets/app2.js"></script><script src="/assets/app3.js"></script>', 'text/html'); }
        return send(gate, 'text/html');
      }
      if (url.pathname.startsWith('/assets/')) { await wait(10000); return send('void 0;', 'application/javascript'); }
      if (url.pathname.startsWith('/api/')) return send(JSON.stringify({ status: 200, data: {} }));
      return send('fixture', 'image/png');
    });
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    const host = 'http://127.0.0.1:' + server.address().port;
    let run;
    try {
      run = sandbox.HojLoadtest.createRun({ host, users: 400, rampMs: scenario.ramp, maxQueueMs: 1800000, loadTimeoutMs: 1800000, requestTimeoutMs: 1800000 });
      await wait(120000);
      assert.equal(run.users.length, 400); assert.equal(members.size, 400);
      const successes = run.users.filter(user => user.state === 'success').length;
      const failed = run.users.filter(user => ['failed', 'timeout'].includes(user.state)).length;
      if (scenario.legacy) {
        assert(expired.length >= 100, '旧调度应稳定复现大量身份过期');
        assert(failed >= 100, '旧调度应稳定复现大量失败');
      } else {
        assert.equal(expired.length, 0); assert.equal(run.errorCount, 0); assert.equal(failed, 0);
        assert(successes > 0); assert(maxHeartbeatMs < 20000); assert(maxLoadingHeartbeatMs < 60000);
      }
      console.log(`PASS dispatch-${scenario.label}: 400 identities, ramp ${scenario.ramp/1000}s, 120 simulated seconds, 3 concurrent 10-second resources/page; successes ${successes}, expired ${expired.length}, failures ${failed}, errors ${run.errorCount}, waiting heartbeat gap ${(maxHeartbeatMs/1000).toFixed(2)}s, loading heartbeat gap ${(maxLoadingHeartbeatMs/1000).toFixed(2)}s`);
      run.stop(); await run.done;
      assert.equal(run.activeRequests, 0); assert.equal(run.activeControlRequests, 0); assert.equal(run.activeResourceRequests, 0);
      assert.equal([...members.values()].filter(member => member.state !== 'closed').length, 0);
      if (!scenario.legacy) {
        const first = sandbox.HojLoadtest.createRun({ host, users: 30, rampMs: 0, maxQueueMs: 1800000 });
        await wait(200);
        const second = sandbox.HojLoadtest.createRun({ host, users: 1, rampMs: 0 });
        first.stop(); second.stop(); await Promise.all([first.done, second.done]);
        assert(first.users.every(user => user.state === 'cancelled')); assert(second.users.every(user => user.state === 'cancelled'));
        assert.equal(first.activeRequests, 0); assert.equal(second.activeRequests, 0);
        assert.equal([...members.values()].filter(member => member.state !== 'closed').length, 0);
      }
    } finally {
      run?.stop(); if (run) await run.done;
      server.closeAllConnections(); await new Promise(resolve => server.close(resolve));
    }
  }
  console.log('PASS dispatch-cancel-and-restart: all requests cancelled and identities cleaned; no leaked requests');
}

async function main() {
  const args=argumentsFor(process.argv.slice(2));
  if (args.help) { console.log('node scripts/loadtest.mjs --users 25 --ramp 10 --host https://bingoj.cn [--prefetch] [--out report.md]\nnode scripts/loadtest.mjs --self-test\n--max-queue-wait 秒 --load-timeout 秒 --timeout 资源请求超时秒 --poll 毫秒 --sla 秒 --use-cache'); return; }
  if (args['self-test']) return selfTest();
  const number=(name,fallback)=>args[name]===undefined?fallback:Number(args[name]);
  const sla=number('sla',10); if(!Number.isFinite(sla)||sla<=0)throw new Error('--sla 必须为正数');
  const run=createRun({host:args.host||'https://bingoj.cn',users:number('users',1),rampMs:number('ramp',0)*1000,maxQueueMs:number('max-queue-wait',300)*1000,loadTimeoutMs:number('load-timeout',120)*1000,requestTimeoutMs:number('timeout',15)*1000,pollMs:number('poll',1000),prefetch:!!args.prefetch&&!args['no-prefetch'],noCache:!args['use-cache']});
  console.log(`开始资源请求模拟：${run.options.host}，${run.users.length} 个虚拟会话。Ctrl+C 后等待清理。`);
  const interrupt=()=>{console.log('已中止，正在清理已到达会话。');run.stop();};
  process.on('SIGINT',interrupt);process.on('SIGTERM',interrupt);
  await run.done;process.off('SIGINT',interrupt);process.off('SIGTERM',interrupt);
  const output=report(run,sla);console.log(output);
  if(args.out)fs.writeFileSync(args.out,output+'\n');
  const success=run.users.filter(u=>u.state==='success');
  process.exitCode=success.length===run.users.length && percentile(success.map(u=>u.firstScreenMs),.95)<=sla ? 0 : 1;
}
main().catch(error=>{console.error(error.stack||error.message);process.exitCode=1;});
