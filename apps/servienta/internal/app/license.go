package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/servienta/servienta/apps/servienta/internal/license"
)

// FreeStands is what the engine runs without a license file (D10, closed by
// D16): the HTTP file server alone — enough for the import-check use case.
var FreeStands = []string{"http"}

// LicenseStatus is reported over the contract (R12) and read by the console.
type LicenseStatus struct {
	Mode      string   `json:"mode"` // "free" | "licensed"
	Stands    []string `json:"stands"`
	Customer  string   `json:"customer,omitempty"`
	ExpiresAt int64    `json:"expires_at,omitempty"`
	Error     string   `json:"error,omitempty"` // set when a mounted license was rejected
}

// resolveLicense decides which stands to run. A missing file means Free mode;
// a present-but-invalid file is refused explicitly (R12: never silent
// degradation) — the engine runs Free and reports the error.
func resolveLicense(path, pubKey string) LicenseStatus {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("license read failed", "err", err)
		}
		return LicenseStatus{Mode: "free", Stands: FreeStands}
	}
	var f license.File
	if err := jsonUnmarshal(data, &f); err != nil {
		return LicenseStatus{Mode: "free", Stands: FreeStands, Error: "license file is not valid JSON"}
	}
	p, err := license.Verify(f, pubKey, time.Now())
	if err != nil {
		return LicenseStatus{Mode: "free", Stands: FreeStands, Error: err.Error()}
	}
	return LicenseStatus{Mode: "licensed", Stands: p.Stands, Customer: p.Name, ExpiresAt: p.Exp}
}
