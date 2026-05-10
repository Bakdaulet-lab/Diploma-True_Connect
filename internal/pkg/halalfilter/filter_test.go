package halalfilter_test

import (
	"testing"

	"github.com/trueconnect/backend/internal/pkg/halalfilter"
)

func TestCheckMessage_Clean(t *testing.T) {
	blocked, warned := halalfilter.CheckMessage("Привет, как дела?")
	if blocked || warned {
		t.Errorf("clean message should not trigger filter: blocked=%v warned=%v", blocked, warned)
	}
}

func TestCheckMessage_Blocked(t *testing.T) {
	cases := []string{"секс", "порно", "sex with me", "PORN site"}
	for _, tc := range cases {
		blocked, _ := halalfilter.CheckMessage(tc)
		if !blocked {
			t.Errorf("expected blocked=true for %q", tc)
		}
	}
}

func TestCheckMessage_Warned(t *testing.T) {
	cases := []string{"badword here", "this is a scam", "abuse"}
	for _, tc := range cases {
		blocked, warned := halalfilter.CheckMessage(tc)
		if blocked {
			t.Errorf("expected blocked=false for %q", tc)
		}
		if !warned {
			t.Errorf("expected warned=true for %q", tc)
		}
	}
}

func TestCheckMessage_CaseInsensitive(t *testing.T) {
	blocked, _ := halalfilter.CheckMessage("ПОРНО")
	if !blocked {
		t.Error("filter must be case-insensitive")
	}
}
