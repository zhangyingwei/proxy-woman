// 时间工具函数

/**
 * 格式化相对时间（x秒前、x分钟前等）
 */
export function formatRelativeTime(timestamp: string | number | Date): string {
  const now = new Date();
  const time = new Date(timestamp);
  
  // 如果时间无效，返回默认值
  if (isNaN(time.getTime())) {
    return '-';
  }
  
  const diffMs = now.getTime() - time.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);
  
  // 如果是未来时间，显示为"刚刚"
  if (diffMs < 0) {
    return '刚刚';
  }
  
  // 小于1分钟
  if (diffSeconds < 60) {
    return diffSeconds <= 0 ? '刚刚' : `${diffSeconds}秒前`;
  }
  
  // 小于1小时
  if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`;
  }
  
  // 小于24小时
  if (diffHours < 24) {
    return `${diffHours}小时前`;
  }
  
  // 小于7天
  if (diffDays < 7) {
    return `${diffDays}天前`;
  }
  
  // 超过7天，显示具体日期
  return time.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

/**
 * 格式化绝对时间
 */
export function formatAbsoluteTime(timestamp: string | number | Date): string {
  const time = new Date(timestamp);
  
  if (isNaN(time.getTime())) {
    return '-';
  }
  
  return time.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

/**
 * 格式化持续时间（毫秒转换为可读格式）
 */
export function formatDuration(nanoseconds: number): string {
  const ms = nanoseconds / 1000000;
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

/**
 * 格式化文件大小
 */
export function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}
