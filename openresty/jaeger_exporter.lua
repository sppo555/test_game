local _M = {}

local http = require "resty.http"
local cjson = require "cjson.safe"  -- OpenResty 內建的 cjson
local uuid = require "resty.jit-uuid"

-- 使用環境變數中的 Zipkin 端點
local ZIPKIN_ENDPOINT = os.getenv("ZIPKIN_ENDPOINT")
if not ZIPKIN_ENDPOINT then
    ngx.log(ngx.ERR, "ZIPKIN_ENDPOINT environment variable is required")
    ZIPKIN_ENDPOINT = "http://jaeger:9411/api/v2/spans"  -- 使用默認值
end

-- 禁用科學記數法
cjson.encode_number_precision(16)

-- 生成指定長度的隨機十六進制字串
local function generate_random_hex(len)
    return string.sub(string.gsub(uuid(), "-", ""), 1, len)
end

-- 將時間轉換為微秒的字符串格式
local function to_microseconds(seconds)
    -- 先轉換為整數微秒
    local microseconds = math.floor(seconds * 1000000)
    -- 轉換為字符串以避免科學記數法
    return tostring(microseconds)
end

-- 創建 Zipkin span
local function create_span(trace_id, parent_id, operation_name, start_time, duration, tags)
    -- 生成新的 span ID
    local span_id = generate_random_hex(16)
    -- 確保 span ID 不全為 0
    if span_id:match("^0+$") then
        span_id = "1" .. string.sub(span_id, 2)
    end

    -- 從 W3C trace context 格式轉換為 Zipkin 格式
    -- 移除版本前綴 "00-" 和標誌後綴 "-01"
    local zipkin_trace_id = string.match(trace_id, "^00%-([a-f0-9]+)%-") or trace_id
    local zipkin_parent_id = string.match(parent_id, "^([a-f0-9]+)") or parent_id

    ngx.log(ngx.INFO, "Zipkin Trace ID: ", zipkin_trace_id)
    ngx.log(ngx.INFO, "Zipkin Span ID: ", span_id)
    ngx.log(ngx.INFO, "Zipkin Parent ID: ", zipkin_parent_id)

    -- 轉換標籤格式
    local zipkin_tags = {}
    for _, tag in ipairs(tags or {}) do
        zipkin_tags[tag.key] = tostring(tag.vStr or tag.vInt64 or "")
    end

    -- 添加 span.kind 標籤
    zipkin_tags["span.kind"] = "SERVER"

    -- 轉換時間為字符串格式的微秒
    local timestamp_str = to_microseconds(start_time)
    local duration_str = to_microseconds(duration)

    local span = {
        id = span_id,
        traceId = zipkin_trace_id,
        parentId = zipkin_parent_id,
        name = operation_name,
        timestamp = tonumber(timestamp_str),  -- 轉回數字，但已經是正確格式
        duration = tonumber(duration_str),    -- 轉回數字，但已經是正確格式
        tags = zipkin_tags,
        localEndpoint = {
            serviceName = "openresty",
            ipv4 = ngx.var.server_addr
        }
    }
    return span
end

-- 在計時器中發送 span 數據
local function send_span(premature, span_data)
    if premature then
        return
    end

    -- 打印請求數據和端點
    ngx.log(ngx.INFO, "Using Zipkin endpoint: ", ZIPKIN_ENDPOINT)
    ngx.log(ngx.INFO, "Sending span data to Zipkin: ", span_data)

    local httpc = http.new()
    local res, err = httpc:request_uri(ZIPKIN_ENDPOINT, {
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

    if res.status ~= 202 and res.status ~= 200 then
        ngx.log(ngx.ERR, "Zipkin returned unexpected status: ", res.status)
        ngx.log(ngx.ERR, "Response body: ", res.body)
    else
        ngx.log(ngx.INFO, "Successfully sent span to Zipkin")
    end
end

-- 導出 span 到 Zipkin
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

    -- Zipkin 需要一個數組
    local spans = {span}

    -- 將 spans 轉換為 JSON
    local span_data = cjson.encode(spans)
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
