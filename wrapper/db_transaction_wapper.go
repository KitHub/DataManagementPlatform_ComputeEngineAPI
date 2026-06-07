package wrapper

import (
	"context"
	"log/slog"

	"xorm.io/xorm"
)

// TransactionWrapper is a wrapper for database transaction. It will begin a transaction, execute the given function, and commit or rollback the transaction based on the result of the function.
// NOTE: nested transaction is not supported, which means that if the given function calls TransactionWrapper again, the inner transaction will not be executed in the existed transaction, but will start a new transaction.
func TransactionWrapper(ctx context.Context, dbEngine *xorm.Engine,
	fn func(session *xorm.Session) error) (err error) {
	panicked := true

	session := dbEngine.NewSession()
	slog.DebugContext(ctx, "create db session")
	defer session.Close()

	err = session.Begin()
	slog.DebugContext(ctx, "begin transaction")
	defer func() {
		if panicked || err != nil {
			slog.ErrorContext(ctx, "transaction failed, rollback transaction",
				slog.Any("error", err))
			err = session.Rollback()
			if err != nil {
				slog.ErrorContext(ctx, "rollback transaction failed",
					slog.Any("error", err))
			}
			return
		}
	}()
	if err != nil {
		slog.ErrorContext(ctx, "create db session failed",
			slog.Any("error", err))
		return err
	}

	err = fn(session)
	if err != nil {
		slog.ErrorContext(ctx, "transaction failed", slog.Any("error", err))
		return err
	}

	err = session.Commit()
	slog.DebugContext(ctx, "commit transaction")
	if err != nil {
		slog.ErrorContext(ctx, "commit transaction failed",
			slog.Any("error", err))
		return err
	}

	panicked = false
	slog.DebugContext(ctx, "commit transaction done")
	return nil
}
