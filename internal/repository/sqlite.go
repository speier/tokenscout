package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/speier/tokenscout/internal/models"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLite(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) runMigrations() error {
	schema := `
	CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		side TEXT NOT NULL,
		mint TEXT NOT NULL,
		quantity TEXT NOT NULL,
		price_usd REAL,
		tx_sig TEXT,
		status TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS positions (
		mint TEXT PRIMARY KEY,
		quantity TEXT NOT NULL,
		avg_price_usd REAL,
		opened_at INTEGER NOT NULL,
		last_update_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		mint TEXT NOT NULL,
		pair TEXT,
		lp_address TEXT,
		timestamp INTEGER NOT NULL,
		raw TEXT
	);

	CREATE TABLE IF NOT EXISTS configs (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS blacklist (
		mint TEXT PRIMARY KEY
	);

	CREATE TABLE IF NOT EXISTS whitelist (
		mint TEXT PRIMARY KEY
	);

	CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
	`

	_, err := r.db.Exec(schema)
	return err
}

func (r *SQLiteRepository) CreateTrade(ctx context.Context, trade *models.Trade) error {
	query := `INSERT INTO trades (timestamp, side, mint, quantity, price_usd, tx_sig, status)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		trade.Timestamp.Unix(),
		trade.Side,
		trade.Mint,
		trade.Quantity,
		trade.PriceUSD,
		trade.TxSig,
		trade.Status,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	trade.ID = id
	return nil
}

func (r *SQLiteRepository) GetTrades(ctx context.Context, limit int) ([]models.Trade, error) {
	query := `SELECT id, timestamp, side, mint, quantity, price_usd, tx_sig, status
			  FROM trades ORDER BY timestamp DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		var ts int64
		err := rows.Scan(&t.ID, &ts, &t.Side, &t.Mint, &t.Quantity, &t.PriceUSD, &t.TxSig, &t.Status)
		if err != nil {
			return nil, err
		}
		t.Timestamp = time.Unix(ts, 0)
		trades = append(trades, t)
	}
	return trades, nil
}

func (r *SQLiteRepository) GetTradeByID(ctx context.Context, id int64) (*models.Trade, error) {
	query := `SELECT id, timestamp, side, mint, quantity, price_usd, tx_sig, status
			  FROM trades WHERE id = ?`
	var t models.Trade
	var ts int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &ts, &t.Side, &t.Mint, &t.Quantity, &t.PriceUSD, &t.TxSig, &t.Status,
	)
	if err != nil {
		return nil, err
	}
	t.Timestamp = time.Unix(ts, 0)
	return &t, nil
}

func (r *SQLiteRepository) UpdateTradeStatus(ctx context.Context, id int64, status models.TradeStatus, txSig string) error {
	query := `UPDATE trades SET status = ?, tx_sig = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, txSig, id)
	return err
}

func (r *SQLiteRepository) CreatePosition(ctx context.Context, position *models.Position) error {
	query := `INSERT INTO positions (mint, quantity, avg_price_usd, opened_at, last_update_at)
			  VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		position.Mint,
		position.Quantity,
		position.AvgPriceUSD,
		position.OpenedAt.Unix(),
		position.LastUpdateAt.Unix(),
	)
	return err
}

func (r *SQLiteRepository) GetPosition(ctx context.Context, mint string) (*models.Position, error) {
	query := `SELECT mint, quantity, avg_price_usd, opened_at, last_update_at
			  FROM positions WHERE mint = ?`
	var p models.Position
	var openedAt, lastUpdateAt int64
	err := r.db.QueryRowContext(ctx, query, mint).Scan(
		&p.Mint, &p.Quantity, &p.AvgPriceUSD, &openedAt, &lastUpdateAt,
	)
	if err != nil {
		return nil, err
	}
	p.OpenedAt = time.Unix(openedAt, 0)
	p.LastUpdateAt = time.Unix(lastUpdateAt, 0)
	return &p, nil
}

func (r *SQLiteRepository) GetAllPositions(ctx context.Context) ([]models.Position, error) {
	query := `SELECT mint, quantity, avg_price_usd, opened_at, last_update_at FROM positions`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var p models.Position
		var openedAt, lastUpdateAt int64
		err := rows.Scan(&p.Mint, &p.Quantity, &p.AvgPriceUSD, &openedAt, &lastUpdateAt)
		if err != nil {
			return nil, err
		}
		p.OpenedAt = time.Unix(openedAt, 0)
		p.LastUpdateAt = time.Unix(lastUpdateAt, 0)
		positions = append(positions, p)
	}
	return positions, nil
}

func (r *SQLiteRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	query := `UPDATE positions SET quantity = ?, avg_price_usd = ?, last_update_at = ?
			  WHERE mint = ?`
	_, err := r.db.ExecContext(ctx, query,
		position.Quantity,
		position.AvgPriceUSD,
		position.LastUpdateAt.Unix(),
		position.Mint,
	)
	return err
}

func (r *SQLiteRepository) DeletePosition(ctx context.Context, mint string) error {
	query := `DELETE FROM positions WHERE mint = ?`
	_, err := r.db.ExecContext(ctx, query, mint)
	return err
}

func (r *SQLiteRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	query := `INSERT INTO events (type, mint, pair, lp_address, timestamp, raw)
			  VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		event.Type,
		event.Mint,
		event.Pair,
		event.LPAddress,
		event.Timestamp.Unix(),
		event.Raw,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

func (r *SQLiteRepository) GetRecentEvents(ctx context.Context, limit int) ([]models.Event, error) {
	query := `SELECT id, type, mint, pair, lp_address, timestamp, raw
			  FROM events ORDER BY timestamp DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var ts int64
		err := rows.Scan(&e.ID, &e.Type, &e.Mint, &e.Pair, &e.LPAddress, &ts, &e.Raw)
		if err != nil {
			return nil, err
		}
		e.Timestamp = time.Unix(ts, 0)
		events = append(events, e)
	}
	return events, nil
}

func (r *SQLiteRepository) GetConfig(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM configs WHERE key = ?`
	var value string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&value)
	return value, err
}

func (r *SQLiteRepository) SetConfig(ctx context.Context, key, value string) error {
	query := `INSERT OR REPLACE INTO configs (key, value) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, key, value)
	return err
}

func (r *SQLiteRepository) IsBlacklisted(ctx context.Context, mint string) (bool, error) {
	query := `SELECT COUNT(*) FROM blacklist WHERE mint = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, mint).Scan(&count)
	return count > 0, err
}

func (r *SQLiteRepository) IsWhitelisted(ctx context.Context, mint string) (bool, error) {
	query := `SELECT COUNT(*) FROM whitelist WHERE mint = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, mint).Scan(&count)
	return count > 0, err
}

func (r *SQLiteRepository) AddToBlacklist(ctx context.Context, mint string) error {
	query := `INSERT OR IGNORE INTO blacklist (mint) VALUES (?)`
	_, err := r.db.ExecContext(ctx, query, mint)
	return err
}

func (r *SQLiteRepository) AddToWhitelist(ctx context.Context, mint string) error {
	query := `INSERT OR IGNORE INTO whitelist (mint) VALUES (?)`
	_, err := r.db.ExecContext(ctx, query, mint)
	return err
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
