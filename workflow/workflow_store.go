package workflow

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// sql statements
var (
	createWorkflowStmt = `INSERT INTO workflows (name) VALUES ($1) 
						 ON CONFLICT ON CONSTRAINT workflows_name_key DO NOTHING
						 RETURNING workflow_uuid;`
	selectWorkflowsStmt = `SELECT workflow_uuid, name FROM workflows`
)

// WrkflowUUID is a filter we can add as a parameter to narrow down the list of returned results
type WrkflowUUID struct {
	UUID string
}

func (p WrkflowUUID) where() string {
	return fmt.Sprintf("workflow_uuid = '%s'", p.UUID)
}

// Create stores a new workflow in Postgres
func (store pgStore) CreateWorkflow(wf *Workflow) (*Workflow, error) {
	err := store.QueryRow(createWorkflowStmt, wf.Name).Scan(&wf.UUID)
	if err == sql.ErrNoRows {
		return nil, ErrExists
	}
	if err != nil {
		return nil, errors.Wrap(err, "pgStore create workflow")
	}
	profiles := wf.Profiles
	if err := store.addProfiles(wf.UUID, profiles...); err != nil {
		return nil, err
	}
	return wf, nil
}

func (store pgStore) addProfiles(wfUUID string, profiles ...Profile) error {
	if len(profiles) == 0 {
		return nil
	}
	for _, prf := range profiles {
		if err := store.addProfile(wfUUID, prf.UUID); err != nil {
			return err
		}
	}
	return nil
}

func (store pgStore) addProfile(wfUUID, pfUUID string) error {
	addProfileStmt := `INSERT INTO workflow_profile (workflow_uuid, profile_uuid) VALUES ($1, $2)
								  ON CONFLICT ON CONSTRAINT workflow_profile_pkey DO NOTHING;`

	_, err := store.Exec(addProfileStmt, wfUUID, pfUUID)
	if err != nil {
		return errors.Wrap(err, "pgStore add profile to workflow")
	}
	return nil
}

func (store pgStore) Workflows(params ...interface{}) ([]Workflow, error) {
	stmt := selectWorkflowsStmt
	stmt = addWhereFilters(stmt, params...)
	var workflows []Workflow
	err := store.Select(&workflows, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "pgStore Workflows")
	}
	return workflows, nil
}
