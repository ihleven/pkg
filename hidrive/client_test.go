package hidrive

import (
	"fmt"
	"testing"
)

func TestHome(t *testing.T) {

	oap := NewAuthProvider("", "")
	if oap.Token == nil {
		t.Errorf("oap.Token == nil, want it to be a Token")
	}

	hidrive := NewClient()

	if hidrive == nil {
		t.Errorf("hidrive == nil, want it to be not nil")
	}
	hidrive.GetDir("/users/matt.ihle/data")

}

func TestShare(t *testing.T) {

	hidrive := NewClient(10, nil)

	if hidrive == nil {
		t.Errorf("hidrive == nil, want it to be not nil")
	}
	share, err := hidrive.GetShare("")
	fmt.Println(share, err)
}

func TestShareToken(t *testing.T) {

	hidrive := NewClient(10, nil)

	if hidrive == nil {
		t.Errorf("hidrive == nil, want it to be not nil")
	}
	share, err := hidrive.GetShareToken("https://my.hidrive.com/share/j2mblrkm0w", "sylvia")
	fmt.Println(share, err)
}
