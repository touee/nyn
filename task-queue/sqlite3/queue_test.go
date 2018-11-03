package s3queue_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	s3queue "github.com/touee/nyn/task-queue/sqlite3"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func getTempFileName(testName string) string {
	return path.Join(os.TempDir(), fmt.Sprintf("_temp_%s_%d.s3db", testName, rand.Int()))
}

func TestOpenAndClose(t *testing.T) {
	var tempFileName = getTempFileName(t.Name())
	defer func() {
		os.Remove(tempFileName)
	}()

	var (
		err error
		q   *s3queue.Queue
	)
	q, err = s3queue.Open(tempFileName)
	if err != nil {
		t.Fatal(err)
	}

	var newDefaultMaxAttempts = 5
	func() {
		defer func() {
			err = q.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		if a := q.GetDefaultMaxAttempts(); a != s3queue.DefaultDefaultMaxAttempts {
			t.Fatal(a)
		}

		q.SetDefaultMaxAttempts(newDefaultMaxAttempts)
		if a := q.GetDefaultMaxAttempts(); a != newDefaultMaxAttempts {
			t.Fatal(a)
		}
	}()

	q, err = s3queue.Open(tempFileName)
	if err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			err = q.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		if a := q.GetDefaultMaxAttempts(); a != newDefaultMaxAttempts {
			t.Fatal(a)
		}
	}()
}
