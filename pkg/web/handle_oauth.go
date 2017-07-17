package web

import (
	"log"
	"strings"

	"github.com/RangelReale/osin"
	"github.com/gin-gonic/gin"

	"lcgc/platform/staffio/pkg/models"
)

// Authorization code endpoint
func (s *server) oauth2Authorize(c *gin.Context) {
	resp := s.osvr.NewResponse()
	defer resp.Close()

	r := c.Request
	user := UserWithContext(c)
	store := s.service.OSIN()

	if ar := s.osvr.HandleAuthorizeRequest(resp, r); ar != nil {
		log.Printf("client: %v", ar.Client)
		if store.IsAuthorized(ar.Client.GetId(), user.Uid) {
			ar.UserData = user.Uid
			ar.Authorized = true
			s.osvr.FinishAuthorizeRequest(resp, r, ar)
		} else {
			if r.Method == "GET" {
				scopes, err := store.LoadScopes()
				if err != nil {
					c.AbortWithError(404, err)
					return
				}
				Render(c, "authorize.html", map[string]interface{}{
					"link":          r.RequestURI,
					"response_type": ar.Type,
					"scopes":        scopes,
					"client":        ar.Client.(*models.Client),
					"ctx":           c,
				})
			}

			if r.PostForm.Get("authorize") == "1" {
				ar.UserData = user.Uid
				ar.Authorized = true
				s.osvr.FinishAuthorizeRequest(resp, r, ar)
				if r.PostForm.Get("remember") != "" {
					err := store.SaveAuthorized(ar.Client.GetId(), user.Uid)
					if err != nil {
						log.Printf("remember ERR %s", err)
					}
				}
			} else {
				resp.SetRedirect("/")
			}

		}

	}

	if resp.IsError && resp.InternalError != nil {
		log.Printf("authorize ERROR: %s\n", resp.InternalError)
	}
	// if !resp.IsError {
	// 	resp.Output["uid"] = c.User.Uid
	// }

	debug("oauthAuthorize resp: %v", resp)
	osin.OutputJSON(resp, c.Writer, r)
}

// Access token endpoint
func (s *server) oauth2Token(c *gin.Context) {
	resp := s.osvr.NewResponse()
	defer resp.Close()
	r := c.Request

	var (
		uid   string = ""
		user  *User
		staff *models.Staff
		err   error
	)
	if ar := s.osvr.HandleAccessRequest(resp, r); ar != nil {
		debug("ar Code %s Scope %s", ar.Code, ar.Scope)
		switch ar.Type {
		case osin.AUTHORIZATION_CODE:
			uid = ar.UserData.(string)
			staff, err = s.service.Get(uid)
			if err != nil {
				resp.SetError("get_user_error", "staff not found")
				resp.InternalError = err
			} else {
				user = UserFromStaff(staff)
			}
			ar.Authorized = true
		case osin.REFRESH_TOKEN:
			// TODO: load refresh
			ar.Authorized = true
		case osin.PASSWORD:
			if err = s.service.Authenticate(ar.Username, ar.Password); err != nil {
				resp.SetError("authentication_failed", err.Error())
				break
			}
			staff, err := s.service.Get(ar.Username)
			if err != nil {
				// resp.InternalError = err
				resp.SetError("get_user_failed", err.Error())
				break
			}
			ar.Authorized = true
			ar.UserData = staff.Uid
			user = UserFromStaff(staff)

		case osin.CLIENT_CREDENTIALS:
			ar.Authorized = true
		case osin.ASSERTION:
			if ar.AssertionType == "urn:osin.example.complete" && ar.Assertion == "osin.data" {
				ar.Authorized = true
			}
		}
		s.osvr.FinishAccessRequest(resp, r, ar)
	}

	if resp.IsError && resp.InternalError != nil {
		log.Printf("token ERROR: %s\n", resp.InternalError)
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

	debug("oauthToken resp: %v", resp)

	osin.OutputJSON(resp, c.Writer, r)
}

// Information endpoint
func (s *server) oauth2Info(c *gin.Context) {
	resp := s.osvr.NewResponse()
	defer resp.Close()
	r := c.Request

	if ir := s.osvr.HandleInfoRequest(resp, r); ir != nil {
		debug("ir Code %s Token %s", ir.Code, ir.AccessData.AccessToken)
		var (
			uid   string
			topic = c.Param("topic")
		)
		log.Printf("topic %s", topic)
		uid = ir.AccessData.UserData.(string)
		staff, err := s.service.Get(uid)
		if err != nil {
			resp.SetError("get_user_error", "staff not found")
			resp.InternalError = err
		} else {
			resp.Output["uid"] = uid
			if strings.HasPrefix(topic, "me") {
				resp.Output["me"] = staff
				if len(topic) > 2 && strings.Index(topic, "+") == 2 {
					// TODO: search group topic[2:]
					gn := topic[3:]
					resp.Output[gn] = s.InGroup(gn, uid)
				}
			} else if topic == "staff" {
				resp.Output["staff"] = staff
			} else if topic == "grafana" || topic == "generic" {
				resp.Output["name"] = staff.Name()
				resp.Output["login"] = staff.Uid
				resp.Output["username"] = staff.Uid
				resp.Output["email"] = staff.Email
				resp.Output["attributes"] = map[string][]string{} // TODO: fill attributes
			}

		}
		s.osvr.FinishInfoRequest(resp, r, ir)
	}

	if resp.IsError && resp.InternalError != nil {
		log.Printf("info ERROR: %s\n", resp.InternalError)
	}

	osin.OutputJSON(resp, c.Writer, r)
}
