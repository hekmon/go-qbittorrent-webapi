package qbtapi

import (
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/hekmon/cunits/v3"
)

// String returns a pointer to the string value passed in.
// Useful for the many *string fields in the API model.
func String(value string) *string {
	return &value
}

// Int returns a pointer to the int value passed in.
// Useful for the many *int fields in the API model.
func Int(value int) *int {
	return &value
}

// Bool returns a pointer to the bool value passed in.
// Useful for the many *bool fields in the API model.
func Bool(value bool) *bool {
	return &value
}

var (
	unlimitedSpeedLimit = Speed{cunits.Speed{Bits: math.MaxUint64}}
)

// Speed wraps cunits.Speed to provide a custom String() method to handle "unlimited" value representation.
type Speed struct {
	cunits.Speed
}

func (s Speed) Unlimited() bool {
	return s == unlimitedSpeedLimit
}

func (s Speed) ToBytes() int {
	if s.Unlimited() {
		return -1
	}
	return int(s.Bytes())
}

func (s Speed) String() string {
	if s.Unlimited() {
		return "unlimited"
	}
	return s.Speed.String()
}

// GetSpeedFromBytes is an helper to get the Speed type from integer bytes.
// It handles the special value -1 as unlimited.
func GetSpeedFromBytes(bytes int) Speed {
	switch bytes {
	case -1:
		return unlimitedSpeedLimit
	default:
		return Speed{cunits.Speed{Bits: cunits.ImportInBytes(float64(bytes))}}
	}
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func createBtFormFile(w *multipart.Writer, filename string) (io.Writer, error) {
	h := make(textproto.MIMEHeader, 2)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="torrents"; filename="%s"`, quoteEscaper.Replace(filename)))
	h.Set("Content-Type", "application/x-bittorrent")
	return w.CreatePart(h)
}
