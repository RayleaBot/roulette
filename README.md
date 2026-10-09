# 轮盘

RayleaBot 独立插件 `raylea.roulette`，仅提供 OneBot11 群聊俄罗斯轮盘与 Web 设置。需要 Core **0.4.0 或更高版本**、群内机器人管理权限。

玩法参考 [MuteGames](https://github.com/EvolvedGhost/MuteGames/tree/master/src/main/kotlin/roulette)，使用 RayleaBot Go SDK 独立实现。没有自裁、决斗、21 点或其他游戏功能。

## 安装与使用

1. 在 Web 管理面「插件列表」中安装本机平台对应的 ZIP 安装包。
2. 启用「轮盘」，打开插件详情中的「轮盘设置」。
3. 确认机器人有群管理权限，在允许使用该插件的群里发送 `/轮盘`。

| 命令 | 权限 | 行为 |
| --- | --- | --- |
| `/轮盘`、`/俄罗斯轮盘`、`/roulette` | 所有人 | 没有对局时开局并开枪；已有对局时推进一格 |
| `/停止轮盘` | 群管理员或超级管理员 | 停止本群轮盘，不施加超时惩罚 |

命令前缀、群启用范围、名单与冷却由宿主治理配置控制。上表使用默认 `/` 前缀；开枪触发词可以在插件设置页修改。群主或管理员如果无法被机器人禁言，仍可参与，命中时明确提示豁免；机器人是群主且有权限禁言管理员时会正常执行。

## 玩法

- 每个机器人实例、机器人账号与群分别维护一局；同一局任何群成员都可以继续开枪。
- 默认随机 **6–8 个弹膛**、固定 **1 发子弹**，命中禁言随机 **1–300 秒**。数量、位置和禁言时长在开局时确定。
- 子弹耗尽后结束；默认在剩余弹膛全是实弹时提前结束，可关闭。
- 默认开局 **600 秒**后超时。继续开枪不延长时间，默认不惩罚发起人；可开启超时禁言。
- 禁言接口失败或结果未确认时明确提示，本枪仍计入，不自动重试。发送失败也不会重新开枪或回滚子弹位置。
- 插件重启、重载、停用或机器人账号身份变化会结束旧局，不恢复进度、不补禁言。已开始的平台操作无法撤销。

## Web 设置

设置页分三栏。左侧「规则范围」选择默认规则或某个单独设置的群；「为某个群单独设置」打开悬浮面板，按机器人头像、昵称与 QQ 号选择机器人，再按群头像、群名与群号查找并添加群，已单独设置的群会标出并可直接打开；中间「推演一局」按当前范围的规则和台词模拟一局群聊，可切换命中禁言、命中豁免、提前结束、超时和管理员停止等场景，点机器人消息即可就地修改该场景的台词；右侧维护开局口令、弹膛、子弹、禁言、时限与结束条件，底部保存。

单独设置的群按规则单元覆盖：弹膛、子弹、禁言各自整组，时限与两个结束开关单独设置；台词通过「本群专用台词」开关覆盖。未覆盖的项目自动跟随默认规则。

保存会同时校验默认规则与所有群的最终有效规则。删除本群设置并保存后恢复跟随默认；暂时离线的群不会自动被移除。保存配置只影响新局，进行中的对局继续使用开局快照。恢复默认先修改草稿，点击保存后才生效。

台词中的 `<target>` 转换为结构化 @ 提及，其他文本按普通文本发送。支持初始与剩余弹膛/子弹数、禁言秒数与可读时长、超时秒数与可读时长，编辑时可点按钮插入。推演中的群友和抽取值都是示例，不会发送群消息。

机器人与群头像由浏览器直接从 QQ 头像服务（`q1.qlogo.cn`、`p.qlogo.cn`）加载，需要宿主插件页 CSP 允许 HTTPS 图片；不允许或加载失败时显示名称首字头像。

保存需要插件处于运行状态。若显示「已保存，待应用」，配置已经持久化，请使用「重试应用」；也可重载插件，但会结束进行中的对局。重新读取不会把已知应用失败显示为成功。

### 后台等待限制

轮盘使用现有 `event.Detach`，不创建计划任务，也不会占住后续群消息的处理队列。每个进行中的对局使用一个后台事件。

宿主默认允许每插件 **8 个后台事件**、每个 **900 秒**。插件为收尾预留 **30 秒**，因此默认宿主下轮盘超时可设为 **1–870 秒**；宿主后台期限提高后，插件支持的上限是 **3570 秒**。设置页读取当前运行限制，开局仍以宿主实际返回的期限和容量为准。不足时拒绝开局，不会静默缩短设置，也不修改宿主限制。

## 开发与验证

当前版本使用 Go 1.27.2、Node 26.10.0、pnpm 11.25.0；Go SDK 固定为 v0.6.0。Web 使用 Vue 3、TypeScript 与 Vite，产物自带运行时和样式，无外部 CDN。

本地关联 RayleaBot 源码 SDK：

```powershell
node scripts/prepare-local.mjs
corepack pnpm --dir ui install --frozen-lockfile
go test ./...
go vet ./...
corepack pnpm --dir ui typecheck
corepack pnpm --dir ui test
corepack pnpm --dir ui build
```

`go.work` 与 `.rayleabot/` 仅用于本地开发，不进入版本控制。插件无额外 Go 依赖；支持 C 编译器和 CGO 的环境还需执行 `go test -race ./...`。

构建三个正式平台安装包：

```powershell
$env:RAYLEA_PLUGIN_BUILD_USE_WORKSPACE = '1'
go run github.com/RayleaBot/RayleaBot/sdk/go/cmd/raylea-plugin build-go --plugin . --backend ./cmd/roulette --target windows-x64 --out dist --skip-ui-install --skip-ui-build --include README.md=README.md --include docs/interface.md=docs/interface.md --include docs/validation.md=docs/validation.md
go run github.com/RayleaBot/RayleaBot/sdk/go/cmd/raylea-plugin build-go --plugin . --backend ./cmd/roulette --target linux-x64 --out dist --skip-ui-install --skip-ui-build --include README.md=README.md --include docs/interface.md=docs/interface.md --include docs/validation.md=docs/validation.md
go run github.com/RayleaBot/RayleaBot/sdk/go/cmd/raylea-plugin build-go --plugin . --backend ./cmd/roulette --target macos-arm64 --out dist --skip-ui-install --skip-ui-build --include README.md=README.md --include docs/interface.md=docs/interface.md --include docs/validation.md=docs/validation.md
```

使用宿主本地开发流程时，将插件路径加入 RayleaBot 根目录的 `plugin-workspace.local.json`。插件 ID 从 `info.json` 推导，主程序不会将业务插件源码纳入自身构建。

浏览器联调使用本地测试宿主，运行真实插件进程，但机器人身份、群列表和 OneBot 动作全部是测试数据：

```powershell
go build -o dist/smoke/roulette.exe ./cmd/roulette
node scripts/smoke-host.mjs
```

脚本输出仅绑定 `127.0.0.1` 的临时页面地址；按 Ctrl+C 停止。无需真实账号。当前交付验证边界见 [验证记录](docs/validation.md)，设置和生命周期约定见 [插件接口](docs/interface.md)。
