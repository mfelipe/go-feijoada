package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Semver is a custom type for semantic versioning. https://semver.org/
type Semver struct {
	Major uint `json:"MAJOR"` // version when you make incompatible API changes
	Minor uint `json:"MINOR"` // version when you add functionality in a backward compatible manner
	Patch uint `json:"PATCH"` // version when you make backward compatible bug fixes
}

func (s *Semver) String() string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

// UnmarshalParam Binds the version parameter into Semver format.
// MINOR and PATCH are optionals, defaults to zero
func (s *Semver) UnmarshalParam(param string) error {
	split := strings.Split(param, ".")

	if len(split[0]) == 0 || len(split) > 3 {
		return &json.UnmarshalTypeError{Value: param, Type: reflect.TypeOf(s)}
	}

	intVersions := make([]uint, 3)
	for i, sv := range split {
		iv, err := strconv.ParseUint(sv, 10, 10)
		if err != nil {
			sT := reflect.TypeOf(Semver{})
			f := sT.Field(i)
			return &json.UnmarshalTypeError{Value: sv, Type: sT, Field: f.Name, Struct: f.Type.String()}
		}
		intVersions[i] = uint(iv)
	}

	*s = Semver{Major: intVersions[0], Minor: intVersions[1], Patch: intVersions[2]}
	return nil
}
