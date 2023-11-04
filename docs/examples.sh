HOST=${HOST:-localhost:8080}

curl -XPOST ${HOST}/users -H "Content-Type: application/json" -d '{
"name": "foo",
"email": "foo@example.com",
"password": "12345678"
}'

curl -XPOST ${HOST}/urls -H "Content-Type: application/json" -d '{
"original": "https://example.com/very/long/url/that/needs/to/be/shortened",
"user_id": 1
}'
curl -XPOST ${HOST}/urls -H "Content-Type: application/json" -d '{
"original": "https://www.granviaje.ch/travels-with-mitzi/",
"user_id": 1
}'
curl -XPOST ${HOST}/urls -H "Content-Type: application/json" -d '{
"original": "https://www.granviaje.ch/goodbye-brazil/",
"user_id": 1
}'

curl -XGET ${HOST}/s/WOjYdEcJyRv
curl -XGET ${HOST}/s/HxDxQxDHAVp