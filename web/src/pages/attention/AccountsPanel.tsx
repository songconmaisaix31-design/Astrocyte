import { CollectionOverview } from '../../components/CollectionOverview';
import { SectionCard } from '../../components/SectionCard';
import styles from './AccountsPanel.module.css';

const platforms = ['B站', '抖音', '小红书', 'YouTube', 'X', '微信公众号'];

/** This initial panel has no connection state: binding and list contracts are still pending. */
export function AccountsPanel({ fixture }: { fixture: boolean }) {
  return <SectionCard title="账号与更新清单" tabs={['overview']} padded>
    <CollectionOverview title="把关注留给你选择" description="账号更新形成待选清单，由你决定哪些内容进入资料库。" metrics={[
      { label: '已绑定账号', value: '未接入' }, { label: '待选更新', value: '未获取' },
    ]}>
      <ul className={styles.platforms}>{platforms.map((platform, index) => <li key={platform}><strong>{platform}</strong><span>{index < 2 ? '绑定方式待确定' : '待接入'}</span></li>)}</ul>
      <p className={styles.note}>{fixture ? '示例模式不绑定账号或提交导入。' : '公开来源或浏览器登录的绑定方式尚未确定。账号列表未接入，当前没有可勾选更新。'}</p>
      <p className={styles.note}>入库由人工选择；Agent 反馈及正文获取的先后顺序待确定。</p>
    </CollectionOverview>
  </SectionCard>;
}
