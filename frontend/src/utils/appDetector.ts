// 应用检测工具

export interface AppInfo {
  name: string;
  icon: string;
  category: string;
}

// 应用检测规则
const APP_RULES: Array<{
  domains: string[];
  userAgents?: string[];
  headers?: Record<string, string>;
  app: AppInfo;
}> = [
  // 浏览器 - Chrome
  {
    domains: ['google.com', 'googleapis.com', 'gstatic.com', 'googleusercontent.com'],
    userAgents: ['Chrome'],
    app: { name: 'Chrome', icon: '🌐', category: 'Browser' }
  },
  // 浏览器 - Firefox
  {
    domains: ['mozilla.org', 'firefox.com'],
    userAgents: ['Firefox'],
    app: { name: 'Firefox', icon: '🦊', category: 'Browser' }
  },
  // 浏览器 - Safari
  {
    domains: ['apple.com', 'icloud.com'],
    userAgents: ['Safari', 'Version/'],
    app: { name: 'Safari', icon: '🧭', category: 'Browser' }
  },
  // 浏览器 - Edge
  {
    domains: ['microsoft.com', 'bing.com'],
    userAgents: ['Edge', 'Edg/'],
    app: { name: 'Microsoft Edge', icon: '🌐', category: 'Browser' }
  },
  // 通用浏览器
  {
    domains: ['google.com', 'bing.com', 'baidu.com', 'duckduckgo.com'],
    userAgents: ['Mozilla'],
    app: { name: 'Web Browser', icon: '🌐', category: 'Browser' }
  },
  
  // 社交媒体
  {
    domains: ['twitter.com', 'x.com', 't.co'],
    app: { name: 'Twitter/X', icon: '🐦', category: 'Social' }
  },
  {
    domains: ['facebook.com', 'fb.com', 'instagram.com'],
    app: { name: 'Meta Apps', icon: '📘', category: 'Social' }
  },
  {
    domains: ['linkedin.com'],
    app: { name: 'LinkedIn', icon: '💼', category: 'Social' }
  },
  {
    domains: ['tiktok.com', 'bytedance.com'],
    app: { name: 'TikTok', icon: '🎵', category: 'Social' }
  },
  {
    domains: ['weibo.com', 'sina.com.cn'],
    app: { name: 'Weibo', icon: '🐦', category: 'Social' }
  },
  
  // 音乐
  {
    domains: ['music.163.com', 'netease.com'],
    app: { name: 'NetEase Music', icon: '🎵', category: 'Music' }
  },
  {
    domains: ['spotify.com', 'scdn.co'],
    app: { name: 'Spotify', icon: '🎵', category: 'Music' }
  },
  {
    domains: ['music.apple.com', 'itunes.apple.com'],
    app: { name: 'Apple Music', icon: '🎵', category: 'Music' }
  },
  {
    domains: ['music.youtube.com'],
    app: { name: 'YouTube Music', icon: '🎵', category: 'Music' }
  },
  
  // 视频
  {
    domains: ['youtube.com', 'youtu.be', 'googlevideo.com'],
    app: { name: 'YouTube', icon: '📺', category: 'Video' }
  },
  {
    domains: ['netflix.com', 'nflxvideo.net'],
    app: { name: 'Netflix', icon: '📺', category: 'Video' }
  },
  {
    domains: ['bilibili.com', 'bilivideo.com'],
    app: { name: 'Bilibili', icon: '📺', category: 'Video' }
  },
  {
    domains: ['twitch.tv', 'ttvnw.net'],
    app: { name: 'Twitch', icon: '📺', category: 'Video' }
  },
  
  // 办公
  {
    domains: ['office.com', 'outlook.com', 'sharepoint.com', 'onedrive.com'],
    app: { name: 'Microsoft Office', icon: '📄', category: 'Office' }
  },
  {
    domains: ['google.com', 'googleapis.com', 'googleusercontent.com'],
    userAgents: ['Google'],
    app: { name: 'Google Workspace', icon: '📄', category: 'Office' }
  },
  {
    domains: ['slack.com', 'slack-edge.com'],
    app: { name: 'Slack', icon: '💬', category: 'Office' }
  },
  {
    domains: ['zoom.us', 'zoom.com'],
    app: { name: 'Zoom', icon: '📹', category: 'Office' }
  },
  {
    domains: ['teams.microsoft.com'],
    app: { name: 'Microsoft Teams', icon: '💬', category: 'Office' }
  },
  
  // 开发工具
  {
    domains: ['github.com', 'githubusercontent.com'],
    app: { name: 'GitHub', icon: '🐙', category: 'Development' }
  },
  {
    domains: ['gitlab.com'],
    app: { name: 'GitLab', icon: '🦊', category: 'Development' }
  },
  {
    domains: ['stackoverflow.com', 'stackexchange.com'],
    app: { name: 'Stack Overflow', icon: '📚', category: 'Development' }
  },
  {
    domains: ['npmjs.com', 'npm.im'],
    app: { name: 'NPM', icon: '📦', category: 'Development' }
  },
  
  // 游戏
  {
    domains: ['steam.com', 'steamcommunity.com', 'steamstatic.com'],
    app: { name: 'Steam', icon: '🎮', category: 'Gaming' }
  },
  {
    domains: ['epicgames.com', 'unrealengine.com'],
    app: { name: 'Epic Games', icon: '🎮', category: 'Gaming' }
  },
  {
    domains: ['battle.net', 'blizzard.com'],
    app: { name: 'Battle.net', icon: '🎮', category: 'Gaming' }
  },
  
  // 购物
  {
    domains: ['amazon.com', 'amazon.cn', 'amazonaws.com'],
    app: { name: 'Amazon', icon: '🛒', category: 'Shopping' }
  },
  {
    domains: ['taobao.com', 'tmall.com', 'alibaba.com'],
    app: { name: 'Alibaba', icon: '🛒', category: 'Shopping' }
  },
  {
    domains: ['jd.com', '360buyimg.com'],
    app: { name: 'JD.com', icon: '🛒', category: 'Shopping' }
  },
  
  // 新闻
  {
    domains: ['reddit.com', 'redd.it'],
    app: { name: 'Reddit', icon: '📰', category: 'News' }
  },
  {
    domains: ['news.ycombinator.com'],
    app: { name: 'Hacker News', icon: '📰', category: 'News' }
  },
  
  // 云服务
  {
    domains: ['icloud.com', 'apple.com'],
    app: { name: 'iCloud', icon: '☁️', category: 'Cloud' }
  },
  {
    domains: ['dropbox.com', 'dropboxapi.com'],
    app: { name: 'Dropbox', icon: '☁️', category: 'Cloud' }
  },
  
  // 通讯 - 微信
  {
    domains: ['weixin.qq.com', 'wx.qq.com', 'wechat.com', 'qpic.cn', 'gtimg.cn'],
    userAgents: ['MicroMessenger', 'WeChat'],
    app: { name: '微信', icon: '💬', category: 'Communication' }
  },
  // 通讯 - QQ
  {
    domains: ['qq.com', 'qzone.qq.com', 'qlogo.cn', 'gtimg.cn'],
    userAgents: ['QQ/', 'TencentTraveler'],
    app: { name: 'QQ', icon: '🐧', category: 'Communication' }
  },
  // 通讯 - 钉钉
  {
    domains: ['dingtalk.com', 'alicdn.com'],
    userAgents: ['DingTalk'],
    app: { name: '钉钉', icon: '💼', category: 'Communication' }
  },
  // 通讯 - 飞书
  {
    domains: ['feishu.cn', 'larksuite.com', 'bytedance.com'],
    userAgents: ['Lark', 'Feishu'],
    app: { name: '飞书', icon: '🚀', category: 'Communication' }
  },
  {
    domains: ['whatsapp.com', 'whatsapp.net'],
    app: { name: 'WhatsApp', icon: '💬', category: 'Communication' }
  },
  {
    domains: ['telegram.org', 'telegram.me'],
    app: { name: 'Telegram', icon: '💬', category: 'Communication' }
  },
  {
    domains: ['discord.com', 'discordapp.com'],
    app: { name: 'Discord', icon: '💬', category: 'Communication' }
  },
  
  // 移动应用
  {
    domains: ['douyin.com', 'bytedance.com', 'amemv.com'],
    userAgents: ['aweme', 'Aweme'],
    app: { name: '抖音', icon: '🎵', category: 'Social' }
  },
  {
    domains: ['xiaohongshu.com', 'xhscdn.com'],
    userAgents: ['XiaoHongShu'],
    app: { name: '小红书', icon: '📖', category: 'Social' }
  },
  {
    domains: ['zhihu.com', 'zhimg.com'],
    userAgents: ['ZhihuHybrid'],
    app: { name: '知乎', icon: '🤔', category: 'Social' }
  },
  {
    domains: ['taobao.com', 'tmall.com'],
    userAgents: ['AliApp'],
    app: { name: '淘宝/天猫', icon: '🛒', category: 'Shopping' }
  },
  {
    domains: ['meituan.com', 'sankuai.com'],
    userAgents: ['MeiTuan'],
    app: { name: '美团', icon: '🍔', category: 'Shopping' }
  },
  {
    domains: ['ele.me', 'eleme.cn'],
    userAgents: ['Eleme'],
    app: { name: '饿了么', icon: '🍜', category: 'Shopping' }
  },

  // 桌面应用
  {
    domains: ['vscode.dev', 'github.com'],
    userAgents: ['VSCode'],
    app: { name: 'VS Code', icon: '💻', category: 'Development' }
  },
  {
    domains: ['jetbrains.com'],
    userAgents: ['IntelliJ', 'WebStorm', 'PyCharm'],
    app: { name: 'JetBrains IDE', icon: '💻', category: 'Development' }
  },
  {
    domains: ['postman.com', 'getpostman.com'],
    userAgents: ['PostmanRuntime'],
    app: { name: 'Postman', icon: '📮', category: 'Development' }
  },

  // 系统服务
  {
    domains: ['apple.com', 'icloud.com', 'mzstatic.com'],
    userAgents: ['Darwin', 'CFNetwork'],
    app: { name: 'macOS System', icon: '🍎', category: 'System' }
  },
  {
    domains: ['microsoft.com', 'windows.com', 'msftconnecttest.com'],
    userAgents: ['Windows'],
    app: { name: 'Windows System', icon: '🪟', category: 'System' }
  },
  {
    domains: ['android.com', 'googleapis.com'],
    userAgents: ['Android'],
    app: { name: 'Android System', icon: '🤖', category: 'System' }
  }
];

export function detectApp(domain: string, userAgent?: string, headers?: Record<string, string>): AppInfo {
  // 标准化域名
  const normalizedDomain = domain.toLowerCase().replace(/^www\./, '');
  const normalizedUserAgent = userAgent?.toLowerCase() || '';

  // 优先通过User-Agent识别应用
  if (userAgent) {
    // 微信相关
    if (normalizedUserAgent.includes('micromessenger')) {
      return { name: '微信', icon: '💬', category: 'Communication' };
    }
    if (normalizedUserAgent.includes('qq/')) {
      return { name: 'QQ', icon: '🐧', category: 'Communication' };
    }
    if (normalizedUserAgent.includes('dingtalk')) {
      return { name: '钉钉', icon: '💼', category: 'Communication' };
    }

    // 浏览器识别
    if (normalizedUserAgent.includes('chrome') && !normalizedUserAgent.includes('edge')) {
      return { name: 'Chrome', icon: '🌐', category: 'Browser' };
    }
    if (normalizedUserAgent.includes('firefox')) {
      return { name: 'Firefox', icon: '🦊', category: 'Browser' };
    }
    if (normalizedUserAgent.includes('safari') && !normalizedUserAgent.includes('chrome')) {
      return { name: 'Safari', icon: '🧭', category: 'Browser' };
    }
    if (normalizedUserAgent.includes('edge') || normalizedUserAgent.includes('edg/')) {
      return { name: 'Microsoft Edge', icon: '🌐', category: 'Browser' };
    }

    // 移动应用
    if (normalizedUserAgent.includes('aweme')) {
      return { name: '抖音', icon: '🎵', category: 'Social' };
    }
    if (normalizedUserAgent.includes('xiaohongshu')) {
      return { name: '小红书', icon: '📖', category: 'Social' };
    }
    if (normalizedUserAgent.includes('zhihuhybrid')) {
      return { name: '知乎', icon: '🤔', category: 'Social' };
    }

    // 开发工具
    if (normalizedUserAgent.includes('postmanruntime')) {
      return { name: 'Postman', icon: '📮', category: 'Development' };
    }
    if (normalizedUserAgent.includes('vscode')) {
      return { name: 'VS Code', icon: '💻', category: 'Development' };
    }
  }

  // 查找匹配的规则
  for (const rule of APP_RULES) {
    // 检查域名匹配
    const domainMatch = rule.domains.some(ruleDomain =>
      normalizedDomain.includes(ruleDomain) || ruleDomain.includes(normalizedDomain)
    );

    if (domainMatch) {
      // 如果有User-Agent要求，检查是否匹配
      if (rule.userAgents && userAgent) {
        const userAgentMatch = rule.userAgents.some(ua =>
          normalizedUserAgent.includes(ua.toLowerCase())
        );
        if (!userAgentMatch) continue;
      }

      // 如果有Header要求，检查是否匹配
      if (rule.headers && headers) {
        const headerMatch = Object.entries(rule.headers).every(([key, value]) =>
          headers[key]?.toLowerCase().includes(value.toLowerCase())
        );
        if (!headerMatch) continue;
      }

      return rule.app;
    }
  }

  // 基于域名的智能识别
  if (normalizedDomain.includes('cdn') || normalizedDomain.includes('static') ||
      normalizedDomain.includes('assets') || normalizedDomain.includes('img')) {
    return { name: 'CDN/Static', icon: '📦', category: 'Infrastructure' };
  }

  if (normalizedDomain.includes('api') || normalizedDomain.includes('service')) {
    return { name: 'API Service', icon: '🔌', category: 'API' };
  }

  if (normalizedDomain.includes('analytics') || normalizedDomain.includes('tracking') ||
      normalizedDomain.includes('metrics') || normalizedDomain.includes('stats')) {
    return { name: 'Analytics', icon: '📊', category: 'Analytics' };
  }

  // 未知应用
  return { name: 'Unknown App', icon: '❓', category: 'Unknown' };
}

export function getAppCategories(): string[] {
  const categories = new Set<string>();
  APP_RULES.forEach(rule => categories.add(rule.app.category));
  categories.add('Infrastructure');
  categories.add('API');
  categories.add('Analytics');
  categories.add('Unknown');
  return Array.from(categories).sort();
}
