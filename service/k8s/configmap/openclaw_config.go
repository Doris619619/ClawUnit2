package configmap

import (
	"fmt"
	"strings"
)

// OpenClawConfigOptions 是生成 openclaw.json 时可调的字段。
//
// ToolProfile 决定 OpenClaw 工具集的权限级别（full / coding /
// messaging / minimal），空字符串视为 "full"。SearchProvider 和
// SearchApiKey 必须同时提供才会生效。SystemPrompt 不为空时会被注入
// 默认 agent 的 instructions 字段。
//
// AllowPrivateNetwork 字段当前未在生成的 JSON 中真正使用，但配置
// 已经透传到下层 —— 详见 networkpolicy 包文档。
type OpenClawConfigOptions struct {
	ModelID             string
	ToolProfile         string // full/coding/messaging/minimal
	SearchProvider      string // brave/perplexity/duckduckgo 等
	SearchApiKey        string // 搜索引擎 API Key
	SystemPrompt        string // 自定义系统提示词
	AllowPrivateNetwork bool   // 允许浏览器/fetch 访问私有网络
}

// OpenClawConfig 把 OpenClawConfigOptions 渲染成 openclaw.json 文本。
//
// 输出包含 gateway / agents / tools / browser / plugins / models 全部
// 段落。Gateway token 用 ${OPENCLAW_GATEWAY_TOKEN} 占位，由 Pod
// 启动时的环境变量替换；同样 LLM 的 ${CUSTOM_API_KEY} 和
// ${CUSTOM_BASE_URL} 也来自环境变量。
//
// 这些环境变量由 lifecycle.Create 根据 ApiMode 注入，详见 Create()。
func OpenClawConfig(opts OpenClawConfigOptions) string {
	toolProfile := opts.ToolProfile
	if toolProfile == "" {
		toolProfile = "full"
	}

	// 搜索引擎配置
	searchBlock := ""
	if opts.SearchProvider != "" && opts.SearchApiKey != "" {
		searchBlock = fmt.Sprintf(`,
    "%s": {
      "config": {
        "webSearch": {
          "apiKey": "%s"
        }
      }
    }`, opts.SearchProvider, opts.SearchApiKey)
	}

	// 系统提示词
	agentsBlock := ""

	if opts.SystemPrompt != "" {
		// 转义 JSON 字符串中的特殊字符
		escaped := opts.SystemPrompt
		escaped = strings.ReplaceAll(escaped, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		agentsBlock = fmt.Sprintf(`,
    "list": [{
      "id": "default",
      "name": "OpenClaw Assistant",
      "workspace": "/home/node/.openclaw/workspace",
      "instructions": "%s"
    }]`, escaped)
	}

	return fmt.Sprintf(`{
  "gateway": {
    "mode": "local",
    "port": 18789,
    "bind": "lan",
    "auth": { "mode": "token", "token": "${OPENCLAW_GATEWAY_TOKEN}" },
    "http": { "endpoints": { "chatCompletions": { "enabled": true } } },
    "controlUi": { "dangerouslyDisableDeviceAuth": true, "allowInsecureAuth": true }
  },
  "agents": {
    "defaults": {
      "model": { "primary": "sglang/%s" },
      "workspace": "/home/node/.openclaw/workspace",
      "sandbox": { "mode": "off" }
    }%s
  },
  "tools": {
    "profile": "%s",
    "exec": { "security": "full" },
    "fs": { "workspaceOnly": false }
  },
  "browser": {
    "enabled": true,
    "executablePath": "/home/node/.cache/ms-playwright/chromium-1208/chrome-linux64/chrome",
    "defaultProfile": "openclaw",
    "headless": true,
    "noSandbox": true,
    "color": "#FF4500",
    "profiles": {
      "openclaw": { "cdpPort": 18800, "color": "#FF4500" }
    },
    "ssrfPolicy": {
      "dangerouslyAllowPrivateNetwork": true
    }
  },
  "plugins": {
    "entries": {%s}
  },
  "cron": { "enabled": false },
  "models": {
    "providers": {
      "sglang": {
        "baseUrl": "${CUSTOM_BASE_URL}/v1",
        "apiKey": "${CUSTOM_API_KEY}",
        "api": "openai-completions",
        "models": [{
          "id": "%s",
          "name": "%s",
          "reasoning": true,
          "contextWindow": 200000,
          "maxTokens": 8192
        }]
      }
    }
  }
}`, opts.ModelID, agentsBlock, toolProfile, searchBlock, opts.ModelID, opts.ModelID)
}
