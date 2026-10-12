# 插件接口

本插件使用 RayleaBot manifest v4 / JSONL v4 / artifact v2，不新增宿主接口。`info.json.default_config` 同时提供宿主、后端和页面的默认值。

## 配置

| 顶层键 | 内容 |
| --- | --- |
| `trigger_commands` | 1–20 个全局开枪触发词，每词 1–32 字，不含空白或 `/`、`\`，不可重复或占用「停止轮盘」 |
| `rules` | 默认玩法参数 |
| `replies` | 默认消息文案 |
| `group_overrides` | 最多 500 个群覆盖 |

每个群覆盖由 `bot_id`、`group_id` 唯一确定，协议固定为 OneBot11，不绑定连接实例。旧配置中只带 `source_adapter` 的记录在加载时按该连接当前登录的账号补齐 `bot_id` 并去掉 `source_adapter`；补不齐的记录原样保留，管理页显示为待绑定账号，添加同一群时绑定。可选 `group_name` 仅供展示。`rules` 和 `replies` 只保存显式覆盖的字段，缺少字段表示继承，不用 `null` 表示继承。

| 规则 | 说明 |
| --- | --- |
| `random_chambers` / `chambers` / `chambers_max` | 随机开关、固定值或下限、随机上限；数量 2–100 |
| `random_bullets` / `bullets` / `bullets_max` | 随机开关、固定值或下限、随机上限；数量 1–99 |
| `random_mute` / `mute_seconds` / `mute_max_seconds` | 随机开关、固定值或下限、随机上限；秒数 1–2592000 |
| `timeout_seconds` | 从开局计算的期限，整数 1–3570 秒；还受宿主当前后台期限约束 |
| `timeout_mute` | 超时是否禁言发起人 |
| `end_when_all_loaded` | 剩余均为实弹时是否提前结束 |

随机上下限允许相等。启用随机时上限不小于下限；关闭随机时仍校验上限本身的类型与数值范围，但不参与实际生成和跨参数比较。每个有效作用域的最大子弹数必须小于最小弹膛数。

所有已声明字段保持严格类型校验，拒绝 `null` 和非整数数值；未知字段在所有层级（包括数组中的群覆盖对象）忽略，不写回。

回复场景为 `start`、`miss`、`hit`、`immune`、`end`、`all_loaded`、`timeout`、`timeout_hit`、`timeout_immune`、`stopped`，每条非空且不超过 1000 字。占位符固定为 `<bullet>`、`<chamber>`、`<remain-bullet>`、`<remain-chamber>`、`<mute-s>`、`<mute-f>`、`<timeout-s>`、`<timeout-f>`、`<target>`。拒绝未知占位符，不执行 HTML 或 CQ 消息代码。

## 管理操作

管理动作按页面范围划分：

| 动作 | 页面 | 语义 |
| --- | --- | --- |
| `settings.get` / `settings.save` | 轮盘设置（全局） | 读取或保存全局设置，`values` 不含 `group_overrides`，带群覆盖的保存被拒绝；成功返回 `{values}` |
| `groups.get` / `groups.save` | 群覆盖（按机器人账号） | 账号取自宿主核对过的管理事件，只支持 OneBot11；`groups.save` 校验后只替换当前账号的群覆盖，保留其他账号的覆盖与全局设置，全部群覆盖不超过 500 个 |
| `history.list` / `history.delete` | 轮盘设置（全局） | 列出全部群覆盖记录及其 ID，按 ID 删除单条记录；ID 包含记录位置与内容，记录变化后需重新读取 |
| `history.bind` | 以群覆盖页的账号发起 | 把缺少 `bot_id` 的记录绑定到宿主核对过的 OneBot11 账号；已有账号的记录不能改挂 |
| `history.retry` | 轮盘设置（全局） | 重新提交当前群覆盖，用于已保存待应用 |

所有写入在同一把锁内读取插件最近一次提交的配置后合并，排队中的旧配置快照与插件自身写入产生的 `config.changed` 不会覆盖更新的写入。按账号的动作缺少账号、账号与业务参数中的 `bot_id` 不符或协议不是 OneBot11 时，返回 `platform.invalid_request`。

无效设置返回 `platform.invalid_request`，`details.field` 定位配置项。宿主返回的 `plugin.settings_apply_failed` 和 `details.committed` 原样传递，页面保留草稿并提供重试。未知管理动作不产生副作用。

## 对局生命周期

每局在进程内持有不可变规则/回复快照、弹膛排列、位置、剩余子弹和绝对截止时间。按实例、账号、群进行互斥访问，不持久化进度。配置由宿主持久化，页面只有编辑草稿。

开局先验证身份、权限和配置，再转入后台并验证截止时间。每局只转入后台一次，后续消息使用新的普通事件。开局和首枪使用非终态消息动作，后台等待期间不持有群锁或前台并发许可。超时、耗尽、提前结束、管理员停止或身份变化唤醒等待并结束事件；旧局清理仅清理自身。

枪位消费先于禁言动作，失败不回退枪位或自动重试。管理员停止不施加超时惩罚。身份变化或进程停止结束旧局，已提交的平台副作用无法回滚。
