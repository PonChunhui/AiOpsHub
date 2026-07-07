-- Migration: Add mcp_server_ids field to agents table
-- Purpose: Allow agents to specify which MCP Servers they want to use
-- Date: 2026-07-06

ALTER TABLE agents ADD COLUMN IF NOT EXISTS mcp_server_ids TEXT DEFAULT '';

COMMENT ON COLUMN agents.mcp_server_ids IS 'MCP Server ID list in JSON array format, e.g. ["server-id-1", "server-id-2"]';

-- For existing agents, set mcp_server_ids to empty array
UPDATE agents SET mcp_server_ids = '[]' WHERE mcp_server_ids IS NULL OR mcp_server_ids = '';