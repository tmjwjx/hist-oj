import axios from 'axios';

// Anonymous requests must never attach account credentials or trigger login dialogs.
const anonymous = axios.create({ timeout: 30000, withCredentials: false });

async function request(config, admin = false) {
  try {
    const response = await (admin ? axios : anonymous)(config);
    const result = response.data;
    const status = result.status === undefined ? result.code : result.status;
    if (status !== undefined && status !== 200) {
      const error = new Error(result.msg || result.message || '请求失败');
      error.status = status;
      throw error;
    }
    return result.data;
  } catch (error) {
    const response = error.response || (error.data ? error : null);
    if (response) {
      const data = response.data || {};
      const normalized = new Error(data.msg || data.message || '请求失败，请稍后重试');
      normalized.status = data.status || data.code || response.status;
      throw normalized;
    }
    throw error;
  }
}

const path = token => '/api/external-verification/' + encodeURIComponent(token);

export default {
  getLink(pid) {
    return request({ url: '/api/admin/problem/external-verification', params: { pid } }, true);
  },
  setLink(pid, enabled) {
    return request({ method: 'put', url: '/api/admin/problem/external-verification', data: { pid, enabled } }, true);
  },
  getProblem(token) {
    return request({ url: path(token) });
  },
  submit(token, data) {
    return request({ method: 'post', url: path(token) + '/submit', data });
  },
  getSubmissions({ pid, token, visitorToken, currentPage, limit }) {
    return request(pid
      ? { url: '/api/admin/problem/external-submissions', params: { pid, currentPage, limit } }
      : { url: path(token) + '/submissions', params: { currentPage, limit }, headers: { 'X-External-Visitor-Token': visitorToken } }, Boolean(pid));
  },
  getSubmission({ pid, token, visitorToken, submitId }) {
    return request(pid
      ? { url: '/api/admin/problem/external-submission', params: { pid, submitId } }
      : { url: path(token) + '/submissions/' + encodeURIComponent(submitId), headers: { 'X-External-Visitor-Token': visitorToken } }, Boolean(pid));
  },
};
