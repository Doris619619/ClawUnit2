// Package v1 是技能管理 API 的 request/response 定义。
//
// 路由列表：
//
//	ListSystemReq    GET  /api/skills/v1/system/list   列出管理员预置的系统技能
//	ListUserReq      GET  /api/skills/v1/user/list     列出当前用户的自定义技能
//	UploadReq        POST /api/skills/v1/user/upload   上传技能到用户 PVC
//	DeleteSkillReq   POST /api/skills/v1/user/delete   删除用户技能
//
// # 技能的物理位置
//
// 系统技能和用户技能挂载到 OpenClaw Pod 的不同路径：
//
//   - 系统技能：来自 SystemSkillsPVC（管理员预创建，跨用户共享，ReadOnly），
//     挂载到 /skills/system。
//   - 用户技能：来自用户数据 PVC 的 skills/ 子目录，挂载到 /home/node/.openclaw/skills，
//     可读可写。
//
// Upload 接口通过 multipart/form-data 上传文件，并把元数据写入 skills 表。
package v1
