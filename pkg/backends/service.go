package backends

import (
	"log"

	"lcgc/platform/staffio/pkg/backends/ldap"
	"lcgc/platform/staffio/pkg/models"
	"lcgc/platform/staffio/pkg/models/cas"
	"lcgc/platform/staffio/pkg/models/common"
	. "lcgc/platform/staffio/pkg/settings"
)

type Servicer interface {
	models.Authenticator
	models.StaffStore
	models.PasswordStore
	models.GroupStore
	cas.TicketStore
	OSIN() OSINStore
	CloseAll()
	StoreStaff(*models.Staff) error
	InGroup(gn, uid string) bool
	ProfileModify(uid, password string, staff *models.Staff) error
	PasswordForgot(at common.AliasType, target, uid string) error
	PasswordResetTokenVerify(token string) (uid string, err error)
	PasswordResetWithToken(login, token, passwd string) (err error)
}

type serviceImpl struct {
	*ldap.LDAPStore
	osinStore *DbStorage
}

var _ Servicer = (*serviceImpl)(nil)

func NewService() *serviceImpl {

	cfg := &ldap.Config{
		Addr:   Settings.LDAP.Hosts,
		Base:   Settings.LDAP.Base,
		Bind:   Settings.LDAP.BindDN,
		Passwd: Settings.LDAP.Password,
	}
	store, err := ldap.NewStore(cfg)
	if err != nil {
		log.Fatalf("new service ERR %s", err)
	}
	// LDAP is a special store
	return &serviceImpl{
		LDAPStore: store,
		osinStore: NewStorage(),
	}

}

func (s *serviceImpl) OSIN() OSINStore {
	return s.osinStore
}

func (s *serviceImpl) CloseAll() {

}