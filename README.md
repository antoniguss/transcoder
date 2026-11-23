# Transcoder: distributed media transcoding and streaming

Technologies used: Go

```bash
curl -F file=@nature.mp4 -F width=480 localhost:8080/upload   # → {"id": "abc123"}
curl localhost:8080/jobs/abc123                               # → {"state": "transcoding", "progress": 0.42}
curl -o out.mp4 localhost:8080/jobs/abc123/out                # → the file
```
