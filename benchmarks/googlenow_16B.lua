-- wrk -t12 -c400 -d30s -s ./benchmarks/googlenow_16B.lua http://localhost:8093/googlenow/add-credentials
-- go-torch -u http://localhost:8093/ -t5
wrk.method = "POST"
wrk.body = "{\"authCode\":\"deadbeef\"}"
wrk.headers["x-uuid"] = "some-uuid"
wrk.headers["x-token"] = "some-token"
