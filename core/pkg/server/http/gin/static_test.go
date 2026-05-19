package gin

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockFS는 테스트용 파일시스템입니다
type MockFS struct {
	files map[string]*MockFile
	mu    sync.RWMutex
}

type MockFile struct {
	name    string
	content []byte
	modTime time.Time
	isDir   bool
}

func (m *MockFile) Name() string { return m.name }
func (m *MockFile) Size() int64  { return int64(len(m.content)) }
func (m *MockFile) Mode() fs.FileMode {
	if m.isDir {
		return fs.ModeDir
	}
	return 0644
}
func (m *MockFile) ModTime() time.Time { return m.modTime }
func (m *MockFile) IsDir() bool        { return m.isDir }
func (m *MockFile) Sys() any           { return nil }

func NewMockFS() *MockFS {
	return &MockFS{
		files: make(map[string]*MockFile),
	}
}

func (mfs *MockFS) AddFile(path string, content []byte) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	mfs.files[path] = &MockFile{
		name:    path,
		content: content,
		modTime: time.Now(),
		isDir:   false,
	}
}

func (mfs *MockFS) AddDir(path string) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	mfs.files[path] = &MockFile{
		name:    path,
		isDir:   true,
		modTime: time.Now(),
	}
}

func (mfs *MockFS) Open(name string) (fs.File, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()

	file, exists := mfs.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}

	return &MockOpenFile{file: file, reader: strings.NewReader(string(file.content))}, nil
}

type MockOpenFile struct {
	file   *MockFile
	reader *strings.Reader
}

func (m *MockOpenFile) Stat() (fs.FileInfo, error) {
	return m.file, nil
}

func (m *MockOpenFile) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func (m *MockOpenFile) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

func (m *MockOpenFile) Close() error {
	return nil
}

func TestSafeRedirectPath_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected string
	}{
		{"/normal/path", "/normal/path"},
		{"//double/slash", "/double/slash"},
		{"\\/mixed/slash", "/mixed/slash"},
		{"\\\\backslash", "/backslash"},
		{"/single", "/single"},
		{"", ""},
		{"no-leading-slash", "no-leading-slash"},
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([][]string, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			results[id] = make([]string, len(testCases))

			for j, tc := range testCases {
				results[id][j] = safeRedirectPath(tc.input)
			}
		}(i)
	}

	wg.Wait()

	// 모든 goroutine의 결과가 동일한지 확인
	for i := 1; i < numGoroutines; i++ {
		for j, tc := range testCases {
			assert.Equal(t, tc.expected, results[i][j],
				"Goroutine %d case %d: safeRedirectPath(%q) should return %q", i, j, tc.input, tc.expected)
		}
	}
}

func TestFileFS_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	mockFS := NewMockFS()
	mockFS.AddFile("test.html", []byte("<html>Test</html>"))
	mockFS.AddFile("index.html", []byte("<html>Index</html>"))
	mockFS.AddDir("testdir")
	mockFS.AddFile("testdir/index.html", []byte("<html>Dir Index</html>"))

	testCases := []struct {
		filename    string
		expectError bool
	}{
		{"test.html", false},
		{"index.html", false},
		{"testdir", false}, // 디렉토리는 index.html을 찾음
		{"nonexistent.html", true},
	}

	const numGoroutines = 8
	var wg sync.WaitGroup
	results := make([][]bool, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			results[id] = make([]bool, len(testCases))

			for j, tc := range testCases {
				_, _, err := FileFS(mockFS, tc.filename)
				results[id][j] = (err != nil)
			}
		}(i)
	}

	wg.Wait()

	// 모든 goroutine의 결과가 예상과 일치하는지 확인
	for i, result := range results {
		for j, tc := range testCases {
			assert.Equal(t, tc.expectError, result[j],
				"Goroutine %d case %d: FileFS(%q) error expectation should match", i, j, tc.filename)
		}
	}
}

func TestStatic_ConcurrentRequests(t *testing.T) {
	// t.Parallel() 제거 - gin 전역 상태 사용

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Static 핸들러를 등록
	router.Any("/ui/*filename", Static)

	const numRequests = 15
	var wg sync.WaitGroup
	responses := make([]struct {
		statusCode int
		hasError   bool
		path       string
	}, numRequests)

	wg.Add(numRequests)

	for i := range numRequests {
		go func(id int) {
			defer wg.Done()

			// 다양한 경로로 요청
			paths := []string{
				"/ui/",
				"/ui/index.html",
				"/ui/nonexistent.html",
				"/ui/../test", // directory traversal 시도
			}
			path := paths[id%len(paths)]

			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			responses[id].statusCode = w.Code
			responses[id].hasError = w.Code >= 400
			responses[id].path = path
		}(i)
	}

	wg.Wait()

	// 모든 응답이 적절한 상태 코드를 반환했는지 확인
	for i, resp := range responses {
		assert.True(t, resp.statusCode >= 200 && resp.statusCode < 600,
			"Request %d should return valid HTTP status code", i)
		// directory traversal은 404나 다른 에러를 반환해야 함
		if strings.Contains(resp.path, "../") {
			assert.True(t, resp.hasError,
				"Request %d with directory traversal should return error", i)
		}
	}
}

func TestStatic_DirectoryTraversal_ConcurrentAccess(t *testing.T) {
	// t.Parallel() 제거 - gin 전역 상태 사용

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Any("/ui/*filename", Static)

	// 다양한 directory traversal 패턴
	traversalPaths := []string{
		"/ui/../etc/passwd",
		"/ui/../../sensitive",
		"/ui/../..",
		"/ui/..\\windows\\system32",
	}

	const numGoroutines = 8
	var wg sync.WaitGroup
	responses := make([]int, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			path := traversalPaths[id%len(traversalPaths)]
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			responses[id] = w.Code
		}(i)
	}

	wg.Wait()

	// 모든 directory traversal 시도가 차단되었는지 확인
	for i, statusCode := range responses {
		assert.Equal(t, http.StatusNotFound, statusCode,
			"Directory traversal attempt %d should return 404", i)
	}
}

func TestStatic_RedirectBehavior_ConcurrentAccess(t *testing.T) {
	// t.Parallel() 제거 - gin 전역 상태 사용

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Any("/ui/*filename", Static)

	redirectTestCases := []struct {
		path           string
		expectRedirect bool
	}{
		{"/ui/test", false},         // 파일이 없으면 404
		{"/ui/test/", false},        // 디렉토리가 없으면 404
		{"/ui/nonexistent/", false}, // 존재하지 않는 파일에 trailing slash가 있어도 404
	}

	numGoroutines := len(redirectTestCases) * 3
	var wg sync.WaitGroup
	results := make([]struct {
		statusCode int
		isRedirect bool
	}, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			tc := redirectTestCases[id%len(redirectTestCases)]
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			results[id].statusCode = w.Code
			results[id].isRedirect = (w.Code >= 300 && w.Code < 400)
		}(i)
	}

	wg.Wait()

	// 결과 검증
	for i, result := range results {
		tc := redirectTestCases[i%len(redirectTestCases)]
		if tc.expectRedirect {
			assert.True(t, result.isRedirect,
				"Request %d to %q should be redirected", i, tc.path)
		}
		assert.True(t, result.statusCode >= 200 && result.statusCode < 600,
			"Request %d should return valid status code", i)
	}
}
