package gin

// pocketbase 참조(https://github.com/pocketbase/pocketbase)
import (
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	ui "ntels.com/pharos/core/frontend"
)

const IndexPage = "index.html"

func FileFS(fsys fs.FS, filename string) (io.ReadSeeker, fs.FileInfo, error) {
	f, err := fsys.Open(filename)
	if err != nil {
		return nil, nil, external.ErrorNotExistFile
	}
	defer func(f fs.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	fi, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	// if it is a directory try to open its index.html file
	if fi.IsDir() {
		checkFile := filepath.ToSlash(filepath.Base(filename) + ".html")
		f, err = fsys.Open(checkFile)
		if err == nil {
			goto FILE_EXISTS
		}

		checkFile = filepath.ToSlash(filepath.Join(filename, IndexPage))
		f, err = fsys.Open(checkFile)
		if err != nil {
			return nil, nil, external.ErrorNotExistFile
		}
		defer func(f fs.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)
	FILE_EXISTS:

		fi, err = f.Stat()
		if err != nil {
			return nil, nil, err
		}
	}

	ff, ok := f.(io.ReadSeeker)
	if !ok {
		return nil, nil, external.ErrorInvalidReadSeekerType
	}

	return ff, fi, nil
}

// safeRedirectPath normalizes the path string by replacing all beginning slashes
// (`\\`, `//`, `\/`) with a single forward slash to prevent open redirect attacks
func safeRedirectPath(path string) string {
	if len(path) > 1 && (path[0] == '\\' || path[0] == '/') && (path[1] == '\\' || path[1] == '/') {
		path = "/" + strings.TrimLeft(path, `/\`)
	}
	return path
}

func Static(c *gin.Context) {

	filename := c.Param("filename")
	filename = filepath.ToSlash(filepath.Clean(strings.TrimPrefix(filename, "/")))

	// eagerly check for directory traversal
	//
	// note: this is just out of an abundance of caution because the fs.FS implementation could be non-std,
	// but usually shouldn't be necessary since os.DirFS.Open is expected to fail if the filename starts with dots
	if len(filename) > 2 && filename[0] == '.' && filename[1] == '.' && (filename[2] == '/' || filename[2] == '\\') {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	fi, err := fs.Stat(ui.DistDirFS, filename)
	if err != nil {
		filename = filename + ".html"
		fi, err = fs.Stat(ui.DistDirFS, filename)
		if err != nil {
			// SPA fallback: serve index.html for all non-existent routes
			filename = IndexPage
			fi, err = fs.Stat(ui.DistDirFS, filename)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
				return
			}
		}
	}

	if fi.IsDir() {
		// redirect to a canonical dir url, aka. with trailing slash
		if !strings.HasSuffix(c.Request.URL.Path, "/") {
			c.Redirect(http.StatusMovedPermanently, safeRedirectPath(c.Request.URL.Path+"/"))
			return
		}
	} else {
		urlPath := c.Request.URL.Path
		if strings.HasSuffix(urlPath, "/") {
			// redirect to a non-trailing slash file route
			urlPath = strings.TrimRight(urlPath, "/")
			if len(urlPath) > 0 {
				c.Redirect(http.StatusMovedPermanently, safeRedirectPath(urlPath))
				return
			}
		} else if stripped, ok := strings.CutSuffix(urlPath, IndexPage); ok {
			// redirect without the index.html
			c.Redirect(http.StatusMovedPermanently, safeRedirectPath(stripped))
			return
		}
	}

	openFile, fileInfo, err := FileFS(ui.DistDirFS, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	readFile := make([]byte, fileInfo.Size())
	_, err = openFile.Read(readFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	http.ServeContent(c.Writer, c.Request, fileInfo.Name(), fileInfo.ModTime(), openFile)

}
