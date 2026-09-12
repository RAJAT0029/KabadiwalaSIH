package users

import (
	"testing"
	"time"
)

func TestParticipationPolicy(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	cases := []struct {
		name, role, status  string
		active, demo, allow bool
		from, until         *time.Time
		want                bool
	}{
		{"authorized", "recycler", "authorized", true, false, false, nil, &future, true},
		{"demo explicitly allowed", "recycler", "demo_verified", true, true, true, nil, nil, true},
		{"demo in production", "recycler", "demo_verified", true, true, false, nil, nil, false},
		{"unmarked demo", "recycler", "demo_verified", true, false, true, nil, nil, false},
		{"demo claims authorized", "recycler", "authorized", true, true, false, nil, nil, false},
		{"collector", "kabadiwala", "authorized", true, false, true, nil, nil, false},
		{"inactive", "recycler", "authorized", false, false, false, nil, nil, false},
		{"expired", "recycler", "authorized", true, false, false, nil, &past, false},
		{"expiry boundary", "recycler", "authorized", true, false, false, nil, &now, false},
		{"future validity", "recycler", "authorized", true, false, false, &future, nil, false},
		{"pending", "recycler", "pending_verification", true, false, true, nil, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u := User{Role: c.role, IsActive: c.active, Authorization: Authorization{Status: c.status, IsDemo: c.demo, ValidFrom: c.from, ValidUntil: c.until}}
			if u.CanParticipate(now, c.allow) != c.want {
				t.Fatal("incorrect participation decision")
			}
		})
	}
}

func TestProfileCannotWriteVerificationFields(t *testing.T) {
	now := time.Now()
	for _, a := range []Authorization{{Status: "authorized"}, {IsDemo: true}, {ValidFrom: &now}, {ValidUntil: &now}} {
		if ValidateProfilePatch(ProfilePatch{Authorization: &a}) == nil {
			t.Fatal("verification field accepted")
		}
	}
	if err := ValidateProfilePatch(ProfilePatch{Authorization: &Authorization{RegistrationNumber: "submitted reference"}}); err != nil {
		t.Fatal(err)
	}
}
