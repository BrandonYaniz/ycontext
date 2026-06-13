# ycontextd

`ycontextd` is the local daemon. It listens on a Unix socket, accepts JSON Lines requests, writes to SQLite and the document store, manages jobs, and calls yllmd for LLM work.

Do not add direct model serving here. ycontextd uses yllmd.
