# Configuration

Configuration uses YAML.

Default user config:

```text
~/.config/ycontext/ycontext.yaml
```

Example:

```yaml
server:
  socket_path: "~/.local/run/ycontext/ycontextd.sock"

store:
  database_path: "~/.local/share/ycontext/ycontext.db"
  document_path: "~/.local/share/ycontext/documents"

llm:
  provider: "yllmd"
  socket_path: "~/.local/run/yllmd/yllmd.sock"
  model: "local-default"

jobs:
  workers: 1
  idle_validation: true
```

System mode is opt-in and should be configured explicitly.
