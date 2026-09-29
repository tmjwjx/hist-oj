-- All keys share one Redis Cluster slot. TIME is server-owned; no client clock or rank is trusted.
redis.replicate_commands()
local time = redis.call('TIME')
local now = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)
local waiting, seen, reserved, loading = KEYS[1], KEYS[2], KEYS[3], KEYS[4]
local departing, closed, sequence, revision = KEYS[5], KEYS[6], KEYS[7], KEYS[8]
local warm = KEYS[9]
local operation, id, capacity = ARGV[1], ARGV[2], tonumber(ARGV[3])
local claimed, completed = 0, 0

redis.call('ZREMRANGEBYSCORE', warm, '-inf', now)
if operation == 'reuse' then
    -- A completed browser may reuse its cache without acquiring or extending a slot.
    return {redis.call('ZSCORE', warm, id) and 1 or 0}
end

local function close(ticket)
    redis.call('ZREM', waiting, ticket)
    redis.call('ZREM', seen, ticket)
    redis.call('ZREM', reserved, ticket)
    redis.call('ZREM', loading, ticket)
    redis.call('ZREM', departing, ticket)
    redis.call('ZADD', closed, now + 120000, ticket)
end

-- Read only expired members, never scan the full queue.
redis.call('ZREMRANGEBYSCORE', closed, '-inf', now)
for _, key in ipairs({seen, reserved, loading, departing}) do
    for _, ticket in ipairs(redis.call('ZRANGEBYSCORE', key, '-inf', now)) do
        close(ticket)
    end
end

if operation == 'status' then
    if not redis.call('ZSCORE', closed, id) then
        redis.call('ZREM', departing, id)
        if redis.call('ZSCORE', reserved, id) then
            -- Receipt of the offer may be lost; its owner's next poll is still a heartbeat.
            redis.call('ZADD', seen, now + 20000, id)
            redis.call('ZADD', reserved, now + 15000, id)
        elseif not redis.call('ZSCORE', loading, id) then
            if not redis.call('ZSCORE', waiting, id) then
                redis.call('ZADD', waiting, redis.call('INCR', sequence), id)
            end
            redis.call('ZADD', seen, now + 20000, id)
        end
    end
elseif operation == 'heartbeat' then
    -- A late heartbeat can never create or resurrect a slot.
    if redis.call('ZSCORE', loading, id) then
        redis.call('ZADD', loading, now + 60000, id)
    end
elseif operation == 'complete' then
    -- Only a live loading lease can issue a new, server-generated completion token.
    if redis.call('ZSCORE', loading, id) then
        redis.call('ZADD', warm, now + 86400000, ARGV[5])
        completed = 1
    end
    close(id)
elseif operation == 'release' or (operation == 'leave' and ARGV[4] == '1') then
    -- Also remember unknown tickets: an in-flight status must not join after cancellation.
    close(id)
elseif operation == 'leave' then
    if redis.call('ZSCORE', waiting, id) or redis.call('ZSCORE', reserved, id) then
        -- Duplicate unload notifications must not extend the refresh grace period.
        redis.call('ZADD', departing, 'NX', now + 5000, id)
    elseif not redis.call('ZSCORE', loading, id) then
        close(id)
    end
end

-- Reserve every free slot in Redis arrival order, irrespective of browser polling order.
local free = capacity - redis.call('ZCARD', reserved) - redis.call('ZCARD', loading)
if free > 0 then
    for _, ticket in ipairs(redis.call('ZRANGE', waiting, 0, free - 1)) do
        redis.call('ZREM', waiting, ticket)
        -- Promotion is not a heartbeat: an absent page must still expire at its old deadline.
        local deadline = tonumber(redis.call('ZSCORE', seen, ticket))
        redis.call('ZADD', reserved, math.min(now + 15000, deadline), ticket)
    end
end

if operation == 'claim' and redis.call('ZSCORE', reserved, id) then
    redis.call('ZREM', reserved, id)
    redis.call('ZREM', seen, id)
    redis.call('ZREM', departing, id)
    redis.call('ZADD', loading, now + 60000, id)
    claimed = 1
end

local state, position = 'closed', 0
if redis.call('ZSCORE', loading, id) then
    state = 'loading'
elseif redis.call('ZSCORE', reserved, id) then
    state = 'reserved'
else
    local rank = redis.call('ZRANK', waiting, id)
    if rank then
        state, position = 'waiting', rank + 1
    end
end
-- Preserve response ordering across Redis restarts as well as concurrent requests.
local nextRevision = math.max(tonumber(time[1]) * 1000000 + tonumber(time[2]),
        tonumber(redis.call('GET', revision) or '0') + 1)
redis.call('SET', revision, string.format('%.0f', nextRevision))
return {state, position, redis.call('ZCARD', waiting), redis.call('ZCARD', reserved),
        redis.call('ZCARD', loading), nextRevision, claimed, completed}
