// Package remote implements S3-compatible object stores (AWS S3, MinIO,
// Cloudflare R2, Backblaze B2) using DVC's content-addressed key layout.
package remote

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getlode/lode/internal/repo"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 is an S3-compatible remote.
type S3 struct {
	client *minio.Client
	bucket string
	prefix string // key prefix under the bucket (the <key> in s3://bucket/<key>)
}

// NewS3 builds an S3 store from a remote config.
func NewS3(r repo.Remote) (*S3, error) {
	if r.URL == "" {
		return nil, errors.New("remote has no url")
	}
	bucket, prefix, err := parseS3URL(r.URL)
	if err != nil {
		return nil, err
	}

	endpoint, secure := "s3.amazonaws.com", true
	if r.EndpointURL != "" {
		u, err := url.Parse(r.EndpointURL)
		if err != nil {
			return nil, err
		}
		endpoint = u.Host
		secure = u.Scheme != "http"
	}

	var creds *credentials.Credentials
	if r.AccessKeyID != "" {
		creds = credentials.NewStaticV4(r.AccessKeyID, r.SecretAccessKey, r.SessionToken)
	} else {
		// Predictable precedence: environment -> shared-credentials file (honoring
		// the configured profile / AWS_PROFILE) -> IAM (EC2 instance role, ECS, and
		// EKS/IRSA via the container/web-identity endpoints). The IAM provider is
		// given an explicit timeout-bounded HTTP client so that off-cloud (no
		// metadata endpoint) it fails fast instead of hanging — and never panics on
		// a nil client.
		creds = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.EnvAWS{},
			&credentials.FileAWSCredentials{Profile: r.Profile},
			&credentials.IAM{Client: &http.Client{Timeout: 5 * time.Second}},
		})
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: secure,
		Region: r.Region,
	})
	if err != nil {
		return nil, err
	}
	return &S3{client: client, bucket: bucket, prefix: prefix}, nil
}

// key maps an object id to its full S3 key: <prefix>/files/md5/<2>/<rest>.
func (s *S3) key(oid string) string {
	return path.Join(s.prefix, "files", "md5", oid[:2], oid[2:])
}

func (s *S3) Has(ctx context.Context, oid string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, s.key(oid), minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		resp := minio.ToErrorResponse(err)
		if resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3) ListOIDs(ctx context.Context) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	base := path.Join(s.prefix, "files", "md5") + "/"
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    base,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		// key tail is <2>/<rest>; reconstruct the oid.
		rel := strings.TrimPrefix(obj.Key, base)
		rel = strings.ReplaceAll(rel, "/", "")
		if rel != "" {
			out[rel] = struct{}{}
		}
	}
	return out, nil
}

func (s *S3) Put(ctx context.Context, oid, localPath string) error {
	_, err := s.client.FPutObject(ctx, s.bucket, s.key(oid), localPath, minio.PutObjectOptions{})
	return classifyS3(err)
}

func (s *S3) Get(ctx context.Context, oid, localPath string) error {
	return classifyS3(s.client.FGetObject(ctx, s.bucket, s.key(oid), localPath, minio.GetObjectOptions{}))
}

// retryableS3 wraps an S3 error the transfer layer should retry, exposing
// Temporary() so retries are classified by HTTP status/code, not error strings.
type retryableS3 struct{ error }

func (retryableS3) Temporary() bool { return true }

// classifyS3 marks throttling and 5xx responses as temporary (retryable);
// everything else (auth, NoSuchKey, bad request) is returned unchanged. Raw
// network errors (StatusCode 0) pass through to the transfer layer's fallback.
func classifyS3(err error) error {
	if err == nil {
		return nil
	}
	resp := minio.ToErrorResponse(err)
	switch resp.StatusCode {
	case 429, 500, 502, 503, 504:
		return retryableS3{err}
	}
	switch resp.Code {
	case "SlowDown", "RequestTimeout", "InternalError", "ServiceUnavailable",
		"Throttling", "ThrottlingException", "RequestThrottled", "RequestThrottledException":
		return retryableS3{err}
	}
	return err
}

func (s *S3) Delete(ctx context.Context, oid string) error {
	return s.client.RemoveObject(ctx, s.bucket, s.key(oid), minio.RemoveObjectOptions{})
}

// EnsureBucket creates the bucket if it does not exist (used in tests / first
// push to a fresh MinIO).
func (s *S3) EnsureBucket(ctx context.Context) error {
	ok, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !ok {
		return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	}
	return nil
}

func parseS3URL(raw string) (bucket, prefix string, err error) {
	rest := strings.TrimPrefix(raw, "s3://")
	if rest == raw {
		return "", "", errors.New("remote url must start with s3://")
	}
	parts := strings.SplitN(rest, "/", 2)
	bucket = parts[0]
	if bucket == "" {
		return "", "", errors.New("remote url has no bucket")
	}
	if len(parts) == 2 {
		prefix = strings.Trim(parts[1], "/")
	}
	return bucket, prefix, nil
}

// Reachable verifies connectivity and authentication by listing one object
// under the prefix. A non-nil error means the remote is unreachable or the
// credentials are invalid (used by `lode doctor`).
func (s *S3) Reachable(ctx context.Context) error {
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:  s.prefix,
		MaxKeys: 1,
	}) {
		// First result: nil Err means the listing (auth + connectivity) succeeded.
		return obj.Err
	}
	return ctx.Err() // empty listing, but reachable
}
