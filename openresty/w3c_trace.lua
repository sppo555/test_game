local _M = {}

-- 基礎依賴
local string_format = string.format
local string_len = string.len
local string_match = string.match
local uuid = require "resty.jit-uuid"

-- 常量定義
local TRACEPARENT_PATTERN = "^(%d%d)%-([a-f0-9]+)%-([a-f0-9]+)%-(%d%d)$"
local DEFAULT_VERSION = "00"
local DEFAULT_FLAGS = "01"
local TRACE_ID_LENGTH = 32
local PARENT_ID_LENGTH = 16

-- ngx API 本地化
local ngx_log = ngx.log
local ngx_ERR = ngx.ERR
local ngx_DEBUG = ngx.DEBUG
local ngx_INFO = ngx.INFO

-- 基礎配置
local _config = nil

local function init_config()
    if _config then
        return _config
    end

    _config = {
        enable = (ngx.var.enable_trace_context or "on") == "on",
        service_name = ngx.var.service_name or "nginx-service"
    }

    return _config
end

-- 生成指定長度的隨機十六進制字串
local function generate_random_hex(len)
    return string.sub(string.gsub(uuid(), "-", ""), 1, len)
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
    
    -- 版本檢查
    if version ~= DEFAULT_VERSION then 
        return false 
    end
    
    -- trace_id 不能全為 0
    if trace_id:match("^0+$") then
        return false
    end
    
    -- parent_id 不能全為 0
    if parent_id:match("^0+$") then
        return false
    end
    
    -- 長度檢查
    if string_len(trace_id) ~= TRACE_ID_LENGTH then 
        return false 
    end
    if string_len(parent_id) ~= PARENT_ID_LENGTH then 
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
        return nil, "Invalid traceparent format"
    end
    
    local version, trace_id, parent_id, flags = string_match(traceparent, TRACEPARENT_PATTERN)
    
    return {
        version = version,
        trace_id = trace_id,
        parent_id = parent_id,
        flags = flags
    }
end

-- 生成新的 trace context
function _M.generate_trace_context()
    local trace_id = generate_random_hex(TRACE_ID_LENGTH)
    local parent_id = generate_random_hex(PARENT_ID_LENGTH)
    
    -- 確保生成的 ID 不全為 0
    if trace_id:match("^0+$") then
        trace_id = "1" .. string.sub(trace_id, 2)
    end
    if parent_id:match("^0+$") then
        parent_id = "1" .. string.sub(parent_id, 2)
    end
    
    return {
        version = DEFAULT_VERSION,
        trace_id = trace_id,
        parent_id = parent_id,
        flags = DEFAULT_FLAGS
    }
end

-- 獲取或創建 trace context
function _M.get_trace_context()
    -- 檢查是否啟用
    if not init_config().enable then
        return nil, "Trace context is disabled"
    end

    -- 檢查請求頭中是否已有 traceparent
    local headers = ngx.req.get_headers()
    local traceparent = headers["traceparent"]
    if traceparent then
        ngx_log(ngx_INFO, "Found existing traceparent: ", traceparent)
        local parsed, err = _M.parse_traceparent(traceparent)
        if parsed then
            -- 生成新的 parent_id，保持 trace_id
            parsed.parent_id = generate_random_hex(PARENT_ID_LENGTH)
            if parsed.parent_id:match("^0+$") then
                parsed.parent_id = "1" .. string.sub(parsed.parent_id, 2)
            end
            return parsed
        end
        ngx_log(ngx_DEBUG, "Invalid traceparent in request: ", err or "unknown error")
    end

    -- 如果沒有有效的 traceparent，創建新的
    local ctx = _M.generate_trace_context()
    if not ctx then
        return nil, "Failed to generate trace context"
    end
    ngx_log(ngx_INFO, "Generated new trace context: ",
        "version=", ctx.version,
        " trace_id=", ctx.trace_id,
        " parent_id=", ctx.parent_id,
        " flags=", ctx.flags)

    return ctx
end

-- 設置響應頭
function _M.set_response_headers(ctx)
    if not ctx or not ctx.trace_id or not ctx.parent_id then
        return
    end

    -- 設置 traceparent 響應頭
    local traceparent = string_format("%s-%s-%s-%s", 
        ctx.version or DEFAULT_VERSION, 
        ctx.trace_id, 
        ctx.parent_id, 
        ctx.flags or DEFAULT_FLAGS)

    if is_valid_traceparent(traceparent) then
        ngx.header["traceparent"] = traceparent
        ngx_log(ngx_INFO, "Set response traceparent: ", traceparent)
    else
        ngx_log(ngx_ERR, "Generated invalid traceparent: ", traceparent)
    end
end

-- 獲取當前配置
function _M.get_config()
    return {
        enable = init_config().enable,
        service_name = init_config().service_name
    }
end

return _M