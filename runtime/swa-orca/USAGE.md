# SWA 无界面 CLI

Windows x64 包包含`swa.exe`、`server.exe`、固定Node24.21.0、私有Orca运行文件、安装入口与许可。`VERSION`和`swa.exe --version`给出同一准确源码提交。当前首版尚未通过同一安装候选的完整真实验收，实际验证范围以随交付的记录为准。

## 安装

使用PowerShell7.2以上，明确选择已有候选目录、尚不存在的安装目录和本机Orca及账户保护目录。安装目标父目录必须存在。下列变量由使用者设置为实际绝对路径，端口由使用者明确选择；命令不会修改PATH或注册后台服务。

```powershell
pwsh -NoProfile -File "$packageRoot/install-swa.ps1" -PackageRoot $packageRoot -InstallRoot $installRoot -ProtectedRoots $protectedRoots
$swa = Join-Path $installRoot 'swa.exe'
& $swa --version
& $swa --help
```

`$protectedRoots`是字符串数组，包含本机Orca安装目录、用户数据/profile目录及实际账户配置目录。安装目录与候选目录均须和保护目录互不包含；途经reparse point的路径会被拒绝。安装仅创建明确目标并设置当前用户私有ACL，已有目标不会覆盖。

## 启动私有服务

`$dataRoot`为产品独占绝对目录，须与安装目录、原Orca及所有保护目录互不包含。默认发现本机原生命令和配置。需要采用既有Orca的Agent设置时，显式提供`--agent-profile $agentProfile`，其中`$agentProfile`是已经存在的原profile绝对目录；该来源只读，缺失或损坏时明确失败。

```powershell
& $swa serve --data-dir $dataRoot --port $port --protected-roots ($protectedRoots -join ';') --json
```

该命令在前台持有所属stdin，输出实际Node/HTTP readiness后持续运行。关闭该stdin或Ctrl+C请求关闭本次所属服务；不要把stdin预先重定向为空。服务监听显式`127.0.0.1`端口，冲突返回退出码5。缺少资源额度、依赖或所属证据会明确拒绝；不会另选端口或启动本机Orca。进程、SQLite、profile、IPC、会话、缓存和临时目录均归本次私有根。

## 项目授权与调用身份

命令端需要该私有HTTP服务中已经登记的真实项目ID和当前human批准的项目Agent token。原Orca的token、终端`--from`、进程PID及本机开发Agent身份均不能代替项目批准。新数据根没有预置Grant，服务启动不会发行token。

使用者须通过现有Workspace的人工授权路径在该服务中登记明确公开项目，检查设置并签发token：项目登记接口为`POST /api/v1/local-projects`，带实际`name/root/space_id`；设置为`PUT /api/v1/local-projects/{id}/settings`，带当前`expected_revision`及完整`settings`；批准为`POST /api/v1/local-projects/{id}/grants`，带实际`agent_id`和`actions`；签发为`POST /api/v1/local-projects/{id}/agent-token`。人工请求沿用实际human会话cookie及`X-CSRF-Token`，请求正文保留既有`schema_version/request_id/expected_version`，Agent不能自动取得human会话或替使用者批准。

批准范围须覆盖当前固定公开项目、允许读取的文件/子目录，以及`read_context/observe/start/send/stop`中实际需要的动作。`allow_agent_control`、`allow_directory`、`allowed_subdirs`、`allowed_actions`和`allowed_tools`由当前批准确定，`external_model_cli`继续只批准一个实际CLI标识。`capabilities`分别显示注册、启用、安装、配置和各项控制能力及来源；声明能力和实际启动结果分别记录。更改设置会改变revision，旧上下文须按当前批准重新取得。缺少人工批准时保留退出码3及required_action。

普通终端与现有Agent原生命令工具均调用同一安装后的绝对路径；只在调用进程内提供环境，不把token写入输入文件、命令正文或验证日志。

```powershell
$env:SWA_ENDPOINT = "http://127.0.0.1:$port"
$env:SWA_PROJECT_ID = $projectId
$env:SWA_PROJECT_TOKEN = $projectToken
& $swa status --json
& $swa capabilities --json
& $swa controller open --project $projectId --workdir $projectRoot --request-id ([Guid]::NewGuid().ToString()) --json
```

后继controller、Run、根Task及Dispatch全部使用真实回执。固定`task.json`包含`goal/source_repo/source_revision/success_criteria/max_workers:2/context`；context引用批准的实际公开文件。`a.txt/b.txt`和后继输入位于该项目的允许目录，文件内容作为真实模型输入。

```powershell
& $swa task create --controller $controller --spec-file $specFile --request-id $createRequest --json
& $swa worker start --controller $controller --task $rootTask --name A --agent $agent --model $model --input-file $inputA --request-id $startARequest --json
& $swa worker start --controller $controller --task $rootTask --name B --agent $agent --model $model --input-file $inputB --request-id $startBRequest --json
& $swa worker show --controller $controller --dispatch $dispatchA --json
& $swa worker read --controller $controller --dispatch $dispatchB --source transcript --limit 100 --json
& $swa worker stop --controller $controller --dispatch $dispatchA --request-id $stopARequest --json
& $swa worker send --controller $controller --dispatch $dispatchB --input-file $followupFile --request-id $sendBRequest --json
& $swa task result --controller $controller --task $rootTask --output-dir $newOutputDirectory --json
```

每个写动作的request变量为首次调用前保存的不同规范UUID。查询结果包含真实原生状态，查询成功不代表任务成功。退出码6表示效果未知：保留原request，使用`request show --controller ... --request-id ... --json`查询；controller尚未取得时使用`--project`形式。未知输入不能换键重发。

指定停止仅处理该Worker的专有Job；root退出、tree退出和资源释放分别记录。全部相关Worker确认退出并按原生流程归档释放后，可显式提交固定成功条件的真实结果，再关闭controller：

```powershell
& $swa worker stop --controller $controller --dispatch $dispatchB --request-id $stopBRequest --json
& $swa worker release --controller $controller --dispatch $dispatchA --request-id $releaseARequest --json
& $swa worker release --controller $controller --dispatch $dispatchB --request-id $releaseBRequest --json
& $swa task complete --controller $controller --task $rootTask --result-file $checkedResultFile --request-id $completeRequest --json
& $swa controller close --controller $controller --request-id $closeRequest --json
```

`$agent`和`$model`必须是当前真实发现、明确选择并获准使用的本机Agent和模型。后续输入沿原Dispatch的启动选择，不能更换Agent或模型。任一退出/归档证据未知时保留占用并检查原请求；模型调用、原生工具和账号读取须满足当前任务的准确批准及已验证的隔离能力。
