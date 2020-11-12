package web

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openshift/osin"

	"github.com/liut/staffio/pkg/models"
	"github.com/liut/staffio/pkg/models/oauth"
)

// Authorization code endpoint
func (s *server) oauth2Authorize(c *gin.Context) {
	resp := s.osvr.NewResponse()
	defer resp.Close()

	r := c.Request
	user := UserWithContext(c)
	store := s.service.OSIN()

	if ar := s.osvr.HandleAuthorizeRequest(resp, r); ar != nil {
		logger().Debugw("HandleAuthorizeRequest", "client", ar.Client)
		if store.IsAuthorized(ar.Client.GetId(), user.UID) {
			ar.UserData = oauth.JSONKV{"uid": user.UID}
			ar.Authorized = true
			s.osvr.FinishAuthorizeRequest(resp, r, ar)
		} else {
			if r.Method == "GET" {
				scopes, err := store.LoadScopes()
				if err != nil {
					c.AbortWithError(404, err)
					return
				}
				s.Render(c, "authorize.html", map[string]interface{}{
					"link":          r.RequestURI,
					"response_type": ar.Type,
					"scopes":        scopes,
					"client":        ar.Client.(*oauth.Client),
					"ctx":           c,
				})
				return
			}

			if r.PostForm.Get("authorize") == "1" {
				ar.UserData = oauth.JSONKV{"uid": user.UID}
				ar.Authorized = true
				s.osvr.FinishAuthorizeRequest(resp, r, ar)
				if r.PostForm.Get("remember") != "" {
					err := store.SaveAuthorized(ar.Client.GetId(), user.UID)
					if err != nil {
						logger().Infow("SaveAuthorized fail", "err", err)
					}
				}
			} else {
				resp.SetRedirect("/")
			}

		}

	}

	if resp.IsError && resp.InternalError != nil {
		logger().Infow("authorize ERROR", "err", resp.InternalError)
	}
	// if !resp.IsError {
	// 	resp.Output["uid"] = c.user.UID
	// }

	logger().Debugw("oauthAuthorize", "resp", resp)
	osin.OutputJSON(resp, c.Writer, r)
}

// Access token endpoint
func (s *server) oauth2Token(c *gin.Context) {
	resp := s.osvr.NewResponse()
	defer resp.Close()
	r := c.Request

	var (
		uid   string
		user  *User
		staff *models.Staff
		err   error
	)
	if ar := s.osvr.HandleAccessRequest(resp, r); ar != nil {
		logger().Debugw("HandleAccessRequest", "code", ar.Code, "scope", ar.Scope)
		switch ar.Type {
		case osin.AUTHORIZATION_CODE:
			kv, _ := oauth.ToJSONKV(ar.UserData)
			if v, ok := kv["uid"]; ok {
				uid = v.(string)
			}

			staff, err = s.service.Get(uid)
			if err != nil {
				resp.SetError("get_user_error", "staff not found")
				resp.InternalError = err
			} else {
				user = UserFromStaff(staff)
			}
			ar.Authorized = true
		case osin.REFRESH_TOKEN:
			ar.UserData = nil
			// TODO: load refresh
			ar.Authorized = true
		case osin.PASSWORD:
			var staff *models.Staff
			if staff, err = s.service.Authenticate(ar.Username, ar.Password); err != nil {
				resp.SetError("authentication_failed", err.Error())
				break
			}
			ar.Authorized = true
			ar.UserData = oauth.JSONKV{"uid": staff.UID}
			user = UserFromStaff(staff)

		case osin.CLIENT_CREDENTIALS:
			ar.UserData = nil
			ar.Authorized = true
		case osin.ASSERTION:
			ar.UserData = nil
			if ar.AssertionType == "urn:osin.example.complete" && ar.Assertion == "osin.data" {
				ar.Authorized = true
			}
		}
		s.osvr.FinishAccessRequest(resp, r, ar)
	}

	if resp.IsError && resp.InternalError != nil {
		logger().Infow("token ERROR", "err", resp.InternalError)
	}
	if !resp.IsError {
		if uid != "" {
			resp.Output["uid"] = uid
			resp.Output["is_keeper"] = s.IsKeeper(uid)
		}
		if user != nil {
			resp.Output["user"] = user
		}

	}

	logger().Infow("oauthToken", "resp", resp)

	osin.OutputJSON(resp, c.Writer, r)
}

// Information endpoint
func (s *server) oauth2Info(c *gin.Context) {
	resp := s.osvr.NewResponse()
	defer resp.Close()
	r := c.Request

	if ir := s.osvr.HandleInfoRequest(resp, r); ir != nil {
		logger().Debugw("HandleInfoRequest", "code", ir.Code, "accessToken", ir.AccessData.AccessToken)
		var (
			uid   string
			topic = c.Param("topic")
		)
		logger().Infow("param", "topic", topic)
		kv, _ := oauth.ToJSONKV(ir.AccessData.UserData)
		if v, ok := kv["uid"]; ok {
			uid = v.(string)
		}
		staff, err := s.service.Get(uid)
		if err != nil {
			resp.SetError("get_user_error", "staff not found")
			resp.InternalError = err
		} else {
			resp.Output["uid"] = uid
			if strings.HasPrefix(topic, "me") {
				resp.Output["me"] = staff
				if len(topic) > 3 && topic[2] == '+' {
					if arr := strings.Split(topic[3:], "+"); len(arr) > 0 {
						logger().Infow("search groups", "arr", arr)
						gm := make(map[string]interface{})
						for _, gn := range arr {
							if gn == "" {
								continue
							}
							gm[gn] = s.InGroup(gn, uid)
						}
						logger().Infow("result", "gm", gm)
						resp.Output["group"] = gm
					}
				} else if len(topic) > 3 && topic[2] == '|' {
					if arr := strings.Split(topic[3:], "|"); len(arr) > 0 {
						var roles []string
						for _, gn := range arr {
							if s.InGroup(gn, uid) {
								roles = append(roles, gn)
							}
						}
						logger().Infow("result", "roles", roles)
						resp.Output["group"] = roles
					}
				}

			} else if topic == "staff" {
				resp.Output["staff"] = staff
			} else if topic == "grafana" || topic == "generic" {
				resp.Output["name"] = staff.GetName()
				resp.Output["login"] = staff.UID
				resp.Output["username"] = staff.UID
				resp.Output["email"] = staff.Email
				resp.Output["attributes"] = map[string][]string{} // TODO: fill attributes
			}

		}
		s.osvr.FinishInfoRequest(resp, r, ir)
	}

	if resp.IsError && resp.InternalError != nil {
		logger().Infow("info ERROR", "err", resp.InternalError)
	}

	osin.OutputJSON(resp, c.Writer, r)
}
