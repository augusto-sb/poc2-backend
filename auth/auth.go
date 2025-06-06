package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type IntrospectionResponse = struct {
	Active         bool `json:"active"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

var introspectionEndpoint string = "https://example.com/auth/realms/REALM/protocol/openid-connect/token/introspect"
var privateClientId string = "client"
var privateClientSecret string = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
var httpClient http.Client = http.Client{}
var skipAuth bool = false

func init() {
	//using rbac for simplicity, could use resource/scope based authorization for large project
	introspectionEndpoint = os.Getenv("AUTH_INTROSPECTION_ENDPOINT")
	privateClientId = os.Getenv("AUTH_CLIENT_ID")
	privateClientSecret = os.Getenv("AUTH_CLIENT_SECRET")
	if os.Getenv("AUTH_ENABLED") == "false" {
		skipAuth = true
	}
}

func Middleware(next http.HandlerFunc, role string) http.HandlerFunc {
	if skipAuth {
		return next
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		//req.Header = map[string][]string
		authHeader, ok := req.Header["Authorization"]
		if !ok {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Authorization header not present"))
			return
		}
		if len(authHeader) != 1 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Authorization header present multiple times"))
			return
		}
		bearer := authHeader[0]
		if strings.HasPrefix(strings.ToLower(bearer), "bearer ") {
			bearer = bearer[7:len(bearer)]
		}
		//some validation
		jwt := strings.Split(bearer, ".")
		if len(jwt) != 3 {
			// header, payload, signature
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("invalid JWT"))
			return
		}
		//introspect
		data := url.Values{}
		data.Set("token", bearer)
		data.Set("client_id", privateClientId)
		data.Set("client_secret", privateClientSecret)
		kcReq, err := http.NewRequest(http.MethodPost, introspectionEndpoint, strings.NewReader(data.Encode()))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		kcReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		kcReq.Host = "localhost:8080" //fake its not a direct request
		resp, err := httpClient.Do(kcReq)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != http.StatusOK {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		var respJson IntrospectionResponse
		err = json.Unmarshal(bodyBytes, &respJson)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		//kc tells if token is active and decodes it (base64 payload) to a json
		if !respJson.Active {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("inactive token"))
			return
		}
		var hasAccess bool = false
		for k, v := range respJson.ResourceAccess {
			if k == privateClientId {
				for _, v2 := range v.Roles {
					if v2 == role {
						hasAccess = true
						break
					}
				}
			}
		}
		if !hasAccess {
			rw.WriteHeader(http.StatusForbidden)
			rw.Write([]byte("role required: " + role))
			return
		}
		next.ServeHTTP(rw, req)
	})
}
