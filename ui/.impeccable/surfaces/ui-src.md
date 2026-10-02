---
version: 1
slug: "ui-src"
primary_target: "ui/src"
related_targets: []
---

# 轮盘设置页

## Scope

插件管理页 `settings`（`ui/src`），宿主 iframe 内的完整设置界面。访问模式：Operate。

## Audience and job

自托管 RayleaBot 的机器人管理员，低频打开，多数时候只调全局默认。任务：看懂当前规则下一局在群里会怎样发生，调整弹膛、子弹、禁言、时限与台词，按群单独设置少数差异，保存并确认是否已应用。

## Constraints

- 只在宿主桌面工作区内验证（1920×1080 宿主，iframe 约 1880×960），iframe 高度固定，页面自己管理滚动。
- 推演只用示例群友、示例群名和示例数值，始终标注不会发送；不连接真实开局、消息或禁言。
- 不加载外部字体与脚本；头像直接显示 QQ 头像服务的 HTTPS 图片，失败时显示首字；亮暗主题跟随宿主 `data-theme`。
- 保留现有能力：校验、草稿/已保存/待应用三态、重新读取、恢复默认、群覆盖继承、离线群可编辑、运行限制读取。

## Memorable moment

在右侧把禁言上限从 300 改到 600，中间那条「老王 被禁言 2分30秒」立刻变成「5分」并闪一下。

## Unresolved

- 群覆盖的继承粒度改为按规则单元（弹膛、子弹、禁言各自整组），旧数据中只覆盖单个字段的群仍按「单独设置」显示并保留原字段。

## Direction contract

THESIS：设置页的主画面是一局在群里的推演回放；规则改到哪里，群消息和禁言提示就跟到哪里。拒绝「左说明右控件的灰色表单行 + 顶部页签」的默认排法。

OWN-WORLD：聊天应用的语言。浅灰聊天底、白色成员气泡、淡蓝机器人气泡、居中的灰色系统提示条、蓝色 @；左侧会话列表式的范围切换，右侧白色规则检查器；系统中文字体与等宽数字；蓝色只给操作、提及与焦点。

STORY：管理员先选「所有群」或某个群，看这一局在群里怎样发生；在检查器里改数值时，推演里对应的数字同步高亮；点机器人消息就地改台词；最后在检查器底部保存，并看到已同步、未保存或待应用。

FIRST VIEWPORT：三栏铺满 iframe。左 288px 规则范围列表：默认规则置顶，下面是单独设置的群与「为某个群单独设置」入口。中间是推演群聊：顶部场景切换与「换一组示例数值」，底部示例取值说明。右 452px 规则检查器：首枪中弹、禁言、时限速览，口令、装填、惩罚、收局分组，底部固定保存区与「保存设置」主按钮。

FORM：群聊推演，方向列表第 1 位（IMPECCABLE’S PICK），seed key 7c24eb90。Signature interaction：规则联动高亮，聚焦或修改检查器中的任一项，推演中对应的数字与提示即时高亮，减少动态效果时改为静态标记。

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
