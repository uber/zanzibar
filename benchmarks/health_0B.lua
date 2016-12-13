-- wrk -t12 -c400 -d30s -s ./benchmarks/health_0B.lua http://localhost:8093/health
-- go-torch -u http://localhost:8093/ -t5
wrk.method = "GET"

