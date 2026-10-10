import { BrandIcon } from '../../components/BrandIcon';
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

const placeholders = [{ name: '小红书', icon: 'xiaohongshu' }, { name: 'YouTube', icon: 'youtube' }, { name: 'X', icon: 'x' }, { name: '微信公众号', icon: 'wechat' }, { name: '快手', icon: 'kuaishou' }, { name: '微博', icon: 'weibo' }, { name: '知乎', icon: 'zhihu' }, { name: '微信视频号', icon: 'wechat' }, { name: 'Instagram', icon: 'instagram' }, { name: 'TikTok', icon: 'tiktok' }, { name: 'Facebook', icon: 'facebook' }, { name: 'Reddit', icon: 'reddit' }];
type Source = components['schemas']['TrackingSourceV1'];
export function AccountsPanel({ fixture, onChanged, onMaterial }: { fixture: boolean; onChanged: () => void; onMaterial: (id: string) => void }) {
  const sources = useReadApi(signal => trackingApi.listSources({ signal }), { enabled: !fixture });
  const [id, setID] = useState('');
  const [accessMode, setAccessMode] = useState<'public' | 'browser_selected'>('public');
  const [platform, setPlatform] = useState<Source['platform']>('bilibili');
  const [kind, setKind] = useState<Source['source_kind']>('uploads');
  const [externalID, setExternalID] = useState('');
  const [ownerID, setOwnerID] = useState('');
  const [locator, setLocator] = useState('');
  const [title, setTitle] = useState('');
  const [collections, setCollections] = useState<components['schemas']['SourceCollectionV1'][] | null>(null);
  const command = useCommand(fixture || sources.loading || sources.stale || !sources.data);
  const setSourceLocator = (value: string) => {
    setLocator(value); setCollections(null); setExternalID(''); setOwnerID('');
    try {
      const url = new URL(value);
      if (platform === 'bilibili' && url.hostname === 'space.bilibili.com') {
        const owner = url.pathname.split('/').filter(Boolean)[0] ?? '';
        if (/^\d+$/.test(owner)) { setOwnerID(owner); setExternalID(kind === 'favorites' ? url.searchParams.get('fid') ?? '' : owner); }
      }
      if (platform === 'douyin' && ['www.douyin.com', 'douyin.com'].includes(url.hostname)) {
        const owner = url.pathname.match(/^\/user\/([^/]+)/)?.[1];
        if (owner && owner !== 'self') { setOwnerID(owner); if (kind === 'uploads') setExternalID(owner); }
      }
    } catch { /* An incomplete URL stays editable; no source is submitted. */ }
  };
  const discover = () => {
    const request = command.prepare({ expected_version: 1, platform, owner_id: accessMode === 'browser_selected' ? 'self' : ownerID.trim() || externalID.trim(), access_mode: accessMode });
    void command.run(() => trackingApi.discoverCollections(request.body, request.key), result => setCollections(result.items), '已读取所选范围的收藏夹清单；选择后才保存绑定');
  };
  let requiresPublicURL = false;
  try { requiresPublicURL = accessMode === 'public' && platform === 'douyin' && new URL(locator).pathname.replace(/\/$/, '') === '/user/self'; } catch { /* Actual locator validated by form/backend. */ }
  return <SectionCard title="账号与更新清单" tabs={['overview']} padded>
    <div className="ac-platform-heading"><BrandIcon name="bilibili" /><span>B站</span><BrandIcon name="douyin" /><span>抖音</span></div>
    <p className={styles.note}>每个平台可绑定多个创作者或公开收藏夹。公开来源与明确选定的已登录抖音收藏分别绑定。先同步标题与简介，再取得 Agent 建议，由你勾选后用 summarize 获取正文。首次默认100条，服务启动同步一次。</p>
    {fixture ? <p className={styles.note}>示例模式不绑定账号、读取更新清单或提交导入。</p> : <>
      <button className="ac-button secondary compact" type="button" disabled={sources.loading} onClick={sources.retry}>重载绑定列表</button>
      <QueryState state={sources} empty={data => !data.items.length} emptyTitle="尚未绑定公开来源">{data => <div className={formStyles.form}><SelectField label="查看公开来源" value={id} onChange={setID} options={[{ value: '', label: '选择一个来源…' }, ...data.items.map(source => ({ value: source.id, label: `${source.platform === 'bilibili' ? 'B站' : '抖音'} · ${source.source_kind === 'uploads' ? '创作者发布' : '收藏夹'} · ${source.access_mode === 'browser_selected' ? '已选浏览器范围' : '公开'} · ${source.title || source.external_id}` }))]} />{data.next_cursor && <p>还有未加载来源，当前列表仅覆盖本页。</p>}</div>}</QueryState>
      <details><summary>绑定公开创作者 / 收藏夹</summary><form className={formStyles.form} onSubmit={event => {
        event.preventDefault(); if (requiresPublicURL) return;
        const request = command.prepare({ expected_version: 1, platform, source_kind: kind, access_mode: accessMode, external_id: externalID.trim(), owner_id: ownerID.trim(), locator: locator.trim(), title: title.trim() });
        void command.run(() => trackingApi.bindSource(request.body, request.key), result => { sources.retry(); setID(result.source.id); }, '已保存公开来源；同步只获取标题与简介，不自动入库');
      }}><fieldset disabled={fixture || command.pending || sources.loading || sources.stale || !sources.data}><legend>公开来源</legend>
        <SelectField label="来源平台" value={platform} onChange={value => { setPlatform(value as Source['platform']); setAccessMode('public'); setExternalID(''); setOwnerID(''); setLocator(''); setCollections(null); }} options={[{ value: 'bilibili', label: 'B站' }, { value: 'douyin', label: '抖音' }]} />
        <SelectField label="追踪内容" value={kind} onChange={value => { setKind(value as Source['source_kind']); setAccessMode('public'); setOwnerID(value === 'favorites' && platform === 'bilibili' ? externalID : ''); setExternalID(''); setLocator(''); setTitle(''); setCollections(null); }} options={[{ value: 'uploads', label: '创作者发布' }, { value: 'favorites', label: '公开收藏夹' }]} />
        {platform === 'douyin' && kind === 'favorites' && <SelectField label="收藏夹读取范围" value={accessMode} onChange={value => { setAccessMode(value as typeof accessMode); setExternalID(''); setOwnerID(''); setLocator(''); setCollections(null); }} options={[{ value: 'public', label: '公开收藏夹' }, { value: 'browser_selected', label: '已登录浏览器 · 仅明确选定收藏夹' }]} />}
        {accessMode === 'browser_selected' && <p className={styles.note}>先正常登录并连接 OpenCLI 浏览器扩展。服务只读取主控交接的选定收藏夹；不读取其他收藏、私信、观看历史或 Cookie。未连接、未获准或平台挑战会保留实际错误。</p>}
        {accessMode === 'public' && <TextField label="公开主页 / 收藏夹链接" value={locator} onChange={setSourceLocator} required type="url" hint="粘贴公开主页或明确收藏夹链接；程序识别来源身份。收藏夹列表需你主动读取并选择。" />}
        {kind === 'favorites' && <>
          <button className="ac-button secondary" type="button" disabled={command.pending || (accessMode === 'public' && !ownerID.trim() && !externalID.trim())} onClick={discover}>{accessMode === 'browser_selected' ? '读取已连接浏览器的选定收藏夹' : '查找此账号的公开收藏夹'}</button>
          {collections && <div className="ac-folder-choices">{collections.length ? collections.filter(collection => (collection.access_mode ?? 'public') === accessMode).map(collection => <button className="ac-button secondary" type="button" key={`${collection.owner_id}:${collection.external_id}`} onClick={() => { setAccessMode(collection.access_mode ?? 'public'); setExternalID(collection.external_id); setOwnerID(collection.owner_id); setLocator(collection.locator); setTitle(collection.title); }}><BrandIcon name={platform} size={22} />{collection.title || '未提供收藏夹名称'} · {collection.item_count == null ? '数量未知' : `${collection.item_count} 条`}</button>) : <p role="status">此范围本次没有返回收藏夹；不代表整个账号为空。请确认选定范围后重新读取。</p>}</div>}
          {accessMode === 'browser_selected' && externalID && <p role="status">已选择 · {title || '未提供名称'}；点击绑定后才保存此收藏夹。</p>}
        </>}
        <details><summary>手动填写来源身份（高级）</summary>
          <TextField label={kind === 'uploads' ? '公开账号 ID' : '公开收藏夹 ID'} value={externalID} onChange={setExternalID} hint="优先使用链接识别或列表选择。仅在已有确切身份时手动填写，不使用显示名或 self。" />
          {kind === 'favorites' && <TextField label="收藏夹所属账号 ID" value={ownerID} onChange={value => { setOwnerID(value); setCollections(null); }} />}
          {accessMode === 'browser_selected' && <TextField label="公开主页 / 收藏夹链接" value={locator} onChange={setLocator} type="url" />}
        </details>
        <TextField label="来源备注（可选）" value={title} onChange={setTitle} />
      </fieldset>{requiresPublicURL && <p role="alert">此 self 收藏地址不能作为公开来源。请提供公开主页，或切换到明确选定的浏览器收藏夹并填写稳定身份。</p>}<CommandState {...command} /><button className="ac-button" type="submit" disabled={fixture || command.pending || requiresPublicURL || sources.loading || sources.stale || !sources.data || !externalID.trim() || !locator.trim() || (kind === 'favorites' && !ownerID.trim())}>{accessMode === 'browser_selected' ? '绑定此选定浏览器收藏夹' : '绑定公开来源'}</button></form></details>
      {id && <SourceReview key={id} id={id} onChanged={() => { sources.retry(); onChanged(); }} onMaterial={onMaterial} />}
    </>}
    <details><summary>其他平台 · 待接入</summary><ul className={styles.platforms}>{placeholders.map(platformItem => <li key={platformItem.name}><strong className="ac-brand-heading"><BrandIcon name={platformItem.icon} size={24} />{platformItem.name}</strong><span>待接入</span></li>)}</ul></details>
  </SectionCard>;
}
