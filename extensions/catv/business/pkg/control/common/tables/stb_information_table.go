package tables

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

const distStbInformationTableName = "dist_stb_information"

// Area types for STB list queries
const (
	AreaTypeAll       = "all"
	AreaTypeSO        = "so"
	AreaTypeL3        = "l3"
	AreaTypeCell      = "cell"
	AreaTypeSettopbox = "settopbox"
)

// StbInformationRaw represents a set-top box record in the database.
// Contains Cable Modem (CM) and Set-Top Box (STB) information along with network details.
type StbInformationRaw struct {
	StbMdlNm      string       `json:"stb_mdl_nm" db:"STB_MDL_NM"`
	CmMacAddr     string       `json:"cm_mac_addr" db:"CM_MAC_ADDR"`
	CmIpAddr      string       `json:"cm_ip_addr" db:"CM_IP_ADDR"`
	StbMacAddr    string       `json:"stb_mac_addr" db:"STB_MAC_ADDR"`
	SrcIpAddr     string       `json:"src_ip_addr" db:"SRC_IP_ADDR"`
	CCellNum      string       `json:"ccell_num" db:"CCELL_NUM"`
	L3EquipTidVal string       `json:"l3_equip_tid_val" db:"L3_EQUIP_TID_VAL"`
	ScrbrSoNm     string       `json:"scrbr_so_nm" db:"SCRBR_SO_NM"`
	UpdateTime    orm.Datetime `json:"update_time" db:"UPDATE_TIME"`
}

// StbInformationTable handles database operations for STB information.
type StbInformationTable struct {
	config common.Config
}

// NewStbInformationTable creates a new STB information repository instance.
//
// Parameters:
//   - config: system configuration including database settings
//
// Returns:
//   - *StbInformationTable: repository instance for STB information operations
func NewStbInformationTable(config common.Config) *StbInformationTable {
	return &StbInformationTable{
		config: config,
	}
}

// GetSos retrieves a distinct list of all subscriber SO (System Operator) names.
//
// Returns:
//   - []string: sorted list of SO names in ascending order
//   - error: if database query fails
func (r *StbInformationTable) GetSos() ([]string, error) {
	var results []string
	var query = "SELECT DISTINCT SCRBR_SO_NM FROM " + distStbInformationTableName + " FINAL ORDER BY SCRBR_SO_NM ASC"

	handler := func(db *sqlx.DB) error {
		return db.Select(&results, query)
	}

	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return nil, fmt.Errorf("failed to get SOs: %w", err)
	}

	return results, nil
}

// GetL3s retrieves a distinct list of L3 equipment TID values for a specific SO.
//
// Parameters:
//   - so: subscriber SO name to filter by
//
// Returns:
//   - []string: sorted list of L3 equipment TID values in ascending order
//   - error: if database query fails
func (r *StbInformationTable) GetL3s(so string) ([]string, error) {
	var results []string
	var query = "SELECT DISTINCT L3_EQUIP_TID_VAL FROM " + distStbInformationTableName + " FINAL WHERE SCRBR_SO_NM = ? ORDER BY L3_EQUIP_TID_VAL ASC"
	var args = []any{so}

	handler := func(db *sqlx.DB) error {
		return db.Select(&results, query, args...)
	}

	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return nil, fmt.Errorf("failed to get L3s for SO %s: %w", so, err)
	}

	return results, nil
}

// GetCells retrieves a distinct list of cell numbers for a specific SO and L3 equipment.
//
// Parameters:
//   - so: subscriber SO name to filter by
//   - l3: L3 equipment TID value to filter by
//
// Returns:
//   - []string: sorted list of cell numbers in ascending order
//   - error: if database query fails
func (r *StbInformationTable) GetCells(so, l3 string) ([]string, error) {
	var results []string
	var query = "SELECT DISTINCT CCELL_NUM FROM " + distStbInformationTableName + " FINAL WHERE SCRBR_SO_NM = ? AND L3_EQUIP_TID_VAL = ? ORDER BY CCELL_NUM ASC"
	var args = []any{so, l3}

	handler := func(db *sqlx.DB) error {
		return db.Select(&results, query, args...)
	}

	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return nil, fmt.Errorf("failed to get cells for SO %s and L3 %s: %w", so, l3, err)
	}

	return results, nil
}

// GetSettopboxes retrieves a distinct list of Cable Modem MAC addresses for a specific SO, L3, and cell.
//
// Parameters:
//   - so: subscriber SO name to filter by
//   - l3: L3 equipment TID value to filter by
//   - cell: cell number to filter by
//
// Returns:
//   - []string: sorted list of CM MAC addresses in ascending order
//   - error: if database query fails
func (r *StbInformationTable) GetSettopboxes(so, l3, cell string) ([]string, error) {
	var results []string
	var query = "SELECT DISTINCT CM_MAC_ADDR FROM " + distStbInformationTableName + " FINAL WHERE SCRBR_SO_NM = ? AND L3_EQUIP_TID_VAL = ? AND CCELL_NUM = ? ORDER BY CM_MAC_ADDR ASC"
	var args = []any{so, l3, cell}

	handler := func(db *sqlx.DB) error {
		return db.Select(&results, query, args...)
	}

	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return nil, fmt.Errorf("failed to get settopboxes for SO %s, L3 %s, cell %s: %w", so, l3, cell, err)
	}

	return results, nil
}

// GetStbInformationRaws retrieves STB information records filtered by area type and IDs.
//
// Parameters:
//   - area: area type to filter by (use AreaType constants)
//   - ids: list of IDs to filter by (ignored for AreaTypeAll)
//
// Returns:
//   - []StbInformation: list of STB records matching the criteria
//   - error: if area type is invalid or database query fails
func (r *StbInformationTable) GetStbInformationRaws(area string, ids []string) ([]StbInformationRaw, error) {
	var query string
	var args []any

	switch area {
	case AreaTypeAll:
		query = "SELECT * FROM " + distStbInformationTableName + " FINAL ORDER BY CM_MAC_ADDR ASC"
		args = []any{}
	case AreaTypeSO:
		if len(ids) == 0 {
			return []StbInformationRaw{}, nil
		}
		query = "SELECT * FROM " + distStbInformationTableName + " FINAL WHERE SCRBR_SO_NM IN (?) ORDER BY SCRBR_SO_NM ASC"
		args = []any{ids}
	case AreaTypeL3:
		if len(ids) == 0 {
			return []StbInformationRaw{}, nil
		}
		query = "SELECT * FROM " + distStbInformationTableName + " FINAL WHERE L3_EQUIP_TID_VAL IN (?) ORDER BY L3_EQUIP_TID_VAL ASC"
		args = []any{ids}
	case AreaTypeCell:
		if len(ids) == 0 {
			return []StbInformationRaw{}, nil
		}
		query = "SELECT * FROM " + distStbInformationTableName + " FINAL WHERE CCELL_NUM IN (?) ORDER BY CCELL_NUM ASC"
		args = []any{ids}
	case AreaTypeSettopbox:
		if len(ids) == 0 {
			return []StbInformationRaw{}, nil
		}
		query = "SELECT * FROM " + distStbInformationTableName + " FINAL WHERE CM_MAC_ADDR IN (?) ORDER BY CM_MAC_ADDR ASC"
		args = []any{ids}
	default:
		return nil, fmt.Errorf("invalid area type: %s (use AreaType constants)", area)
	}

	results := []StbInformationRaw{}
	handler := func(db *sqlx.DB) error {
		// For IN queries with slice parameters, use sqlx.In to expand the placeholders
		if len(args) > 0 && len(ids) > 0 {
			q, a, err := sqlx.In(query, args...)
			if err != nil {
				return fmt.Errorf("failed to build IN query: %w", err)
			}
			query = db.Rebind(q)
			args = a
		}
		return db.Select(&results, query, args...)
	}

	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return nil, fmt.Errorf("failed to get STB lists for area %s: %w", area, err)
	}

	return results, nil
}
