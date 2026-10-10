import { Icon } from './DesignIcons';

// Original assets served locally; sources and failed downloads are public/brands/provenance.json.
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
  "kimi": "kimi.ico",
  "cursor-agent": "cursor.svg"
};

export function BrandIcon({ name, size = 28 }: { name: string; size?: number }) {
  const asset = assets[name.toLowerCase()];
  return asset ? <img className="ac-brand-icon" src={`/brands/${asset}`} alt="" aria-hidden="true" width={size} height={size} /> : <span className="ac-brand-icon ac-brand-unavailable" title="官方图标暂未取得"><Icon name="layers" size={size} /></span>;
}
