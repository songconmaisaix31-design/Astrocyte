// Evaluated by OpenCLI in the explicitly bound tab. This expression never
// reads global stores, cookies, hidden sidebars, storage, or other tabs and never
// invokes page callbacks. All returned values belong to the selected folder.
(() => {
  const request = __ASTROCYTE_SELECTED_REQUEST__;
  const fail = (code) => ({ error: { code } });
  if (location.origin !== 'https://www.douyin.com' ||
      !location.pathname.startsWith('/user/') ||
      new URLSearchParams(location.search).get('showTab') !== 'favorite_collection' ||
      new URLSearchParams(location.search).get('showSubTab') !== 'favorite_folder') {
    return fail('selected_page_required');
  }
  const fiber = (element) => element?.[Object.keys(element).find(key => key.startsWith('__reactFiber'))];
  const findProps = (element, predicate, maxDepth = 18) => {
    let current = fiber(element);
    for (let depth = 0; current && depth < maxDepth; depth++, current = current.return) {
      const props = current.memoizedProps;
      if (props && predicate(props)) return props;
    }
    return null;
  };
  const userProps = findProps(document.querySelector('[data-e2e="user-info"]'),
    props => typeof props.userInfo?.secUid === 'string', 4);
  if (!userProps || userProps.userInfo.secUid !== request.owner) return fail('selected_account_changed');
  const navigation = document.querySelector('#collection-navigation');
  if (!navigation) return fail('selected_folders_missing');
  const folders = new Map();
  for (const node of navigation.querySelectorAll('[data-tip]')) {
    const props = findProps(node, p => typeof p.id === 'string' && typeof p.name === 'string' && 'active' in p, 5);
    if (props && request.folders.includes(props.id)) {
      if (!Number.isSafeInteger(props.total) || props.total < 0) return fail('selected_folder_count_missing');
      if (folders.has(props.id)) return fail('selected_folder_duplicated');
      folders.set(props.id, { id: props.id, title: props.name, count: props.total, active: props.active === true });
    }
  }
  if (folders.size !== request.folders.length) return fail('selected_folders_missing');
  if (request.operation === 'catalog') {
    return { owner_id: request.owner, folders: request.folders.map(id => {
      const { active, ...metadata } = folders.get(id);
      return metadata;
    }) };
  }
  const selected = folders.get(request.folder);
  if (!selected?.active) return fail('selected_folder_not_open');
  const detail = document.querySelector('[data-e2e="user-detail"]');
  if (!detail) return fail('selected_page_required');
  let matched = null;
  for (const list of detail.querySelectorAll('[data-e2e="scroll-list"]')) {
    const props = findProps(list, p => p.collectionFolderInfo?.collectionFolderId === request.folder && p.defaultRes);
    if (props) {
      if (matched) return fail('selected_folder_ambiguous');
      matched = props;
    }
  }
  if (!matched) return fail('selected_folder_metadata_missing');
  const result = matched.defaultRes;
  if (result.statusCode !== 0 || !Array.isArray(result.data) ||
      ![0, 1, false, true].includes(result.hasMore)) return fail('selected_folder_metadata_missing');
  // The mounted component is evidence for a complete list only if the provider
  // reports completion and its count agrees. Never invent a next network page.
  if (result.hasMore || result.data.length !== selected.count) return fail('selected_folder_partial');
  const start = request.cursor === '' ? 0 : Number(request.cursor);
  if (!Number.isSafeInteger(start) || start < 0 || start > result.data.length) return fail('selected_cursor_invalid');
  const end = Math.min(start + request.limit, result.data.length);
  const videos = [];
  for (const item of result.data.slice(start, end)) {
    if (typeof item.awemeId !== 'string' || !/^[1-9][0-9]{0,19}$/.test(item.awemeId)) return fail('selected_video_identity_missing');
    const description = typeof item.desc === 'string' ? item.desc : '';
    const title = typeof item.itemTitle === 'string' && item.itemTitle ? item.itemTitle : description;
    const author = item.authorInfo;
    const cover = item.video?.cover;
    videos.push({ id: item.awemeId, title, description,
      author: typeof author?.nickname === 'string' ? author.nickname : '',
      cover: typeof cover === 'string' && cover.startsWith('https://') ? cover : '',
      published_at: Number.isSafeInteger(item.createTime) && item.createTime >= 0 ? item.createTime : 0 });
  }
  return { owner_id: request.owner, folder_id: request.folder, status_code: result.statusCode,
    has_more: end < result.data.length, cursor: String(end), videos };
})()
