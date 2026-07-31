-- =============================================================================
-- Migration: 001_initial_schema
-- 描述:      PlumeBot 初始数据库表结构，共 8 张业务表 + 1 个索引。
-- 创建时间:  2026-07
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. messages — 聊天消息（全量持久化）
-- 作用:   长期存储所有群聊/私聊消息，供记忆检索和上下文补全。
-- 索引:   idx_messages_group_time — 按群 + 时间排序查询最近消息。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS messages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id  TEXT    NOT NULL,                -- OneBot 消息唯一 ID
    group_id    TEXT    NOT NULL DEFAULT '',     -- 群 ID，私聊时为空
    user_id     TEXT    NOT NULL,                -- 发送者 QQ 号
    content     TEXT    NOT NULL DEFAULT '',     -- 消息文本
    timestamp   INTEGER NOT NULL DEFAULT 0,     -- Unix 时间戳（秒）
    message_type TEXT   NOT NULL DEFAULT 'group' -- group / private
);

CREATE INDEX IF NOT EXISTS idx_messages_group_time
    ON messages(group_id, timestamp);

-- -----------------------------------------------------------------------------
-- 2. group_profile — 群聊画像（每个群一条）
-- 作用:   群文化特征、主流话题、活跃时段、群规、氛围标签。
-- 注意:   黑话词典走 group_jargon 表，不在此处冗余。
-- 召回:   group_id 精确查询。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS group_profile (
    group_id     TEXT PRIMARY KEY,               -- 群 ID
    culture      TEXT NOT NULL DEFAULT '',       -- 群文化特征描述
    topics       TEXT NOT NULL DEFAULT '[]',     -- 主流话题标签 (JSON array)
    active_hours TEXT NOT NULL DEFAULT '',       -- 活跃时段描述
    rules        TEXT NOT NULL DEFAULT '[]',     -- 群规 (JSON array)
    atmosphere   TEXT NOT NULL DEFAULT '[]'      -- 氛围标签 (JSON array)
);

-- -----------------------------------------------------------------------------
-- 3. group_jargon — 群黑话词典
-- 作用:   存储群内特有词汇/梗，供 Agent 理解上下文时参考。
-- 约束:   (group_id, jargon) 唯一。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS group_jargon (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT    NOT NULL,                   -- 群 ID
    jargon   TEXT    NOT NULL,                   -- 黑话/梗文本
    UNIQUE(group_id, jargon)
);

-- -----------------------------------------------------------------------------
-- 4. member_profile — 群聊个人画像
-- 作用:   按 (group_id, user_id) 唯一，记录某人在某个群的表现。
-- 字段:   活跃度、亲密度。
-- 注意:   事实记忆走 member_facts 表，不在此处冗余。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS member_profile (
    group_id  TEXT  NOT NULL,                    -- 群 ID
    user_id   TEXT  NOT NULL,                    -- 用户 QQ 号
    activity  REAL  NOT NULL DEFAULT 0,          -- 活跃度 (0~1)
    intimacy  REAL  NOT NULL DEFAULT 0,          -- 与 bot 亲密度 (0~1)
    PRIMARY KEY (group_id, user_id)
);

-- -----------------------------------------------------------------------------
-- 5. member_facts — 个人事实记忆
-- 作用:   存储对某个用户的零散事实（如"喜欢猫"、"是程序员"）。
-- 约束:   (group_id, user_id, fact) 唯一。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS member_facts (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT    NOT NULL,                   -- 群 ID
    user_id  TEXT    NOT NULL,                   -- 用户 QQ 号
    fact     TEXT    NOT NULL,                   -- 事实描述
    UNIQUE(group_id, user_id, fact)
);

-- -----------------------------------------------------------------------------
-- 6. persona — 人格模板
-- 作用:   定义 bot 回复风格。groupid=0 为全局默认人格，
--         groupid!=0 为群专属人格。extend 字段实现继承链。
-- 种子:   启动时自动插入一条 (userid=0, groupid=0, extend=0, traits='[]')。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS persona (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    userid  INTEGER NOT NULL DEFAULT 0,          -- 所属用户 ID
    groupid INTEGER NOT NULL DEFAULT 0,          -- 0=全局，其他=群 ID
    extend  INTEGER NOT NULL DEFAULT 0,          -- 继承父级 Persona.ID
    traits  TEXT    NOT NULL DEFAULT '[]'        -- 性格标签 (JSON array)
);

-- -----------------------------------------------------------------------------
-- 7. bot_state — Bot 群内状态
-- 作用:   每个群一条，存储 bot 在本群的运行时状态
--         （精力值、连续回复计数、冷却时间等），JSON blob。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bot_state (
    group_id TEXT PRIMARY KEY,                   -- 群 ID
    state    TEXT NOT NULL DEFAULT '{}'          -- 状态 JSON
);

-- -----------------------------------------------------------------------------
-- 8. plugin_config — 插件配置
-- 作用:   按 (group_id, plugin_name) 唯一，存储某插件在某群的配置。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS plugin_config (
    group_id    TEXT NOT NULL,                   -- 群 ID
    plugin_name TEXT NOT NULL,                   -- 插件名称
    config      TEXT NOT NULL DEFAULT '{}',      -- 配置 JSON
    PRIMARY KEY (group_id, plugin_name)
);
