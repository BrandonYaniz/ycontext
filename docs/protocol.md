# Protocol

`ycontextd` uses JSON Lines over a Unix socket.

Each request and event is one JSON object followed by a newline.

## Request

```json
{"id":"req_1","method":"corpus.create","params":{"name":"Moby Dick"}}
```

## Response

```json
{"id":"req_1","ok":true,"result":{"corpus_id":"cor_..."}}
```

## Error

```json
{"id":"req_1","ok":false,"error":{"code":"invalid_request","message":"missing corpus name"}}
```

## Job events

Long-running methods may return events before the final response.

```json
{"id":"req_2","event":"job.created","job_id":"job_..."}
{"id":"req_2","event":"job.progress","stage":"chunking","percent":20}
{"id":"req_2","ok":true,"result":{"job_id":"job_...","status":"complete"}}
```

## Method naming

Use lower-case dotted method names:

```text
workspace.create
knowledge_set.create
corpus.create
source.add_text
source.add_file
ingest.start
job.get
job.list
node.get
tree.get
context.assemble
context.render
```

Keep the wire format boring. Avoid clever protocol features until the daemon needs them.
