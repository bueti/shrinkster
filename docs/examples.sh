HOST=${HOST:-localhost:8080}

admin_id="63920346-70d0-40ec-8f53-f8d019628804"
non_privileged_id="a6ba3348-6bb3-49fa-81f2-87fb9f984b56"
owner_id="af138557-d47b-480a-917e-2a51437a13f7"

curl -XPOST ${HOST}/signup -H "Content-Type: application/json" -d '{
"name": "non-privileged user",
"email": "sad-user@example.com",
"password": "12345678",
"password_confirm": "12345678"
}'

token=$(curl -XPOST ${HOST}/login -H "Content-Type: application/json" -d '{"email": "foo@example.com","password": "12345678"}' | jq -r .token)
auth="-H \"Authorization: Bearer $token\""

curl -XPOST ${HOST}/urls -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d '{
"original": "https://example.com/very/long/url/that/needs/to/be/shortened",
"user_id": "3c9a1029-2fc3-4b26-a10b-e83ee5106188"
}'
curl -XPOST ${HOST}/urls $auth -H "Content-Type: application/json" -d '{
"original": "https://www.granviaje.ch/travels-with-mitzi/",
"user_id": "3c9a1029-2fc3-4b26-a10b-e83ee5106188"
}'
curl -XPOST ${HOST}/urls -H "Content-Type: application/json" -d '{
"original": "https://www.granviaje.ch/goodbye-brazil/",
"user_id": 1
}'
curl -XPOST ${HOST}/urls -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d '{
"original": "https://www.granviaje.ch/goodbye-brazil/",
"short_code": "foo",
"user_id": "63920346-70d0-40ec-8f53-f8d019628804"
}'

curl -XGET ${HOST}/s/audNtvP2MAm
curl -XGET ${HOST}/s/HxDxQxDHAVp

# get all urls for user
owner="63920346-70d0-40ec-8f53-f8d019628804"
not_owner="af138557-d47b-480a-917e-2a51437a13f7"
curl ${HOST}/urls/$not_owner -H "Authorization: Bearer $token" -H "Content-Type: application/json"