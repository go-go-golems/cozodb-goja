---
Title: "Module 6 - JavaScript Integration with Goja"
Ticket: CO-03
Status: active
Topics:
    - cozodb
    - course
    - javascript
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go:Go CLI that creates Goja VM and runs JS plugins"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go:Plugin loading and validation"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js:Plugin example"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_extractor_factory.js:Extractor factory"
ExternalSources:
    - https://github.com/dop251/goja
    - https://github.com/go-go-golems/go-go-goja
Summary: "How the Go runtime embeds JavaScript via Goja, module system, and how to write extraction plugins"
LastUpdated: 2026-02-24T19:00:00-05:00
WhatFor: "Understand the Go<->JavaScript bridge and write custom extraction plugins"
WhenToUse: "When building or modifying JavaScript extraction plugins"
---

# Module 6: JavaScript Integration with Goja

## Learning Objectives

By the end of this module you will:
- Understand what Goja is and why it's used instead of Node.js
- Know how Go creates and configures a JavaScript runtime
- Understand the module system (`require()`)
- Know how to register native (Go) modules accessible from JS
- Understand the Geppetto middleware system for LLM access
- Be able to write a custom extraction plugin

## 6.1 What is Goja?

**Goja** is a pure-Go implementation of ECMAScript 5.1 (with some ES6+ features). It lets you embed a JavaScript runtime inside a Go application without CGO or external dependencies.

**Why Goja instead of Node.js?**

| Feature | Goja | Node.js |
|---|---|---|
| Deployment | Single Go binary | Requires Node.js install |
| Go integration | Direct, no FFI | Requires child_process or FFI |
| Performance | Good for scripts | Better for heavy computation |
| npm ecosystem | Limited (no native modules) | Full access |
| Concurrency model | Go goroutines + event loop | Single-threaded event loop |
| ECMAScript version | ES5.1 + some ES6 | Full ES2023+ |

**Key limitation**: Goja runs ES5.1 primarily. No `async/await`, no `import/export`. But `require()` works via the `goja_nodejs` package, and the event loop package provides `setTimeout/setInterval`.

## 6.2 Creating a Goja Runtime

The Go host creates and configures the JS runtime (`main.go:408-446`):

```go
// Create the event loop (for setTimeout, Promises, etc.)
loop := eventloop.NewEventLoop()
go loop.Start()
defer loop.Stop()

// Create the JavaScript runtime
vm := goja.New()

// Create a runner that wraps the VM with panic recovery
runner := runtimeowner.NewRunner(vm, loop, runtimeowner.Options{
    Name:          "cozo-relationship-js-runner",
    RecoverPanics: true,
})
```

The `runner` wraps the raw Goja VM with safety features (panic recovery, named identification).

## 6.3 The Module System

### Setting Up `require()`

```go
// Configure module search paths
reg := require.NewRegistry(
    require.WithGlobalFolders(
        scriptRoot,                                    // e.g., ./scripts/
        filepath.Join(scriptRoot, "node_modules"),     // e.g., ./scripts/node_modules/
    ),
)

// Enable require() in the VM
reqMod := reg.Enable(vm)
```

After this, JS code can use `require("./some_module")` and modules are resolved from the configured paths.

### Registering Native (Go) Modules

Go code can register modules that appear as regular `require()` targets in JS:

```go
// Register the database module
scriptDBModule := &databasemod.DBModule{}
scriptDBModule.Configure("sqlite3", scriptDB)
reg.RegisterNativeModule("runnerdb", scriptDBModule.Loader)
reg.RegisterNativeModule("database", scriptDBModule.Loader)

// Register the Geppetto module (LLM integration)
gp.Register(reg, gpOptions)
```

Now JS can do:
```javascript
const db = require("runnerdb");
const gp = require("geppetto");
```

These are not npm packages -- they're Go functions exposed as JS modules.

### Module Resolution Order

When JS calls `require("foo")`:
1. Check registered native modules (Go-implemented)
2. Search `scriptRoot/foo.js`
3. Search `scriptRoot/node_modules/foo/index.js`
4. Search global folders

## 6.4 Global Variables

The Go host injects global variables before running any JS:

```go
vm.Set("RELATIONSHIP_RUN_ID", runID)
vm.Set("RELATIONSHIP_PROFILE", profile)
vm.Set("RELATIONSHIP_ENGINE_OPTIONS", engineOptions)
vm.Set("RELATIONSHIP_PROMPT", prompt)
vm.Set("RELATIONSHIP_TIMEOUT_MS", timeoutMs)
vm.Set("RELATIONSHIP_TRANSCRIPT", transcriptText)
vm.Set("RELATIONSHIP_SCRIPT_ROOT", scriptRoot)
vm.Set("RELATIONSHIP_SCRIPT_DB_DSN", scriptDB)
```

These are accessible in JS as global variables:
```javascript
console.log(RELATIONSHIP_TRANSCRIPT);  // The text to extract from
console.log(RELATIONSHIP_PROMPT);      // The system prompt
console.log(RELATIONSHIP_TIMEOUT_MS);  // Timeout in milliseconds
```

Additionally, the Go host installs utility globals:
```go
vm.Set("ENV", mapEnv())       // All environment variables as a map
vm.Set("console", consoleObj) // console.log, console.error
```

And an `assert` function:
```javascript
assert(condition, "error message");  // Throws if condition is falsy
```

## 6.5 The Geppetto Module

Geppetto is the LLM middleware framework. When registered, it provides:

```javascript
const gp = require("geppetto");

// Create an LLM engine from config or profile
const engine = gp.engines.fromConfig({
    apiType: "openai",
    model: "gpt-4.1-mini",
    maxTokens: 4096
});
// OR from a Pinocchio profile
const engine = gp.engines.fromProfile("my-profile", { timeoutMs: 120000 });

// Create a builder for configuring the LLM pipeline
const builder = gp.createBuilder(options)
    .withEngine(engine)
    .useGoMiddleware("systemPrompt", {
        prompt: "You are an extraction expert..."
    });

// Build a session
const session = builder.buildSession();

// Create a turn (conversation message)
const seed = gp.turns.newTurn({
    blocks: [
        gp.turns.newUserBlock("Extract entities from: " + text)
    ],
    data: {
        [gp.consts.TurnDataKeys.STRUCTURED_OUTPUT_CONFIG]: {
            mode: "json_schema",
            name: "relation_extraction",
            schema: { /* JSON Schema */ },
            strict: true,
            require_valid: true,
        }
    }
});

// Run the turn (synchronous in Goja, backed by Go goroutines)
const result = session.run(seed, {
    timeoutMs: 120000,
    tags: { app: "my-app", kind: "extraction" }
});

// Extract the assistant's response
const blocks = result.blocks || [];
const assistantText = blocks
    .filter(b => b.kind === gp.consts.BlockKind.LLM_TEXT)
    .map(b => b.payload.text)
    .join("\n");
```

### Key Geppetto Concepts

| Concept | Description |
|---|---|
| **Engine** | LLM provider configuration (OpenAI, Claude, Gemini) |
| **Builder** | Configures middleware chain for the LLM pipeline |
| **Middleware** | Processing steps (system prompt, tool calling, etc.) |
| **Session** | An execution context created from a builder |
| **Turn** | A unit of conversation (user message + optional config) |
| **Block** | A piece of content in a turn (user text, assistant text, tool calls) |

### Structured Output

The structured output config tells the LLM to return JSON matching a specific schema:

```javascript
const structuredOutput = {
    mode: "json_schema",
    name: "relation_extraction",
    description: "Transcript relation extraction payload",
    schema: RELATIONSHIP_EXTRACTION_SCHEMA,
    strict: true,          // Enforce schema compliance
    require_valid: true,   // Error if output doesn't validate
};
```

## 6.6 The Plugin Descriptor Pattern

Every extraction plugin follows this contract:

```javascript
module.exports = {
    // Required metadata
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "unique.plugin.identifier",
    name: "Human Readable Plugin Name",

    // Factory function: called once with host context
    create(hostContext) {
        // hostContext contains:
        // - app: "cozo-relationship-js-runner"
        // - runId: UUID string
        // - scriptPath: absolute path to this script
        // - scriptRoot: module resolution root
        // - profile: Pinocchio profile name
        // - scriptDB: path to script database
        // - recording: boolean
        // - recordDB: path to recording database

        return {
            // Execution function: called for each extraction
            run(input, options) {
                // input contains:
                // - transcript: string (the text to extract from)
                // - prompt: string (the system prompt)
                // - profile: string (AI profile name)
                // - timeoutMs: number (execution timeout)
                // - engineOptions: object (LLM configuration)

                // options contains:
                // - timeoutMs: number
                // - tags: { app, source }

                // Must return: JSON string or JavaScript object
                return JSON.stringify({ /* extraction result */ });
            }
        };
    }
};
```

### Validation

The Go host validates the descriptor (`plugin_loader.go:161-189`):
- `apiVersion` must be exactly `"cozo.extractor/v1"`
- `kind` must be `"extractor"`
- `id` and `name` must be non-empty strings
- `create` must be a function
- `create()` must return an object with a `run` function

### Helper Functions

The Geppetto module provides helper functions:

```javascript
const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");

// defineExtractorPlugin: validates and normalizes the descriptor
module.exports = defineExtractorPlugin({
    id: "my-plugin",
    name: "My Plugin",
    create() { /* ... */ }
});

// wrapExtractorRun: adds error handling and normalization
return {
    run: wrapExtractorRun((input) => {
        // Your extraction logic here
        return extractedData;
    })
};
```

## 6.7 Writing a Custom Plugin: Step by Step

### Step 1: Create the File

Create `scripts/my_extractor.js`:

```javascript
const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");

module.exports = defineExtractorPlugin({
    id: "my.custom-extractor",
    name: "My Custom Extractor",
    create(ctx) {
        console.log("Plugin created for run:", ctx.runId);

        return {
            run: wrapExtractorRun(function(input) {
                console.log("Processing transcript of length:", input.transcript.length);

                // Your extraction logic here
                return { message: "Hello from custom extractor!" };
            })
        };
    }
});
```

### Step 2: Add LLM Integration

```javascript
const gp = require("geppetto");
const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");

module.exports = defineExtractorPlugin({
    id: "my.llm-extractor",
    name: "My LLM Extractor",
    create(ctx) {
        return {
            run: wrapExtractorRun(function(input) {
                // Create engine from the resolved config
                var engine = input.engineOptions
                    ? gp.engines.fromConfig(input.engineOptions)
                    : gp.engines.fromProfile(input.profile || "", {
                        timeoutMs: input.timeoutMs || 120000
                    });

                // Build the LLM pipeline
                var builder = gp.createBuilder()
                    .withEngine(engine)
                    .useGoMiddleware("systemPrompt", {
                        prompt: input.prompt || "Extract key facts from the text."
                    });

                var session = builder.buildSession();

                // Create the user turn
                var seed = gp.turns.newTurn({
                    blocks: [
                        gp.turns.newUserBlock(
                            "Extract all facts from:\n\n" + input.transcript
                        )
                    ],
                    data: {
                        [gp.consts.TurnDataKeys.STRUCTURED_OUTPUT_CONFIG]: {
                            mode: "json_schema",
                            name: "fact_extraction",
                            schema: {
                                type: "object",
                                properties: {
                                    facts: {
                                        type: "array",
                                        items: {
                                            type: "object",
                                            properties: {
                                                subject: { type: "string" },
                                                predicate: { type: "string" },
                                                object: { type: "string" }
                                            },
                                            required: ["subject", "predicate", "object"]
                                        }
                                    }
                                },
                                required: ["facts"]
                            },
                            strict: true,
                            require_valid: true
                        }
                    }
                });

                // Run inference
                var result = session.run(seed, {
                    timeoutMs: input.timeoutMs || 120000
                });

                // Extract assistant text
                var blocks = (result && result.blocks) ? result.blocks : [];
                var text = blocks
                    .filter(function(b) { return b.payload && b.payload.text; })
                    .map(function(b) { return b.payload.text; })
                    .join("\n")
                    .trim();

                if (!text) {
                    throw new Error("No extraction output from LLM");
                }

                return JSON.parse(text);
            })
        };
    }
});
```

### Step 3: Use the Database Module

```javascript
var db = require("runnerdb");

// In your plugin:
create(ctx) {
    return {
        run: wrapExtractorRun(function(input) {
            // Query existing data
            var rows = db.query("SELECT * FROM some_table WHERE key = ?", ["value"]);

            // Insert results
            db.exec(
                "INSERT INTO extractions (run_id, data) VALUES (?, ?)",
                [ctx.runId, JSON.stringify(extractedData)]
            );

            return extractedData;
        })
    };
}
```

### Step 4: Run It

```bash
cozo-relationship-js-runner extract \
    scripts/my_extractor.js \
    demo_text.txt \
    --ai-api-type openai \
    --ai-engine gpt-4.1-mini \
    --stream-events \
    --pretty
```

## 6.8 Go-JS Data Type Mapping

| Go Type | JavaScript Type | Notes |
|---|---|---|
| `string` | `string` | Direct mapping |
| `int`, `int64` | `number` | Precision loss for very large ints |
| `float64` | `number` | Direct mapping |
| `bool` | `boolean` | Direct mapping |
| `nil` | `null` | Direct mapping |
| `map[string]any` | `Object` | Direct mapping |
| `[]any` | `Array` | Direct mapping |
| `func(...)` | `function` | Go functions callable from JS |
| `struct` | `Object` | Exported fields become properties |

### Calling Go from JS

When Go registers a function:
```go
vm.Set("myGoFunc", func(call goja.FunctionCall) goja.Value {
    name := call.Argument(0).String()
    return vm.ToValue("Hello, " + name)
})
```

JS can call it:
```javascript
var result = myGoFunc("world");  // "Hello, world"
```

### Calling JS from Go

```go
// Get a JS function
runFn, ok := goja.AssertFunction(obj.Get("run"))
if !ok {
    return errors.New("run is not a function")
}

// Call it with arguments
result, err := runFn(obj, vm.ToValue(input), vm.ToValue(options))
if err != nil {
    return err
}

// Export the result to Go
goValue := result.Export()
```

## 6.9 Error Handling

### JS Errors in Go

When JS throws an error, Go receives it:
```go
result, err := runFn(obj, vm.ToValue(input))
if err != nil {
    // err contains the JS error message
    // Including stack trace if available
    fmt.Println("JS error:", err.Error())
}
```

### Go Errors in JS

When a Go-registered function returns an error, it becomes a JS exception:
```javascript
try {
    var result = someGoFunction();
} catch (e) {
    console.error("Go function failed:", e.message);
}
```

### Panic Recovery

The `runtimeowner.Runner` wraps JS execution with panic recovery:
```go
runner := runtimeowner.NewRunner(vm, loop, runtimeowner.Options{
    RecoverPanics: true,  // Catches Go panics, returns them as errors
})
```

---

## Exercises

### Exercise 6.1: Minimal Plugin

Write a minimal plugin that:
1. Follows the descriptor contract
2. Returns a JSON object with the transcript length and word count
3. Does NOT use any LLM -- just string processing

<details>
<summary>Solution</summary>

```javascript
module.exports = {
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "exercise.word-counter",
    name: "Word Counter (Exercise 6.1)",
    create: function(ctx) {
        return {
            run: function(input) {
                var text = input.transcript || "";
                var words = text.split(/\s+/).filter(function(w) { return w.length > 0; });
                var sentences = text.split(/[.!?]+/).filter(function(s) { return s.trim().length > 0; });

                return JSON.stringify({
                    characterCount: text.length,
                    wordCount: words.length,
                    sentenceCount: sentences.length,
                    averageWordLength: words.length > 0
                        ? Math.round(words.reduce(function(sum, w) { return sum + w.length; }, 0) / words.length * 10) / 10
                        : 0
                });
            }
        };
    }
};
```
</details>

### Exercise 6.2: Name Extractor

Write a plugin that extracts person names from text using simple heuristics (capitalized words, "Mr./Ms./Dr." prefixes). No LLM needed.

<details>
<summary>Solution</summary>

```javascript
module.exports = {
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "exercise.name-extractor",
    name: "Simple Name Extractor (Exercise 6.2)",
    create: function() {
        return {
            run: function(input) {
                var text = input.transcript || "";
                var namePattern = /(?:(?:Mr|Ms|Mrs|Dr|Prof)\.?\s+)?([A-Z][a-z]+(?:\s+[A-Z][a-z]+)+)/g;
                var names = {};
                var match;

                while ((match = namePattern.exec(text)) !== null) {
                    var name = match[1] || match[0];
                    name = name.replace(/^(?:Mr|Ms|Mrs|Dr|Prof)\.?\s+/, '');
                    if (name.split(/\s+/).length >= 2) {
                        var id = name.toLowerCase().replace(/\s+/g, '_');
                        if (!names[id]) {
                            names[id] = { id: id, name: name, mentions: 0 };
                        }
                        names[id].mentions++;
                    }
                }

                var persons = [];
                for (var id in names) {
                    persons.push(names[id]);
                }
                persons.sort(function(a, b) { return b.mentions - a.mentions; });

                return JSON.stringify({ persons: persons });
            }
        };
    }
};
```
</details>

### Exercise 6.3: Plugin with Database

Write a plugin that:
1. Extracts word frequency from the transcript
2. Stores the results in a SQLite table via `require("runnerdb")`
3. Returns the top 10 most frequent words

### Exercise 6.4: Using Globals

Write a plugin that reads all the `RELATIONSHIP_*` global variables and returns them as a diagnostic object, without doing any extraction. This is useful for debugging the runtime environment.

<details>
<summary>Solution</summary>

```javascript
module.exports = {
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "exercise.env-inspector",
    name: "Environment Inspector (Exercise 6.4)",
    create: function(ctx) {
        return {
            run: function(input) {
                return JSON.stringify({
                    hostContext: ctx,
                    globals: {
                        runId: typeof RELATIONSHIP_RUN_ID !== 'undefined' ? RELATIONSHIP_RUN_ID : null,
                        profile: typeof RELATIONSHIP_PROFILE !== 'undefined' ? RELATIONSHIP_PROFILE : null,
                        prompt: typeof RELATIONSHIP_PROMPT !== 'undefined'
                            ? RELATIONSHIP_PROMPT.substring(0, 100) + '...'
                            : null,
                        timeoutMs: typeof RELATIONSHIP_TIMEOUT_MS !== 'undefined' ? RELATIONSHIP_TIMEOUT_MS : null,
                        transcriptLength: typeof RELATIONSHIP_TRANSCRIPT !== 'undefined'
                            ? RELATIONSHIP_TRANSCRIPT.length
                            : null,
                        scriptRoot: typeof RELATIONSHIP_SCRIPT_ROOT !== 'undefined' ? RELATIONSHIP_SCRIPT_ROOT : null,
                    },
                    input: {
                        transcriptLength: (input.transcript || '').length,
                        promptLength: (input.prompt || '').length,
                        profile: input.profile,
                        timeoutMs: input.timeoutMs,
                        engineOptionsKeys: input.engineOptions ? Object.keys(input.engineOptions) : []
                    }
                }, null, 2);
            }
        };
    }
};
```
</details>

### Exercise 6.5: Reflective Extractor

Study `relation_extractor_reflective.js` in the repo. How does it differ from the base template? What additional middleware or processing does the "reflective" variant add?

---

## Key Takeaways

1. **Goja** is a pure-Go JavaScript runtime -- no Node.js dependency
2. **`require()`** works via the `goja_nodejs` package with configurable search paths
3. **Native modules** let Go code appear as regular `require()` targets
4. **Global variables** pass configuration from Go CLI into JS runtime
5. **Geppetto** provides LLM access (engines, builders, sessions, turns)
6. **Plugin descriptor pattern**: `apiVersion` + `kind` + `id` + `name` + `create()` -> `run()`
7. **Structured output** enforces JSON Schema on LLM responses
8. Use **ES5.1 syntax** (no arrow functions, no `const`/`let` in older Goja, no async/await)

## Next Module

In **Module 7: Capstone Project**, you'll put everything together to build a complete extraction pipeline from scratch.
