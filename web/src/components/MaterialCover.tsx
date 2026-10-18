import { Icon } from './DesignIcons';
import { materialKindLabel } from '../utils/format';
import { BrandIcon } from './BrandIcon';

export function MaterialCover({ kind, sourceLocator }: { kind: 'paper' | 'video' | 'text' | 'file'; sourceLocator?: string }) {
  const url = sourceLocator && URL.canParse(sourceLocator) ? new URL(sourceLocator) : null;
  const host = url?.hostname.toLowerCase();
  const brands = [
    { domain: 'arxiv.org', name: 'arxiv' }, { domain: 'bilibili.com', name: 'bilibili' }, { domain: 'b23.tv', name: 'bilibili' },
    { domain: 'douyin.com', name: 'douyin' }, { domain: 'youtube.com', name: 'youtube' }, { domain: 'youtu.be', name: 'youtube' },
    { domain: 'github.com', name: 'github' }, { domain: 'xiaohongshu.com', name: 'xiaohongshu' }, { domain: 'xhslink.com', name: 'xiaohongshu' },
    { domain: 'x.com', name: 'x' }, { domain: 'twitter.com', name: 'x' }, { domain: 'weixin.qq.com', name: 'wechat' },
    { domain: 'kuaishou.com', name: 'kuaishou' }, { domain: 'weibo.com', name: 'weibo' }, { domain: 'weibo.cn', name: 'weibo' },
    { domain: 'zhihu.com', name: 'zhihu' }, { domain: 'instagram.com', name: 'instagram' }, { domain: 'facebook.com', name: 'facebook' },
    { domain: 'reddit.com', name: 'reddit' }, { domain: 'tiktok.com', name: 'tiktok' },
  ];
  const brand = brands.find(item => host === item.domain || host?.endsWith(`.${item.domain}`));
  if (brand) return <div className={`ac-cover ac-source-cover ${kind}`} aria-hidden="true"><BrandIcon name={brand.name} size={32} /><span className="ac-cover-label">{materialKindLabel(kind)}</span></div>;
  const art = kind === 'paper' ? 'paper' : kind === 'video' ? 'video' : 'note';
  return <div className={`ac-cover ${art}`} aria-hidden="true">
    <span className="ac-cover-label">{materialKindLabel(kind)}</span>
    {art === 'paper' ? <div className="ac-paper-art"><span>RESEARCH NOTES</span><b>Ideas<br />into<br />Knowledge</b><div className="ac-formula">∞ · ∑ · Δ</div><i /><i /></div> : art === 'video' ? <div className="ac-video-art"><div className="ac-wave">{[14, 30, 47, 22, 63, 37, 52, 27, 16].map((height, index) => <i key={index} style={{ height }} />)}</div><span className="ac-play"><Icon name="video" size={24} /></span></div> : <div className="ac-note-art"><Icon name="file" size={28} /><span>记录一个<br />值得继续的问题</span></div>}
  </div>;
}
