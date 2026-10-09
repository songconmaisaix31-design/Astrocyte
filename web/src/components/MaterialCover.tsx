import { Icon } from './DesignIcons';
import { materialKindLabel } from '../utils/format';

export function MaterialCover({ kind }: { kind: 'paper' | 'video' | 'text' | 'file' }) {
  const art = kind === 'paper' ? 'paper' : kind === 'video' ? 'video' : 'note';
  return <div className={`ac-cover ${art}`} aria-hidden="true">
    <span className="ac-cover-label">{materialKindLabel(kind)}</span>
    {art === 'paper' ? <div className="ac-paper-art"><span>RESEARCH NOTES</span><b>Ideas<br />into<br />Knowledge</b><div className="ac-formula">∞ · ∑ · Δ</div><i /><i /></div> : art === 'video' ? <div className="ac-video-art"><div className="ac-wave">{[14, 30, 47, 22, 63, 37, 52, 27, 16].map((height, index) => <i key={index} style={{ height }} />)}</div><span className="ac-play"><Icon name="video" size={24} /></span></div> : <div className="ac-note-art"><Icon name="file" size={28} /><span>记录一个<br />值得继续的问题</span></div>}
  </div>;
}
