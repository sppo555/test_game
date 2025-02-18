local _M = {}

-- 基礎依賴
local string_format = string.format
local string_len = string.len
local string_lower = string.lower
local string_match = string.match
local uuid = require "resty.jit-uuid"
local http = require "resty.http"
local cjson = require "cjson.safe"

-- 常量定義
local TRACEPARENT_PATTERN = "^(%d%d)%-([a-f0-9]+)%-([a-f0-9]+)%-(%d%d)$"
local DEFAULT_VERSION = "00"
local DEFAULT_FLAGS = "01"

-- ngx API 本地化
local ngx_log = ngx.log
local ngx_ERR = ngx.ERR
local ngx_DEBUG = ngx.DEBUG
local ngx_INFO = ngx.INFO
local ngx_WARN = ngx.WARN

-- 基礎配置
local _config = {
    enable = (ngx.var.enable_trace_context or "on") == "on",
    timeout = tonumber(ngx.var.trace_timeout) or 1000,
    backend_url = ngx.var.trace_backend_url or ngx.var.upstream_url or "http://127.0.0.1:8080",
    service_name = ngx.var.service_name or "nginx-service",
    log_level = ngx.var.log_level or "INFO",
    retry_times = tonumber(ngx.var.trace_retry_times) or 3,
    retry_delay = tonumber(ngx.var.trace_retry_delay) or 1,
    batch_size = tonumber(ngx.var.trace_batch_size) or 100,
    flush_interval = tonumber(ngx.var.trace_flush_interval) or 5,
    -- 預設關閉監控指標和資料脫敏
    enable_metrics = (ngx.var.enable_trace_metrics or "off") == "on",
    enable_sanitization = (ngx.var.enable_trace_sanitization or "off") == "on",
    -- 可配置脫敏規則
    sanitization_rules = {
        auth = (ngx.var.sanitize_auth or "on") == "on",
        cookie = (ngx.var.sanitize_cookie or "on") == "on",
        token = (ngx.var.sanitize_token or "on") == "on",
        custom_fields = ngx.var.sanitize_custom_fields or "" -- 自定義脫敏欄位，以逗號分隔
    }
}

-- 解析自定義脫敏欄位
local _custom_sanitize_fields = {}
if _config.sanitization_rules.custom_fields ~= "" then
    for field in string.gmatch(_config.sanitization_rules.custom_fields, "([^,]+)") do
        _custom_sanitize_fields[string_lower(field:match("^%s*(.-)%s*$"))] = true
    end
end

-- 批次處理佇列
local _batch_queue = {}
local _batch_timer_running = false

-- 監控指標（只有在啟用時才會被更新）
local _metrics = {
    requests_total = 0,
    errors_total = 0,
    backend_errors = 0,
    processing_time_total = 0,
    batch_size_current = 0,
    traces_sent_total = 0
}

-- 生成指定長度的隨機十六進制字串
local function generate_random_hex(len)
    return string.sub(uuid(), 1, len * 2)
end

-- 驗證 traceparent 格式
local function is_valid_traceparent(traceparent)
    if not traceparent then 
        return false 
    end
    
    local version, trace_id, parent_id, flags = string_match(traceparent, TRACEPARENT_PATTERN)
    
    if not (version and trace_id and parent_id and flags) then
        return false
    end
    
    if version ~= DEFAULT_VERSION then 
        return false 
    end
    if string_len(trace_id) ~= 32 then 
        return false 
    end
    if string_len(parent_id) ~= 16 then 
        return false 
    end
    if string_len(flags) ~= 2 then 
        return false 
    end
    
    return true
end

-- 解析 traceparent
function _M.parse_traceparent(traceparent)
    if not is_valid_traceparent(traceparent) then
        return nil
    end
    
    local version, trace_id, parent_id, flags = string_match(traceparent, TRACEPARENT_PATTERN)
    
    return {
        version = version,
        trace_id = trace_id,
        parent_id = parent_id,
        flags = flags
    }
end

-- 生成新的 traceparent
function _M.generate_trace_id()
    local trace_id = generate_random_hex(16)  -- 32字符
    local parent_id = generate_random_hex(8)  -- 16字符
    return string_format("%s-%s-%s-%s", DEFAULT_VERSION, trace_id, parent_id, DEFAULT_FLAGS)
end

-- 資料脫敏（根據配置決定是否執行）
local function sanitize_headers(headers)
    -- 如果未啟用脫敏，直接返回原始 headers
    if not _config.enable_sanitization then
        return headers
    end
    
    local sanitized = {}
    for k, v in pairs(headers) do
        local key_lower = string_lower(k)
        local should_sanitize = false
        
        -- 檢查是否需要脫敏
        if (_config.sanitization_rules.auth and string_match(key_lower, "auth")) or
           (_config.sanitization_rules.cookie and string_match(key_lower, "cookie")) or
           (_config.sanitization_rules.token and string_match(key_lower, "token")) or
           _custom_sanitize_fields[key_lower] then
            should_sanitize = true
        end
        
        if should_sanitize then
            sanitized[k] = "[REDACTED]"
        else
            sanitized[k] = v
        end
    end
    return sanitized
end

-- 驗證追蹤資料
local function validate_trace_data(trace_data)
    if not trace_data then
        return false, "Trace data is nil"
    end
    
    if not trace_data.trace_id or string_len(trace_data.trace_id) ~= 32 then
        return false, "Invalid trace_id"
    end
    
    if not trace_data.parent_id or string_len(trace_data.parent_id) ~= 16 then
        return false, "Invalid parent_id"
    end
    
    return true
end

-- 更新監控指標（根據配置決定是否執行）
local function update_metrics(start_time, success, batch_size)
    if not _config.enable_metrics then
        return
    end
    
    _metrics.requests_total = _metrics.requests_total + 1
    _metrics.processing_time_total = _metrics.processing_time_total + (ngx.now() - start_time)
    _metrics.batch_size_current = #_batch_queue
    
    if batch_size then
        _metrics.traces_sent_total = _metrics.traces_sent_total + batch_size
    end
    
    if not success then
        _metrics.errors_total = _metrics.errors_total + 1
    end
end

-- [其餘代碼保持不變...]

-- 獲取監控指標
function _M.get_metrics()
    if not _config.enable_metrics then
        return nil, "Metrics collection is disabled"
    end
    return _metrics
end

-- 獲取當前配置
function _M.get_config()
    return {
        enable = _config.enable,
        enable_metrics = _config.enable_metrics,
        enable_sanitization = _config.enable_sanitization,
        sanitization_rules = _config.sanitization_rules,
        batch_size = _config.batch_size,
        flush_interval = _config.flush_interval,
        service_name = _config.service_name
    }
end

return _M