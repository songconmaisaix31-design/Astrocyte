import { useState } from 'react';
import { trackingApi } from '../../api/s1';
import type { components } from '../../api/schema';
import { useReadApi } from '../../hooks/useReadApi';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SectionCard } from '../../components/SectionCard';
import { SelectField, TextField } from './FormControls';
import { SourceReview } from './SourceReview';
import styles from './AccountsPanel.module.css';
import formStyles from './AttentionPage.module.css';

const placeholders = ['小红书', 'YouTube', 'X', '微信公众号', '快手', '微博', '知乎', '微信视频号', 'Instagram', 'TikTok', 'Facebook', 'Reddit'];
type Source = components['schemas']['TrackingSourceV1'];
export function AccountsPanel({ fixture, onChanged, onMaterial }: { fixture: boolean; onChanged: () => void; onMaterial: (id: string) => void }) {
  const sources = useReadApi(signal => trackingApi.listSources({ signal }), { enabled: !fixture });
  const [id, setID] = useState('');
  const [platform, setPlatform] = useState<Source['platform']>('bilibili');
  const [kind, setKind] = useState<Source['source_kind']>('uploads');
  const [externalID, setExternalID] = useState('3494358764489275');
  const [ownerID, setOwnerID] = useState('');
  const [locator, setLocator] = useState('https://space.bilibili.com/3494358764489275');
  const [title, setTitle] = useState('');
  const [collections, setCollections] = useState<components['schemas']['SourceCollectionV1'][] | null>(null);
  const command = useCommand(fixture || sources.loading || sources.stale || !sources.data);
  let requiresPublicURL = false;
  try { requiresPublicURL = platform === 'douyin' && new URL(locator).pathname.replace(/\/$/, '') === '/user/self'; } catch { /* Actual locator validated by form/backend. */ }
  return <SectionCard title="账号与更新清单" tabs={['overview']} padded>
    <p className={styles.note}>每个平台可绑定多个创作者或公开收藏夹。先同步标题与简介，再取得 Agent 建议，由你勾选后用 summarize 获取正文。首次默认100条，服务启动同步一次。</p>
    {fixture ? <p className={styles.note}>示例模式不绑定账号、读取更新清单或提交导入。</p> : <>
      <button className="ac-button secondary compact" type="button" disabled={sources.loading} onClick={sources.retry}>重载绑定列表</button>
      <QueryState state={sources} empty={data => !data.items.length} emptyTitle="尚未绑定公开来源">{data => <div className={formStyles.form}><SelectField label="查看公开来源" value={id} onChange={setID} options={[{ value: '', label: '选择一个来源…' }, ...data.items.map(source => ({ value: source.id, label: `${source.platform === 'bilibili' ? 'B站' : '抖音'} · ${source.source_kind === 'uploads' ? '创作者发布' : '公开收藏夹'} · ${source.title || source.external_id}` }))]} />{data.next_cursor && <p>还有未加载来源，当前列表仅覆盖本页。</p>}</div>}</QueryState>
      <details><summary>绑定公开创作者 / 收藏夹</summary><form className={formStyles.form} onSubmit={event => {
        event.preventDefault(); if (requiresPublicURL) return;
        const request = command.prepare({ expected_version: 1, platform, source_kind: kind, external_id: externalID.trim(), owner_id: ownerID.trim(), locator: locator.trim(), title: title.trim() });
        void command.run(() => trackingApi.bindSource(request.body, request.key), result => { sources.retry(); setID(result.source.id); }, '已保存公开来源；同步只获取标题与简介，不自动入库');
      }}><fieldset disabled={fixture || command.pending || sources.loading || sources.stale || !sources.data}><legend>公开来源</legend>
        <SelectField label="来源平台" value={platform} onChange={value => { setPlatform(value as Source['platform']); setExternalID(''); setOwnerID(''); setLocator(''); setCollections(null); }} options={[{ value: 'bilibili', label: 'B站' }, { value: 'douyin', label: '抖音' }]} />
        <SelectField label="追踪内容" value={kind} onChange={value => { setKind(value as Source['source_kind']); setOwnerID(value === 'favorites' && platform === 'bilibili' ? externalID : ''); setExternalID(''); setLocator(''); setTitle(''); setCollections(null); }} options={[{ value: 'uploads', label: '创作者发布' }, { value: 'favorites', label: '公开收藏夹' }]} />
        <TextField label={kind === 'uploads' ? '公开账号 ID' : '公开收藏夹 ID'} value={externalID} onChange={setExternalID} required hint={platform === 'bilibili' ? '创作者填写 UID；收藏夹填写公开列表的 ID。' : '填写可公开访问的账号或收藏夹身份，不使用登录态的 self 地址。'} />
        {kind === 'favorites' && <TextField label="收藏夹所属账号 ID" value={ownerID} onChange={value => { setOwnerID(value); setCollections(null); }} hint="填写公开页面显示的所属账号，不读取登录信息。" />}
        {kind === 'favorites' && <><button className="ac-button secondary compact" type="button" disabled={!ownerID.trim()} onClick={() => { void command.run(() => trackingApi.listCollections(platform, ownerID.trim()), result => setCollections(result.items), '已获取公开收藏夹列表；请选择一个明确绑定'); }}>查找此账号的公开收藏夹</button>{collections && <div>{collections.length ? collections.map(collection => <button className="ac-button secondary compact" type="button" key={collection.external_id} onClick={() => { setExternalID(collection.external_id); setOwnerID(collection.owner_id); setLocator(collection.locator); setTitle(collection.title); }}>{collection.title || collection.external_id} · 内容数量 {collection.item_count ?? '未提供'}</button>) : <p>该账号本次没有返回公开收藏夹；请提供另一个公开账号或明确收藏夹链接。</p>}</div>}</>}
        <TextField label="公开主页 / 收藏夹链接" value={locator} onChange={setLocator} required type="url" />
        <TextField label="来源备注（可选）" value={title} onChange={setTitle} />
      </fieldset>{requiresPublicURL && <p role="alert">此 self 收藏地址需要登录。请提供可公开访问的抖音主页或公开收藏夹链接；本轮只处理公开来源。</p>}<CommandState {...command} /><button className="ac-button" type="submit" disabled={fixture || command.pending || requiresPublicURL || sources.loading || sources.stale || !sources.data}>绑定公开来源</button></form></details>
      {id && <SourceReview key={id} id={id} onChanged={() => { sources.retry(); onChanged(); }} onMaterial={onMaterial} />}
    </>}
    <details><summary>其他平台 · 待接入</summary><ul className={styles.platforms}>{placeholders.map(platformName => <li key={platformName}><strong>{platformName}</strong><span>待接入</span></li>)}</ul></details>
  </SectionCard>;
}
