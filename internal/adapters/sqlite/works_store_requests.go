package sqlite

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

var _ workspaceapp.WorksRequestRepository = (*DB)(nil)

func (db *DB) BeginWorksRequest(ctx context.Context, caller, command, key string, input []byte) (workspaceapp.WorksRequestReceipt, bool, error) {
	var receipt workspaceapp.WorksRequestReceipt
	result, err := db.conn.ExecContext(ctx, "INSERT INTO works_requests(caller_id,command,request_key,input,status) VALUES(?,?,?,?,'pending') ON CONFLICT(caller_id,command,request_key) DO NOTHING", caller, command, key, input)
	if err != nil {
		return receipt, false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return receipt, false, err
	}
	var errorData, resultData []byte
	err = db.conn.QueryRowContext(ctx, "SELECT input,status,result,error FROM works_requests WHERE caller_id=? AND command=? AND request_key=?", caller, command, key).Scan(&receipt.Input, &receipt.Status, &resultData, &errorData)
	if err != nil {
		return receipt, false, err
	}
	if !bytes.Equal(receipt.Input, input) {
		return receipt, false, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "works request key binds different input", RequiredAction: "review_original_works_request"}
	}
	receipt.Result = resultData
	if len(errorData) > 0 {
		if err := json.Unmarshal(errorData, &receipt.Error); err != nil {
			return receipt, false, err
		}
	}
	return receipt, count == 1, nil
}

func (db *DB) FinishWorksRequest(ctx context.Context, caller, command, key string, receipt workspaceapp.WorksRequestReceipt) error {
	var errorData []byte
	if receipt.Error != nil {
		var err error
		errorData, err = json.Marshal(receipt.Error)
		if err != nil {
			return err
		}
	}
	result, err := db.conn.ExecContext(ctx, "UPDATE works_requests SET status=?,result=?,error=? WHERE caller_id=? AND command=? AND request_key=? AND status='pending'", receipt.Status, []byte(receipt.Result), errorData, caller, command, key)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "works request was already finalized", RequiredAction: "read_works_request_result"}
	}
	return nil
}
