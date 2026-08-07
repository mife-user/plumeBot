package onebot

import (
	"errors"
	"testing"

	"plumebot/internal/domain"
)

func TestDecideReply(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"限流超时", domain.ErrRateLimited, rateLimitedReply},
		{"敏感词命中", &domain.SensitiveWordError{Word: "敏感"}, sensitiveWordReply},
		{"其他错误", errors.New("boom"), ""},
		{"无错误", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := decideReply(tc.err); got != tc.want {
				t.Errorf("decideReply(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}
