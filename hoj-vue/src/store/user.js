import { USER_TYPE } from '@/common/constants'
import storage from '@/common/storage'
import api from '@/common/api'
let storedToken = localStorage.getItem('token')
try {
  if (storedToken) {
    const payload = JSON.parse(atob(storedToken.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    // 只剔除已明确过期的凭据；权限仍由服务器验证。
    if (typeof payload.exp === 'number' && payload.exp * 1000 <= Date.now()) {
      storage.remove('token')
      storage.remove('userInfo')
      storedToken = null
    }
  }
} catch (_) { /* 无法解析的凭据交由服务器验证。 */ }

let authRequest = null
const state = {
  userInfo: storage.get('userInfo'),
  token: storedToken,
  loginFailNum:0,
  unreadMessage:{
    comment:0,
    reply:0,
    like:0,
    sys:0,
    mine:0
  }
}

const getters = {
  userInfo: state => state.userInfo || {},
  token: state => state.token ||'',
  unreadMessage:state => state.unreadMessage || {},
  loginFailNum:state=>state.loginFailNum || 0,
  isAuthenticated: (state, getters) => {
    return !!getters.token
  },
  isAdminRole: (state, getters) => {
    if(getters.userInfo.roleList){
      return getters.userInfo.roleList.indexOf(USER_TYPE.ADMIN)!=-1 ||
        getters.userInfo.roleList.indexOf(USER_TYPE.PROBLEM_ADMIN)!=-1 ||
        getters.userInfo.roleList.indexOf(USER_TYPE.SUPER_ADMIN)!=-1
    }else{
      return false;
    }
  },
  isSuperAdmin: (state, getters) => {
    if(getters.userInfo.roleList){
      return getters.userInfo.roleList.indexOf(USER_TYPE.SUPER_ADMIN) !=-1
    }else{
      return false;
    }
  },
  isProblemAdmin:(state, getters) => {
    if(getters.userInfo.roleList){
      return getters.userInfo.roleList.indexOf(USER_TYPE.PROBLEM_ADMIN) !=-1
    }else{
      return false;
    }
  },
}

const mutations = {
  changeUserInfo(state, {userInfo}) {
    state.userInfo = userInfo
    storage.set('userInfo',userInfo)
  },
  changeUserToken(state,token){
    state.token = token
    localStorage.setItem("token",token)
  },
  incrLoginFailNum(state,{success}){
    if(!success){
      state.loginFailNum +=1
    }else{
      state.loginFailNum = 0
    }
  },
  
  clearUserInfoAndToken(state){
    state.token = ''
    state.userInfo = {}
    state.loginFailNum = 0
    state.unreadMessage = { comment: 0, reply: 0, like: 0, sys: 0, mine: 0 }
    storage.remove('token')
    storage.remove('userInfo')
  },
  updateUnreadMessageCount(state, {unreadMessage}){
    state.unreadMessage = unreadMessage
  },
  substractUnreadMessageCount(state,{needSubstractMsg}){
    // 负数也没关系
    state.unreadMessage[needSubstractMsg.name] = state.unreadMessage[needSubstractMsg.name]-needSubstractMsg.num;
  },
  changeUserAuthInfo(state, {roles}){
    state.userInfo.roleList = roles;
    storage.set('userInfo', state.userInfo);
  }
}

const actions = {
  setUserInfo ({commit},userInfo) {
      commit('changeUserInfo', {
        userInfo: userInfo
      })
  },
  incrLoginFailNum({commit},success){
    commit('incrLoginFailNum',{success:success})
  },
  clearUserInfoAndToken ({commit}) {
    commit('clearUserInfoAndToken')
  },
  updateUnreadMessageCount({commit},unreadMessage){
    commit('updateUnreadMessageCount', {
      unreadMessage: unreadMessage
    })
  },
  substractUnreadMessageCount({commit},needSubstractMsg){
    commit('substractUnreadMessageCount', {
      needSubstractMsg: needSubstractMsg
    })
  },
  refreshUserAuthInfo({commit, state}){
    const token = state.token
    const uid = state.userInfo && state.userInfo.uid
    if (!token) return Promise.resolve(null)
    if (authRequest && authRequest.token === token && authRequest.uid === uid) {
      return authRequest.promise
    }
    const request = { token, uid }
    request.promise = api.getUserAuthInfo().then((res) => {
      const refreshedToken = res.headers && res.headers['refresh-token'] && res.headers.authorization
      if (!state.token || (state.token !== token && state.token !== refreshedToken) ||
          !state.userInfo || state.userInfo.uid !== uid) return null
      commit('changeUserAuthInfo', {roles: res.data.data.roles})
      return res
    }).finally(() => {
      if (authRequest === request) authRequest = null
    })
    authRequest = request
    return request.promise
  }
}

export default {
  state,
  getters,
  actions,
  mutations
}
