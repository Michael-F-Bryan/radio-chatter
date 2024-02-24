package on_disk_storage

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"go.uber.org/zap"
)

func New(logger *zap.Logger, rootDir string) (*OnDiskStorage, error) {
	conn, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	addr := conn.Addr()
	server := &http.Server{
		Handler:     server(logger, rootDir),
		Addr:        addr.String(),
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	errChan := make(chan error, 2)

	go func() {
		defer close(errChan)

		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			errChan <- err
		}
	}()

	go func() {
		if err := server.Serve(conn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	storage := &OnDiskStorage{
		rootDir: rootDir,
		logger:  logger,
		cancel:  cancel,
		errors:  errChan,
		addr:    addr,
	}

	// HACK: Make sure we don't leave dangling servers if the caller forgets to
	// shut down manually
	runtime.SetFinalizer(storage, func(s *OnDiskStorage) {
		select {
		case <-ctx.Done():
			// The caller has already triggered a shutdown
		default:
			// Manually shut down the server
			logger.DPanic(
				"On Disk Storage server wasn't shutdown. Forcing it to shutdown to avoid a dangling server",
				zap.Stringer("addr", addr),
				zap.String("root", rootDir),
			)
			s.cancel()
		}
	})

	return storage, nil
}

type OnDiskStorage struct {
	rootDir string
	logger  *zap.Logger
	cancel  context.CancelFunc
	errors  <-chan error
	addr    net.Addr
}

func (s *OnDiskStorage) Close() error {
	s.cancel()
	return <-s.errors
}

func (s *OnDiskStorage) Addr() net.Addr {
	return s.addr
}

func (s *OnDiskStorage) path(item blob.Key) string {
	return path.Join(s.rootDir, item.String())
}

func (s *OnDiskStorage) Link(ctx context.Context, key blob.Key, validFor time.Duration) (*url.URL, error) {
	filename := s.path(key)

	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to find %q: %w", filename, blob.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to find %q: %w", filename, err)
	}

	raw := fmt.Sprintf("http://%s/%s", s.addr, key)
	return url.Parse(raw)
}

func (s *OnDiskStorage) Store(ctx context.Context, data []byte) (blob.Key, error) {
	key := blob.KeyForBytes(data)
	filename := s.path(key)

	parent := path.Dir(filename)
	if err := os.MkdirAll(parent, 0766); err != nil {
		return blob.Key{}, fmt.Errorf("unable to create %s/: %w", s.rootDir, err)
	}

	if _, err := os.Stat(filename); err == nil {
		s.logger.Debug(
			"Already exists",
			zap.String("filename", filename),
			zap.Stringer("key", key),
		)

		return key, nil
	}

	s.logger.Debug(
		"Saving blob",
		zap.String("filename", filename),
		zap.Stringer("key", key),
		zap.Int("bytes", len(data)),
	)
	// FIXME: Should probably write to a temporary file and move to the final
	// destination. We might also want to use singleflight so we don't write
	// to the same file multiple times.
	if err := os.WriteFile(filename, data, 0766); err != nil {
		return blob.Key{}, fmt.Errorf("unable to save to %s: %w", filename, err)
	}

	return key, nil
}
