# CLIProxyAPI Plus

[English](README.md) | 中文

这是 [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 的 Plus 版本，在原有基础上增加了第三方供应商的支持。

所有的第三方供应商支持都由第三方社区维护者提供，CLIProxyAPI 不提供技术支持。如需取得支持，请与对应的社区维护者联系。

该 Plus 版本的主线功能与主线功能强制同步。

## 与主线版本版本差异

- 新增 GitHub Copilot 支持（OAuth 登录），由[em4go](https://github.com/em4go/CLIProxyAPI/tree/feature/github-copilot-auth)提供
- 新增 Kiro (AWS CodeWhisperer) 支持 (OAuth 登录), 由[fuko2935](https://github.com/fuko2935/CLIProxyAPI/tree/feature/kiro-integration)、[Ravens2121](https://github.com/Ravens2121/CLIProxyAPIPlus/)提供

## 贡献

该项目仅接受第三方供应商支持的 Pull Request。任何非第三方供应商支持的 Pull Request 都将被拒绝。

如果需要提交任何非第三方供应商支持的 Pull Request，请提交到主线版本。

## 谁与我们在一起？

这些项目基于 CLIProxyAPI:

### [vibeproxy](https://github.com/automazeio/vibeproxy)

一个原生 macOS 菜单栏应用，让您可以使用 Claude Code & ChatGPT 订阅服务和 AI 编程工具，无需 API 密钥。

### [Subtitle Translator](https://github.com/VjayC/SRT-Subtitle-Translator-Validator)

一款基于浏览器的 SRT 字幕翻译工具，可通过 CLI 代理 API 使用您的 Gemini 订阅。内置自动验证与错误修正功能，无需 API 密钥。

### [CCS (Claude Code Switch)](https://github.com/kaitranntt/ccs)

CLI 封装器，用于通过 CLIProxyAPI OAuth 即时切换多个 Claude 账户和替代模型（Gemini, Codex, Antigravity），无需 API 密钥。

### [ProxyPal](https://github.com/heyhuynhgiabuu/proxypal)

基于 macOS 平台的原生 CLIProxyAPI GUI：配置供应商、模型映射以及OAuth端点，无需 API 密钥。

### [Quotio](https://github.com/nguyenphutrong/quotio)

原生 macOS 菜单栏应用，统一管理 Claude、Gemini、OpenAI、Qwen 和 Antigravity 订阅，提供实时配额追踪和智能自动故障转移，支持 Claude Code、OpenCode 和 Droid 等 AI 编程工具，无需 API 密钥。

### [CodMate](https://github.com/loocor/CodMate)

原生 macOS SwiftUI 应用，用于管理 CLI AI 会话（Claude Code、Codex、Gemini CLI），提供统一的提供商管理、Git 审查、项目组织、全局搜索和终端集成。集成 CLIProxyAPI 为 Codex、Claude、Gemini、Antigravity 和 Qwen Code 提供统一的 OAuth 认证，支持内置和第三方提供商通过单一代理端点重路由 - OAuth 提供商无需 API 密钥。

### [ProxyPilot](https://github.com/Finesssee/ProxyPilot)

原生 Windows CLIProxyAPI 分支，集成 TUI、系统托盘及多服务商 OAuth 认证，专为 AI 编程工具打造，无需 API 密钥。

### [Claude Proxy VSCode](https://github.com/uzhao/claude-proxy-vscode)

一款 VSCode 扩展，提供了在 VSCode 中快速切换 Claude Code 模型的功能，内置 CLIProxyAPI 作为其后端，支持后台自动启动和关闭。

### [ZeroLimit](https://github.com/0xtbug/zero-limit)

Windows 桌面应用，基于 Tauri + React 构建，用于通过 CLIProxyAPI 监控 AI 编程助手配额。支持跨 Gemini、Claude、OpenAI Codex 和 Antigravity 账户的使用量追踪，提供实时仪表盘、系统托盘集成和一键代理控制，无需 API 密钥。

### [CPA-XXX Panel](https://github.com/ferretgeek/CPA-X)

面向 CLIProxyAPI 的 Web 管理面板，提供健康检查、资源监控、日志查看、自动更新、请求统计与定价展示，支持一键安装与 systemd 服务。

> [!NOTE]
> 如果你开发了基于 CLIProxyAPI 的项目，请提交一个 PR（拉取请求）将其添加到此列表中。

## 更多选择

以下项目是 CLIProxyAPI 的移植版或受其启发：

### [9Router](https://github.com/decolua/9router)

基于 Next.js 的实现，灵感来自 CLIProxyAPI，易于安装使用；自研格式转换（OpenAI/Claude/Gemini/Ollama）、组合系统与自动回退、多账户管理（指数退避）、Next.js Web 控制台，并支持 Cursor、Claude Code、Cline、RooCode 等 CLI 工具，无需 API 密钥。

> [!NOTE]
> 如果你开发了 CLIProxyAPI 的移植或衍生项目，请提交 PR 将其添加到此列表中。

## 许可证

此项目根据 MIT 许可证授权 - 有关详细信息，请参阅 [LICENSE](LICENSE) 文件。
