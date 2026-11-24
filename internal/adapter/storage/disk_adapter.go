// Package storage provides thread-safe local filesystem persistence adapters for snapshots and historical caches.
package storage

// Import standard library packages for file operations, JSON encoding, path manipulation, and concurrency.
import (
	// encoding/json handles serializing Go structs to JSON bytes and deserializing them back.
	"encoding/json"
	// fmt allows formatted string generation for error messages.
	"fmt"
	// os provides file system manipulation primitives like Open, Create, Rename, and Remove.
	"os"
	// path/filepath ensures cross-platform file path manipulation on Windows and Linux.
	"path/filepath"
	// strings provides string manipulation and lowercase normalization.
	"strings"
	// sync provides read-write mutexes for thread-safe concurrent file access.
	"sync"
	// time handles timestamps and file modification age evaluations.
	"time"
)

// DiskAdapter coordinates atomic reads and writes of JSON snapshots to local disk.
type DiskAdapter struct {
	// mu provides mutual exclusion locks to prevent concurrent file corruption.
	mu sync.RWMutex
	// baseDir specifies the root directory containing snapshots and historical cache files.
	baseDir string
}

// NewDiskAdapter constructs and initializes a new DiskAdapter rooted at the specified directory.
func NewDiskAdapter(baseDir string) (*DiskAdapter, error) {
	// Ensure the root base directory exists on disk with full read/write/execute permissions.
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		// Return error if base directory cannot be initialized.
		return nil, fmt.Errorf("failed to create disk adapter base directory: %w", err)
	}

	// Create subdirectories for live snapshots and historical weekly archives.
	snapshotsDir := filepath.Join(baseDir, "snapshots")
	// Attempt creating snapshots directory.
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		// Return error if snapshots subdirectory creation fails.
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Define historical archives directory path.
	historyDir := filepath.Join(baseDir, "history")
	// Attempt creating history directory.
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		// Return error if history subdirectory creation fails.
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Construct and return initialized DiskAdapter instance.
	return &DiskAdapter{
		// Store base directory path.
		baseDir: baseDir,
	}, nil
}

// GetSnapshotPath computes the full filesystem path for a given domain snapshot.
func (d *DiskAdapter) GetSnapshotPath(domain string, market ...string) string {
	// Candidate file names to check.
	var candidateNames []string

	// If a market sub-identifier is provided.
	if len(market) > 0 && market[0] != "" {
		m := strings.ToLower(strings.TrimSpace(market[0]))
		// Map aliases.
		shortCode := m
		if m == "saudi" {
			shortCode = "sa"
		} else if m == "uae" {
			shortCode = "ae"
		} else if m == "qatar" {
			shortCode = "qa"
		} else if m == "kuwait" {
			shortCode = "kw"
		}
		// Add underscore and hyphen candidates.
		candidateNames = append(candidateNames,
			fmt.Sprintf("%s_%s.json", domain, shortCode),
			fmt.Sprintf("%s_%s.json", domain, m),
			fmt.Sprintf("%s-%s.json", domain, shortCode),
			fmt.Sprintf("%s-%s.json", domain, m),
		)
	} else {
		// Single domain filename.
		candidateNames = append(candidateNames, domain+".json")
		// For rates, add legacy fallback.
		if domain == "rates" {
			candidateNames = append(candidateNames, "latest-snapshot.json")
		}
	}

	// Directories to probe.
	candidateDirs := []string{
		d.baseDir,
		filepath.Join(d.baseDir, "snapshots"),
	}

	// Check existing file candidates.
	for _, dir := range candidateDirs {
		for _, name := range candidateNames {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// Default fallback path for saving new snapshots.
	defaultName := domain + ".json"
	if len(market) > 0 && market[0] != "" {
		m := strings.ToLower(strings.TrimSpace(market[0]))
		if m == "saudi" {
			m = "sa"
		} else if m == "uae" {
			m = "ae"
		} else if m == "qatar" {
			m = "qa"
		} else if m == "kuwait" {
			m = "kw"
		}
		defaultName = fmt.Sprintf("%s_%s.json", domain, m)
	}
	return filepath.Join(d.baseDir, defaultName)
}

// SaveSnapshot atomically writes a domain snapshot envelope to disk using write-and-rename.
func (d *DiskAdapter) SaveSnapshot(domain string, data any, market ...string) error {
	// Acquire exclusive write lock to prevent concurrent collisions on the same target file.
	d.mu.Lock()
	// Defer release of write lock upon function exit.
	defer d.mu.Unlock()

	// Compute final target destination path.
	targetPath := d.GetSnapshotPath(domain, market...)
	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		// Return directory creation error.
		return fmt.Errorf("failed to create directory for snapshot %s: %w", targetPath, err)
	}

	// Create temporary file path in the same directory for atomic rename.
	tmpPath := fmt.Sprintf("%s.tmp.%d", targetPath, time.Now().UnixNano())

	// Serialize data into formatted JSON bytes.
	bytes, err := json.MarshalIndent(data, "", "  ")
	// Check for JSON serialization error.
	if err != nil {
		// Return JSON encoding error.
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	// Write payload bytes to temporary file with 0644 permissions.
	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		// Return write error.
		return fmt.Errorf("failed to write temp snapshot file: %w", err)
	}

	// Atomically rename temporary file to target destination path.
	if err := os.Rename(tmpPath, targetPath); err != nil {
		// Remove temporary file on failure.
		_ = os.Remove(tmpPath)
		// Return rename error.
		return fmt.Errorf("failed to atomically rename snapshot file: %w", err)
	}

	// Return success.
	return nil
}

// LoadSnapshot reads and deserializes a domain snapshot from disk into the provided target pointer.
func (d *DiskAdapter) LoadSnapshot(domain string, target any, market ...string) error {
	// Acquire shared read lock.
	d.mu.RLock()
	// Defer release of read lock upon function return.
	defer d.mu.RUnlock()

	// Compute snapshot filesystem path.
	targetPath := d.GetSnapshotPath(domain, market...)

	// Read file contents from disk.
	bytes, err := os.ReadFile(targetPath)
	// Check if reading file failed.
	if err != nil {
		// Return read error.
		return err
	}

	// Check if the JSON payload is wrapped inside an outer metadata envelope containing a "data" property.
	var envelope struct {
		// Data holds the raw unparsed inner JSON bytes for the domain quote payload.
		Data json.RawMessage `json:"data"`
	}
	// Attempt deserializing into the outer envelope structure first.
	if err := json.Unmarshal(bytes, &envelope); err == nil && len(envelope.Data) > 0 {
		// Attempt unmarshaling inner data bytes into the caller's target struct.
		if err := json.Unmarshal(envelope.Data, target); err == nil {
			// Successfully deserialized unwrapped data payload.
			return nil
		}
	}

	// Deserialize JSON bytes directly into the provided target struct if no envelope was present.
	if err := json.Unmarshal(bytes, target); err != nil {
		// Return JSON unmarshal error.
		return fmt.Errorf("failed to unmarshal snapshot %s: %w", targetPath, err)
	}

	// Return success.
	return nil
}

// GetHistoryPath computes the full filesystem path for a weekly historical bucket file.
func (d *DiskAdapter) GetHistoryPath(collection, weekKey string) string {
	// Return joined path inside history subdirectory.
	return filepath.Join(d.baseDir, "history", collection, fmt.Sprintf("%s.json", weekKey))
}

// SaveHistoryBucket atomically writes a weekly historical bucket to disk with 30-day lifecycle.
func (d *DiskAdapter) SaveHistoryBucket(collection, weekKey string, data any) error {
	// Acquire exclusive write lock.
	d.mu.Lock()
	// Defer release of write lock upon function exit.
	defer d.mu.Unlock()

	// Determine destination target path.
	targetPath := d.GetHistoryPath(collection, weekKey)
	// Ensure parent collection directory exists.
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		// Return directory creation error.
		return fmt.Errorf("failed to create directory for history: %w", err)
	}

	// Compute temp file path.
	tmpPath := fmt.Sprintf("%s.tmp.%d", targetPath, time.Now().UnixNano())

	// Serialize data into JSON bytes.
	bytes, err := json.Marshal(data)
	// Check for marshaling errors.
	if err != nil {
		// Return marshal error.
		return fmt.Errorf("failed to marshal history bucket: %w", err)
	}

	// Write to temporary file.
	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		// Return write error.
		return fmt.Errorf("failed to write temp history file: %w", err)
	}

	// Atomically rename temporary file to destination path.
	if err := os.Rename(tmpPath, targetPath); err != nil {
		// Cleanup temporary file on failure.
		_ = os.Remove(tmpPath)
		// Return rename error.
		return fmt.Errorf("failed to rename history file: %w", err)
	}

	// Return success.
	return nil
}

// LoadHistoryBucket reads a weekly historical bucket from disk if it exists.
func (d *DiskAdapter) LoadHistoryBucket(collection, weekKey string, target any) error {
	// Acquire read lock.
	d.mu.RLock()
	// Defer release of read lock upon function return.
	defer d.mu.RUnlock()

	// Determine file path.
	targetPath := d.GetHistoryPath(collection, weekKey)

	// Read raw file bytes.
	bytes, err := os.ReadFile(targetPath)
	// Check if read failed.
	if err != nil {
		// Return error.
		return err
	}

	// Deserialize JSON bytes into target pointer.
	if err := json.Unmarshal(bytes, target); err != nil {
		// Return deserialization error.
		return fmt.Errorf("failed to unmarshal history bucket %s: %w", targetPath, err)
	}

	// Return success.
	return nil
}
