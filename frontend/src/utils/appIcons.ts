// 应用图标映射工具

export interface AppIconInfo {
  icon: string;
  color: string;
  name: string;
}

// 应用图标映射表
const APP_ICONS: Record<string, AppIconInfo> = {
  // 浏览器
  'chrome': { icon: '🌐', color: '#4285F4', name: 'Chrome' },
  'firefox': { icon: '🦊', color: '#FF7139', name: 'Firefox' },
  'safari': { icon: '🧭', color: '#006CFF', name: 'Safari' },
  'edge': { icon: '🌊', color: '#0078D4', name: 'Edge' },
  'opera': { icon: '🎭', color: '#FF1B2D', name: 'Opera' },
  
  // 开发工具
  'vscode': { icon: '💻', color: '#007ACC', name: 'VS Code' },
  'webstorm': { icon: '🔧', color: '#000000', name: 'WebStorm' },
  'sublime': { icon: '📝', color: '#FF9800', name: 'Sublime Text' },
  'atom': { icon: '⚛️', color: '#66595C', name: 'Atom' },
  'vim': { icon: '📄', color: '#019733', name: 'Vim' },
  
  // 网络工具
  'postman': { icon: '📮', color: '#FF6C37', name: 'Postman' },
  'insomnia': { icon: '😴', color: '#4000BF', name: 'Insomnia' },
  'curl': { icon: '🌀', color: '#073551', name: 'cURL' },
  'wget': { icon: '⬇️', color: '#2E8B57', name: 'Wget' },
  
  // 移动应用
  'ios': { icon: '📱', color: '#007AFF', name: 'iOS App' },
  'android': { icon: '🤖', color: '#3DDC84', name: 'Android App' },
  'flutter': { icon: '🐦', color: '#02569B', name: 'Flutter' },
  'react-native': { icon: '⚛️', color: '#61DAFB', name: 'React Native' },
  
  // 桌面应用
  'electron': { icon: '⚡', color: '#47848F', name: 'Electron' },
  'tauri': { icon: '🦀', color: '#FFC131', name: 'Tauri' },
  'qt': { icon: '🔷', color: '#41CD52', name: 'Qt' },
  'gtk': { icon: '🏠', color: '#4A90E2', name: 'GTK' },
  
  // 编程语言/运行时
  'node': { icon: '🟢', color: '#339933', name: 'Node.js' },
  'python': { icon: '🐍', color: '#3776AB', name: 'Python' },
  'java': { icon: '☕', color: '#ED8B00', name: 'Java' },
  'go': { icon: '🐹', color: '#00ADD8', name: 'Go' },
  'rust': { icon: '🦀', color: '#000000', name: 'Rust' },
  'php': { icon: '🐘', color: '#777BB4', name: 'PHP' },
  'ruby': { icon: '💎', color: '#CC342D', name: 'Ruby' },
  'dotnet': { icon: '🔷', color: '#512BD4', name: '.NET' },
  
  // 框架
  'react': { icon: '⚛️', color: '#61DAFB', name: 'React' },
  'vue': { icon: '💚', color: '#4FC08D', name: 'Vue.js' },
  'angular': { icon: '🅰️', color: '#DD0031', name: 'Angular' },
  'svelte': { icon: '🔥', color: '#FF3E00', name: 'Svelte' },
  'next': { icon: '▲', color: '#000000', name: 'Next.js' },
  'nuxt': { icon: '💚', color: '#00C58E', name: 'Nuxt.js' },
  
  // 数据库
  'mysql': { icon: '🐬', color: '#4479A1', name: 'MySQL' },
  'postgresql': { icon: '🐘', color: '#336791', name: 'PostgreSQL' },
  'mongodb': { icon: '🍃', color: '#47A248', name: 'MongoDB' },
  'redis': { icon: '🔴', color: '#DC382D', name: 'Redis' },
  'sqlite': { icon: '💾', color: '#003B57', name: 'SQLite' },
  
  // 云服务
  'aws': { icon: '☁️', color: '#FF9900', name: 'AWS' },
  'azure': { icon: '☁️', color: '#0078D4', name: 'Azure' },
  'gcp': { icon: '☁️', color: '#4285F4', name: 'Google Cloud' },
  'docker': { icon: '🐳', color: '#2496ED', name: 'Docker' },
  'kubernetes': { icon: '⚓', color: '#326CE5', name: 'Kubernetes' },
  
  // 通信工具
  'slack': { icon: '💬', color: '#4A154B', name: 'Slack' },
  'discord': { icon: '🎮', color: '#5865F2', name: 'Discord' },
  'telegram': { icon: '✈️', color: '#0088CC', name: 'Telegram' },
  'whatsapp': { icon: '💬', color: '#25D366', name: 'WhatsApp' },
  
  // 媒体工具
  'spotify': { icon: '🎵', color: '#1DB954', name: 'Spotify' },
  'youtube': { icon: '📺', color: '#FF0000', name: 'YouTube' },
  'netflix': { icon: '🎬', color: '#E50914', name: 'Netflix' },
  'twitch': { icon: '🎮', color: '#9146FF', name: 'Twitch' },
  
  // 系统工具
  'terminal': { icon: '💻', color: '#000000', name: 'Terminal' },
  'powershell': { icon: '💙', color: '#012456', name: 'PowerShell' },
  'cmd': { icon: '⚫', color: '#000000', name: 'Command Prompt' },
  'bash': { icon: '🐚', color: '#4EAA25', name: 'Bash' },
  
  // 默认分类
  'unknown': { icon: '❓', color: '#666666', name: 'Unknown' },
  'system': { icon: '⚙️', color: '#666666', name: 'System' },
  'network': { icon: '🌐', color: '#007ACC', name: 'Network' },
  'api': { icon: '🔌', color: '#FF6B6B', name: 'API' },
  'web': { icon: '🌍', color: '#4CAF50', name: 'Web' },
  'mobile': { icon: '📱', color: '#2196F3', name: 'Mobile' },
  'desktop': { icon: '🖥️', color: '#9C27B0', name: 'Desktop' }
};

/**
 * 根据应用名称获取图标信息
 */
export function getAppIcon(appName: string, appCategory?: string): AppIconInfo {
  if (!appName) {
    return APP_ICONS.unknown;
  }

  const normalizedName = appName.toLowerCase().trim();
  
  // 直接匹配
  if (APP_ICONS[normalizedName]) {
    return APP_ICONS[normalizedName];
  }
  
  // 模糊匹配
  for (const [key, iconInfo] of Object.entries(APP_ICONS)) {
    if (normalizedName.includes(key) || key.includes(normalizedName)) {
      return iconInfo;
    }
  }
  
  // 根据应用分类返回默认图标
  if (appCategory) {
    const normalizedCategory = appCategory.toLowerCase();
    if (APP_ICONS[normalizedCategory]) {
      return APP_ICONS[normalizedCategory];
    }
  }
  
  // 根据名称特征推断
  if (normalizedName.includes('browser') || normalizedName.includes('chrome') || normalizedName.includes('firefox')) {
    return APP_ICONS.web;
  }
  
  if (normalizedName.includes('mobile') || normalizedName.includes('ios') || normalizedName.includes('android')) {
    return APP_ICONS.mobile;
  }
  
  if (normalizedName.includes('api') || normalizedName.includes('rest') || normalizedName.includes('graphql')) {
    return APP_ICONS.api;
  }
  
  if (normalizedName.includes('terminal') || normalizedName.includes('shell') || normalizedName.includes('cmd')) {
    return APP_ICONS.terminal;
  }
  
  // 默认返回未知图标
  return APP_ICONS.unknown;
}

/**
 * 获取所有可用的应用图标
 */
export function getAllAppIcons(): Record<string, AppIconInfo> {
  return { ...APP_ICONS };
}

/**
 * 根据User-Agent推断应用信息
 */
export function getAppFromUserAgent(userAgent: string): AppIconInfo {
  if (!userAgent) {
    return APP_ICONS.unknown;
  }
  
  const ua = userAgent.toLowerCase();
  
  // 浏览器检测
  if (ua.includes('chrome') && !ua.includes('edge')) {
    return APP_ICONS.chrome;
  }
  if (ua.includes('firefox')) {
    return APP_ICONS.firefox;
  }
  if (ua.includes('safari') && !ua.includes('chrome')) {
    return APP_ICONS.safari;
  }
  if (ua.includes('edge')) {
    return APP_ICONS.edge;
  }
  if (ua.includes('opera')) {
    return APP_ICONS.opera;
  }
  
  // 移动应用检测
  if (ua.includes('iphone') || ua.includes('ipad')) {
    return APP_ICONS.ios;
  }
  if (ua.includes('android')) {
    return APP_ICONS.android;
  }
  
  // 工具检测
  if (ua.includes('postman')) {
    return APP_ICONS.postman;
  }
  if (ua.includes('insomnia')) {
    return APP_ICONS.insomnia;
  }
  if (ua.includes('curl')) {
    return APP_ICONS.curl;
  }
  
  return APP_ICONS.web;
}
