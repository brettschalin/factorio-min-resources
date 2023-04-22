-- defines a table of simple FIFO queues. Combined with the task definitions this creates a DAG
-- that specifies everything that needs to be done and in what order
local queues = {}

-- on push, we assign an ID to each item, of the form "queuename_0".
-- This keeps track of what number to append
local ids = {}

local finished_tasks = {}

local function parse_id(id)
    patt = "_%d+$"
    idx, id_end, match = string.find(id, patt)    
    q_name = string.sub(id, 1, idx - 1)
    id_num = string.sub(id, idx + 1)
    return q_name, id_num
end

function queues.add_queue(name)
    queues[name] = {}
    ids[name] = 0
end

function queues.push(q_name, value)
    if not queues[q_name] then
        queues.add_queue(q_name)
    end

    value.id = q_name .. "_" .. ids[q_name]
    ids[q_name] = ids[q_name] + 1

    table.insert(queues[q_name], value)
    return value.id
end

function queues.mark_done(id)
    finished_tasks[id] = true
end

function queues.is_done(id)
    return finished_tasks[id] == true
end

-- looks at the next item without removing it
function queues.peek(q_name)
    len = #queues[q_name]
    if len > 0 then
        return queues[q_name][len]
    end
    return nil
end

-- looks at the next item and removes it
function queues.pop(q_name)
    value = nil
    if queues[q_name] then
        value = table.remove(queues[q_name], 1)
    end
    return value
end

function queues.len(q_name)
    if not queues[q_name] then
        return 0
    end
    return #queues[q_name]
end

function queues.is_empty(q_name)
    return queues.len(q_name) == 0
end

return queues