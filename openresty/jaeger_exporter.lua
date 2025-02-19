local _M = {}

local http = require "resty.http"
local cjson = require "cjson.safe"  -- OpenResty 內建的 cjson
local uuid = require "resty.jit-uuid"

local JAEGER_ENDPOINT = os.getenv("JAEGER_ENDPOINT") or "http://jaeger:14268/api/traces"

-- 生成指定長度的隨機十六進制字串
local function generate_random_hex(len)
    return string.sub(string.gsub(uuid(), "-", ""), 1, len)
end

-- 創建 Jaeger Thrift span
local function create_span(trace_id, parent_id, operation_name, start_time, duration, tags)
    -- 生成新的 span ID
    local span_id = generate_random_hex(16)
    -- 確保 span ID 不全為 0
    if span_id:match("^0+$") then
        span_id = "1" .. string.sub(span_id, 2)
    end

    local span = {
        traceIdLow = tonumber(string.sub(trace_id, 17, 32), 16),
        traceIdHigh = tonumber(string.sub(trace_id, 1, 16), 16),
        spanId = tonumber(span_id, 16),
        parentSpanId = parent_id and tonumber(parent_id, 16) or 0,
        operationName = operation_name,
        startTime = start_time * 1000000, -- convert to microseconds
        duration = duration * 1000000,    -- convert to microseconds
        tags = tags or {}
    }
    return span
end

-- 在計時器中發送 span 數據
local function send_span(premature, span_data)
    if premature then
        return
    end

    local httpc = http.new()
    local res, err = httpc:request_uri(JAEGER_ENDPOINT, {
        method = "POST",
        body = span_data,
        headers = {
            ["Content-Type"] = "application/json"
        }
    })

    if not res then
        ngx.log(ngx.ERR, "Failed to export span: ", err)
        return
    end

    if res.status ~= 202 then
        ngx.log(ngx.ERR, "Jaeger returned unexpected status: ", res.status)
    end
end

-- 導出 span 到 Jaeger
function _M.export_span(trace_ctx, operation_name, start_time, duration, tags)
    if not trace_ctx or not trace_ctx.trace_id or not trace_ctx.parent_id then
        return nil, "Invalid trace context"
    end

    local span = create_span(
        trace_ctx.trace_id,
        trace_ctx.parent_id,
        operation_name,
        start_time,
        duration,
        tags
    )

    -- 創建 Jaeger 格式的請求體
    local batch = {
        process = {
            serviceName = "openresty",
            tags = {}
        },
        spans = {span}
    }

    -- 將 batch 轉換為 JSON
    local span_data = cjson.encode(batch)
    if not span_data then
        return nil, "Failed to encode span data"
    end

    -- 使用計時器發送 span 數據
    local ok, err = ngx.timer.at(0, send_span, span_data)
    if not ok then
        ngx.log(ngx.ERR, "Failed to create timer: ", err)
        return nil, err
    end

    return true
end

return _M
