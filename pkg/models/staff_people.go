package models

import (
	schema "github.com/liut/staffio-backend/model"
	"github.com/liut/staffio/pkg/common"
)

type Gender = common.Gender

const (
	Unknown = common.Unknown
	Male    = common.Male
	Female  = common.Female
)

type Staff = schema.People
type Staffs = schema.Peoples
type Spec = schema.Spec
