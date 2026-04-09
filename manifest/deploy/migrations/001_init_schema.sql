-- ClawUnit 初始数据库结构
-- PostgreSQL

-- 实例表：每个 OpenClaw 实例对应一条记录
CREATE TABLE instances (
    id              BIGSERIAL       PRIMARY KEY,
    owner_upn       TEXT            NOT NULL,
    name            TEXT            NOT NULL,
    description     TEXT            NOT NULL DEFAULT '',
    status          TEXT            NOT NULL DEFAULT 'creating',
    image           TEXT            NOT NULL,
    storage_class   TEXT            NOT NULL DEFAULT 'standard',
    mount_path      TEXT            NOT NULL DEFAULT '/home/user/data',
    pod_name        TEXT            NOT NULL DEFAULT '',
    pod_namespace   TEXT            NOT NULL DEFAULT '',
    pod_ip          TEXT            NOT NULL DEFAULT '',
    access_token    TEXT            NOT NULL DEFAULT '',
    api_key_hash    TEXT            NOT NULL DEFAULT '',
    api_mode        TEXT            NOT NULL DEFAULT 'manual',
    model_id        TEXT            NOT NULL DEFAULT '',
    provider        TEXT            NOT NULL DEFAULT 'openrouter',
    api_key         TEXT            NOT NULL DEFAULT '',
    base_url        TEXT            NOT NULL DEFAULT '',
    gateway_token   TEXT            NOT NULL DEFAULT '',
    cpu_cores       TEXT            NOT NULL,
    memory_gb       TEXT            NOT NULL,
    disk_gb         TEXT            NOT NULL,
    gpu_count       INT             NOT NULL DEFAULT 0,
    container_port  INT             NOT NULL DEFAULT 3001,
    gpu_enabled     BOOLEAN         NOT NULL DEFAULT FALSE,
    allow_private_network BOOLEAN  NOT NULL DEFAULT FALSE,
    pvc_name        TEXT            NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    stopped_at      TIMESTAMPTZ,
    UNIQUE (owner_upn, name)
);

CREATE INDEX idx_instances_owner ON instances (owner_upn);
CREATE INDEX idx_instances_status ON instances (status);

-- 用户配额表：每个用户的资源上限
CREATE TABLE user_quotas (
    id              BIGSERIAL       PRIMARY KEY,
    owner_upn       TEXT            NOT NULL UNIQUE,
    max_instances   INT             NOT NULL DEFAULT 3,
    max_cpu_cores   INT             NOT NULL DEFAULT 8,
    max_memory_gb   INT             NOT NULL DEFAULT 16,
    max_storage_gb  INT             NOT NULL DEFAULT 50,
    max_gpu_count   INT             NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 技能表：系统级和用户级技能的元数据
CREATE TABLE skills (
    id              BIGSERIAL       PRIMARY KEY,
    name            TEXT            NOT NULL,
    description     TEXT            NOT NULL DEFAULT '',
    scope           TEXT            NOT NULL DEFAULT 'system',
    owner_upn       TEXT,
    pvc_path        TEXT            NOT NULL,
    version         TEXT            NOT NULL DEFAULT '',
    enabled         BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE (scope, owner_upn, name)
);

CREATE INDEX idx_skills_scope ON skills (scope);
CREATE INDEX idx_skills_owner ON skills (owner_upn);

-- API Key 配置记录：ClawUnit 为实例在 UniAuth 中创建的 API Key
CREATE TABLE api_key_provisions (
    id              BIGSERIAL       PRIMARY KEY,
    instance_id     BIGINT          NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    owner_upn       TEXT            NOT NULL,
    api_key_hash    TEXT            NOT NULL,
    quota_pool      TEXT            NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX idx_akp_instance ON api_key_provisions (instance_id);

-- 审计事件表：记录关键操作
CREATE TABLE audit_events (
    id              BIGSERIAL       PRIMARY KEY,
    actor_upn       TEXT            NOT NULL,
    action          TEXT            NOT NULL,
    resource_type   TEXT            NOT NULL,
    resource_id     BIGINT,
    details         JSONB,
    ip_address      TEXT            NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor ON audit_events (actor_upn);
CREATE INDEX idx_audit_action ON audit_events (action);
CREATE INDEX idx_audit_created ON audit_events (created_at);
