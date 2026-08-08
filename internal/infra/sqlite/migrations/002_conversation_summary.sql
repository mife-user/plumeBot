-- =============================================================================
-- Migration: 002_conversation_summary
-- 描述:      摘要长程归档表（P3-003 扩展）。
--           窗口压缩产生的摘要，离开内存热链（被二级融合覆盖 / FIFO 淘汰）时落库，
--           供重启后回灌热链底，让 AI 保留"很久以前"的浓缩历史。
-- 创建时间:  2026-08
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 9. conversation_summary — 摘要归档（1 会话 N 条）
-- 作用:   存储被淘汰/被融合覆盖的窗口摘要（长程记忆）。
-- 约束:   (chat_id, seq) 唯一 —— seq 为会话内递增序号，upsert 幂等，
--         回灌的旧摘要再次被覆盖时重复落库不产生冗余。
-- 索引:   idx_conversation_summary_chat — 按会话 + seq 排序读取最新 N 条回灌。
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS conversation_summary (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id    TEXT    NOT NULL,                -- 会话键：群聊=GroupID，私聊="private:"+UserID
    seq        INTEGER NOT NULL,                -- 会话内递增序号（排序 + 幂等键）
    text       TEXT    NOT NULL DEFAULT '',     -- 摘要文本
    keywords   TEXT    NOT NULL DEFAULT '[]',   -- 关键词标签 (JSON array)
    decisions  TEXT    NOT NULL DEFAULT '[]',   -- 关键决定 (JSON array)
    created_at INTEGER NOT NULL DEFAULT 0,      -- 生成时间（Unix 秒）
    UNIQUE(chat_id, seq)
);

CREATE INDEX IF NOT EXISTS idx_conversation_summary_chat
    ON conversation_summary(chat_id, seq);
