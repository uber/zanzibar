-- wrk -t12 -c400 -d30s -s ./benchmarks/contacts_10B.lua http://localhost:8093/contacts/foo/contacts
-- go-torch -u http://localhost:8093/ -t5
wrk.method = "POST"
wrk.body = "{\"userUUID\":\"some-uuid\", \"contacts\":[]}"

