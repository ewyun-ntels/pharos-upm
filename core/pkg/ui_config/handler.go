package ui_config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/orderedmap"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	ui "ntels.com/pharos/core/frontend"
	"ntels.com/pharos/core/pkg/common"
)

type ConfigForUI struct {
	Config string `json:"-" db:"config"`

	config common.Config
}

func (configForUI *ConfigForUI) GetHandler(_ *gin.Context) (int, any) {
	handler := func(db *sqlx.DB) error {
		return db.Get(configForUI, `SELECT * FROM ui_config_config;`)
	}
	if err := orm.Handler(orm.DriverDefault, &configForUI.config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	config := orderedmap.New()
	config.SetEscapeHTML(false)
	if err := config.UnmarshalJSON([]byte(configForUI.Config)); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, config
}

func (configForUI *ConfigForUI) PostHandler(c *gin.Context) (int, any) {
	if body, err := io.ReadAll(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else {
		configForUI.Config = string(body)
	}

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(`INSERT INTO ui_config_config(config) VALUES (:config);`, configForUI)
		return err
	}
	if err := orm.Handler(orm.DriverDefault, &configForUI.config.Database, handler); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (configForUI *ConfigForUI) PutHandler(c *gin.Context) (int, any) {
	if body, err := io.ReadAll(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else {
		configForUI.Config = string(body)
	}

	handler := func(db *sqlx.DB) error {
		// UPSERT: 데이터가 없으면 INSERT, 있으면 UPDATE
		result, err := db.NamedExec(`UPDATE ui_config_config SET config = :config;`, configForUI)
		if err != nil {
			return err
		}

		// UPDATE로 영향받은 행이 없으면 INSERT
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			_, err = db.NamedExec(`INSERT INTO ui_config_config(config) VALUES (:config);`, configForUI)
			return err
		}

		return nil
	}
	if err := orm.Handler(orm.DriverDefault, &configForUI.config.Database, handler); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (configForUI *ConfigForUI) DeleteHandler(_ *gin.Context) (int, any) {
	handler := func(db *sqlx.DB) error {
		_, err := db.Exec(`DELETE FROM ui_config_config;`)
		return err
	}
	if err := orm.Handler(orm.DriverDefault, &configForUI.config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func getConfigHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		configForUI := ConfigForUI{config: config}

		switch c.Request.Method {
		case http.MethodGet:
			c.JSON(configForUI.GetHandler(c))
		case http.MethodPost:
			c.JSON(configForUI.PostHandler(c))
		case http.MethodPut:
			c.JSON(configForUI.PutHandler(c))
		case http.MethodDelete:
			c.JSON(configForUI.DeleteHandler(c))
		default:
			c.JSON(http.StatusNotImplemented, nil)
		}
	}
}

type Images struct {
	basePath string
}

func (images *Images) GetHandler(c *gin.Context) (int, string, []byte) {
	if len(c.Param("name")) != 0 {
		return images.getHandler(c)
	} else {
		return images.getsHandler(c)
	}
}

func (images *Images) getHandler(c *gin.Context) (int, string, []byte) {
	name := images.basePath + string(os.PathSeparator) + c.Param("name")
	bytes, err := ui.UiConfigImagesFS.ReadFile(name)
	if err != nil {
		return http.StatusInternalServerError, restful.MIME_JSON, []byte(`{"message":"` + err.Error() + `"}`)
	}

	infos := map[string]string{
		".svg": "image/svg+xml",
	}

	for extension, contentType := range infos {
		if strings.HasSuffix(name, strings.ToLower(extension)) {
			return http.StatusOK, contentType, bytes
		}
	}

	return http.StatusOK, http.DetectContentType(bytes), bytes
}

func (images *Images) getsHandler(_ *gin.Context) (int, string, []byte) {
	var body []string

	walkDirFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		after, found := strings.CutPrefix(path, images.basePath+string(os.PathSeparator))
		if found && len(after) != 0 {
			body = append(body, after)
		}

		return nil
	}
	err := fs.WalkDir(ui.UiConfigImagesFS, images.basePath, walkDirFunc)
	if err != nil {
		return http.StatusInternalServerError, restful.MIME_JSON, []byte(`{"message":"` + err.Error() + `"}`)
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		return http.StatusInternalServerError, restful.MIME_JSON, []byte(`{"message":"` + err.Error() + `"}`)
	}

	return http.StatusOK, restful.MIME_JSON, bytes
}

func imagesHandler(c *gin.Context) {
	images := Images{basePath: "images"}

	switch c.Request.Method {
	case http.MethodGet:
		c.Data(images.GetHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}
