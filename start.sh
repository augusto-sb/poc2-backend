AUTH_INTROSPECTION_ENDPOINT="http://localhost:8080/auth/realms/poc/protocol/openid-connect/token/introspect" \
AUTH_CLIENT_ID="private" \
AUTH_CLIENT_SECRET="VaSyUQksKSDf1Suv1lLVrmxaRQNaczHW" \
CORS_ORIGIN="http://localhost:4200" \
PORT='8081' \
MONGODB_URI='mongodb://rootuser:rootpass@172.18.0.3:27017/?timeoutMS=5000&directConnection=true' \
CONTEXT_PATH='/backend' \
go run .