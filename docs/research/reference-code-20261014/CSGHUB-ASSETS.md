# CSGHub 资产、权限与版本代码索引

来源、SHA、许可证和 REA 覆盖见 [SOURCE.md](SOURCE.md)。以下服务端代码统一来自 `a67b3024a9553aecf3379c9f6bccfcdada2034f5`，前端代码来自 `bbb9bf4b346a199a387090de76fcd6c1fffe98f6`。结论均为实际源码阅读或注明的 REA 静态证据，未验证上游部署行为。

## 资产接口与实现入口

| 接口或需求 | 固定源码入口 | 可确认的行为 |
| --- | --- | --- |
| 模型资产 Git 分支、提交、目录和文件 | [api/router/api.go:679](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/router/api.go#L679) | `/api/v1/models/:namespace/:name` 下提供 `branches`、`tags`、`last_commit`、`commit/:commit_id`、`commit/:revision`、`refs/:ref/tree/*path`、`commits`、`raw/*file_path`。路由对应真实 repo handler；登录和手机号中间件随操作不同。 |
| 数据集与代码资产 | [api/router/api.go:813](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/router/api.go#L813)、[api/router/api.go:902](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/router/api.go#L902) | 数据集和代码具有相应 Git 目录/提交/文件操作。数据集部分读接口额外要求登录，不能把全部资产接口视为匿名可读。 |
| 上传与镜像操作 | [api/router/api.go:713](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/api/router/api.go#L713) | 模型 `upload_file`；`mirror` 的 POST/GET/PUT/DELETE，以及 POST `mirror/sync`。数据集和代码有对应入口。 |
| 镜像导入 | [component/mirror_repo.go:123](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/mirror_repo.go#L123) | `CreateMirrorRepo` 要求 current user，规范化 source URL、目标 namespace/name 和 priority；已有目标仓库检查 `RepositoryCanWrite`；同一来源的已有 mirror 可以重新加入同步任务。 |
| 导入记录事务 | [builder/store/database/mirror_repo.go:91](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/builder/store/database/mirror_repo.go#L91) | `RunInTx` 创建或更新 repository、typed repo、mirror、mirror task 与 current task ID。配置 jobClient 时，同事务插入对应任务。关系授权写入由 [component/mirror_repo.go:601](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/mirror_repo.go#L601) 在该调用之后执行；外部 Git、授权与数据库并非一个事务。 |
| 创建、编辑、删除、fork 的权限 | [component/repo.go:255](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L255)、[component/repo.go:380](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L380)、[component/repo.go:538](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L538)、[component/repo.go:579](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L579) | 创建检查 namespace write；编辑检查 repository write；删除检查 repository admin；fork 检查来源 read 和目标 namespace write。 |

## 权限执行与版本语义

[CheckUserRepoPermission，component/repo.go:2334](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L2334)只接受 read/write/admin 权限；匿名请求仅公开仓库 read 可通过。登录用户先查询真实用户，平台管理员可以通过，其余请求调用 ReBAC `Check`，使用 `ConsistencyHigher`。公开仓库允许 read；私有仓库 read 拒绝时清理用户访问缓存。这证明应用层存在资源权限检查；尚未运行授权后端，也不能据此断言缓存失效后的实时撤销延迟。

前端 [RepoDetailStore.js:1](https://github.com/OpenCSGs/csghub/blob/bbb9bf4b346a199a387090de76fcd6c1fffe98f6/frontend/src/stores/RepoDetailStore.js#L1)保存 `canManage`、`canWrite`、`private`、`defaultBranch`、license 与镜像状态，适合作为显示状态参考。业务权限仍须在服务端执行。REA 将该文件纳入静态解析，不证明真实接口权限行为。

[DownloadRepoZip，component/repo.go:3795](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L3795)先检查 read 权限，使用请求 revision 或 default branch；成功解析分支 commit 后，以精确 commit 请求 Git archive，并按 commit 查询/写入包缓存。[resolveRepositoryArchiveCommit:3862](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/repo.go#L3862)读取 Git branch 对象的 `Commit.ID`。解析失败时该上游实现记录警告，继续按原 revision 下载。Astrocyte 如要求固定版本，应在解析失败处返回错误，不能复制这一继续执行路径。

[resolveModelVersions，component/evaluation.go:314](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/evaluation.go#L314)解析模型版本；[resolveRepoRevision:372](https://github.com/OpenCSGs/csghub-server/blob/a67b3024a9553aecf3379c9f6bccfcdada2034f5/component/evaluation.go#L372)获取实际最后 commit，空结果直接失败；模型按 repo ID 和 SHA 区分。数据集的 path/revision/commit 也写入评估请求。此模式能为后续结果关联提供固定输入版本。

## 前端 API 和 REA 调用证据

[useFetchApi.js:44](https://github.com/OpenCSGs/csghub/blob/bbb9bf4b346a199a387090de76fcd6c1fffe98f6/frontend/src/packs/useFetchApi.js#L44)在登录 cookie 存在时等待 `refreshJWT`，从 cookie 获取 bearer token，追加语言头；文件末尾统一使用 `${CSGHUB_SERVER}/api/v1`。REA 输出识别本地 `acceptLanguage` 调用和 401 对 `popupReloginDialog` 的调用；跨未选择依赖和真实 token 轮换保持未知。

[csrfFetch.js:1](https://github.com/OpenCSGs/csghub/blob/bbb9bf4b346a199a387090de76fcd6c1fffe98f6/frontend/src/packs/csrfFetch.js#L1)仅把参数传给原生 fetch，没有注入 CSRF token。不能按函数名称推断它提供 CSRF 防护。Astrocyte 现有 HTTP session/origin 检查仍按当前实现与验收要求执行。

## 给 Astrocyte 的可执行 Handoff

本表供 M-product 交给现有负责人，不形成新的开发授权，不改变冻结范围。

| Astrocyte 现有位置（当前基线） | 可执行检查/修改候选 | 验收需要的真实行为 |
| --- | --- | --- |
| `internal/workspace/domain/repositories.go`，`internal/adapters/sqlite/local_agents_repositories.go:53` | 使用已有 metadata revision、placement revision、clone status 和 Head 字段，检查 Git 元数据观察、下载状态、实际 checkout commit 分别来自哪个真实结果；参考镜像导入事务中的记录一致性。 | 下载成功后 Head 来自实际 checkout；失败状态保留准确错误；旧 metadata revision 的更新返回冲突；无需引入上游新调度器。 |
| `internal/attention/app/materials.go:49`、`:86` | 检查 Material 修改版本与附件指定 revision 读取；参考固定 commit 的输入关联，把来源版本放入现有允许的材料来源信息。具体字段变化由现有规格决定。 | 同一来源后续版本变化不会把已记录材料的附件悄悄替换；读取权限在后端校验；版本解析失败立即返回错误。 |
| `internal/adapters/httpapi/session.go` 与已有材料/作品接口 | 检查每个上传、下载、放置和修改请求的 session、资源授权与 revision；前端权限字段仅用于显示。 | 无权请求真实返回拒绝；CSRF/origin 检查保持有效；写权限和 read 权限分别验证。 |
| 评估/研究结果的既有 Event、Knowledge 关联 | 参考模型与数据集固定版本模式，核查结果是否能找到实际输入来源和原始 event；已有域中没有对应字段时先由负责人提出决定。 | 引用能在真实库中找到，不从文件 SHA 推断完整仓库 commit，不以未运行的上游接口作为已验证能力。 |

可直接复用的是源码中的权限检查位置、事务边界、输入版本关联和状态显示方法。第三方代码移植及依赖引入尚未执行；任何复制应先按根许可证及相关文件声明核查。
