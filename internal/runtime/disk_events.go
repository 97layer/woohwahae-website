package runtime

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

func (d *diskStore) loadEvents() ([]EventEnvelope, error) {
	return readJSON(d.eventsPath(), []EventEnvelope{})
}

func (d *diskStore) saveEvents(items []EventEnvelope) error {
	return writeJSONAtomic(d.eventsPath(), items)
}

func (d *diskStore) countEventArchive() (int, error) {
	count := 0
	err := d.streamEventArchive(func(EventEnvelope) error {
		count++
		return nil
	})
	return count, err
}

func (d *diskStore) readEventArchive() ([]EventEnvelope, error) {
	items := []EventEnvelope{}
	err := d.streamEventArchive(func(item EventEnvelope) error {
		items = append(items, item)
		return nil
	})
	return items, err
}

func (d *diskStore) appendEventArchive(items []EventEnvelope) error {
	if len(items) == 0 {
		return nil
	}
	if err := d.ensureEventArchiveNDJSON(); err != nil {
		return err
	}
	file, err := os.OpenFile(d.eventArchivePath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err := writer.Write(raw); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (d *diskStore) eventsPath() string {
	return filepath.Join(d.baseDir, "events.json")
}

func (d *diskStore) eventArchivePath() string {
	return filepath.Join(d.baseDir, "events_archive.json")
}

func (d *diskStore) streamEventArchive(visit func(EventEnvelope) error) error {
	file, err := os.Open(d.eventArchivePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		b, err := reader.Peek(1)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if isSpaceByte(b[0]) {
			_, _ = reader.ReadByte()
			continue
		}
		break
	}

	b, err := reader.Peek(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	decoder := json.NewDecoder(reader)
	if b[0] == '[' {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); !ok || delim != '[' {
			return errors.New("invalid event archive format")
		}
		for decoder.More() {
			var item EventEnvelope
			if err := decoder.Decode(&item); err != nil {
				return err
			}
			if err := visit(item); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		if err != nil {
			return err
		}
		return nil
	}
	skipped := 0
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				var item EventEnvelope
				if unmarshalErr := json.Unmarshal(trimmed, &item); unmarshalErr != nil {
					skipped++
				} else if visitErr := visit(item); visitErr != nil {
					return visitErr
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				if skipped > 0 {
					log.Printf("event archive: skipped %d malformed ndjson line(s)", skipped)
				}
				return nil
			}
			return err
		}
	}
}

func (d *diskStore) ensureEventArchiveNDJSON() error {
	file, err := os.Open(d.eventArchivePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	reader := bufio.NewReader(file)
	first := byte(0)
	for {
		b, err := reader.Peek(1)
		if err != nil {
			_ = file.Close()
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if isSpaceByte(b[0]) {
			_, _ = reader.ReadByte()
			continue
		}
		first = b[0]
		break
	}
	_ = file.Close()
	if first != '[' {
		return nil
	}
	items, err := d.readEventArchive()
	if err != nil {
		return err
	}
	tmp := d.eventArchivePath() + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(out)
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			_ = out.Close()
			return err
		}
		if _, err := writer.Write(raw); err != nil {
			_ = out.Close()
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			_ = out.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, d.eventArchivePath())
}

func isSpaceByte(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}
