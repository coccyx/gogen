# Gogen Performance Release: FastPath Generation

## Overview

This release introduces **FastPath generation**, a major performance optimization that can deliver **4-6x faster event generation** and **99% fewer memory allocations** for compatible samples.

## Performance Improvements

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Events per second | ~7,000/s | ~28,000/s | **4x** |
| Memory allocations | 17 per event | 0.003 per event | **99.98%** |
| Throughput | 66 MB/s | 437 MB/s | **6.6x** |

## How It Works

FastPath pre-compiles your sample's output format at startup, eliminating runtime overhead:

**Traditional Path (per event):**
```
Copy event map → Find tokens → Replace strings → Marshal to JSON → Write
```

**FastPath (per event):**
```
Write pre-compiled bytes → Generate token value → Write more bytes → Done
```

No maps. No JSON marshaling. No string concatenation. Just direct byte writes.

---

## Samples That ARE Faster (FastPath Enabled)

### ✅ Static token samples

```yaml
name: weblog
tokens:
  - name: ip
    type: static
    replacement: "192.168.1.100"
    format: template
    token: $ip$
  - name: user
    type: static
    replacement: "admin"
    format: template
    token: $user$
lines:
  - _raw: "$ip$ - $user$ - GET /index.html 200"
```
**Result: ~4x faster**

### ✅ Random value tokens

```yaml
name: firewall
tokens:
  - name: src_ip
    type: random
    replacement: ipv4
    format: template
    token: $src$
  - name: dst_ip
    type: random
    replacement: ipv4
    format: template
    token: $dst$
  - name: bytes
    type: random
    replacement: integer
    lower: 100
    upper: 50000
    format: template
    token: $bytes$
lines:
  - _raw: "src=$src$ dst=$dst$ bytes=$bytes$"
```
**Result: ~4x faster**

### ✅ Choice tokens

```yaml
name: auth-log
tokens:
  - name: action
    type: choice
    choice:
      - "login"
      - "logout"
      - "failed"
    format: template
    token: $action$
  - name: user
    type: choice
    choice:
      - "alice"
      - "bob"
      - "charlie"
    format: template
    token: $user$
lines:
  - _raw: "user=$user$ action=$action$"
```
**Result: ~4x faster**

### ✅ Weighted choice tokens

```yaml
name: http-status
tokens:
  - name: status
    type: weightedChoice
    weightedChoice:
      - weight: 90
        choice: "200"
      - weight: 5
        choice: "404"
      - weight: 5
        choice: "500"
    format: template
    token: $status$
lines:
  - _raw: "GET /api/endpoint HTTP/1.1 $status$"
```
**Result: ~4x faster**

### ✅ Timestamp tokens

```yaml
name: syslog
tokens:
  - name: ts
    type: timestamp
    replacement: "%Y-%m-%dT%H:%M:%S"
    format: template
    token: $ts$
lines:
  - _raw: "$ts$ myhost sshd[1234]: Connection accepted"
```
**Result: ~4x faster**

### ✅ File/CSV-based samples with template tokens

```yaml
name: user-activity
tokens:
  - name: username
    type: file
    sample: usernames
    format: template
    token: $user$
lines:
  - _raw: "User $user$ logged in"
```
**Result: ~4x faster**

### ✅ Multiple output formats

FastPath supports these output formats:
- `raw` - Plain text output
- `json` - JSON formatted output
- `splunkhec` - Splunk HTTP Event Collector format
- `elasticsearch` - Elasticsearch bulk format

---

## Samples That Are NOT Faster (Traditional Path)

### ❌ Regex format tokens

```yaml
name: log-with-regex
tokens:
  - name: timestamp
    type: timestamp
    replacement: "%Y-%m-%d"
    format: regex                    # ← Regex format not supported
    token: '\d{4}-\d{2}-\d{2}'
lines:
  - _raw: "2024-01-15 Event occurred"
```
**Why:** Regex replacement requires runtime pattern matching that can't be pre-compiled.

**Workaround:** Convert to template format if possible:
```yaml
tokens:
  - name: timestamp
    type: timestamp
    replacement: "%Y-%m-%d"
    format: template                 # ← Use template instead
    token: $timestamp$
lines:
  - _raw: "$timestamp$ Event occurred"
```

### ❌ Script tokens (Lua)

```yaml
name: custom-logic
tokens:
  - name: computed
    type: script                     # ← Script tokens not supported
    script: |
      return math.random(1, 100) * 2
    format: template
    token: $computed$
lines:
  - _raw: "Value: $computed$"
```
**Why:** Lua scripts have side effects and complex state that prevents pre-compilation.

**Workaround:** If your script is simple, consider using built-in token types:
```yaml
tokens:
  - name: computed
    type: random
    replacement: integer
    lower: 2
    upper: 200
    format: template
    token: $computed$
```

### ❌ SinglePass mode samples

```yaml
name: singlepass-sample
singlepass: true                     # ← SinglePass not supported
tokens:
  - name: field
    type: static
    replacement: "value"
lines:
  - _raw: "Event with $field$"
```
**Why:** SinglePass uses a different internal representation (BrokenLines) that's already optimized differently.

### ❌ Custom Lua generators

```yaml
name: lua-generated
generator: my-custom-generator       # ← Custom generators not supported
```
**Why:** Custom generators have their own logic that FastPath can't optimize.

### ❌ Unsupported output formats

These output formats use the traditional path:
- `rfc3164` - BSD syslog format
- `rfc5424` - IETF syslog format
- Custom templates

**Why:** These formats require field-by-field formatting that differs from the pre-compiled approach.

---

## How to Check if FastPath is Enabled

When gogen starts, look for log messages:

```
INFO Setting sample 'my-sample' to FAST PATH generator
```

vs.

```
INFO Setting sample 'my-sample' to generator 'sample'
```

You can also check at debug level for initialization:

```
INFO FastPath enabled for sample 'my-sample' with output 'json'
```

---

## Migration Guide

### Converting Regex to Template Format

**Before (slow):**
```yaml
tokens:
  - name: ip
    format: regex
    token: '\d+\.\d+\.\d+\.\d+'
    type: random
    replacement: ipv4
lines:
  - _raw: "Connection from 10.0.0.1 accepted"
```

**After (fast):**
```yaml
tokens:
  - name: ip
    format: template
    token: $ip$
    type: random
    replacement: ipv4
lines:
  - _raw: "Connection from $ip$ accepted"
```

### Converting Script to Built-in Types

**Before (slow):**
```yaml
tokens:
  - name: guid
    type: script
    script: |
      local hex = "0123456789abcdef"
      local result = ""
      for i = 1, 32 do
        result = result .. hex:sub(math.random(1, 16), math.random(1, 16))
      end
      return result
```

**After (fast):**
```yaml
tokens:
  - name: guid
    type: random
    replacement: guid
    format: template
    token: $guid$
```

---

## Benchmark Your Samples

Run benchmarks to see the improvement for your specific samples:

```bash
# Run the built-in benchmarks
go test -bench=. -benchmem ./generator/...

# Example output:
# BenchmarkTraditionalGen100    7796    138546 ns/op    99425 B/op    1703 allocs/op
# BenchmarkFastGen100          34972     35821 ns/op    18580 B/op       3 allocs/op
```

---

## Summary

| Sample Type | FastPath | Expected Speedup |
|-------------|----------|------------------|
| Template tokens (static, random, choice, timestamp) | ✅ Yes | 4-6x |
| File/CSV tokens with template format | ✅ Yes | 4-6x |
| Regex format tokens | ❌ No | - |
| Script (Lua) tokens | ❌ No | - |
| SinglePass samples | ❌ No | - |
| Custom generators | ❌ No | - |
| RFC3164/RFC5424 output | ❌ No | - |

**To maximize performance:** Use template format tokens with `raw`, `json`, `splunkhec`, or `elasticsearch` output formats.
