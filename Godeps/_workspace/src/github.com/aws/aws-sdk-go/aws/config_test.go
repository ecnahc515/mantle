package aws

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/credentials"
)

var testCredentials = credentials.NewStaticCredentials("AKID", "SECRET", "SESSION")

var copyTestConfig = Config{
	Credentials:             testCredentials,
	Endpoint:                String("CopyTestEndpoint"),
	Region:                  String("COPY_TEST_AWS_REGION"),
	DisableSSL:              Bool(true),
	HTTPClient:              http.DefaultClient,
	LogLevel:                LogLevel(LogDebug),
	Logger:                  NewDefaultLogger(),
	MaxRetries:              Int(DefaultRetries),
	DisableParamValidation:  Bool(true),
	DisableComputeChecksums: Bool(true),
	S3ForcePathStyle:        Bool(true),
}

func TestCopy(t *testing.T) {
	want := copyTestConfig
	got := copyTestConfig.Copy()
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("Copy() = %+v", got)
		t.Errorf("    want %+v", want)
	}
}

func TestCopyReturnsNewInstance(t *testing.T) {
	want := copyTestConfig
	got := copyTestConfig.Copy()
	if got == &want {
		t.Errorf("Copy() = %p; want different instance as source %p", got, &want)
	}
}

var mergeTestZeroValueConfig = Config{}

var mergeTestConfig = Config{
	Credentials:             testCredentials,
	Endpoint:                String("MergeTestEndpoint"),
	Region:                  String("MERGE_TEST_AWS_REGION"),
	DisableSSL:              Bool(true),
	HTTPClient:              http.DefaultClient,
	LogLevel:                LogLevel(LogDebug),
	Logger:                  NewDefaultLogger(),
	MaxRetries:              Int(10),
	DisableParamValidation:  Bool(true),
	DisableComputeChecksums: Bool(true),
	S3ForcePathStyle:        Bool(true),
}

var mergeTests = []struct {
	cfg  *Config
	in   *Config
	want *Config
}{
	{&Config{}, nil, &Config{}},
	{&Config{}, &mergeTestZeroValueConfig, &Config{}},
	{&Config{}, &mergeTestConfig, &mergeTestConfig},
}

func TestMerge(t *testing.T) {
	for i, tt := range mergeTests {
		got := tt.cfg.Merge(tt.in)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Config %d %+v", i, tt.cfg)
			t.Errorf("   Merge(%+v)", tt.in)
			t.Errorf("     got %+v", got)
			t.Errorf("    want %+v", tt.want)
		}
	}
}
