import axios from 'axios';
import store from '@/store';

// 创建专门用于对战API的axios实例
const battleRequest = axios.create({
  baseURL: '/battle-api',
  timeout: 30000
});

// 请求拦截器
battleRequest.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = token;
    }

    // 自动添加用户信息
    const userInfo = store.state.user.userInfo;
    if (userInfo && userInfo.uid) {
      // POST 请求：添加到请求体
      if ((config.method === 'post' || config.method === 'POST') && config.data) {
        if (!config.data.userId) {
          config.data.userId = userInfo.uid;
        }
        if (!config.data.username && userInfo.username) {
          config.data.username = userInfo.username;
        }
      }
      // GET 请求：添加到查询参数
      else if ((config.method === 'get' || config.method === 'GET') && config.params) {
        if (!config.params.userId) {
          config.params.userId = userInfo.uid;
        }
      }
    }

    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 响应拦截器
battleRequest.interceptors.response.use(
  response => {
    return response;
  },
  error => {
    // 静默处理业务逻辑错误（400），只在服务器错误（500）时打印
    if (error.response && error.response.status >= 500) {
      console.error('Battle API Server Error:', error);
    }
    return Promise.reject(error);
  }
);

/**
 * 创建对战房间
 */
export function createRoom() {
  return battleRequest({
    url: '/battle/create-room',
    method: 'post',
    data: {}
  });
}

/**
 * 加入对战房间
 * @param {Object} data - { roomId: string }
 */
export function joinRoom(data) {
  return battleRequest({
    url: '/battle/join-room',
    method: 'post',
    data
  });
}

/**
 * 获取房间信息
 * @param {String} roomId - 房间号
 */
export function getRoomInfo(roomId) {
  return battleRequest({
    url: '/battle/room-info',
    method: 'get',
    params: { roomId }
  });
}

/**
 * 开始对战
 * @param {Object} data - { roomId: string }
 */
export function startBattle(data) {
  return battleRequest({
    url: '/battle/start-battle',
    method: 'post',
    data
  });
}

/**
 * 放弃对战
 * @param {Object} data - { roomId: string }
 */
export function giveupBattle(data) {
  return battleRequest({
    url: '/battle/giveup',
    method: 'post',
    data
  });
}

/**
 * 解散房间
 * @param {Object} data - { roomId: string }
 */
export function dissolveRoom(data) {
  return battleRequest({
    url: '/battle/dissolve-room',
    method: 'post',
    data
  });
}

/**
 * 退出房间（挑战者退出等待中的房间）
 * @param {Object} data - { roomId: string }
 */
export function leaveRoom(data) {
  return battleRequest({
    url: '/battle/leave-room',
    method: 'post',
    data
  });
}

/**
 * AC提交（用户在题目中AC后调用）
 * @param {Object} data - { roomId: string, problemId: string }
 */
export function submitAC(data) {
  return battleRequest({
    url: '/battle/submit-ac',
    method: 'post',
    data
  });
}

/**
 * 获取我的对战记录
 * @param {Object} params - { limit: number, currentPage: number }
 */
export function getMyRecords(params) {
  return battleRequest({
    url: '/battle/my-records',
    method: 'get',
    params
  });
}

/**
 * 获取所有对战记录（管理员功能）
 * @param {Object} params - { limit: number, currentPage: number, username?: string, roomId?: string, isWinner?: string }
 */
export function getAllBattleRecords(params) {
  return battleRequest({
    url: '/battle/all-records',
    method: 'get',
    params
  });
}

/**
 * 获取对战排行榜
 * @param {Object} params - { limit: number, currentPage: number, username?: string }
 */
export function getBattleRank(params) {
  return battleRequest({
    url: '/battle/rank',
    method: 'get',
    params
  });
}

/**
 * 重置房间（再来一局）
 * @param {Object} data - { roomId: string }
 */
export function resetRoom(data) {
  return battleRequest({
    url: '/battle/reset-room',
    method: 'post',
    data
  });
}
