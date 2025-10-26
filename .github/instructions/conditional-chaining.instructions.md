---
description: "Instructions for chaining conditional statements"
applyTo: "**/*.go"
---
# Conditional Statement Chaining Instructions

## If/Else If/Else Chaining

- Chain separate `if` statements into `if/else if/else` blocks when they are mutually exclusive
- Reduces vertical scrolling and improves code readability
- Makes control flow more explicit and easier to understand

### Pattern: Validation Checks

**Avoid separate if statements:**
```go
if ctx == nil {
    return nil, fmt.Errorf("context cannot be nil")
}

if logger == nil {
    return nil, fmt.Errorf("logger cannot be nil")
}

if description == "" {
    return nil, fmt.Errorf("description cannot be empty")
}
```

**Prefer chained if/else if:**
```go
if ctx == nil {
    return nil, fmt.Errorf("context cannot be nil")
} else if logger == nil {
    return nil, fmt.Errorf("logger cannot be nil")
} else if description == "" {
    return nil, fmt.Errorf("description cannot be empty")
}
```

### Pattern: Switch-Like Logic

**Avoid separate if statements:**
```go
if description == "start" {
    Log(logger, "Starting service")
}

if description == "stop" {
    Log(logger, "Stopping service")
}

if description == "restart" {
    Log(logger, "Restarting service")
}
```

**Prefer chained if/else if/else:**
```go
if description == "start" {
    Log(logger, "Starting service")
} else if description == "stop" {
    Log(logger, "Stopping service")
} else if description == "restart" {
    Log(logger, "Restarting service")
}
```

### When NOT to Chain

- Independent conditions that should all be evaluated (not mutually exclusive)
- Error accumulation patterns where multiple errors are collected
- Cases where early returns prevent subsequent checks naturally
