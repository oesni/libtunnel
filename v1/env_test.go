package v1

import (
	"testing"
)

type recordingEnv struct {
	vars map[string]string
	set  map[string]string
}

func (e *recordingEnv) LookupEnv(key string) (string, bool) {
	v, ok := e.vars[key]
	return v, ok
}

func (e *recordingEnv) Setenv(key, value string) error {
	e.set[key] = value
	return nil
}

func TestSetEnvironmentRoutesReadsAndWrites(t *testing.T) {
	t.Setenv(TokenEnv, "from-process")
	e := &recordingEnv{vars: map[string]string{TokenEnv: "from-installed"}, set: map[string]string{}}
	SetEnvironment(e)
	t.Cleanup(func() { SetEnvironment(nil) })

	if got := Token(); got != "from-installed" {
		t.Errorf("Token() = %q, want the installed environment's value", got)
	}
	if err := Setenv(SpecEnv, "spec"); err != nil {
		t.Fatal(err)
	}
	if e.set[SpecEnv] != "spec" {
		t.Errorf("Setenv did not reach the installed environment")
	}
}

func TestIsolatedReadsNothingAndWritesNothing(t *testing.T) {
	t.Setenv(TokenEnv, "inherited")
	t.Setenv(ActionsTokenEnv, "inherited-actions")
	t.Setenv(LocalURLEnv, "http://elsewhere.invalid")
	SetEnvironment(Isolated)
	t.Cleanup(func() { SetEnvironment(nil) })

	if got := Token(); got != "" {
		t.Errorf("Token() = %q under Isolated, want empty", got)
	}
	if v, ok := LookupEnv(LocalURLEnv); ok || v != "" {
		t.Errorf("LookupEnv(%s) = %q, %v under Isolated, want unset", LocalURLEnv, v, ok)
	}
	if err := Setenv(HostnameEnv, "x.example"); err != nil {
		t.Fatal(err)
	}
	SetEnvironment(nil)
	if got := Getenv(HostnameEnv); got == "x.example" {
		t.Errorf("Isolated Setenv reached the process environment")
	}
	if got := Token(); got != "inherited" {
		t.Errorf("Token() = %q after SetEnvironment(nil), want the process value back", got)
	}
}
