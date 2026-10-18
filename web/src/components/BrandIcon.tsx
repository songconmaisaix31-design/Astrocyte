// 使用本地原始素材；来源与下载失败记录见 public/brands/provenance.json。
const assets: Record<string, string> = {
  "github": "github.svg",
  "arxiv": "arxiv.png",
  "bilibili": "bilibili.ico",
  "douyin": "douyin.ico",
  "opencode": "opencode.png",
  "pi": "pi.svg",
  "gemini": "gemini.png",
  "qwen": "qwen.png",
  "amp": "amp.svg",
  "cursor": "cursor.svg",
  "claude": "claude.ico",
  "goose": "goose.png",
  "codex": "codex.png",
  "grok": "grok.svg",
  "kimi": "kimi.ico",
  "xiaohongshu": "xiaohongshu.ico",
  "youtube": "youtube.png",
  "x": "x.png",
  "wechat": "wechat.png",
  "kuaishou": "kuaishou.ico",
  "weibo": "weibo.ico",
  "zhihu": "zhihu.ico",
  "instagram": "instagram.webp",
  "facebook": "facebook.ico",
  "reddit": "reddit.png",
  "tiktok": "tiktok.png",
  "cursor-agent": "cursor.svg"
};

const aliases: Record<string, string> = {
  claude_code: 'claude', 'claude-code': 'claude', 'claude code': 'claude',
  codex_cli: 'codex', 'codex-cli': 'codex', 'codex cli': 'codex',
  gemini_cli: 'gemini', 'gemini-cli': 'gemini', 'gemini cli': 'gemini',
  qwen_code: 'qwen', 'qwen-code': 'qwen', 'qwen code': 'qwen',
  kimi_code: 'kimi', 'kimi-code': 'kimi', 'kimi code': 'kimi',
  'open-code': 'opencode', 'cursor_agent': 'cursor-agent',
  twitter: 'x', 'b站': 'bilibili', '抖音': 'douyin',
};

export function BrandIcon({ name, size = 28 }: { name: string; size?: number }) {
  const normalized = name.trim().toLowerCase();
  const asset = assets[aliases[normalized] ?? normalized];
  return asset ? <img className="ac-brand-icon" src={`/brands/${asset}`} alt="" aria-hidden="true" width={size} height={size} /> : <span className="ac-brand-unavailable" title={`${name}：官方图标暂未取得`}>图标未取得</span>;
}
