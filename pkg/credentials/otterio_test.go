package credentials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvOtterIOIsolationAndFallback(t *testing.T) {
	for _, name := range []string{"OTTERIO_ROOT_USER", "OTTERIO_ROOT_PASSWORD", "OTTERIO_ACCESS_KEY", "OTTERIO_SECRET_KEY", "MINIO_ROOT_USER", "MINIO_ROOT_PASSWORD", "MINIO_ACCESS_KEY", "MINIO_SECRET_KEY"} {
		t.Setenv(name, "")
	}
	t.Setenv("MINIO_ROOT_USER", "upstream-user")
	t.Setenv("MINIO_ROOT_PASSWORD", "upstream-secret")
	provider := &EnvOtterIO{}
	value, err := provider.RetrieveWithCredContext(nil)
	if err != nil || value.SignerType != SignatureAnonymous {
		t.Fatalf("must not adopt MinIO credentials: %#v, %v", value, err)
	}
	t.Setenv("OTTERIO_ACCESS_KEY", "legacy-user")
	t.Setenv("OTTERIO_SECRET_KEY", "legacy-secret")
	value, err = provider.RetrieveWithCredContext(nil)
	if err != nil || value.AccessKeyID != "legacy-user" || value.SecretAccessKey != "legacy-secret" {
		t.Fatalf("legacy OtterIO pair: %#v, %v", value, err)
	}
	t.Setenv("OTTERIO_ROOT_USER", "root-user")
	t.Setenv("OTTERIO_ROOT_PASSWORD", "root-secret")
	value, err = provider.RetrieveWithCredContext(nil)
	if err != nil || value.AccessKeyID != "root-user" || value.SecretAccessKey != "root-secret" {
		t.Fatalf("root pair must take priority: %#v, %v", value, err)
	}
	old, err := (&EnvMinio{}).RetrieveWithCredContext(nil)
	if err != nil || old.AccessKeyID != "upstream-user" {
		t.Fatalf("legacy provider changed: %#v, %v", old, err)
	}
}

func TestFileOCDefaultAndExplicitConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("OC_SHARED_CREDENTIALS_FILE", "")
	t.Setenv("OC_ALIAS", "store")
	for _, dir := range []string{".oc", "oc"} {
		if err := os.MkdirAll(filepath.Join(home, dir), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, dir, "config.json"), []byte(`{"version":"10","aliases":{"store":{"url":"http://localhost:9000","accessKey":"oc-user","secretKey":"oc-secret","api":"S3v4"}}}`), 0600); err != nil {
			t.Fatal(err)
		}
	}
	value, err := (&FileOC{}).RetrieveWithCredContext(nil)
	if err != nil || value.AccessKeyID != "oc-user" || value.SecretAccessKey != "oc-secret" {
		t.Fatalf("default OC config: %#v, %v", value, err)
	}
	file := filepath.Join(home, "explicit.json")
	if err := os.WriteFile(file, []byte(`{"version":"10","aliases":{"other":{"accessKey":"explicit-user","secretKey":"explicit-secret","api":"S3v4"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	value, err = (&FileOC{Filename: file, Alias: "other"}).RetrieveWithCredContext(nil)
	if err != nil || value.AccessKeyID != "explicit-user" {
		t.Fatalf("explicit config must override environment: %#v, %v", value, err)
	}
}
