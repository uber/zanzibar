-- wrk -t12 -c400 -d30s -s ./benchmarks/baz_call_16B.lua http://localhost:8093/baz/call
-- go-torch -u http://localhost:8093/ -t5
wrk.method = "POST"
wrk.body = "{\"arg\":{\"b1\":true,\"s2\":\"hello\",\"i3\":42}}"
wrk.headers["x-uuid"] = "some-uuid"
wrk.headers["x-token"] = "some-token"
