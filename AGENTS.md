# Agent Guidelines — go-qbittorrent-webapi

> **Audience**: This document is for **developers and agents** who write or modify code in this repository. It contains global architectural rules, coding patterns, and conventions. It is **not** end-user documentation.
>
> **Scope**: `AGENTS.md` defines **how to code** (structure, patterns, naming). `README.md` defines **how to use** the library (installation, examples, API surface). Keep these concerns separate: never explain implementation details in `README.md`, and never add usage tutorials in `AGENTS.md`.

## Table of Contents

- [Interaction Rules (Critical)](#interaction-rules-critical)
- [Project Overview](#project-overview)
- [File Layout & Naming](#file-layout--naming)
- [Go Style, Idioms, and Error Handling](#go-style-idioms-and-error-handling)
- [Data Types and Structs](#data-types-and-structs)
  - [Optional Fields](#optional-fields)
  - [Enum-like Types](#enum-like-types)
  - [Undocumented Fields](#undocumented-fields)
  - [Large Flat Structs](#large-flat-structs)
- [Custom JSON Marshaling](#custom-json-marshaling)
- [Implementing a New Endpoint](#implementing-a-new-endpoint)
- [HTTP Internals (request.go)](#http-internals-requestgo)
- [Dependencies](#dependencies)

## Interaction Rules (Critical)

- **Never** write, edit, delete, or create any file in the repository unless the user **explicitly** asks for it.
- If the user says "propose me something", "what do you think", "suggest", etc., respond **in the conversation only**.
- An action on the repository (write, edit, delete, move, git commit, etc.) must **always** be preceded by an explicit request from the user.
- When in doubt, ask for confirmation before modifying anything.

## Project Overview

This is a **Go client library** for the [qBittorrent v5 Web API](https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)).

It exposes a stateful `Client` object that handles HTTP cookie-based session management and automatic re-authentication. It is **not** an application or a service: there is no `main`, no CLI flags, and no internal logging.

## File Layout & Naming

### Directory & File Layout

| File | Role |
|------|------|
| `client.go` | Constructor (`New`), `Client` struct definition, HTTP client setup (cookie jar, cleanhttp). Also contains `APIReferenceVersion`, an exported constant tracking the qBittorrent wiki version this library targets. |
| `request.go` | **Internal** HTTP plumbing: `requestBuild`, `requestExecute`, `requestExtract`. These must never be exposed publicly. |
| `api_auth.go` | Authentication domain (`Login`, `Logout`). |
| `api_app.go` | Application domain (`GetApplicationVersion`, `SetApplicationPreferences`, etc.). |
| `api_torrents.go` | Torrents domain (`GetTorrentList`, `AddNewTorrents`, etc.). |
| `api_<domain>.go` | Other API domains (`transfer`, `sync`, `log`, `rss`, `search`). One file per domain. |
| `helpers.go` | Public helpers (`String`, `Int`, `Bool`) and domain-specific value types (e.g. `Speed`). |
| `go.mod` / `go.sum` | Go module definitions. Keep external dependencies to a minimum. |

**Rule**: when adding a new endpoint, identify its qBittorrent API domain and place it in the corresponding `api_<domain>.go` file. Create a new file only for a domain that does not exist yet.

### Naming Conventions

- **Package name**: `qbtapi`.
- **API domain constant** (private): `<domain>APIName` (e.g. `torrentsAPIName = "torrents"`).
- **Public methods** on `*Client`: use descriptive action names matching the API operation, e.g. `GetTorrentList`, `SetApplicationPreferences`, `DeleteTorrents`.
- **Context first**: every public method that performs an HTTP call must take `ctx context.Context` as its first argument.

## Go Style, Idioms, and Error Handling

### Named Return Values

Use **named return values**: `(result SomeType, err error)`. Use bare `return` when populating them; do **not** write `return nil, err`. Named returns may be omitted for trivial helpers (e.g. `String()`, `Int()`, `Bool()`) and standard-library interface implementations (e.g. `json.Unmarshaler`, `fmt.Stringer`) where the return values are self-evident.

### Error Wrapping

Wrap every error with `%w` before returning it. The error prefix should describe the **intended action** that failed (e.g. `"building request failed"`, `"executing request failed"`, `"reading file %q failed"`), not the name of the function returning the error.

```go
if err != nil {
    err = fmt.Errorf("building request failed: %w", err)
    return
}
```

In private helpers with many sub-actions, wrap each sub-action individually so the caller knows exactly which step failed:

```go
if err = mp.WriteField("savepath", *options.SavePath); err != nil {
    err = fmt.Errorf("writing savepath to form field failed: %w", err)
    return
}
if err = mp.WriteField("upLimit", strconv.Itoa(options.UploadLimit.ToBytes())); err != nil {
    err = fmt.Errorf("writing upLimit to form field failed: %w", err)
    return
}
```

### Logging and Panic

- **No logging** inside the library. Errors are always returned to the caller.
- Avoid `panic` for external or runtime errors. Validate inputs explicitly and return errors.
- `panic` is acceptable only for internal "should never happen" invariants (e.g., `json.Marshal` on a known-good struct failing) to signal a bug in the library itself, not the remote server.
- Do not use generics or overly abstract helpers unless the existing codebase already uses them for the same pattern.

### Error Types

- `HTTPError(int)` — for unexpected HTTP status codes.
- `InternalError(string)` — for states that should never happen (e.g. unsupported `Content-Type`, invalid output pointer kind).
- Always wrap the underlying error with context. Never discard it.

## Data Types and Structs

### Optional Fields

qBittorrent uses sparse JSON objects for updates. Represent optional fields as **pointers** with `omitempty`:

```go
type SomePreferences struct {
    SavePath *string `json:"save_path,omitempty"`
    Paused   *bool   `json:"paused,omitempty"`
}
```

**Exception**: slices and maps do **not** need pointer wrapping because they are already reference types with a native `nil` value (e.g. `Tags []string`, `Hashes []string`).

**Exception**: structs used exclusively for multipart/form-data serialization (e.g. `AddNewTorrentsOptions`) do **not** need `json` tags at all, since they are manually translated into form fields rather than JSON-encoded.

Provide the existing helpers so callers can easily create pointers:
- `String(v string) *string`
- `Int(v int) *int`
- `Bool(v bool) *bool`

### Enum-like Types

Use a dedicated type with typed constants and a `String()` method.

The qBittorrent API uses both string-encoded and integer-encoded enums. The underlying type is dictated by the API documentation — do not choose based on preference.

**String enums** represent API values verbatim. They do **not** need a `String()` method because their raw value is already meaningful. Every constant should have a trailing `//` comment describing what the value represents; the text should be copied straight from the official qBittorrent wiki documentation for that endpoint.

```go
type FilterState string

const (
    FilterStateAll    FilterState = "all"    // No filtering
    FilterStateActive FilterState = "active" // Only active torrents
    // ...
)
```

**Numeric enums** represent API values as integers. They **must** implement `String()` to return a human-readable description. Hardcode the exact integer values from the official qBittorrent wiki documentation — **do not use `iota`**. Explicit values prevent silent bugs if the API changes or if a value is skipped (e.g. `0` is absent or a negative value is required). Constant comments should likewise be copied from the wiki:

```go
type TorrentTrackerStatus uint8

const (
    TorrentTrackerDisabled TorrentTrackerStatus = 0 // Tracker is disabled
    TorrentTrackerWorking  TorrentTrackerStatus = 2 // Tracker is working
    // ...
)

func (tts TorrentTrackerStatus) String() string { ... }
```

Every enum-like type must also implement a `Ptr()` method with a **value receiver**, so it can be used in optional struct fields (see [Optional Fields](#optional-fields)). Do not make this conditional — always provide it:

```go
func (fs FilterState) Ptr() *FilterState {
    return &fs
}
```

#### When the server behaves differently from the documented API

The official qBittorrent wiki documents the API for a specific version. Newer qBittorrent releases may change field serializations without updating the wiki. For example, `proxy_type` is documented as an integer, yet qBittorrent v5.0+ sends and accepts a **string** (`"None"`, `"SOCKS4"`, etc.).

When a numeric enum field changes its wire format to a string in a newer server version, add custom `UnmarshalJSON` / `MarshalJSON` methods that accept **both** representations. This preserves backward compatibility with older servers while fixing unmarshaling against newer ones. See `ProxyType` in `api_app.go` for the reference implementation:

- `UnmarshalJSON` tries the original `int` first; if that fails, falls back to `string`.
- `MarshalJSON` writes the new `string` format for forward compatibility.

Only apply this pattern when a test or user report proves the discrepancy — do not add it speculatively to other enums.

### Undocumented Fields

If a field exists in the API response but is absent from the official qBittorrent wiki, mark it with a `// undocumented` comment on the line immediately above the field:

```go
// undocumented
Platform string `json:"platform"` // Platform (e.g. Linux)
```

This signals that the field was discovered empirically and may not be stable.

**Development workflow**: when adding a new endpoint, unmarshal the response into a `map[string]any` first to inspect the actual payload. Use `reflect.TypeOf()` on each value to determine its JSON type. Compare every key against the wiki documentation. Any key present in the response but missing from the wiki becomes an `// undocumented` field in the final struct.

### Large Flat Structs

Very large flat preference structs (e.g. `ApplicationPreferences`) may implement `String()` and `GoString()` for human-readable output. Raw struct formatting is useless for 100+ fields, so round-tripping through JSON to `map[string]any` is acceptable:

```go
func (ap ApplicationPreferences) getMap() (data map[string]any) {
    payload, err := json.Marshal(ap)
    if err != nil {
        panic(err)
    }
    if err = json.Unmarshal(payload, &data); err != nil {
        panic(err)
    }
    return
}

func (ap ApplicationPreferences) String() string {
    return fmt.Sprintf("%v", ap.getMap())
}

func (ap ApplicationPreferences) GoString() string {
    return fmt.Sprintf("%+v", ap.getMap())
}
```

This is the only situation where `panic` in JSON handling is justified — it signals an internal library bug, not an external server error.

## Custom JSON Marshaling

When the API returns a value that does not map directly to a Go type (Unix timestamps, raw strings that need parsing, etc.), **do not** expose the raw form in public structs. Use the established `mask` pattern in custom `UnmarshalJSON` / `MarshalJSON` to convert to the proper Go type.

`UnmarshalJSON` and `MarshalJSON` must always be implemented as a **paired set** so that round-tripping through JSON does not corrupt data.

Fields that are transformed by custom JSON methods should carry a `json:"-"` tag as a **visual indicator** that they are not serialized as-is.

### The Mask Pattern

The `mask` alias avoids infinite recursion when calling `json.Unmarshal` / `json.Marshal` inside the custom methods. There are two variants:

1. **Standalone `mask` struct** — use when the transformed field has a different Go type but the same JSON key as the public field. Because the names clash, you cannot embed `mask`; instead declare a fresh struct that lists every raw field and copy them manually.
2. **Embedded `type mask MyStruct`** — use when you only need to override a subset of fields. Embed `*mask` (pointer) in `UnmarshalJSON` so writes go directly into the receiver, and embed `mask` (value) in `MarshalJSON` to avoid mutating the receiver during serialization.

### Example 1 — Unix timestamp → `time.Time` (standalone mask)

```go
func (t *MyStruct) UnmarshalJSON(data []byte) error {
    type mask struct {
        CreatedAt int64 `json:"created_at"`
        // ... other raw fields
    }
    var m mask
    if err := json.Unmarshal(data, &m); err != nil {
        return err
    }
    t.CreatedAt = time.Unix(m.CreatedAt, 0)
    // ...
    return nil
}
```

### Example 2 — raw string → `*url.URL` (embedded mask)

```go
func (t *MyStruct) UnmarshalJSON(data []byte) (err error) {
    type mask MyStruct
    tmp := struct {
        *mask
        URL string `json:"url"`
    }{
        mask: (*mask)(t),
    }
    if err = json.Unmarshal(data, &tmp); err != nil {
        return
    }
    if t.URL, err = url.Parse(tmp.URL); err != nil {
        err = fmt.Errorf("parsing URL failed: %w", err)
        return
    }
    return
}

func (t MyStruct) MarshalJSON() ([]byte, error) {
    type mask MyStruct
    tmp := struct {
        mask
        URL string `json:"url"`
    }{
        mask: mask(t),
        URL:  t.URL.String(),
    }
    return json.Marshal(tmp)
}
```

### Example 3 — raw integer seconds → `time.Duration` (embedded mask)

```go
func (t *MyStruct) UnmarshalJSON(data []byte) (err error) {
    type mask MyStruct
    tmp := struct {
        *mask
        ETA int `json:"eta"` // seconds
    }{
        mask: (*mask)(t),
    }
    if err = json.Unmarshal(data, &tmp); err != nil {
        return
    }
    t.ETA = time.Duration(tmp.ETA) * time.Second
    return
}

func (t MyStruct) MarshalJSON() ([]byte, error) {
    type mask MyStruct
    tmp := struct {
        mask
        ETA int `json:"eta"` // seconds
    }{
        mask: mask(t),
        ETA:  int(t.ETA.Seconds()),
    }
    return json.Marshal(tmp)
}
```

### Example 4 — raw bytes / bytes-per-second → `cunits.Bits` and `Speed` (embedded mask)

The API returns sizes as raw bytes and speeds as bytes/second (with `-1` meaning "unlimited"). Use `cunits.Bits` for sizes and the wrapper `Speed` type for speeds.

```go
func (t *MyStruct) UnmarshalJSON(data []byte) (err error) {
    type mask MyStruct
    tmp := struct {
        *mask
        TotalSize  int `json:"total_size"`  // bytes
        SpeedLimit int `json:"speed_limit"` // bytes/s, -1 if unlimited
    }{
        mask: (*mask)(t),
    }
    if err = json.Unmarshal(data, &tmp); err != nil {
        return
    }
    t.TotalSize = cunits.ImportInBytes(float64(tmp.TotalSize))
    t.SpeedLimit = GetSpeedFromBytes(tmp.SpeedLimit)
    return
}

func (t MyStruct) MarshalJSON() ([]byte, error) {
    type mask MyStruct
    tmp := struct {
        mask
        TotalSize  int `json:"total_size"`  // bytes
        SpeedLimit int `json:"speed_limit"` // bytes/s, -1 if unlimited
    }{
        mask:       mask(t),
        TotalSize:  int(t.TotalSize.Bytes()),
        SpeedLimit: t.SpeedLimit.ToBytes(),
    }
    return json.Marshal(tmp)
}
```

### Example 5 — comma-separated string ↔ `[]string` (embedded mask)

Some API fields return tags or lists as a comma-separated string. Convert them to a `[]string` on unmarshal and join them back on marshal. Extract the separator as an unexported constant so that split and join remain perfectly symmetric.

```go
const tagListSeparator = ","

func (t *MyStruct) UnmarshalJSON(data []byte) (err error) {
    type mask MyStruct
    tmp := struct {
        *mask
        Tags string `json:"tags"` // Comma-separated tag list
    }{
        mask: (*mask)(t),
    }
    if err = json.Unmarshal(data, &tmp); err != nil {
        return
    }
    if tmp.Tags != "" {
        t.Tags = strings.Split(tmp.Tags, tagListSeparator)
        for i := range t.Tags {
            t.Tags[i] = strings.TrimSpace(t.Tags[i])
        }
    }
    return
}

func (t MyStruct) MarshalJSON() ([]byte, error) {
    type mask MyStruct
    tmp := struct {
        mask
        Tags string `json:"tags"` // Comma-separated tag list
    }{
        mask: mask(t),
        Tags:  strings.Join(t.Tags, tagListSeparator),
    }
    return json.Marshal(tmp)
}
```

## Implementing a New Endpoint

Follow this exact flow when adding a public API method:

1. **Add the domain header and constant** if the domain file is new.
   The first comment line must be the **exact section title** from the qBittorrent wiki documentation. The constant value must be the corresponding **URL path segment** used in the API route (`/api/v2/<segment>/methodName`).

   ```go
   /*
   	Transfer
   	https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#transfer
   */

   const (
   	transferAPIName = "transfer"
   )
   ```

   | Wiki Section Title | URL Segment | Constant |
   |--------------------|-------------|----------|
   | Application | `app` | `applicationAPIName = "app"` |
   | Authentication | `auth` | `authenticationAPIName = "auth"` |
   | Torrent management | `torrents` | `torrentsAPIName = "torrents"` |

2. **Define request/response structs** if needed. Use pointers for optional fields. Every exported struct field should have a trailing `//` comment explaining its purpose and units where applicable.
   Within a domain file, group related methods and types with a block comment sub-section header matching the qBittorrent wiki sub-section name:
   ```go
   /*
   	Listing
   */
   ```

3. **Implement the method** on `*Client`:
   - Validate method inputs **before** calling `requestBuild`. Return an error early for invalid arguments.
   - If the endpoint has no query parameters or form data, pass `nil` for `parameters` (not an empty map).
   - If the endpoint returns no meaningful body, pass `nil` as the `output` argument to `requestExecute`.
   - If the endpoint returns `text/plain` with a known success marker (e.g. `"Ok."`), validate against the `expectedSuccessResponse` constant.
   - If an endpoint requires complex input preparation (e.g. reading files from disk), provide a dedicated exported helper function and reference it in the method's godoc.
   - If an endpoint requires headers or a body format that `requestBuild` does not produce by default (e.g. multipart uploads, CSRF `Origin` headers), mutate the `*http.Request` after `requestBuild` and before `requestExecute`.
   - If the endpoint accepts files, URLs, or other complex payloads, validate format, scheme, extension, and non-emptiness **before** building the request body (see `torrentAddGeneratePayload` in `api_torrents.go`).
   - If the API expresses a duration as an integer with an implicit unit (seconds, minutes, etc.), expose `time.Duration` in the public struct and convert to the API unit at the serialization boundary.

   ```go
   // GetGlobalTransferInfo returns ...
   // https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-global-transfer-info
   func (c *Client) GetGlobalTransferInfo(ctx context.Context) (info GlobalTransferInfo, err error) {
       req, err := c.requestBuild(ctx, "GET", transferAPIName, "info", nil, nil)
       if err != nil {
           err = fmt.Errorf("building request failed: %w", err)
           return
       }
       if err = c.requestExecute(req, &info, true); err != nil {
           err = fmt.Errorf("executing request failed: %w", err)
       }
       return
   }
   ```

4. **Parameters / filters**: if the endpoint accepts query parameters or form data, build a `map[string]string`. For complex filters, define a struct with a private `getLowLevelRepr() map[string]string` method that translates typed fields into the flat string map (see `ListFilters` in `api_torrents.go`).
   - When an endpoint accepts multiple values in a single field, extract the separator as a named constant from the wiki documentation (e.g. `hashListSeparator = "|"`). Do not reuse separators across unrelated endpoints unless the wiki explicitly confirms they are the same.
   - If a JSON field also uses the same separator encoding (e.g. comma-separated tags), use the same named constant in both `getLowLevelRepr()` and custom `MarshalJSON` / `UnmarshalJSON` methods so split and join stay perfectly symmetric.

5. **Documentation**: every exported method must have a godoc comment ending with the link to the official qBittorrent wiki anchor.
   - If a field's purpose is unclear from the wiki (or differs from it), use a `// TODO` comment explaining the open question.
   - If an endpoint is implemented but known not to work (e.g. due to unclear upstream requirements), add a `// NOT WORKING FOR NOW.` comment at the method level with a brief explanation.

6. **Update `README.md`**: tick the endpoint in the implementation checklist.

7. **Add integration tests** in the domain's `_test.go` file:
   - Tests must be part of the existing domain-scoped scenario function (e.g. `TestTorrentsDomain`), not a new top-level `TestXxx` function.
   - Every mutating action must be followed by a read-back that verifies the change was actually applied on the server (state verification, not just "no error").
   - Use `t.Skipf` for transient failures (network issues, server config dependencies like queueing disabled).
   - Clean up any side effects (categories, tags, added torrents) within the same scenario flow.

## HTTP Internals (request.go)

- `requestBuild(ctx, method, apiName, methodName, parameters, body)` — builds the `*http.Request`. `parameters` is always `map[string]string` (use `nil` when there are no parameters). `body` can be `nil`.
  - qBittorrent has two special parameter encodings: `parameters["cookies"] = rawJSON` sends `cookies=<rawJSON>`, and `parameters["json"] = rawJSON` sends `json=<rawJSON>`. Use these only when the endpoint requires them.
  - When `parameters` are encoded as a POST body (i.e. method is `POST` and no explicit `body` was provided), `requestBuild` automatically sets `Content-Type: application/x-www-form-urlencoded` and `Content-Length` on the returned request.
  - For multipart uploads (e.g. file upload), build the payload locally into a `bytes.Buffer`, obtain the `Content-Type` from the multipart writer, pass the buffer as `body` to `requestBuild` with `nil` parameters, then manually set `req.Header.Set(contentTypeHeader, contentType)` before calling `requestExecute`.
- `requestExecute(req, output, autoAuth)` — executes the request. `output` can be `nil` when the caller does not need to read the response body. If `autoAuth` is `true` and a `403` is received, it closes the response body, calls `Login`, resets the request body via `request.GetBody()`, and retries once with `autoAuth = false`. Use `autoAuth = true` for all endpoints except auth-lifecycle methods (e.g. `Logout`) that must not trigger a recursive login.
- `requestExtract(response, output)` — unmarshals the body based on `Content-Type`:
  - `text/plain` → `*string`
  - `application/json` → struct pointer, slice pointer, or `*string` (some endpoints return a JSON-encoded string such as `"Ok."`)
  - Unsupported content types produce an `InternalError`.

**Do not** modify the core retry or auto-auth logic without a strong reason. If you add an endpoint that behaves differently (e.g. file upload), handle the special body/header logic locally in the domain file (see `AddNewTorrents` as an example).

Some endpoints historically returned `text/plain` (e.g. `"Ok."`) but newer qBittorrent versions may return `application/json` instead. For write operations where the response body carries no meaningful data, pass `nil` for `output` and rely on the HTTP status code. This is the pattern used by `AddNewTorrents`.

The `Login` endpoint is another example: it manually constructs and sets an `Origin` header to satisfy qBittorrent's CSRF protection. This is done in `api_auth.go` after calling `requestBuild`.

## Dependencies

Keep the dependency tree small. Current allowed dependencies:
- `github.com/hashicorp/go-cleanhttp`
- `github.com/hekmon/cunits/v3`
- `golang.org/x/net/publicsuffix`

Do not add new external libraries unless the user explicitly validates the choice.
