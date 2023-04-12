package handlers

import (
	"context"
	"database/sql"
	"github.com/educ-educ/handlers-service/internal/pkg/common"
	"github.com/educ-educ/handlers-service/internal/pkg/http_tools"
	"github.com/google/uuid"
	"time"
)

type PostgresHandlersRepository struct {
	logger common.Logger
	ctx    context.Context
	db     *sql.DB
}

func NewPostgresHandlersRepository(logger common.Logger, ctx context.Context, db *sql.DB) *PostgresHandlersRepository {
	return &PostgresHandlersRepository{
		logger: logger,
		ctx:    ctx,
		db:     db,
	}
}

func (repo *PostgresHandlersRepository) GetUsedSockets() ([]string, *http_tools.Error) {
	queryCtx, queryCancelFunc := context.WithTimeout(repo.ctx, time.Second)
	defer queryCancelFunc()

	rows, err := repo.db.QueryContext(queryCtx, `SELECT socket_address FROM handlers`)
	if err != nil {
		repo.logger.Error(err)
		return nil, &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}
	defer func() {
		if err = rows.Close(); err != nil {
			repo.logger.Error(err)
		}
	}()

	sockets := make([]string, 0)
	socket := ""
	for rows.Next() {
		err = rows.Scan(&socket)
		if err != nil {
			repo.logger.Error(err)
			return nil, &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
		}
		sockets = append(sockets, socket)
	}

	return sockets, nil
}

func (repo *PostgresHandlersRepository) GetSpecification(handlerID string) (Specification, *http_tools.Error) {
	queryCtx, queryCancelFunc := context.WithTimeout(repo.ctx, time.Second)
	defer queryCancelFunc()

	var socketAddress string
	err := repo.db.QueryRowContext(queryCtx,
		`SELECT socket_address FROM handlers WHERE id = $1`, handlerID).Scan(&socketAddress)
	if err != nil {
		repo.logger.Error(err)
		return Specification{}, &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}

	rows, err := repo.db.QueryContext(queryCtx,
		`SELECT path_part, method_type FROM methods WHERE handler_id = $1`, handlerID)
	if err != nil {
		repo.logger.Error(err)
		return Specification{}, &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}
	defer func() {
		if err = rows.Close(); err != nil {
			repo.logger.Error(err)
		}
	}()

	methods := make([]Method, 0)
	method := Method{}
	for rows.Next() {
		err = rows.Scan(&method.PathPart, &method.MethodType)
		if err != nil {
			repo.logger.Error(err)
			return Specification{}, &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
		}
		methods = append(methods, method)
	}

	return Specification{
		Socket:  socketAddress,
		Methods: methods,
	}, nil
}

func (repo *PostgresHandlersRepository) AddHandlerInstance(socket string) (string, *http_tools.Error) {
	queryCtx, queryCancelFunc := context.WithTimeout(repo.ctx, time.Second)
	defer queryCancelFunc()

	id := uuid.New().String()
	_, err := repo.db.ExecContext(queryCtx,
		`INSERT INTO handlers (id, socket_address) VALUES ($1, $2)`, id, socket)
	if err != nil {
		repo.logger.Error(err)
		return "", &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}

	return id, nil
}

func (repo *PostgresHandlersRepository) AddMethods(handlerID string, methods []Method) *http_tools.Error {
	queryCtx, queryCancelFunc := context.WithTimeout(repo.ctx, time.Second)
	defer queryCancelFunc()

	for _, method := range methods {
		_, err := repo.db.ExecContext(queryCtx,
			`INSERT INTO methods (handler_id, path_part, method_type) VALUES ($1, $2, $3)`,
			handlerID, method.PathPart, method.MethodType)
		if err != nil {
			repo.logger.Error(err)
			return &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
		}
	}

	return nil
}

func (repo *PostgresHandlersRepository) RemoveHandler(handlerID string) *http_tools.Error {
	queryCtx, queryCancelFunc := context.WithTimeout(repo.ctx, time.Second)
	defer queryCancelFunc()

	_, err := repo.db.ExecContext(queryCtx,
		`DELETE FROM handlers WHERE id = $1`, handlerID)
	if err != nil {
		repo.logger.Error(err)
		return &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}

	return nil
}

func (repo *PostgresHandlersRepository) UpdateHandler(handlerID string, specification Specification) *http_tools.Error {
	queryCtx, queryCancelFunc := context.WithTimeout(repo.ctx, time.Second)
	defer queryCancelFunc()

	_, err := repo.db.ExecContext(queryCtx,
		`UPDATE handlers SET socket_address = $1 WHERE id = $2`,
		specification.Socket, handlerID)
	if err != nil {
		repo.logger.Error(err)
		return &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}

	_, err = repo.db.ExecContext(queryCtx,
		`DELETE FROM methods WHERE handler_id = $1`, handlerID)
	if err != nil {
		repo.logger.Error(err)
		return &http_tools.Error{Type: http_tools.DatabaseError, Info: err.Error()}
	}

	queryCancelFunc()

	return repo.AddMethods(handlerID, specification.Methods)
}
