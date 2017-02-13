function countlines(t)
  count = 0
  for k, v in pairs(t) do
    count = count + 1
  end
  return count
end

function countentries(ud)
  count = 0
  for k, v in ud() do
    count = count + 1
  end
  return count
end

function getline(t, i)
  count = 0
  for k,v in pairsByKeys(t) do
    if count >= i then
      return v
    else 
      count = count + 1
    end
  end
end

function getentry(ud, i)
  count = 0
  for k, v in ud() do
    if count >= i then
      return v
    else
      count = count + 1
    end
  end
end

function getudvalue(ud, key)
  for k, v in ud() do
    if tostring(k) == key then
      return v
    end
  end
end

function pairsByKeys (t, f)
  local a = {}
  for n in pairs(t) do table.insert(a, n) end
  if f ~= nil then
    table.sort(a, f)
  else
    table.sort(a)
  end
  local i = 0      -- iterator variable
  local iter = function ()   -- iterator function
    i = i + 1
    if a[i] == nil then return nil
    else return a[i], t[a[i]]
    end
  end
  return iter
end

function sendevent(i, choices)
  l = getline(lines, tonumber(i))
  if choices == nil then
    l, ret = replaceTokens(l)
  else
    l, ret = replaceTokens(l, choices)
  end
  sendEvent(l)
  return ret
end

function setcountdown()
  -- Countdown a random amount of seconds
  upper = getudvalue(options["countdown"], tostring(state["stage"]))
  countdown =  math.random(1, upper)
  state["countdown"] = countdown
end

if state["countdown"] == nil or state["countdown"] == 0 then
  if state["stage"] == 0 then
    debug("stage 0")
    -- Pick a user, output first event
    if state["user"] == nil then
      userline = math.random( 0, countentries(options["users"])-1 )
      debug("userline: "..userline)
      user = getentry(options["users"], userline)
      setToken("user", user)
      debug("setToken for user: "..user)
      state["user"] = user
    end
    sendevent(0)

    setcountdown()
  else if state["stage"] == 1 then
    debug("stage 1")
    if state["stage1line"] == 0 then
      state["repeats"] = 3
    end
    state["stage1line"] = state["stage1line"] + 1
    if state["stage1choices"] == nil then
      state["stage1choices"] = sendevent(state["stage1line"])
    else
      sendevent(state["stage1line"], state["stage1choices"])
    end
    -- If we go past line 4, reset
    if state["stage1line"] > 3 then
      state["stage1line"] = 0
    end
    setcountdown()
  else if state["stage"] == 2 then
    debug("stage 2")
    sendevent(5)
    setcountdown()
  end end end
else
  state["countdown"] = state["countdown"] - 1
  if state["countdown"] == 0 then
    if state["repeats"] == 0 then
      state["stage"] = state["stage"] + 1
      -- If we're over two, start over
      if state["stage"] > 2 then
        debug("reset!")
        state["stage"] = 0 
      end
    else
      debug("repeats: "..state["repeats"])
      state["repeats"] = state["repeats"] - 1
    end
  end
end