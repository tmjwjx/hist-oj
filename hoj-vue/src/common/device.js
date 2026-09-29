/**
 * 设备ID：首次访问时生成的随机UUID，持久化在 localStorage。
 *
 * 故意不使用任何浏览器指纹成分（UA/屏幕/时区/Canvas 等）：
 * 机房电脑软硬件配置完全相同，指纹哈希会把整间机房判成"同一台设备"。
 * 随机UUID下每台电脑/浏览器各不相同，天然满足机房场景；
 * 换浏览器或清空站点数据会得到新ID，反而触发"一账号多设备"规则，无法用于绕过检测。
 */

const DEVICE_ID_KEY = 'bingoj_device_id'

export function getDeviceId() {
  try {
    let deviceId = localStorage.getItem(DEVICE_ID_KEY)
    if (!deviceId) {
      deviceId = generateUuid()
      localStorage.setItem(DEVICE_ID_KEY, deviceId)
    }
    return deviceId
  } catch (e) {
    // localStorage 不可用（隐私模式极端情况）时退化为会话级随机ID
    return generateUuid()
  }
}

function generateUuid() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  // RFC4122 v4 兜底实现
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : ((r & 0x3) | 0x8)
    return v.toString(16)
  })
}

export default { getDeviceId }
