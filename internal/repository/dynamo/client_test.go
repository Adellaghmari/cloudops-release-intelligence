package dynamo

import (
	"context"
	"testing"
)

func TestUsesLocalStaticCredentials(t *testing.T) {
	t.Parallel()
	if usesLocalStaticCredentials("") {
		t.Fatal("empty endpoint must use the AWS default credential chain")
	}
	if usesLocalStaticCredentials("   ") {
		t.Fatal("whitespace-only endpoint must use the AWS default credential chain")
	}
	if !usesLocalStaticCredentials("http://127.0.0.1:8000") {
		t.Fatal("explicit local endpoint must select DynamoDB Local static credentials")
	}
}

func TestLoadAWSConfigProductionUsesEnvCredentialChain(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIATESTPRODONLY")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-for-unit-test-only")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "")
	t.Setenv("AWS_CONFIG_FILE", "")

	cfg, err := loadAWSConfig(context.Background(), ClientOptions{
		Region:   "eu-west-1",
		Endpoint: "",
	})
	if err != nil {
		t.Fatalf("loadAWSConfig: %v", err)
	}
	if cfg.Region != "eu-west-1" {
		t.Fatalf("region=%q", cfg.Region)
	}
	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if creds.AccessKeyID == "local" || creds.SecretAccessKey == "local" {
		t.Fatal("production path must not inject static local/local credentials")
	}
	if creds.AccessKeyID != "AKIATESTPRODONLY" {
		t.Fatalf("expected default chain env credentials, got access key %q", creds.AccessKeyID)
	}
}

func TestLoadAWSConfigLocalEndpointUsesStaticLocal(t *testing.T) {
	// Even if env credentials exist, DynamoDB Local must use dummy local/local.
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIATESTSHOULDNOTWIN")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-for-unit-test-only")

	cfg, err := loadAWSConfig(context.Background(), ClientOptions{
		Region:   "eu-west-1",
		Endpoint: "http://127.0.0.1:8000",
	})
	if err != nil {
		t.Fatalf("loadAWSConfig: %v", err)
	}
	if cfg.Region != "eu-west-1" {
		t.Fatalf("region=%q", cfg.Region)
	}
	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if creds.AccessKeyID != "local" || creds.SecretAccessKey != "local" {
		t.Fatalf("local endpoint must use static local/local, got access=%q", creds.AccessKeyID)
	}
}

func TestLoadAWSConfigDefaultRegion(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIATESTPRODONLY")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-for-unit-test-only")
	t.Setenv("AWS_SESSION_TOKEN", "")

	cfg, err := loadAWSConfig(context.Background(), ClientOptions{})
	if err != nil {
		t.Fatalf("loadAWSConfig: %v", err)
	}
	if cfg.Region != "eu-west-1" {
		t.Fatalf("expected default region eu-west-1, got %q", cfg.Region)
	}
}
