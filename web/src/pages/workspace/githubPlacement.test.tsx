import { expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { PlacementForm } from './GitHubRepositoriesPanel';
import type { components } from '../../api/schema';
import type { ReadApiState } from '../../hooks/useReadApi';

type S = components['schemas'];
// 这些参数只验证按钮、空间锁定及未知状态显示，不表示实际克隆观察。
const repository: S['GitHubRepositoryV1'] = { id: 'input-repository', revision: 5, metadata_revision: 1, metadata: { github_id: 0, full_name: 'input/repository', html_url: 'input-url', clone_url: 'input-clone-url', description: '', default_branch: '', language: '', stars: 0, archived: false, fork: false, updated_at: '', pushed_at: '', metadata_source: 'github_public_html', unknown_fields: ['stars'] }, sync_status: 'synced', synced_at: '2026-01-01T00:00:00Z', space_id: 'input-space', project_id: '', clone_status: 'cloning', root: '', head: '', clone_attempts: 1, last_error: '' };
const never = () => { throw new Error('显示契约不得执行命令'); };
const spaces: ReadApiState<S['ProjectSpaceListV1']> = { data: { schema_version: 1, items: [{ id: 'input-space', title: '输入空间', version: 1, material_refs: [], created_at: '2026-01-01T00:00:00Z' }], next_cursor: null }, loading: false, stale: false, error: null, retry: never };

it('cloning 输入保留原空间，恢复命令只允许人工操作', () => {
  const html = renderToStaticMarkup(<PlacementForm repository={repository} spaces={spaces} disabled={false} onPlaced={never} />);
  expect(html).toContain('value="input-space" selected=""');
  expect(html).toContain('disabled=""');
  expect(html).toContain('<button class="ac-button" type="submit">核对并恢复此空间克隆</button>');
});
it('unknown 输入禁用提交并要求核对原操作', () => {
  const html = renderToStaticMarkup(<PlacementForm repository={{ ...repository, clone_status: 'unknown' }} spaces={spaces} disabled={false} onPlaced={never} />);
  expect(html).toContain('<button class="ac-button" type="submit" disabled="">结果未知，请先核对原操作</button>');
});
it('ready 输入禁用重复纳入', () => {
  const html = renderToStaticMarkup(<PlacementForm repository={{ ...repository, clone_status: 'ready' }} spaces={spaces} disabled={false} onPlaced={never} />);
  expect(html).toContain('<button class="ac-button" type="submit" disabled="">已纳入并克隆</button>');
});
