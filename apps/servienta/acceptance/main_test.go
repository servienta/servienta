package acceptance

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// go-ipfix logs via klog; silence it in tests.
	_ = flag.Set("logtostderr", "false")
	_ = flag.Set("stderrthreshold", "FATAL")
	os.Exit(m.Run())
}
