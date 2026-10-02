---
name: "RayleaBot 轮盘插件设置"
description: "群聊推演：用一局模拟群聊呈现规则，检查器改值即在推演中联动高亮；独立于宿主的聊天应用视觉。"
colors:
  canvas: "#ffffff"
  panel: "#ffffff"
  chat: "#eef0f3"
  sunken: "#f2f3f5"
  ink: "#1b1d22"
  ink-2: "#5a606b"
  line: "#e3e6ea"
  line-strong: "#d0d4da"
  control: "#8a909a"
  accent: "#1f6feb"
  accent-hover: "#175cd3"
  accent-ink: "#ffffff"
  accent-text: "#1559c9"
  bubble: "#ffffff"
  bot-bubble: "#e2ecff"
  notice: "#e0e3e8"
  notice-ink: "#474d57"
  lit: "#ffe9a8"
  flash: "#ffd45c"
  lit-edge: "#9a6b00"
  ok: "#1d7a43"
  warn: "#c27a00"
  danger: "#b42318"
  danger-soft: "#fdeceb"
  danger-line: "#e7a39c"
  admin: "#f2c14e"
  admin-ink: "#3a2a00"
  selected: "#e5e8ed"
  bot-avatar: "#1b1d22"
  bot-avatar-ink: "#ffffff"
  switch-off-ink: "#12141a"
  backdrop: "rgb(15 17 21 / 0.42)"
  canvas-dark: "#15171a"
  panel-dark: "#1b1e22"
  chat-dark: "#101214"
  sunken-dark: "#24282d"
  ink-dark: "#eceef1"
  ink-2-dark: "#a8aeb8"
  line-dark: "#2b3036"
  line-strong-dark: "#3b4149"
  control-dark: "#78808b"
  accent-dark: "#6a9fff"
  accent-hover-dark: "#87b2ff"
  accent-ink-dark: "#0b1220"
  accent-text-dark: "#8fb6ff"
  bubble-dark: "#262a30"
  bot-bubble-dark: "#1c2a44"
  notice-dark: "#2a2e35"
  notice-ink-dark: "#b7bdc6"
  lit-dark: "#4a3a0c"
  flash-dark: "#6e570f"
  lit-edge-dark: "#f2b84b"
  ok-dark: "#5cc98a"
  warn-dark: "#f2b84b"
  danger-dark: "#ff8a80"
  danger-soft-dark: "#3a1f1d"
  danger-line-dark: "#7a3a35"
  selected-dark: "#2b3139"
  bot-avatar-dark: "#39404a"
  bot-avatar-ink-dark: "#dfe4ea"
  switch-off-ink-dark: "#0e1013"
  backdrop-dark: "rgb(0 0 0 / 0.6)"
  tone-1: "#2f6fd0"
  tone-2: "#a85a17"
  tone-3: "#2f7d47"
  tone-4: "#b03a74"
  tone-5: "#6c51c4"
  tone-6: "#1d7680"
typography:
  dialog:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "18px"
    fontWeight: 700
    lineHeight: 1.6
  headline:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "17px"
    fontWeight: 700
    lineHeight: 1.6
  figure:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "17px"
    fontWeight: 700
    lineHeight: 1.6
    letterSpacing: "-0.01em"
    fontFeature: "\"tnum\""
  title:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "16px"
    fontWeight: 700
    lineHeight: 1.6
  section:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "15px"
    fontWeight: 700
    lineHeight: 1.6
  message:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.7
  body:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.6
  control:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "14px"
    fontWeight: 600
    lineHeight: 1.6
  label:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "13px"
    fontWeight: 600
    lineHeight: 1.6
  caption:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: 1.6
  tag:
    fontFamily: "\"Microsoft YaHei UI\",\"PingFang SC\",\"Hiragino Sans GB\",system-ui,sans-serif"
    fontSize: "11px"
    fontWeight: 700
    lineHeight: "18px"
rounded:
  mark: "4px"
  tag: "5px"
  inner: "8px"
  field: "9px"
  control: "10px"
  row: "12px"
  card: "14px"
  bubble: "16px"
  sheet: "18px"
  pill: "999px"
spacing:
  inline-gap: "6px"
  stack-gap: "8px"
  message-gap: "10px"
  replay-gap: "14px"
  block-y: "16px"
  head-x: "24px"
  pane-x: "26px"
  stream-x: "32px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.accent-ink}"
    typography: "{typography.control}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "38px"
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
  button-secondary:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.control}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "38px"
  button-secondary-hover:
    backgroundColor: "{colors.sunken}"
  button-quiet:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    typography: "{typography.control}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "38px"
  button-danger:
    backgroundColor: "transparent"
    textColor: "{colors.danger}"
    typography: "{typography.control}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "38px"
  button-danger-hover:
    backgroundColor: "{colors.danger-soft}"
  button-small:
    typography: "{typography.label}"
    padding: "0 12px"
    height: "32px"
  text-link:
    backgroundColor: "transparent"
    textColor: "{colors.accent-text}"
    typography: "{typography.label}"
  segmented:
    backgroundColor: "{colors.sunken}"
    rounded: "{rounded.control}"
    padding: "3px"
  segmented-option:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    typography: "{typography.label}"
    rounded: "{rounded.inner}"
    padding: "0 10px"
    height: "30px"
  segmented-option-pressed:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
  number-field:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.control}"
    rounded: "{rounded.field}"
    height: "36px"
  reply-editor:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.message}"
    rounded: "{rounded.control}"
    padding: "10px 12px"
  switch:
    backgroundColor: "{colors.control}"
    rounded: "{rounded.pill}"
    width: "48px"
    height: "26px"
  switch-on:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.accent-ink}"
  scope-row:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    rounded: "{rounded.row}"
    padding: "10px"
  scope-row-hover:
    backgroundColor: "{colors.sunken}"
  scope-row-current:
    backgroundColor: "{colors.selected}"
  member-bubble:
    backgroundColor: "{colors.bubble}"
    textColor: "{colors.ink}"
    typography: "{typography.message}"
    rounded: "{rounded.bubble}"
    padding: "9px 14px"
  bot-bubble:
    backgroundColor: "{colors.bot-bubble}"
    textColor: "{colors.ink}"
    typography: "{typography.message}"
    rounded: "{rounded.bubble}"
    padding: "9px 14px"
  system-notice:
    backgroundColor: "{colors.notice}"
    textColor: "{colors.notice-ink}"
    typography: "{typography.caption}"
    rounded: "{rounded.pill}"
    padding: "4px 12px"
  rule-lit:
    backgroundColor: "{colors.lit}"
    rounded: "{rounded.mark}"
  rule-flash:
    backgroundColor: "{colors.flash}"
    rounded: "{rounded.mark}"
  glance:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.figure}"
    rounded: "{rounded.card}"
    padding: "10px 14px"
  glance-lit:
    backgroundColor: "{colors.lit}"
  chat-tag:
    backgroundColor: "{colors.notice}"
    textColor: "{colors.notice-ink}"
    typography: "{typography.tag}"
    rounded: "{rounded.tag}"
    padding: "0 6px"
  chat-tag-admin:
    backgroundColor: "{colors.admin}"
    textColor: "{colors.admin-ink}"
  chat-tag-own:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.canvas}"
  chat-tag-error:
    backgroundColor: "{colors.danger-soft}"
    textColor: "{colors.danger}"
  inherit-tag:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    rounded: "{rounded.pill}"
    padding: "0 7px"
  inherit-tag-own:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.canvas}"
  count-badge:
    backgroundColor: "{colors.notice}"
    textColor: "{colors.ink}"
    rounded: "{rounded.pill}"
    padding: "0 7px"
    height: "22px"
  scope-chip:
    backgroundColor: "{colors.notice}"
    textColor: "{colors.ink}"
    rounded: "{rounded.pill}"
    padding: "2px 9px"
  trigger-chip:
    backgroundColor: "{colors.sunken}"
    textColor: "{colors.ink}"
    typography: "{typography.control}"
    rounded: "{rounded.field}"
    padding: "0 4px 0 11px"
    height: "32px"
  branch-option:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    typography: "{typography.label}"
    rounded: "{rounded.pill}"
    padding: "0 12px"
    height: "28px"
  branch-option-pressed:
    backgroundColor: "{colors.sunken}"
    textColor: "{colors.ink}"
  timeout-meter:
    backgroundColor: "{colors.line}"
    rounded: "{rounded.pill}"
    height: "6px"
  confirm-strip:
    backgroundColor: "{colors.danger-soft}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "10px 12px"
  avatar:
    rounded: "50%"
    size: "42px"
  avatar-lg:
    size: "48px"
  avatar-sm:
    size: "36px"
  search-field:
    backgroundColor: "{colors.sunken}"
    textColor: "{colors.ink-2}"
    rounded: "{rounded.control}"
    padding: "0 12px"
    height: "36px"
  icon-close:
    backgroundColor: "transparent"
    textColor: "{colors.ink-2}"
    rounded: "{rounded.control}"
    size: "36px"
  dialog-backdrop:
    backgroundColor: "{colors.backdrop}"
  add-sheet:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.sheet}"
    width: "940px"
    height: "660px"
  bot-pane:
    backgroundColor: "{colors.chat}"
    padding: "16px 14px"
    width: "300px"
  bot-card:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.card}"
    padding: "12px"
  bot-card-disabled:
    backgroundColor: "transparent"
  group-card:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.card}"
    padding: "10px 12px"
  group-card-hover:
    backgroundColor: "{colors.sunken}"
  group-card-selected:
    backgroundColor: "{colors.selected}"
---

# Design System: RayleaBot 轮盘插件设置

## Overview

**Creative North Star: "群聊推演"**

设置页把规则放进一局模拟群聊里演给管理员看。中间是聊天应用式的推演回放，左侧像会话列表一样切换「默认规则」和单独设置的群，右侧是白色规则检查器；聚焦或改动检查器里的任一项，推演里由它决定的数字或消息随之亮起。视觉身份独立于宿主管理面，材料取自聊天应用本身：浅灰聊天底、白色成员气泡、淡蓝机器人气泡、居中的灰色系统提示条和蓝色 @。

密度安静而紧凑：系统中文字体，最大字号 18px（只给添加群面板标题），变化的数字一律等宽。色彩分工严格：蓝色只属于动作、提及、链接、开关开启与焦点；淡蓝只属于机器人气泡；联动黄只用于规则联动高亮；状态、计数、范围标签和时限条保持中性，群专属用墨色实心标签。层次来自聊天底与白色面板的明暗、1px 细线和聊天底上白色物件（成员气泡、机器人卡片）的一层轻影；添加群面板是页面唯一弹出的一层。

动效只服务功能：规则联动的点亮与改值闪烁，悬停底色、开关滑块、时限条等控件状态的短过渡，以及添加群面板 0.22s 的进场和读取中骨架卡片的呼吸；减少动态效果时全部取消，闪烁改为静态标记。会丢弃内容的操作在页面内的确认条里完成；为某个群单独设置时，在模态的添加群面板里挑选机器人和群。这个方向明确拒绝「左说明右控件的灰色表单行 + 顶部页签」的默认设置页排法。

本文件记录 `ui/src` 中已实现的视觉系统，token 数值以 `ui/src/style.css` 为准；页面结构与交互约定见 `docs/settings-design.md`，方向约定见 `ui/.impeccable/surfaces/ui-src.md`。

**Key Characteristics:**

- 三栏铺满 iframe：会话列表式的规则范围、群聊推演回放、白色规则检查器。
- 聊天应用材质：浅灰聊天底、白色成员气泡、淡蓝机器人气泡、居中的灰色系统提示条。
- 蓝色只给动作、提及、链接、开关与焦点；状态与计数保持中性，「本群」用墨色实心。
- 规则联动高亮：检查器每一项都对应推演中的数字或整条消息，悬停或聚焦时点亮，改值时闪一次。
- 系统中文字体与等宽数字；不加载外部字体或脚本，只有 QQ 头像图片经 HTTPS 直接加载，图标为内联 SVG 线描。

## Colors

冷灰中性色承托聊天场景，有职责的色彩只有三种：操作蓝、机器人淡蓝和联动黄；状态色只出现在状态点、错误提示和角色标签上。前置 token 是数值来源：无后缀项属于浅色主题，带 `-dark` 后缀的同名项属于暗色主题；`admin`、`admin-ink` 与 `tone-1` 至 `tone-6` 两套主题共用。主题优先采用宿主同步到根元素的 `data-theme`，未指定时跟随系统。

### Primary

- **操作蓝**（`accent`）：主按钮底色、开关开启、焦点轮廓、文本选区和原生控件强调色；悬停用 `accent-hover`，其上的文字用 `accent-ink`。
- **链接与提及蓝**（`accent-text`）：文字链接、推演中的 @ 提及，以及编辑入口、添加入口和占位符按钮的悬停文字。浅色主题中比操作蓝深一档、暗色主题中浅一档，保证小字在面板和淡蓝气泡上可读。

### Secondary

- **机器人淡蓝**（`bot-bubble`）：只作机器人消息气泡的底色，让台词在聊天流里一眼可辨；不用于按钮、选中态或提示。

### Tertiary

- **联动浅黄**（`lit`）：规则联动的点亮色。检查器某一项被悬停或聚焦时，推演中由它决定的数字和速览条对应格铺上这层底色。
- **联动描边**（`lit-edge`）：规则联动的边线色，比 `lit` 深得多，让联动不只靠底色：点亮的数字下加一道下划线，收局开关决定的整条机器人气泡或系统提示外加一圈外环。
- **闪烁金黄**（`flash`）：改值后对应数字闪一次的峰值色，随后淡出。

### Neutral

- **画布与面板**（`canvas`、`panel`）：`panel` 是范围列、检查器、推演头部和底部说明条的白色面板；`canvas` 是页面底，也是数值框、口令输入框和文本框的底色。
- **聊天底**（`chat`）：推演对话区、载入页与添加群面板机器人栏的底色，比面板深一档，让白色成员气泡和机器人卡片浮出来。
- **凹陷面**（`sunken`）：悬停底、分段切换轨道、搜索框、口令标签、占位符按钮和按下的结局选项。
- **当前范围**（`selected`）：范围列表中的当前项，以及添加群面板中选中的群卡片。
- **墨色**（`ink`、`ink-2`）：正文与次级文字；`ink` 也作「本群」「已单独设置」标签的实心底和选中卡片的描边，`ink-2` 也作时限条的填充。
- **线与控件边界**（`line`、`line-strong`、`control`）：`line` 用于面板分隔、速览格线、时限条轨道、群卡片描边和骨架占位；`line-strong` 用于折叠分隔线、结局选项与继承标签描边、文本框边界、机器人气泡的悬停外环和卡片的悬停描边；`control` 用于按钮、数值框、口令输入框的边界、虚线添加入口和关闭状态的开关轨道。
- **成员气泡**（`bubble`）：群友消息气泡。
- **系统提示灰**（`notice`、`notice-ink`）：居中的系统提示条；`notice` 也是机器人标签、计数徽标和推演头部范围标签的底色。
- **机器人头像**（`bot-avatar`、`bot-avatar-ink`）：机器人头像的底色与转轮线描。
- **开关关闭字**（`switch-off-ink`）：关闭状态开关轨道上的「关」字，深色压在 `control` 轨道上。
- **遮罩**（`backdrop`）：模态的添加群面板打开时压暗整页的半透明底，暗色主题更深。

### 状态色

- **已同步绿、待处理橙、错误红**（`ok`、`warn`、`danger`）：保存区状态点与添加群面板里机器人实例的状态点，状态点总与一句状态文字同时出现。`danger` 也用于字段错误、危险按钮文字和无效输入边框。
- **错误浅底与危险描边**（`danger-soft`、`danger-line`）：`danger-soft` 是确认条、待修正清单、推演暂停提示和「需修正」标签的底色，也是危险按钮的悬停底；`danger-line` 是危险按钮的描边。
- **管理员黄**（`admin`、`admin-ink`）：推演中群管理员的角色标签，沿用聊天应用的惯例。

### 成员头像色

- **六色头像**（`tone-1` 至 `tone-6`）：所有圆形头像的底色与后备：底层是色盘加白色首字，真实 QQ 头像加载后盖在上面，加载中或失败时露出首字；推演里的示例群友只用首字。颜色按群号或 QQ 号固定取色，同一个群始终同色；这组颜色只用于头像。

### Named Rules

**「蓝色只给动作」规则。** 蓝色只出现在可操作的东西和被 @ 的人身上：主按钮、链接、提及、开关开启、焦点与选区。状态点、计数徽标、各类标签、时限条和装饰一律不用蓝；淡蓝只属于机器人气泡。

**「点亮只指对应」规则。** `lit`、`lit-edge` 与 `flash` 只表示「检查器这一项决定了推演里这个数字或这条消息」，不用于警告、选中、搜索命中或装饰。

**「本群用墨」规则。** 群专属用墨色实心标签，跟随默认用中性描边标签；两者靠明度反转区分，不另加颜色。

## Typography

**Display Font:** 无。页面没有展示型字号；最大的 18px 只给添加群面板标题，其次是检查器标题和速览数值（17px）。
**Body Font:** 系统中文字体栈（Microsoft YaHei UI，回退 PingFang SC、Hiragino Sans GB、system-ui、sans-serif），所有角色共用，不下载字体。
**Label/Mono Font:** 无独立字体；数字对齐靠等宽数字特性。

**Character:** 聊天应用自己的系统字体，清楚、紧凑、不表演；层级只靠 11–18px 的字号和 400 / 600 / 700 三档字重拉开，开关轨道内的「开 / 关」另用 10–11px。

### Hierarchy

- **Dialog**（700，18px，1.6）：添加群悬浮面板的标题。
- **Headline**（700，17px，1.6）：检查器顶部的范围标题，即「默认规则」或群名，过长时省略。
- **Figure**（700，17px，1.6，字距 -0.01em，等宽数字）：速览条中的首枪中弹、禁言与时限数值。
- **Title**（700，16px，1.6）：栏标题，如「规则范围」「推演一局」。
- **Section**（700，15px，1.6）：检查器分组标题：开局口令、装填、惩罚、收局。
- **Message**（400，15px，1.7）：聊天气泡与台词编辑框，保留换行，长串可在任意处断行。
- **Body**（400，14px，1.6）：默认正文。
- **Control**（600，14px，1.6）：按钮、规则行标签、口令标签、数值输入与群卡片的群名；范围列表与机器人卡片中的名称、面板栏标题加重到 700。
- **Label**（600，13px，1.6）：小号按钮、分段切换、文字链接与结局选项；确认说明、面板描述与机器人 QQ 号同字号用 400。
- **Caption**（400，12px，1.6）：说明、元信息、单位、时间戳、系统提示与示例取值说明；字段错误同字号加重到 600。
- **Tag**（700，11px，行高 18px）：聊天消息上的机器人、管理员、本群、正在编辑、需修正标签；继承标签同字号用 600。

### Named Rules

**「数字等宽」规则。** 会随设置变化的数字——数值输入、速览数值、计数徽标、群号与账号行、示例取值说明、字数计数——一律使用等宽数字，改值时版面不跳动。

## Layout

页面是铺满宿主 iframe 的三栏网格：左侧规则范围列（288px），中间推演回放占据剩余宽度，右侧规则检查器（452px）；最小高度（520px）。推演的对话列居中，最大宽度（840px），就地台词编辑器最大宽度（640px）。推演顶部是标题、范围标签、示例标注、场景切换与「换一组示例数值」（最小高度 64px），超时场景下多一条结局选择条，底部固定一条示例取值说明。检查器的规则内容在固定保存区上方滚动；范围列底部固定「为某个群单独设置」入口。

宽度不超过（1360px）时两侧栏收窄为（248px）与（408px），对话列水平内边距收至（22px），检查器与保存区收至（20px）；放不下的上限输入换到下一行。宽度不超过（1080px）时三栏上下堆叠，页面恢复整体滚动：范围列表最高（280px），推演对话区最高（70vh），保存区粘在视口底部。验证基准是宿主 1920×1080 下约（1880×960）的 iframe；1440 与 1280 宽度保持三栏。

添加群悬浮面板居中于视口，最大（940×660px），与视口四边至少留（24px）。面板分上中下三段：头部、主体、底部操作；主体左侧是机器人栏（300px），右侧群栏的群卡片按最小（250px）自动填充成网格、间距 `stack-gap`，两栏各自滚动。

间距按 spacing token 组织：同一行内的图标、标签与控件用 `inline-gap`；保存区和确认条的纵向堆叠、操作按钮之间用 `stack-gap`；头像与气泡之间、编辑器内部用 `message-gap`；推演消息之间用 `replay-gap`；检查器分组以 `block-y` 起头。推演头部与结局条的水平内边距为 `head-x`，检查器与保存区为 `pane-x`，对话列与示例说明为 `stream-x`。规则行和开关行向两侧外扩（10px），悬停底色铺满而文字仍与分组标题对齐。

粗指针环境下，按钮、分段选项、数值框、口令输入框、「添加」口令入口、「为某个群单独设置」入口、面板关闭按钮、结局选项和占位符按钮最小高度（44px），范围行最小（56px），文字链接、编辑入口和口令删除按钮最小（44×44px）。

**「三栏各自滚动」规则。** 桌面宽度下页面本身不滚动，三栏各自滚动，推演头部、示例取值说明和保存区始终可见；窄于 1080px 才改为整页滚动。

## Elevation & Depth

整体是平的，靠明暗分层：聊天底比白色面板深一档（暗色主题中同样更深），面板之间只用 1px 细线分隔，检查器分组用顶部细线而不是卡片。投影只给少数真正浮起的东西；高亮外扩、整行点亮外环、机器人气泡悬停外环和「正在编辑」内描边虽用 box-shadow 绘制，但属于描边，不表示高度。添加群面板是唯一的弹出层：`backdrop` 遮罩压暗整页，面板自身只有一层弹出投影，内部用聊天底的机器人栏和细线分区，不再叠加层级。

### Shadow Vocabulary

- **气泡轻影**（`box-shadow: var(--shadow-1)`；浅色 `0 1px 2px rgb(16 24 40 / 0.06)`，暗色 `0 1px 2px rgb(0 0 0 / 0.35)`）：白色成员气泡、添加群面板机器人栏里的白色机器人卡片，以及分段切换中被按下的选项（另加 1px `line` 外环）。机器人气泡不带影，靠淡蓝底区分。
- **浮起投影**（`box-shadow: var(--shadow-2)`；浅色 `0 4px 16px rgb(16 24 40 / 0.1)`，暗色 `0 4px 16px rgb(0 0 0 / 0.4)`）：就地展开的台词编辑器和载入卡。
- **弹出投影**（`box-shadow: var(--shadow-pop)`；浅色 `0 24px 64px rgb(16 24 40 / 0.24)`，暗色 `0 24px 64px rgb(0 0 0 / 0.6)`）：添加群悬浮面板，页面里最高也是唯一的弹出层。
- **保存区上沿**（`box-shadow: var(--shadow-up)`；浅色 `0 -6px 16px rgb(16 24 40 / 0.04)`，暗色 `0 -6px 16px rgb(0 0 0 / 0.3)`）：检查器底部固定保存区与上方滚动内容的分界，配合顶部细线。
- **开关滑块**（`box-shadow: 0 1px 2px rgb(0 0 0 / 0.25)`）：开关的白色圆形滑块。

### Named Rules

**「气泡才有影」规则。** 静止的面板、规则行、标签、速览条和群卡片都是平的；常驻轻影只给聊天底上的白色物件（成员气泡、机器人卡片）和按下的分段选项，浮起投影只给就地编辑器与载入卡，弹出投影只给添加群面板。

## Shapes

圆角随物件尺寸递增：高亮标记（`mark`）、聊天标签（`tag`）、分段内选项（`inner`）、数值框与口令（`field`）、按钮、文本框与确认条（`control`）、范围行与规则行（`row`）、速览条与选择卡片（`card`）、气泡与编辑器（`bubble`）、添加群面板（`sheet`）依次放大。表示状态或归属的小件——系统提示条、继承标签、计数徽标、范围标签、结局选项、开关和时限条——用胶囊形（`pill`）。头像与状态点是正圆：头像默认（42px，范围列表、检查器标题、群卡片），机器人卡片用大号（48px），推演与面板底部的已选群用小号（36px），真实头像在圆内裁切铺满；状态点在保存区为（9px），在机器人卡片里为（7px）。

聊天气泡左上角收成 `mark` 小角、其余三角为 `bubble`，形成指向头像的气泡尾；就地编辑器沿用同一轮廓。高亮标记用同色外扩（点亮 2px、闪烁 3px）留出边距，不推动文字，点亮时另有 2px 下划线（偏移 3px）；整行点亮沿整条气泡或系统提示的轮廓外加 2px 外环。

边界一律 1px 实线：可操作控件用较深的 `control`，分隔与容器用 `line` 或 `line-strong`。键盘焦点统一为操作蓝 3px 外轮廓、2px 偏移；数值框和搜索框把焦点画在整个外框上。图标是内联 SVG 线描：16px、2px 描边、圆头圆角、跟随文字颜色，系统提示内缩至 14px；机器人头像是转轮弹巢线描。

**「虚线表示新增」规则。** 虚线边框只用于新增入口：「为某个群单独设置」（1.5px，悬停时虚线转为操作蓝）与「添加」口令（1px，悬停转为实线）。

## Components

### Buttons

- **形状：** `control` 圆角、1px 边界、600 字重；常规高（38px），小号高（32px）、13px 字。图标与文字间距 `inline-gap`。
- **主按钮：** 操作蓝底、`accent-ink` 字，悬停 `accent-hover`。用于保存设置（待应用时为「重试应用」）、添加群面板的「添加这个群」（已设置过的群为「打开已有设置」）和编辑器的「完成」；一个区域只放一个。
- **次按钮：** 面板底、墨色字、`control` 边界，悬停为凹陷面。用于重新读取、换一组示例数值、恢复插件默认与取消。
- **安静按钮：** 无边框无底色、次级墨色字；保存区的「恢复默认」靠右放置。
- **图标关闭按钮：** （36px）见方、`control` 圆角、无边框无底色，次级墨色 18px 叉号，悬停铺凹陷面并转为墨色；用于添加群面板右上角。
- **危险按钮：** 透明底、`danger` 字、`danger-line` 描边，悬停铺 `danger-soft`。用于删除本群设置和确认条中的破坏性确认。
- **文字链接：** `accent-text`、13px/600，悬停才出现下划线；在错误浅底上改用 `danger` 并常显下划线。
- **禁用：** 透明度（0.5），停止悬停反馈。

### Chips

- **口令标签：** 凹陷面底、`line` 描边、`field` 圆角、高（32px）；命令前缀「/」用次级墨色，右侧是（24px）删除按钮。「添加」是同尺寸的虚线入口，悬停转实线并变为 `accent-text`；输入态是（150px）输入框，回车添加、Esc 收起。
- **继承标签：** 群范围下每个规则单元前的胶囊：「跟随默认」为 `line-strong` 描边、次级墨色字；「本群」为墨色实心、`canvas` 字。
- **聊天标签：** `tag` 圆角：机器人（系统提示灰）、管理员（管理员黄）、本群（墨色实心）、正在编辑（1px 墨色内描边）、需修正（错误浅底、`danger` 字）；添加群面板群卡片上的「已单独设置」沿用本群的墨色实心样式。
- **计数徽标：** 系统提示灰胶囊，高（22px），12px/700 等宽数字；为 0 时改为透明底加 `line-strong` 描边。添加群面板群栏标题旁的数量用同色胶囊，高（20px）。
- **范围标签：** 推演标题旁的系统提示灰胶囊，12px/600，超过（220px）省略。

### Cards / Containers

- 规则不装进卡片：检查器分组是带顶部细线的段落，规则行只在悬停或聚焦时铺凹陷面底色。
- **速览条：** 三格横排（首枪中弹、禁言、时限），1px `line` 外框、`card` 圆角、格间细线；格内上为 12px 次级说明，下为 Figure 数值。
- **选择卡片：** 添加群面板里的机器人卡片与群卡片，`card` 圆角，选中靠墨色描边表达，见「添加群悬浮面板」。
- **载入卡：** 聊天底上居中的面板色卡片，`bubble` 圆角、浮起投影，含状态文字与「重新读取」。

### Inputs / Fields

- **数值框：** `canvas` 底、`control` 边界、`field` 圆角、高（36px）；数字右对齐、600 字重、等宽，单位以 12px 次级字放在框内右侧，秒数用加宽的输入。随机时上下限之间用「到」连接。
- **分段切换：** 凹陷面轨道（`control` 圆角、3px 内边距），选项 13px/600 次级墨色；按下的选项为面板底、墨色字、气泡轻影加 1px `line` 外环。用于「固定 / 随机」与推演场景切换。
- **开关：** （48×26px）胶囊，台词编辑器内为（42×22px）；关闭为 `control` 轨道，开启为操作蓝；轨道内写「开 / 关」（关闭时用 `switch-off-ink`，开启时用 `accent-ink`），白色滑块（0.18s）滑动。跟随默认时禁用，透明度（0.45）。
- **台词文本框：** `canvas` 底、`line-strong` 边界、`control` 圆角、Message 字号，可纵向拉伸；跟随默认台词时禁用并转为凹陷面底。
- **搜索框：** 凹陷面底、`control` 圆角、高（36px），前置放大镜图标，输入框本身无边框；聚焦时整框加操作蓝 3px 外轮廓。用于范围列表（超过 5 个群时出现）与添加群面板的群栏（240px）。
- **错误：** 无效输入的边框转为 `danger`，下方 12px/600 `danger` 文字通过语义属性关联到字段；聚焦时仍保持操作蓝焦点轮廓。

### Navigation

- **规则范围列表：** 会话列表式的行：头像（42px，群行优先显示真实 QQ 群头像）、名称（700，单行省略）、群号与机器人账号（Caption）、覆盖摘要（Caption，墨色）和计数徽标。行为 `row` 圆角、（10px）内边距；悬停铺凹陷面，当前项铺 `selected`。「默认规则」置顶并使用机器人头像；其下是「单独设置的群」分组标题与数量，超过 5 个群时出现搜索框。
- **添加入口：** 列底部高（40px）的虚线按钮「为某个群单独设置」，打开添加群悬浮面板；处理中禁用。
- 页面没有顶部页签；场景切换和结局选择是推演内部的分段切换与胶囊选项。

### 头像

- 统一的圆形头像，对读屏隐藏，名字始终以文字出现在旁边。底层是六色色盘加白色首字（按群号或 QQ 号固定取色），真实 QQ 头像在其上裁切铺满整圆；图片懒加载、不发送来源页信息，加载失败即移除，露出首字。
- 图片来源：机器人用身份信息里的头像地址，缺省时按 QQ 号取 `q1.qlogo.cn`；群按群号取 `p.qlogo.cn`。宿主 CSP 只对图片和媒体放开 HTTPS，脚本、样式、字体与数据请求仍限同源。
- 用在范围列表的群行与群范围的检查器标题（42px）、添加群面板的机器人卡片（48px）与群卡片（42px），以及面板底部的已选群（36px）。「默认规则」行和推演里的机器人保留转轮线描头像，推演里的示例群友只用首字。

### 推演回放

- 聊天底上的对话列：居中的时间戳（Caption、次级墨色）；群友消息为彩色头像、名字行和白色气泡，命令文字加重；机器人消息为转轮头像、「示例机器人」名字行、机器人标签、「· 某某台词」场景名和淡蓝气泡。@ 提及用 `accent-text`、600 字重。
- 系统提示是居中的灰色胶囊，前置锁形图标（如「老王 被禁言 2分30秒」）；略过的多枪用两侧 `line-strong` 细线夹住的折叠说明。
- 机器人气泡可点击：悬停出现 2px `line-strong` 外环，名字行的「编辑」入口变为 `accent-text`。点开后气泡原位换成编辑器：面板底、气泡轮廓、浮起投影，含本群专用台词开关、文本框、占位符按钮行（凹陷面、`line-strong` 描边、7px 圆角、高 28px，悬停为操作蓝描边与 `accent-text` 字）、字数计数与「完成」；Esc 关闭并把焦点还给编辑入口。
- 头部始终标注「示例，不会发送到群里」，底部说明条列出示例取值。规则有误时对话列换成错误浅底的「推演暂停」提示，附「去修正」链接。
- 超时场景的结局选择条用胶囊选项：按下为墨色描边加凹陷面底，符合当前设置的选项带「当前设置」小胶囊。

### 规则联动高亮

- 检查器每一项都连到推演。数值与口令连到片段：开局口令对应群友发出的命令；弹膛、子弹对应台词里的弹膛数、子弹数和折叠说明里的计数；禁言对应禁言时长；时限对应台词与时间行里的时限。悬停或聚焦该项时，对应片段铺上 `lit`（`mark` 圆角、2px 同色外扩）并加 2px `lit-edge` 下划线（偏移 3px），速览条对应格整格铺 `lit`；离开即撤除，过渡（0.2s）。
- 两个收局开关连到整条消息：「剩余全是实弹时提前结束」对应提前结束台词或「当前设置不提前结束」等提示，「超时后禁言发起人」对应超时结局台词。悬停或聚焦时，整条机器人气泡或系统提示加一圈 2px `lit-edge` 外环。
- 数值改变时，对应数字以 `flash` 闪一次：前 25% 保持峰值，随后在（1.4s）内淡出，缓动 cubic-bezier(0.2, 0.8, 0.2, 1)。
- 减少动态效果时取消所有过渡与动画，闪烁改为同时长的静态 `lit` 标记；强制颜色模式下片段与整条消息的点亮、闪烁都改为 2px Highlight 轮廓（偏移 2px）。

**「改值闪一次」规则。** 数值改变只让对应数字闪一次后回落；连续修改重新开始而不叠加，从不循环。

### 规则检查器与继承

- 顶部是范围标题与次级说明（群范围时标题前有该群头像，说明为群号、机器人账号与连接名，并带「删除本群设置」危险按钮），其下是速览条，再往下是开局口令、装填、惩罚、收局四组。
- 规则行左侧是（44px）行标签，右侧依次为「固定 / 随机」分段切换、数值框（随机时加上限框）、说明与错误。
- 群范围下按规则单元继承：弹膛、子弹、禁言各自整组，时限与两个收局开关各自单独。每个单元顶部一行显示继承标签、单独设置时的默认值说明，以及右对齐的「单独设置 / 恢复默认」链接；跟随默认时数值显示为有效值的次级文字，开关显示当前状态并禁用。
- 时限框旁是（6px）高的中性时限条：`line` 轨道、`ink-2` 填充，按时限占当前上限的比例伸缩。

### 保存区与确认

- 固定在检查器底部：状态行（9px 状态点加粗体文字：设置已同步、有未保存更改、已保存，待应用、正在处理）、改动摘要（最多两行）、处理说明（礼貌播报）、待修正清单（错误浅底，可滚动，条目是跳转到字段的链接），以及操作行：主按钮、重新读取、靠右的恢复默认。
- **确认条：** 恢复默认、有未保存更改时重新读取、删除本群设置三种操作在原位展开错误浅底确认条（`control` 圆角、10px 12px 内边距）：一句后果说明，小号危险按钮在前、小号次按钮「取消」在后。打开时焦点落在确认按钮，Esc 或取消后焦点回到触发按钮。

**「页面内确认」规则。** 会丢弃内容的操作只在原位确认条里确认，不弹原生 confirm 或模态框。选择类任务（如为某个群单独设置）可以使用添加群悬浮面板这类模态面板。

### 添加群悬浮面板

- 「为某个群单独设置」入口（声明会弹出对话框）以原生对话框的模态方式打开面板，`backdrop` 遮罩压暗整页。面板为 `sheet` 圆角、面板底、弹出投影，最大（940×660px）。
- **头部：** Dialog 标题「为某个群单独设置」、13px 次级描述、右上角图标关闭按钮；下方细线。
- **机器人栏：** 聊天底上的白色机器人卡片（`card` 圆角、12px 内边距、气泡轻影、1px 透明描边）：大头像、昵称（700）、QQ 号（13px 等宽）、实例名与连接状态（7px 状态点加文字，如已连接、连接中、鉴权失败、已停用）。悬停为 `line-strong` 描边，选中为墨色描边加右侧对勾；已停用或身份未确认的机器人禁用（透明底、无影、`line` 描边、透明度 0.6）。栏标题旁有「刷新」链接。
- **群栏：** 标题「群」与数量胶囊，右侧搜索框按群名或群号筛选。群卡片为 `card` 圆角、`line` 描边、面板底、10px 12px 内边距，含头像、群名（600）、群号（Caption 等宽）；已单独设置的群带墨色实心「已单独设置」标签。悬停为 `line-strong` 描边加凹陷面底，选中为墨色描边加 `selected` 底，双击直接确认。读取中显示 8 张骨架卡片（`line` 色圆与两条横条，1.2s 呼吸）；尚未选机器人或没有匹配结果时，居中显示次级文字。
- **底部：** 已选群摘要（小头像、群名、「群号 · 机器人 昵称（QQ）」）或提示「选择一个群；双击可直接添加」；达到 500 个群上限时显示错误文字；右侧次按钮「取消」与主按钮「添加这个群」，已设置过的群改为「打开已有设置」。
- **键盘与焦点：** 打开时焦点落在第一个可用机器人（已选过机器人时落在搜索框）；只有一个可用机器人时自动选中。选中机器人、群列表读完后，焦点移到搜索框。Esc 或点击遮罩关闭面板，焦点回到「为某个群单独设置」。
- **动效与强制颜色：** 面板以（0.22s）进场：透明度从 0、上移 10px、缩放 0.985，缓动 cubic-bezier(0.2, 0.8, 0.2, 1)；减少动态效果时取消进场与骨架呼吸。强制颜色模式下面板加 CanvasText 边框，选中的机器人卡片与群卡片加 2px Highlight 轮廓。

## Do's and Don'ts

### Do:

- **Do** 让检查器的每一项都接入规则联动：数值点亮推演里由它决定的数字（`lit` 底加 `lit-edge` 下划线），开关为它决定出现的整条消息加 `lit-edge` 外环；改值闪 `flash` 一次（1.4s），减少动态效果时改为静态标记。
- **Do** 用聊天应用的材质表达推演：白色成员气泡、淡蓝机器人气泡、居中的灰色系统提示条、时间戳与折叠分隔线。
- **Do** 头像先画六色首字圆底，再叠真实 QQ 头像，加载失败时退回首字；头像对读屏隐藏，名字始终以文字出现。
- **Do** 在群范围下逐个规则单元显示继承标签、默认值和「单独设置 / 恢复默认」；跟随默认的数值以有效值文字呈现。
- **Do** 同步维护亮暗两套 token：显式 `data-theme="dark"` 与跟随系统的暗色回退定义同一组变量。
- **Do** 保持粗指针下（44px）的最小交互尺寸（范围行 56px），焦点始终是操作蓝 3px 外轮廓。

### Don't:

- **Don't** 把蓝色用在状态点、计数、标签、时限条或装饰上；淡蓝只属于机器人气泡。
- **Don't** 把 `lit`、`lit-edge` 或 `flash` 用于警告、选中、搜索命中或装饰。
- **Don't** 回到「左说明右控件的灰色表单行 + 顶部页签」的设置页排法。
- **Don't** 用原生 confirm、alert 或模态框确认丢弃类操作。
- **Don't** 只用颜色表达状态：状态点要配文字，继承关系要写成「跟随默认 / 本群」。
- **Don't** 加载外部字体、脚本、样式或 CDN 资源，也不用字符或 emoji 代替图标；外部图片只限经 HTTPS 加载的 QQ 头像。
- **Don't** 在页面内重复宿主标题、插件名称或全局导航。
- **Don't** 让推演失去「示例」标注或示例取值说明。
